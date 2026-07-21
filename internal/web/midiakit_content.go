package web

// Conteúdo curado do mídia kit (Fase 8). Structs Go, não JSON externo, pelo mesmo
// motivo do sobreCarrossel em pages.go: o projeto prioriza type safety (struct,
// nunca map -- ver nota em server.go) e o dado é curado, não editado por não-dev.
// O §2 do plano lista "struct Go" como forma válida de "arquivos simples". Quando
// a Fase 2 do plano (SQLite editável) chegar, isto migra pro banco.
//
// Números e preços portados do mídia kit estático original (Midiakit_isabelPontes).

// whatsappMidiaKit é o MESMO número do /contato (pages.go). QUESTÃO ABERTA (§11):
// 554191977011 tem 12 dígitos; confirmar o E.164 correto com a Isabel.
const whatsappMidiaKit = "554191977011"

// ── Investimento (pacotes) ────────────────────────────────────────────────────

type pacote struct {
	Nome     string
	Resumo   string   // o que é, ex.: "1 vídeo vertical"
	Itens    []string // entregáveis
	Preco    string   // ex.: "R$ 300"
	Destaque bool     // "mais popular"
}

type blocoPacotes struct {
	Titulo    string
	Subtitulo string
	Tag       string // curto, pra mensagem do WhatsApp: "UGC" / "Influencer"
	Pacotes   []pacote
}

var investimentos = []blocoPacotes{
	{
		Titulo:    "Investimento UGC",
		Subtitulo: "Escolha a jornada ideal para sua marca.",
		Tag:       "UGC",
		Pacotes: []pacote{
			{
				Nome: "Pacote 1", Resumo: "1 vídeo vertical", Preco: "R$ 300",
				Itens: []string{
					"Roteiro simples", "1 foto", "Edição padrão", "Legendas",
					"1 revisão simples",
					"Tráfego pago por 3 meses (90 dias) + orgânico no mesmo período",
				},
			},
			{
				Nome: "Pacote 2", Resumo: "2 vídeos verticais", Preco: "R$ 550", Destaque: true,
				Itens: []string{
					"2 roteiros humanizados", "2 fotos lifestyle", "1 revisão simples",
					"Edição completa", "Legendas",
					"Efeitos visuais/sonoros leves (dinâmica, transições, textos na tela)",
					"Tráfego pago por 3 meses (90 dias) + orgânico no mesmo período",
				},
			},
			{
				Nome: "Pacote 3", Resumo: "4 vídeos verticais", Preco: "R$ 1100",
				Itens: []string{
					"4 roteiros humanizados (estratégia de campanha)",
					"Edição premium (mais dinâmica + acabamento)", "2 revisões",
					"Legendas",
					"Efeitos visuais/sonoros leves (dinâmica, transições, textos na tela)",
					"Tráfego pago por 6 meses (180 dias) + orgânico no mesmo período",
				},
			},
		},
	},
	{
		Titulo: "Investimento Influencer",
		Tag:    "Influencer",
		Pacotes: []pacote{
			{
				Nome: "Pacote 1", Resumo: "Sequência de 3 stories", Preco: "R$ 150",
				Itens: []string{"Copy curta", "Marcação", "Link/CTA (se tiver)"},
			},
			{
				Nome: "Pacote 2", Resumo: "1 Reels + 1 foto de capa", Preco: "R$ 300",
				Itens: []string{"1 foto de capa", "Sequência de 3 stories com CTA", "Marcação"},
			},
			{
				Nome: "Pacote 3", Resumo: "2 Reels/mês + 1 story semanal (4 no mês)", Preco: "R$ 700/mês",
				Itens: []string{"Planejamento do mês (datas) + CTA", "Marcação"},
			},
		},
	},
}

// ── Processo de contratação ───────────────────────────────────────────────────

type passoProcesso struct {
	Num      string   // "01"
	Titulo   string
	Resumo   string
	Prazo    string   // "~10 min", "1–2 dias"
	Detalhes []string // itens do <details> "Entenda"
}

const processoTempoMedio = "Tempo médio: 7–12 dias (dependendo da demanda)"

