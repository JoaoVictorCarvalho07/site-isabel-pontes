import { mkdirSync, writeFileSync } from "node:fs";
import { join } from "node:path";
import { parseArgs } from "node:util";
import { chromium } from "playwright";

const { values } = parseArgs({
  options: {
    base: { type: "string", default: "http://127.0.0.1:8099" },
    saida: { type: "string", default: "paridade/referencia" },
    paginas: { type: "string" },
    idiomas: { type: "string" },
  },
});

const TODAS_PAGINAS = [
  ["home", "/"],
  ["galeria", "/galeria"],
  ["galeria-casal", "/galeria?tema=casal"],
  ["sobre", "/sobre"],
  ["contato", "/contato"],
  ["briefing", "/briefing"],
  ["midiakit", "/midiakit"],
  ["404", "/nao-existe"],
];
const TODOS_IDIOMAS = [
  ["pt", ""],
  ["en", "/en"],
];
const filtrar = (lista, nomes) => (nomes ? lista.filter(([nome]) => nomes.split(",").includes(nome)) : lista);
const PAGINAS = filtrar(TODAS_PAGINAS, values.paginas);
const IDIOMAS = filtrar(TODOS_IDIOMAS, values.idiomas);
const TEMAS = ["light", "dark"];
const LARGURAS = [375, 480, 768, 900, 1100, 1440];
const LARGURA_A4 = 794;
const ALTURA_MAXIMA = 14000;

function url(prefixo, caminho) {
  if (!prefixo) return values.base + caminho;
  return values.base + prefixo + (caminho === "/" ? "" : caminho);
}

async function estabilizar(pagina) {
  await pagina.evaluate(async () => {
    const altura = document.documentElement.scrollHeight;
    for (let y = 0; y < altura; y += window.innerHeight) {
      window.scrollTo(0, y);
      await new Promise((r) => setTimeout(r, 120));
    }
    window.scrollTo(0, 0);
    document.querySelectorAll("img[loading=lazy]").forEach((img) => (img.loading = "eager"));
    await Promise.all(
      [...document.images].map((img) =>
        img.complete ? null : new Promise((r) => ((img.onload = r), (img.onerror = r))),
      ),
    );
    await Promise.all([...document.images].map((img) => img.decode().catch(() => null)));
    await document.fonts.ready;
    document.querySelectorAll("video").forEach((video) => {
      video.pause();
      video.removeAttribute("autoplay");
    });
  });
  await pagina.waitForTimeout(600);
}

async function fotografar(pagina, caminho, mascara = []) {
  const altura = await pagina.evaluate(() => document.documentElement.scrollHeight);
  const largura = pagina.viewportSize().width;
  await pagina.screenshot({
    path: caminho,
    fullPage: true,
    clip: { x: 0, y: 0, width: largura, height: Math.min(altura, ALTURA_MAXIMA) },
    mask: mascara,
    timeout: 90000,
  });
}

async function abrir(contexto, endereco, largura) {
  const pagina = await contexto.newPage();
  await pagina.setViewportSize({ width: largura, height: 900 });
  await pagina.goto(endereco, { waitUntil: "networkidle" });
  await estabilizar(pagina);
  return pagina;
}

const navegador = await chromium.launch();
let total = 0;

for (const tema of TEMAS) {
  const contexto = await navegador.newContext({
    colorScheme: tema,
    reducedMotion: "reduce",
    deviceScaleFactor: 1,
  });
  await contexto.addInitScript((valor) => localStorage.setItem("tema", valor), tema);

  for (const [idioma, prefixo] of IDIOMAS) {
    const pasta = join(values.saida, tema, idioma);
    mkdirSync(pasta, { recursive: true });

    for (const [nome, caminho] of PAGINAS) {
      for (const largura of LARGURAS) {
        const pagina = await abrir(contexto, url(prefixo, caminho), largura);
        await fotografar(pagina, join(pasta, `${nome}-${largura}.png`), [pagina.locator("video")]);
        if (tema === "light" && largura === 1440) {
          const html = await (await contexto.request.get(url(prefixo, caminho))).text();
          mkdirSync(join(values.saida, "html", idioma), { recursive: true });
          writeFileSync(join(values.saida, "html", idioma, `${nome}.html`), html);
        }
        await pagina.close();
        total++;
      }
    }

    if (!PAGINAS.some(([nome]) => nome === "midiakit")) continue;
    const impressao = await abrir(contexto, url(prefixo, "/midiakit"), LARGURA_A4);
    await impressao.evaluate(() => document.querySelectorAll("details").forEach((d) => (d.open = true)));
    await impressao.emulateMedia({ media: "print", colorScheme: tema });
    await impressao.waitForTimeout(300);
    await fotografar(impressao, join(pasta, "midiakit-impressao.png"));
    await impressao.close();
    total++;
    console.log(`${tema}/${idioma} pronto`);
  }
  await contexto.close();
}

await navegador.close();
console.log(`${total} capturas em ${values.saida}`);
