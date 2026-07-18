# Fallback de deploy: Fly.io (ou qualquer host Docker), caso a VM Oracle demore.
# O alvo primario e a VM Oracle via systemd -- ver deploy/README.md.
#
# Imagem final ~20MB: alpine (precisa do shell para o entrypoint restaurar o
# banco), binario Go estatico (CGO_ENABLED=0, gracas ao driver modernc) e o
# Litestream.

# --- build do binario ---
FROM golang:1.26-alpine AS build
WORKDIR /src
COPY go.mod go.sum ./
RUN go mod download
COPY . .
RUN CGO_ENABLED=0 go build -ldflags="-s -w" -o /site ./cmd/site

# --- Litestream ---
FROM litestream/litestream:0.3 AS litestream

# --- runtime ---
FROM alpine:3.20
RUN apk add --no-cache ca-certificates
WORKDIR /app
COPY --from=build /site /app/site
COPY --from=litestream /usr/local/bin/litestream /usr/local/bin/litestream
COPY deploy/litestream.yml /etc/litestream.yml
COPY deploy/docker-entrypoint.sh /app/entrypoint.sh
# o manifest vai embutido na imagem para o primeiro `sync` funcionar sem volume
# populado. (gitignored, mas presente no contexto de build.)
COPY data/manifest.json /app/manifest.json
RUN chmod +x /app/entrypoint.sh

# Litestream para backup no R2 usa este bucket (litestream.yml).
ENV MANIFEST_PATH=/app/manifest.json
EXPOSE 8080
ENTRYPOINT ["/app/entrypoint.sh"]
