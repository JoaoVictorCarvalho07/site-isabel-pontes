#!/usr/bin/env bash
#
# Backup do site: banco + imagens de referencia dos briefings, para o R2.
#
# Roda pelo systemd timer (deploy/backup.timer), uma vez por dia. Ver
# deploy/README.md para a instalacao.
#
# POR QUE ISTO EXISTE
# O Litestream cobriria o SQLite continuamente, mas nao cobre ARQUIVO -- e os
# anexos de briefing sao arquivos em data/uploads/. Um backup so do banco
# deixaria a Isa com o texto do briefing e sem as referencias visuais que o
# cliente mandou. Este script cobre os dois, no mesmo lugar.
#
# O QUE ELE NAO E
# Nao substitui o Litestream para RPO baixo: o pior caso aqui e perder ate 24h.
# Para o volume atual (alguns briefings por semana) esta e a troca certa; se um
# dia isso mudar, o Litestream entra POR CIMA disto, nao no lugar.

set -euo pipefail

# --- Configuracao ------------------------------------------------------------
# Ajuste se o deploy nao seguir o layout padrao.
DIR_APP="${DIR_APP:-/home/joao/apps/site-test}"
DIR_DADOS="$DIR_APP/data"
# Nome do remote configurado no rclone (rclone config -> tipo s3, provider
# Cloudflare). Ver README.
REMOTE="${REMOTE:-r2backup}"
BUCKET="${BUCKET:-isabel-site-backup}"
# Quantos snapshots diarios do banco manter no destino.
MANTER_DIAS="${MANTER_DIAS:-30}"

log() { echo "[backup] $*"; }

# --- 1. Snapshot consistente do banco ---------------------------------------
# `cp site.db` NAO serve: em modo WAL as escritas recentes vivem no -wal, e a
# copia sai corrompida sem aviso. O subcomando usa VACUUM INTO, que e a forma
# oficial de copiar um SQLite a quente. Ver internal/store/store.go.
CARIMBO="$(date -u +%Y-%m-%d)"
SNAPSHOT="/tmp/site-$CARIMBO.db"

log "gerando snapshot do banco..."
cd "$DIR_APP"
./site snapshot "$SNAPSHOT"

# O snapshot vai para /tmp e some no fim: nao deixamos uma copia do banco
# (com dados pessoais dos briefings) parada no disco da VM sem necessidade.
trap 'rm -f "$SNAPSHOT"' EXIT

# --- 2. Envia o banco --------------------------------------------------------
log "enviando banco para $REMOTE:$BUCKET/db/ ..."
rclone copyto "$SNAPSHOT" "$REMOTE:$BUCKET/db/site-$CARIMBO.db" \
  --s3-no-check-bucket --stats-one-line

# --- 3. Espelha os uploads ---------------------------------------------------
# `copy`, nao `sync`: sync APAGA no destino o que sumiu na origem. Se o disco da
# VM corromper e esvaziar data/uploads/, um `sync` propagaria a perda para o
# backup -- exatamente quando ele mais importa. Os anexos nunca sao editados,
# entao copy incremental basta e nada e sobrescrito.
if [ -d "$DIR_DADOS/uploads" ]; then
  log "enviando anexos..."
  rclone copy "$DIR_DADOS/uploads" "$REMOTE:$BUCKET/uploads" \
    --s3-no-check-bucket --stats-one-line --immutable
else
  log "sem uploads ainda, pulando"
fi

# --- 4. Retencao dos snapshots do banco --------------------------------------
# Só os snapshots diarios do banco tem validade; anexo nunca e apagado.
log "limpando snapshots com mais de $MANTER_DIAS dias..."
rclone delete "$REMOTE:$BUCKET/db" --min-age "${MANTER_DIAS}d" --s3-no-check-bucket

log "concluido."
