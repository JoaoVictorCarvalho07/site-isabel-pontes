import { mkdir, rm, writeFile } from "node:fs/promises";
import { dirname, join, resolve } from "node:path";
import { AwsClient } from "aws4fetch";
import { BASE_URL } from "./config";

const conta = process.env.R2_ACCOUNT_ID;
const chave = process.env.R2_ACCESS_KEY_ID;
const segredo = process.env.R2_SECRET_ACCESS_KEY;
const bucketPublico = process.env.R2_BUCKET;
const bucketOriginais = process.env.R2_BUCKET_ORIGINAIS ?? bucketPublico;

export const usaR2 = Boolean(conta && chave && segredo && bucketPublico && process.env.R2_PUBLIC_BASE);
export const PASTA_LOCAL = resolve(/* turbopackIgnore: true */ process.env.R2_LOCAL_DIR ?? "data/r2-local");
export const BASE_PUBLICA = (usaR2 ? (process.env.R2_PUBLIC_BASE as string) : `${BASE_URL}/dev-r2`).replace(/\/$/, "");

export type Destino = "publico" | "originais";

const cliente = usaR2 ? new AwsClient({ accessKeyId: chave as string, secretAccessKey: segredo as string, service: "s3", region: "auto" }) : null;

function endereco(destino: Destino, objeto: string): string {
  const bucket = destino === "publico" ? bucketPublico : bucketOriginais;
  return `https://${conta}.r2.cloudflarestorage.com/${bucket}/${objeto.split("/").map(encodeURIComponent).join("/")}`;
}

function caminhoLocal(destino: Destino, objeto: string): string {
  const caminho = resolve(/* turbopackIgnore: true */ PASTA_LOCAL, destino, objeto);
  if (!caminho.startsWith(join(/* turbopackIgnore: true */ PASTA_LOCAL, destino))) throw new Error(`objeto fora da pasta: ${objeto}`);
  return caminho;
}

export async function gravar(destino: Destino, objeto: string, bytes: Uint8Array, tipo: string): Promise<void> {
  if (!cliente) {
    const caminho = caminhoLocal(destino, objeto);
    await mkdir(dirname(caminho), { recursive: true });
    await writeFile(caminho, bytes);
    return;
  }
  const assinado = await cliente.sign(endereco(destino, objeto), {
    method: "PUT",
    headers: { "Content-Type": tipo, "Cache-Control": "public, max-age=31536000, immutable", "X-Amz-Content-Sha256": "UNSIGNED-PAYLOAD" },
  });
  const resposta = await fetch(assinado.url, { method: "PUT", headers: assinado.headers, body: new Uint8Array(bytes) });
  if (!resposta.ok) throw new Error(`R2 recusou ${objeto}: ${resposta.status} ${(await resposta.text()).slice(0, 300)}`);
}

export async function apagar(destino: Destino, objeto: string): Promise<void> {
  if (!cliente) {
    await rm(caminhoLocal(destino, objeto), { force: true });
    return;
  }
  const resposta = await cliente.fetch(endereco(destino, objeto), { method: "DELETE" });
  if (!resposta.ok && resposta.status !== 404) throw new Error(`R2 nao apagou ${objeto}: ${resposta.status}`);
}

export function urlPublica(objeto: string): string {
  return `${BASE_PUBLICA}/${objeto}`;
}

export function objetoDaUrl(url: string): string | null {
  return url.startsWith(`${BASE_PUBLICA}/`) ? url.slice(BASE_PUBLICA.length + 1) : null;
}
