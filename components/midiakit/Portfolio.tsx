"use client";

import { useState, type ReactNode } from "react";

export type Grupo = { slug: string; nome: string; conteudo: ReactNode };

type Props = { grupos: Grupo[]; rotuloTodos: string; rotuloNav: string };

export function Portfolio({ grupos, rotuloTodos, rotuloNav }: Props) {
  const [ativa, setAtiva] = useState("");
  const filtros = [{ slug: "", nome: rotuloTodos }, ...grupos];
  return (
    <div id="mk-portfolio">
      <nav className="mk-filtros" aria-label={rotuloNav}>
        {filtros.map((filtro) => (
          <button
            key={filtro.slug}
            type="button"
            className={filtro.slug === ativa ? "mk-filtro mk-filtro--ativo" : "mk-filtro"}
            aria-current={filtro.slug === ativa ? "true" : undefined}
            onClick={() => setAtiva(filtro.slug)}
          >
            {filtro.nome}
          </button>
        ))}
      </nav>
      {grupos
        .filter((grupo) => ativa === "" || grupo.slug === ativa)
        .map((grupo) => (
          <section className="mk-vgrupo" key={`${ativa}:${grupo.slug}`}>
            <h3 className="mk-vgrupo__nome">{grupo.nome}</h3>
            {grupo.conteudo}
          </section>
        ))}
    </div>
  );
}
