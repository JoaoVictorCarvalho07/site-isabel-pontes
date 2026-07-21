package web

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

// TestPaginasRenderizam exercita as paginas com ViewModel proprio pelos handlers
// reais (com um servidor populado), o que o teste generico de template nao faz.
// Pega, por construcao, um campo que o handler esqueca de preencher ou um
// template que referencie campo inexistente.
func TestPaginasRenderizam(t *testing.T) {
	s := servidorComFotos(t)
	h := s.Handler()

	casos := []struct {
		rota   string
		contem []string
	}{
		{"/", []string{"</html>", "hero", "Bem-vindo", "cards-editoriais"}},
		{"/sobre", []string{"</html>", "Meu nome é Isabel Pontes", "carrossel"}},
		{"/contato", []string{"</html>", "wa.me", "instagram.com", "mailto:"}},
		{"/midiakit", []string{
			"</html>", "mk-hero", "Mídia kit",
			"Investimento UGC", "R$ 550", "Mais popular", // pacotes
			"wa.me/554191977011", // CTA WhatsApp montado no servidor
			"conic-gradient",     // donut de métrica
			"Briefing &amp; objetivos", // processo
			"mk-filtro", "beleza-loreal.avif", "data-video-src", // portfolio de videos
		}},
	}

	for _, c := range casos {
		t.Run(c.rota, func(t *testing.T) {
			w := httptest.NewRecorder()
			h.ServeHTTP(w, httptest.NewRequest(http.MethodGet, c.rota, nil))
			if w.Code != http.StatusOK {
				t.Fatalf("status %d, quero 200. Corpo: %.400s", w.Code, w.Body.String())
			}
			corpo := w.Body.String()
			for _, marca := range c.contem {
				if !strings.Contains(corpo, marca) {
					t.Errorf("%s nao contem %q", c.rota, marca)
				}
			}
		})
	}
}

// TestHomeUsaFotoDoDia: o card 1 da home mostra uma foto da galeria (a primeira
// do shuffle do dia). Se o snapshot tiver fotos, a URL do bucket tem que aparecer.
func TestHomeUsaFotoDoDia(t *testing.T) {
	s := servidorComFotos(t)
	w := httptest.NewRecorder()
	s.Handler().ServeHTTP(w, httptest.NewRequest(http.MethodGet, "/", nil))
	if !strings.Contains(w.Body.String(), "https://cdn.test/") {
		t.Error("a home devia embutir a foto do dia (URL do bucket)")
	}
}

// TestPortfolioVideosFragmento: o filtro de categoria do portfolio (HTMX) devolve
// so a categoria pedida, com o poster, e 404 para categoria inventada -- o mesmo
// contrato da galeria de fotos.
func TestPortfolioVideosFragmento(t *testing.T) {
	s := servidorComFotos(t)
	h := s.Handler()

	w := httptest.NewRecorder()
	h.ServeHTTP(w, httptest.NewRequest(http.MethodGet, "/midiakit/videos?cat=beleza", nil))
	if w.Code != http.StatusOK {
		t.Fatalf("status %d, quero 200. Corpo: %.300s", w.Code, w.Body.String())
	}
	corpo := w.Body.String()
	for _, m := range []string{"mk-portfolio", "Beleza", "beleza-loreal.avif", "data-video-src"} {
		if !strings.Contains(corpo, m) {
			t.Errorf("fragmento nao contem %q", m)
		}
	}
	// "Eventos" aparece nos botões de filtro (correto); o que NÃO pode aparecer é
	// um vídeo de outra categoria no grid.
	if strings.Contains(corpo, "eventos-natura-yoga.avif") {
		t.Error("filtro 'beleza' nao devia renderizar videos de 'Eventos'")
	}

	// categoria inexistente -> 404 (nao lista vazia)
	w = httptest.NewRecorder()
	h.ServeHTTP(w, httptest.NewRequest(http.MethodGet, "/midiakit/videos?cat=naoexiste", nil))
	if w.Code != http.StatusNotFound {
		t.Errorf("status %d, quero 404 para categoria invalida", w.Code)
	}
}

// TestMailtoEncodado: o corpo do mailto/wa.me precisa de %20 (nao '+') e %0A,
// como o encodeURIComponent do JS produzia.
func TestMailtoEncodado(t *testing.T) {
	s := servidorComFotos(t)
	w := httptest.NewRecorder()
	s.Handler().ServeHTTP(w, httptest.NewRequest(http.MethodGet, "/contato", nil))
	corpo := w.Body.String()
	if strings.Contains(corpo, "Vim+pelo") {
		t.Error("espaco codificado como '+' -- deveria ser %20")
	}
	if !strings.Contains(corpo, "%20") {
		t.Error("faltou %20 no href de contato")
	}
}
