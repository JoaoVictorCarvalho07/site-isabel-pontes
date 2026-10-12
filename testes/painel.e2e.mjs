import assert from "node:assert/strict";
import { parseArgs } from "node:util";
import Database from "better-sqlite3";
import { encode } from "next-auth/jwt";
import { chromium } from "playwright";

const { values } = parseArgs({ options: { base: { type: "string", default: "http://127.0.0.1:8098" } } });
const SEGREDO = "segredo-apenas-para-testes-locais-0123456789";
const COOKIE = "authjs.session-token";

const banco = new Database("data/teste.db", { readonly: true });
const pergunta = (chave) => banco.prepare("SELECT * FROM form_questions WHERE form_slug = 'fotografia' AND chave = ?").get(chave);
const ordem = () => banco.prepare("SELECT chave FROM form_questions WHERE form_slug = 'fotografia' ORDER BY sort_order, id").all().map((l) => l.chave);

const navegador = await chromium.launch();
const JPEG = Buffer.concat([Buffer.from([0xff, 0xd8, 0xff, 0xe0]), Buffer.alloc(3000, 7)]);

const visitante = await navegador.newContext({ extraHTTPHeaders: { "X-Forwarded-For": `10.7.${Date.now() % 250}.${Math.floor(Date.now() / 250) % 250}` } });
const publico = await visitante.newPage();
await publico.goto(`${values.base}/briefing`, { waitUntil: "networkidle" });
await publico.locator('[name="nome"]').fill("Marina Painel");
await publico.locator('[name="contato"]').fill("marina@exemplo.com");
await publico.locator('[name="tipo"]').selectOption({ index: 1 });
await publico.locator('[name="motivacao"]').fill("Refs em https://pin.it/abc123. <b>negrito</b> javascript:alert(1)");
await publico.locator('[name="observacoes"]').fill("Pão de queijo");
await publico.locator('[name="referencias_imagens"]').setInputFiles([{ name: "ref.jpg", mimeType: "image/jpeg", buffer: JPEG }]);
await publico.locator(".form-briefing__enviar").click();
await publico.locator(".form-briefing__sucesso").waitFor();
const idBriefing = banco.prepare("SELECT max(id) AS id FROM submissions").get().id;
const idAnexo = banco.prepare("SELECT id FROM submission_files WHERE submission_id = ?").get(idBriefing).id;

for (const caminho of ["/admin/briefings", "/admin/formularios", "/admin/formularios/fotografia"]) {
  await publico.goto(values.base + caminho);
  assert.equal(new URL(publico.url()).pathname, "/admin/entrar", `${caminho} exige login`);
}
assert.equal((await visitante.request.get(`${values.base}/admin/arquivo/${idAnexo}`)).status(), 404, "anexo nao e servido sem login");
await visitante.close();

const contexto = await navegador.newContext({ viewport: { width: 390, height: 844 } });
const valor = await encode({ token: { email: "isa@teste.dev", sub: "1" }, secret: SEGREDO, salt: COOKIE });
await contexto.addCookies([{ name: COOKIE, value: valor, url: values.base, httpOnly: true }]);
const pagina = await contexto.newPage();
const erros = [];
pagina.on("pageerror", (erro) => erros.push(erro.message));
pagina.on("dialog", (dialogo) => dialogo.accept());

const esperar = async (condicao, mensagem) => {
  for (let i = 0; i < 50; i++) {
    if (condicao()) return;
    await pagina.waitForTimeout(100);
  }
  assert.fail(mensagem);
};

await pagina.goto(`${values.base}/admin/briefings`, { waitUntil: "networkidle" });
assert.equal(await pagina.evaluate(() => document.documentElement.scrollWidth <= window.innerWidth), true, "sem rolagem horizontal no celular");
const cartao = pagina.locator(`[data-briefing="${idBriefing}"]`);
assert.equal(await cartao.locator("h2").textContent(), "Marina Painel", "o destaque e o primeiro texto obrigatorio");
const naoLidas = banco.prepare("SELECT count(*) AS n FROM submissions WHERE lida = 0").get().n;
assert.ok((await pagina.locator(".adm-titulo").textContent()).includes(`${naoLidas} não lida`));
assert.match(await cartao.locator("dl").textContent(), /Mais alguma coisa/, "a resposta aparece com o rotulo da pergunta");
const links = await cartao.locator("dd a").evaluateAll((as) => as.map((a) => [a.getAttribute("href"), a.rel]));
assert.deepEqual(links, [["https://pin.it/abc123", "noopener noreferrer nofollow"]], "so o endereco http vira link");
assert.equal(await cartao.locator("dd b").count(), 0, "HTML da resposta nao e interpretado");
assert.match(await cartao.locator("dl").textContent(), /<b>negrito<\/b> javascript:alert\(1\)/);

