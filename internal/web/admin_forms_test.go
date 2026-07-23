package web

import (
	"context"
	"net/http"
	"net/http/httptest"
	"net/url"
	"strconv"
	"strings"
	"testing"
)

// Testes do editor de formularios (Fase B, docs §8b). O foco sao as garantias
// que protegem os dados: a chave gerada e opaca, editar PT marca en_stale sem
// derrubar o EN de quem so mexeu num toggle, opcoes vazias sao recusadas, e o
// CSRF barra POST de outra origem.

func servidorAdmin(t *testing.T) *Server {
	t.Helper()
	s := novoServidor(t)
	s.cfg.AdminUser, s.cfg.AdminSenha = "isa", "segredo"
	return s
}

// postAdmin faz um POST autenticado E com Origin da propria pagina (senao o
// CSRF barra). httptest usa host "example.com".
func postAdmin(t *testing.T, s *Server, caminho string, form url.Values) *httptest.ResponseRecorder {
	t.Helper()
	r := httptest.NewRequest(http.MethodPost, caminho, strings.NewReader(form.Encode()))
	r.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	r.Header.Set("Origin", "http://"+r.Host)
	r.SetBasicAuth("isa", "segredo")
	w := httptest.NewRecorder()
	s.Handler().ServeHTTP(w, r)
	return w
}

// qid acha o id de uma pergunta pela chave.
func qid(t *testing.T, s *Server, slug, chave string) int64 {
	t.Helper()
	def, ok, err := s.store.LoadForm(context.Background(), slug)
	if err != nil || !ok {
		t.Fatalf("form %q nao carrega: %v", slug, err)
	}
	for _, q := range def.Perguntas {
		if q.Chave == chave {
			return q.ID
		}
	}
	t.Fatalf("pergunta %q nao existe em %q", chave, slug)
	return 0
}

func questionDef(t *testing.T, s *Server, slug, chave string) (q QuestionDefLike) {
	t.Helper()
	def, _, _ := s.store.LoadForm(context.Background(), slug)
	for _, x := range def.Perguntas {
		if x.Chave == chave {
			return QuestionDefLike{x.LabelPT, x.LabelEN, x.Obrigatoria, x.ENStale, x.Chave}
		}
	}
	t.Fatalf("pergunta %q sumiu", chave)
	return
}

type QuestionDefLike struct {
	LabelPT, LabelEN string
	Obrigatoria      bool
	ENStale          bool
	Chave            string
}

func TestEditorSalvaTextosDoForm(t *testing.T) {
	s := servidorAdmin(t)
	w := postAdmin(t, s, "/admin/forms/fotografia", url.Values{
		"titulo":    {"Vamos criar juntas?"},
		"subtitulo": {"Conta tudo."},
		"enviar":    {"Mandar"},
		"sucesso":   {"Recebido!"},
	})
	if w.Code != http.StatusSeeOther {
		t.Fatalf("status = %d, quero 303. Corpo: %s", w.Code, w.Body.String())
	}
	// O cache ja recarregou (aposEditar): o titulo novo esta no ar.
	foto, _ := s.forms.get("fotografia")
	if foto.Titulo.T("pt") != "Vamos criar juntas?" {
		t.Errorf("titulo nao salvou: %q", foto.Titulo.T("pt"))
	}
	if foto.Enviar.T("pt") != "Mandar" {
		t.Errorf("botao nao salvou: %q", foto.Enviar.T("pt"))
	}
}

func TestEditorTituloVazioRejeitado(t *testing.T) {
	s := servidorAdmin(t)
	w := postAdmin(t, s, "/admin/forms/fotografia", url.Values{
		"titulo": {""}, "enviar": {"Enviar"},
	})
	if w.Code != http.StatusUnprocessableEntity {
		t.Errorf("status = %d, quero 422 (titulo vazio)", w.Code)
	}
	foto, _ := s.forms.get("fotografia")
	if foto.Titulo.T("pt") == "" {
		t.Error("titulo foi apagado apesar da recusa")
	}
}

