package web

import (
	"errors"
	"net/http"
	"strconv"
	"strings"

	"github.com/JoaoVictorCarvalho07/site-da-bebel-htmx/internal/store"
)

var (
	errCampoObrigatorio  = errors.New("preencha os campos obrigatórios")
	errSelectSemOpcoes   = errors.New("uma lista precisa de ao menos uma opção")
	errApagarComResposta = errors.New("essa pergunta já tem respostas — esconda em vez de apagar, para não perder os briefings antigos")
)

// Editor de formularios de briefing (Fases B+C, docs §8b). A Isa edita textos,
// rotulos, ajuda, opcoes, obrigatoriedade e ordem, adiciona perguntas (Fase B),
// e ainda esconde/apaga/troca o tipo (Fase C) -- tudo pelo /admin, sem redeploy.
// Idioma: ela edita so PT; ao mudar um texto, o EN daquele campo e limpo e a
// pergunta e marcada en_stale (o dev revisa depois; o visitante EN ve o PT nesse
// meio-tempo, via fallback em ml.T).
//
// Modelo de UI: cada bloco (textos do form, cada pergunta, "adicionar") e um
// <form> proprio com POST + redirect (PRG). Sem JS obrigatorio, sem perder
// edicao de um bloco ao mexer noutro, e cada acao e atomica.
//
// A politica destrutiva (Fase C) e simples: esconder e trocar tipo sao sempre
// permitidos (reversiveis, nao perdem dado); APAGAR so com zero respostas -- com
// respostas, apagar orfanaria briefings, entao a Isa e mandada a esconder. A
// contagem de respostas aparece em cada pergunta para ela decidir informada.
// Ainda FORA: criar formulario novo do zero (precisa de pagina generica; §8b).

// tiposDisponiveis alimenta o seletor de "adicionar pergunta". A ordem e a que a
// Isa ve; o rotulo e amigavel, o valor e o tipoCampo interno.
var tiposDisponiveis = []struct{ Valor, Rotulo string }{
	{string(campoTexto), "Texto curto"},
	{string(campoTextarea), "Texto longo (parágrafo)"},
	{string(campoSelect), "Lista de opções"},
	{string(campoContato), "WhatsApp ou e-mail"},
	{string(campoEmail), "E-mail"},
	{string(campoData), "Data"},
	{string(campoArquivo), "Anexar imagens"},
}

// --- View models -------------------------------------------------------------

type adminFormsView struct {
	pageData
	Forms []formResumo
}

type formResumo struct {
	Slug         string
	Titulo       string
	NumPerguntas int
	Visivel      bool
}

type adminFormEditView struct {
	pageData
	Slug       string
	Titulo     string // PT, para o <h1>
	Textos     store.FormDef
	Perguntas  []perguntaEdit
	Tipos      []struct{ Valor, Rotulo string }
	PreviewURL string
	Salvo      bool // flash de sucesso apos o redirect
}

type perguntaEdit struct {
	store.QuestionDef
	TipoRotulo  string
	EhSelect    bool
	OpcoesTexto string                           // opcoes PT, uma por linha, para o <textarea>
	Primeira    bool                             // desabilita "subir"
	Ultima      bool                             // desabilita "descer"
	Respostas   int                              // quantas submissions responderam esta pergunta
	Tipos       []struct{ Valor, Rotulo string } // para o seletor de trocar tipo
}

// PodeApagar: so sem respostas. Com respostas, o caminho seguro e esconder.
func (p perguntaEdit) PodeApagar() bool { return p.Respostas == 0 }

// --- Handlers: leitura -------------------------------------------------------

func (s *Server) adminForms(w http.ResponseWriter, r *http.Request) {
	defs, err := s.store.LoadForms(r.Context())
	if err != nil {
		s.erroAdmin(w, "listar formularios", err)
		return
	}
	resumos := make([]formResumo, 0, len(defs))
	for _, d := range defs {
		resumos = append(resumos, formResumo{
			Slug: d.Slug, Titulo: d.TituloPT, NumPerguntas: len(d.Perguntas), Visivel: d.Visivel,
		})
	}
	dados := s.novaPagina("pt", "Formulários — painel", "", "/admin/forms")
	dados.OcultarRodape = true
	s.render.Page(w, r, "admin_forms", http.StatusOK, adminFormsView{pageData: dados, Forms: resumos})
}

