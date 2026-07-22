package web

import (
	"bytes"
	"context"
	"mime/multipart"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"testing"
)

// Cabecalhos reais dos formatos. O farejamento olha bytes, entao o teste
// precisa mandar bytes de verdade -- string "fake png" nao exercita nada.
var (
	bytesPNG  = []byte("\x89PNG\r\n\x1a\n\x00\x00\x00\rIHDR")
	bytesJPEG = []byte("\xff\xd8\xff\xe0\x00\x10JFIF\x00\x01")
	bytesGIF  = []byte("GIF89a\x01\x00\x01\x00\x00\x00\x00")
	// ISO-BMFF: tamanho do box, "ftyp", marca. E como iPhone entrega foto.
	bytesHEIC = []byte("\x00\x00\x00\x18ftypheic\x00\x00\x00\x00heic")
	bytesAVIF = []byte("\x00\x00\x00\x1cftypavif\x00\x00\x00\x00avif")
)

// postMultipart monta um POST multipart com campos de texto e arquivos.
func postMultipart(t *testing.T, h http.Handler, caminho string, campos map[string]string, arquivos map[string][]byte) *httptest.ResponseRecorder {
	t.Helper()
	var corpo bytes.Buffer
	escritor := multipart.NewWriter(&corpo)

	for chave, valor := range campos {
		if err := escritor.WriteField(chave, valor); err != nil {
			t.Fatal(err)
		}
	}
	for nome, conteudo := range arquivos {
		parte, err := escritor.CreateFormFile("referencias_imagens", nome)
		if err != nil {
			t.Fatal(err)
		}
		if _, err := parte.Write(conteudo); err != nil {
			t.Fatal(err)
		}
	}
	if err := escritor.Close(); err != nil {
		t.Fatal(err)
	}

	r := httptest.NewRequest(http.MethodPost, caminho, &corpo)
	r.Header.Set("Content-Type", escritor.FormDataContentType())
	r.Header.Set("HX-Request", "true")
	r.RemoteAddr = randomIP(t)
	w := httptest.NewRecorder()
	h.ServeHTTP(w, r)
	return w
}

func camposValidos() map[string]string {
	return map[string]string{
		"nome":      "Maria Souza",
		"contato":   "41999998888",
		"tipo":      "Ensaio individual",
		"motivacao": "Quero registrar meus 30 anos.",
	}
}

// TestUploadGravaImagem: o caminho feliz de ponta a ponta -- o arquivo chega ao
// disco, a linha chega ao banco, e o nome gravado NAO e o que o visitante mandou.
func TestUploadGravaImagem(t *testing.T) {
	s := novoServidor(t)
	dir := t.TempDir()
	s.cfg.UploadDir = dir

	w := postMultipart(t, s.Handler(), "/briefing/fotografia", camposValidos(),
		map[string][]byte{"minha foto.png": bytesPNG})

	if !strings.Contains(w.Body.String(), "Recebi seu briefing") {
		t.Fatalf("envio falhou:\n%s", w.Body.String())
	}

	subs, _ := s.store.ListarSubmissions(context.Background(), "", 10)
	if len(subs) != 1 {
		t.Fatalf("gravou %d briefings, quero 1", len(subs))
	}
	anexos, err := s.store.ArquivosDe(context.Background(), []int64{subs[0].ID})
	if err != nil {
		t.Fatal(err)
	}
	lista := anexos[subs[0].ID]
	if len(lista) != 1 {
		t.Fatalf("gravou %d anexos, quero 1", len(lista))
	}

	a := lista[0]
	if a.Mime != "image/png" {
		t.Errorf("mime = %q, quero image/png", a.Mime)
	}
	if a.NomeOriginal != "minha foto.png" {
		t.Errorf("nome de exibicao = %q", a.NomeOriginal)
	}
	// O nome EM DISCO tem que ser nosso, nao o do visitante.
	if strings.Contains(a.Caminho, "minha foto") {
		t.Errorf("o caminho em disco usou o nome enviado: %q", a.Caminho)
	}
	if _, err := os.Stat(filepath.Join(dir, filepath.FromSlash(a.Caminho))); err != nil {
		t.Errorf("arquivo nao chegou ao disco: %v", err)
	}
}

