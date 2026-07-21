package web

// GERADO por scripts/gen_video_posters.py a partir de videos-index.json +
// labels.json do midia kit estatico. NAO edite a mao -- rode o script de novo.
//
// Videos ficam no R2 (§6 do plano); posters AVIF em static/img/midiakit/videos/.
// O portfolio (#portfolio) mostra o poster e so carrega o .webm no clique --
// os 26 videos somam 322 MB, autoplayar tudo seria inviavel.

type videoUGC struct {
	Titulo string
	Src    string // .webm no R2
	Poster string // base do poster local (sem ext); .avif + .webp
}

type categoriaVideos struct {
	Slug   string // filtro (ex.: "beleza")
	Nome   string // exibido
	Videos []videoUGC
}

var portfolioVideos = []categoriaVideos{
	{
		Slug: "beleza", Nome: "Beleza",
		Videos: []videoUGC{
			{Titulo: "L'Oréal Paris", Src: "https://pub-2584b081d14449de91b2354653b1d10f.r2.dev/videos-ugc/beleza/Loreal.webm", Poster: "/static/img/midiakit/videos/beleza-loreal"},
			{Titulo: "Dove", Src: "https://pub-2584b081d14449de91b2354653b1d10f.r2.dev/videos-ugc/beleza/Dove.webm", Poster: "/static/img/midiakit/videos/beleza-dove"},
			{Titulo: "Pantene", Src: "https://pub-2584b081d14449de91b2354653b1d10f.r2.dev/videos-ugc/beleza/Pantene.webm", Poster: "/static/img/midiakit/videos/beleza-pantene"},
			{Titulo: "Maquiagem", Src: "https://pub-2584b081d14449de91b2354653b1d10f.r2.dev/videos-ugc/beleza/Maquiagem.webm", Poster: "/static/img/midiakit/videos/beleza-maquiagem"},
			{Titulo: "Elseve", Src: "https://pub-2584b081d14449de91b2354653b1d10f.r2.dev/videos-ugc/beleza/Elseve.webm", Poster: "/static/img/midiakit/videos/beleza-elseve"},
		},
	},
	{
		Slug: "eventos", Nome: "Eventos",
		Videos: []videoUGC{
			{Titulo: "Natura Yoga", Src: "https://pub-2584b081d14449de91b2354653b1d10f.r2.dev/videos-ugc/eventos/Natura_Yoga.webm", Poster: "/static/img/midiakit/videos/eventos-natura-yoga"},
			{Titulo: "Mercado Livre", Src: "https://pub-2584b081d14449de91b2354653b1d10f.r2.dev/videos-ugc/eventos/Mercado_Livre.webm", Poster: "/static/img/midiakit/videos/eventos-mercado-livre"},
			{Titulo: "Top Girls Experience", Src: "https://pub-2584b081d14449de91b2354653b1d10f.r2.dev/videos-ugc/eventos/Top_Girls_Experience.webm", Poster: "/static/img/midiakit/videos/eventos-top-girls-experience"},
			{Titulo: "Roda das Deusas", Src: "https://pub-2584b081d14449de91b2354653b1d10f.r2.dev/videos-ugc/eventos/Roda_Das_Deusas.webm", Poster: "/static/img/midiakit/videos/eventos-roda-das-deusas"},
		},
	},
	{
		Slug: "fitness", Nome: "Fitness",
		Videos: []videoUGC{
			{Titulo: "Probiótica", Src: "https://pub-2584b081d14449de91b2354653b1d10f.r2.dev/videos-ugc/fitness/Probiotica.webm", Poster: "/static/img/midiakit/videos/fitness-probiotica"},
			{Titulo: "Crossfit", Src: "https://pub-2584b081d14449de91b2354653b1d10f.r2.dev/videos-ugc/fitness/Crossfit.webm", Poster: "/static/img/midiakit/videos/fitness-crossfit"},
		},
	},
	{
		Slug: "gastronomia", Nome: "Gastronomia",
		Videos: []videoUGC{
			{Titulo: "Café Onírico", Src: "https://pub-2584b081d14449de91b2354653b1d10f.r2.dev/videos-ugc/gastronomia/Cafe_Onirico.webm", Poster: "/static/img/midiakit/videos/gastronomia-cafe-onirico"},
			{Titulo: "Doces Bliss", Src: "https://pub-2584b081d14449de91b2354653b1d10f.r2.dev/videos-ugc/gastronomia/Doces_Bliss.webm", Poster: "/static/img/midiakit/videos/gastronomia-doces-bliss"},
		},
	},
	{
		Slug: "moda", Nome: "Moda",
		Videos: []videoUGC{
			{Titulo: "Pé de Algodão", Src: "https://pub-2584b081d14449de91b2354653b1d10f.r2.dev/videos-ugc/moda/Pe_De_Algodao.webm", Poster: "/static/img/midiakit/videos/moda-pe-de-algodao"},
			{Titulo: "Pri Joias", Src: "https://pub-2584b081d14449de91b2354653b1d10f.r2.dev/videos-ugc/moda/Pri_Joias.webm", Poster: "/static/img/midiakit/videos/moda-pri-joias"},
			{Titulo: "Atelie Cris Tavares", Src: "https://pub-2584b081d14449de91b2354653b1d10f.r2.dev/videos-ugc/moda/Atelie_Cris_Tavares.webm", Poster: "/static/img/midiakit/videos/moda-atelie-cris-tavares"},
		},
	},
	{
		Slug: "saude", Nome: "Saúde",
		Videos: []videoUGC{
			{Titulo: "Redoxon", Src: "https://pub-2584b081d14449de91b2354653b1d10f.r2.dev/videos-ugc/saude/Redoxon.webm", Poster: "/static/img/midiakit/videos/saude-redoxon"},
			{Titulo: "Gino-canisten", Src: "https://pub-2584b081d14449de91b2354653b1d10f.r2.dev/videos-ugc/saude/Ginocanisten.webm", Poster: "/static/img/midiakit/videos/saude-ginocanisten"},
		},
	},
	{
		Slug: "parcerias", Nome: "Parcerias",
		Videos: []videoUGC{
			{Titulo: "Comu Delas", Src: "https://pub-2584b081d14449de91b2354653b1d10f.r2.dev/videos-ugc/parcerias/Comu_Delas.webm", Poster: "/static/img/midiakit/videos/parcerias-comu-delas"},
		},
	},
	{
		Slug: "diario", Nome: "Diário de uma fotógrafa",
		Videos: []videoUGC{
			{Titulo: "Isaresponde", Src: "https://pub-2584b081d14449de91b2354653b1d10f.r2.dev/videos-ugc/diario%20de%20uma%20fotografa/Isaresponde.webm", Poster: "/static/img/midiakit/videos/diario-isaresponde"},
			{Titulo: "Bate Papo", Src: "https://pub-2584b081d14449de91b2354653b1d10f.r2.dev/videos-ugc/diario%20de%20uma%20fotografa/Bate_Papo.webm", Poster: "/static/img/midiakit/videos/diario-bate-papo"},
			{Titulo: "Musa Day Afrodite Lançamento", Src: "https://pub-2584b081d14449de91b2354653b1d10f.r2.dev/videos-ugc/diario%20de%20uma%20fotografa/Musa%20Day%20Afrodite%20Lan%C3%A7amento.webm", Poster: "/static/img/midiakit/videos/diario-musa-day-afrodite-lancamento"},
			{Titulo: "Musaday", Src: "https://pub-2584b081d14449de91b2354653b1d10f.r2.dev/videos-ugc/diario%20de%20uma%20fotografa/Musaday.webm", Poster: "/static/img/midiakit/videos/diario-musaday"},
			{Titulo: "Projeto Metamorfose", Src: "https://pub-2584b081d14449de91b2354653b1d10f.r2.dev/videos-ugc/diario%20de%20uma%20fotografa/Projeto_Metamorfose.webm", Poster: "/static/img/midiakit/videos/diario-projeto-metamorfose"},
			{Titulo: "Organizando acessorios", Src: "https://pub-2584b081d14449de91b2354653b1d10f.r2.dev/videos-ugc/diario%20de%20uma%20fotografa/Organizando_acessorios.webm", Poster: "/static/img/midiakit/videos/diario-organizando-acessorios"},
			{Titulo: "Dicas de Camera", Src: "https://pub-2584b081d14449de91b2354653b1d10f.r2.dev/videos-ugc/diario%20de%20uma%20fotografa/Dicas_de_Camera.webm", Poster: "/static/img/midiakit/videos/diario-dicas-de-camera"},
		},
	},
}