func (s *Server) adminFormEdit(w http.ResponseWriter, r *http.Request) {
	slug := r.PathValue("slug")
	def, ok, err := s.store.LoadForm(r.Context(), slug)
	if err != nil {
		s.erroAdmin(w, "carregar formulario", err)
		return
	}
	if !ok {
		s.notFound(w, r)
		return
	}

	// Contagem de respostas por pergunta: alimenta a decisao de apagar/trocar
	// tipo. Uma query so para o form inteiro.
	contagem, err := s.store.ContarRespostasPorChave(r.Context(), slug)
	if err != nil {
		logger.Error("admin editor: contar respostas", "erro", err)
		contagem = map[string]int{} // sem contagem, a tela ainda funciona
	}

	perguntas := make([]perguntaEdit, len(def.Perguntas))
	for i, q := range def.Perguntas {
		perguntas[i] = perguntaEdit{
			QuestionDef: q,
			TipoRotulo:  rotuloTipo(q.Tipo),
			EhSelect:    tipoCampo(q.Tipo) == campoSelect,
			OpcoesTexto: opcoesParaTexto(q.OpcoesJSON),
			Primeira:    i == 0,
			Ultima:      i == len(def.Perguntas)-1,
			Respostas:   contagem[q.Chave],
			Tipos:       tiposDisponiveis,
		}
	}

	dados := s.novaPagina("pt", "Editar "+def.TituloPT, "", "/admin/forms/"+slug)
	dados.OcultarRodape = true
	s.render.Page(w, r, "admin_form_edit", http.StatusOK, adminFormEditView{
		pageData:   dados,
		Slug:       slug,
		Titulo:     def.TituloPT,
		Textos:     def,
		Perguntas:  perguntas,
		Tipos:      tiposDisponiveis,
		PreviewURL: previewURL(slug),
		Salvo:      r.URL.Query().Get("salvo") == "1",
	})
}

// --- Handlers: escrita -------------------------------------------------------

// adminFormSalvar grava os textos de nivel de formulario (PT). O EN de um campo
// e limpo se o PT mudou.
func (s *Server) adminFormSalvar(w http.ResponseWriter, r *http.Request) {
	slug := r.PathValue("slug")
	def, ok, err := s.store.LoadForm(r.Context(), slug)
	if err != nil || !ok {
		s.notFound(w, r)
		return
	}
	_ = r.ParseForm()

	def.TituloPT, def.TituloEN = casarPTeEN(def.TituloPT, def.TituloEN, campo(r, "titulo"))
	def.SubtituloPT, def.SubtituloEN = casarPTeEN(def.SubtituloPT, def.SubtituloEN, campo(r, "subtitulo"))
	def.EnviarPT, def.EnviarEN = casarPTeEN(def.EnviarPT, def.EnviarEN, campo(r, "enviar"))
	def.SucessoPT, def.SucessoEN = casarPTeEN(def.SucessoPT, def.SucessoEN, campo(r, "sucesso"))

	// Um formulario sem titulo ou sem rotulo de botao fica quebrado; recusa.
	if def.TituloPT == "" || def.EnviarPT == "" {
		s.erroAdmin(w, "validar formulario", errCampoObrigatorio)
		return
	}

	if err := s.store.AtualizarForm(r.Context(), def); err != nil {
		s.erroAdmin(w, "salvar formulario", err)
		return
	}
	s.aposEditar(w, r, slug)
}

// adminPerguntaSalvar grava os campos editaveis de UMA pergunta.
func (s *Server) adminPerguntaSalvar(w http.ResponseWriter, r *http.Request) {
	slug := r.PathValue("slug")
	id, err := strconv.ParseInt(r.PathValue("id"), 10, 64)
	if err != nil {
		http.Error(w, "id inválido", http.StatusBadRequest)
		return
	}
	q, ok := s.acharQuestion(r, slug, id)
	if !ok {
		s.notFound(w, r)
		return
	}
	_ = r.ParseForm()

	// Compara o PT novo com o atual: mudou => limpa o EN daquele campo e marca
	// en_stale. So-toggle-de-obrigatorio nao mexe no EN.
	novoLabel := campo(r, "label")
	novaAjuda := campo(r, "ajuda")
	if novoLabel == "" {
		s.erroAdmin(w, "validar pergunta", errCampoObrigatorio)
		return
	}

	if novoLabel != q.LabelPT {
		q.LabelPT, q.LabelEN, q.ENStale = novoLabel, "", true
	}
	if novaAjuda != q.AjudaPT {
		q.AjudaPT, q.AjudaEN, q.ENStale = novaAjuda, "", true
	}
	q.Obrigatoria = r.PostFormValue("obrigatoria") != ""

	if tipoCampo(q.Tipo) == campoSelect {
		novasOpcoes := textoParaOpcoes(campo(r, "opcoes"))
		if len(novasOpcoes) == 0 {
			s.erroAdmin(w, "validar opcoes", errSelectSemOpcoes)
			return
		}
		novoJSON := opcoesToJSON(novasOpcoes)
		if novoJSON != q.OpcoesJSON {
			q.OpcoesJSON, q.ENStale = novoJSON, true
		}
	}

	if err := s.store.AtualizarPergunta(r.Context(), q); err != nil {
		s.erroAdmin(w, "salvar pergunta", err)
		return
	}
	s.aposEditar(w, r, slug)
}

