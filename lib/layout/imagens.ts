import { asset } from "../asset";
import { banco } from "../banco";
import { SLOTS, type NomeSlot } from "./slots";

export type ImagemFixa = { base: string; w: number; h: number };

export type ItemCarrossel = ImagemFixa & { id: number };

export function imagemFixa(slot: NomeSlot): ImagemFixa {
  const linha = banco().prepare("SELECT base, w, h FROM layout_imagens WHERE slot = ?").get(slot) as ImagemFixa | undefined;
  return linha ?? SLOTS[slot].padrao;
}

export function carrossel(): ItemCarrossel[] {
  return banco().prepare("SELECT id, base, w, h FROM sobre_carrossel ORDER BY sort_order, id").all() as ItemCarrossel[];
}

export function endereco(caminho: string): string {
  return caminho.startsWith("/static/") ? asset(caminho) : caminho;
}