// TestUploadRecusaNaoImagem: o coracao da defesa. Um HTML disfarcado de .png
// servido do nosso dominio seria XSS -- com a sessao do painel da Isa junto.
// A extensao e o Content-Type sao do remetente; so os bytes valem.
func TestUploadRecusaNaoImagem(t *testing.T) {
	perigosos := map[string][]byte{
		"a.png":  []byte("<html><script>alert(document.cookie)</script></html>"),
		"b.jpg":  []byte("#!/bin/sh\nrm -rf /\n"),
		"c.webp": []byte("%PDF-1.4\n%\xe2\xe3\xcf\xd3\n"),
		// SVG e imagem de verdade, mas e XML executavel: fora da allowlist.
		"d.svg": []byte(`<svg xmlns="http://www.w3.org/2000/svg"><script>alert(1)</script></svg>`),
	}

	for nome, conteudo := range perigosos {
		t.Run(nome, func(t *testing.T) {
			s := novoServidor(t)
			s.cfg.UploadDir = t.TempDir()

			w := postMultipart(t, s.Handler(), "/briefing/fotografia", camposValidos(),
				map[string][]byte{nome: conteudo})

			// O briefing em si e aceito (o texto vale); so o anexo e descartado.
			if !strings.Contains(w.Body.String(), "Recebi seu briefing") {
				t.Fatalf("o briefing deveria ser aceito mesmo com anexo ruim:\n%s", w.Body.String())
			}
			subs, _ := s.store.ListarSubmissions(context.Background(), "", 10)
			anexos, _ := s.store.ArquivosDe(context.Background(), []int64{subs[0].ID})
			if len(anexos[subs[0].ID]) != 0 {
				t.Errorf("aceitou %q como imagem", nome)
			}
		})
	}
}

// TestUploadAceitaFormatosDeCelular: HEIC e AVIF sao ISO-BMFF e o
// http.DetectContentType do Go nao os conhece. Sem o reconhecimento do box
// ftyp, foto de iPhone -- o caso mais comum de quem responde pelo celular --
// seria recusada.
func TestUploadAceitaFormatosDeCelular(t *testing.T) {
	casos := map[string]struct {
		conteudo []byte
		mime     string
	}{
		"foto.heic": {bytesHEIC, "image/heic"},
		"foto.avif": {bytesAVIF, "image/avif"},
		"foto.jpg":  {bytesJPEG, "image/jpeg"},
		"foto.gif":  {bytesGIF, "image/gif"},
	}

	for nome, caso := range casos {
		t.Run(nome, func(t *testing.T) {
			s := novoServidor(t)
			s.cfg.UploadDir = t.TempDir()

			postMultipart(t, s.Handler(), "/briefing/fotografia", camposValidos(),
				map[string][]byte{nome: caso.conteudo})

			subs, _ := s.store.ListarSubmissions(context.Background(), "", 10)
			anexos, _ := s.store.ArquivosDe(context.Background(), []int64{subs[0].ID})
			lista := anexos[subs[0].ID]
			if len(lista) != 1 {
				t.Fatalf("recusou %s", nome)
			}
			if lista[0].Mime != caso.mime {
				t.Errorf("mime = %q, quero %q", lista[0].Mime, caso.mime)
			}
		})
	}
}

// TestUploadNomeMaliciosoNaoViraCaminho: o nome enviado nunca pode escapar do
// diretorio de uploads. E o path traversal classico.
func TestUploadNomeMaliciosoNaoViraCaminho(t *testing.T) {
	s := novoServidor(t)
	dir := t.TempDir()
	s.cfg.UploadDir = dir

	postMultipart(t, s.Handler(), "/briefing/fotografia", camposValidos(),
		map[string][]byte{`../../../../etc/cron.d/x.png`: bytesPNG})

	subs, _ := s.store.ListarSubmissions(context.Background(), "", 10)
	anexos, _ := s.store.ArquivosDe(context.Background(), []int64{subs[0].ID})
	lista := anexos[subs[0].ID]
	if len(lista) != 1 {
		t.Fatal("o anexo deveria ter sido aceito (o nome e que precisa ser neutralizado)")
	}
	if strings.Contains(lista[0].Caminho, "..") {
		t.Errorf("o caminho gravado contem '..': %q", lista[0].Caminho)
	}
	// A prova final: o arquivo esta DENTRO do dir de uploads.
	destino, err := filepath.Abs(filepath.Join(dir, filepath.FromSlash(lista[0].Caminho)))
	if err != nil {
		t.Fatal(err)
	}
	base, _ := filepath.Abs(dir)
	if !strings.HasPrefix(destino, base+string(os.PathSeparator)) {
		t.Errorf("arquivo gravado fora do dir de uploads: %q", destino)
	}
}

