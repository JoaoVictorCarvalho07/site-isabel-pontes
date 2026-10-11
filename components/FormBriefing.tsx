import type { Formulario, Pergunta } from "@/lib/forms";
import { prefixo, resolver, t, type Idioma } from "@/lib/i18n";

type Props = { idioma: Idioma; form: Formulario };

function Campo({ idioma, slug, pergunta }: { idioma: Idioma; slug: string; pergunta: Pergunta }) {
  const id = `campo-${slug}-${pergunta.chave}`;
  const rotulo = resolver(idioma, pergunta.label);
  const ajuda = resolver(idioma, pergunta.ajuda);
  const descricao = ajuda ? `${id}-ajuda` : undefined;
  const comum = { id, name: pergunta.chave, "aria-describedby": descricao };
  return (
    <div className="campo">
      <label htmlFor={id}>
        {pergunta.obrigatoria ? rotulo : `${rotulo} `}
        {!pergunta.obrigatoria && <span className="campo__opcional">{`(${t(idioma, "form.opcional")})`}</span>}
      </label>
      {pergunta.tipo === "textarea" && <textarea {...comum} rows={4} maxLength={4000} required={pergunta.obrigatoria} defaultValue="" />}
      {pergunta.tipo === "select" && (
        <select {...comum} required={pergunta.obrigatoria} defaultValue="">
          <option value="">{t(idioma, "form.escolha")}</option>
          {pergunta.opcoes.map((opcao) => {
            const texto = resolver(idioma, opcao);
            return (
              <option key={texto} value={texto}>
                {texto}
              </option>
            );
          })}
        </select>
      )}
      {pergunta.tipo === "file" && <input {...comum} type="file" accept="image/*" multiple />}
      {pergunta.tipo !== "textarea" && pergunta.tipo !== "select" && pergunta.tipo !== "file" && (
        <input
          {...comum}
          type={pergunta.tipo === "contato" ? "text" : pergunta.tipo}
          defaultValue=""
          maxLength={4000}
          required={pergunta.obrigatoria}
        />
      )}
      {ajuda && (
        <small className="campo__ajuda" id={`${id}-ajuda`}>
          {ajuda}
        </small>
      )}
    </div>
  );
}

export function FormBriefing({ idioma, form }: Props) {
  return (
    <div className="form-briefing" id="form-briefing">
      <form className="form-briefing__form" method="post" action={`${prefixo(idioma)}/briefing/${form.slug}`} encType="multipart/form-data">
        {form.perguntas.map((pergunta) => (
          <Campo key={pergunta.id} idioma={idioma} slug={form.slug} pergunta={pergunta} />
        ))}
        <div className="campo-armadilha" aria-hidden="true">
          <label htmlFor={`website-${form.slug}`}>Não preencha este campo</label>
          <input id={`website-${form.slug}`} name="website" type="text" tabIndex={-1} autoComplete="off" />
        </div>
        <button className="form-briefing__enviar" type="submit">
          <span className="form-briefing__enviar-texto">{resolver(idioma, form.enviar)}</span>
          <span className="form-briefing__enviando" aria-hidden="true">
            {t(idioma, "form.enviando")}
          </span>
        </button>
      </form>
    </div>
  );
}
