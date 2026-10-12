import { readFile } from "node:fs/promises";
import { resolve, sep } from "node:path";
import { anexoPorId } from "@/lib/admin/briefings";
import { emailAdmin } from "@/lib/admin/sessao";
import { UPLOAD_DIR } from "@/lib/config";
import { codificar } from "@/lib/url";

export async function GET(_requisicao: Request, { params }: { params: Promise<{ id: string }> }) {
  if (!(await emailAdmin())) return new Response(null, { status: 404 });
  const anexo = anexoPorId(Number((await params).id));
  if (!anexo) return new Response(null, { status: 404 });
  const base = resolve(/* turbopackIgnore: true */ UPLOAD_DIR);
  const caminho = resolve(/* turbopackIgnore: true */ base, anexo.caminho);
  if (!caminho.startsWith(base + sep)) {
    console.error("anexo: caminho fora da pasta de uploads", { caminho: anexo.caminho });
    return new Response(null, { status: 404 });
  }
  try {
    const bytes = await readFile(caminho);
    return new Response(new Uint8Array(bytes), {
      headers: {
        "Content-Type": anexo.mime,
        "X-Content-Type-Options": "nosniff",
        "Content-Security-Policy": "default-src 'none'; sandbox",
        "Content-Disposition": `inline; filename*=UTF-8''${codificar(anexo.nome)}`,
        "Cache-Control": "private, max-age=300",
      },
    });
  } catch {
    return new Response(null, { status: 404 });
  }
}
