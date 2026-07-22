package web

import (
	"context"
	"io/fs"
	"net/http"
	"net/http/httptest"
	"path/filepath"
	"strings"
	"testing"

	site "github.com/JoaoVictorCarvalho07/site-da-bebel-htmx"
	"github.com/JoaoVictorCarvalho07/site-da-bebel-htmx/internal/config"
	"github.com/JoaoVictorCarvalho07/site-da-bebel-htmx/internal/store"
)

func novoServidor(t *testing.T) *Server {
	t.Helper()
	// snapshot vazio de um banco temporario: basta para os testes de layout/SEO;
	// os testes da galeria em si vivem no pacote store e num teste dedicado.
	st, err := store.Open(filepath.Join(t.TempDir(), "t.db"))
	if err != nil {
		t.Fatalf("store.Open: %v", err)
	}
	t.Cleanup(func() { st.Close() })
	snap, err := st.LoadSnapshot(context.Background())
	if err != nil {
		t.Fatalf("LoadSnapshot: %v", err)
	}
	s, err := New(&config.Config{BaseURL: "https://exemplo.test"}, site.Assets, snap)
	if err != nil {
		t.Fatalf("New: %v", err)
	}
	return s
}

// TestTodosOsTemplatesExecutam e a defesa contra o unico buraco real de trocar
// React por html/template: renomear um campo do struct e esquecer o template
// nao da erro de compilacao -- da pagina quebrada em producao.
//
// Executa TODA pagina com dado preenchido. Se um template referenciar campo que
// nao existe no pageData, ExecuteTemplate falha e o teste pega aqui, nao o
// visitante. Toda pagina nova entra nesta lista.
func TestTodosOsTemplatesExecutam(t *testing.T) {
	s := novoServidor(t)

	paginas, err := fs.Glob(site.Assets, "templates/pages/*.gohtml")
	if err != nil {
		t.Fatal(err)
	}
	if len(paginas) == 0 {
		t.Fatal("nenhum template encontrado -- o //go:embed pegou?")
	}

	// Paginas com ViewModel proprio (nao o pageData generico) sao testadas pelos
	// handlers reais em TestPaginasRenderizam / TestGaleriaPaginaEFragmento. Aqui
	// fica so o que renderiza com pageData puro (ex.: 404).
	viewProprio := map[string]bool{"home": true, "sobre": true, "contato": true, "galeria": true, "midiakit": true}

	for _, p := range paginas {
		nome := strings.TrimSuffix(strings.TrimPrefix(p, "templates/pages/"), ".gohtml")
		if viewProprio[nome] {
			continue
		}
		t.Run(nome, func(t *testing.T) {
			w := httptest.NewRecorder()
			r := httptest.NewRequest(http.MethodGet, "/", nil)
			// Dado preenchido de proposito: campo vazio esconde erro de tipo.
			s.render.Page(w, r, nome, http.StatusOK, pageData{
				Titulo:    "titulo",
				Descricao: "descricao",
				Canonical: "https://exemplo.test/x",
				BodyClass: "classe",
				Ano:       2026,
			})
			if w.Code != http.StatusOK {
				t.Fatalf("status = %d, quero 200. Corpo: %s", w.Code, w.Body.String())
			}
			if !strings.Contains(w.Body.String(), "</html>") {
				t.Errorf("HTML truncado -- o template falhou no meio:\n%s", w.Body.String())
			}
		})
	}
}

// TestI18nEN protege o espelho em ingles: /en serve HTML em EN, com hreflang para
// as duas versoes, e o conteudo estruturado do midia kit resolve para o ingles --
// enquanto a raiz (PT) continua intacta.
func TestI18nEN(t *testing.T) {
	s := novoServidor(t)
	h := s.Handler()

	w := httptest.NewRecorder()
	h.ServeHTTP(w, httptest.NewRequest(http.MethodGet, "/en", nil))
	if w.Code != http.StatusOK {
		t.Fatalf("/en status %d, quero 200", w.Code)
	}
	corpo := w.Body.String()
	for _, m := range []string{
		`<html lang="en">`, "Welcome to my universe",
		`hreflang="pt-BR"`, `hreflang="en"`, `hreflang="x-default"`,
		">Gallery<", // nav traduzida
	} {
		if !strings.Contains(corpo, m) {
			t.Errorf("/en nao contem %q", m)
		}
	}

	w = httptest.NewRecorder()
	h.ServeHTTP(w, httptest.NewRequest(http.MethodGet, "/en/midiakit", nil))
	mk := w.Body.String()
	for _, m := range []string{"UGC Investment", "Most popular", "Beauty", "/en/midiakit/videos"} {
		if !strings.Contains(mk, m) {
			t.Errorf("/en/midiakit nao contem %q", m)
		}
	}

	// regressao: a raiz continua em PT
	w = httptest.NewRecorder()
	h.ServeHTTP(w, httptest.NewRequest(http.MethodGet, "/", nil))
	if !strings.Contains(w.Body.String(), `<html lang="pt-br">`) {
		t.Error("/ deveria continuar em pt-br")
	}
}

