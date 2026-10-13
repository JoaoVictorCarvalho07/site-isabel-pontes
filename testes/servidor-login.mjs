import { spawn } from "node:child_process";
import { copyFileSync, readFileSync, rmSync } from "node:fs";
import { resolve } from "node:path";

const CHAVES = ["AUTH_GOOGLE_ID", "AUTH_GOOGLE_SECRET", "AUTH_SECRET", "ADMIN_EMAILS"];
const login = Object.fromEntries(
  readFileSync(".env.local", "utf8")
    .split(/\r?\n/)
    .map((linha) => [linha.slice(0, linha.indexOf("=")), linha.slice(linha.indexOf("=") + 1).replace(/^"|"$/g, "")])
    .filter(([chave]) => CHAVES.includes(chave)),
);
const faltando = CHAVES.filter((chave) => !login[chave]);
if (faltando.length > 0) throw new Error(`faltam no .env.local: ${faltando.join(", ")}`);

const banco = resolve("data/teste.db");
for (const sufixo of ["", "-wal", "-shm"]) rmSync(banco + sufixo, { force: true });
copyFileSync("data/site.db", banco);

const servidor = spawn(process.execPath, [".next/standalone/server.js"], {
  stdio: "inherit",
  env: {
    ...process.env,
    ...login,
    PORT: "3000",
    HOSTNAME: "localhost",
    BASE_URL: "http://localhost:3000",
    AUTH_URL: "http://localhost:3000",
    DB_PATH: banco,
    UPLOAD_DIR: resolve("data/teste-uploads"),
    R2_LOCAL_DIR: resolve("data/teste-r2"),
    R2_ACCOUNT_ID: "",
    RESEND_API_KEY: "",
  },
});

servidor.on("exit", (codigo) => process.exit(codigo ?? 0));
