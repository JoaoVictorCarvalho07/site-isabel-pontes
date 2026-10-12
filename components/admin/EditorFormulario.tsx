"use client";

import { useRouter } from "next/navigation";
import { useState, useTransition, type FormEvent } from "react";
import * as acoes from "@/lib/admin/acoes";
import type { FormularioEdicao, PerguntaEdicao } from "@/lib/admin/formularios";

type Tipo = { valor: string; rotulo: string };

type Props = { form: FormularioEdicao; tipos: Tipo[]; previa: string };

type Executar = (acao: () => Promise<{ erro?: string }>, aoConcluir?: () => void) => void;

function texto(dados: FormData, nome: string): string {
  return String(dados.get(nome) ?? "");
}

function Pergunta({ slug, pergunta, tipos, primeira, ultima, ocupado, executar }: {
  slug: string;
  pergunta: PerguntaEdicao;
  tipos: Tipo[];
  primeira: boolean;
  ultima: boolean;
  ocupado: boolean;
  executar: Executar;
}) {
  const [novoTipo, setNovoTipo] = useState(pergunta.tipo);
  const rotuloTipo = tipos.find((tipo) => tipo.valor === pergunta.tipo)?.rotulo ?? pergunta.tipo;

  function salvar(evento: FormEvent<HTMLFormElement>) {
    evento.preventDefault();
    const dados = new FormData(evento.currentTarget);
    executar(() =>
      acoes.salvarPergunta(slug, pergunta.id, {
        label: texto(dados, "label"),
        ajuda: texto(dados, "ajuda"),
        obrigatoria: dados.get("obrigatoria") !== null,
        opcoes: texto(dados, "opcoes"),
      }),
    );
  }

  function trocar(evento: FormEvent<HTMLFormElement>) {
    evento.preventDefault();
    const dados = new FormData(evento.currentTarget);
    executar(() => acoes.trocarTipoPergunta(slug, pergunta.id, novoTipo, texto(dados, "opcoes")));
  }

  return (
    <li className={pergunta.oculta ? "adm-pergunta adm-pergunta--oculta" : "adm-pergunta"} data-pergunta={pergunta.chave}>
      <div className="adm-pergunta__selos">
        <span>{rotuloTipo}</span>
        {pergunta.oculta && <span>oculta</span>}
        {pergunta.ingles && <span>EN a revisar</span>}
        <span>{`${pergunta.respostas} ${pergunta.respostas === 1 ? "resposta" : "respostas"}`}</span>
      </div>
      <form onSubmit={salvar}>
        <label>
          Pergunta
          <input className="adm-campo" name="label" defaultValue={pergunta.label} maxLength={300} required />
        </label>
        <label>
          Texto de ajuda (opcional)
          <input className="adm-campo" name="ajuda" defaultValue={pergunta.ajuda} maxLength={300} />
        </label>
        {pergunta.tipo === "select" && (
          <label>
            Opções, uma por linha
            <textarea className="adm-campo" name="opcoes" rows={4} defaultValue={pergunta.opcoes} />
          </label>
        )}
        <label className="adm-marcar">
          <input type="checkbox" name="obrigatoria" defaultChecked={pergunta.obrigatoria} />
          Resposta obrigatória
        </label>
        <button className="adm-botao adm-botao--principal" type="submit" disabled={ocupado}>
          Salvar pergunta
        </button>
      </form>
      <div className="adm-pergunta__acoes">
        <button className="adm-botao" aria-label="Mover para antes" disabled={ocupado || primeira} onClick={() => executar(() => acoes.moverPergunta(slug, pergunta.id, "antes"))}>
          ↑
        </button>
        <button className="adm-botao" aria-label="Mover para depois" disabled={ocupado || ultima} onClick={() => executar(() => acoes.moverPergunta(slug, pergunta.id, "depois"))}>
          ↓
        </button>
        <button className="adm-botao" disabled={ocupado} onClick={() => executar(() => acoes.ocultarPergunta(slug, pergunta.id, !pergunta.oculta))}>
          {pergunta.oculta ? "Mostrar no formulário" : "Ocultar do formulário"}
        </button>
      </div>
      <details>
        <summary>Mais ações</summary>
        <form className="adm-pergunta__tipo" onSubmit={trocar}>
          <label>
            Trocar o tipo
            <select className="adm-campo" value={novoTipo} onChange={(evento) => setNovoTipo(evento.target.value)}>
              {tipos.map((tipo) => (
                <option key={tipo.valor} value={tipo.valor}>
                  {tipo.rotulo}
                </option>
              ))}
            </select>
          </label>
          {novoTipo === "select" && (
            <label>
              Opções, uma por linha
              <textarea className="adm-campo" name="opcoes" rows={3} defaultValue={pergunta.opcoes} />
            </label>
          )}
          <button className="adm-botao" type="submit" disabled={ocupado || novoTipo === pergunta.tipo}>
            Trocar tipo
          </button>
        </form>
        <button
          className="adm-botao adm-botao--perigo"
          disabled={ocupado || pergunta.respostas > 0}
          onClick={() => confirm("Apagar esta pergunta?") && executar(() => acoes.apagarPergunta(slug, pergunta.id))}
        >
          Apagar pergunta
        </button>
        {pergunta.respostas > 0 && <p className="adm-nota">Pergunta com respostas não pode ser apagada; oculte para tirá-la do formulário.</p>}
      </details>
    </li>
  );
}

