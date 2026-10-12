import { notFound } from "next/navigation";
import { Casca } from "@/components/Casca";
import { FormBriefing, type RetornoSemJS } from "@/components/FormBriefing";
import { formularioPublico } from "@/lib/forms";
import { prefixo, resolver, type Idioma } from "@/lib/i18n";
import { metadados } from "@/lib/metadados";

export function metadadosBriefing(idioma: Idioma) {
  return metadados({
    idioma,
    caminhoPT: "/briefing",
    titulo: "Briefing — Isabel Pontes",
    descricao:
      idioma === "en"
        ? "Tell me about the shoot you have in mind: theme, date, references and what inspired you."
        : "Conte sobre o ensaio que você quer fazer: tema, data, referências e o que te motivou.",
  });
}

export type BuscaBriefing = Promise<{ briefing?: string | string[] }>;

export async function Briefing({ idioma, busca }: { idioma: Idioma; busca: BuscaBriefing }) {
  const form = formularioPublico("fotografia");
  if (!form) notFound();
  const en = idioma === "en";
  const { briefing } = await busca;
  return (
    <Casca>
      <section className="briefing">
        <header className="briefing__head">
          <h1>{resolver(idioma, form.titulo)}</h1>
          <p>{resolver(idioma, form.subtitulo)}</p>
        </header>
        <FormBriefing idioma={idioma} form={form} retorno={briefing as RetornoSemJS} />
        <p className="briefing__alternativa">
          {en ? "Prefer to talk first? " : "Prefere conversar antes? "}
          <a href={`${prefixo(idioma)}/contato`}>{en ? "Message me on WhatsApp" : "Me chama no WhatsApp"}</a>
          {"."}
        </p>
      </section>
    </Casca>
  );
}
