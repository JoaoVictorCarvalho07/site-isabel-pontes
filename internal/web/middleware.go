package web

import (
	"compress/gzip"
	"io"
	"net/http"
	"strings"
	"sync"
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

var gzipPool = sync.Pool{New: func() any { return gzip.NewWriter(io.Discard) }}

// compressivel: so os tipos textuais. woff2 e avif ja sao comprimidos; gzipa-los
// desperdicaria CPU e quebraria o Content-Length/range do FileServer sem ganho.
func compressivel(ct string) bool {
	ct = strings.ToLower(ct)
	for _, p := range []string{
		"text/html", "text/css", "text/plain",
		"application/javascript", "text/javascript",
		"application/json", "image/svg+xml", "application/xml", "text/xml",
	} {
		if strings.HasPrefix(ct, p) {
			return true
		}
	}
	return false
}

// gzipWriter decide comprimir ou nao pelo Content-Type da RESPOSTA, no momento
// do WriteHeader (ou do primeiro Write). Isso e o que evita gzipar imagem/fonte
// e preserva Content-Length/Accept-Ranges dos assets estaticos.
type gzipWriter struct {
	http.ResponseWriter
	gz       *gzip.Writer
	decidido bool
	usarGz   bool
}

func (w *gzipWriter) decidir() {
	if w.decidido {
		return
	}
	w.decidido = true
	if compressivel(w.Header().Get("Content-Type")) {
		w.usarGz = true
		w.Header().Set("Content-Encoding", "gzip")
		w.Header().Add("Vary", "Accept-Encoding")
		w.Header().Del("Content-Length") // errado sob gzip
		w.gz.Reset(w.ResponseWriter)
	}
}

func (w *gzipWriter) WriteHeader(code int) {
	w.decidir()
	w.ResponseWriter.WriteHeader(code)
}

func (w *gzipWriter) Write(b []byte) (int, error) {
	w.decidir() // handlers que escrevem sem WriteHeader (200 implicito)
	if w.usarGz {
		return w.gz.Write(b)
	}
	return w.ResponseWriter.Write(b)
}

func gzipMW(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if !strings.Contains(r.Header.Get("Accept-Encoding"), "gzip") {
			next.ServeHTTP(w, r)
			return
		}
		gz := gzipPool.Get().(*gzip.Writer)
		defer gzipPool.Put(gz)
		gw := &gzipWriter{ResponseWriter: w, gz: gz}
		defer func() {
			if gw.usarGz {
				_ = gz.Close()
			}
		}()
		next.ServeHTTP(gw, r)
	})
}
