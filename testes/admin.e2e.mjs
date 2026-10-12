import assert from "node:assert/strict";
import { existsSync } from "node:fs";
import { parseArgs } from "node:util";
import Database from "better-sqlite3";
import { encode } from "next-auth/jwt";
import { chromium } from "playwright";
import sharp from "sharp";

const { values } = parseArgs({ options: { base: { type: "string", default: "http://127.0.0.1:8098" } } });
const SEGREDO = "segredo-apenas-para-testes-locais-0123456789";
const COOKIE = "authjs.session-token";

const banco = new Database("data/teste.db", { readonly: true });
const foto = (id) => banco.prepare("SELECT * FROM photos WHERE id = ?").get(id);
const total = () => banco.prepare("SELECT count(*) AS n FROM photos").get().n;

async function sessao(email) {
  const valor = await encode({ token: { email, name: "Teste", sub: "1" }, secret: SEGREDO, salt: COOKIE });
  return { name: COOKIE, value: valor, url: values.base, httpOnly: true };
}

async function imagem(largura, altura, cor, orientacao) {
  const base = sharp({ create: { width: largura, height: altura, channels: 3, background: cor } }).jpeg({ quality: 90 });
  return (orientacao ? base.withMetadata({ orientation: orientacao }) : base).toBuffer();
}

const navegador = await chromium.launch();

{
  const anonimo = await navegador.newContext();
  const pagina = await anonimo.newPage();
  for (const caminho of ["/admin", "/admin/fotos", "/admin/temas", "/admin/lixeira"]) {
    await pagina.goto(values.base + caminho);
    assert.equal(new URL(pagina.url()).pathname, "/admin/entrar", `${caminho} exige login`);
  }
  const envio = await anonimo.request.post(`${values.base}/admin/api/fotos`, { multipart: { tema: "casal", arquivo: { name: "a.jpg", mimeType: "image/jpeg", buffer: await imagem(10, 10, "#000") } } });
  assert.equal(envio.status(), 401, "upload sem login e recusado");
  assert.equal(total(), 190);
  await anonimo.close();

  const intruso = await navegador.newContext();
  await intruso.addCookies([await sessao("outra@pessoa.dev")]);
  const paginaIntruso = await intruso.newPage();
  await paginaIntruso.goto(`${values.base}/admin/fotos`);
  assert.equal(new URL(paginaIntruso.url()).pathname, "/admin/entrar", "e-mail fora da lista nao entra");
  await intruso.close();

  const semJS = await navegador.newContext({ javaScriptEnabled: false });
  const paginaSemJS = await semJS.newPage();
  await paginaSemJS.goto(`${values.base}/admin/entrar`);
  await paginaSemJS.getByRole("button", { name: "Entrar com Google" }).click({ noWaitAfter: true });
  await paginaSemJS.waitForTimeout(3000);
  const vivo = await fetch(`${values.base}/sobre`, { signal: AbortSignal.timeout(8000) });
  assert.equal(vivo.status, 200, "o botao de entrar sem JavaScript nao trava o servidor");
  await semJS.close();
}

const contexto = await navegador.newContext({ viewport: { width: 390, height: 844 } });
await contexto.addCookies([await sessao("isa@teste.dev")]);
const pagina = await contexto.newPage();
const erros = [];
pagina.on("pageerror", (erro) => erros.push(erro.message));
pagina.on("dialog", (dialogo) => dialogo.accept());

const galeria = async (tema = "") => {
  const html = await (await fetch(`${values.base}/galeria${tema ? `?tema=${tema}` : ""}`)).text();
  return [...html.matchAll(/data-full="([^"]+)" data-alt="([^"]*)"/g)].map(([, full, alt]) => ({ full, alt }));
};
const cartao = (id) => pagina.locator(`[data-foto="${id}"]`);
const esperarBanco = async (condicao, mensagem) => {
  for (let i = 0; i < 50; i++) {
    if (condicao()) return;
    await pagina.waitForTimeout(100);
  }
  assert.fail(mensagem);
};

await pagina.goto(`${values.base}/admin/fotos`, { waitUntil: "networkidle" });
assert.match(await pagina.locator(".adm-titulo").textContent(), /190/);
assert.equal(await pagina.evaluate(() => document.documentElement.scrollWidth <= window.innerWidth), true, "sem rolagem horizontal no celular");