// TestEditorEditarPTMarcaStale: mudar o texto PT limpa o EN daquele campo e liga
// en_stale -- e o mecanismo que avisa o dev que falta retraduzir.
func TestEditorEditarPTMarcaStale(t *testing.T) {
	s := servidorAdmin(t)
	id := qid(t, s, "fotografia", "nome")

	w := postAdmin(t, s, "/admin/forms/fotografia/perguntas/"+itoaID(id), url.Values{
		"label":       {"Qual é o seu nome completo?"},
		"obrigatoria": {"on"},
	})
	if w.Code != http.StatusSeeOther {
		t.Fatalf("status = %d, quero 303. Corpo: %s", w.Code, w.Body.String())
	}

	q := questionDef(t, s, "fotografia", "nome")
	if q.LabelPT != "Qual é o seu nome completo?" {
		t.Errorf("label PT nao salvou: %q", q.LabelPT)
	}
	if q.LabelEN != "" {
		t.Errorf("label EN deveria ter sido limpo, esta %q", q.LabelEN)
	}
	if !q.ENStale {
		t.Error("en_stale deveria estar ligado apos editar o PT")
	}
}

// TestEditorToggleNaoMexeNoEN: mexer so no obrigatorio (texto PT igual) NAO pode
// derrubar a traducao EN nem marcar stale.
func TestEditorToggleNaoMexeNoEN(t *testing.T) {
	s := servidorAdmin(t)
	antes := questionDef(t, s, "fotografia", "nome") // LabelEN = "What's your name?"
	if antes.LabelEN == "" {
		t.Fatal("pre-condicao: 'nome' deveria ter EN no seed")
	}
	id := qid(t, s, "fotografia", "nome")

	// Reenvia o MESMO label PT, so desmarcando obrigatoria (campo omitido).
	w := postAdmin(t, s, "/admin/forms/fotografia/perguntas/"+itoaID(id), url.Values{
		"label": {antes.LabelPT},
	})
	if w.Code != http.StatusSeeOther {
		t.Fatalf("status = %d, quero 303", w.Code)
	}

	depois := questionDef(t, s, "fotografia", "nome")
	if depois.LabelEN != antes.LabelEN {
		t.Errorf("EN foi mexido sem o PT mudar: %q -> %q", antes.LabelEN, depois.LabelEN)
	}
	if depois.ENStale {
		t.Error("en_stale ligou sem o PT mudar")
	}
	if depois.Obrigatoria {
		t.Error("obrigatoria deveria ter sido desligada")
	}
}

// TestEditorAdicionaPerguntaComChaveOpaca: a pergunta nova nasce com chave
// gerada por nos (q_...), invisivel para a Isa -- e o que torna seguro renomear
// depois sem orfanar resposta.
func TestEditorAdicionaPerguntaComChaveOpaca(t *testing.T) {
	s := servidorAdmin(t)
	antes, _ := s.forms.get("fotografia")
	n := len(antes.Perguntas)

	w := postAdmin(t, s, "/admin/forms/fotografia/perguntas", url.Values{
		"label": {"Você tem algum pet?"},
		"tipo":  {"text"},
	})
	if w.Code != http.StatusSeeOther {
		t.Fatalf("status = %d, quero 303. Corpo: %s", w.Code, w.Body.String())
	}

	depois, _ := s.forms.get("fotografia")
	if len(depois.Perguntas) != n+1 {
		t.Fatalf("tinha %d perguntas, agora %d", n, len(depois.Perguntas))
	}
	nova := depois.Perguntas[len(depois.Perguntas)-1] // entrou no fim
	if nova.Label.T("pt") != "Você tem algum pet?" {
		t.Errorf("label da nova errado: %q", nova.Label.T("pt"))
	}
	if !strings.HasPrefix(nova.Chave, "q_") {
		t.Errorf("chave deveria ser opaca (q_...), e %q", nova.Chave)
	}
}

func TestEditorSelectSemOpcaoRejeitado(t *testing.T) {
	s := servidorAdmin(t)
	antes, _ := s.forms.get("fotografia")
	n := len(antes.Perguntas)

	w := postAdmin(t, s, "/admin/forms/fotografia/perguntas", url.Values{
		"label":  {"Escolha uma cor"},
		"tipo":   {"select"},
		"opcoes": {"   \n  \n"}, // so espaco/linhas vazias
	})
	if w.Code != http.StatusUnprocessableEntity {
		t.Errorf("status = %d, quero 422 (select sem opcao)", w.Code)
	}
	depois, _ := s.forms.get("fotografia")
	if len(depois.Perguntas) != n {
		t.Error("adicionou um select invalido")
	}
}