// TestSeletorIdiomaERelativo protege contra a regressao que quebrou o site em
// producao: o seletor PT/EN usava HrefPT/HrefEN, que carregam o BASE_URL. Com o
// BASE_URL apontando para outro host que nao o de acesso (o caso classico: o
// binario compilado com localhost e servido por IP), clicar em "EN" mandava o
// visitante para um endereco morto.
//
// Regra: link de NAVEGACAO e relativo; so canonical/hreflang/og_url -- lidos por
// bots que nao resolvem caminho relativo -- sao absolutos.
func TestSeletorIdiomaERelativo(t *testing.T) {
	h := novoServidor(t).Handler() // BASE_URL = https://exemplo.test

	casos := map[string]string{"/midiakit": `href="/en/midiakit"`, "/en/midiakit": `href="/midiakit"`}
	for caminho, esperado := range casos {
		w := httptest.NewRecorder()
		h.ServeHTTP(w, httptest.NewRequest(http.MethodGet, caminho, nil))
		corpo := w.Body.String()

		if !strings.Contains(corpo, `class="navbar__lang" `+esperado) {
			t.Errorf("%s: seletor de idioma deveria apontar para %s (relativo)", caminho, esperado)
		}
		if strings.Contains(corpo, `class="navbar__lang" href="https://exemplo.test`) {
			t.Errorf("%s: seletor de idioma voltou a usar URL absoluta (BASE_URL)", caminho)
		}
		// contraprova: o canonical PRECISA continuar absoluto.
		if !strings.Contains(corpo, `<link rel="canonical" href="https://exemplo.test`) {
			t.Errorf("%s: canonical deveria ser absoluto", caminho)
		}
	}
}

// TestPaginaTemSEO protege a razao de ser do projeto. A SPA servia
// <div id="root"></div> e o titulo "siteisabelpontes", entao o card de preview
// no WhatsApp e no Instagram (de onde vem o cliente da Isabel) saia vazio --
// aqueles bots nao executam JavaScript.
//
// Se alguem quebrar o <title> ou o canonical, isso tem que falhar no CI, nao
// ser descoberto quando um cliente nao clicar no link.
func TestPaginaTemSEO(t *testing.T) {
	s := novoServidor(t)
	w := httptest.NewRecorder()
	s.Handler().ServeHTTP(w, httptest.NewRequest(http.MethodGet, "/", nil))

	corpo := w.Body.String()
	for _, esperado := range []string{
		"<title>Isabel Pontes",
		`<meta name="description"`,
		`<link rel="canonical" href="https://exemplo.test/"`,
		"<h1>",
	} {
		if !strings.Contains(corpo, esperado) {
			t.Errorf("falta no HTML: %q", esperado)
		}
	}
}

// TestRotaInexistenteDa404: a SPA devolvia 200 em qualquer URL, o que faz o
// Google indexar pagina de erro como se fosse conteudo.
func TestRotaInexistenteDa404(t *testing.T) {
	s := novoServidor(t)
	w := httptest.NewRecorder()
	s.Handler().ServeHTTP(w, httptest.NewRequest(http.MethodGet, "/nao-existe", nil))
	if w.Code != http.StatusNotFound {
		t.Errorf("status = %d, quero 404", w.Code)
	}
}

// TestAssetTemHash: sem identificador de conteudo na URL, o Cache-Control
// immutable de um ano deixaria o visitante preso ao CSS antigo apos o deploy.
func TestAssetTemHash(t *testing.T) {
	s := novoServidor(t)
	w := httptest.NewRecorder()
	s.Handler().ServeHTTP(w, httptest.NewRequest(http.MethodGet, "/", nil))

	if !strings.Contains(w.Body.String(), "/static/css/app.css?v=") {
		t.Error("o <link> do CSS saiu sem hash de conteudo")
	}
}

// TestAssetFaltandoNaoDerrubaPagina: asset ausente deve degradar para um 404
// daquele recurso, nao para uma pagina em branco.
func TestAssetFaltandoNaoDerrubaPagina(t *testing.T) {
	h := newAssetHasher(site.Assets)
	if got := h.url("/static/nao-existe.css"); got != "/static/nao-existe.css" {
		t.Errorf("url = %q, quero o caminho original de volta", got)
	}
}
