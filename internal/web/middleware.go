package web

import (
	"net/http"
	"time"
)

// recoverer impede que um panic num handler derrube o processo inteiro.
// Num binario unico servindo o site todo, sem isto um nil pointer numa foto
// tira o site do ar -- nao so aquela request.
func recoverer(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		defer func() {
			if rec := recover(); rec != nil {
				logger.Error("panic no handler", "erro", rec, "path", r.URL.Path)
				http.Error(w, "Erro interno", http.StatusInternalServerError)
			}
		}()
		next.ServeHTTP(w, r)
	})
}

type statusWriter struct {
	http.ResponseWriter
	status int
}

func (w *statusWriter) WriteHeader(c int) {
	w.status = c
	w.ResponseWriter.WriteHeader(c)
}

func requestLog(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		inicio := time.Now()
		sw := &statusWriter{ResponseWriter: w, status: http.StatusOK}
		next.ServeHTTP(sw, r)
		logger.Info("req",
			"metodo", r.Method,
			"path", r.URL.RequestURI(),
			"status", sw.status,
			"ms", time.Since(inicio).Milliseconds(),
		)
	})
}

// cacheForever so pode ser aplicado a caminhos com hash de conteudo na URL.
// Aplicar isto num caminho estavel deixaria o browser com CSS velho por um ano.
func cacheForever(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Cache-Control", "public, max-age=31536000, immutable")
		next.ServeHTTP(w, r)
	})
}