// adminPerguntaAdd cria uma pergunta nova no fim do formulario.
func (s *Server) adminPerguntaAdd(w http.ResponseWriter, r *http.Request) {
	slug := r.PathValue("slug")
	if !s.forms.existe(slug) {
		s.notFound(w, r)
		return
	}
	_ = r.ParseForm()

	label := campo(r, "label")
	tipo := campo(r, "tipo")
	if label == "" || !tipoValido(tipo) {
		s.erroAdmin(w, "validar nova pergunta", errCampoObrigatorio)
		return
	}

	q := store.QuestionDef{
		LabelPT:     label,
		AjudaPT:     campo(r, "ajuda"),
		Tipo:        tipo,
		Obrigatoria: r.PostFormValue("obrigatoria") != "",
		ENStale:     true, // recem-criada: sem traducao ainda
	}
	if tipoCampo(tipo) == campoSelect {
		opcoes := textoParaOpcoes(campo(r, "opcoes"))
		if len(opcoes) == 0 {
			s.erroAdmin(w, "validar opcoes", errSelectSemOpcoes)
			return
		}
		q.OpcoesJSON = opcoesToJSON(opcoes)
	}

	if _, err := s.store.AdicionarPergunta(r.Context(), slug, q); err != nil {
		s.erroAdmin(w, "adicionar pergunta", err)
		return
	}
	s.aposEditar(w, r, slug)
}

// adminPerguntaOcultar liga/desliga o soft-delete (esconder/mostrar). Sempre
// permitido: e reversivel e nao perde dado.
func (s *Server) adminPerguntaOcultar(w http.ResponseWriter, r *http.Request) {
	slug := r.PathValue("slug")
	id, err := strconv.ParseInt(r.PathValue("id"), 10, 64)
	if err != nil {
		http.Error(w, "id inválido", http.StatusBadRequest)
		return
	}
	ocultar := r.PostFormValue("oculta") == "1"
	if err := s.store.OcultarPergunta(r.Context(), id, ocultar); err != nil {
		s.erroAdmin(w, "esconder pergunta", err)
		return
	}
	s.aposEditar(w, r, slug)
}

// adminPerguntaApagar remove a pergunta de vez -- mas SO se ela nao tiver
// resposta nenhuma. Com respostas, a irreversibilidade orfanaria briefings; a
// Isa e mandada a esconder. Checa no servidor, nunca confia so na UI.
func (s *Server) adminPerguntaApagar(w http.ResponseWriter, r *http.Request) {
	slug := r.PathValue("slug")
	id, err := strconv.ParseInt(r.PathValue("id"), 10, 64)
	if err != nil {
		http.Error(w, "id inválido", http.StatusBadRequest)
		return
	}
	q, ok := s.acharQuestion(r, slug, id)
	if !ok {
		s.notFound(w, r)
		return
	}
	contagem, err := s.store.ContarRespostasPorChave(r.Context(), slug)
	if err != nil {
		s.erroAdmin(w, "contar respostas", err)
		return
	}
	if contagem[q.Chave] > 0 {
		s.erroAdmin(w, "apagar pergunta", errApagarComResposta)
		return
	}
	if err := s.store.ApagarPergunta(r.Context(), id); err != nil {
		s.erroAdmin(w, "apagar pergunta", err)
		return
	}
	s.aposEditar(w, r, slug)
}

// adminPerguntaTrocarTipo muda o tipo de uma pergunta existente. Permitido mesmo
// com respostas (elas mantem o texto gravado); a contagem so aparece na tela
// para a Isa decidir informada. Trocar PARA lista exige opcoes.
func (s *Server) adminPerguntaTrocarTipo(w http.ResponseWriter, r *http.Request) {
	slug := r.PathValue("slug")
	id, err := strconv.ParseInt(r.PathValue("id"), 10, 64)
	if err != nil {
		http.Error(w, "id inválido", http.StatusBadRequest)
		return
	}
	if _, ok := s.acharQuestion(r, slug, id); !ok {
		s.notFound(w, r)
		return
	}
	_ = r.ParseForm()

	tipo := campo(r, "tipo")
	if !tipoValido(tipo) {
		s.erroAdmin(w, "validar tipo", errCampoObrigatorio)
		return
	}
	opcoesJSON := ""
	if tipoCampo(tipo) == campoSelect {
		opcoes := textoParaOpcoes(campo(r, "opcoes"))
		if len(opcoes) == 0 {
			s.erroAdmin(w, "validar opcoes", errSelectSemOpcoes)
			return
		}
		opcoesJSON = opcoesToJSON(opcoes)
	}
	if err := s.store.TrocarTipoPergunta(r.Context(), id, tipo, opcoesJSON); err != nil {
		s.erroAdmin(w, "trocar tipo", err)
		return
	}
	s.aposEditar(w, r, slug)
}

