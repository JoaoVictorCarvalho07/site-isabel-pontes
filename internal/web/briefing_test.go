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

// Respostas validas minimas do briefing de fotografia (so os obrigatorios).
func formValido() url.Values {
	return url.Values{
		"nome":      {"Maria Souza"},
		"contato":   {"41999998888"},
		"tipo":      {"Ensaio individual"},
		"motivacao": {"Quero registrar meus 30 anos."},
	}
}

func postBriefing(t *testing.T, h http.Handler, caminho string, dados url.Values, htmx bool) *httptest.ResponseRecorder {
	t.Helper()
	r := httptest.NewRequest(http.MethodPost, caminho, strings.NewReader(dados.Encode()))
	r.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	if htmx {
		r.Header.Set("HX-Request", "true")
	}
	// IP unico por chamada: sem isto o rate limit de 30s derruba o segundo
	// POST de cada teste, e o motivo da falha ficaria completamente obscuro.
	r.RemoteAddr = randomIP(t)
	w := httptest.NewRecorder()
	h.ServeHTTP(w, r)
	return w
}

var contadorIP int

func randomIP(t *testing.T) string {
	t.Helper()
	contadorIP++
	return "10.0.0." + strconv.Itoa(contadorIP) + ":1234"
}

// TestBriefingSalvaResposta e o teste central: um POST valido tem que ATERRISSAR
// NO BANCO. O e-mail pode falhar, o painel pode estar fora -- perder a resposta
// e a unica falha inaceitavel deste recurso.
func TestBriefingSalvaResposta(t *testing.T) {
	s := novoServidor(t)
	w := postBriefing(t, s.Handler(), "/briefing/fotografia", formValido(), true)

	if w.Code != http.StatusOK {
		t.Fatalf("status = %d, quero 200. Corpo: %s", w.Code, w.Body.String())
	}
	if !strings.Contains(w.Body.String(), "Recebi seu briefing") {
		t.Errorf("faltou a confirmacao no HTML devolvido:\n%s", w.Body.String())
	}

	subs, err := s.store.ListarSubmissions(context.Background(), "", 10)
	if err != nil {
		t.Fatalf("ListarSubmissions: %v", err)
	}
	if len(subs) != 1 {
		t.Fatalf("gravou %d respostas, quero 1", len(subs))
	}
	if subs[0].Form != "fotografia" {
		t.Errorf("form = %q, quero \"fotografia\"", subs[0].Form)
	}
	if subs[0].Respostas["motivacao"] != "Quero registrar meus 30 anos." {
		t.Errorf("motivacao nao bateu: %q", subs[0].Respostas["motivacao"])
	}
}

// TestBriefingValidacao: campo obrigatorio faltando nao pode gravar nada, e o
// visitante precisa receber o formulario DE VOLTA com o que ja tinha digitado --
// mandar alguem reescrever uma resposta longa e o jeito mais rapido de perder o
// preenchimento.
func TestBriefingValidacao(t *testing.T) {
	s := novoServidor(t)
	dados := formValido()
	dados.Del("motivacao") // obrigatorio

	w := postBriefing(t, s.Handler(), "/briefing/fotografia", dados, true)

	// 200 mesmo com erro: o HTMX so troca o DOM em 2xx (ver comentario em
	// briefing.go). Um 422 aqui faria o botao "enviar" nao fazer nada visivel.
	if w.Code != http.StatusOK {
		t.Errorf("status = %d, quero 200 (fragmento HTMX)", w.Code)
	}
	corpo := w.Body.String()
	if !strings.Contains(corpo, "Esse campo é obrigatório") {
		t.Errorf("faltou a mensagem de erro:\n%s", corpo)
	}
	if !strings.Contains(corpo, "Maria Souza") {
		t.Error("o formulario voltou sem os valores ja preenchidos")
	}
	if subs, _ := s.store.ListarSubmissions(context.Background(), "", 10); len(subs) != 0 {
		t.Errorf("gravou %d respostas apesar do erro de validacao", len(subs))
	}
}

// TestBriefingHoneypot: bot preenche o campo escondido -> ve "obrigada", banco
// continua vazio. Responder sucesso e proposital (ver briefing.go).
func TestBriefingHoneypot(t *testing.T) {
	s := novoServidor(t)
	dados := formValido()
	dados.Set(campoHoneypot, "http://spam.example")

	w := postBriefing(t, s.Handler(), "/briefing/fotografia", dados, true)

	if w.Code != http.StatusOK {
		t.Errorf("status = %d, quero 200 (o bot precisa achar que deu certo)", w.Code)
	}
	if subs, _ := s.store.ListarSubmissions(context.Background(), "", 10); len(subs) != 0 {
		t.Errorf("o honeypot deixou passar: %d respostas gravadas", len(subs))
	}
}

