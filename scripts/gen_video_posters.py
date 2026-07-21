#!/usr/bin/env python3
"""
Gera os posters do portfolio de videos do midia kit e o catalogo Go.

Os videos ficam no R2 (Cloudflare, publicos). Autoplaya-los todos seria 322 MB;
em vez disso mostramos um POSTER por video e so carregamos o .webm no clique.

Para cada video listado em videos-index.json:
  1. ffmpeg extrai 1 frame (~1s) lendo direto da URL do R2 (range-read, sem
     baixar o video inteiro) e ja codifica em AVIF (opaco -> AVIF e o padrao).
  2. Salva em static/img/midiakit/videos/<slug>.avif.
Depois escreve internal/web/midiakit_videos.go com o catalogo (structs, type-safe).

Fontes (do repo do midia kit estatico): videos-index.json + labels.json.
Requisitos: ffmpeg com libaom-av1 no PATH.

Uso:  python scripts/gen_video_posters.py
"""

import json
import re
import subprocess
import sys
import unicodedata
import urllib.parse
from pathlib import Path

try:
    sys.stdout.reconfigure(encoding="utf-8")
except (AttributeError, ValueError):
    pass

# Fonte dos dados: o repo do midia kit estatico, ao lado deste projeto.
FONTE = Path(r"D:/projetos/Midiakit_isabelPontes")
R2_BASE = "https://pub-2584b081d14449de91b2354653b1d10f.r2.dev/videos-ugc"

RAIZ = Path(__file__).resolve().parent.parent
DIR_POSTERS = RAIZ / "static" / "img" / "midiakit" / "videos"
ARQ_GO = RAIZ / "internal" / "web" / "midiakit_videos.go"
LARGURA_POSTER = 400  # 2x do tamanho exibido (~200px no mockup)

# Rotulos das categorias: slug de filtro + nome exibido. Chave = como vem no JSON.
CATEGORIAS = {
    "BELEZA":                 ("beleza",      "Beleza"),
    "EVENTOS":                ("eventos",     "Eventos"),
    "FITNESS":                ("fitness",     "Fitness"),
    "GASTRONOMIA":            ("gastronomia", "Gastronomia"),
    "MODA":                   ("moda",        "Moda"),
    "SAUDE":                  ("saude",       "Saúde"),
    "PARCERIAS":              ("parcerias",   "Parcerias"),
    "diario de uma fotografa":("diario",      "Diário de uma fotógrafa"),
}


def sem_acento(s: str) -> str:
    return "".join(c for c in unicodedata.normalize("NFKD", s) if not unicodedata.combining(c))


def slugify(s: str) -> str:
    s = sem_acento(s).lower()
    s = re.sub(r"[^a-z0-9]+", "-", s).strip("-")
    return s


def sem_ext(nome: str) -> str:
    return re.sub(r"\.[^.]+$", "", nome)


def prettify(nome: str) -> str:
    return sem_ext(nome).replace("_", " ").strip()


def escapar_go(s: str) -> str:
    return s.replace("\\", "\\\\").replace('"', '\\"')


