import { randomBytes } from "node:crypto";
import { mkdir, writeFile } from "node:fs/promises";
import { dirname, join } from "node:path";
import { banco } from "../banco";
import { UPLOAD_DIR } from "../config";
import { farejarTipo, MAXIMO_ARQUIVOS, nomeParaExibir, TAMANHO_MAXIMO_ARQUIVO } from "./tipos";

export async function salvarAnexos(idResposta: number, campo: string, arquivos: File[]): Promise<number> {
  let salvos = 0;
  for (const arquivo of arquivos.slice(0, MAXIMO_ARQUIVOS)) {
    if (arquivo.size <= 0 || arquivo.size > TAMANHO_MAXIMO_ARQUIVO) continue;
    const bytes = new Uint8Array(await arquivo.arrayBuffer());
    const tipo = farejarTipo(bytes);
    if (!tipo) {
      console.warn("briefing: anexo recusado", { idResposta, nome: nomeParaExibir(arquivo.name) });
      continue;
    }
    const agora = new Date();
    const relativo = [
      String(agora.getFullYear()),
      String(agora.getMonth() + 1).padStart(2, "0"),
      `${idResposta}-${randomBytes(8).toString("hex")}${tipo.extensao}`,
    ].join("/");
    const destino = join(UPLOAD_DIR, relativo);
    await mkdir(dirname(destino), { recursive: true });
    await writeFile(destino, bytes, { flag: "wx", mode: 0o640 });
    banco()
      .prepare("INSERT INTO submission_files (submission_id, campo, caminho, nome_original, mime, tamanho) VALUES (?,?,?,?,?,?)")
      .run(idResposta, campo, relativo, nomeParaExibir(arquivo.name), tipo.mime, bytes.length);
    salvos++;
  }
  return salvos;
}
