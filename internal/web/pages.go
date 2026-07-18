package web

import (
	"net/http"
	"net/url"
	"strings"
)

// urlEncode imita o encodeURIComponent do JS: espaco vira %20 (nao '+'), como o
// wa.me e o mailto esperam. QueryEscape sozinho usaria '+'.
func urlEncode(s string) string {
	return strings.ReplaceAll(url.QueryEscape(s), "+", "%20")
}

// ── Home ──────────────────────────────────────────────────────────────────────

type homeView struct {
	pageData
	// FotoDoDia e a primeira foto do shuffle do dia -- estavel dentro do dia,
	// entao cacheavel, e "muda" todo dia sozinha. No React o card usava uma foto
	// aleatoria por load (nao cacheavel).
	FotoDoDia string
}

func (s *Server) home(w http.ResponseWriter, r *http.Request) {
	d := s.novaPagina(
		"Isabel Pontes — Fotografia artística e direção criativa em Curitiba",
		"Ensaios conceituais, fotografia artística e direção criativa em Curitiba.",
		"/",
	)
	// Hero escuro -> logo/nav brancos nos dois temas. No React era useState +
	// matchMedia + dois useEffect, com flash de logo errada ate hidratar; aqui o
	// servidor ja decide.
	d.BodyClass = "hero-escuro"

	foto := ""
	if fs := s.snap.Fotos(""); len(fs) > 0 {
		foto = fs[0].Full
	}
	s.render.Page(w, r, "home", http.StatusOK, homeView{pageData: d, FotoDoDia: foto})
}

// ── Sobre ─────────────────────────────────────────────────────────────────────

type sobreImg struct {
	Base string // '/static/img/sobre/SobreMim1'
	W, H int
}

type sobreView struct {
	pageData
	Carrossel []sobreImg
}

// dimensoes reais das variantes (max 800px); width/height contra layout shift.
var sobreCarrossel = []sobreImg{
	{"/static/img/sobre/SobreMim1", 800, 1066},
	{"/static/img/sobre/SobreMim2", 800, 600},
	{"/static/img/sobre/SobreMim3", 800, 450},
	{"/static/img/sobre/SobreMim4", 800, 1200},
	{"/static/img/sobre/SobreMim5", 800, 1200},
	{"/static/img/sobre/SobreMim6", 433, 728},
	{"/static/img/sobre/SobreMim7", 361, 560},
	{"/static/img/sobre/SobreMim8", 706, 1280},
}

func (s *Server) sobre(w http.ResponseWriter, r *http.Request) {
	d := s.novaPagina(
		"Sobre — Isabel Pontes",
		"Fotógrafa artística e criadora visual em Curitiba. Conheça a trajetória de Isabel Pontes.",
		"/sobre",
	)
	s.render.Page(w, r, "sobre", http.StatusOK, sobreView{pageData: d, Carrossel: sobreCarrossel})
}

// ── Contato ───────────────────────────────────────────────────────────────────

type contatoView struct {
	pageData
	WhatsAppHref  string
	InstagramHref string
	Email         string
	MailtoHref    string
	Cidade        string
}

func (s *Server) contato(w http.ResponseWriter, r *http.Request) {
	d := s.novaPagina(
		"Contato — Isabel Pontes",
		"Fale com Isabel Pontes por WhatsApp, Instagram ou e-mail. Base em Curitiba.",
		"/contato",
	)
	// Mesmos dados do React. NOTA: o numero 554191977011 tem 12 digitos; celular
	// brasileiro tem 13 no E.164 (falta um digito) -- provavelmente o wa.me nao
	// resolve. Portado como esta; confirmar com a Isabel.
	s.render.Page(w, r, "contato", http.StatusOK, contatoView{
		pageData: d,
		WhatsAppHref: "https://wa.me/554191977011?text=" +
			urlEncode("Oi! Vim pelo seu site ✨\n\nQuero falar sobre: (ensaio / evento / UGC)\nData/cidade: \nReferências (se tiver): \n"),
		InstagramHref: "https://instagram.com/isapontesfoto",
		Email:         "isabelpontesfotografia@gmail.com",
		MailtoHref: "mailto:isabelpontesfotografia@gmail.com?subject=" +
			urlEncode("Contato pelo site") + "&body=" +
			urlEncode("Oi! Vim pelo seu site ✨\n\nQuero falar sobre:\n"),
		Cidade: "Curitiba, PR",
	})
}
