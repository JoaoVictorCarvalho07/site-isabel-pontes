import assert from "node:assert/strict";
import { parseArgs } from "node:util";
import { chromium } from "playwright";

const { values } = parseArgs({ options: { base: { type: "string", default: "http://127.0.0.1:8098" } } });

const navegador = await chromium.launch();
const pagina = await navegador.newPage({ viewport: { width: 1100, height: 800 } });
const erros = [];
pagina.on("pageerror", (erro) => erros.push(erro.message));

await pagina.goto(`${values.base}/galeria`, { waitUntil: "networkidle" });
const total = await pagina.locator(".masonry__item").count();
assert.ok(total > 100, "a galeria completa lista as fotos");

await pagina.evaluate(() => {
  window.__marca = true;
  window.scrollTo(0, 300);
});
await pagina.getByRole("button", { name: "Casal", exact: true }).evaluate((botao) => botao.click());
await pagina.waitForURL("**/galeria?tema=casal");
await pagina.waitForFunction((antes) => document.querySelectorAll(".masonry__item").length < antes, total);

const filtrados = await pagina.locator(".masonry__item").count();
assert.equal(await pagina.locator('.masonry__item:not([data-tema-label="Casal"])').count(), 0, "o filtro nao vaza outro tema");
assert.equal(await pagina.evaluate(() => window.__marca), true, "o filtro nao recarrega o documento");
assert.equal(await pagina.evaluate(() => window.scrollY), 300, "o filtro nao rola a pagina");
assert.equal(await pagina.locator(".filtro--ativo").textContent(), "Casal");
assert.match(await pagina.title(), /^Casal — Galeria/);
await pagina.waitForFunction(() => document.querySelector(".masonry__img")?.classList.contains("is-loaded"));

await pagina.locator(".masonry__item").nth(1).click();
const lightbox = pagina.locator("#lightbox");
assert.equal(await lightbox.evaluate((d) => d.open), true, "o clique abre o lightbox");
assert.equal(await pagina.locator("[data-lb-count]").textContent(), `2 / ${filtrados}`);
const segunda = await pagina.locator("[data-lb-img]").getAttribute("src");
assert.equal(segunda, await pagina.locator(".masonry__item").nth(1).getAttribute("data-full"));

await pagina.keyboard.press("ArrowRight");
assert.equal(await pagina.locator("[data-lb-count]").textContent(), `3 / ${filtrados}`);
await pagina.keyboard.press("ArrowLeft");
await pagina.keyboard.press("ArrowLeft");
await pagina.keyboard.press("ArrowLeft");
assert.equal(await pagina.locator("[data-lb-count]").textContent(), `${filtrados} / ${filtrados}`, "a navegacao e circular");
await pagina.keyboard.press("Escape");
assert.equal(await lightbox.evaluate((d) => d.open), false, "Esc fecha o lightbox");

await pagina.goBack();
await pagina.waitForFunction((esperado) => document.querySelectorAll(".masonry__item").length === esperado, total);
assert.equal(await pagina.locator(".filtro--ativo").textContent(), "Todos");

await pagina.locator(".masonry__item").first().click();
assert.equal(await lightbox.evaluate((d) => d.open), true, "o lightbox funciona depois do voltar");
assert.equal(await pagina.locator("[data-lb-count]").textContent(), `1 / ${total}`);

assert.deepEqual(erros, [], "sem erros de JavaScript");
await navegador.close();
console.log(`galeria ok: ${total} fotos, ${filtrados} em Casal`);
