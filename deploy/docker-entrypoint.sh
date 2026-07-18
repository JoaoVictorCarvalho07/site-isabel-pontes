#!/bin/sh
# Entrypoint da imagem Docker (Fly.io). Cobre os 3 cenarios de boot:
#   1. volume perdido        -> restaura o banco do R2 (Litestream)
#   2. primeiro deploy        -> popula o banco do manifest (sync)
#   3. boot normal            -> so replica
set -e

: "${DB_PATH:=/data/site.db}"
: "${MANIFEST_PATH:=/app/manifest.json}"

# 1. restaura do R2 se o arquivo sumiu mas ha backup replicado
litestream restore -if-db-not-exists -if-replica-exists "$DB_PATH" || true

# 2. se ainda nao existe (primeiro deploy, sem backup), popula do manifest
if [ ! -f "$DB_PATH" ]; then
  echo "banco ausente -- populando do manifest $MANIFEST_PATH"
  /app/site sync "$MANIFEST_PATH"
fi

# 3. serve com replicacao continua para o R2
exec litestream replicate -exec "/app/site serve"
