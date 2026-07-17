#!/usr/bin/env python3
"""
Migra as fotos de um bucket R2 de qualidade unica para um bucket novo com
variantes responsivas por foto, e produz o manifest + LQIP que o site consome.

POR QUE UM BUCKET NOVO
  - A chave do bucket antigo vazou no historico do git. Criando um bucket novo
    com uma chave nova, a chave velha morre junto com o bucket velho quando voce
    deletar -- em vez de ter que rotacionar e torcer.
  - Voce inspeciona o bucket novo inteiro antes de destruir o antigo.

O QUE ESTE SCRIPT FAZ (por foto canonica)
  original (lido pela URL PUBLICA, sem chave)  ->
    lqip   24px  webp  ~0.2KB  -> data URI, vai no manifest (NAO no bucket)
    grid   400px avif  ~10KB   -> bucket novo, e o thumb do grid
    grid2x 800px avif  ~28KB   -> bucket novo, retina / coluna larga no mobile
    full  1600px avif ~120KB   -> bucket novo, o que o lightbox abre
  Os originais 4000px NAO sao copiados: o site nunca serve resolucao de arquivo.
  >>> Garanta que a Isabel tem os 4000px em outro lugar antes de deletar o
      bucket antigo. <<<

SEGURANCA
  Le originais por HTTP publico. So o UPLOAD usa credencial, e so de ESCRITA no
  bucket NOVO. A chave comprometida do bucket antigo nunca e usada aqui.

DEDUPE
  Os 17 .jpg que sao par redundante de um .webp existente sao descartados na
  migracao. O bucket novo nasce com 190 fotos, nao 207.

IDEMPOTENTE
  Reexecutar pula o que ja existe no bucket novo (HeadObject). Seguro rerodar
  apos falha de rede. Use --force para reescrever tudo.

USO
  # valida a geracao em 6 fotos, sem subir nada e sem precisar de chave:
  python scripts/migrate_media.py --dry-run --limit 6

  # migracao real (precisa das vars R2_NEW_* no ambiente):
  python scripts/migrate_media.py
"""
from __future__ import annotations

import argparse
import base64
import io
import json
import os
import sys
import time
from concurrent.futures import ThreadPoolExecutor, as_completed
from urllib.request import Request, urlopen

from PIL import Image, ImageOps

# ---------------------------------------------------------------------------
# Configuracao. As variantes: um TAMANHO por uso, nao varias qualidades do mesmo
# tamanho. O browser escolhe grid vs grid2x pelo srcset; o lightbox usa full.
# Qualidades calibradas na mao contra as fotos reais (ver o commit).
# ---------------------------------------------------------------------------
VARIANTES = {
    # nome:   (prefixo no bucket novo, largura, qualidade AVIF)
    "grid":   ("variants/grid400",  400, 50),
    "grid2x": ("variants/grid800",  800, 55),
    "full":   ("variants/full1600", 1600, 60),
}
LQIP_LARGURA = 24
LQIP_QUALIDADE = 40
AVIF_SPEED = 6  # 0=melhor/lento .. 10=pior/rapido. 6 e o joelho da curva.

EXTS_IMAGEM = (".webp", ".jpg", ".jpeg", ".png")
CACHE_IMUTAVEL = "public, max-age=31536000, immutable"


def log(*a):
    print(*a, file=sys.stderr, flush=True)


# ---------------------------------------------------------------------------
# Fonte das keys: o manifest.json atual do site React. Usa-lo em vez de listar o
# bucket evita tocar na chave comprometida. O `sync` do Go depois lista o bucket
# de forma autoritativa; para esta migracao unica o manifest basta.
# ---------------------------------------------------------------------------
def carregar_canonicas(caminho_manifest: str) -> list[dict]:
    with open(caminho_manifest, encoding="utf-8") as f:
        entradas = json.load(f)

    # Agrupa por caminho-sem-extensao. Num grupo com .webp e .jpg, o .webp vence
    # e o .jpg (redundante) e descartado -- e aqui que 207 vira 190.
    grupos: dict[str, dict] = {}
    for e in entradas:
        key = e["key"]
        stem = key.rsplit(".", 1)[0]          # "Eventos/Eventos_1"
        ext = "." + key.rsplit(".", 1)[1].lower()
        if ext not in EXTS_IMAGEM:
            continue
        atual = grupos.get(stem)
        # prefere webp; senao mantem o primeiro visto
        if atual is None or (ext == ".webp" and not atual["key"].endswith(".webp")):
            grupos[stem] = e

    return sorted(grupos.values(), key=lambda e: e["key"])


