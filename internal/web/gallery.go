package web

import (
	"fmt"
	"html/template"
	"net/http"
)

// temaView e um botao do filtro. Active marca o selecionado.
type temaView struct {
	Slug   string
	Label  string
	Active bool
}

// photoView e um item do grid, com tudo pronto para o template (o template fica
// burro; toda logica vive aqui).
type photoView struct {
	Grid   string
	Grid2x string
	Full   string
	// LQIP e um data: URI. Precisa ser template.URL: o filtro do html/template
	// troca data: em src por "#ZgotmplZ" (protecao anti-XSS), o que apagaria o
	// placeholder. Vem do nosso migrate_media.py, entao marcar como confiavel e
	// seguro. Os grid/full sao https e passam pelo filtro normal.
	LQIP      template.URL
	W, H      int
	Alt       string
	TemaLabel string
	Idx       int
	DelayMS   int // fadeUp escalonado; capado para nao dar 11s no ultimo item
}

// galleryView embute pageData (Titulo/Canonical/... para o layout) e carrega o
// que os partials filtros/grid precisam. O MESMO struct alimenta a pagina
// completa e o fragmento -- e isso que garante que os dois rendam identico.
type galleryView struct {
	pageData
	Temas     []temaView
	TemaAtivo string // slug; "" = Todos
	Fotos     []photoView
	Total     int
}

// buildGalleryView monta a ViewModel para um tema (slug vazio = Todos).
func (s *Server) buildGalleryView(temaSlug string) galleryView {
	temas := s.snap.Temas()

	// mapa slug->label para nao repetir a busca a cada foto
	labelPorSlug := map[string]string{}

	// os botoes do filtro: "Todos" + um por tema
	botoesFiltro := make([]temaView, 0, len(temas)+1)
	botoesFiltro = append(botoesFiltro, temaView{Slug: "", Label: "Todos", Active: temaSlug == ""})
	for _, tema := range temas {
		labelPorSlug[tema.Slug] = tema.Label
		botoesFiltro = append(botoesFiltro, temaView{
			Slug: tema.Slug, Label: tema.Label, Active: tema.Slug == temaSlug,
		})
	}

	// os itens do grid, ja prontos para o template
	fotosDoTema := s.snap.Fotos(temaSlug)
	itensGrid := make([]photoView, len(fotosDoTema))
	for i, foto := range fotosDoTema {
		label := labelPorSlug[foto.TemaSlug]
		alt := foto.Alt
		if alt == "" {
			alt = "Ensaio de " + label + " — Isabel Pontes"
		}
		itensGrid[i] = photoView{
			Grid: foto.Grid, Grid2x: foto.Grid2x, Full: foto.Full,
			LQIP: template.URL(foto.LQIP), //nolint:gosec // data: URI proprio, do migrate_media
			W:    foto.W, H: foto.H, Alt: alt, TemaLabel: label, Idx: i,
			DelayMS: (i % 12) * 45, // onda por grupo de 12; o React usava i*60 (11s no fim)
		}
	}

	titulo := "Galeria — Isabel Pontes"
	caminhoCanonico := "/galeria"
	if temaSlug != "" {
		titulo = labelPorSlug[temaSlug] + " — Galeria — Isabel Pontes"
		caminhoCanonico = "/galeria?tema=" + temaSlug
	}

	dadosBase := s.novaPagina(titulo,
		"Ensaios de fotografia artística por Isabel Pontes, em Curitiba.",
		caminhoCanonico)

	return galleryView{
		pageData:  dadosBase,
		Temas:     botoesFiltro,
		TemaAtivo: temaSlug,
		Fotos:     itensGrid,
		Total:     len(itensGrid),
	}
}

// galeria serve a pagina completa. ?tema=slug filtra; tema inexistente -> 404
// (nao lista vazia), para o Google nao indexar URL de filtro invalida como
// pagina real.
func (s *Server) galeria(w http.ResponseWriter, r *http.Request) {
	s.snap.EnsureDay()
	tema := r.URL.Query().Get("tema")
	if tema != "" && !s.snap.TemaValido(tema) {
		s.notFound(w, r)
		return
	}
	if s.setGalleryCache(w, r, tema) {
		return // 304
	}
	s.render.Page(w, r, "galeria", http.StatusOK, s.buildGalleryView(tema))
}

// galeriaFragment serve SO o corpo (#galeria-body), sem layout -- a resposta do
// HTMX ao clicar num filtro. Renderiza a mesma partial que a pagina usa.
func (s *Server) galeriaFragment(w http.ResponseWriter, r *http.Request) {
	s.snap.EnsureDay()
	tema := r.URL.Query().Get("tema")
	if tema != "" && !s.snap.TemaValido(tema) {
		http.Error(w, "tema desconhecido", http.StatusNotFound)
		return
	}
	if s.setGalleryCache(w, r, tema) {
		return
	}
	s.render.Fragment(w, "galeria", "galeria_body", s.buildGalleryView(tema))
}

// setGalleryCache põe o ETag (tema + dia do shuffle + data_version) e o
// Cache-Control, e responde 304 se o cliente ja tem a versao. O ETag muda quando
// o shuffle do dia vira ou quando um sync incrementa data_version -- exatamente
// as duas coisas que mudam o HTML. Retorna true se ja respondeu 304.
func (s *Server) setGalleryCache(w http.ResponseWriter, r *http.Request, tema string) bool {
	etag := fmt.Sprintf(`W/"g-%s-%s-v%d-b%s"`, tema, s.snap.DiaSeed(), s.snap.DataVersion(), s.boot)
	w.Header().Set("ETag", etag)
	// s-maxage deixa a CDN cachear por 1h; stale-while-revalidate serve o antigo
	// enquanto revalida. max-age=0 obriga o browser a revalidar (barato: 304).
	w.Header().Set("Cache-Control", "public, max-age=0, s-maxage=3600, stale-while-revalidate=86400")
	if match := r.Header.Get("If-None-Match"); match != "" && match == etag {
		w.WriteHeader(http.StatusNotModified)
		return true
	}
	return false
}
