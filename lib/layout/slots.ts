export type Formato = "avif" | "webp" | "jpg";

export type Slot = {
  rotulo: string;
  dica: string;
  larguras: number[];
  formatos: Formato[];
  sufixoDeLargura: boolean;
  padrao: { base: string; w: number; h: number };
};

export const SLOTS = {
  "hero-retrato": {
    rotulo: "Capa da home no celular",
    dica: "Foto em pé. Aparece na abertura do site em telas estreitas.",
    larguras: [640, 960, 1280, 1920],
    formatos: ["avif", "webp"],
    sufixoDeLargura: true,
    padrao: { base: "/static/img/hero/hero", w: 4000, h: 6000 },
  },
  "hero-paisagem": {
    rotulo: "Capa da home no computador",
    dica: "Foto deitada. Aparece na abertura do site em telas largas.",
    larguras: [960, 1280, 1920, 2560],
    formatos: ["avif", "webp"],
    sufixoDeLargura: true,
    padrao: { base: "/static/img/hero/hero-land", w: 2560, h: 1440 },
  },
  "card-midiakit": {
    rotulo: "Cartão do mídia kit na home",
    dica: "Foto em pé, ao lado do cartão da galeria.",
    larguras: [1080],
    formatos: ["jpg"],
    sufixoDeLargura: false,
    padrao: { base: "/static/img/MidiaKit", w: 1080, h: 1900 },
  },
  "sobre-bio-1": {
    rotulo: "Sobre: primeira foto do texto",
    dica: "Fica à esquerda do primeiro parágrafo.",
    larguras: [800],
    formatos: ["avif", "webp"],
    sufixoDeLargura: false,
    padrao: { base: "/static/img/sobre/SobreMim4", w: 800, h: 1200 },
  },
  "sobre-bio-2": {
    rotulo: "Sobre: segunda foto do texto",
    dica: "Fica à direita do segundo parágrafo.",
    larguras: [800],
    formatos: ["avif", "webp"],
    sufixoDeLargura: false,
    padrao: { base: "/static/img/sobre/SobreMim6", w: 433, h: 728 },
  },
} as const satisfies Record<string, Slot>;

export type NomeSlot = keyof typeof SLOTS;

export const CARROSSEL: Pick<Slot, "larguras" | "formatos" | "sufixoDeLargura"> = { larguras: [800], formatos: ["avif", "webp"], sufixoDeLargura: false };

export function ehSlot(nome: string): nome is NomeSlot {
  return Object.hasOwn(SLOTS, nome);
}

export function arquivo(base: string, slot: Pick<Slot, "sufixoDeLargura">, largura: number, formato: Formato): string {
  return slot.sufixoDeLargura ? `${base}-${largura}.${formato}` : `${base}.${formato}`;
}
