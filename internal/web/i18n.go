package web

import (
	"net/http"
	"strings"
)

// i18n — PT-BR (padrão, sem prefixo) e EN (prefixo /en). Ver docs §7.
//
// Duas mecânicas, escolhidas por caberem melhor a cada caso:
//   - `t(lang, chave)` para strings curtas de UI (nav, botões, rótulos).
//   - tipo `ml` (pt/en) nos campos de conteúdo estruturado (pacotes, processo),
//     resolvido no template por `.Campo.T $.Lang` -- um dado só, sem duplicar
//     dataset. Prosa longa (bio, Sobre) fica com {{ if eq $.Lang "en" }} no template.
//
// NOTA: a copy EN é um rascunho meu; a Isa deve revisar a voz de marca (§ do plano).

// ml é uma string bilíngue. Campos minúsculos + método T exportado: o template
// acessa via {{ .Campo.T $.Lang }}, e construir ml{"pt","en"} fica curto.
type ml struct{ pt, en string }

// T devolve a string no idioma pedido (fallback PT).
func (m ml) T(lang string) string {
	if lang == "en" {
		return m.en
	}
	return m.pt
}

// idioma resolve o idioma pela URL: /en e /en/... são EN; o resto é PT.
func idioma(r *http.Request) string {
	p := r.URL.Path
	if p == "/en" || strings.HasPrefix(p, "/en/") {
		return "en"
	}
	return "pt"
}

// enPath mapeia um caminho PT para o equivalente EN. "/" -> "/en"; "/x" -> "/en/x".
func enPath(pathPT string) string {
	if pathPT == "/" {
		return "/en"
	}
	return "/en" + pathPT
}