// TestBriefingSelectRejeitaValorForaDaLista: um POST cru nao pode inventar
// categoria. Sem isto a Isa leria "tipo: <script>" como se fosse uma escolha.
func TestBriefingSelectRejeitaValorForaDaLista(t *testing.T) {
	s := novoServidor(t)
	dados := formValido()
	dados.Set("tipo", "Ensaio de gato de botas")

	w := postBriefing(t, s.Handler(), "/briefing/fotografia", dados, true)

	if !strings.Contains(w.Body.String(), "Escolha uma das opções") {
		t.Errorf("aceitou opcao inventada:\n%s", w.Body.String())
	}
	if subs, _ := s.store.ListarSubmissions(context.Background(), "", 10); len(subs) != 0 {
		t.Error("gravou uma opcao que nao existe no formulario")
	}
}

// TestBriefingRateLimit: dois envios seguidos do MESMO IP -- o segundo e barrado.
func TestBriefingRateLimit(t *testing.T) {
	s := novoServidor(t)
	h := s.Handler()

	enviar := func() *httptest.ResponseRecorder {
		r := httptest.NewRequest(http.MethodPost, "/briefing/fotografia",
			strings.NewReader(formValido().Encode()))
		r.Header.Set("Content-Type", "application/x-www-form-urlencoded")
		r.Header.Set("HX-Request", "true")
		r.RemoteAddr = "203.0.113.7:5555" // fixo: e o ponto do teste
		w := httptest.NewRecorder()
		h.ServeHTTP(w, r)
		return w
	}

	if corpo := enviar().Body.String(); !strings.Contains(corpo, "Recebi seu briefing") {
		t.Fatalf("o primeiro envio deveria passar:\n%s", corpo)
	}
	if corpo := enviar().Body.String(); !strings.Contains(corpo, "Espera um minutinho") {
		t.Errorf("o segundo envio deveria ser barrado:\n%s", corpo)
	}
	if subs, _ := s.store.ListarSubmissions(context.Background(), "", 10); len(subs) != 1 {
		t.Errorf("gravou %d respostas, quero 1 (a segunda foi barrada)", len(subs))
	}
}

// TestBriefingSemJS: sem o header HX-Request o POST devolve a PAGINA inteira,
// com status semantico. E a prova de que o form funciona com JavaScript
// desligado -- a promessa do progressive enhancement do projeto.
func TestBriefingSemJS(t *testing.T) {
	s := novoServidor(t)

	w := postBriefing(t, s.Handler(), "/briefing/fotografia", formValido(), false)
	corpo := w.Body.String()
	if w.Code != http.StatusOK {
		t.Errorf("status = %d, quero 200", w.Code)
	}
	if !strings.Contains(corpo, "<!doctype html") && !strings.Contains(corpo, "<!DOCTYPE html") {
		t.Error("sem HTMX o POST deveria devolver a pagina completa, nao um fragmento")
	}

	// Erro de validacao, sem JS: aqui o status HTTP semantico VALE (nao ha swap
	// de DOM para proteger).
	dados := formValido()
	dados.Del("nome")
	w = postBriefing(t, s.Handler(), "/briefing/fotografia", dados, false)
	if w.Code != http.StatusUnprocessableEntity {
		t.Errorf("status = %d, quero 422 no caminho sem JS", w.Code)
	}
}

// TestBriefingUGCNaOnePage: o formulario de UGC tem que estar EMBUTIDO no mídia
// kit (decisao "misto"), apontando para o proprio slug -- e nao para o de
// fotografia, que gravaria a resposta no formulario errado.
func TestBriefingUGCNaOnePage(t *testing.T) {
	s := novoServidor(t)
	w := httptest.NewRecorder()
	s.Handler().ServeHTTP(w, httptest.NewRequest(http.MethodGet, "/midiakit", nil))

	corpo := w.Body.String()
	for _, m := range []string{
		`action="/briefing/ugc"`,
		"Por que você me escolheu como creator?",
		`name="marca"`,
	} {
		if !strings.Contains(corpo, m) {
			t.Errorf("/midiakit nao contem %q", m)
		}
	}
	// A pergunta do lanche e do briefing de FOTOGRAFIA; nao pode vazar para o
	// formulario que a marca ve.
	if strings.Contains(corpo, "O que você gosta de comer?") {
		t.Error("pergunta de fotografia vazou para o formulario de UGC")
	}
}

// TestBriefingSlugDesconhecido: POST num formulario que nao existe e 404, nao
// um registro com `form` invalido.
func TestBriefingSlugDesconhecido(t *testing.T) {
	s := novoServidor(t)
	w := postBriefing(t, s.Handler(), "/briefing/inventado", formValido(), true)
	if w.Code != http.StatusNotFound {
		t.Errorf("status = %d, quero 404", w.Code)
	}
}

