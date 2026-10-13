"use client";

import { useRouter } from "next/navigation";
import { useState, useTransition, type ChangeEvent } from "react";
import * as acoes from "@/lib/admin/acoes";

export type CartaoSlot = { nome: string; rotulo: string; dica: string; miniatura: string; original: boolean };

export type CartaoCarrossel = { id: number; miniatura: string };

type Props = { slots: CartaoSlot[]; carrossel: CartaoCarrossel[] };

async function enviar(destino: string, arquivo: File): Promise<string | undefined> {
  const dados = new FormData();
  dados.append("destino", destino);
  dados.append("arquivo", arquivo);
  try {
    const resposta = await fetch("/admin/api/layout", { method: "POST", body: dados });
    return resposta.ok ? undefined : ((await resposta.json()) as { erro?: string }).erro ?? "Não foi possível enviar.";
  } catch {
    return "A conexão caiu durante o envio.";
  }
}

export function ImagensDoSite({ slots, carrossel }: Props) {
  const router = useRouter();
  const [ocupado, iniciar] = useTransition();
  const [enviando, setEnviando] = useState("");
  const [aviso, setAviso] = useState("");
  const parado = ocupado || enviando !== "";

  async function escolher(destino: string, evento: ChangeEvent<HTMLInputElement>) {
    const escolhido = evento.target.files?.[0];
    evento.target.value = "";
    if (!escolhido) return;
    setEnviando(destino);
    setAviso((await enviar(destino, escolhido)) ?? "");
    setEnviando("");
    router.refresh();
  }

  function executar(acao: () => Promise<{ erro?: string }>) {
    iniciar(async () => {
      setAviso((await acao()).erro ?? "");
      router.refresh();
    });
  }

  return (
    <>
      {aviso && (
        <p className="adm-aviso" role="alert">
          {aviso}
        </p>
      )}
      <ul className="adm-grade">
        {slots.map((slot) => (
          <li key={slot.nome} className="adm-foto" data-slot={slot.nome}>
            <img className="adm-foto__img" src={slot.miniatura} alt="" loading="lazy" decoding="async" />
            <div className="adm-foto__dados">
              <strong>{slot.rotulo}</strong>
              <span className="adm-nota">{slot.dica}</span>
            </div>
            <div className="adm-foto__acoes">
              <label className="adm-botao adm-botao--principal" aria-disabled={parado}>
                {enviando === slot.nome ? "Preparando…" : "Trocar foto"}
                <input type="file" accept="image/jpeg,image/png,image/webp" hidden disabled={parado} onChange={(evento) => escolher(slot.nome, evento)} />
              </label>
              {!slot.original && (
                <button className="adm-botao" disabled={parado} onClick={() => executar(() => acoes.restaurarImagemFixa(slot.nome))}>
                  Voltar à original
                </button>
              )}
            </div>
          </li>
        ))}
      </ul>

      <h2 className="adm-titulo">{`Carrossel da página Sobre · ${carrossel.length}`}</h2>
      <label className="adm-botao adm-botao--principal" aria-disabled={parado}>
        {enviando === "carrossel" ? "Preparando…" : "Adicionar ao carrossel"}
        <input type="file" accept="image/jpeg,image/png,image/webp" hidden disabled={parado} onChange={(evento) => escolher("carrossel", evento)} />
      </label>
      <ul className="adm-grade adm-grade--carrossel">
        {carrossel.map((item, indice) => (
          <li key={item.id} className="adm-foto" data-carrossel={item.id}>
            <img className="adm-foto__img" src={item.miniatura} alt="" loading="lazy" decoding="async" />
            <div className="adm-foto__dados">
              <strong>{`${indice + 1}ª foto`}</strong>
            </div>
            <div className="adm-foto__acoes">
              <button className="adm-botao" aria-label="Mover para antes" disabled={parado || indice === 0} onClick={() => executar(() => acoes.moverNoCarrossel(item.id, "antes"))}>
                ↑
              </button>
              <button
                className="adm-botao"
                aria-label="Mover para depois"
                disabled={parado || indice === carrossel.length - 1}
                onClick={() => executar(() => acoes.moverNoCarrossel(item.id, "depois"))}
              >
                ↓
              </button>
              <button
                className="adm-botao adm-botao--perigo"
                disabled={parado || carrossel.length <= 1}
                onClick={() => confirm("Tirar esta foto do carrossel?") && executar(() => acoes.removerDoCarrossel(item.id))}
              >
                Remover
              </button>
            </div>
          </li>
        ))}
      </ul>
    </>
  );
}
