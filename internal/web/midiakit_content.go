package web

// Conteúdo curado do mídia kit. Structs Go, não JSON externo, pelo mesmo motivo
// do sobreCarrossel em pages.go: o projeto prioriza type safety e o dado é curado.
//
// i18n (Fase 10): campos traduzíveis são do tipo `ml` (pt/en, ver i18n.go); o
// template resolve por {{ .Campo.T $.Lang }}. A copy EN é rascunho -- a Isa revisa.
// Números/preços/marcas não mudam entre idiomas.

// whatsappMidiaKit é o MESMO número do /contato (pages.go). QUESTÃO ABERTA (§11):
// 554191977011 tem 12 dígitos; confirmar o E.164 correto com a Isabel.
const whatsappMidiaKit = "554191977011"

// tags da bio (o ponto forte da Isa).
var midiakitTags = []ml{
	{"Storytelling", "Storytelling"},
	{"Direção de Arte", "Art Direction"},
	{"Cinema", "Film"},
	{"Edição Caprichada", "Meticulous Editing"},
}

// ── Investimento (pacotes) ────────────────────────────────────────────────────

type pacote struct {
	Nome     ml
	Resumo   ml   // o que é, ex.: "1 vídeo vertical"
	Itens    []ml // entregáveis
	Preco    string
	Destaque bool // "mais popular"
}

type blocoPacotes struct {
	Titulo    ml
	Subtitulo ml
	Tag       string // curto e estável, pra mensagem do WhatsApp: "UGC" / "Influencer"
	Pacotes   []pacote
}

var investimentos = []blocoPacotes{
	{
		Titulo:    ml{"Investimento UGC", "UGC Investment"},
		Subtitulo: ml{"Escolha a jornada ideal para sua marca.", "Choose the ideal journey for your brand."},
		Tag:       "UGC",
		Pacotes: []pacote{
			{
				Nome: ml{"Pacote 1", "Package 1"}, Resumo: ml{"1 vídeo vertical", "1 vertical video"}, Preco: "R$ 300",
				Itens: []ml{
					{"Roteiro simples", "Simple script"},
					{"1 foto", "1 photo"},
					{"Edição padrão", "Standard editing"},
					{"Legendas", "Captions"},
					{"1 revisão simples", "1 simple revision"},
					{"Tráfego pago por 3 meses (90 dias) + orgânico no mesmo período", "Paid traffic for 3 months (90 days) + organic in the same period"},
				},
			},
			{
				Nome: ml{"Pacote 2", "Package 2"}, Resumo: ml{"2 vídeos verticais", "2 vertical videos"}, Preco: "R$ 550", Destaque: true,
				Itens: []ml{
					{"2 roteiros humanizados", "2 humanized scripts"},
					{"2 fotos lifestyle", "2 lifestyle photos"},
					{"1 revisão simples", "1 simple revision"},
					{"Edição completa", "Full editing"},
					{"Legendas", "Captions"},
					{"Efeitos visuais/sonoros leves (dinâmica, transições, textos na tela)", "Light visual/sound effects (pacing, transitions, on-screen text)"},
					{"Tráfego pago por 3 meses (90 dias) + orgânico no mesmo período", "Paid traffic for 3 months (90 days) + organic in the same period"},
				},
			},
			{
				Nome: ml{"Pacote 3", "Package 3"}, Resumo: ml{"4 vídeos verticais", "4 vertical videos"}, Preco: "R$ 1100",
				Itens: []ml{
					{"4 roteiros humanizados (estratégia de campanha)", "4 humanized scripts (campaign strategy)"},
					{"Edição premium (mais dinâmica + acabamento)", "Premium editing (more pacing + finishing)"},
					{"2 revisões", "2 revisions"},
					{"Legendas", "Captions"},
					{"Efeitos visuais/sonoros leves (dinâmica, transições, textos na tela)", "Light visual/sound effects (pacing, transitions, on-screen text)"},
					{"Tráfego pago por 6 meses (180 dias) + orgânico no mesmo período", "Paid traffic for 6 months (180 days) + organic in the same period"},
				},
			},
		},
	},
	{
		Titulo: ml{"Investimento Influencer", "Influencer Investment"},
		Tag:    "Influencer",
		Pacotes: []pacote{
			{
				Nome: ml{"Pacote 1", "Package 1"}, Resumo: ml{"Sequência de 3 stories", "Sequence of 3 stories"}, Preco: "R$ 150",
				Itens: []ml{
					{"Copy curta", "Short copy"},
					{"Marcação", "Tag/mention"},
					{"Link/CTA (se tiver)", "Link/CTA (if any)"},
				},
			},
			{
				Nome: ml{"Pacote 2", "Package 2"}, Resumo: ml{"1 Reels + 1 foto de capa", "1 Reels + 1 cover photo"}, Preco: "R$ 300",
				Itens: []ml{
					{"1 foto de capa", "1 cover photo"},
					{"Sequência de 3 stories com CTA", "Sequence of 3 stories with CTA"},
					{"Marcação", "Tag/mention"},
				},
			},
			{
				Nome: ml{"Pacote 3", "Package 3"}, Resumo: ml{"2 Reels/mês + 1 story semanal (4 no mês)", "2 Reels/month + 1 weekly story (4 a month)"}, Preco: "R$ 700/mês",
				Itens: []ml{
					{"Planejamento do mês (datas) + CTA", "Monthly planning (dates) + CTA"},
					{"Marcação", "Tag/mention"},
				},
			},
		},
	},
}

