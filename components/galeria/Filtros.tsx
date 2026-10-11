"use client";

import { useRouter } from "next/navigation";
import { useTransition } from "react";

export type Filtro = { slug: string; label: string; ativo: boolean };

type Props = { filtros: Filtro[]; caminho: string; rotulo: string };

export function Filtros({ filtros, caminho, rotulo }: Props) {
  const router = useRouter();
  const [carregando, iniciar] = useTransition();

  function filtrar(slug: string) {
    const destino = slug ? `${caminho}?tema=${slug}` : caminho;
    iniciar(() => router.push(destino, { scroll: false }));
  }

  return (
    <>
      <nav className="filtros" aria-label={rotulo}>
        {filtros.map((filtro) => (
          <button
            key={filtro.slug}
            type="button"
            className={filtro.ativo ? "filtro filtro--ativo" : "filtro"}
            aria-current={filtro.ativo ? "true" : undefined}
            onClick={() => filtrar(filtro.slug)}
          >
            {filtro.label}
          </button>
        ))}
      </nav>
      <div id="galeria-load" className={carregando ? "galeria-load htmx-request" : "galeria-load"} aria-hidden="true"></div>
    </>
  );
}
