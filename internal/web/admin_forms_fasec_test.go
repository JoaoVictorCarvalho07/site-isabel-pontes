package web

import (
	"context"
	"net/http"
	"net/http/httptest"
	"net/url"
	"strings"
	"testing"
)

// Testes das operacoes destrutivas (Fase C, docs §8b). O foco e a politica que
// protege os dados: apagar so sem respostas; com respostas, esconder. Trocar
// tipo e reversivel e sempre permitido.

// respondeuLanche envia um briefing publico valido que responde a pergunta
// "lanche", para as perguntas passarem a ter resposta.
func respondeuLanche(t *testing.T, s *Server) {
	t.Helper()
	dados := url.Values{
		"nome":      {"Cliente Teste"},
		"contato":   {"41999998888"},
		"tipo":      {"Ensaio individual"},
		"motivacao": {"Quero muito."},
		"lanche":    {"Brigadeiro"},
	}
	w := postBriefing(t, s.Handler(), "/briefing/fotografia", dados, true)
	if !strings.Contains(w.Body.String(), "Recebi seu briefing") {
		t.Fatalf("nao consegui semear uma resposta: %s", w.Body.String())
	}
}

// TestApagarComRespostaRecusado e a garantia central da Fase C: uma pergunta com
// resposta NAO pode ser apagada (orfanaria o briefing). O servidor recusa mesmo
// que a UI deixasse -- nunca confia so na tela.
func TestApagarComRespostaRecusado(t *testing.T) {
	s := servidorAdmin(t)
	respondeuLanche(t, s)
	id := qid(t, s, "fotografia", "lanche")

	w := postAdmin(t, s, "/admin/forms/fotografia/perguntas/"+itoaID(id)+"/apagar", url.Values{})
	if w.Code != http.StatusUnprocessableEntity {
		t.Fatalf("status = %d, quero 422 (apagar com resposta)", w.Code)
	}
	// A pergunta continua existindo.
	if _, ok := s.forms.perguntasDe("fotografia"); !ok {
		t.Fatal("form sumiu")
	}
	perguntas, _ := s.forms.perguntasDe("fotografia")
	achou := false
	for _, p := range perguntas {
		if p.Chave == "lanche" {
			achou = true
		}
	}
	if !achou {
		t.Error("a pergunta 'lanche' foi apagada apesar de ter resposta")
	}
}

// TestApagarSemRespostaFunciona: uma pergunta recem-criada, sem resposta, pode
// ser apagada de vez.
func TestApagarSemRespostaFunciona(t *testing.T) {
	s := servidorAdmin(t)
	// adiciona uma pergunta nova (sem respostas)
	postAdmin(t, s, "/admin/forms/fotografia/perguntas", url.Values{
		"label": {"Pergunta descartável"}, "tipo": {"text"},
	})
	foto, _ := s.forms.get("fotografia")
	nova := foto.Perguntas[len(foto.Perguntas)-1]
	id := qid(t, s, "fotografia", nova.Chave)

	w := postAdmin(t, s, "/admin/forms/fotografia/perguntas/"+itoaID(id)+"/apagar", url.Values{})
	if w.Code != http.StatusSeeOther {
		t.Fatalf("status = %d, quero 303", w.Code)
	}
	foto, _ = s.forms.get("fotografia")
	for _, p := range foto.Perguntas {
		if p.Chave == nova.Chave {
			t.Error("a pergunta sem resposta nao foi apagada")
		}
	}
}

