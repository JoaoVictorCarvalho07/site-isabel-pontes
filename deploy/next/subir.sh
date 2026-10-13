#!/usr/bin/env bash
set -euo pipefail

RAIZ="${RAIZ:-$HOME/apps/site-next}"
PACOTE="${1:?uso: subir.sh <pacote.tgz>}"
VERSAO="$(date -u +%Y%m%d-%H%M%S)"
DESTINO="$RAIZ/versoes/$VERSAO"

mkdir -p "$DESTINO" "$RAIZ/data/uploads"
tar -xzf "$PACOTE" -C "$DESTINO" --strip-components=1
test -f "$DESTINO/server.js"
test -f "$RAIZ/.env"
test -f "$RAIZ/data/site.db"

docker rm -f site-next >/dev/null 2>&1 || true
docker run -d --name site-next --restart unless-stopped --memory 1g \
  --user "$(id -u):$(id -g)" \
  -p 127.0.0.1:8082:3000 \
  --env-file "$RAIZ/.env" \
  -v "$DESTINO:/app" -v "$RAIZ/data:/dados" -w /app \
  node:22-slim node server.js >/dev/null

for tentativa in $(seq 1 30); do
  if curl -fsS -o /dev/null http://127.0.0.1:8082/sobre; then
    ls -1dt "$RAIZ"/versoes/* | tail -n +4 | xargs -r rm -rf
    echo "no ar: $VERSAO"
    exit 0
  fi
  sleep 1
done

docker logs --tail 40 site-next
exit 1
