package web

import (
	"crypto/rand"
	"encoding/hex"
	"errors"
	"fmt"
	"io"
	"mime/multipart"
	"net/http"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"time"

	"github.com/JoaoVictorCarvalho07/site-da-bebel-htmx/internal/store"
)

// Anexos de imagem nos briefings.
//
// Toda a postura deste arquivo parte de uma premissa: o arquivo vem de um
// ESTRANHO. Nada que acompanha o upload -- nome, extensao, Content-Type -- e
// confiavel; sao todos campos que o remetente escolhe. Por isso:
//
//   - o nome do arquivo em disco e GERADO por nos, nunca derivado do enviado;
//   - o tipo e decidido farejando os bytes, nunca lendo o cabecalho;
//   - o arquivo e servido so pela rota autenticada, com nosniff e
//     Content-Disposition (ver servirArquivo).

const (
	// 5 MB por imagem: uma foto de celular moderna cabe; um "zip disfarcado
	// de jpg" de 500 MB, nao.
	maxTamanhoArquivo = 5 << 20
	maxArquivos       = 4
	// Teto do corpo inteiro. Sem isto, um POST multipart gigante consome disco
	// (o Go faz spool em arquivo temporario) antes de qualquer validacao nossa.
	maxCorpoUpload = maxArquivos*maxTamanhoArquivo + (1 << 20)
	// Quanto lemos para farejar o tipo. http.DetectContentType so olha os
	// primeiros 512 bytes; para o box ftyp do HEIC isso e mais que suficiente.
	bytesParaFarejar = 512
)

// tiposAceitos mapeia o MIME farejado para a extensao que vamos gravar.
// Allowlist, nunca denylist: um formato desconhecido e recusado por padrao.
//
// SVG esta deliberadamente FORA. SVG e XML e pode conter <script>; servido do
// nosso dominio, viraria XSS com a sessao do painel da Isa junto.
var tiposAceitos = map[string]string{
	"image/jpeg": ".jpg",
	"image/png":  ".png",
	"image/webp": ".webp",
	"image/gif":  ".gif",
	"image/avif": ".avif",
	"image/heic": ".heic", // iPhone; ver farejarTipo
	"image/heif": ".heif",
}

// farejarTipo decide o tipo pelos BYTES. Devolve o mime e a extensao.
func farejarTipo(cabeca []byte) (string, string, bool) {
	mime := http.DetectContentType(cabeca)
	// DetectContentType devolve "image/jpeg; charset=..." em alguns casos.
	if i := strings.IndexByte(mime, ';'); i >= 0 {
		mime = strings.TrimSpace(mime[:i])
	}

	// O Go nao conhece HEIC/AVIF (container ISO-BMFF). Ambos se identificam
	// pelo box `ftyp` com a marca do formato nos bytes 4..12. Sem isto, foto
	// de iPhone -- o caso mais comum de quem responde pelo celular -- seria
	// recusada como "application/octet-stream".
	if len(cabeca) >= 12 && string(cabeca[4:8]) == "ftyp" {
		switch string(cabeca[8:12]) {
		case "heic", "heix", "hevc", "heim", "heis":
			mime = "image/heic"
		case "mif1", "msf1":
			mime = "image/heif"
		case "avif", "avis":
			mime = "image/avif"
		}
	}

	ext, ok := tiposAceitos[mime]
	return mime, ext, ok
}

// salvarAnexos grava os arquivos de UM campo e devolve os registros prontos.
// Erro aqui NAO invalida o briefing: o texto ja esta salvo, e perder a resposta
// inteira porque um anexo falhou seria trocar um problema pequeno por um grande.
func (s *Server) salvarAnexos(submissionID int64, campo string, arquivos []*multipart.FileHeader) []store.Arquivo {
	var salvos []store.Arquivo

	for i, cabecalho := range arquivos {
		if i >= maxArquivos {
			logger.Warn("upload: excedeu o limite de arquivos", "id", submissionID, "enviados", len(arquivos))
			break
		}
		if cabecalho.Size <= 0 || cabecalho.Size > maxTamanhoArquivo {
			logger.Warn("upload: tamanho recusado", "id", submissionID, "bytes", cabecalho.Size)
			continue
		}

		registro, err := s.gravarUmAnexo(submissionID, campo, cabecalho)
		if err != nil {
			logger.Error("upload: falhou", "id", submissionID, "arquivo", cabecalho.Filename, "erro", err)
			continue
		}
		salvos = append(salvos, registro)
	}
	return salvos
}

