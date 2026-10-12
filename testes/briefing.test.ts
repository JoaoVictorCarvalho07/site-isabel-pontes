import assert from "node:assert/strict";
import { test } from "node:test";
import { permitir } from "../lib/briefing/limitador.ts";
import { farejarTipo, nomeParaExibir } from "../lib/briefing/tipos.ts";
import { contatoValido, validar } from "../lib/briefing/validar.ts";

const bytes = (texto: string) => new TextEncoder().encode(texto);
const ftyp = (marca: string) => bytes(`\0\0\0\x18ftyp${marca}\0\0\0\0`);

test("aceita os formatos de imagem de celular e de computador", () => {
  assert.equal(farejarTipo(new Uint8Array([0xff, 0xd8, 0xff, 0xe0]))?.extensao, ".jpg");
  assert.equal(farejarTipo(new Uint8Array([0x89, 0x50, 0x4e, 0x47, 0x0d, 0x0a, 0x1a, 0x0a]))?.extensao, ".png");
  assert.equal(farejarTipo(bytes("GIF89a..."))?.mime, "image/gif");
  assert.equal(farejarTipo(bytes("RIFF\0\0\0\0WEBPVP8 "))?.mime, "image/webp");
  assert.equal(farejarTipo(ftyp("heic"))?.mime, "image/heic");
  assert.equal(farejarTipo(ftyp("mif1"))?.mime, "image/heif");
  assert.equal(farejarTipo(ftyp("avif"))?.mime, "image/avif");
});

test("recusa o que nao e imagem, mesmo disfarcado", () => {
  assert.equal(farejarTipo(bytes("<html><script>alert(1)</script>")), null);
  assert.equal(farejarTipo(bytes("#!/bin/sh\nrm -rf /")), null);
  assert.equal(farejarTipo(bytes("%PDF-1.7")), null);
  assert.equal(farejarTipo(bytes('<svg xmlns="http://www.w3.org/2000/svg">')), null);
  assert.equal(farejarTipo(ftyp("isom")), null);
  assert.equal(farejarTipo(new Uint8Array()), null);
});

test("o nome enviado nunca vira caminho", () => {
  assert.equal(nomeParaExibir("../../etc/passwd"), "passwd");
  assert.equal(nomeParaExibir("C:\\Users\\x\\foto.jpg"), "foto.jpg");
  assert.equal(nomeParaExibir(".."), "imagem");
  assert.equal(nomeParaExibir("a\nb.png"), "ab.png");
  assert.equal(nomeParaExibir("x".repeat(300)).length, 120);
});

test("contato aceita e-mail ou telefone brasileiro", () => {
  for (const valido of ["a@b.co", "(41) 99999-8888", "4133334444", "5541999998888", "554133334444"]) {
    assert.equal(contatoValido(valido), true, valido);
  }
  for (const invalido of ["a@b", "123", "994199999888", "99941999998888", "@x.com"]) {
    assert.equal(contatoValido(invalido), false, invalido);
  }
});

const form = {
  slug: "x",
  titulo: ["", ""],
  subtitulo: ["", ""],
  enviar: ["", ""],
  sucesso: ["", ""],
  perguntas: [
    { id: 1, chave: "nome", label: ["", ""], ajuda: ["", ""], tipo: "text", obrigatoria: true, opcoes: [] },
    { id: 2, chave: "email", label: ["", ""], ajuda: ["", ""], tipo: "email", obrigatoria: false, opcoes: [] },
    { id: 3, chave: "tipo", label: ["", ""], ajuda: ["", ""], tipo: "select", obrigatoria: true, opcoes: [["Ensaio", "Shoot"]] },
    { id: 4, chave: "fotos", label: ["", ""], ajuda: ["", ""], tipo: "file", obrigatoria: false, opcoes: [] },
  ],
} as const;

const responder = (dados: Record<string, string>) => validar(form as never, (chave) => dados[chave] ?? "");

test("validacao: obrigatorio, e-mail, opcao da lista e corte de tamanho", () => {
  assert.deepEqual(responder({ nome: " Ana ", tipo: "Shoot" }), { valores: { nome: "Ana", email: "", tipo: "Shoot" }, erros: {} });
  assert.deepEqual(responder({ tipo: "Ensaio" }).erros, { nome: "form.obrigatorio" });
  assert.deepEqual(responder({ nome: "Ana", tipo: "Outro" }).erros, { tipo: "form.opcaoInvalida" });
  assert.deepEqual(responder({ nome: "Ana", tipo: "Ensaio", email: "ana@" }).erros, { email: "form.emailInvalido" });
  assert.equal(responder({ nome: "x".repeat(5000), tipo: "Ensaio" }).valores.nome.length, 4000);
});

test("limite de um envio a cada 30 segundos por IP", () => {
  assert.equal(permitir("1.1.1.1", 1_000), true);
  assert.equal(permitir("1.1.1.1", 20_000), false);
  assert.equal(permitir("2.2.2.2", 20_000), true);
  assert.equal(permitir("1.1.1.1", 32_000), true);
});
