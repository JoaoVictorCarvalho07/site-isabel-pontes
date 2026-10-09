import { copyFileSync } from "node:fs";
import { DatabaseSync } from "node:sqlite";

const origem = "data/site.db";
const destino = "data/paridade.db";

copyFileSync(origem, destino);
const banco = new DatabaseSync(destino);
banco.exec("UPDATE photos SET sort_order = id");
const { total } = banco.prepare("SELECT count(*) AS total FROM photos").get();
banco.close();
console.log(`${destino}: ${total} fotos com ordem fixa`);