func (s *Server) gravarUmAnexo(submissionID int64, campo string, cabecalho *multipart.FileHeader) (store.Arquivo, error) {
	origem, err := cabecalho.Open()
	if err != nil {
		return store.Arquivo{}, err
	}
	defer origem.Close()

	cabeca := make([]byte, bytesParaFarejar)
	n, err := io.ReadFull(origem, cabeca)
	if err != nil && !errors.Is(err, io.ErrUnexpectedEOF) && !errors.Is(err, io.EOF) {
		return store.Arquivo{}, err
	}
	mime, ext, ok := farejarTipo(cabeca[:n])
	if !ok {
		return store.Arquivo{}, fmt.Errorf("tipo nao aceito: %s", mime)
	}
	if _, err := origem.Seek(0, io.SeekStart); err != nil {
		return store.Arquivo{}, err
	}

	// Nome proprio: <submissionID>-<aleatorio>.<ext>. Nenhum pedaco vem do
	// visitante, entao nao existe "../" nem colisao proposital.
	sufixo := make([]byte, 8)
	if _, err := rand.Read(sufixo); err != nil {
		return store.Arquivo{}, err
	}
	agora := time.Now()
	relativo := filepath.Join(
		agora.Format("2006"), agora.Format("01"),
		fmt.Sprintf("%d-%s%s", submissionID, hex.EncodeToString(sufixo), ext),
	)
	destino := filepath.Join(s.cfg.UploadDir, relativo)

	if err := os.MkdirAll(filepath.Dir(destino), 0o750); err != nil {
		return store.Arquivo{}, err
	}
	// O_EXCL: se o caminho ja existir (nao deveria), falha em vez de sobrescrever.
	f, err := os.OpenFile(destino, os.O_WRONLY|os.O_CREATE|os.O_EXCL, 0o640)
	if err != nil {
		return store.Arquivo{}, err
	}
	defer f.Close()

	// Copia com teto, mesmo o Size ja tendo sido conferido: o Size vem do
	// cabecalho multipart, que tambem e do remetente.
	escritos, err := io.Copy(f, io.LimitReader(origem, maxTamanhoArquivo))
	if err != nil {
		_ = os.Remove(destino)
		return store.Arquivo{}, err
	}

	return store.Arquivo{
		SubmissionID: submissionID,
		Campo:        campo,
		Caminho:      filepath.ToSlash(relativo),
		NomeOriginal: nomeSeguroParaExibir(cabecalho.Filename),
		Mime:         mime,
		Tamanho:      escritos,
	}, nil
}

// nomeSeguroParaExibir limpa o nome enviado para ele poder aparecer na tela da
// Isa. Nao vira caminho de arquivo em lugar nenhum -- e rotulo, so isso.
func nomeSeguroParaExibir(nome string) string {
	nome = filepath.Base(strings.ReplaceAll(nome, `\`, "/"))
	nome = strings.Map(func(r rune) rune {
		if r < 32 || r == 127 { // controle, incluindo o \n que quebraria o log
			return -1
		}
		return r
	}, nome)
	if len(nome) > 120 {
		nome = nome[:120]
	}
	if nome == "" || nome == "." || nome == ".." {
		return "imagem"
	}
	return nome
}

// servirArquivo entrega um anexo para a Isa. Rota SOB basic-auth (ver
// server.go): sao fotos de clientes, nunca podem ficar publicas.
func (s *Server) servirArquivo(w http.ResponseWriter, r *http.Request) {
	id, err := strconv.ParseInt(r.PathValue("id"), 10, 64)
	if err != nil {
		http.Error(w, "id inválido", http.StatusBadRequest)
		return
	}
	arquivo, err := s.store.Arquivo(r.Context(), id)
	if err != nil {
		s.notFound(w, r)
		return
	}

	// Defesa em profundidade: o caminho veio do NOSSO banco, mas conferir que
	// ele continua dentro do dir de uploads custa nada e fecha a porta caso
	// alguem, um dia, deixe um caminho entrar por outro lugar.
	base, err := filepath.Abs(s.cfg.UploadDir)
	if err != nil {
		http.Error(w, "Erro interno", http.StatusInternalServerError)
		return
	}
	caminho, err := filepath.Abs(filepath.Join(base, filepath.FromSlash(arquivo.Caminho)))
	if err != nil || !strings.HasPrefix(caminho, base+string(os.PathSeparator)) {
		logger.Error("upload: caminho fora do dir de uploads", "id", id, "caminho", arquivo.Caminho)
		s.notFound(w, r)
		return
	}

	f, err := os.Open(caminho)
	if err != nil {
		logger.Error("upload: arquivo sumiu do disco", "id", id, "caminho", caminho, "erro", err)
		s.notFound(w, r)
		return
	}
	defer f.Close()
	info, err := f.Stat()
	if err != nil {
		http.Error(w, "Erro interno", http.StatusInternalServerError)
		return
	}

	// nosniff: sem isto o navegador pode reinterpretar o conteudo e executar
	// algo que passou pelo nosso farejamento. inline + filename da a Isa o nome
	// certo ao salvar. Cache privado: nunca em proxy compartilhado.
	w.Header().Set("Content-Type", arquivo.Mime)
	w.Header().Set("X-Content-Type-Options", "nosniff")
	w.Header().Set("Content-Security-Policy", "default-src 'none'; sandbox")
	w.Header().Set("Content-Disposition",
		fmt.Sprintf("inline; filename*=UTF-8''%s", urlEncode(arquivo.NomeOriginal)))
	w.Header().Set("Cache-Control", "private, max-age=300")
	http.ServeContent(w, r, arquivo.NomeOriginal, info.ModTime(), f)
}
