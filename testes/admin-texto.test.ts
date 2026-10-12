import assert from "node:assert/strict";
import { test } from "node:test";
import { casar, opcoesDoTexto, trechosComLinks } from "../lib/admin/texto.ts";

test("so http e https viram link, e a pontuacao final fica de fora", () => {
  assert.deepEqual(trechosComLinks("veja https://pin.it/abc, e http://x.co/y?z=1."), [
    { texto: "veja " },
    { texto: "https://pin.it/abc", href: "https://pin.it/abc" },
    { texto: "," },
    { texto: " e " },
    { texto: "http://x.co/y?z=1", href: "http://x.co/y?z=1" },
    { texto: "." },
  ]);
  assert.deepEqual(trechosComLinks("javascript:alert(1) <script>x</script>"), [{ texto: "javascript:alert(1) <script>x</script>" }]);
  assert.deepEqual(trechosComLinks(""), []);
});

test("opcoes: uma por linha, mantendo a traducao das que nao mudaram", () => {
  const anteriores = [
    { pt: "Ensaio", en: "Shoot" },
    { pt: "Evento", en: "Event" },
  ];
  assert.deepEqual(opcoesDoTexto(" Ensaio \n\nCasamento\n", anteriores), [
    { pt: "Ensaio", en: "Shoot" },
    { pt: "Casamento", en: "" },
  ]);
  assert.deepEqual(opcoesDoTexto("  \n "), []);
});

test("mudar o portugues limpa o ingles; manter preserva", () => {
  assert.deepEqual(casar("Nome", "Name", "Nome"), ["Nome", "Name"]);
  assert.deepEqual(casar("Nome", "Name", "Seu nome"), ["Seu nome", ""]);
});
