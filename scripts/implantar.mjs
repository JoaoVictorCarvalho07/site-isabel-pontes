import { execFileSync } from "node:child_process";
import { mkdtempSync, rmSync } from "node:fs";
import { tmpdir } from "node:os";
import { join } from "node:path";

const alvo = process.env.SITE_HOST ?? "oracle";
const executar = (comando, argumentos) => execFileSync(comando, argumentos, { stdio: "inherit" });
const ler = (comando, argumentos) => execFileSync(comando, argumentos, { encoding: "utf8" }).trim();

const commit = ler("git", ["rev-parse", "HEAD"]);
const execucao = ler("gh", ["run", "list", "--workflow", "pacote", "--commit", commit, "--status", "success", "--limit", "1", "--json", "databaseId", "--jq", ".[0].databaseId"]);
if (!execucao) throw new Error(`nenhum pacote pronto para o commit ${commit.slice(0, 7)}; envie a branch e espere o workflow terminar`);

const pasta = mkdtempSync(join(tmpdir(), "site-next-"));
try {
  executar("gh", ["run", "download", execucao, "--name", "site-next-linux-arm64", "--dir", pasta]);
  executar("scp", ["-q", join(pasta, "site-next-linux-arm64.tgz"), `${alvo}:/tmp/site-next.tgz`]);
  executar("scp", ["-q", "deploy/next/subir.sh", `${alvo}:apps/site-next/subir.sh`]);
  executar("ssh", [alvo, "chmod +x ~/apps/site-next/subir.sh && ~/apps/site-next/subir.sh /tmp/site-next.tgz && rm /tmp/site-next.tgz"]);
} finally {
  rmSync(pasta, { recursive: true, force: true });
}