export function EditorFormulario({ form, tipos, previa }: Props) {
  const router = useRouter();
  const [ocupado, iniciar] = useTransition();
  const [aviso, setAviso] = useState<{ erro: boolean; texto: string } | null>(null);
  const [tipoNovo, setTipoNovo] = useState("text");

  const executar: Executar = (acao, aoConcluir) => {
    iniciar(async () => {
      const { erro } = await acao();
      setAviso(erro ? { erro: true, texto: erro } : { erro: false, texto: "Salvo. Já está valendo no site." });
      if (!erro) aoConcluir?.();
      router.refresh();
    });
  };

  function salvarTextos(evento: FormEvent<HTMLFormElement>) {
    evento.preventDefault();
    const dados = new FormData(evento.currentTarget);
    executar(() =>
      acoes.salvarTextosFormulario(form.slug, {
        titulo: texto(dados, "titulo"),
        subtitulo: texto(dados, "subtitulo"),
        enviar: texto(dados, "enviar"),
        sucesso: texto(dados, "sucesso"),
      }),
    );
  }

  function adicionar(evento: FormEvent<HTMLFormElement>) {
    evento.preventDefault();
    const formulario = evento.currentTarget;
    const dados = new FormData(formulario);
    executar(
      () =>
        acoes.adicionarPergunta(form.slug, {
          label: texto(dados, "label"),
          ajuda: texto(dados, "ajuda"),
          tipo: tipoNovo,
          obrigatoria: dados.get("obrigatoria") !== null,
          opcoes: texto(dados, "opcoes"),
        }),
      () => formulario.reset(),
    );
  }

  return (
    <>
      {aviso && (
        <p className={aviso.erro ? "adm-aviso" : "adm-aviso adm-aviso--ok"} role={aviso.erro ? "alert" : "status"}>
          {aviso.texto}
        </p>
      )}
      <p>
        <a href={previa} target="_blank" rel="noopener">
          Ver o formulário no site
        </a>
      </p>

      <form className="adm-bloco" onSubmit={salvarTextos} aria-label="Textos do formulário">
        <label>
          Título
          <input className="adm-campo" name="titulo" defaultValue={form.titulo} maxLength={200} required />
        </label>
        <label>
          Subtítulo
          <input className="adm-campo" name="subtitulo" defaultValue={form.subtitulo} maxLength={300} />
        </label>
        <label>
          Texto do botão
          <input className="adm-campo" name="enviar" defaultValue={form.enviar} maxLength={60} required />
        </label>
        <label>
          Mensagem depois do envio
          <input className="adm-campo" name="sucesso" defaultValue={form.sucesso} maxLength={300} />
        </label>
        <button className="adm-botao adm-botao--principal" type="submit" disabled={ocupado}>
          Salvar textos
        </button>
      </form>

      <ul className="adm-perguntas">
        {form.perguntas.map((pergunta, indice) => (
          <Pergunta
            key={pergunta.id}
            slug={form.slug}
            pergunta={pergunta}
            tipos={tipos}
            primeira={indice === 0}
            ultima={indice === form.perguntas.length - 1}
            ocupado={ocupado}
            executar={executar}
          />
        ))}
      </ul>

      <form className="adm-bloco adm-bloco--novo" onSubmit={adicionar} aria-label="Nova pergunta">
        <h2>Nova pergunta</h2>
        <label>
          Pergunta
          <input className="adm-campo" name="label" maxLength={300} required />
        </label>
        <label>
          Texto de ajuda (opcional)
          <input className="adm-campo" name="ajuda" maxLength={300} />
        </label>
        <label>
          Tipo
          <select className="adm-campo" value={tipoNovo} onChange={(evento) => setTipoNovo(evento.target.value)}>
            {tipos.map((tipo) => (
              <option key={tipo.valor} value={tipo.valor}>
                {tipo.rotulo}
              </option>
            ))}
          </select>
        </label>
        {tipoNovo === "select" && (
          <label>
            Opções, uma por linha
            <textarea className="adm-campo" name="opcoes" rows={4} />
          </label>
        )}
        <label className="adm-marcar">
          <input type="checkbox" name="obrigatoria" />
          Resposta obrigatória
        </label>
        <button className="adm-botao adm-botao--principal" type="submit" disabled={ocupado}>
          Adicionar pergunta
        </button>
      </form>
    </>
  );
}
