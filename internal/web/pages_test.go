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
