import { readFileSync } from "node:fs";
import { join } from "node:path";

const cache = new Map<string, string>();

export function svg(caminho: string): string {
  let conteudo = cache.get(caminho);
  if (conteudo === undefined) {
    conteudo = readFileSync(join(process.cwd(), "public/static", caminho), "utf8").trim();
    cache.set(caminho, conteudo);
  }
  return conteudo;
}
