"use client";

import { useState, useTransition, type FormEvent } from "react";
import { enviarBriefing, type EstadoBriefing } from "@/lib/briefing/acao";
import { MAXIMO_ARQUIVOS, TAMANHO_MAXIMO_ARQUIVO } from "@/lib/briefing/tipos";
import type { Formulario, Pergunta } from "@/lib/forms";
import { prefixo, resolver, t, type Idioma } from "@/lib/i18n";

export type RetornoSemJS = "enviado" | "corrija" | "espere" | "erro" | undefined;

type Props = { idioma: Idioma; form: Formulario; retorno?: RetornoSemJS };

const MENSAGEM_RETORNO = { corrija: "form.corrija", espere: "form.muitosEnvios", erro: "form.erroGeral" } as const;

function estadoInicial(idioma: Idioma, retorno: RetornoSemJS): EstadoBriefing {
  if (retorno === "enviado") return { enviado: true, erros: {} };
  if (retorno) return { enviado: false, erros: { _geral: t(idioma, MENSAGEM_RETORNO[retorno]) } };
  return { enviado: false, erros: {} };
}

type CampoProps = { idioma: Idioma; slug: string; pergunta: Pergunta; erro?: string; houveErro: boolean };

function Campo({ idioma, slug, pergunta, erro, houveErro }: CampoProps) {
  const id = `campo-${slug}-${pergunta.chave}`;
  const rotulo = resolver(idioma, pergunta.label);
  const ajuda = resolver(idioma, pergunta.ajuda);
  const descricao = [ajuda && `${id}-ajuda`, erro && `${id}-erro`].filter(Boolean).join(" ") || undefined;
  const comum = { id, name: pergunta.chave, "aria-describedby": descricao };
  const validacao = { required: pergunta.obrigatoria, "aria-invalid": erro ? ("true" as const) : undefined };
  return (
    <div className={erro ? "campo campo--erro" : "campo"}>
      <label htmlFor={id}>
        {pergunta.obrigatoria ? rotulo : `${rotulo} `}
        {!pergunta.obrigatoria && <span className="campo__opcional">{`(${t(idioma, "form.opcional")})`}</span>}
      </label>
      {pergunta.tipo === "textarea" && <textarea {...comum} {...validacao} rows={4} maxLength={4000} defaultValue="" />}
      {pergunta.tipo === "select" && (
        <select {...comum} {...validacao} defaultValue="">
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
      {pergunta.tipo === "file" && (
        <>
          <input {...comum} type="file" accept="image/*" multiple />
          {houveErro && <small className="campo__realerta">{t(idioma, "form.reanexar")}</small>}
        </>
      )}
      {pergunta.tipo !== "textarea" && pergunta.tipo !== "select" && pergunta.tipo !== "file" && (
        <input {...comum} {...validacao} type={pergunta.tipo === "contato" ? "text" : pergunta.tipo} defaultValue="" maxLength={4000} />
      )}
      {ajuda && (
        <small className="campo__ajuda" id={`${id}-ajuda`}>
          {ajuda}
        </small>
      )}
      {erro && (
        <small className="campo__erro" id={`${id}-erro`}>
          {erro}
        </small>
      )}
    </div>
  );
}

function anexosGrandesDemais(dados: FormData): boolean {
  const arquivos = [...dados.values()].filter((valor): valor is File => valor instanceof File);
  return arquivos.reduce((soma, arquivo) => soma + arquivo.size, 0) > MAXIMO_ARQUIVOS * TAMANHO_MAXIMO_ARQUIVO;
}

export function FormBriefing({ idioma, form, retorno }: Props) {
  const [estado, setEstado] = useState(() => estadoInicial(idioma, retorno));
  const [enviando, iniciar] = useTransition();

  function enviar(evento: FormEvent<HTMLFormElement>) {
    evento.preventDefault();
    const dados = new FormData(evento.currentTarget);
    if (anexosGrandesDemais(dados)) {
      setEstado({ enviado: false, erros: { _geral: t(idioma, "form.arquivoGrande") } });
      return;
    }
    iniciar(async () => {
      try {
        setEstado(await enviarBriefing(form.slug, idioma, dados));
      } catch {
        setEstado({ enviado: false, erros: { _geral: t(idioma, "form.erroGeral") } });
      }
    });
  }

  if (estado.enviado) {
    return (
      <div className="form-briefing" id="form-briefing">
        <div className="form-briefing__sucesso" role="status">
          <span className="form-briefing__sucesso-icone" aria-hidden="true">
            ✿
          </span>
          <p>{resolver(idioma, form.sucesso)}</p>
        </div>
      </div>
    );
  }

  const houveErro = Object.keys(estado.erros).length > 0;
  const erroGeral = estado.erros._geral ?? (houveErro ? t(idioma, "form.corrija") : "");

  return (
    <div className="form-briefing" id="form-briefing">
      <form
        className="form-briefing__form"
        method="post"
        action={`${prefixo(idioma)}/briefing/${form.slug}`}
        encType="multipart/form-data"
        onSubmit={enviar}
      >
        {erroGeral && (
          <p className="form-briefing__erro-geral" role="alert">
            {erroGeral}
          </p>
        )}
        {form.perguntas.map((pergunta) => (
          <Campo key={pergunta.id} idioma={idioma} slug={form.slug} pergunta={pergunta} erro={estado.erros[pergunta.chave]} houveErro={houveErro} />
        ))}
        <div className="campo-armadilha" aria-hidden="true">
          <label htmlFor={`website-${form.slug}`}>Não preencha este campo</label>
          <input id={`website-${form.slug}`} name="website" type="text" tabIndex={-1} autoComplete="off" />
        </div>
        <button className="form-briefing__enviar" type="submit" disabled={enviando}>
          <span className="form-briefing__enviar-texto">{resolver(idioma, form.enviar)}</span>
          <span className="form-briefing__enviando" aria-hidden="true">
            {t(idioma, "form.enviando")}
          </span>
        </button>
      </form>
    </div>
  );
}
