package web

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

// A Fase A do editor (docs §8b) move os formularios de structs Go para o SQLite,
// carregados em RAM. Estes testes travam o contrato: o seed reproduz fielmente o
// conteudo de antes, e o cache recarrega uma edicao feita direto no banco -- que
// e exatamente o que a UI da Fase B vai fazer.

// TestSeedReproduzOsForms: um banco novo nasce com os dois formularios, na ordem
// certa, com as opcoes de select intactas apos o round-trip por JSON.
func TestSeedReproduzOsForms(t *testing.T) {
	s := novoServidor(t)

	foto, ok := s.forms.get("fotografia")
	if !ok {
		t.Fatal("form 'fotografia' nao foi semeado")
	}
	ugc, ok := s.forms.get("ugc")
	if !ok {
		t.Fatal("form 'ugc' nao foi semeado")
	}

	// Titulo bilingue sobreviveu.
	if foto.Titulo.T("pt") != "Vamos combinar seu ensaio" {
		t.Errorf("titulo PT errado: %q", foto.Titulo.T("pt"))
	}
	if ugc.Titulo.T("en") != "Tell me about the project" {
		t.Errorf("titulo EN errado: %q", ugc.Titulo.T("en"))
	}

	// A primeira pergunta e 'nome', obrigatoria.
	if foto.Perguntas[0].Chave != "nome" || !foto.Perguntas[0].Obrigatoria {
		t.Errorf("primeira pergunta inesperada: %+v", foto.Perguntas[0])
	}

	// O select 'tipo' manteve as 5 opcoes, bilingues.
	tipo := acharPergunta(t, foto, "tipo")
	if tipo.Tipo != campoSelect {
		t.Errorf("'tipo' deveria ser select, e %q", tipo.Tipo)
	}
	if len(tipo.Opcoes) != 5 {
		t.Fatalf("'tipo' tem %d opcoes, quero 5", len(tipo.Opcoes))
	}
	if tipo.Opcoes[0].T("pt") != "Ensaio individual" || tipo.Opcoes[0].T("en") != "Individual shoot" {
		t.Errorf("primeira opcao de 'tipo' nao sobreviveu ao JSON: %+v", tipo.Opcoes[0])
	}

	// O campo de contato manteve o tipo proprio (nao virou texto puro).
	if acharPergunta(t, foto, "contato").Tipo != campoContato {
		t.Error("'contato' perdeu o tipo campoContato no seed")
	}
}

// TestSeedIdempotente: subir o servidor duas vezes sobre o MESMO banco nao
// duplica formularios -- o seed so roda com a tabela vazia.
func TestSeedIdempotente(t *testing.T) {
	s := novoServidor(t)
	// Reusa o mesmo store num cache novo: simula um segundo boot.
	c2, err := novoFormsCache(s.store)
	if err != nil {
		t.Fatal(err)
	}
	if _, ok := c2.get("fotografia"); !ok {
		t.Fatal("segundo boot nao enxerga o form")
	}
	// Conta forms no banco: tem que ser 2, nao 4.
	var n int
	if err := s.store.DB().QueryRow("SELECT COUNT(*) FROM forms").Scan(&n); err != nil {
		t.Fatal(err)
	}
	if n != 2 {
		t.Errorf("seed rodou de novo: %d forms no banco, quero 2", n)
	}
}

// TestReloadPegaEdicaoNoBanco e o coracao da Fase A: editar direto no banco e
// recarregar reflete no formulario servido -- prova que o caminho editavel
// funciona, antes mesmo de existir UI.
func TestReloadPegaEdicaoNoBanco(t *testing.T) {
	s := novoServidor(t)

	_, err := s.store.DB().Exec(
		`UPDATE form_questions SET label_pt = 'Qual o seu nome, lindeza?' WHERE form_slug='fotografia' AND chave='nome'`)
	if err != nil {
		t.Fatal(err)
	}

	// Antes do reload, o cache ainda tem o valor antigo.
	if foto, _ := s.forms.get("fotografia"); foto.Perguntas[0].Label.T("pt") == "Qual o seu nome, lindeza?" {
		t.Fatal("o cache mudou sem reload -- deveria ler da RAM, nao do banco")
	}

	if err := s.forms.reload(); err != nil {
		t.Fatal(err)
	}
	foto, _ := s.forms.get("fotografia")
	if foto.Perguntas[0].Label.T("pt") != "Qual o seu nome, lindeza?" {
		t.Errorf("reload nao pegou a edicao: %q", foto.Perguntas[0].Label.T("pt"))
	}
}

// TestPerguntaOcultaSumeDoFormMasNaoDoPainel protege as respostas antigas: uma
// pergunta escondida (oculta=1) desaparece do formulario que o visitante ve, mas
// continua resolvendo rotulo no painel -- senao esconder uma pergunta apagaria
// da tela as respostas ja recebidas para ela.
func TestPerguntaOcultaSumeDoFormMasNaoDoPainel(t *testing.T) {
	s := novoServidor(t)

	if _, err := s.store.DB().Exec(
		`UPDATE form_questions SET oculta=1 WHERE form_slug='fotografia' AND chave='lanche'`); err != nil {
		t.Fatal(err)
	}
	if err := s.forms.reload(); err != nil {
		t.Fatal(err)
	}

	// Some do formulario publico.
	foto, _ := s.forms.get("fotografia")
	for _, p := range foto.Perguntas {
		if p.Chave == "lanche" {
			t.Error("pergunta oculta ainda aparece no formulario publico")
		}
	}
	// Mas o painel ainda a conhece (para rotular respostas antigas).
	perguntas, _ := s.forms.perguntasDe("fotografia")
	achou := false
	for _, p := range perguntas {
		if p.Chave == "lanche" {
			achou = true
		}
	}
	if !achou {
		t.Error("pergunta oculta sumiu tambem do painel -- respostas antigas ficariam sem rotulo")
	}
}

// TestFormInvisivelSaiDasRotas: um form desligado (visivel=0) nao e servido
// (get devolve false), mas continua conhecido (existe true) para o painel
// rotular respostas antigas.
func TestFormInvisivelSaiDasRotas(t *testing.T) {
	s := novoServidor(t)

	if _, err := s.store.DB().Exec(`UPDATE forms SET visivel=0 WHERE slug='ugc'`); err != nil {
		t.Fatal(err)
	}
	if err := s.forms.reload(); err != nil {
		t.Fatal(err)
	}

	if _, ok := s.forms.get("ugc"); ok {
		t.Error("form invisivel ainda e servido por get()")
	}
	if !s.forms.existe("ugc") {
		t.Error("form invisivel deveria continuar 'conhecido' para o painel")
	}

	// A pagina do midia kit ainda renderiza, so sem a secao do briefing.
	w := httptest.NewRecorder()
	s.Handler().ServeHTTP(w, httptest.NewRequest(http.MethodGet, "/midiakit", nil))
	if w.Code != 200 {
		t.Fatalf("/midiakit quebrou com o form de UGC desligado: %d", w.Code)
	}
	if strings.Contains(w.Body.String(), `id="briefing"`) {
		t.Error("a secao do briefing deveria sumir com o form de UGC desligado")
	}
}

func acharPergunta(t *testing.T, f *formBriefing, chave string) pergunta {
	t.Helper()
	for _, p := range f.Perguntas {
		if p.Chave == chave {
			return p
		}
	}
	t.Fatalf("pergunta %q nao encontrada em %q", chave, f.Slug)
	return pergunta{}
}