// ── Processo de contratação ───────────────────────────────────────────────────

type passoProcesso struct {
	Num      string
	Titulo   ml
	Resumo   ml
	Prazo    ml
	Detalhes []ml
}

var processoTempoMedio = ml{
	"Tempo médio: 7–12 dias (dependendo da demanda)",
	"Average time: 7–12 days (depending on demand)",
}

var processo = []passoProcesso{
	{
		Num: "01", Titulo: ml{"Briefing & objetivos", "Briefing & goals"}, Prazo: ml{"~10 min", "~10 min"},
		Resumo: ml{"Alinhamos produto, público e meta da campanha.", "We align product, audience and campaign goal."},
		Detalhes: []ml{
			{"Você envia briefing, referências, tom de voz e infos do produto.", "You send the briefing, references, tone of voice and product info."},
			{"Definimos objetivo (awareness/conversão), formato (9:16) e CTA.", "We set the goal (awareness/conversion), format (9:16) and CTA."},
			{"Eu confirmo entregáveis e prazo ideal para gravação e edição.", "I confirm deliverables and the ideal timeline for shooting and editing."},
		},
	},
	{
		Num: "02", Titulo: ml{"Proposta & pacote ideal", "Proposal & ideal package"}, Prazo: ml{"rápido", "quick"},
		Resumo: ml{"Eu te envio opções de entregáveis com valores e prazos.", "I send you deliverable options with prices and timelines."},
		Detalhes: []ml{
			{"Pacotes com variações (hook/ângulo), versões com e sem legenda.", "Packages with variations (hook/angle), captioned and uncaptioned versions."},
			{"Definição de uso (orgânico/ads/whitelisting, se houver).", "Usage definition (organic/ads/whitelisting, if any)."},
			{"Se quiser: opções extras (UGC foto, cortes, thumbnails).", "If you want: extra options (photo UGC, cuts, thumbnails)."},
		},
	},
	{
		Num: "03", Titulo: ml{"Roteiro & aprovação rápida", "Script & quick approval"}, Prazo: ml{"1–2 dias", "1–2 days"},
		Resumo: ml{"Eu proponho hooks + estrutura do vídeo pra aprovar antes de gravar.", "I propose hooks + video structure for approval before shooting."},
		Detalhes: []ml{
			{"Hooks + bullet points (o que precisa ser dito / mostrado).", "Hooks + bullet points (what needs to be said / shown)."},
			{"Você aprova ou ajusta (sem burocracia).", "You approve or adjust (no bureaucracy)."},
			{"Depois disso, a gravação segue alinhada e eficiente.", "After that, the shoot runs aligned and efficient."},
		},
	},
	{
		Num: "04", Titulo: ml{"Produção & gravação", "Production & shooting"}, Prazo: ml{"1–3 dias", "1–3 days"},
		Resumo: ml{"Eu gravo com estética alinhada ao produto + tendência.", "I shoot with an aesthetic aligned to the product + trend."},
		Detalhes: []ml{
			{"Gravação em luz natural/estúdio (o que combinar mais).", "Shooting in natural light/studio (whichever fits best)."},
			{"Cenas de uso real + close do produto + prova social (quando faz sentido).", "Real-use scenes + product close-ups + social proof (when it makes sense)."},
			{"Captação pensando em cortes pra ads e variações.", "Footage planned for ad cuts and variations."},
		},
	},
	{
		Num: "05", Titulo: ml{"Edição & entrega final", "Editing & final delivery"}, Prazo: ml{"2–5 dias", "2–5 days"},
		Resumo: ml{"Eu edito (dinâmico + legendas) e te entrego tudo organizado.", "I edit (dynamic + captions) and deliver everything organized."},
		Detalhes: []ml{
			{"Entrega em alta + versão otimizada pra plataformas (Reels/TikTok).", "High-res delivery + version optimized for platforms (Reels/TikTok)."},
			{"Organização por pasta e nomes fáceis (ex.: Hook1_V1, Hook1_V2...).", "Organized by folder with clear names (e.g. Hook1_V1, Hook1_V2...)."},
			{"1 rodada de ajustes leves incluída (ex.: texto/legenda/corte).", "1 round of light adjustments included (e.g. text/caption/cut)."},
		},
	},
}

