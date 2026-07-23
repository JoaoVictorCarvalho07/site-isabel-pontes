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
	Tags          []ml
	TempoMedio    ml
	Processo      []passoProcesso
	Metricas      []metricaRede
	Investimentos []blocoPacotes
	Portfolio     portfolioView
	// FormUGC e o briefing embutido na one-page. Fica DENTRO do mídia kit (e
	// nao numa /midiakit/briefing) para nao tirar a marca do documento no
	// momento exato em que ela decidiu falar com a Isa. Ver docs §8.
	FormUGC briefingView
}

func (s *Server) midiakit(w http.ResponseWriter, r *http.Request) {
	lang := idioma(r)
	titulo := "Mídia Kit — Isabel Pontes | UGC & Creator"
	descr := "Conteúdo autêntico com estética elevada: UGC e IGC para marcas, por Isabel Pontes, criadora e diretora criativa em Curitiba."
	if lang == "en" {
		titulo = "Media Kit — Isabel Pontes | UGC & Creator"
		descr = "Authentic content with elevated aesthetics: UGC and IGC for brands, by Isabel Pontes, creator and creative director in Curitiba."
	}
	dados := s.novaPagina(lang, titulo, descr, "/midiakit")
	// Hero de topo escuro -> navbar e logo brancos, como na home.
	dados.BodyClass = "hero-escuro"
	// O midia kit fecha com rodape proprio (contato + redes); esconde o padrao.
	dados.OcultarRodape = true
	// ... e e a unica pagina com o botao de baixar em PDF na navbar.
	dados.MostrarPdf = true
	s.render.Page(w, r, "midiakit", http.StatusOK, midiaKitView{
		pageData:      dados,
		Tags:          midiakitTags,
		TempoMedio:    processoTempoMedio,
		Processo:      processo,
		Metricas:      metricas,
		Investimentos: investimentos,
		Portfolio:     buildPortfolio(lang, ""), // "" = todas as categorias
		FormUGC:       s.formUGCView(dados, lang),
	})
}

// formUGCView monta o briefing embutido na one-page. Se o form de UGC estiver
// desligado no painel, devolve um view com Form=nil -- o template esconde a
// secao (ver o {{ with .FormUGC.Form }} em midiakit.gohtml).
func (s *Server) formUGCView(dados pageData, lang string) briefingView {
	f, ok := s.forms.get("ugc")
	if !ok {
		return briefingView{pageData: dados}
	}
	return briefingView{
		pageData: dados,
		Form:     f,
		Action:   caminhoEnvio(lang, f.Slug),
		Valores:  map[string]string{},
		Erros:    map[string]string{},
	}
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
	Nome   string // já traduzido
	Active bool
}

// grupoVideosView é uma categoria pronta pro template, com o nome já no idioma.
type grupoVideosView struct {
	Nome   string
	Videos []videoUGC
}

type portfolioView struct {
	Lang    string
	Base    string // "/midiakit/videos" ou "/en/midiakit/videos" (hx-get)
	Filtros []portfolioFiltro // "Todos" + uma por categoria
	Grupos  []grupoVideosView // o que renderizar: todas, ou só a filtrada
}

// nomeCategoria traduz o nome exibido da categoria (o slug é estável).
func nomeCategoria(lang, slug, nomePT string) string {
	if lang == "en" {
		if en, ok := nomeCategoriaEN[slug]; ok {
			return en
		}
	}
	return nomePT
}

// buildPortfolio monta a view para uma categoria (slug vazio = todas), no idioma.
func buildPortfolio(lang, catSlug string) portfolioView {
	base := "/midiakit/videos"
	if lang == "en" {
		base = "/en/midiakit/videos"
	}
	filtros := []portfolioFiltro{{Slug: "", Nome: t(lang, "comum.todos"), Active: catSlug == ""}}
	grupos := make([]grupoVideosView, 0, len(portfolioVideos))
	for _, c := range portfolioVideos {
		nome := nomeCategoria(lang, c.Slug, c.Nome)
		filtros = append(filtros, portfolioFiltro{Slug: c.Slug, Nome: nome, Active: c.Slug == catSlug})
		if catSlug == "" || c.Slug == catSlug {
			grupos = append(grupos, grupoVideosView{Nome: nome, Videos: c.Videos})
		}
	}
	return portfolioView{Lang: lang, Base: base, Filtros: filtros, Grupos: grupos}
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
	s.render.Fragment(w, "midiakit", "mk_portfolio", buildPortfolio(idioma(r), cat))
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
