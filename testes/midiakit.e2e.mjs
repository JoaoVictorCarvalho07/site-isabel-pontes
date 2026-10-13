import assert from "node:assert/strict";
import { parseArgs } from "node:util";
import { chromium } from "playwright";

const { values } = parseArgs({ options: { base: { type: "string", default: "http://127.0.0.1:8098" } } });

const resposta = await fetch(`${values.base}/midiakit`);
const html = await resposta.text();
assert.equal(resposta.status, 200);
for (const marcador of ["R$ 550", "Mais popular", "wa.me/554191977011", "conic-gradient", 'action="/briefing/ugc"']) {
  assert.ok(html.includes(marcador), `o HTML do servidor tem ${marcador}`);
}
const links = [...html.matchAll(/href="(https:\/\/wa\.me\/[^"]+)"/g)].map((m) => m[1]);
assert.equal(links.length, 8, "seis pacotes, o contato geral e o rodape");
assert.ok(links.every((link) => !link.includes("+")), "os links nao usam + como espaco");
assert.ok(!html.includes('name="lanche"'), "o formulario UGC nao traz pergunta de fotografia");

const ingles = await (await fetch(`${values.base}/en/midiakit`)).text();
assert.ok(ingles.includes("UGC Investment"));
assert.ok(ingles.includes('action="/en/briefing/ugc"'));

const navegador = await chromium.launch();
const pagina = await navegador.newPage({ viewport: { width: 1100, height: 800 } });
const erros = [];
pagina.on("pageerror", (erro) => erros.push(erro.message));
await pagina.goto(`${values.base}/midiakit`, { waitUntil: "networkidle" });

const grupos = pagina.locator(".mk-vgrupo");
const total = await grupos.count();
assert.equal(total, 8, "todas as categorias aparecem");
assert.equal(await pagina.locator(".mk-video").count(), 26);

await pagina.locator(".mk-filtro", { hasText: "Fitness" }).evaluate((botao) => botao.click());
assert.equal(await grupos.count(), 1, "o filtro mostra so a categoria pedida");
assert.equal(await pagina.locator(".mk-vgrupo__nome").textContent(), "Fitness");
assert.equal(await pagina.locator(".mk-filtro--ativo").textContent(), "Fitness");
assert.equal(new URL(pagina.url()).search, "", "o filtro de videos nao muda a URL");

const primeiro = pagina.locator(".mk-video__frame").first();
await primeiro.locator(".mk-video__mock").evaluate((botao) => botao.click());
assert.equal(await primeiro.locator("video.mk-video__player").count(), 1, "o clique troca o poster pelo video");
assert.match(await primeiro.locator("video").getAttribute("src"), /\.webm$/);

await pagina.locator(".mk-filtro", { hasText: "Todos" }).evaluate((botao) => botao.click());
assert.equal(await grupos.count(), 8);
assert.equal(await pagina.locator("video.mk-video__player").count(), 0, "trocar o filtro descarta o video aberto");

assert.equal(await pagina.locator("details[open]").count(), 0);
await pagina.evaluate(() => window.dispatchEvent(new Event("beforeprint")));
assert.equal(await pagina.locator("details:not([open])").count(), 0, "a impressao abre todos os detalhes");
await pagina.evaluate(() => window.dispatchEvent(new Event("afterprint")));
assert.equal(await pagina.locator("details[open]").count(), 0);

assert.deepEqual(erros, [], "sem erros de JavaScript");
await navegador.close();
console.log("midiakit ok");
