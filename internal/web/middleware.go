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
			if panico := recover(); panico != nil {
				logger.Error("panic no handler", "erro", panico, "path", r.URL.Path)
				http.Error(w, "Erro interno", http.StatusInternalServerError)
			}
		}()
		next.ServeHTTP(w, r)
	})
}

// statusWriter embrulha o ResponseWriter so para lembrar qual status foi escrito,
// para o log conseguir registra-lo (o ResponseWriter padrao nao expoe isso).
type statusWriter struct {
	http.ResponseWriter
	status int
}

func (w *statusWriter) WriteHeader(codigo int) {
	w.status = codigo
	w.ResponseWriter.WriteHeader(codigo)
}

func requestLog(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		inicio := time.Now()
		resp := &statusWriter{ResponseWriter: w, status: http.StatusOK}
		next.ServeHTTP(resp, r)
		logger.Info("req",
			"metodo", r.Method,
			"path", r.URL.RequestURI(),
			"status", resp.status,
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

// poolCompressores reaproveita gzip.Writers entre requests (criar um a cada
// request geraria lixo pro GC). O io.Discard e so um destino temporario -- o
// Reset() troca pelo ResponseWriter real na hora de usar.
var poolCompressores = sync.Pool{New: func() any { return gzip.NewWriter(io.Discard) }}

// compressivel: so os tipos textuais valem a pena. woff2 e avif ja sao
// comprimidos; gzipa-los desperdicaria CPU e quebraria o Content-Length/range
// do FileServer sem ganho.
func compressivel(tipoConteudo string) bool {
	tipoConteudo = strings.ToLower(tipoConteudo)
	for _, prefixo := range []string{
		"text/html", "text/css", "text/plain",
		"application/javascript", "text/javascript",
		"application/json", "image/svg+xml", "application/xml", "text/xml",
	} {
		if strings.HasPrefix(tipoConteudo, prefixo) {
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
	compressor *gzip.Writer
	jaDecidiu  bool
	comprimir  bool
}

func (w *gzipWriter) decidir() {
	if w.jaDecidiu {
		return
	}
	w.jaDecidiu = true
	if compressivel(w.Header().Get("Content-Type")) {
		w.comprimir = true
		w.Header().Set("Content-Encoding", "gzip")
		w.Header().Add("Vary", "Accept-Encoding")
		w.Header().Del("Content-Length") // errado sob gzip
		w.compressor.Reset(w.ResponseWriter)
	}
}

func (w *gzipWriter) WriteHeader(codigo int) {
	w.decidir()
	w.ResponseWriter.WriteHeader(codigo)
}

func (w *gzipWriter) Write(dados []byte) (int, error) {
	w.decidir() // handlers que escrevem sem WriteHeader (200 implicito)
	if w.comprimir {
		return w.compressor.Write(dados)
	}
	return w.ResponseWriter.Write(dados)
}

func gzipMW(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if !strings.Contains(r.Header.Get("Accept-Encoding"), "gzip") {
			next.ServeHTTP(w, r)
			return
		}
		compressor := poolCompressores.Get().(*gzip.Writer)
		defer poolCompressores.Put(compressor)
		respGzip := &gzipWriter{ResponseWriter: w, compressor: compressor}
		defer func() {
			if respGzip.comprimir {
				_ = compressor.Close()
			}
		}()
		next.ServeHTTP(respGzip, r)
	})
}
