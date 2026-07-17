package web

import (
	"context"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strings"
	"testing"

	site "github.com/JoaoVictorCarvalho07/site-da-bebel-htmx"
	"github.com/JoaoVictorCarvalho07/site-da-bebel-htmx/internal/config"
	"github.com/JoaoVictorCarvalho07/site-da-bebel-htmx/internal/store"
)

// fixture com 2 temas; 'casal' e 'epoca' existem em labelsTema.
const galeriaFixture = `[
  {"key":"casal/casal_1.webp","tema":"casal","w":800,"h":1200,"grid":"https://cdn.test/g1","grid2x":"https://cdn.test/g1x","full":"https://cdn.test/f1","lqip":"data:,1"},
  {"key":"casal/casal_2.webp","tema":"casal","w":800,"h":1200,"grid":"https://cdn.test/g2","grid2x":"https://cdn.test/g2x","full":"https://cdn.test/f2","lqip":"data:,2"},
  {"key":"epoca/epoca_1.webp","tema":"epoca","w":800,"h":1200,"grid":"https://cdn.test/e1","grid2x":"https://cdn.test/e1x","full":"https://cdn.test/ef1","lqip":"data:,e1"}
]`

func servidorComFotos(t *testing.T) *Server {
	t.Helper()
	dir := t.TempDir()
	mf := filepath.Join(dir, "m.json")
	if err := os.WriteFile(mf, []byte(galeriaFixture), 0o644); err != nil {
		t.Fatal(err)
	}
	st, err := store.Open(filepath.Join(dir, "t.db"))
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { st.Close() })
	if _, err := st.SyncFromManifest(context.Background(), mf); err != nil {
		t.Fatal(err)
	}
	snap, err := st.LoadSnapshot(context.Background())
	if err != nil {
		t.Fatal(err)
	}
	s, err := New(&config.Config{BaseURL: "https://exemplo.test"}, site.Assets, snap)
	if err != nil {
		t.Fatal(err)
	}
	return s
}

// TestGaleriaPaginaEFragmento e a prova da REGRA DE OURO: a pagina completa e o
// fragmento HTMX tem que conter o mesmo #galeria-body com as mesmas fotos. Se
// divergirem, /galeria?tema=X colado no WhatsApp mostraria coisa diferente do
// clique no filtro, e o SEO quebraria em silencio.
func TestGaleriaPaginaEFragmento(t *testing.T) {
	s := servidorComFotos(t)
	h := s.Handler()

	wp := httptest.NewRecorder()
	h.ServeHTTP(wp, httptest.NewRequest(http.MethodGet, "/galeria", nil))
	if wp.Code != http.StatusOK {
		t.Fatalf("pagina: status %d, quero 200", wp.Code)
	}
	pagina := wp.Body.String()

	wf := httptest.NewRecorder()
	h.ServeHTTP(wf, httptest.NewRequest(http.MethodGet, "/galeria/fotos", nil))
	if wf.Code != http.StatusOK {
		t.Fatalf("fragmento: status %d, quero 200", wf.Code)
	}
	frag := wf.Body.String()

	// a pagina tem layout; o fragmento NAO.
	if !strings.Contains(pagina, "</html>") {
		t.Error("a pagina completa devia ter <html>")
	}
	if strings.Contains(frag, "<html") {
		t.Error("o fragmento NAO devia ter layout (<html>)")
	}

	// os dois contem o mesmo corpo e as mesmas fotos.
	for _, marca := range []string{
		`id="galeria-body"`,
		"https://cdn.test/f1", // full da foto 1 (lightbox)
		"https://cdn.test/g1", // grid da foto 1
		"data:,1",             // lqip da foto 1
	} {
		if !strings.Contains(pagina, marca) {
			t.Errorf("pagina nao contem %q", marca)
		}
		if !strings.Contains(frag, marca) {
			t.Errorf("fragmento nao contem %q", marca)
		}
	}
}

func TestGaleriaFiltraPorTema(t *testing.T) {
	s := servidorComFotos(t)
	w := httptest.NewRecorder()
	s.Handler().ServeHTTP(w, httptest.NewRequest(http.MethodGet, "/galeria?tema=epoca", nil))
	corpo := w.Body.String()
	// epoca tem 1 foto (ef1); as de casal (f1/f2) nao podem aparecer.
	if !strings.Contains(corpo, "https://cdn.test/ef1") {
		t.Error("filtro epoca devia mostrar a foto de epoca")
	}
	if strings.Contains(corpo, "https://cdn.test/f1") {
		t.Error("filtro epoca vazou uma foto de casal")
	}
	// o label acentuado aparece no botao ativo
	if !strings.Contains(corpo, "Época") {
		t.Error("o label acentuado 'Época' devia aparecer")
	}
}

func TestGaleriaTemaInvalidoDa404(t *testing.T) {
	s := servidorComFotos(t)
	w := httptest.NewRecorder()
	s.Handler().ServeHTTP(w, httptest.NewRequest(http.MethodGet, "/galeria?tema=naoexiste", nil))
	if w.Code != http.StatusNotFound {
		t.Errorf("tema inventado: status %d, quero 404 (nao lista vazia)", w.Code)
	}
}

func TestGaleriaETagE304(t *testing.T) {
	s := servidorComFotos(t)
	h := s.Handler()

	w1 := httptest.NewRecorder()
	h.ServeHTTP(w1, httptest.NewRequest(http.MethodGet, "/galeria", nil))
	etag := w1.Header().Get("ETag")
	if etag == "" {
		t.Fatal("sem ETag na resposta da galeria")
	}

	r2 := httptest.NewRequest(http.MethodGet, "/galeria", nil)
	r2.Header.Set("If-None-Match", etag)
	w2 := httptest.NewRecorder()
	h.ServeHTTP(w2, r2)
	if w2.Code != http.StatusNotModified {
		t.Errorf("com If-None-Match casando: status %d, quero 304", w2.Code)
	}
}
