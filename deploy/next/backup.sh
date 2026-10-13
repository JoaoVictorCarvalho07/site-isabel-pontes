#!/usr/bin/env bash
set -euo pipefail

RAIZ="${RAIZ:-$HOME/apps/site-next}"
REMOTE="${REMOTE:-r2backup}"
BUCKET="${BUCKET:-isabel-site-backup}"
CARIMBO="$(date -u +%Y-%m-%d)"
COPIA="$RAIZ/data/.copia-do-dia.db"

rm -f "$COPIA"
trap 'rm -f "$COPIA"' EXIT

docker exec site-next node -e "
  const banco = require('/app/node_modules/better-sqlite3')('/dados/site.db', { readonly: true });
  banco.backup('/dados/.copia-do-dia.db').then(() => process.exit(0), (erro) => { console.error(erro); process.exit(1); });
"

rclone copyto "$COPIA" "$REMOTE:$BUCKET/db/site-$CARIMBO.db" --s3-no-check-bucket --stats-one-line
rclone copy "$RAIZ/data/uploads" "$REMOTE:$BUCKET/uploads" --s3-no-check-bucket --stats-one-line --immutable
echo "[backup] concluido: site-$CARIMBO.db"
