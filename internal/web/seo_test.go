package web

import (
	"encoding/xml"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

// TestOGTags protege a razao de ser do projeto: o card de preview no
// WhatsApp/Instagram. Se og:image/og:title/og:url sumirem, o preview volta a
// sair vazio -- e isso tem que quebrar o CI, nao ser descoberto por um cliente
// que nao clicou no link.
func TestOGTags(t *testing.T) {
	s := servidorComFotos(t)
	for _, rota := range []string{"/", "/galeria", "/sobre", "/contato"} {
		t.Run(rota, func(t *testing.T) {
			w := httptest.NewRecorder()
			s.Handler().ServeHTTP(w, httptest.NewRequest(http.MethodGet, rota, nil))
			corpo := w.Body.String()
			for _, marca := range []string{
				`property="og:title"`,
				`property="og:image" content="https://exemplo.test/static/img/og-image.jpg"`,
				`property="og:url" content="https://exemplo.test` + rota,
				`name="twitter:card" content="summary_large_image"`,
				`application/ld+json`,
				`"@type": "ProfessionalService"`,
			} {
				if !strings.Contains(corpo, marca) {
					t.Errorf("%s: falta %q", rota, marca)
				}
			}
		})
	}
}

func TestRobots(t *testing.T) {
	s := servidorComFotos(t)
	w := httptest.NewRecorder()
	s.Handler().ServeHTTP(w, httptest.NewRequest(http.MethodGet, "/robots.txt", nil))
	if w.Code != http.StatusOK {
		t.Fatalf("status %d", w.Code)
	}
	if ct := w.Header().Get("Content-Type"); !strings.HasPrefix(ct, "text/plain") {
		t.Errorf("content-type = %q, quero text/plain", ct)
	}
	corpo := w.Body.String()
	if !strings.Contains(corpo, "Sitemap: https://exemplo.test/sitemap.xml") {
		t.Errorf("robots sem linha de Sitemap:\n%s", corpo)
	}
}

// TestSitemapValido: XML bem-formado e com as rotas esperadas, inclusive uma por
// tema (que sao URLs reais).
func TestSitemapValido(t *testing.T) {
	s := servidorComFotos(t)
	w := httptest.NewRecorder()
	s.Handler().ServeHTTP(w, httptest.NewRequest(http.MethodGet, "/sitemap.xml", nil))
	if w.Code != http.StatusOK {
		t.Fatalf("status %d", w.Code)
	}

	// bem-formado?
	var urlset struct {
		URLs []struct {
			Loc string `xml:"loc"`
		} `xml:"url"`
	}
	if err := xml.Unmarshal(w.Body.Bytes(), &urlset); err != nil {
		t.Fatalf("sitemap nao e XML valido: %v", err)
	}

	locs := map[string]bool{}
	for _, u := range urlset.URLs {
		locs[u.Loc] = true
	}
	for _, esperado := range []string{
		"https://exemplo.test/",
		"https://exemplo.test/galeria",
		"https://exemplo.test/galeria?tema=casal", // fixture tem casal
		"https://exemplo.test/sobre",
		"https://exemplo.test/contato",
	} {
		if !locs[esperado] {
			t.Errorf("sitemap sem %q", esperado)
		}
	}
}
