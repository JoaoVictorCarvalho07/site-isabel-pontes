import assert from "node:assert/strict";
import { existsSync, readdirSync } from "node:fs";
import { parseArgs } from "node:util";
import Database from "better-sqlite3";
import { chromium } from "playwright";

const { values } = parseArgs({ options: { base: { type: "string", default: "http://127.0.0.1:8098" } } });

const banco = new Database("data/teste.db", { readonly: true });
const respostas = () => banco.prepare("SELECT id, form, lang, payload, ip FROM submissions ORDER BY id").all();
const anexos = () => banco.prepare("SELECT campo, caminho, nome_original, mime FROM submission_files ORDER BY id").all();

const JPEG = Buffer.concat([Buffer.from([0xff, 0xd8, 0xff, 0xe0]), Buffer.alloc(2048, 1)]);
const arquivo = (name, buffer, mimeType = "image/jpeg") => ({ name, mimeType, buffer });

const navegador = await chromium.launch();
let ip = 10;

async function abrir(caminho, opcoes = {}) {
  const contexto = await navegador.newContext({ extraHTTPHeaders: { "X-Forwarded-For": `10.0.0.${ip++}` }, ...opcoes });
  const pagina = await contexto.newPage();
  await pagina.goto(values.base + caminho, { waitUntil: "networkidle" });
  return pagina;
}

async function preencherFotografia(pagina) {
  await pagina.locator('[name="nome"]').fill("Ana Teste");
  await pagina.locator('[name="contato"]').fill("(41) 99999-8888");
  await pagina.locator('[name="tipo"]').selectOption({ index: 1 });
  await pagina.locator('[name="motivacao"]').fill("Quero um ensaio de época.\nSegunda linha.");
}

{
  const pagina = await abrir("/briefing");
  await pagina.locator('[name="nome"]').fill("Ana Teste");
  await pagina.locator('[name="contato"]').fill("123");
  await pagina.locator("form.form-briefing__form").evaluate((f) => f.setAttribute("novalidate", ""));
  await pagina.locator(".form-briefing__enviar").click();
  await pagina.locator(".form-briefing__erro-geral").waitFor();
  assert.equal(await pagina.locator(".form-briefing__erro-geral").textContent(), "Confere os campos marcados abaixo.");
  assert.equal(await pagina.locator('[name="nome"]').inputValue(), "Ana Teste", "o valor digitado e preservado");
  assert.equal(await pagina.locator('[name="contato"]').getAttribute("aria-invalid"), "true");
  assert.match(await pagina.locator("#campo-fotografia-contato-erro").textContent(), /WhatsApp com DDD/);
  assert.match(await pagina.locator("#campo-fotografia-tipo-erro").textContent(), /obrigatório/);
  assert.equal(await pagina.locator(".campo__realerta").count(), 1, "avisa que os anexos precisam ser reenviados");
  assert.equal(respostas().length, 0, "erro de validacao nao grava");

  await preencherFotografia(pagina);
  await pagina.locator('[name="referencias_imagens"]').setInputFiles([
    arquivo("../../segredo.jpg", JPEG),
    arquivo("disfarce.jpg", Buffer.from("<html><script>alert(1)</script></html>")),
    arquivo("desenho.svg", Buffer.from('<svg xmlns="http://www.w3.org/2000/svg"/>'), "image/svg+xml"),
    arquivo("b.jpg", JPEG),
    arquivo("c.jpg", JPEG),
  ]);
  await pagina.locator(".form-briefing__enviar").click();
  await pagina.locator(".form-briefing__sucesso").waitFor();
  assert.match(await pagina.locator(".form-briefing__sucesso").textContent(), /Recebi seu briefing/);

  const [resposta] = respostas();
  assert.equal(resposta.form, "fotografia");
  assert.equal(resposta.lang, "pt");
  assert.equal(resposta.ip, "10.0.0.10");
  const payload = JSON.parse(resposta.payload);
  assert.equal(payload.nome, "Ana Teste");
  assert.equal(payload.motivacao, "Quero um ensaio de época.\nSegunda linha.");

  const salvos = anexos();
  assert.equal(salvos.length, 2, "so as imagens de verdade entre os 4 primeiros arquivos sao gravadas");
  assert.deepEqual(salvos.map((a) => a.nome_original), ["segredo.jpg", "b.jpg"]);
  for (const anexo of salvos) {
    assert.match(anexo.caminho, /^\d{4}\/\d{2}\/\d+-[0-9a-f]{16}\.jpg$/, "nome gerado, dentro da pasta de anexos");
    assert.equal(anexo.mime, "image/jpeg");
    assert.ok(existsSync(`data/teste-uploads/${anexo.caminho}`));
  }
  assert.ok(!existsSync("data/segredo.jpg") && !existsSync("segredo.jpg"));
  await pagina.context().close();
}

