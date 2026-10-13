import assert from "node:assert/strict";
import { parseArgs } from "node:util";
import Database from "better-sqlite3";
import { encode } from "next-auth/jwt";
import { chromium } from "playwright";
import sharp from "sharp";

const { values } = parseArgs({ options: { base: { type: "string", default: "http://127.0.0.1:8098" } } });
const SEGREDO = "segredo-apenas-para-testes-locais-0123456789";
const COOKIE = "authjs.session-token";

const banco = new Database("data/teste.db", { readonly: true });
const slot = (nome) => banco.prepare("SELECT * FROM layout_imagens WHERE slot = ?").get(nome);
const carrossel = () => banco.prepare("SELECT id, base FROM sobre_carrossel ORDER BY sort_order, id").all();

const jpeg = (largura, altura, cor) => sharp({ create: { width: largura, height: altura, channels: 3, background: cor } }).jpeg().toBuffer();
const html = async (caminho) => (await fetch(values.base + caminho)).text();
const largura = async (url) => (await sharp(Buffer.from(await (await fetch(url)).arrayBuffer())).metadata()).width;

const navegador = await chromium.launch();

const anonimo = await navegador.newContext();
const semLogin = await anonimo.request.post(`${values.base}/admin/api/layout`, {
  multipart: { destino: "hero-retrato", arquivo: { name: "a.jpg", mimeType: "image/jpeg", buffer: await jpeg(10, 10, "#000") } },
});
assert.equal(semLogin.status(), 401, "trocar imagem sem login e recusado");
const paginaAnonima = await anonimo.newPage();
await paginaAnonima.goto(`${values.base}/admin/site`);
assert.equal(new URL(paginaAnonima.url()).pathname, "/admin/entrar");
await anonimo.close();

const contexto = await navegador.newContext({ viewport: { width: 390, height: 844 } });
const valor = await encode({ token: { email: "isa@teste.dev", sub: "1" }, secret: SEGREDO, salt: COOKIE });
await contexto.addCookies([{ name: COOKIE, value: valor, url: values.base, httpOnly: true }]);
const pagina = await contexto.newPage();
const erros = [];
pagina.on("pageerror", (erro) => erros.push(erro.message));
pagina.on("dialog", (dialogo) => dialogo.accept());

const esperar = async (condicao, mensagem) => {
  for (let i = 0; i < 300; i++) {
    if (condicao()) return;
    await pagina.waitForTimeout(100);
  }
  assert.fail(mensagem);
};

assert.ok((await html("/")).includes("/static/img/hero/hero-640.avif"), "sem troca, a home usa a capa original");
assert.equal(carrossel().length, 8, "o carrossel comeca com as oito fotos originais");

await pagina.goto(`${values.base}/admin/site`, { waitUntil: "networkidle" });
assert.equal(await pagina.evaluate(() => document.documentElement.scrollWidth <= window.innerWidth), true, "sem rolagem horizontal no celular");
const cartao = (nome) => pagina.locator(`[data-slot="${nome}"]`);
assert.equal(await cartao("hero-retrato").getByRole("button", { name: "Voltar à original" }).count(), 0);

await cartao("hero-retrato").locator('input[type="file"]').setInputFiles({ name: "capa.jpg", mimeType: "image/jpeg", buffer: await jpeg(2400, 3600, "#553311") });
await esperar(() => slot("hero-retrato") !== undefined, "a capa nova e gravada");
const capa = slot("hero-retrato");
assert.deepEqual([capa.w, capa.h], [2400, 3600]);
const home = await html("/");
for (const tamanho of [640, 960, 1280, 1920]) {
  for (const formato of ["avif", "webp"]) {
    const url = `${capa.base}-${tamanho}.${formato}`;
    assert.ok(home.includes(`${url} ${tamanho}w`), `a home aponta para ${tamanho}.${formato}`);
    assert.equal(await largura(url), tamanho);
  }
}
assert.ok(home.includes('width="2400" height="3600"'));
assert.ok(home.includes("/static/img/hero/hero-land-960.avif"), "a capa de computador nao foi tocada");