// TestEsconderPergunta: some do formulario publico, mas continua no editor e
// rotulando respostas no painel. Reversivel.
func TestEsconderPergunta(t *testing.T) {
	s := servidorAdmin(t)
	respondeuLanche(t, s) // "lanche" agora tem resposta
	id := qid(t, s, "fotografia", "lanche")

	// Esconde.
	w := postAdmin(t, s, "/admin/forms/fotografia/perguntas/"+itoaID(id)+"/ocultar", url.Values{"oculta": {"1"}})
	if w.Code != http.StatusSeeOther {
		t.Fatalf("status = %d, quero 303", w.Code)
	}
	// Some do formulario publico.
	foto, _ := s.forms.get("fotografia")
	for _, p := range foto.Perguntas {
		if p.Chave == "lanche" {
			t.Error("pergunta escondida ainda aparece no formulario publico")
		}
	}
	// Mas o painel de respostas ainda a rotula.
	perguntas, _ := s.forms.perguntasDe("fotografia")
	achou := false
	for _, p := range perguntas {
		if p.Chave == "lanche" {
			achou = true
		}
	}
	if !achou {
		t.Error("pergunta escondida sumiu do painel -- respostas ficariam sem rotulo")
	}

	// Mostra de novo.
	w = postAdmin(t, s, "/admin/forms/fotografia/perguntas/"+itoaID(id)+"/ocultar", url.Values{"oculta": {"0"}})
	if w.Code != http.StatusSeeOther {
		t.Fatalf("mostrar: status = %d, quero 303", w.Code)
	}
	foto, _ = s.forms.get("fotografia")
	achou = false
	for _, p := range foto.Perguntas {
		if p.Chave == "lanche" {
			achou = true
		}
	}
	if !achou {
		t.Error("pergunta nao voltou ao formulario apos mostrar")
	}
}

// TestTrocarTipo: muda o tipo e marca en_stale. As respostas antigas continuam.
func TestTrocarTipo(t *testing.T) {
	s := servidorAdmin(t)
	respondeuLanche(t, s)
	id := qid(t, s, "fotografia", "lanche") // era campoTexto

	w := postAdmin(t, s, "/admin/forms/fotografia/perguntas/"+itoaID(id)+"/tipo", url.Values{
		"tipo": {"textarea"},
	})
	if w.Code != http.StatusSeeOther {
		t.Fatalf("status = %d, quero 303. Corpo: %s", w.Code, w.Body.String())
	}
	q := questionDef(t, s, "fotografia", "lanche")
	if q.Chave != "lanche" {
		t.Fatal("a chave nao pode mudar ao trocar tipo")
	}
	foto, _ := s.forms.get("fotografia")
	for _, p := range foto.Perguntas {
		if p.Chave == "lanche" && p.Tipo != campoTextarea {
			t.Errorf("tipo nao trocou: %q", p.Tipo)
		}
	}
	if !q.ENStale {
		t.Error("trocar tipo deveria marcar en_stale")
	}

	// A resposta antiga sobreviveu (o briefing continua no banco).
	subs, _ := s.store.ListarSubmissions(context.Background(), "fotografia", 10)
	if len(subs) != 1 || subs[0].Respostas["lanche"] != "Brigadeiro" {
		t.Error("a resposta antiga se perdeu ao trocar o tipo")
	}
}

// TestTrocarTipoParaSelectSemOpcao: virar lista sem opcoes e recusado.
func TestTrocarTipoParaSelectSemOpcao(t *testing.T) {
	s := servidorAdmin(t)
	id := qid(t, s, "fotografia", "lanche")
	w := postAdmin(t, s, "/admin/forms/fotografia/perguntas/"+itoaID(id)+"/tipo", url.Values{
		"tipo": {"select"}, "opcoes": {"  \n "},
	})
	if w.Code != http.StatusUnprocessableEntity {
		t.Errorf("status = %d, quero 422 (select sem opcao)", w.Code)
	}
}

// TestEditorMostraContagem: a contagem de respostas aparece na tela de edicao,
// para a Isa decidir informada antes de apagar/trocar tipo.
func TestEditorMostraContagem(t *testing.T) {
	s := servidorAdmin(t)
	respondeuLanche(t, s)

	w := httptest.NewRecorder()
	r := httptest.NewRequest(http.MethodGet, "/admin/forms/fotografia", nil)
	r.SetBasicAuth("isa", "segredo")
	s.Handler().ServeHTTP(w, r)

	if !strings.Contains(w.Body.String(), "resposta") {
		t.Error("a tela de edicao nao mostra a contagem de respostas")
	}
}
