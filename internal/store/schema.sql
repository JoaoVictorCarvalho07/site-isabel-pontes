-- Schema do site. Embutido no binario via //go:embed e aplicado no boot.
--
-- Lembrete do modelo: o `sync` escreve aqui, o boot carrega tudo para a RAM, e
-- os requests NUNCA tocam o banco. Por isso indices sao minimos -- a query
-- quente roda em memoria, nao em SQL.

PRAGMA journal_mode = WAL;
PRAGMA foreign_keys = ON;

-- Temas: o slug e a pasta no R2 ('epoca'); o label e o que o usuario ve ('Época').
-- Essa separacao conserta um bug visivel no React: os botoes do filtro mostravam
-- 'epoca', 'artistico', 'personalizado' -- crus, minusculos, sem acento.
CREATE TABLE IF NOT EXISTS temas (
  slug       TEXT PRIMARY KEY,
  label      TEXT NOT NULL,
  sort_order INTEGER NOT NULL DEFAULT 100,
  visible    INTEGER NOT NULL DEFAULT 1
);

-- Fotos. As variantes (grid/grid2x/full) e o lqip vem prontos do manifest que o
-- migrate_media.py gera -- o Go nao processa imagem, so serve as URLs.
CREATE TABLE IF NOT EXISTS photos (
  id         INTEGER PRIMARY KEY,
  key        TEXT NOT NULL UNIQUE,   -- 'Eventos/Eventos_1.webp' (identidade)
  tema_slug  TEXT NOT NULL REFERENCES temas(slug),
  w          INTEGER NOT NULL,       -- dimensoes do original -> aspect ratio -> sem CLS
  h          INTEGER NOT NULL,
  grid       TEXT NOT NULL,          -- AVIF 400px
  grid2x     TEXT NOT NULL,          -- AVIF 800px
  full       TEXT NOT NULL,          -- AVIF 1600px (lightbox)
  lqip       TEXT NOT NULL,          -- data URI, placeholder inline
  alt        TEXT NOT NULL DEFAULT '',
  published  INTEGER NOT NULL DEFAULT 1,
  -- sort_order NULL => entra no shuffle diario; preenchido => fixo no topo.
  -- E o caminho para "ordem curada > aleatoria" sem mexer em codigo.
  sort_order INTEGER,
  synced_at  TEXT NOT NULL DEFAULT (datetime('now'))
);

CREATE INDEX IF NOT EXISTS idx_photos_feed ON photos (tema_slug, published);

-- Chave-valor de sistema. data_version incrementa a cada sync e entra no ETag,
-- entao um sync invalida todo o cache HTTP de uma vez.
CREATE TABLE IF NOT EXISTS meta (
  k TEXT PRIMARY KEY,
  v TEXT NOT NULL
);