def baixar(url: str) -> bytes:
    # UA de browser: o r2.dev responde 403 para o User-Agent padrao do urllib.
    req = Request(url, headers={"User-Agent": "Mozilla/5.0"})
    with urlopen(req, timeout=60) as r:
        return r.read()


def gerar_variantes(dados: bytes) -> tuple[dict[str, bytes], str, int, int]:
    """Retorna (variantes_bytes, lqip_data_uri, largura_original, altura_original)."""
    src = Image.open(io.BytesIO(dados))
    src = ImageOps.exif_transpose(src)        # respeita rotacao EXIF
    src = src.convert("RGB")
    w0, h0 = src.size

    saidas: dict[str, bytes] = {}
    for nome, (_prefixo, largura, q) in VARIANTES.items():
        im = src.copy()
        im.thumbnail((largura, largura * 100), Image.LANCZOS)  # preserva aspecto
        buf = io.BytesIO()
        im.save(buf, "AVIF", quality=q, speed=AVIF_SPEED)
        saidas[nome] = buf.getvalue()

    # LQIP: minusculo, webp, vira data URI inline no HTML (0 requests).
    lq = src.copy()
    lq.thumbnail((LQIP_LARGURA, LQIP_LARGURA * 100), Image.LANCZOS)
    b = io.BytesIO()
    lq.save(b, "WEBP", quality=LQIP_QUALIDADE)
    lqip = "data:image/webp;base64," + base64.b64encode(b.getvalue()).decode()

    return saidas, lqip, w0, h0


def variant_key(nome: str, key_original: str) -> str:
    """conceitual/conceitual_28.webp -> variants/grid400/conceitual/conceitual_28.avif"""
    prefixo = VARIANTES[nome][0]
    stem = key_original.rsplit(".", 1)[0]
    return f"{prefixo}/{stem}.avif"


class Uploader:
    """Envolve o boto3. Em dry-run, nao sobe nada e nunca pede credencial."""

    def __init__(self, dry_run: bool, force: bool):
        self.dry_run = dry_run
        self.force = force
        self.bucket = os.getenv("R2_NEW_BUCKET", "")
        self._existentes: set[str] = set()
        if dry_run:
            self.s3 = None
            return

        import boto3  # importa so quando vai subir de verdade

        faltando = [k for k in ("R2_NEW_ACCOUNT_ID", "R2_NEW_ACCESS_KEY_ID",
                                "R2_NEW_SECRET_ACCESS_KEY", "R2_NEW_BUCKET")
                    if not os.getenv(k)]
        if faltando:
            log("ERRO: faltam variaveis de ambiente:", ", ".join(faltando))
            log("      (defina-as apontando para o BUCKET NOVO; veja .env.example)")
            sys.exit(1)

        self.s3 = boto3.client(
            "s3",
            endpoint_url=f"https://{os.getenv('R2_NEW_ACCOUNT_ID')}.r2.cloudflarestorage.com",
            aws_access_key_id=os.getenv("R2_NEW_ACCESS_KEY_ID"),
            aws_secret_access_key=os.getenv("R2_NEW_SECRET_ACCESS_KEY"),
            region_name="auto",
        )

    def ja_existe(self, key: str) -> bool:
        if self.dry_run or self.force:
            return False
        try:
            self.s3.head_object(Bucket=self.bucket, Key=key)
            return True
        except Exception:
            return False

    def enviar(self, key: str, dados: bytes, content_type: str):
        if self.dry_run:
            return
        self.s3.put_object(Bucket=self.bucket, Key=key, Body=dados,
                           ContentType=content_type, CacheControl=CACHE_IMUTAVEL)


