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
	// /admin fica fora do indice. Nao e seguranca (o basic-auth e que protege),
	// e higiene: um painel indexado convida tentativa de login.
	fmt.Fprintf(w, "User-agent: *\nAllow: /\nDisallow: /admin\n\nSitemap: %s/sitemap.xml\n", s.cfg.BaseURL)
}

// sitemap lista as URLs para o Google achar todas as paginas, inclusive uma por
// tema da galeria (que sao URLs reais, gracas ao par pagina/fragmento).
func (s *Server) sitemap(w http.ResponseWriter, _ *http.Request) {
	var xml strings.Builder
	xml.WriteString(`<?xml version="1.0" encoding="UTF-8"?>` + "\n")
	xml.WriteString(`<urlset xmlns="http://www.sitemaps.org/schemas/sitemap/0.9">` + "\n")

	// cada pagina entra nas duas versoes (PT e EN, via enPath).
	adicionar := func(caminhoPT string) {
		xml.WriteString("  <url><loc>" + xmlEscape(s.cfg.BaseURL+caminhoPT) + "</loc></url>\n")
		xml.WriteString("  <url><loc>" + xmlEscape(s.cfg.BaseURL+enPath(caminhoPT)) + "</loc></url>\n")
	}
	adicionar("/")
	adicionar("/galeria")
	for _, tema := range s.snap.Temas() {
		adicionar("/galeria?tema=" + tema.Slug)
	}
	adicionar("/sobre")
	adicionar("/midiakit")
	adicionar("/contato")
	adicionar("/briefing")

	xml.WriteString("</urlset>\n")

	w.Header().Set("Content-Type", "application/xml; charset=utf-8")
	w.Header().Set("Cache-Control", "public, max-age=3600")
	_, _ = w.Write([]byte(xml.String()))
}

func xmlEscape(texto string) string {
	texto = strings.ReplaceAll(texto, "&", "&amp;") // primeiro, senao re-escapa os outros
	texto = strings.ReplaceAll(texto, "<", "&lt;")
	texto = strings.ReplaceAll(texto, ">", "&gt;")
	return texto
}
