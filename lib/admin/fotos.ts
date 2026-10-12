import { randomBytes } from "node:crypto";
import { apagar, gravar, objetoDaUrl, urlPublica } from "../armazenamento";
import { banco } from "../banco";
import { farejarTipo } from "../briefing/tipos";
import { invalidarCatalogo } from "../fotos";
import { processarImagem } from "./imagem";
import { slugificar } from "./slug";

export const TAMANHO_MAXIMO_FOTO = 40 * 1024 * 1024;

export type FotoAdmin = {
  id: number;
  tema: string;
  w: number;
  h: number;
  grid: string;
  full: string;
  lqip: string;
  alt: string;
  publicada: boolean;
  ordem: number | null;
  naLixeira: boolean;
};

export type TemaAdmin = { slug: string; label: string; ordem: number; visivel: boolean; fotos: number };

type Linha = Omit<FotoAdmin, "publicada" | "naLixeira"> & { published: number; excluida_em: string | null };

const COLUNAS = "id, tema_slug AS tema, w, h, grid, full, lqip, alt, published, sort_order AS ordem, excluida_em";

function paraFoto({ published, excluida_em, ...resto }: Linha): FotoAdmin {
  return { ...resto, publicada: published === 1, naLixeira: excluida_em !== null };
}

export function listarFotos(): FotoAdmin[] {
  const linhas = banco().prepare(`SELECT ${COLUNAS} FROM photos ORDER BY sort_order IS NULL, sort_order, id DESC`).all() as Linha[];
  return linhas.map(paraFoto);
}

function buscarFoto(id: number): FotoAdmin {
  const linha = banco().prepare(`SELECT ${COLUNAS} FROM photos WHERE id = ?`).get(id) as Linha | undefined;
  if (!linha) throw new Error(`foto ${id} nao existe`);
  return paraFoto(linha);
}

export function listarTemas(): TemaAdmin[] {
  return (
    banco()
      .prepare(
        `SELECT t.slug, t.label, t.sort_order AS ordem, t.visible AS visivel,
                (SELECT count(*) FROM photos p WHERE p.tema_slug = t.slug AND p.excluida_em IS NULL) AS fotos
           FROM temas t ORDER BY t.sort_order, t.label`,
      )
      .all() as (Omit<TemaAdmin, "visivel"> & { visivel: number })[]
  ).map((tema) => ({ ...tema, visivel: tema.visivel === 1 }));
}

function temaExiste(slug: string): boolean {
  return banco().prepare("SELECT 1 FROM temas WHERE slug = ?").get(slug) !== undefined;
}

export class FotoRecusada extends Error {}

export async function adicionarFoto(nomeEnviado: string, bytes: Uint8Array, tema: string): Promise<FotoAdmin> {
  if (!temaExiste(tema)) throw new FotoRecusada("Escolha um tema da lista.");
  if (bytes.length === 0 || bytes.length > TAMANHO_MAXIMO_FOTO) throw new FotoRecusada("A foto passa do limite de 40 MB.");
  const tipo = farejarTipo(bytes);
  if (!tipo) throw new FotoRecusada("Esse arquivo não é uma imagem que eu consiga abrir. Envie em JPG, PNG ou WebP.");

  let imagem;
  try {
    imagem = await processarImagem(bytes);
  } catch (erro) {
    console.warn("admin: imagem ilegivel", { nomeEnviado, erro });
    throw new FotoRecusada("Não consegui ler essa imagem. Envie em JPG, PNG ou WebP.");
  }

  const nome = `${Date.now().toString(36)}-${randomBytes(4).toString("hex")}`;
  const original = `originais/${tema}/${nome}${tipo.extensao}`;
  const objetos = Object.fromEntries(imagem.variantes.map(({ pasta }) => [pasta, `variants/${pasta}/${tema}/${nome}.avif`]));

  await gravar("originais", original, bytes, tipo.mime);
  for (const { pasta, bytes: variante } of imagem.variantes) await gravar("publico", objetos[pasta], variante, "image/avif");

  const gravada = banco()
    .prepare(
      `INSERT INTO photos (key, tema_slug, w, h, grid, grid2x, full, lqip, original)
       VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?)`,
    )
    .run(
      `${tema}/${nome}${tipo.extensao}`,
      tema,
      imagem.w,
      imagem.h,
      urlPublica(objetos.grid400),
      urlPublica(objetos.grid800),
      urlPublica(objetos.full1600),
      imagem.lqip,
      original,
    );
  invalidarCatalogo();
  return buscarFoto(Number(gravada.lastInsertRowid));
}

function alterar(sql: string, ...valores: unknown[]): void {
  banco().prepare(sql).run(...valores);
  invalidarCatalogo();
}