def processar(entrada: dict, up: Uploader) -> dict:
    key = entrada["key"]
    dados = baixar(entrada["url"])
    variantes, lqip, w0, h0 = gerar_variantes(dados)

    enviados = 0
    for nome, blob in variantes.items():
        vk = variant_key(nome, key)
        if up.ja_existe(vk):
            continue
        up.enviar(vk, blob, "image/avif")
        enviados += 1

    base = os.getenv("R2_NEW_PUBLIC_BASE", "https://NOVO-BUCKET.r2.dev").rstrip("/")
    return {
        "registro": {
            "key": key,
            "tema": entrada.get("tema", key.split("/")[0]),
            "w": w0,
            "h": h0,
            "grid": f"{base}/{variant_key('grid', key)}",
            "grid2x": f"{base}/{variant_key('grid2x', key)}",
            "full": f"{base}/{variant_key('full', key)}",
            "lqip": lqip,
        },
        "enviados": enviados,
        "pesos": {n: len(b) for n, b in variantes.items()},
    }


def main():
    ap = argparse.ArgumentParser()
    ap.add_argument("--manifest", default="../site-da-bebel/siteIsabelPontes/public/manifest.json",
                    help="manifest de origem (keys + URLs publicas)")
    ap.add_argument("--out", default="out/manifest.json", help="manifest novo a gerar")
    ap.add_argument("--dry-run", action="store_true", help="gera e mede, nao sobe nada")
    ap.add_argument("--force", action="store_true", help="reescreve variantes ja existentes")
    ap.add_argument("--limit", type=int, default=0, help="processa so N fotos (validacao)")
    ap.add_argument("--workers", type=int, default=6)
    args = ap.parse_args()

    canonicas = carregar_canonicas(args.manifest)
    if args.limit:
        canonicas = canonicas[:args.limit]
    log(f"{len(canonicas)} fotos canonicas" + (" (LIMITADO)" if args.limit else ""))

    up = Uploader(args.dry_run, args.force)
    if args.dry_run:
        log(">>> DRY-RUN: nada sera enviado, nenhuma credencial necessaria\n")

    registros, total_pesos, enviados = [], {n: 0 for n in VARIANTES}, 0
    t0 = time.time()

    with ThreadPoolExecutor(max_workers=args.workers) as ex:
        futuros = {ex.submit(processar, e, up): e for e in canonicas}
        for i, fut in enumerate(as_completed(futuros), 1):
            e = futuros[fut]
            try:
                r = fut.result()
            except Exception as ex_:
                log(f"  FALHA {e['key']}: {ex_}")
                continue
            registros.append(r["registro"])
            enviados += r["enviados"]
            for n, p in r["pesos"].items():
                total_pesos[n] += p
            log(f"  [{i}/{len(canonicas)}] {e['key']}")

    if not registros:
        log("\nNENHUMA foto processada -- veja os erros acima.")
        sys.exit(1)

    registros.sort(key=lambda r: r["key"])
    os.makedirs(os.path.dirname(args.out), exist_ok=True)
    with open(args.out, "w", encoding="utf-8") as f:
        json.dump(registros, f, ensure_ascii=False, indent=2)

    dt = time.time() - t0
    log(f"\n{'='*54}")
    log(f"fotos:          {len(registros)}")
    log(f"objetos subidos:{enviados}" + (" (dry-run: 0)" if args.dry_run else ""))
    for n in VARIANTES:
        mb = total_pesos[n] / 1024 / 1024
        log(f"  {n:<7} total {mb:6.2f} MB   ({total_pesos[n]/len(registros)/1024:5.1f} KB/foto)")
    lqip_kb = sum(len(r['lqip']) for r in registros) / 1024
    log(f"  lqip    total {lqip_kb:6.1f} KB inline (no manifest, 0 requests)")
    log(f"manifest novo:  {args.out}")
    log(f"tempo:          {dt:.1f}s")
    log("="*54)


if __name__ == "__main__":
    main()