// TestUploadLimiteDeQuantidade: acima de maxArquivos, o excedente e ignorado --
// sem derrubar o briefing.
func TestUploadLimiteDeQuantidade(t *testing.T) {
	s := novoServidor(t)
	s.cfg.UploadDir = t.TempDir()

	arquivos := map[string][]byte{}
	for i := 0; i < maxArquivos+3; i++ {
		arquivos["foto"+strconv.Itoa(i)+".png"] = bytesPNG
	}
	postMultipart(t, s.Handler(), "/briefing/fotografia", camposValidos(), arquivos)

	subs, _ := s.store.ListarSubmissions(context.Background(), "", 10)
	anexos, _ := s.store.ArquivosDe(context.Background(), []int64{subs[0].ID})
	if n := len(anexos[subs[0].ID]); n != maxArquivos {
		t.Errorf("gravou %d anexos, quero no maximo %d", n, maxArquivos)
	}
}

// TestArquivoExigeAutenticacao: as imagens sao fotos de clientes. A rota que as
// serve nao pode responder sem credencial -- nem com o id certo.
func TestArquivoExigeAutenticacao(t *testing.T) {
	s := novoServidor(t)
	s.cfg.UploadDir = t.TempDir()
	s.cfg.AdminUser, s.cfg.AdminSenha = "isa", "segredo"
	h := s.Handler()

	postMultipart(t, h, "/briefing/fotografia", camposValidos(),
		map[string][]byte{"foto.png": bytesPNG})

	subs, _ := s.store.ListarSubmissions(context.Background(), "", 10)
	anexos, _ := s.store.ArquivosDe(context.Background(), []int64{subs[0].ID})
	url := "/admin/arquivo/" + strconv.FormatInt(anexos[subs[0].ID][0].ID, 10)

	w := httptest.NewRecorder()
	h.ServeHTTP(w, httptest.NewRequest(http.MethodGet, url, nil))
	if w.Code != http.StatusUnauthorized {
		t.Errorf("sem auth: status = %d, quero 401", w.Code)
	}

	w = httptest.NewRecorder()
	r := httptest.NewRequest(http.MethodGet, url, nil)
	r.SetBasicAuth("isa", "segredo")
	h.ServeHTTP(w, r)
	if w.Code != http.StatusOK {
		t.Fatalf("com auth: status = %d, quero 200", w.Code)
	}
	if !bytes.Equal(w.Body.Bytes(), bytesPNG) {
		t.Error("o conteudo servido nao bate com o enviado")
	}
	// nosniff evita que o navegador reinterprete o conteudo e execute algo que
	// tenha passado pelo farejamento.
	if w.Header().Get("X-Content-Type-Options") != "nosniff" {
		t.Error("falta o X-Content-Type-Options: nosniff")
	}
	if w.Header().Get("Content-Type") != "image/png" {
		t.Errorf("Content-Type = %q", w.Header().Get("Content-Type"))
	}
}