await pagina.locator(".adm-envio select").selectOption("casal");
await pagina.locator('.adm-envio input[type="file"]').setInputFiles([
  { name: "retrato.jpg", mimeType: "image/jpeg", buffer: await imagem(2000, 3000, "#884422") },
  { name: "falsa.jpg", mimeType: "image/jpeg", buffer: Buffer.from("<html>nao sou imagem</html>") },
  { name: "deitada.jpg", mimeType: "image/jpeg", buffer: await imagem(3000, 2000, "#224488", 6) },
]);
await pagina.locator(".adm-botao--principal", { hasText: "Adicionar fotos" }).waitFor({ timeout: 120000 });
const fila = await pagina.locator(".adm-fila li").allTextContents();
assert.equal(fila.filter((linha) => linha.includes("pronta")).length, 2, "duas fotos entraram");
assert.match(fila.find((linha) => linha.includes("falsa.jpg")), /não é uma imagem/);
assert.equal(total(), 192);
await pagina.locator(".adm-titulo", { hasText: "192" }).waitFor();

const [retrato, deitada] = banco.prepare("SELECT * FROM photos WHERE id > 190 ORDER BY id").all();
assert.deepEqual([retrato.w, retrato.h], [2000, 3000]);
assert.deepEqual([deitada.w, deitada.h], [2000, 3000], "a orientacao EXIF e aplicada");
assert.equal(retrato.tema_slug, "casal");
assert.match(retrato.grid, /\/dev-r2\/variants\/grid400\/casal\/[a-z0-9-]+\.avif$/);
assert.match(retrato.lqip, /^data:image\/webp;base64,/);
assert.match(retrato.original, /^originais\/casal\/.+\.jpg$/);
assert.ok(existsSync(`data/teste-r2/originais/${retrato.original}`), "o original e guardado");
for (const url of [retrato.grid, retrato.grid2x, retrato.full]) {
  const resposta = await fetch(url);
  assert.equal(resposta.status, 200);
  assert.equal(resposta.headers.get("content-type"), "image/avif");
  const dimensoes = await sharp(Buffer.from(await resposta.arrayBuffer())).metadata();
  assert.equal(dimensoes.width, Number(url.match(/grid(\d+)|full(\d+)/).slice(1).find(Boolean)));
}
assert.equal((await fetch(`${values.base}/dev-r2/${retrato.original}`)).status, 404, "o original nao e servido ao publico");
assert.equal((await galeria("casal")).length, 17, "as fotos novas aparecem na galeria na hora");

await cartao(retrato.id).getByLabel("Descrição da foto").fill("Casal ao pôr do sol");
await cartao(retrato.id).getByLabel("Descrição da foto").blur();
await esperarBanco(() => foto(retrato.id).alt === "Casal ao pôr do sol", "a descricao e gravada");
assert.ok((await galeria("casal")).some((item) => item.alt === "Casal ao pôr do sol"));

await cartao(deitada.id).getByLabel("Tema").selectOption("epoca");
await esperarBanco(() => foto(deitada.id).tema_slug === "epoca", "o tema e trocado");
assert.equal((await galeria("casal")).length, 16);

await cartao(deitada.id).getByRole("button", { name: "Ocultar" }).click();
await esperarBanco(() => foto(deitada.id).published === 0, "a foto e ocultada");
assert.ok(!(await galeria()).some((item) => item.full === deitada.full), "foto oculta some do site");
await cartao(deitada.id).getByRole("button", { name: "Publicar" }).click();
await esperarBanco(() => foto(deitada.id).published === 1, "a foto volta a ser publicada");

await cartao(retrato.id).getByRole("button", { name: "Fixar no topo" }).click();
await esperarBanco(() => foto(retrato.id).sort_order === 1, "a primeira fixada recebe a posicao 1");
await cartao(deitada.id).getByRole("button", { name: "Fixar no topo" }).click();
await esperarBanco(() => foto(deitada.id).sort_order === 2, "a segunda fixada recebe a posicao 2");
assert.deepEqual((await galeria()).slice(0, 2).map((item) => item.full), [retrato.full, deitada.full], "as fixadas abrem a galeria, na ordem");
await cartao(deitada.id).getByRole("button", { name: "Mover para antes" }).click();
await esperarBanco(() => foto(deitada.id).sort_order === 1 && foto(retrato.id).sort_order === 2, "a ordem das fixadas e trocada");
assert.deepEqual((await galeria()).slice(0, 2).map((item) => item.full), [deitada.full, retrato.full]);

