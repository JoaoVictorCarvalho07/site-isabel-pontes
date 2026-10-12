"use server";

import { signIn, signOut } from "@/auth";
import { marcarLida as gravarLida } from "./briefings";
import * as formularios from "./formularios";
import * as fotos from "./fotos";
import { exigirAdmin } from "./sessao";

export async function entrar(): Promise<void> {
  await signIn("google", { redirectTo: "/admin/fotos" });
}

export async function sair(): Promise<void> {
  await signOut({ redirectTo: "/admin/entrar" });
}

export async function editarFoto(id: number, dados: { tema?: string; alt?: string }): Promise<void> {
  await exigirAdmin();
  fotos.editarFoto(id, dados);
}

export async function publicarFoto(id: number, publicada: boolean): Promise<void> {
  await exigirAdmin();
  fotos.publicarFoto(id, publicada);
}

export async function fixarFoto(id: number, fixada: boolean): Promise<void> {
  await exigirAdmin();
  fotos.fixarFoto(id, fixada);
}

export async function moverFixada(id: number, direcao: "antes" | "depois"): Promise<void> {
  await exigirAdmin();
  fotos.moverFixada(id, direcao);
}

export async function moverParaLixeira(id: number): Promise<void> {
  await exigirAdmin();
  fotos.moverParaLixeira(id);
}

export async function restaurarFoto(id: number): Promise<void> {
  await exigirAdmin();
  fotos.restaurarFoto(id);
}

export async function apagarDeVez(id: number): Promise<void> {
  await exigirAdmin();
  await fotos.apagarDeVez(id);
}

export async function criarTema(label: string): Promise<{ erro?: string }> {
  await exigirAdmin();
  try {
    fotos.criarTema(label);
    return {};
  } catch (erro) {
    return { erro: erro instanceof Error ? erro.message : "nao foi possivel criar o tema" };
  }
}

export async function renomearTema(slug: string, label: string): Promise<void> {
  await exigirAdmin();
  fotos.renomearTema(slug, label);
}

export async function mostrarTema(slug: string, visivel: boolean): Promise<void> {
  await exigirAdmin();
  fotos.mostrarTema(slug, visivel);
}

export async function moverTema(slug: string, direcao: "antes" | "depois"): Promise<void> {
  await exigirAdmin();
  fotos.moverTema(slug, direcao);
}

export async function marcarLida(id: number, lida: boolean): Promise<void> {
  await exigirAdmin();
  gravarLida(id, lida);
}

type Retorno = Promise<{ erro?: string }>;

async function editar(mudanca: () => void): Retorno {
  await exigirAdmin();
  try {
    mudanca();
    return {};
  } catch (erro) {
    if (erro instanceof formularios.EdicaoRecusada) return { erro: erro.message };
    console.error("admin: edicao de formulario falhou", erro);
    return { erro: "Não foi possível salvar. Tente de novo." };
  }
}

export async function salvarTextosFormulario(slug: string, textos: { titulo: string; subtitulo: string; enviar: string; sucesso: string }): Retorno {
  return editar(() => formularios.salvarTextos(slug, textos));
}

export async function salvarPergunta(slug: string, id: number, dados: { label: string; ajuda: string; obrigatoria: boolean; opcoes: string }): Retorno {
  return editar(() => formularios.salvarPergunta(slug, id, dados));
}

export async function adicionarPergunta(
  slug: string,
  dados: { label: string; ajuda: string; tipo: string; obrigatoria: boolean; opcoes: string },
): Retorno {
  return editar(() => formularios.adicionarPergunta(slug, dados));
}

export async function moverPergunta(slug: string, id: number, direcao: "antes" | "depois"): Retorno {
  return editar(() => formularios.moverPergunta(slug, id, direcao));
}

export async function ocultarPergunta(slug: string, id: number, oculta: boolean): Retorno {
  return editar(() => formularios.ocultarPergunta(slug, id, oculta));
}

export async function apagarPergunta(slug: string, id: number): Retorno {
  return editar(() => formularios.apagarPergunta(slug, id));
}

export async function trocarTipoPergunta(slug: string, id: number, tipo: string, opcoes: string): Retorno {
  return editar(() => formularios.trocarTipo(slug, id, tipo, opcoes));
}