var processo = []passoProcesso{
	{
		Num: "01", Titulo: "Briefing & objetivos", Prazo: "~10 min",
		Resumo: "Alinhamos produto, público e meta da campanha.",
		Detalhes: []string{
			"Você envia briefing, referências, tom de voz e infos do produto.",
			"Definimos objetivo (awareness/conversão), formato (9:16) e CTA.",
			"Eu confirmo entregáveis e prazo ideal para gravação e edição.",
		},
	},
	{
		Num: "02", Titulo: "Proposta & pacote ideal", Prazo: "rápido",
		Resumo: "Eu te envio opções de entregáveis com valores e prazos.",
		Detalhes: []string{
			"Pacotes com variações (hook/ângulo), versões com e sem legenda.",
			"Definição de uso (orgânico/ads/whitelisting, se houver).",
			"Se quiser: opções extras (UGC foto, cortes, thumbnails).",
		},
	},
	{
		Num: "03", Titulo: "Roteiro & aprovação rápida", Prazo: "1–2 dias",
		Resumo: "Eu proponho hooks + estrutura do vídeo pra aprovar antes de gravar.",
		Detalhes: []string{
			"Hooks + bullet points (o que precisa ser dito / mostrado).",
			"Você aprova ou ajusta (sem burocracia).",
			"Depois disso, a gravação segue alinhada e eficiente.",
		},
	},
	{
		Num: "04", Titulo: "Produção & gravação", Prazo: "1–3 dias",
		Resumo: "Eu gravo com estética alinhada ao produto + tendência.",
		Detalhes: []string{
			"Gravação em luz natural/estúdio (o que combinar mais).",
			"Cenas de uso real + close do produto + prova social (quando faz sentido).",
			"Captação pensando em cortes pra ads e variações.",
		},
	},
	{
		Num: "05", Titulo: "Edição & entrega final", Prazo: "2–5 dias",
		Resumo: "Eu edito (dinâmico + legendas) e te entrego tudo organizado.",
		Detalhes: []string{
			"Entrega em alta + versão otimizada pra plataformas (Reels/TikTok).",
			"Organização por pasta e nomes fáceis (ex.: Hook1_V1, Hook1_V2...).",
			"1 rodada de ajustes leves incluída (ex.: texto/legenda/corte).",
		},
	},
}

// ── Métricas (Instagram / TikTok) ─────────────────────────────────────────────

type barra struct {
	Rotulo string
	Pct    float64
}

// fatia é um pedaço de um donut (gênero/idade); Cor é uma CSS var do tema.
type fatia struct {
	Rotulo string
	Pct    float64
	Cor    string // ex.: "var(--gold)"
}

type metricaRede struct {
	Rede       string
	Seguidores string
	Views      string
	Locais     []barra
	Genero     []fatia
	Idades     []fatia
	// Foto é o mockup de celular que flutua no card (WebP com alpha -- AVIF do
	// projeto não guarda transparência, e o celular precisa dos cantos e da sombra).
	// Lado: "dir" ou "esq" (alterna entre as redes, como no mídia kit original).
	Foto string
	Lado string
}

// cores dos donuts (mesma ordem das fatias). Hex FIXO, não token: o card de
// métrica é um painel marrom escuro fixo nos dois temas, então as fatias têm que
// ser claras e distintas independentemente do dark mode -- um token que inverte
// sumiria sobre o painel. Dourado / creme / verde claro, mutuamente contrastantes.
const (
	corA = "#d8b45a" // dourado
	corB = "#efe3d0" // creme
	corC = "#5f9184" // verde claro
)

var metricas = []metricaRede{
	{
		Rede: "Instagram", Seguidores: "+3.2K", Views: "+36.1K",
		Locais: []barra{{"Brasil", 96.7}, {"Curitiba", 29.5}},
		Genero: []fatia{{"Feminino", 83.9, corA}, {"Masculino", 16.1, corB}},
		Idades: []fatia{{"18–24", 36.9, corA}, {"25–34", 47.5, corB}, {"+35", 15.6, corC}},
		Foto:   "/static/img/midiakit/phone-instagram.webp", Lado: "dir",
	},
	{
		Rede: "TikTok", Seguidores: "+7K", Views: "+17K",
		Locais: []barra{{"Brasil", 99.4}},
		Genero: []fatia{{"Feminino", 78.0, corA}, {"Masculino", 21.0, corB}},
		Idades: []fatia{{"18–24", 54.2, corA}, {"25–34", 35.9, corB}, {"+35", 9.9, corC}},
		Foto:   "/static/img/midiakit/phone-tiktok.webp", Lado: "esq",
	},
}