await cartao(deitada.id).getByRole("button", { name: "Lixeira" }).click();
await esperarBanco(() => foto(deitada.id).excluida_em !== null, "a foto vai para a lixeira");
assert.ok(!(await galeria()).some((item) => item.full === deitada.full), "foto na lixeira some do site");
assert.equal((await fetch(deitada.full)).status, 200, "a lixeira nao apaga os arquivos");

await pagina.goto(`${values.base}/admin/lixeira`, { waitUntil: "networkidle" });
assert.equal(await pagina.locator("[data-foto]").count(), 1);
await cartao(deitada.id).getByRole("button", { name: "Restaurar" }).click();
await esperarBanco(() => foto(deitada.id).excluida_em === null, "a foto e restaurada");
assert.ok((await galeria()).some((item) => item.full === deitada.full));

await pagina.goto(`${values.base}/admin/fotos`, { waitUntil: "networkidle" });
await cartao(deitada.id).getByRole("button", { name: "Lixeira" }).click();
await esperarBanco(() => foto(deitada.id).excluida_em !== null, "a foto volta para a lixeira");
await pagina.goto(`${values.base}/admin/lixeira`, { waitUntil: "networkidle" });
await cartao(deitada.id).getByRole("button", { name: "Apagar de vez" }).click();
await esperarBanco(() => foto(deitada.id) === undefined, "a foto e apagada do banco");
assert.equal((await fetch(deitada.full)).status, 404, "apagar de vez remove as variantes");
assert.ok(!existsSync(`data/teste-r2/originais/${deitada.original}`), "apagar de vez remove o original");
assert.equal(total(), 191);

await pagina.goto(`${values.base}/admin/temas`, { waitUntil: "networkidle" });
await pagina.getByLabel("Nome do novo tema").fill("Família & Pets");
await pagina.getByRole("button", { name: "Criar" }).click();
await pagina.locator('[data-tema="familia-pets"]').waitFor();
const filtros = async () => [...(await (await fetch(`${values.base}/galeria`)).text()).matchAll(/class="filtro[^"]*"[^>]*>([^<]+)</g)].map((m) => m[1].replaceAll("&amp;", "&"));
assert.equal((await filtros()).at(-1), "Família & Pets", "o tema novo entra no fim do filtro");

await pagina.locator('[data-tema="familia-pets"]').getByRole("button", { name: "Mover para antes" }).click();
await esperarBanco(() => banco.prepare("SELECT sort_order FROM temas WHERE slug = 'familia-pets'").get().sort_order === 70, "o tema sobe uma posicao");
assert.equal((await filtros()).at(-2), "Família & Pets");

await pagina.locator('[data-tema="familia-pets"] input').fill("Família");
await pagina.locator('[data-tema="familia-pets"] input').blur();
await esperarBanco(() => banco.prepare("SELECT label FROM temas WHERE slug = 'familia-pets'").get().label === "Família", "o tema e renomeado");

await pagina.locator('[data-tema="familia-pets"]').getByRole("button", { name: "Ocultar do site" }).click();
await esperarBanco(() => banco.prepare("SELECT visible FROM temas WHERE slug = 'familia-pets'").get().visible === 0, "o tema e ocultado");
assert.ok(!(await filtros()).includes("Família"), "tema oculto some do filtro");
assert.equal((await fetch(`${values.base}/galeria?tema=familia-pets`)).status, 404);

await pagina.getByLabel("Nome do novo tema").fill("Casal");
await pagina.getByRole("button", { name: "Criar" }).click();
await pagina.locator(".adm-aviso").waitFor();
assert.equal(banco.prepare("SELECT count(*) AS n FROM temas").get().n, 8, "tema repetido nao e criado");

assert.deepEqual(erros, [], "sem erros de JavaScript");
await navegador.close();
banco.close();
console.log("admin ok");
