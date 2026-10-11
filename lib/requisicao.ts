import { headers } from "next/headers";
import type { Idioma } from "./i18n";

export type Contexto = { idioma: Idioma; caminhoPT: string };

export async function contexto(): Promise<Contexto> {
  const caminho = (await headers()).get("x-caminho") ?? "/";
  if (caminho === "/en") return { idioma: "en", caminhoPT: "/" };
  if (caminho.startsWith("/en/")) return { idioma: "en", caminhoPT: caminho.slice(3) };
  return { idioma: "pt", caminhoPT: caminho };
}