// TestBriefingEN: o espelho em ingles posta em /en/briefing/... e responde em EN.
func TestBriefingEN(t *testing.T) {
	s := novoServidor(t)
	h := s.Handler()

	w := httptest.NewRecorder()
	h.ServeHTTP(w, httptest.NewRequest(http.MethodGet, "/en/briefing", nil))
	if !strings.Contains(w.Body.String(), `action="/en/briefing/fotografia"`) {
		t.Error("/en/briefing deveria postar para /en/briefing/fotografia")
	}

	w = postBriefing(t, h, "/en/briefing/fotografia", formValido(), true)
	if !strings.Contains(w.Body.String(), "Got your briefing") {
		t.Errorf("a confirmacao deveria vir em EN:\n%s", w.Body.String())
	}
	subs, _ := s.store.ListarSubmissions(context.Background(), "", 10)
	if len(subs) != 1 || subs[0].Lang != "en" {
		t.Errorf("o idioma nao foi gravado junto com a resposta: %+v", subs)
	}
}

// TestAdminExigeSenha protege a caixa de entrada da Isa. Os dois casos importam:
// sem credencial configurada o painel NAO PODE existir (senao um deploy que
// esqueceu a variavel publica todos os briefings), e com credencial ele exige.
func TestAdminExigeSenha(t *testing.T) {
	t.Run("sem credencial configurada -> 404", func(t *testing.T) {
		s := novoServidor(t) // config sem AdminUser/AdminSenha
		w := httptest.NewRecorder()
		s.Handler().ServeHTTP(w, httptest.NewRequest(http.MethodGet, "/admin", nil))
		if w.Code != http.StatusNotFound {
			t.Errorf("status = %d, quero 404 (painel desligado)", w.Code)
		}
	})

	t.Run("com credencial -> exige e aceita", func(t *testing.T) {
		s := novoServidor(t)
		s.cfg.AdminUser, s.cfg.AdminSenha = "isa", "segredo"
		h := s.Handler()

		w := httptest.NewRecorder()
		h.ServeHTTP(w, httptest.NewRequest(http.MethodGet, "/admin", nil))
		if w.Code != http.StatusUnauthorized {
			t.Errorf("sem auth: status = %d, quero 401", w.Code)
		}

		w = httptest.NewRecorder()
		r := httptest.NewRequest(http.MethodGet, "/admin", nil)
		r.SetBasicAuth("isa", "errada")
		h.ServeHTTP(w, r)
		if w.Code != http.StatusUnauthorized {
			t.Errorf("senha errada: status = %d, quero 401", w.Code)
		}

		w = httptest.NewRecorder()
		r = httptest.NewRequest(http.MethodGet, "/admin", nil)
		r.SetBasicAuth("isa", "segredo")
		h.ServeHTTP(w, r)
		if w.Code != http.StatusOK {
			t.Errorf("credencial certa: status = %d, quero 200. Corpo: %s", w.Code, w.Body.String())
		}
	})
}

// TestAdminMostraResposta: o que foi enviado precisa aparecer no painel, casado
// com o TEXTO da pergunta -- nao com a chave crua.
func TestAdminMostraResposta(t *testing.T) {
	s := novoServidor(t)
	s.cfg.AdminUser, s.cfg.AdminSenha = "isa", "segredo"
	h := s.Handler()

	postBriefing(t, h, "/briefing/fotografia", formValido(), true)

	w := httptest.NewRecorder()
	r := httptest.NewRequest(http.MethodGet, "/admin", nil)
	r.SetBasicAuth("isa", "segredo")
	h.ServeHTTP(w, r)

	corpo := w.Body.String()
	for _, m := range []string{
		"Maria Souza",
		"O que te motivou a querer fazer esse ensaio?", // label, nao a chave
		"Quero registrar meus 30 anos.",
	} {
		if !strings.Contains(corpo, m) {
			t.Errorf("/admin nao mostra %q", m)
		}
	}
}

// TestAdminForaDoRobots: painel nao entra no indice nem no sitemap.
func TestAdminForaDoRobots(t *testing.T) {
	s := novoServidor(t)
	h := s.Handler()

	w := httptest.NewRecorder()
	h.ServeHTTP(w, httptest.NewRequest(http.MethodGet, "/robots.txt", nil))
	if !strings.Contains(w.Body.String(), "Disallow: /admin") {
		t.Error("robots.txt deveria bloquear /admin")
	}

	w = httptest.NewRecorder()
	h.ServeHTTP(w, httptest.NewRequest(http.MethodGet, "/sitemap.xml", nil))
	if strings.Contains(w.Body.String(), "/admin") {
		t.Error("/admin nao pode estar no sitemap")
	}
}
