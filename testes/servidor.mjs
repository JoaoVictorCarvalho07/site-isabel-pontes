import { spawn } from "node:child_process";
import { copyFileSync, rmSync } from "node:fs";
import { resolve } from "node:path";

const banco = resolve("data/teste.db");
const anexos = resolve("data/teste-uploads");
const porta = process.env.PORT ?? "8098";

for (const sufixo of ["", "-wal", "-shm"]) rmSync(banco + sufixo, { force: true });
copyFileSync(process.argv[2] ?? "data/site.db", banco);
rmSync(anexos, { recursive: true, force: true });
rmSync(resolve("data/teste-r2"), { recursive: true, force: true });

const servidor = spawn(process.execPath, [".next/standalone/server.js"], {
  stdio: "inherit",
  env: {
    ...process.env,
    PORT: porta,
    HOSTNAME: "127.0.0.1",
    BASE_URL: `http://127.0.0.1:${porta}`,
    DB_PATH: banco,
    UPLOAD_DIR: anexos,
    AUTH_SECRET: "segredo-apenas-para-testes-locais-0123456789",
    AUTH_GOOGLE_ID: "teste",
    AUTH_GOOGLE_SECRET: "teste",
    ADMIN_EMAILS: "isa@teste.dev",
    R2_LOCAL_DIR: resolve("data/teste-r2"),
    R2_ACCOUNT_ID: "",
    RESEND_API_KEY: "",
    MAIL_PARA: "",
  },
});

servidor.on("exit", (codigo) => process.exit(codigo ?? 0));