const html = await pagina.content();
assert.ok(!html.includes("teste-uploads") && !/\d{4}\/\d{2}\/\d+-[0-9a-f]{16}/.test(html), "o caminho em disco nao aparece");
assert.match(await cartao.locator(".adm-briefing__anexos").textContent(), /ref\.jpg · 3 KB/);
const anexo = await contexto.request.get(`${values.base}/admin/arquivo/${idAnexo}`);
assert.equal(anexo.status(), 200);
assert.equal(anexo.headers()["content-type"], "image/jpeg");
assert.equal(anexo.headers()["x-content-type-options"], "nosniff");
assert.match(anexo.headers()["content-security-policy"], /sandbox/);
assert.equal((await anexo.body()).length, JPEG.length);
assert.equal((await contexto.request.get(`${values.base}/admin/arquivo/999999`)).status(), 404);

await cartao.getByRole("button", { name: "Marcar como lida" }).click();
await esperar(() => banco.prepare("SELECT lida FROM submissions WHERE id = ?").get(idBriefing).lida === 1, "o briefing e marcado como lido");
await cartao.getByRole("button", { name: "Marcar como não lida" }).waitFor();

await pagina.goto(`${values.base}/admin/briefings?form=ugc`, { waitUntil: "networkidle" });
assert.equal(await pagina.locator(`[data-briefing="${idBriefing}"]`).count(), 0, "o filtro por formulario separa as respostas");

await pagina.goto(`${values.base}/admin/formularios`, { waitUntil: "networkidle" });
assert.equal(await pagina.locator(".adm-temas li").count(), 2);
await pagina.goto(`${values.base}/admin/formularios/fotografia`, { waitUntil: "networkidle" });
assert.equal(await pagina.evaluate(() => document.documentElement.scrollWidth <= window.innerWidth), true);
const cartaoPergunta = (chave) => pagina.locator(`[data-pergunta="${chave}"]`);
const salvo = () => pagina.locator(".adm-aviso--ok").waitFor();
const formPublico = async (caminho = "/briefing") => (await fetch(values.base + caminho)).text();

const comNome = banco.prepare("SELECT payload FROM submissions WHERE form = 'fotografia'").all().filter((l) => JSON.parse(l.payload).nome).length;
assert.ok((await cartaoPergunta("nome").locator(".adm-pergunta__selos").textContent()).includes(`${comNome} resposta`));
assert.match(await cartaoPergunta("data").locator(".adm-pergunta__selos").textContent(), /0 respostas/);

const textos = pagina.getByRole("form", { name: "Textos do formulário" });
await textos.locator('[name="titulo"]').fill("Vamos planejar seu ensaio");
await textos.getByRole("button", { name: "Salvar textos" }).click();
await salvo();
const form = () => banco.prepare("SELECT * FROM forms WHERE slug = 'fotografia'").get();
assert.equal(form().titulo_pt, "Vamos planejar seu ensaio");
assert.equal(form().titulo_en, "", "mudar o titulo em portugues limpa o ingles");
assert.notEqual(form().enviar_en, "", "o que nao mudou mantem a traducao");
assert.ok((await formPublico()).includes("Vamos planejar seu ensaio"), "a edicao aparece no site na hora");
assert.ok((await formPublico("/en/briefing")).includes("Vamos planejar seu ensaio"), "sem ingles, o site em ingles mostra o portugues");

await textos.locator('[name="titulo"]').evaluate((campo) => campo.removeAttribute("required"));
await textos.locator('[name="titulo"]').fill("  ");
await textos.getByRole("button", { name: "Salvar textos" }).click();
await pagina.locator('.adm-aviso[role="alert"]').waitFor();
assert.equal(form().titulo_pt, "Vamos planejar seu ensaio", "titulo vazio e recusado");
await pagina.reload({ waitUntil: "networkidle" });

const antes = pergunta("local");
await cartaoPergunta("local").locator('[name="obrigatoria"]').check();
await cartaoPergunta("local").getByRole("button", { name: "Salvar pergunta" }).click();
await esperar(() => pergunta("local").obrigatoria === 1, "a pergunta vira obrigatoria");
assert.equal(pergunta("local").label_en, antes.label_en, "mexer so em obrigatoria nao toca o ingles");
assert.equal(pergunta("local").en_stale, antes.en_stale);

await cartaoPergunta("local").locator('[name="label"]').fill("Onde vai ser o ensaio?");
await cartaoPergunta("local").getByRole("button", { name: "Salvar pergunta" }).click();
await esperar(() => pergunta("local").label_pt === "Onde vai ser o ensaio?", "o texto da pergunta e salvo");
assert.equal(pergunta("local").label_en, "");
assert.equal(pergunta("local").en_stale, 1, "editar o portugues marca o ingles para revisao");
assert.equal(pergunta("local").chave, "local", "a chave nao muda");

