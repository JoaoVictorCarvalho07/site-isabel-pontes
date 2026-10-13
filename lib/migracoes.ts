import type Database from "better-sqlite3";

const MIGRACOES: [versao: number, sql: string][] = [
  [
    5,
    `ALTER TABLE photos ADD COLUMN original TEXT;
     ALTER TABLE photos ADD COLUMN excluida_em TEXT;`,
  ],
  [
    6,
    `CREATE TABLE layout_imagens (
       slot TEXT PRIMARY KEY,
       base TEXT NOT NULL,
       w INTEGER NOT NULL,
       h INTEGER NOT NULL,
       atualizado_em TEXT NOT NULL DEFAULT (datetime('now'))
     );
     CREATE TABLE sobre_carrossel (
       id INTEGER PRIMARY KEY,
       base TEXT NOT NULL,
       w INTEGER NOT NULL,
       h INTEGER NOT NULL,
       sort_order INTEGER NOT NULL
     );
     INSERT INTO sobre_carrossel (base, w, h, sort_order) VALUES
       ('/static/img/sobre/SobreMim1', 800, 1066, 10),
       ('/static/img/sobre/SobreMim2', 800, 600, 20),
       ('/static/img/sobre/SobreMim3', 800, 450, 30),
       ('/static/img/sobre/SobreMim4', 800, 1200, 40),
       ('/static/img/sobre/SobreMim5', 800, 1200, 50),
       ('/static/img/sobre/SobreMim6', 433, 728, 60),
       ('/static/img/sobre/SobreMim7', 361, 560, 70),
       ('/static/img/sobre/SobreMim8', 706, 1280, 80);`,
  ],
];

export function migrar(banco: Database.Database): void {
  const linha = banco.prepare("SELECT v FROM meta WHERE k = 'schema_version'").get() as { v: string } | undefined;
  const atual = Number(linha?.v ?? 0);
  const pendentes = MIGRACOES.filter(([versao]) => versao > atual);
  if (pendentes.length === 0) return;
  banco.transaction(() => {
    for (const [versao, sql] of pendentes) {
      banco.exec(sql);
      banco.prepare("INSERT INTO meta (k, v) VALUES ('schema_version', ?) ON CONFLICT(k) DO UPDATE SET v = excluded.v").run(String(versao));
    }
  })();
}