def gerar_poster(url: str, base: Path) -> bool:
    """
    Extrai 1 frame do video (range-read no R2) e salva AVIF + WebP com o mesmo
    nome-base. Dois formatos porque o projeto serve <picture> AVIF + fallback WebP
    (partial 'picture') -- AVIF nao decodifica em 100% dos ambientes.
    """
    from PIL import Image
    import tempfile

    base.parent.mkdir(parents=True, exist_ok=True)
    with tempfile.TemporaryDirectory() as tmp:
        frame = Path(tmp) / "f.png"
        extrair = [
            "ffmpeg", "-y", "-loglevel", "error",
            "-ss", "1", "-i", url, "-frames:v", "1",
            "-vf", f"scale='min({LARGURA_POSTER},iw)':-2",
            str(frame),
        ]
        r = subprocess.run(extrair, capture_output=True, text=True)
        if r.returncode != 0 or not frame.exists():
            print(f"  ✗ {base.name}: {r.stderr.strip().splitlines()[-1] if r.stderr.strip() else 'falha ao extrair'}")
            return False

        # AVIF (padrao) via ffmpeg libaom; WebP (fallback) via Pillow
        avif = [
            "ffmpeg", "-y", "-loglevel", "error", "-i", str(frame),
            "-c:v", "libaom-av1", "-crf", "32", "-cpu-used", "6",
            "-still-picture", "1", "-pix_fmt", "yuv420p",
            str(base.with_suffix(".avif")),
        ]
        if subprocess.run(avif, capture_output=True, text=True).returncode != 0:
            print(f"  ✗ {base.name}: falha no AVIF")
            return False
        Image.open(frame).convert("RGB").save(base.with_suffix(".webp"), "WEBP", quality=80, method=6)
    return True


def main() -> None:
    idx = json.loads((FONTE / "videos-index.json").read_text(encoding="utf-8"))
    labels = json.loads((FONTE / "labels.json").read_text(encoding="utf-8"))

    categorias_go = []
    total = ok = 0

    for cat, arquivos in idx.items():
        cat_slug, cat_nome = CATEGORIAS.get(cat, (slugify(cat), cat.title()))
        pasta_r2 = cat.lower()  # como o R2 organiza
        print(f"\n{cat_nome}:")

        videos_go = []
        for arq in arquivos:
            total += 1
            base = sem_ext(arq)
            key = base.lower()
            titulo = labels.get(key) or prettify(arq)
            slug = f"{cat_slug}-{slugify(base)}"
            src = f"{R2_BASE}/{urllib.parse.quote(pasta_r2)}/{urllib.parse.quote(arq)}"
            # base sem extensao: o partial 'picture' anexa .avif/.webp
            poster_rel = f"/static/img/midiakit/videos/{slug}"

            if gerar_poster(src, DIR_POSTERS / slug):
                ok += 1
                print(f"  ✓ {titulo}")

            videos_go.append((titulo, src, poster_rel))

        campos = ",\n".join(
            f'\t\t\t{{Titulo: "{escapar_go(t)}", Src: "{escapar_go(s)}", Poster: "{escapar_go(p)}"}}'
            for t, s, p in videos_go
        )
        categorias_go.append(
            f'\t{{\n\t\tSlug: "{cat_slug}", Nome: "{escapar_go(cat_nome)}",\n'
            f"\t\tVideos: []videoUGC{{\n{campos},\n\t\t}},\n\t}}"
        )

    conteudo = (
        "package web\n\n"
        "// GERADO por scripts/gen_video_posters.py a partir de videos-index.json +\n"
        "// labels.json do midia kit estatico. NAO edite a mao -- rode o script de novo.\n"
        "//\n"
        "// Videos ficam no R2 (§6 do plano); posters AVIF em static/img/midiakit/videos/.\n"
        "// O portfolio (#portfolio) mostra o poster e so carrega o .webm no clique --\n"
        "// os 26 videos somam 322 MB, autoplayar tudo seria inviavel.\n\n"
        "type videoUGC struct {\n"
        "\tTitulo string\n\tSrc    string // .webm no R2\n\tPoster string // base do poster local (sem ext); .avif + .webp\n}\n\n"
        "type categoriaVideos struct {\n"
        "\tSlug   string // filtro (ex.: \"beleza\")\n\tNome   string // exibido\n\tVideos []videoUGC\n}\n\n"
        "var portfolioVideos = []categoriaVideos{\n"
        + ",\n".join(categorias_go)
        + ",\n}\n"
    )
    ARQ_GO.write_text(conteudo, encoding="utf-8")

    print(f"\nConcluido: {ok}/{total} posters. Catalogo -> {ARQ_GO.relative_to(RAIZ)}")


if __name__ == "__main__":
    main()
