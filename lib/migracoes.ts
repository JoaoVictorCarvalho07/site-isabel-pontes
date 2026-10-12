import type Database from "better-sqlite3";

const MIGRACOES: [versao: number, sql: string][] = [
  [
    5,
    `ALTER TABLE photos ADD COLUMN original TEXT;
     ALTER TABLE photos ADD COLUMN excluida_em TEXT;`,
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