const opcoesAntes = JSON.parse(pergunta("tipo").opcoes_json);
await cartaoPergunta("tipo").locator('form:not(.adm-pergunta__tipo) [name="opcoes"]').fill(`${opcoesAntes[0].pt}\nEnsaio de formatura`);
await cartaoPergunta("tipo").getByRole("button", { name: "Salvar pergunta" }).click();
await esperar(() => JSON.parse(pergunta("tipo").opcoes_json).length === 2, "as opcoes sao salvas");
assert.deepEqual(JSON.parse(pergunta("tipo").opcoes_json), [opcoesAntes[0], { pt: "Ensaio de formatura", en: "" }], "opcao mantida guarda a traducao");

await cartaoPergunta("tipo").locator('form:not(.adm-pergunta__tipo) [name="opcoes"]').fill("   ");
await cartaoPergunta("tipo").getByRole("button", { name: "Salvar pergunta" }).click();
await pagina.locator('.adm-aviso[role="alert"]', { hasText: "pelo menos uma opção" }).waitFor();
assert.equal(JSON.parse(pergunta("tipo").opcoes_json).length, 2, "lista sem opcao e recusada");
await pagina.reload({ waitUntil: "networkidle" });

const ordemAntes = ordem();
await cartaoPergunta("contato").getByRole("button", { name: "Mover para antes" }).click();
await esperar(() => ordem()[0] === "contato", "a pergunta sobe");
assert.deepEqual(ordem().slice(0, 2), [ordemAntes[1], ordemAntes[0]]);
assert.ok((await formPublico()).indexOf('name="contato"') < (await formPublico()).indexOf('name="nome"'), "a ordem nova vale no site");

await cartaoPergunta("observacoes").getByRole("button", { name: "Ocultar do formulário" }).click();
await esperar(() => pergunta("observacoes").oculta === 1, "a pergunta e ocultada");
assert.ok(!(await formPublico()).includes('name="observacoes"'), "pergunta oculta some do formulario publico");
await pagina.goto(`${values.base}/admin/briefings`, { waitUntil: "networkidle" });
assert.match(await pagina.locator(`[data-briefing="${idBriefing}"] dl`).textContent(), /Pão de queijo/, "a resposta antiga continua no painel");
await pagina.goto(`${values.base}/admin/formularios/fotografia`, { waitUntil: "networkidle" });

await cartaoPergunta("nome").locator("summary").click();
assert.equal(await cartaoPergunta("nome").getByRole("button", { name: "Apagar pergunta" }).isDisabled(), true, "pergunta com resposta nao pode ser apagada");
await cartaoPergunta("data").locator("summary").click();
await cartaoPergunta("data").getByRole("button", { name: "Apagar pergunta" }).click();
await esperar(() => pergunta("data") === undefined, "pergunta sem resposta e apagada");

await cartaoPergunta("observacoes").locator("summary").click();
await cartaoPergunta("observacoes").locator(".adm-pergunta__tipo select").selectOption("select");
await cartaoPergunta("observacoes").getByRole("button", { name: "Trocar tipo" }).click();
await pagina.locator('.adm-aviso[role="alert"]', { hasText: "pelo menos uma opção" }).waitFor();
assert.equal(pergunta("observacoes").tipo, "textarea", "trocar para lista sem opcoes e recusado");
await cartaoPergunta("observacoes").locator('.adm-pergunta__tipo [name="opcoes"]').fill("Doce\nSalgado");
await cartaoPergunta("observacoes").getByRole("button", { name: "Trocar tipo" }).click();
await esperar(() => pergunta("observacoes").tipo === "select", "o tipo e trocado");
assert.equal(pergunta("observacoes").chave, "observacoes");
assert.equal(JSON.parse(banco.prepare("SELECT payload FROM submissions WHERE id = ?").get(idBriefing).payload).observacoes, "Pão de queijo", "a resposta antiga e preservada");

const quantas = ordem().length;
const nova = pagina.getByRole("form", { name: "Nova pergunta" });
await nova.locator('[name="label"]').fill("Tem pet que vai participar?");
await nova.locator("select").selectOption("select");
await nova.locator('[name="opcoes"]').fill("Sim\nNão");
await nova.locator('[name="obrigatoria"]').check();
await nova.getByRole("button", { name: "Adicionar pergunta" }).click();
await esperar(() => ordem().length === quantas + 1, "a pergunta nova entra no fim");
const criada = pergunta(ordem().at(-1));
assert.match(criada.chave, /^q_[0-9a-f]{8}$/);
assert.equal(criada.en_stale, 1);
assert.deepEqual(JSON.parse(criada.opcoes_json), [{ pt: "Sim", en: "" }, { pt: "Não", en: "" }]);
assert.ok((await formPublico()).includes("Tem pet que vai participar?"));

assert.deepEqual(erros, [], "sem erros de JavaScript");
await navegador.close();
banco.close();
console.log("painel ok");