await cartao("hero-retrato").locator('input[type="file"]').setInputFiles({ name: "x.jpg", mimeType: "image/jpeg", buffer: Buffer.from("<html>") });
await pagina.locator('.adm-aviso[role="alert"]', { hasText: "não é uma imagem" }).waitFor();
assert.equal(slot("hero-retrato").base, capa.base, "arquivo invalido nao substitui a capa");

await cartao("hero-retrato").locator('input[type="file"]').setInputFiles({ name: "capa2.jpg", mimeType: "image/jpeg", buffer: await jpeg(1300, 1900, "#115533") });
await esperar(() => slot("hero-retrato").base !== capa.base, "a capa e trocada de novo");
assert.equal((await fetch(`${capa.base}-640.avif`)).status, 404, "a capa anterior e apagada do armazenamento");
assert.equal(await largura(`${slot("hero-retrato").base}-1920.avif`), 1300, "foto menor que a largura pedida nao e ampliada");

await cartao("hero-retrato").getByRole("button", { name: "Voltar à original" }).click();
await esperar(() => slot("hero-retrato") === undefined, "a capa volta a original");
assert.ok((await html("/")).includes("/static/img/hero/hero-640.avif"));

await cartao("card-midiakit").locator('input[type="file"]').setInputFiles({ name: "kit.png", mimeType: "image/png", buffer: await sharp({ create: { width: 2000, height: 3000, channels: 3, background: "#333366" } }).png().toBuffer() });
await esperar(() => slot("card-midiakit") !== undefined, "o cartao do midia kit e trocado");
assert.ok((await html("/")).includes(`${slot("card-midiakit").base}.jpg`));
assert.equal(await largura(`${slot("card-midiakit").base}.jpg`), 1080);

await cartao("sobre-bio-1").locator('input[type="file"]').setInputFiles({ name: "bio.jpg", mimeType: "image/jpeg", buffer: await jpeg(1600, 2000, "#664422") });
await esperar(() => slot("sobre-bio-1") !== undefined, "a foto do Sobre e trocada");
const bio = slot("sobre-bio-1");
assert.deepEqual([bio.w, bio.h], [800, 1000], "largura e altura gravadas sao as da variante");
const sobre = await html("/sobre");
assert.ok(sobre.includes(`srcSet="${bio.base}.avif"`) || sobre.includes(`srcset="${bio.base}.avif"`));
assert.ok(sobre.includes('width="800" height="1000"'));

await pagina.locator('label:has-text("Adicionar ao carrossel") input').setInputFiles({ name: "nova.jpg", mimeType: "image/jpeg", buffer: await jpeg(1200, 900, "#226644") });
await esperar(() => carrossel().length === 9, "a foto entra no carrossel");
const nova = carrossel().at(-1);
assert.equal((await html("/sobre")).split('class="carrossel__item"').length - 1, 9);
assert.equal(await largura(`${nova.base}.webp`), 800);

await pagina.locator(`[data-carrossel="${nova.id}"]`).getByRole("button", { name: "Mover para antes" }).click();
await esperar(() => carrossel().at(-2).id === nova.id, "a foto sobe uma posicao");
const original = carrossel()[0];
await pagina.locator(`[data-carrossel="${original.id}"]`).getByRole("button", { name: "Remover" }).click();
await esperar(() => carrossel().length === 8, "foto original sai do carrossel");
assert.ok(!(await html("/sobre")).includes("SobreMim1.avif"));
assert.equal((await fetch(`${values.base}/static/img/sobre/SobreMim1.avif`)).status, 200, "remover do carrossel nao apaga arquivo do pacote");

await pagina.locator(`[data-carrossel="${nova.id}"]`).getByRole("button", { name: "Remover" }).click();
await esperar(() => carrossel().length === 7, "foto enviada sai do carrossel");
assert.equal((await fetch(`${nova.base}.webp`)).status, 404, "foto enviada e apagada do armazenamento");

assert.deepEqual(erros, [], "sem erros de JavaScript");
await navegador.close();
banco.close();
console.log("site ok");