// TestAutoLinkNoPainel protege o ponto onde a garantia migrou.
//
// Antes havia um campo de URL dedicado, validado na ENTRADA. Ao fundi-lo com o
// texto livre de referencias, a unica defesa passou a ser a RENDERIZACAO: o
// autoLink so transforma em <a> o que casa com http(s)://, entao esquema
// perigoso nunca vira link -- e todo o resto sai escapado.
func TestAutoLinkNoPainel(t *testing.T) {
	perigosos := []string{
		"javascript:alert(document.cookie)",
		"data:text/html;base64,PHNjcmlwdD5hbGVydCgxKTwvc2NyaXB0Pg==",
		"file:///etc/passwd",
		`<script>alert(1)</script>`,
		`" onmouseover="alert(1)`,
	}
	for _, ruim := range perigosos {
		saida := string(autoLink(ruim))
		if strings.Contains(saida, "<a href") {
			t.Errorf("virou link: %q -> %q", ruim, saida)
		}
		if strings.Contains(saida, "<script") || strings.Contains(saida, `onmouseover="`) {
			t.Errorf("marcacao nao escapada: %q -> %q", ruim, saida)
		}
	}

	// O caso util: link colado no meio da prosa vira clicavel.
	saida := string(autoLink("adoro esse board https://pinterest.com/ana/x, o que acha?"))
	if !strings.Contains(saida, `<a href="https://pinterest.com/ana/x"`) {
		t.Errorf("nao linkou a URL: %q", saida)
	}
	// A virgula final nao pode entrar no href.
	if strings.Contains(saida, `x,"`) {
		t.Errorf("pontuacao final entrou no link: %q", saida)
	}
}

// TestContatoValidado protege o campo por onde a Isa RESPONDE. Um numero
// incompleto aqui custa o cliente inteiro, e em silencio -- ninguem descobre
// que a resposta nao chegou.
func TestContatoValidado(t *testing.T) {
	aceitos := []string{
		"41999998888",       // celular, so digitos
		"(41) 99999-8888",   // formatado
		"+55 41 99999-8888", // com DDI
		"5541999998888",     // DDI colado
		"4133334444",        // fixo
		"55999887766",       // DDD 55 (Santa Maria) -- nao pode ser lido como DDI
		"maria@exemplo.com",
		"maria.souza@empresa.com.br",
	}
	for _, bom := range aceitos {
		if !contatoValido(bom) {
			t.Errorf("recusou contato valido: %q", bom)
		}
	}

	recusados := []string{
		"41 9999-888", // digitos de menos
		"999998888",   // sem DDD
		"quero conversar",
		"maria@",
		"maria@exemplo", // sem ponto no dominio
		"@exemplo.com",
		"1199999888812345", // digitos demais
	}
	for _, ruim := range recusados {
		if contatoValido(ruim) {
			t.Errorf("aceitou contato invalido: %q", ruim)
		}
	}
}

// TestContatoInvalidoBloqueiaEnvio: a validacao precisa estar LIGADA no fluxo,
// nao so existir como funcao.
func TestContatoInvalidoBloqueiaEnvio(t *testing.T) {
	s := novoServidor(t)
	campos := camposValidos()
	campos["contato"] = "me liga ai"

	w := postMultipart(t, s.Handler(), "/briefing/fotografia", campos, nil)
	if !strings.Contains(w.Body.String(), "Confere esse contato") {
		t.Errorf("aceitou contato invalido:\n%s", w.Body.String())
	}
	if subs, _ := s.store.ListarSubmissions(context.Background(), "", 10); len(subs) != 0 {
		t.Error("gravou o briefing com contato invalido")
	}
}

// TestAdminMostraAnexo: a imagem precisa aparecer no painel, apontando para a
// rota autenticada -- nunca para um caminho de disco.
func TestAdminMostraAnexo(t *testing.T) {
	s := novoServidor(t)
	s.cfg.UploadDir = t.TempDir()
	s.cfg.AdminUser, s.cfg.AdminSenha = "isa", "segredo"
	h := s.Handler()

	postMultipart(t, h, "/briefing/fotografia", camposValidos(),
		map[string][]byte{"referencia.png": bytesPNG})

	w := httptest.NewRecorder()
	r := httptest.NewRequest(http.MethodGet, "/admin", nil)
	r.SetBasicAuth("isa", "segredo")
	h.ServeHTTP(w, r)

	corpo := w.Body.String()
	if !strings.Contains(corpo, "/admin/arquivo/") {
		t.Error("o painel nao linka o anexo")
	}
	if !strings.Contains(corpo, "referencia.png") {
		t.Error("o painel nao mostra o nome do arquivo")
	}
	if strings.Contains(corpo, s.cfg.UploadDir) {
		t.Error("o painel vazou o caminho do disco no HTML")
	}
}
