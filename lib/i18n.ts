export type Idioma = "pt" | "en";

type Par = readonly [pt: string, en: string];

const TRADUCOES: Record<string, Par> = {
  "nav.galeria": ["Galeria", "Gallery"],
  "nav.sobre": ["Sobre", "About"],
  "nav.midiakit": ["Mídia kit", "Media kit"],
  "nav.contato": ["Contato", "Contact"],
  "a11y.pular": ["Pular para o conteúdo", "Skip to content"],
  "a11y.abrir": ["Abrir menu", "Open menu"],
  "a11y.fechar": ["Fechar", "Close"],
  "a11y.menu": ["Menu", "Menu"],
  "a11y.principal": ["Principal", "Main"],
  "a11y.tema": ["Alternar tema claro/escuro", "Toggle light/dark theme"],
  "rodape": ["Todos os direitos reservados", "All rights reserved"],
  "comum.todos": ["Todos", "All"],
  "form.obrigatorio": ["Esse campo é obrigatório.", "This field is required."],
  "form.emailInvalido": ["Confere o e-mail, parece incompleto.", "Please check the email, it looks incomplete."],
  "form.opcaoInvalida": ["Escolha uma das opções da lista.", "Please pick one of the listed options."],
  "form.contatoInvalido": ["Confere esse contato: um WhatsApp com DDD (41 99999-8888) ou um e-mail.", "Please check this: a phone number with area code, or an email address."],
  "form.arquivoGrande": ["As imagens somaram mais que o limite. Manda até 4, de 5 MB cada.", "The images exceeded the limit. Send up to 4, 5 MB each."],
  "form.reanexar": ["Suas imagens precisam ser anexadas de novo — o navegador não guarda arquivos entre envios.", "You'll need to attach your images again — browsers don't keep files between submissions."],
  "form.erroGeral": ["Algo deu errado aqui do meu lado. Tenta de novo ou me chama no WhatsApp.", "Something went wrong on my end. Try again or message me on WhatsApp."],
  "form.muitosEnvios": ["Você acabou de enviar um briefing. Espera um minutinho antes de mandar outro.", "You've just sent a briefing. Please wait a moment before sending another."],
  "form.opcional": ["opcional", "optional"],
  "form.escolha": ["Escolha uma opção", "Choose an option"],
  "form.enviando": ["Enviando…", "Sending…"],
  "form.corrija": ["Confere os campos marcados abaixo.", "Please check the highlighted fields below."],
  "form.abrirBriefing": ["Preencher briefing", "Fill out the briefing"],
  "galeria.vazia": ["Nenhuma foto encontrada.", "No photos found."],
  "galeria.filtrar": ["Filtrar por tema", "Filter by theme"],
  "mk.verTrabalhos": ["Ver trabalhos", "See work"],
  "mk.investimento": ["Investimento", "Investment"],
  "mk.trabalhos": ["Trabalhos", "Work"],
  "mk.curadoria": ["Uma curadoria de vídeos pensados para diferentes marcas nesse mundo UGC / IGC.", "A curated selection of videos made for different brands in this UGC / IGC world."],
  "mk.metricas": ["Métricas", "Metrics"],
  "mk.publico": ["Público", "Audience"],
  "mk.seguidores": ["seguidores", "followers"],
  "mk.visualizacoes": ["visualizações", "views"],
  "mk.comoFunciona": ["Como funciona a contratação (UGC)", "How the hiring works (UGC)"],
  "mk.entenda": ["Entenda", "Learn more"],
  "mk.solicitar": ["Solicitar", "Request"],
  "mk.maisPopular": ["Mais popular", "Most popular"],
  "mk.verInstagram": ["Ver trabalho no Instagram", "See work on Instagram"],
  "mk.assistir": ["Assistir vídeo", "Watch video"],
  "mk.previa": ["Prévia do vídeo", "Video preview"],
  "mk.filtrarCat": ["Filtrar por categoria", "Filter by category"],
  "mk.baixarPdf": ["Baixar em PDF", "Download as PDF"],
  "mk.contatoTitulo": ["Vamos criar algo extraordinário?", "Shall we create something extraordinary?"],
  "mk.contatoSub": ["Estou pronta para elevar a narrativa da sua marca.", "I'm ready to elevate your brand's narrative."],
  "mk.wppGeral": ["Oii Isa! Vim pelo mídia kit e quero conversar sobre um trabalho ✨", "Hi Isa! I came from the media kit and I'd like to talk about a project ✨"],
  "mk.wppPacote": ["Oii Isa! Tenho interesse no %s (%s) — %s. Pode me enviar detalhes, valores e como funciona o briefing? ^-^", "Hi Isa! I'm interested in %s (%s) — %s. Could you send me details, prices and how the briefing works? :)"],
};

export function t(idioma: Idioma, chave: string): string {
  const par = TRADUCOES[chave];
  if (!par) return chave;
  return idioma === "en" && par[1] ? par[1] : par[0];
}

export function resolver(idioma: Idioma, par: readonly [pt: string, en: string]): string {
  return idioma === "en" && par[1] ? par[1] : par[0];
}

export function prefixo(idioma: Idioma): string {
  return idioma === "en" ? "/en" : "";
}

export function caminho(idioma: Idioma, caminhoPT: string): string {
  if (idioma === "pt") return caminhoPT;
  return caminhoPT === "/" ? "/en" : `/en${caminhoPT}`;
}

export const TEMAS_EN: Record<string, string> = {
  Eventos: "Events",
  personalizado: "Custom",
  epoca: "Period",
  artistico: "Artistic",
  casal: "Couples",
  conceitual: "Conceptual",
  outros: "Others",
};

export const CATEGORIAS_EN: Record<string, string> = {
  beleza: "Beauty",
  eventos: "Events",
  fitness: "Fitness",
  gastronomia: "Food",
  moda: "Fashion",
  saude: "Health",
  parcerias: "Partnerships",
  diario: "Diary of a photographer",
};
