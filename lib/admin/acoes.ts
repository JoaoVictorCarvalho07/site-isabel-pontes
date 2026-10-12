"use server";

import { signIn, signOut } from "@/auth";
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