export function editarFoto(id: number, dados: { tema?: string; alt?: string }): void {
  if (dados.tema !== undefined) {
    if (!temaExiste(dados.tema)) throw new Error("tema inexistente");
    alterar("UPDATE photos SET tema_slug = ? WHERE id = ?", dados.tema, id);
  }
  if (dados.alt !== undefined) alterar("UPDATE photos SET alt = ? WHERE id = ?", dados.alt.trim().slice(0, 300), id);
}

export function publicarFoto(id: number, publicada: boolean): void {
  alterar("UPDATE photos SET published = ? WHERE id = ?", publicada ? 1 : 0, id);
}

export function fixarFoto(id: number, fixada: boolean): void {
  if (!fixada) return alterar("UPDATE photos SET sort_order = NULL WHERE id = ?", id);
  const { proxima } = banco().prepare("SELECT coalesce(max(sort_order), 0) + 1 AS proxima FROM photos").get() as { proxima: number };
  alterar("UPDATE photos SET sort_order = ? WHERE id = ? AND sort_order IS NULL", proxima, id);
}

export function moverFixada(id: number, direcao: "antes" | "depois"): void {
  const fixadas = banco().prepare("SELECT id, sort_order AS ordem FROM photos WHERE sort_order IS NOT NULL ORDER BY sort_order, id").all() as {
    id: number;
    ordem: number;
  }[];
  const posicao = fixadas.findIndex((foto) => foto.id === id);
  const vizinha = fixadas[posicao + (direcao === "antes" ? -1 : 1)];
  if (posicao < 0 || !vizinha) return;
  const reordenadas = [...fixadas];
  reordenadas[posicao] = vizinha;
  reordenadas[fixadas.indexOf(vizinha)] = fixadas[posicao];
  banco().transaction(() => {
    reordenadas.forEach((foto, indice) => banco().prepare("UPDATE photos SET sort_order = ? WHERE id = ?").run(indice + 1, foto.id));
  })();
  invalidarCatalogo();
}

export function moverParaLixeira(id: number): void {
  alterar("UPDATE photos SET excluida_em = datetime('now'), sort_order = NULL WHERE id = ?", id);
}

export function restaurarFoto(id: number): void {
  alterar("UPDATE photos SET excluida_em = NULL WHERE id = ?", id);
}

export async function apagarDeVez(id: number): Promise<void> {
  const linha = banco().prepare("SELECT grid, grid2x, full, original, excluida_em FROM photos WHERE id = ?").get(id) as
    | { grid: string; grid2x: string; full: string; original: string | null; excluida_em: string | null }
    | undefined;
  if (!linha || linha.excluida_em === null) throw new Error("so da para apagar o que esta na lixeira");
  for (const url of [linha.grid, linha.grid2x, linha.full]) {
    const objeto = objetoDaUrl(url);
    if (objeto) await apagar("publico", objeto);
  }
  if (linha.original) await apagar("originais", linha.original);
  alterar("DELETE FROM photos WHERE id = ?", id);
}

export function criarTema(label: string): string {
  const limpo = label.trim().slice(0, 40);
  const slug = slugificar(limpo);
  if (!limpo || !slug) throw new Error("nome de tema vazio");
  if (temaExiste(slug)) throw new Error("ja existe um tema com esse nome");
  const { proxima } = banco().prepare("SELECT coalesce(max(sort_order), 0) + 10 AS proxima FROM temas").get() as { proxima: number };
  alterar("INSERT INTO temas (slug, label, sort_order) VALUES (?, ?, ?)", slug, limpo, proxima);
  return slug;
}

export function renomearTema(slug: string, label: string): void {
  const limpo = label.trim().slice(0, 40);
  if (!limpo) throw new Error("nome de tema vazio");
  alterar("UPDATE temas SET label = ? WHERE slug = ?", limpo, slug);
}

export function mostrarTema(slug: string, visivel: boolean): void {
  alterar("UPDATE temas SET visible = ? WHERE slug = ?", visivel ? 1 : 0, slug);
}

export function moverTema(slug: string, direcao: "antes" | "depois"): void {
  const temas = banco().prepare("SELECT slug FROM temas ORDER BY sort_order, label").all() as { slug: string }[];
  const posicao = temas.findIndex((tema) => tema.slug === slug);
  const destino = posicao + (direcao === "antes" ? -1 : 1);
  if (posicao < 0 || destino < 0 || destino >= temas.length) return;
  [temas[posicao], temas[destino]] = [temas[destino], temas[posicao]];
  banco().transaction(() => {
    temas.forEach((tema, indice) => banco().prepare("UPDATE temas SET sort_order = ? WHERE slug = ?").run((indice + 1) * 10, tema.slug));
  })();
  invalidarCatalogo();
}