{
  const contexto = await navegador.newContext({ extraHTTPHeaders: { "X-Forwarded-For": "10.0.0.99" } });
  const pagina = await contexto.newPage();
  for (const esperado of [".form-briefing__sucesso", ".form-briefing__erro-geral"]) {
    await pagina.goto(`${values.base}/briefing`, { waitUntil: "networkidle" });
    await preencherFotografia(pagina);
    await pagina.locator(".form-briefing__enviar").click();
    await pagina.locator(esperado).waitFor();
  }
  assert.match(await pagina.locator(".form-briefing__erro-geral").textContent(), /Espera um minutinho/);
  assert.equal(respostas().length, 2, "o segundo envio do mesmo IP nao grava");
  await contexto.close();
}

{
  const pagina = await abrir("/briefing");
  await preencherFotografia(pagina);
  await pagina.locator('[name="website"]').evaluate((campo) => (campo.value = "http://spam.example"));
  await pagina.locator(".form-briefing__enviar").click();
  await pagina.locator(".form-briefing__sucesso").waitFor();
  assert.equal(respostas().length, 2, "honeypot responde sucesso sem gravar");
  await pagina.context().close();
}

{
  const pagina = await abrir("/en/midiakit");
  await pagina.locator('[name="marca"]').fill("Acme");
  await pagina.locator('[name="responsavel"]').fill("Bob");
  await pagina.locator('[name="email"]').fill("bob@acme.com");
  await pagina.locator('[name="produto"]').fill("Shampoo");
  await pagina.locator('[name="objetivo"]').selectOption("Product launch");
  await pagina.locator('[name="pacote"]').selectOption({ index: 1 });
  const antes = await pagina.evaluate(() => (window.__marca = true));
  await pagina.locator(".form-briefing__enviar").click();
  await pagina.locator(".form-briefing__sucesso").waitFor();
  assert.match(await pagina.locator(".form-briefing__sucesso").textContent(), /Briefing received/);
  assert.equal(await pagina.evaluate(() => window.__marca), antes, "o envio nao recarrega o midia kit");
  const ultima = respostas().at(-1);
  assert.equal(ultima.form, "ugc");
  assert.equal(ultima.lang, "en");
  assert.equal(JSON.parse(ultima.payload).objetivo, "Product launch");
  await pagina.context().close();
}

{
  const pagina = await abrir("/briefing", { javaScriptEnabled: false });
  await pagina.locator('[name="nome"]').fill("Sem JS");
  await pagina.locator('[name="contato"]').fill("semjs@exemplo.com");
  await pagina.locator('[name="tipo"]').selectOption({ index: 1 });
  await pagina.locator('[name="motivacao"]').fill("Funciona sem JavaScript.");
  await pagina.locator(".form-briefing__enviar").click();
  await pagina.locator(".form-briefing__sucesso").waitFor();
  assert.ok(await pagina.locator("header.navbar").count(), "sem JS a resposta e a pagina inteira");
  assert.equal(JSON.parse(respostas().at(-1).payload).nome, "Sem JS");
  await pagina.context().close();
}

{
  const robots = await (await fetch(`${values.base}/robots.txt`)).text();
  assert.match(robots, /Disallow: \/admin/);
  assert.match(robots, /Sitemap: http.*\/sitemap\.xml/);
  const sitemap = await (await fetch(`${values.base}/sitemap.xml`)).text();
  for (const trecho of ["/galeria?tema=casal<", "/en/galeria?tema=casal<", "/briefing<", "/en/midiakit<", "/sobre<"]) {
    assert.ok(sitemap.replaceAll("&amp;", "&").includes(trecho), `sitemap tem ${trecho}`);
  }
  assert.ok(!sitemap.includes("/admin"));
}

assert.equal(readdirSync("data/teste-uploads", { recursive: true }).filter((f) => String(f).endsWith(".jpg")).length, 2);
await navegador.close();
banco.close();
console.log("briefing ok");
