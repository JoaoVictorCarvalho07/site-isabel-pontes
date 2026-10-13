#!/usr/bin/env bash
set -euo pipefail

RAIZ="${RAIZ:-$HOME/apps/site-next}"
mkdir -p "$RAIZ/versoes" "$RAIZ/data/uploads"

if [ $# -ge 1 ]; then
  DESTINO="$RAIZ/versoes/$(date -u +%Y%m%d-%H%M%S)"
  mkdir -p "$DESTINO"
  tar -xzf "$1" -C "$DESTINO" --strip-components=1
else
  DESTINO="$(ls -1dt "$RAIZ"/versoes/* | head -n 1)"
fi

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
    echo "no ar: $(basename "$DESTINO")"
    exit 0
  fi
  sleep 1
done

docker logs --tail 40 site-next
exit 1