// ── Métricas (Instagram / TikTok) ─────────────────────────────────────────────

type barra struct {
	Rotulo ml
	Pct    float64
}

// fatia é um pedaço de um donut (gênero/idade); Cor é um hex fixo (ver abaixo).
type fatia struct {
	Rotulo ml
	Pct    float64
	Cor    string
}

type metricaRede struct {
	Rede       string // marca, igual nos dois idiomas
	Seguidores string
	Views      string
	Locais     []barra
	Genero     []fatia
	Idades     []fatia
	// Foto é o mockup de celular que flutua no card (WebP com alpha). Lado: "dir"/"esq".
	Foto string
	Lado string
}

// cores dos donuts (hex FIXO, não token: o card de métrica é um painel escuro
// fixo nos dois temas, então as fatias têm que ser claras e distintas).
const (
	corA = "#d8b45a" // dourado
	corB = "#efe3d0" // creme
	corC = "#5f9184" // verde claro
)

var (
	fem = ml{"Feminino", "Female"}
	mas = ml{"Masculino", "Male"}
	br  = ml{"Brasil", "Brazil"}
	cwb = ml{"Curitiba", "Curitiba"}
)

var metricas = []metricaRede{
	{
		Rede: "Instagram", Seguidores: "+3.2K", Views: "+36.1K",
		Locais: []barra{{br, 96.7}, {cwb, 29.5}},
		Genero: []fatia{{fem, 83.9, corA}, {mas, 16.1, corB}},
		Idades: []fatia{{ml{"18–24", "18–24"}, 36.9, corA}, {ml{"25–34", "25–34"}, 47.5, corB}, {ml{"+35", "+35"}, 15.6, corC}},
		Foto:   "/static/img/midiakit/phone-instagram.webp", Lado: "dir",
	},
	{
		Rede: "TikTok", Seguidores: "+7K", Views: "+17K",
		Locais: []barra{{br, 99.4}},
		Genero: []fatia{{fem, 78.0, corA}, {mas, 21.0, corB}},
		Idades: []fatia{{ml{"18–24", "18–24"}, 54.2, corA}, {ml{"25–34", "25–34"}, 35.9, corB}, {ml{"+35", "+35"}, 9.9, corC}},
		Foto:   "/static/img/midiakit/phone-tiktok.webp", Lado: "esq",
	},
}
