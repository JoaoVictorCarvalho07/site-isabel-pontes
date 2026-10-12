import { banco } from "./banco";
import { embaralhar } from "./embaralhar";

export type Foto = {
  id: number;
  tema: string;
  w: number;
  h: number;
  grid: string;
  grid2x: string;
  full: string;
  lqip: string;
  alt: string;
  ordem: number | null;
};

export type Tema = { slug: string; label: string };

type Catalogo = {
  dia: string;
  temas: Tema[];
  todas: Foto[];
  porTema: Map<string, Foto[]>;
};

const memoria = globalThis as typeof globalThis & { catalogoDeFotos?: Catalogo };

function diaEmSaoPaulo(): string {
  return new Intl.DateTimeFormat("en-CA", { timeZone: "America/Sao_Paulo" }).format(new Date());
}

function montar(dia: string): Catalogo {
  const temas = banco()
    .prepare("SELECT slug, label FROM temas WHERE visible = 1 ORDER BY sort_order, label")
    .all() as Tema[];
  const fotos = banco()
    .prepare(
      `SELECT id, tema_slug AS tema, w, h, grid, grid2x, full, lqip, alt, sort_order AS ordem
         FROM photos
        WHERE published = 1 AND excluida_em IS NULL
        ORDER BY sort_order IS NULL, sort_order, tema_slug, id`,
    )
    .all() as Foto[];
  const fixas = fotos.filter((foto) => foto.ordem !== null);
  const soltas = fotos.filter((foto) => foto.ordem === null);
  const todas = [...fixas, ...embaralhar(soltas, Number(dia.replaceAll("-", "")))];
  const porTema = new Map<string, Foto[]>();
  for (const foto of todas) {
    const lista = porTema.get(foto.tema);
    if (lista) lista.push(foto);
    else porTema.set(foto.tema, [foto]);
  }
  return { dia, temas, todas, porTema };
}

export function catalogo(): Catalogo {
  const dia = diaEmSaoPaulo();
  if (!memoria.catalogoDeFotos || memoria.catalogoDeFotos.dia !== dia) memoria.catalogoDeFotos = montar(dia);
  return memoria.catalogoDeFotos;
}

export function invalidarCatalogo(): void {
  memoria.catalogoDeFotos = undefined;
}

export function fotosDoTema(slug: string): Foto[] {
  const { todas, porTema } = catalogo();
  return slug ? (porTema.get(slug) ?? []) : todas;
}

export function temaValido(slug: string): boolean {
  return catalogo().temas.some((tema) => tema.slug === slug);
}
