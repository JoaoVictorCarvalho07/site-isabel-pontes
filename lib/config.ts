export const BASE_URL = (process.env.BASE_URL ?? "http://127.0.0.1:3000").replace(/\/$/, "");
export const DB_PATH = process.env.DB_PATH ?? "data/site.db";
export const NOME_SITE = "Isabel Pontes Fotoarte";
export const UPLOAD_DIR = process.env.UPLOAD_DIR ?? "data/uploads";
