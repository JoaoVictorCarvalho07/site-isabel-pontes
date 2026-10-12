import { spawn } from "node:child_process";
import { copyFileSync, rmSync } from "node:fs";
import { resolve } from "node:path";

const banco = resolve("data/teste.db");
const anexos = resolve("data/teste-uploads");
const porta = process.env.PORT ?? "8098";

for (const sufixo of ["", "-wal", "-shm"]) rmSync(banco + sufixo, { force: true });
copyFileSync("data/paridade.db", banco);
rmSync(anexos, { recursive: true, force: true });

const servidor = spawn(process.execPath, [".next/standalone/server.js"], {
  stdio: "inherit",
  env: {
    ...process.env,
    PORT: porta,
    HOSTNAME: "127.0.0.1",
    BASE_URL: `http://127.0.0.1:${porta}`,
    DB_PATH: banco,
    UPLOAD_DIR: anexos,
    RESEND_API_KEY: "",
    MAIL_PARA: "",
  },
});

servidor.on("exit", (codigo) => process.exit(codigo ?? 0));
