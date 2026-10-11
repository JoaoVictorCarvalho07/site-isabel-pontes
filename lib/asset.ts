import { createHash } from "node:crypto";
import { readFileSync } from "node:fs";
import { join } from "node:path";

const cache = new Map<string, string>();

export function asset(caminho: string): string {
  const pronto = cache.get(caminho);
  if (pronto) return pronto;
  let resultado = caminho;
  try {
    const conteudo = readFileSync(join(process.cwd(), "public", caminho));
    resultado = `${caminho}?v=${createHash("sha256").update(conteudo).digest("hex").slice(0, 8)}`;
  } catch {
    resultado = caminho;
  }
  cache.set(caminho, resultado);
  return resultado;
}
