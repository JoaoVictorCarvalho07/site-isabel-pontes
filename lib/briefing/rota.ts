import { NextResponse, type NextRequest } from "next/server";
import { prefixo, type Idioma } from "../i18n";
import { processarBriefing } from "./processar";

const PAGINA: Record<string, string> = { fotografia: "/briefing", ugc: "/midiakit" };

export async function responderEnvio(requisicao: NextRequest, slug: string, idioma: Idioma): Promise<NextResponse> {
  let resultado: Awaited<ReturnType<typeof processarBriefing>>;
  try {
    resultado = await processarBriefing(slug, idioma, await requisicao.formData());
  } catch (erro) {
    console.warn("briefing: corpo recusado", { form: slug, erro });
    resultado = { tipo: "recusado", valores: {}, motivo: "form.erroGeral" };
  }
  if (resultado.tipo === "inexistente") return new NextResponse(null, { status: 404 });

  const estado =
    resultado.tipo === "enviado" ? "enviado" : resultado.tipo === "invalido" ? "corrija" : resultado.motivo === "form.muitosEnvios" ? "espere" : "erro";
  const destino = `${prefixo(idioma)}${PAGINA[slug] ?? "/briefing"}?briefing=${estado}#form-briefing`;
  return new NextResponse(null, { status: 303, headers: { Location: destino } });
}
