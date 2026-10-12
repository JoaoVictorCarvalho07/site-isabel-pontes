import { headers } from "next/headers";
import { banco } from "../banco";
import { formularioPublico } from "../forms";
import type { Idioma } from "../i18n";
import { salvarAnexos } from "./anexos";
import { notificar } from "./email";
import { permitir } from "./limitador";
import { validar } from "./validar";

export type Resultado =
  | { tipo: "enviado" }
  | { tipo: "inexistente" }
  | { tipo: "invalido"; valores: Record<string, string>; erros: Record<string, string> }
  | { tipo: "recusado"; valores: Record<string, string>; motivo: "form.muitosEnvios" | "form.erroGeral" };

async function origem(): Promise<{ ip: string; navegador: string }> {
  const cabecalhos = await headers();
  const encaminhado = cabecalhos.get("x-forwarded-for") ?? "";
  return { ip: encaminhado.split(",")[0].trim() || "local", navegador: cabecalhos.get("user-agent") ?? "" };
}

export async function processarBriefing(slug: string, idioma: Idioma, dados: FormData): Promise<Resultado> {
  const form = formularioPublico(slug);
  if (!form) return { tipo: "inexistente" };

  const { ip, navegador } = await origem();
  if (String(dados.get("website") ?? "").trim() !== "") {
    console.info("briefing: honeypot", { form: slug, ip });
    return { tipo: "enviado" };
  }

  const { valores, erros } = validar(form, (chave) => String(dados.get(chave) ?? "").replace(/\r\n/g, "\n"));
  if (Object.keys(erros).length > 0) return { tipo: "invalido", valores, erros };
  if (!permitir(ip)) return { tipo: "recusado", valores, motivo: "form.muitosEnvios" };

  let id: number;
  try {
    const gravado = banco()
      .prepare("INSERT INTO submissions (form, lang, payload, ip, user_agent) VALUES (?,?,?,?,?)")
      .run(slug, idioma, JSON.stringify(valores), ip, navegador);
    id = Number(gravado.lastInsertRowid);
  } catch (erro) {
    console.error("briefing: salvar falhou", { form: slug, erro, respostas: valores });
    return { tipo: "recusado", valores, motivo: "form.erroGeral" };
  }

  let anexos = 0;
  for (const pergunta of form.perguntas.filter((p) => p.tipo === "file")) {
    const arquivos = dados.getAll(pergunta.chave).filter((item): item is File => item instanceof File);
    try {
      anexos += await salvarAnexos(id, pergunta.chave, arquivos);
    } catch (erro) {
      console.error("briefing: anexo falhou", { id, erro });
    }
  }

  console.info("briefing recebido", { form: slug, id, idioma });
  void notificar(form, idioma, valores, id, anexos);
  return { tipo: "enviado" };
}
