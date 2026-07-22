package web

import (
	"crypto/subtle"
	"net/http"
	"strconv"
	"time"

	"github.com/JoaoVictorCarvalho07/site-da-bebel-htmx/internal/store"
)

// Painel da Isa: le os briefings recebidos direto do SQLite. Existe para as
// respostas nao dependerem do e-mail -- se o Resend nao estiver configurado (ou
// cair), este e o caminho que sempre funciona.

// exigirAdmin protege o painel com basic-auth.
//
// Basic-auth e nao sessao/login: sao duas rotas para uma pessoa so. Uma tabela
// de usuarios, hash de senha e cookie de sessao seriam mais codigo para auditar
// sem ganho real aqui. O que isso EXIGE em troca e HTTPS -- em basic-auth a
// senha viaja no header a cada request, em base64 (que nao e criptografia).
// Enquanto o site estiver em http:// puro, trate esta senha como descartavel.
func (s *Server) exigirAdmin(proximo http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Sem senha configurada o painel NAO existe: 404, nao 500 nem acesso
		// livre. Um deploy que esqueceu a variavel nao pode abrir os briefings
		// para a internet inteira.
		if s.cfg.AdminUser == "" || s.cfg.AdminSenha == "" {
			s.notFound(w, r)
			return
		}

		usuario, senha, temAuth := r.BasicAuth()
		// subtle.ConstantTimeCompare: comparar com == vaza o tamanho do prefixo
		// correto pelo tempo de resposta. O & (nao &&) e proposital -- avalia os
		// dois lados sempre, sem short-circuit.
		okUser := subtle.ConstantTimeCompare([]byte(usuario), []byte(s.cfg.AdminUser)) == 1
		okSenha := subtle.ConstantTimeCompare([]byte(senha), []byte(s.cfg.AdminSenha)) == 1
		if !temAuth || !(okUser && okSenha) {
			w.Header().Set("WWW-Authenticate", `Basic realm="Painel", charset="UTF-8"`)
			http.Error(w, "Não autorizado", http.StatusUnauthorized)
			return
		}
		proximo.ServeHTTP(w, r)
	})
}

// linhaAdmin e uma resposta ja casada com as perguntas do formulario, pronta
// para o template -- o pareamento pergunta/resposta e logica, nao apresentacao.
type linhaAdmin struct {
	ID       int64
	Form     string
	Titulo   string
	Lida     bool
	Quando   string
	Campos   []campoAdmin
	Destaque string // primeiro campo identificador (nome/marca), para o cabecalho
	Anexos   []anexoAdmin
}

type anexoAdmin struct {
	URL  string
	Nome string
	KB   int64
}

type campoAdmin struct {
	Label string
	Valor string
	// Link faz o template renderizar um <a>. O esquema ja foi validado na
	// entrada (urlPlausivel), e o html/template recusa href perigoso na saida --
	// duas camadas, de proposito.
	Link bool
}

type adminView struct {
	pageData
	Linhas    []linhaAdmin
	NaoLidas  int
	Total     int
	FiltroURL string
	// EmailAtivo diz se o Resend esta configurado -- o painel avisa quando nao,
	// senao a Isa fica esperando um e-mail que nunca vem.
	EmailAtivo bool
}