func TestEditorAdicionaSelectComOpcoes(t *testing.T) {
	s := servidorAdmin(t)
	w := postAdmin(t, s, "/admin/forms/fotografia/perguntas", url.Values{
		"label":  {"Período preferido"},
		"tipo":   {"select"},
		"opcoes": {"Manhã\nTarde\nNoite"},
	})
	if w.Code != http.StatusSeeOther {
		t.Fatalf("status = %d, quero 303. Corpo: %s", w.Code, w.Body.String())
	}
	foto, _ := s.forms.get("fotografia")
	nova := foto.Perguntas[len(foto.Perguntas)-1]
	if nova.Tipo != campoSelect || len(nova.Opcoes) != 3 {
		t.Fatalf("select nao entrou certo: tipo=%q opcoes=%d", nova.Tipo, len(nova.Opcoes))
	}
	if nova.Opcoes[1].T("pt") != "Tarde" {
		t.Errorf("opcao errada: %q", nova.Opcoes[1].T("pt"))
	}
}

func TestEditorMovePergunta(t *testing.T) {
	s := servidorAdmin(t)
	foto, _ := s.forms.get("fotografia")
	if foto.Perguntas[0].Chave != "nome" || foto.Perguntas[1].Chave != "contato" {
		t.Fatalf("pre-condicao: esperava nome, contato no topo; veio %q, %q",
			foto.Perguntas[0].Chave, foto.Perguntas[1].Chave)
	}
	id := qid(t, s, "fotografia", "contato")

	w := postAdmin(t, s, "/admin/forms/fotografia/perguntas/"+itoaID(id)+"/mover", url.Values{
		"dir": {"cima"},
	})
	if w.Code != http.StatusSeeOther {
		t.Fatalf("status = %d, quero 303", w.Code)
	}
	foto, _ = s.forms.get("fotografia")
	if foto.Perguntas[0].Chave != "contato" {
		t.Errorf("mover pra cima nao funcionou; topo e %q", foto.Perguntas[0].Chave)
	}
}

// TestEditorCSRF: POST sem Origin/Referer e barrado; com Origin certo passa.
func TestEditorCSRF(t *testing.T) {
	s := servidorAdmin(t)

	// Sem Origin nem Referer: recusa.
	r := httptest.NewRequest(http.MethodPost, "/admin/forms/fotografia", strings.NewReader("titulo=X&enviar=Y"))
	r.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	r.SetBasicAuth("isa", "segredo")
	w := httptest.NewRecorder()
	s.Handler().ServeHTTP(w, r)
	if w.Code != http.StatusForbidden {
		t.Errorf("POST sem Origin: status = %d, quero 403", w.Code)
	}

	// Origin de outro host: recusa.
	r = httptest.NewRequest(http.MethodPost, "/admin/forms/fotografia", strings.NewReader("titulo=X&enviar=Y"))
	r.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	r.Header.Set("Origin", "https://site-malicioso.com")
	r.SetBasicAuth("isa", "segredo")
	w = httptest.NewRecorder()
	s.Handler().ServeHTTP(w, r)
	if w.Code != http.StatusForbidden {
		t.Errorf("POST de outra origem: status = %d, quero 403", w.Code)
	}
}

// TestEditorExigeAuth: o editor esta atras do mesmo basic-auth do painel.
func TestEditorExigeAuth(t *testing.T) {
	s := servidorAdmin(t)
	w := httptest.NewRecorder()
	s.Handler().ServeHTTP(w, httptest.NewRequest(http.MethodGet, "/admin/forms", nil))
	if w.Code != http.StatusUnauthorized {
		t.Errorf("GET /admin/forms sem auth: status = %d, quero 401", w.Code)
	}
}

// TestEditorEdicaoApareceNoSitePublico: a prova de ponta a ponta -- editar no
// painel muda o formulario que o visitante ve, sem redeploy.
func TestEditorEdicaoApareceNoSitePublico(t *testing.T) {
	s := servidorAdmin(t)
	postAdmin(t, s, "/admin/forms/fotografia", url.Values{
		"titulo": {"Bora marcar seu ensaio ✨"},
		"enviar": {"Enviar briefing"},
	})

	w := httptest.NewRecorder()
	s.Handler().ServeHTTP(w, httptest.NewRequest(http.MethodGet, "/briefing", nil))
	if !strings.Contains(w.Body.String(), "Bora marcar seu ensaio") {
		t.Error("a edicao no painel nao apareceu na pagina publica /briefing")
	}
}

func itoaID(id int64) string {
	return strconv.FormatInt(id, 10)
}