// adminPerguntaMover sobe ou desce uma pergunta na ordem.
func (s *Server) adminPerguntaMover(w http.ResponseWriter, r *http.Request) {
	slug := r.PathValue("slug")
	id, err := strconv.ParseInt(r.PathValue("id"), 10, 64)
	if err != nil {
		http.Error(w, "id inválido", http.StatusBadRequest)
		return
	}
	paraCima := r.PostFormValue("dir") == "cima"
	if err := s.store.MoverPergunta(r.Context(), slug, id, paraCima); err != nil {
		s.erroAdmin(w, "mover pergunta", err)
		return
	}
	s.aposEditar(w, r, slug)
}

// --- Apoio -------------------------------------------------------------------

// aposEditar recarrega o cache (para o site publico refletir a edicao na hora) e
// redireciona de volta para a pagina de edicao (PRG: o F5 nao reenvia o form).
func (s *Server) aposEditar(w http.ResponseWriter, r *http.Request, slug string) {
	if err := s.forms.reload(); err != nil {
		// A gravacao ja aconteceu; so o cache ficou velho. Loga e segue -- o
		// proximo boot/edicao conserta, e e melhor que devolver erro a Isa.
		logger.Error("admin: reload do cache falhou apos edicao", "erro", err)
	}
	http.Redirect(w, r, "/admin/forms/"+slug+"?salvo=1", http.StatusSeeOther)
}

// acharQuestion carrega a pergunta atual (do banco) para o handler de save
// partir do estado real, nao do que veio no POST.
func (s *Server) acharQuestion(r *http.Request, slug string, id int64) (store.QuestionDef, bool) {
	def, ok, err := s.store.LoadForm(r.Context(), slug)
	if err != nil || !ok {
		return store.QuestionDef{}, false
	}
	for _, q := range def.Perguntas {
		if q.ID == id {
			return q, true
		}
	}
	return store.QuestionDef{}, false
}

func (s *Server) erroAdmin(w http.ResponseWriter, contexto string, err error) {
	logger.Error("admin editor: "+contexto, "erro", err)
	http.Error(w, "Não foi possível salvar: "+err.Error(), http.StatusUnprocessableEntity)
}

// campo devolve o valor de um campo de texto ja trimado.
func campo(r *http.Request, nome string) string {
	return strings.TrimSpace(r.PostFormValue(nome))
}

// casarPTeEN aplica a edicao PT: se o texto mudou, o EN e limpo (fica stale, cai
// pro PT no site via ml.T). Devolve (novoPT, novoEN).
func casarPTeEN(ptAtual, enAtual, ptNovo string) (string, string) {
	if ptNovo != ptAtual {
		return ptNovo, ""
	}
	return ptAtual, enAtual
}

// textoParaOpcoes quebra o textarea (uma opcao por linha) em opcoes PT, ignorando
// linhas em branco. O EN fica vazio (cai pro PT ate o dev traduzir).
func textoParaOpcoes(texto string) []ml {
	var opcoes []ml
	for _, linha := range strings.Split(texto, "\n") {
		linha = strings.TrimSpace(linha)
		if linha != "" {
			opcoes = append(opcoes, ml{pt: linha})
		}
	}
	return opcoes
}

// opcoesParaTexto faz o caminho inverso, para preencher o textarea de edicao com
// as opcoes PT atuais.
func opcoesParaTexto(opcoesJSON string) string {
	opcoes := opcoesFromJSON(opcoesJSON)
	linhas := make([]string, len(opcoes))
	for i, o := range opcoes {
		linhas[i] = o.pt
	}
	return strings.Join(linhas, "\n")
}

func tipoValido(t string) bool {
	for _, d := range tiposDisponiveis {
		if d.Valor == t {
			return true
		}
	}
	return false
}

func rotuloTipo(t string) string {
	for _, d := range tiposDisponiveis {
		if d.Valor == t {
			return d.Rotulo
		}
	}
	return t
}

// previewURL aponta para onde o formulario aparece ao vivo. O de UGC vive
// embutido no midia kit; os demais tem pagina propria em /briefing/<slug>... mas
// hoje so /briefing (fotografia) e /midiakit existem como paginas. Ver §8b.
func previewURL(slug string) string {
	switch slug {
	case "fotografia":
		return "/briefing"
	case "ugc":
		return "/midiakit#briefing"
	default:
		return "/briefing" // forms novos ainda nao tem pagina propria (Fase C)
	}
}
