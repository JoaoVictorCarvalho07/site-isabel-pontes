import assert from "node:assert/strict";
import { test } from "node:test";
import { embaralhar } from "../lib/embaralhar.ts";

const itens = Array.from({ length: 190 }, (_, i) => i);

test("mesma semente da a mesma ordem", () => {
  assert.deepEqual(embaralhar(itens, 20261011), embaralhar(itens, 20261011));
});

test("dias diferentes dao ordens diferentes", () => {
  assert.notDeepEqual(embaralhar(itens, 20261011), embaralhar(itens, 20261012));
});

test("embaralhar preserva os itens e nao altera a entrada", () => {
  const resultado = embaralhar(itens, 20261011);
  assert.deepEqual([...resultado].sort((a, b) => a - b), itens);
  assert.notDeepEqual(resultado, itens);
  assert.equal(itens[0], 0);
});
