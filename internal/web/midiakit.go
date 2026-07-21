package web

import "net/http"

// ── Mídia kit ─────────────────────────────────────────────────────────────────
//
// One-page de pitch UGC/IGC (ver docs/plano-juncao-midiakit.md §4b). Ao contrario
// do site de fotografia, que e multi-pagina (navegacao de verdade), o midia kit e
// UMA pagina rolavel -- e o formato que marca grande espera receber. As "secoes"
// (bio, portfolio, metricas, processo, investimento) sao ancoras na mesma pagina,
// nao rotas separadas.
//
// FASE 7 (fundacao): so o esqueleto. Layout/navbar/footer compartilhados e as
// secoes com seus ids. O conteudo entra na Fase 8; o portfolio de videos, na
// Fase 9; o download de PDF, na Fase 12.
//
// Sem ViewModel proprio ainda: o esqueleto nao tem campo alem do pageData base,
// entao passa dados direto (igual notFound). Quando a Fase 8 trouxer pacotes e
// metricas, isto ganha um midiaKitView struct -- nunca map (ver nota em pages.go).

func (s *Server) midiakit(w http.ResponseWriter, r *http.Request) {
	dados := s.novaPagina(
		"Mídia Kit — Isabel Pontes | UGC & Creator",
		"Conteúdo autêntico com estética elevada: UGC e IGC para marcas, por Isabel Pontes, criadora e diretora criativa em Curitiba.",
		"/midiakit",
	)
	// Hero de topo escuro -> navbar e logo brancos, como na home.
	dados.BodyClass = "hero-escuro"
	s.render.Page(w, r, "midiakit", http.StatusOK, dados)
}
