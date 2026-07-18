package web

import (
	"fmt"
	"net/http"
	"strings"
)

// robots libera todo o site aos crawlers e aponta o sitemap. Nao existia na SPA
// (o catch-all devolvia a pagina em 200 para /robots.txt).
func (s *Server) robots(w http.ResponseWriter, _ *http.Request) {
	w.Header().Set("Content-Type", "text/plain; charset=utf-8")
	w.Header().Set("Cache-Control", "public, max-age=86400")
	fmt.Fprintf(w, "User-agent: *\nAllow: /\n\nSitemap: %s/sitemap.xml\n", s.cfg.BaseURL)
}

// sitemap lista as URLs para o Google achar todas as paginas, inclusive uma por
// tema da galeria (que sao URLs reais, gracas ao par pagina/fragmento).
func (s *Server) sitemap(w http.ResponseWriter, _ *http.Request) {
	var b strings.Builder
	b.WriteString(`<?xml version="1.0" encoding="UTF-8"?>` + "\n")
	b.WriteString(`<urlset xmlns="http://www.sitemaps.org/schemas/sitemap/0.9">` + "\n")

	add := func(path string) {
		b.WriteString("  <url><loc>" + xmlEscape(s.cfg.BaseURL+path) + "</loc></url>\n")
	}
	add("/")
	add("/galeria")
	for _, t := range s.snap.Temas() {
		add("/galeria?tema=" + t.Slug)
	}
	add("/sobre")
	add("/contato")

	b.WriteString("</urlset>\n")

	w.Header().Set("Content-Type", "application/xml; charset=utf-8")
	w.Header().Set("Cache-Control", "public, max-age=3600")
	_, _ = w.Write([]byte(b.String()))
}

func xmlEscape(s string) string {
	s = strings.ReplaceAll(s, "&", "&amp;") // primeiro, senao re-escapa os outros
	s = strings.ReplaceAll(s, "<", "&lt;")
	s = strings.ReplaceAll(s, ">", "&gt;")
	return s
}
