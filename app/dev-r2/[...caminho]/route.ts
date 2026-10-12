import { readFile } from "node:fs/promises";
import { join, resolve } from "node:path";
import { PASTA_LOCAL, usaR2 } from "@/lib/armazenamento";

const TIPOS: Record<string, string> = { avif: "image/avif", webp: "image/webp", jpg: "image/jpeg", png: "image/png" };

export async function GET(_requisicao: Request, { params }: { params: Promise<{ caminho: string[] }> }) {
  if (usaR2) return new Response(null, { status: 404 });
  const raiz = join(/* turbopackIgnore: true */ PASTA_LOCAL, "publico");
  const arquivo = resolve(/* turbopackIgnore: true */ raiz, ...(await params).caminho);
  if (!arquivo.startsWith(raiz)) return new Response(null, { status: 404 });
  try {
    const bytes = await readFile(arquivo);
    return new Response(bytes as BodyInit, { headers: { "Content-Type": TIPOS[arquivo.split(".").pop() ?? ""] ?? "application/octet-stream" } });
  } catch {
    return new Response(null, { status: 404 });
  }
}