func (s *Server) admin(w http.ResponseWriter, r *http.Request) {
	filtro := r.URL.Query().Get("form")
	if _, existe := formsPorSlug[filtro]; !existe {
		filtro = "" // valor invalido na query nao filtra nada, so e ignorado
	}

	subs, err := s.store.ListarSubmissions(r.Context(), filtro, 200)
	if err != nil {
		logger.Error("admin: listar falhou", "erro", err)
		http.Error(w, "Erro interno", http.StatusInternalServerError)
		return
	}

	// Uma query para TODOS os anexos, nao uma por card: com 200 briefings na
	// tela, o padrao ingenuo viraria 200 idas ao banco.
	ids := make([]int64, len(subs))
	for i, sub := range subs {
		ids[i] = sub.ID
	}
	anexos, err := s.store.ArquivosDe(r.Context(), ids)
	if err != nil {
		// Anexo e complemento: sem ele o painel ainda cumpre o essencial, que e
		// mostrar o texto do briefing.
		logger.Error("admin: carregar anexos falhou", "erro", err)
		anexos = map[int64][]store.Arquivo{}
	}

	linhas := make([]linhaAdmin, 0, len(subs))
	for _, sub := range subs {
		f, conhecido := formsPorSlug[sub.Form]
		linha := linhaAdmin{
			ID: sub.ID, Form: sub.Form, Lida: sub.Lida, Anexos: paraAnexos(anexos[sub.ID]),
			// Horario de Brasilia: o banco grava UTC, e a Isa nao deveria
			// precisar fazer a conta de cabeca.
			Quando: sub.CriadoEm.In(fusoBrasil).Format("02/01/2006 15:04"),
		}
		if !conhecido {
			// Formulario removido do codigo: mostra o cru em vez de sumir com a
			// resposta. Perder um briefing por refactor seria imperdoavel.
			linha.Titulo = sub.Form
			for chave, valor := range sub.Respostas {
				linha.Campos = append(linha.Campos, campoAdmin{Label: chave, Valor: valor})
			}
			linhas = append(linhas, linha)
			continue
		}

		linha.Titulo = f.Titulo.T("pt")
		for _, p := range f.Perguntas {
			valor := sub.Respostas[p.Chave]
			if valor == "" {
				continue
			}
			if linha.Destaque == "" && p.Obrigatoria && p.Tipo == campoTexto {
				linha.Destaque = valor
			}
			linha.Campos = append(linha.Campos, campoAdmin{
				Label: p.Label.T("pt"), Valor: valor, Link: p.Tipo == campoURL,
			})
		}
		linhas = append(linhas, linha)
	}

	dados := s.novaPagina("pt", "Painel — briefings", "", "/admin")
	dados.OcultarRodape = true
	s.render.Page(w, r, "admin", http.StatusOK, adminView{
		pageData:   dados,
		Linhas:     linhas,
		NaoLidas:   s.store.ContarNaoLidas(r.Context()),
		Total:      len(linhas),
		FiltroURL:  filtro,
		EmailAtivo: s.cfg.ResendAPIKey != "" && s.cfg.MailPara != "",
	})
}

func paraAnexos(arquivos []store.Arquivo) []anexoAdmin {
	lista := make([]anexoAdmin, 0, len(arquivos))
	for _, a := range arquivos {
		lista = append(lista, anexoAdmin{
			// Pelo ID, nunca pelo caminho: a URL nao pode ser um ponteiro para
			// o sistema de arquivos.
			URL:  "/admin/arquivo/" + strconv.FormatInt(a.ID, 10),
			Nome: a.NomeOriginal,
			KB:   (a.Tamanho + 1023) / 1024,
		})
	}
	return lista
}

// fusoBrasil resolvido uma vez. LoadLocation depende do tzdata do sistema, que
// pode nao existir num container minimo -- por isso o fallback fixo em -03:00.
// O Brasil nao tem horario de verao desde 2019, entao o offset fixo esta certo.
var fusoBrasil = func() *time.Location {
	if loc, err := time.LoadLocation("America/Sao_Paulo"); err == nil {
		return loc
	}
	return time.FixedZone("BRT", -3*60*60)
}()

// adminMarcarLida alterna o status de lida e devolve a lista atualizada.
func (s *Server) adminMarcarLida(w http.ResponseWriter, r *http.Request) {
	id, err := strconv.ParseInt(r.PathValue("id"), 10, 64)
	if err != nil {
		http.Error(w, "id inválido", http.StatusBadRequest)
		return
	}
	if err := s.store.MarcarLida(r.Context(), id, r.PostFormValue("lida") != "0"); err != nil {
		logger.Error("admin: marcar lida falhou", "id", id, "erro", err)
		http.Error(w, "Erro interno", http.StatusInternalServerError)
		return
	}
	// Redirect apos POST: evita o "reenviar formulario?" do navegador no F5.
	http.Redirect(w, r, "/admin", http.StatusSeeOther)
}
