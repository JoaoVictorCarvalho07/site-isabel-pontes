"use client";

import { useRouter } from "next/navigation";
import { useState, useTransition, type FormEvent } from "react";
import * as acoes from "@/lib/admin/acoes";
import type { TemaAdmin } from "@/lib/admin/fotos";

export function ListaTemas({ temas }: { temas: TemaAdmin[] }) {
  const router = useRouter();
  const [ocupado, iniciar] = useTransition();
  const [aviso, setAviso] = useState("");

  function executar(acao: () => Promise<unknown>) {
    iniciar(async () => {
      await acao();
      router.refresh();
    });
  }

  function criar(evento: FormEvent<HTMLFormElement>) {
    evento.preventDefault();
    const formulario = evento.currentTarget;
    const nome = String(new FormData(formulario).get("nome") ?? "");
    iniciar(async () => {
      const { erro } = await acoes.criarTema(nome);
      setAviso(erro ? "Não deu para criar: confira se o nome já existe." : "");
      if (!erro) formulario.reset();
      router.refresh();
    });
  }

  return (
    <>
      <form className="adm-novo-tema" onSubmit={criar}>
        <input className="adm-campo" name="nome" aria-label="Nome do novo tema" placeholder="Novo tema" maxLength={40} required />
        <button className="adm-botao adm-botao--principal" type="submit" disabled={ocupado}>
          Criar
        </button>
      </form>
      {aviso && <p className="adm-aviso">{aviso}</p>}
      <ul className="adm-temas">
        {temas.map((tema, indice) => (
          <li key={tema.slug} data-tema={tema.slug}>
            <input
              className="adm-campo"
              aria-label={`Nome do tema ${tema.label}`}
              defaultValue={tema.label}
              maxLength={40}
              onBlur={(evento) => {
                const nome = evento.target.value.trim();
                if (nome && nome !== tema.label) executar(() => acoes.renomearTema(tema.slug, nome));
              }}
            />
            <small>{`${tema.fotos} ${tema.fotos === 1 ? "foto" : "fotos"}${tema.visivel ? "" : " · oculto"}`}</small>
            <div className="adm-temas__acoes">
              <button className="adm-botao" aria-label="Mover para antes" disabled={ocupado || indice === 0} onClick={() => executar(() => acoes.moverTema(tema.slug, "antes"))}>
                ↑
              </button>
              <button
                className="adm-botao"
                aria-label="Mover para depois"
                disabled={ocupado || indice === temas.length - 1}
                onClick={() => executar(() => acoes.moverTema(tema.slug, "depois"))}
              >
                ↓
              </button>
              <button className="adm-botao" disabled={ocupado} onClick={() => executar(() => acoes.mostrarTema(tema.slug, !tema.visivel))}>
                {tema.visivel ? "Ocultar do site" : "Mostrar no site"}
              </button>
            </div>
          </li>
        ))}
      </ul>
    </>
  );
}
