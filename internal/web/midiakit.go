package web

import (
	"fmt"
	"html/template"
	"net/http"
	"strings"
)

// ── Mídia kit ─────────────────────────────────────────────────────────────────
//
// One-page de pitch UGC/IGC (ver docs/plano-juncao-midiakit.md §4b). Ao contrario
// do site de fotografia, que e multi-pagina, o midia kit e UMA pagina rolavel --
// e o formato que marca grande espera receber. As secoes (bio, portfolio,
// metricas, processo, investimento) sao ancoras na mesma pagina, nao rotas.
//
// FASE 8: conteudo real (bio, dicionario UGC/IGC, processo, metricas, pacotes).
// O portfolio de videos (#portfolio) segue placeholder ate a Fase 9; o download
// de PDF e a nav hibrida ficam pra Fase 12.

// midiaKitView e o ViewModel. Struct, nunca map (ver nota em server.go): o
// template so acessa campos que existem aqui, e um typo vira erro de teste.
type midiaKitView struct {
	pageData
	Tags          []string
	TempoMedio    string
	Processo      []passoProcesso
	Metricas      []metricaRede
	Investimentos []blocoPacotes
	Portfolio     portfolioView
}

func (s *Server) midiakit(w http.ResponseWriter, r *http.Request) {
	dados := s.novaPagina(
		"Mídia Kit — Isabel Pontes | UGC & Creator",
		"Conteúdo autêntico com estética elevada: UGC e IGC para marcas, por Isabel Pontes, criadora e diretora criativa em Curitiba.",
		"/midiakit",
	)
	// Hero de topo escuro -> navbar e logo brancos, como na home.
	dados.BodyClass = "hero-escuro"
	s.render.Page(w, r, "midiakit", http.StatusOK, midiaKitView{
		pageData:      dados,
		Tags:          []string{"Storytelling", "Direção de Arte", "Cinema", "Edição Caprichada"},
		TempoMedio:    processoTempoMedio,
		Processo:      processo,
		Metricas:      metricas,
		Investimentos: investimentos,
		Portfolio:     buildPortfolio(""), // "" = todas as categorias
	})
}

// ── Portfólio de vídeos (#portfolio, filtro via HTMX) ─────────────────────────
//
// Mesmo padrão da galeria de fotos: um partial (mk_portfolio) com o filtro E o
// grid num wrapper só, trocado por hx-swap=outerHTML para o botão ativo atualizar
// junto. O mesmo partial serve a página inteira e o fragmento -- se divergirem, o
// que o clique mostra difere do estado inicial. Dados: portfolioVideos (gerado em
// midiakit_videos.go). Diferente da galeria, o filtro NÃO muda a URL: é um widget
// interno da one-page, não uma rota com deep-link próprio.

type portfolioFiltro struct {
	Slug   string
	Nome   string
	Active bool
}

type portfolioView struct {
	Filtros []portfolioFiltro // "Todos" + uma por categoria
	Grupos  []categoriaVideos // o que renderizar: todas, ou só a filtrada
}

// buildPortfolio monta a view para uma categoria (slug vazio = todas).
func buildPortfolio(catSlug string) portfolioView {
	filtros := []portfolioFiltro{{Slug: "", Nome: "Todos", Active: catSlug == ""}}
	grupos := make([]categoriaVideos, 0, len(portfolioVideos))
	for _, c := range portfolioVideos {
		filtros = append(filtros, portfolioFiltro{Slug: c.Slug, Nome: c.Nome, Active: c.Slug == catSlug})
		if catSlug == "" || c.Slug == catSlug {
			grupos = append(grupos, c)
		}
	}
	return portfolioView{Filtros: filtros, Grupos: grupos}
}

func categoriaValida(slug string) bool {
	for _, c := range portfolioVideos {
		if c.Slug == slug {
			return true
		}
	}
	return false
}

// midiakitVideos serve SÓ o wrapper #mk-portfolio (filtro + grid) -- a resposta
// do HTMX ao clicar num filtro de categoria.
func (s *Server) midiakitVideos(w http.ResponseWriter, r *http.Request) {
	cat := r.URL.Query().Get("cat")
	if cat != "" && !categoriaValida(cat) {
		http.Error(w, "categoria desconhecida", http.StatusNotFound)
		return
	}
	s.render.Fragment(w, "midiakit", "mk_portfolio", buildPortfolio(cat))
}

// wppURL monta o link do WhatsApp com a mensagem ja pronta, direto no servidor --
// o mIdia kit original fazia isso em JS (data-wpp + listener). Aqui e um href
// normal, sem JS. template.URL porque a mensagem ja vem encodada por urlEncode
// (o filtro de URL do html/template reprocessaria e poderia estragar o %20).
func wppURL(mensagem string) template.URL {
	return template.URL("https://wa.me/" + whatsappMidiaKit + "?text=" + urlEncode(mensagem))
}

// donutGradient monta a string conic-gradient de um donut (genero/idade) a partir
// das fatias. Fica aqui, nao no template, pela regra do projeto: template burro,
// logica no Go. template.CSS porque o valor entra num style= e o html/template
// bloquearia um background computado por seguranca -- e conteudo nosso.
func donutGradient(fatias []fatia) template.CSS {
	var sb strings.Builder
	sb.WriteString("conic-gradient(")
	acumulado := 0.0
	for i, f := range fatias {
		if i > 0 {
			sb.WriteString(", ")
		}
		inicio := acumulado
		acumulado += f.Pct
		fmt.Fprintf(&sb, "%s %.1f%% %.1f%%", f.Cor, inicio, acumulado)
	}
	sb.WriteString(")")
	return template.CSS(sb.String())
}
