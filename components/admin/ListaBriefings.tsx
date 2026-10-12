"use client";

import { useRouter } from "next/navigation";
import { useTransition } from "react";
import { marcarLida } from "@/lib/admin/acoes";
import type { Briefing } from "@/lib/admin/briefings";

export function ListaBriefings({ briefings }: { briefings: Briefing[] }) {
  const router = useRouter();
  const [ocupado, iniciar] = useTransition();

  if (briefings.length === 0) return <p className="adm-vazio">Nenhum briefing por aqui ainda.</p>;

  return (
    <ul className="adm-briefings">
      {briefings.map((briefing) => (
        <li key={briefing.id} className={briefing.lida ? "adm-briefing adm-briefing--lida" : "adm-briefing"} data-briefing={briefing.id}>
          <header>
            <div>
              <h2>{briefing.destaque || `Briefing #${briefing.id}`}</h2>
              <p>
                <span className="adm-briefing__tag">{briefing.form}</span>
                {` ${briefing.quando}`}
              </p>
            </div>
            <button
              className="adm-botao"
              disabled={ocupado}
              onClick={() =>
                iniciar(async () => {
                  await marcarLida(briefing.id, !briefing.lida);
                  router.refresh();
                })
              }
            >
              {briefing.lida ? "Marcar como não lida" : "Marcar como lida"}
            </button>
          </header>
          <dl>
            {briefing.campos.map((campo) => (
              <div key={campo.rotulo}>
                <dt>{campo.rotulo}</dt>
                <dd>
                  {campo.trechos.map((trecho, indice) =>
                    trecho.href ? (
                      <a key={indice} href={trecho.href} target="_blank" rel="noopener noreferrer nofollow">
                        {trecho.texto}
                      </a>
                    ) : (
                      <span key={indice}>{trecho.texto}</span>
                    ),
                  )}
                </dd>
              </div>
            ))}
          </dl>
          {briefing.anexos.length > 0 && (
            <div className="adm-briefing__anexos">
              <p>{`Referências enviadas (${briefing.anexos.length})`}</p>
              <ul>
                {briefing.anexos.map((anexo) => (
                  <li key={anexo.id}>
                    <a href={`/admin/arquivo/${anexo.id}`} target="_blank" rel="noopener">
                      <img src={`/admin/arquivo/${anexo.id}`} alt={anexo.nome} loading="lazy" decoding="async" />
                      <span>{`${anexo.nome} · ${anexo.kb} KB`}</span>
                    </a>
                  </li>
                ))}
              </ul>
            </div>
          )}
        </li>
      ))}
    </ul>
  );
}
