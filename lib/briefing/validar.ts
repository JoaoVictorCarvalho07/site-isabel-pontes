import type { Formulario, Pergunta } from "../forms";

export const TAMANHO_MAXIMO_CAMPO = 4000;

const EMAIL = /^[^@\s]+@[^@\s.]+(\.[^@\s.]+)+$/;

export type Resultado = { valores: Record<string, string>; erros: Record<string, string> };

export function emailValido(valor: string): boolean {
  return EMAIL.test(valor);
}

export function telefoneValido(digitos: string): boolean {
  if (digitos.length === 10 || digitos.length === 11) return true;
  if (digitos.length === 12 || digitos.length === 13) return digitos.startsWith("55");
  return false;
}

export function contatoValido(valor: string): boolean {
  return valor.includes("@") ? emailValido(valor) : telefoneValido(valor.replace(/\D/g, ""));
}

function erroDe(pergunta: Pergunta, valor: string): string | undefined {
  if (pergunta.obrigatoria && valor === "") return "form.obrigatorio";
  if (valor === "") return undefined;
  if (pergunta.tipo === "email" && !emailValido(valor)) return "form.emailInvalido";
  if (pergunta.tipo === "contato" && !contatoValido(valor)) return "form.contatoInvalido";
  if (pergunta.tipo === "select" && !pergunta.opcoes.some(([pt, en]) => valor === pt || valor === en)) return "form.opcaoInvalida";
  return undefined;
}

export function validar(form: Formulario, ler: (chave: string) => string): Resultado {
  const valores: Record<string, string> = {};
  const erros: Record<string, string> = {};
  for (const pergunta of form.perguntas) {
    if (pergunta.tipo === "file") continue;
    const valor = ler(pergunta.chave).trim().slice(0, TAMANHO_MAXIMO_CAMPO);
    valores[pergunta.chave] = valor;
    const erro = erroDe(pergunta, valor);
    if (erro) erros[pergunta.chave] = erro;
  }
  return { valores, erros };
}
