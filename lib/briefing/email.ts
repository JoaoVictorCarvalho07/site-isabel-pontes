import { BASE_URL } from "../config";
import type { Formulario } from "../forms";
import { emailValido } from "./validar";

function escapar(texto: string): string {
  return texto.replace(/&/g, "&amp;").replace(/</g, "&lt;").replace(/>/g, "&gt;").replace(/"/g, "&#34;").replace(/'/g, "&#39;");
}

function resumo(form: Formulario, respostas: Record<string, string>): string {
  const pergunta = form.perguntas.find((p) => p.obrigatoria && p.tipo === "text" && respostas[p.chave]);
  return pergunta ? respostas[pergunta.chave] : "sem identificacao";
}

function emailDoVisitante(form: Formulario, respostas: Record<string, string>): string | undefined {
  return form.perguntas
    .filter((p) => p.tipo === "email" || p.tipo === "contato")
    .map((p) => respostas[p.chave] ?? "")
    .find((valor) => valor.includes("@") && emailValido(valor));
}

function corpo(form: Formulario, idioma: string, respostas: Record<string, string>, id: number, anexos: number): string {
  const campos = form.perguntas
    .filter((p) => p.tipo !== "file" && respostas[p.chave])
    .map(
      (p) =>
        `<p style="margin:14px 0"><strong style="display:block;color:#5f5f5f;font-size:13px">${escapar(p.label[0])}</strong>${escapar(respostas[p.chave]).replaceAll("\n", "<br>")}</p>`,
    )
    .join("");
  const aviso =
    anexos > 0
      ? `<p style="margin:14px 0;padding:10px 12px;background:#f4efe6;border-radius:6px">📎 <strong>${anexos} imagem(ns) de referência</strong> anexada(s) — veja no painel.</p>`
      : "";
  const painel = escapar(`${BASE_URL}/admin`);
  return (
    `<div style="font-family:system-ui,sans-serif;font-size:15px;line-height:1.6;color:#2b2b2b">` +
    `<h2 style="font-weight:600">${escapar(form.titulo[0])}</h2>` +
    `<p style="color:#777;font-size:13px">Briefing #${id} · formulário <strong>${escapar(form.slug)}</strong> · idioma ${escapar(idioma)}</p>` +
    campos +
    aviso +
    `<hr style="border:none;border-top:1px solid #e5e5e5;margin:24px 0">` +
    `<p style="font-size:13px;color:#777">Também disponível em <a href="${painel}">${painel}</a>.</p></div>`
  );
}

export async function notificar(form: Formulario, idioma: string, respostas: Record<string, string>, id: number, anexos: number) {
  const chave = process.env.RESEND_API_KEY;
  const para = process.env.MAIL_PARA;
  if (!chave || !para) {
    console.info("briefing: e-mail desligado (sem RESEND_API_KEY/MAIL_PARA)", { id });
    return;
  }
  try {
    const resposta = await fetch("https://api.resend.com/emails", {
      method: "POST",
      headers: { Authorization: `Bearer ${chave}`, "Content-Type": "application/json" },
      body: JSON.stringify({
        from: process.env.MAIL_DE ?? "briefing@isabelpontes.com.br",
        to: [para],
        subject: `Novo briefing (${form.slug}) — ${resumo(form, respostas)}`,
        html: corpo(form, idioma, respostas, id, anexos),
        reply_to: emailDoVisitante(form, respostas) ?? para,
      }),
      signal: AbortSignal.timeout(15_000),
    });
    if (!resposta.ok) throw new Error(`resend devolveu ${resposta.status}: ${(await resposta.text()).slice(0, 2048)}`);
    console.info("briefing: e-mail enviado", { id });
  } catch (erro) {
    console.error("briefing: e-mail falhou (resposta salva mesmo assim)", { id, erro });
  }
}
