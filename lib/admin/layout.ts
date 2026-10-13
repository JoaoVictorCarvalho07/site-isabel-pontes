import { randomBytes } from "node:crypto";
import sharp from "sharp";
import { apagar, gravar, objetoDaUrl, urlPublica } from "../armazenamento";
import { banco } from "../banco";
import { farejarTipo } from "../briefing/tipos";
import { arquivo, CARROSSEL, SLOTS, type Formato, type NomeSlot, type Slot } from "../layout/slots";
import { FotoRecusada, TAMANHO_MAXIMO_FOTO } from "./fotos";

const QUALIDADE: Record<Formato, number> = { avif: 55, webp: 78, jpg: 82 };
const MIME: Record<Formato, string> = { avif: "image/avif", webp: "image/webp", jpg: "image/jpeg" };

type Receita = Pick<Slot, "larguras" | "formatos" | "sufixoDeLargura">;

async function gerar(bytes: Uint8Array, pasta: string, receita: Receita): Promise<{ base: string; w: number; h: number }> {
  if (bytes.length === 0 || bytes.length > TAMANHO_MAXIMO_FOTO) throw new FotoRecusada("A foto passa do limite de 40 MB.");
  if (!farejarTipo(bytes)) throw new FotoRecusada("Esse arquivo não é uma imagem que eu consiga abrir. Envie em JPG, PNG ou WebP.");
  let dimensoes;
  try {
    dimensoes = await sharp(bytes, { failOn: "error" }).metadata();
  } catch {
    throw new FotoRecusada("Não consegui ler essa imagem. Envie em JPG, PNG ou WebP.");
  }
  const { width = 0, height = 0, orientation = 1 } = dimensoes;
  const [w, h] = orientation >= 5 ? [height, width] : [width, height];
  const objeto = `layout/${pasta}/${Date.now().toString(36)}-${randomBytes(4).toString("hex")}`;
  for (const largura of receita.larguras) {
    for (const formato of receita.formatos) {
      const redimensionada = sharp(bytes).rotate().resize({ width: largura, withoutEnlargement: true });
      const saida =
        formato === "avif"
          ? redimensionada.avif({ quality: QUALIDADE.avif, effort: 4 })
          : formato === "webp"
            ? redimensionada.webp({ quality: QUALIDADE.webp })
            : redimensionada.jpeg({ quality: QUALIDADE.jpg, mozjpeg: true });
      await gravar("publico", arquivo(objeto, receita, largura, formato), await saida.toBuffer(), MIME[formato]);
    }
  }
  const unica = !receita.sufixoDeLargura && w > receita.larguras[0];
  const escala = unica ? receita.larguras[0] / w : 1;
  return { base: urlPublica(objeto), w: Math.round(w * escala), h: Math.round(h * escala) };
}

async function descartar(base: string, receita: Receita): Promise<void> {
  const objeto = objetoDaUrl(base);
  if (!objeto) return;
  for (const largura of receita.larguras) {
    for (const formato of receita.formatos) await apagar("publico", arquivo(objeto, receita, largura, formato));
  }
}

function baseAtual(slot: NomeSlot): string | undefined {
  return (banco().prepare("SELECT base FROM layout_imagens WHERE slot = ?").get(slot) as { base: string } | undefined)?.base;
}

export async function trocarImagem(slot: NomeSlot, bytes: Uint8Array): Promise<void> {
  const nova = await gerar(bytes, slot, SLOTS[slot]);
  const anterior = baseAtual(slot);
  banco()
    .prepare(
      `INSERT INTO layout_imagens (slot, base, w, h) VALUES (?, ?, ?, ?)
       ON CONFLICT(slot) DO UPDATE SET base = excluded.base, w = excluded.w, h = excluded.h, atualizado_em = datetime('now')`,
    )
    .run(slot, nova.base, nova.w, nova.h);
  if (anterior) await descartar(anterior, SLOTS[slot]);
}

export async function restaurarImagem(slot: NomeSlot): Promise<void> {
  const anterior = baseAtual(slot);
  banco().prepare("DELETE FROM layout_imagens WHERE slot = ?").run(slot);
  if (anterior) await descartar(anterior, SLOTS[slot]);
}

export async function adicionarAoCarrossel(bytes: Uint8Array): Promise<void> {
  const nova = await gerar(bytes, "sobre", CARROSSEL);
  const { proxima } = banco().prepare("SELECT coalesce(max(sort_order), 0) + 10 AS proxima FROM sobre_carrossel").get() as { proxima: number };
  banco().prepare("INSERT INTO sobre_carrossel (base, w, h, sort_order) VALUES (?, ?, ?, ?)").run(nova.base, nova.w, nova.h, proxima);
}

export async function removerDoCarrossel(id: number): Promise<void> {
  const linha = banco().prepare("SELECT base FROM sobre_carrossel WHERE id = ?").get(id) as { base: string } | undefined;
  if (!linha) return;
  if ((banco().prepare("SELECT count(*) AS n FROM sobre_carrossel").get() as { n: number }).n <= 1) {
    throw new FotoRecusada("O carrossel precisa de pelo menos uma foto.");
  }
  banco().prepare("DELETE FROM sobre_carrossel WHERE id = ?").run(id);
  await descartar(linha.base, CARROSSEL);
}

export function moverNoCarrossel(id: number, direcao: "antes" | "depois"): void {
  const itens = banco().prepare("SELECT id FROM sobre_carrossel ORDER BY sort_order, id").all() as { id: number }[];
  const posicao = itens.findIndex((item) => item.id === id);
  const destino = posicao + (direcao === "antes" ? -1 : 1);
  if (posicao < 0 || destino < 0 || destino >= itens.length) return;
  [itens[posicao], itens[destino]] = [itens[destino], itens[posicao]];
  banco().transaction(() => {
    itens.forEach((item, indice) => banco().prepare("UPDATE sobre_carrossel SET sort_order = ? WHERE id = ?").run((indice + 1) * 10, item.id));
  })();
}