// traducoes: strings curtas de UI por chave. Falta de chave -> a própria chave
// aparece (visível, pega em teste), nunca "".
var traducoes = map[string]ml{
	"nav.galeria":   {"Galeria", "Gallery"},
	"nav.sobre":     {"Sobre", "About"},
	"nav.midiakit":  {"Mídia kit", "Media kit"},
	"nav.contato":   {"Contato", "Contact"},
	"a11y.pular":    {"Pular para o conteúdo", "Skip to content"},
	"a11y.abrir":    {"Abrir menu", "Open menu"},
	"a11y.fechar":   {"Fechar", "Close"},
	"a11y.menu":     {"Menu", "Menu"},
	"a11y.principal": {"Principal", "Main"},
	"a11y.tema":     {"Alternar tema claro/escuro", "Toggle light/dark theme"},
	"rodape":        {"Todos os direitos reservados", "All rights reserved"},
	"comum.todos":   {"Todos", "All"},

	// Formularios de briefing (as PERGUNTAS vivem em briefing_content.go; aqui
	// so o chrome: validacao, rotulos de UI).
	"form.obrigatorio":   {"Esse campo é obrigatório.", "This field is required."},
	"form.emailInvalido": {"Confere o e-mail, parece incompleto.", "Please check the email, it looks incomplete."},
	"form.opcaoInvalida": {"Escolha uma das opções da lista.", "Please pick one of the listed options."},
	"form.urlInvalida":   {"Cola o link completo, começando com https://", "Paste the full link, starting with https://"},
	"form.arquivoGrande": {"As imagens somaram mais que o limite. Manda até 4, de 5 MB cada.", "The images exceeded the limit. Send up to 4, 5 MB each."},
	"form.reanexar":      {"Suas imagens precisam ser anexadas de novo — o navegador não guarda arquivos entre envios.", "You'll need to attach your images again — browsers don't keep files between submissions."},
	"form.erroGeral":     {"Algo deu errado aqui do meu lado. Tenta de novo ou me chama no WhatsApp.", "Something went wrong on my end. Try again or message me on WhatsApp."},
	"form.muitosEnvios":  {"Você acabou de enviar um briefing. Espera um minutinho antes de mandar outro.", "You've just sent a briefing. Please wait a moment before sending another."},
	"form.opcional":      {"opcional", "optional"},
	"form.escolha":       {"Escolha uma opção", "Choose an option"},
	"form.enviando":      {"Enviando…", "Sending…"},
	"form.corrija":       {"Confere os campos marcados abaixo.", "Please check the highlighted fields below."},
	"form.abrirBriefing": {"Preencher briefing", "Fill out the briefing"},

	"galeria.vazia":   {"Nenhuma foto encontrada.", "No photos found."},
	"galeria.filtrar": {"Filtrar por tema", "Filter by theme"},

	"mk.verTrabalhos": {"Ver trabalhos", "See work"},
	"mk.investimento": {"Investimento", "Investment"},
	"mk.trabalhos":    {"Trabalhos", "Work"},
	"mk.curadoria":    {"Uma curadoria de vídeos pensados para diferentes marcas nesse mundo UGC / IGC.", "A curated selection of videos made for different brands in this UGC / IGC world."},
	"mk.metricas":     {"Métricas", "Metrics"},
	"mk.publico":      {"Público", "Audience"},
	"mk.seguidores":   {"seguidores", "followers"},
	"mk.visualizacoes": {"visualizações", "views"},
	"mk.comoFunciona": {"Como funciona a contratação (UGC)", "How the hiring works (UGC)"},
	"mk.entenda":      {"Entenda", "Learn more"},
	"mk.solicitar":    {"Solicitar", "Request"},
	"mk.maisPopular":  {"Mais popular", "Most popular"},
	"mk.verInstagram": {"Ver trabalho no Instagram", "See work on Instagram"},
	"mk.assistir":     {"Assistir vídeo", "Watch video"},
	"mk.previa":       {"Prévia do vídeo", "Video preview"},
	"mk.filtrarCat":   {"Filtrar por categoria", "Filter by category"},
	"mk.baixarPdf":     {"Baixar em PDF", "Download as PDF"},
	"mk.contatoTitulo": {"Vamos criar algo extraordinário?", "Shall we create something extraordinary?"},
	"mk.contatoSub":   {"Estou pronta para elevar a narrativa da sua marca.", "I'm ready to elevate your brand's narrative."},
	"mk.wppGeral":     {"Oii Isa! Vim pelo mídia kit e quero conversar sobre um trabalho ✨", "Hi Isa! I came from the media kit and I'd like to talk about a project ✨"},
	// mensagem pré-preenchida do WhatsApp nos pacotes; %s = nome, tag, resumo.
	"mk.wppPacote": {
		"Oii Isa! Tenho interesse no %s (%s) — %s. Pode me enviar detalhes, valores e como funciona o briefing? ^-^",
		"Hi Isa! I'm interested in %s (%s) — %s. Could you send me details, prices and how the briefing works? :)",
	},
}

// t é o func de template para strings de UI: {{ t $.Lang "nav.galeria" }}.
func t(lang, chave string) string {
	if m, ok := traducoes[chave]; ok {
		return m.T(lang)
	}
	return chave
}

// cond é um ternário de string para o template: {{ cond $en "yes" "sim" }}. Útil
// em atributos curtos (alt/aria) onde um bloco {{ if }} polui.
func cond(c bool, a, b string) string {
	if c {
		return a
	}
	return b
}

// temaLabelEN traduz os rótulos dos temas da galeria de fotos (slug do R2 -> EN).
// O slug é estável; o label PT vem do SQLite, o EN daqui.
var temaLabelEN = map[string]string{
	"Eventos":       "Events",
	"personalizado": "Custom",
	"epoca":         "Period",
	"artistico":     "Artistic",
	"casal":         "Couples",
	"conceitual":    "Conceptual",
	"outros":        "Others",
}

// nomeCategoriaEN traduz o nome exibido das categorias de vídeo (o slug é estável;
// só o rótulo muda). Categoria sem entrada aqui cai no nome PT.
var nomeCategoriaEN = map[string]string{
	"beleza":      "Beauty",
	"eventos":     "Events",
	"fitness":     "Fitness",
	"gastronomia": "Food",
	"moda":        "Fashion",
	"saude":       "Health",
	"parcerias":   "Partnerships",
	"diario":      "Diary of a photographer",
}
