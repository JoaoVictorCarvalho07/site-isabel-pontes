export type Trecho = { texto: string; href?: string };

const LINK = /https?:\/\/[^\s<>"']+/g;

export function trechosComLinks(texto: string): Trecho[] {
  const trechos: Trecho[] = [];
  let fim = 0;
  for (const achado of texto.matchAll(LINK)) {
    const bruto = achado[0];
    const limpo = bruto.replace(/[.,;:!?)]+$/, "");
    if (achado.index > fim) trechos.push({ texto: texto.slice(fim, achado.index) });
    trechos.push({ texto: limpo, href: limpo });
    if (limpo.length < bruto.length) trechos.push({ texto: bruto.slice(limpo.length) });
    fim = achado.index + bruto.length;
  }
  if (fim < texto.length) trechos.push({ texto: texto.slice(fim) });
  return trechos;
}

export type Opcao = { pt: string; en: string };

export function opcoesDoTexto(texto: string, anteriores: Opcao[] = []): Opcao[] {
  const traducoes = new Map(anteriores.map((opcao) => [opcao.pt, opcao.en]));
  return texto
    .split("\n")
    .map((linha) => linha.trim())
    .filter(Boolean)
    .map((pt) => ({ pt, en: traducoes.get(pt) ?? "" }));
}

export function casar(ptAtual: string, enAtual: string, ptNovo: string): [pt: string, en: string] {
  return ptNovo === ptAtual ? [ptAtual, enAtual] : [ptNovo, ""];
}
