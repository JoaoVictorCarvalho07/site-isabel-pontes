"use client";

import { useRouter } from "next/navigation";
import { useState, useTransition } from "react";
import * as acoes from "@/lib/admin/acoes";
import type { FotoAdmin } from "@/lib/admin/fotos";

type Tema = { slug: string; label: string };

type Props = { fotos: FotoAdmin[]; temas: Tema[]; lixeira?: boolean };

export function ListaFotos({ fotos, temas, lixeira }: Props) {
  const router = useRouter();
  const [ocupado, iniciar] = useTransition();
  const [tema, setTema] = useState("");
  const [situacao, setSituacao] = useState("");

  function executar(acao: () => Promise<unknown>) {
    iniciar(async () => {
      await acao();
      router.refresh();
    });
  }

  const visiveis = fotos.filter(
    (foto) =>
      (tema === "" || foto.tema === tema) &&
      (situacao === "" ||
        (situacao === "publicadas" && foto.publicada) ||
        (situacao === "ocultas" && !foto.publicada) ||
        (situacao === "fixadas" && foto.ordem !== null)),
  );
  const fixadas = fotos.filter((foto) => foto.ordem !== null);

  return (
    <>
      <div className="adm-filtros">
        <select className="adm-campo" aria-label="Filtrar por tema" value={tema} onChange={(evento) => setTema(evento.target.value)}>
          <option value="">Todos os temas</option>
          {temas.map((opcao) => (
            <option key={opcao.slug} value={opcao.slug}>
              {opcao.label}
            </option>
          ))}
        </select>
        {!lixeira && (
          <select className="adm-campo" aria-label="Filtrar por situação" value={situacao} onChange={(evento) => setSituacao(evento.target.value)}>
            <option value="">Todas</option>
            <option value="publicadas">Publicadas</option>
            <option value="ocultas">Ocultas</option>
            <option value="fixadas">Fixadas no topo</option>
          </select>
        )}
      </div>
      {visiveis.length === 0 ? (
        <p className="adm-vazio">{lixeira ? "A lixeira está vazia." : "Nenhuma foto aqui."}</p>
      ) : (
        <ul className="adm-grade">
          {visiveis.map((foto) => (
            <li key={foto.id} className={foto.publicada || lixeira ? "adm-foto" : "adm-foto adm-foto--oculta"} data-foto={foto.id}>
              <img className="adm-foto__img" src={foto.grid} alt={foto.alt} loading="lazy" decoding="async" />
              <div className="adm-foto__dados">
                {lixeira ? (
                  <span>{temas.find((opcao) => opcao.slug === foto.tema)?.label ?? foto.tema}</span>
                ) : (
                  <>
                    <div className="adm-foto__selos">
                      {foto.ordem !== null && <span>{`fixada · ${fixadas.indexOf(foto) + 1}ª`}</span>}
                      {!foto.publicada && <span>oculta</span>}
                    </div>
                    <select
                      className="adm-campo"
                      aria-label="Tema"
                      value={foto.tema}
                      disabled={ocupado}
                      onChange={(evento) => executar(() => acoes.editarFoto(foto.id, { tema: evento.target.value }))}
                    >
                      {temas.map((opcao) => (
                        <option key={opcao.slug} value={opcao.slug}>
                          {opcao.label}
                        </option>
                      ))}
                    </select>
                    <input
                      className="adm-campo"
                      aria-label="Descrição da foto"
                      placeholder="Descrição (opcional)"
                      defaultValue={foto.alt}
                      maxLength={300}
                      onBlur={(evento) => evento.target.value !== foto.alt && executar(() => acoes.editarFoto(foto.id, { alt: evento.target.value }))}
                    />
                  </>
                )}
              </div>
              <div className="adm-foto__acoes">
                {lixeira ? (
                  <>
                    <button className="adm-botao" disabled={ocupado} onClick={() => executar(() => acoes.restaurarFoto(foto.id))}>
                      Restaurar
                    </button>
                    <button
                      className="adm-botao adm-botao--perigo"
                      disabled={ocupado}
                      onClick={() => confirm("Apagar de vez? Isso não tem volta.") && executar(() => acoes.apagarDeVez(foto.id))}
                    >
                      Apagar de vez
                    </button>
                  </>
                ) : (
                  <>
                    <button className="adm-botao" disabled={ocupado} onClick={() => executar(() => acoes.publicarFoto(foto.id, !foto.publicada))}>
                      {foto.publicada ? "Ocultar" : "Publicar"}
                    </button>
                    <button className="adm-botao" disabled={ocupado} onClick={() => executar(() => acoes.fixarFoto(foto.id, foto.ordem === null))}>
                      {foto.ordem === null ? "Fixar no topo" : "Desafixar"}
                    </button>
                    {foto.ordem !== null && (
                      <>
                        <button
                          className="adm-botao"
                          aria-label="Mover para antes"
                          disabled={ocupado || fixadas[0] === foto}
                          onClick={() => executar(() => acoes.moverFixada(foto.id, "antes"))}
                        >
                          ↑
                        </button>
                        <button
                          className="adm-botao"
                          aria-label="Mover para depois"
                          disabled={ocupado || fixadas.at(-1) === foto}
                          onClick={() => executar(() => acoes.moverFixada(foto.id, "depois"))}
                        >
                          ↓
                        </button>
                      </>
                    )}
                    <button className="adm-botao adm-botao--perigo" disabled={ocupado} onClick={() => executar(() => acoes.moverParaLixeira(foto.id))}>
                      Lixeira
                    </button>
                  </>
                )}
              </div>
            </li>
          ))}
        </ul>
      )}
    </>
  );
}
