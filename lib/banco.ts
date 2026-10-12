import Database from "better-sqlite3";
import { DB_PATH } from "./config";
import { migrar } from "./migracoes";

let conexao: Database.Database | undefined;

export function banco(): Database.Database {
  if (!conexao) {
    conexao = new Database(DB_PATH);
    conexao.pragma("journal_mode = WAL");
    conexao.pragma("busy_timeout = 5000");
    conexao.pragma("foreign_keys = ON");
    migrar(conexao);
  }
  return conexao;
}
