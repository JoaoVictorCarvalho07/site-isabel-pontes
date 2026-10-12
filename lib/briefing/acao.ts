"use server";

import { t, type Idioma } from "../i18n";
import { processarBriefing } from "./processar";

export type EstadoBriefing = { enviado: boolean; erros: Record<string, string> };

export async function enviarBriefing(slug: string, idioma: Idioma, dados: FormData): Promise<EstadoBriefing> {
  const resultado = await processarBriefing(slug, idioma, dados);
  if (resultado.tipo === "enviado") return { enviado: true, erros: {} };
  if (resultado.tipo === "invalido") {
    const erros = Object.fromEntries(Object.entries(resultado.erros).map(([chave, erro]) => [chave, t(idioma, erro)]));
    return { enviado: false, erros };
  }
  const motivo = resultado.tipo === "recusado" ? resultado.motivo : "form.erroGeral";
  return { enviado: false, erros: { _geral: t(idioma, motivo) } };
}
