"use client";

import { useRouter } from "next/navigation";
import { useRef, useState, type ChangeEvent } from "react";

type Tema = { slug: string; label: string };

type Item = { id: number; nome: string; progresso: number; estado: "fila" | "enviando" | "processando" | "pronta" | "erro"; erro?: string };

const ROTULO: Record<Item["estado"], string> = {
  fila: "na fila",
  enviando: "enviando",
  processando: "preparando",
  pronta: "pronta",
  erro: "não entrou",
};

function enviar(arquivo: File, tema: string, aoProgredir: (fracao: number) => void): Promise<void> {
  return new Promise((resolver, rejeitar) => {
    const dados = new FormData();
    dados.append("tema", tema);
    dados.append("arquivo", arquivo);
    const pedido = new XMLHttpRequest();
    pedido.open("POST", "/admin/api/fotos");
    pedido.responseType = "json";
    pedido.upload.onprogress = (evento) => evento.lengthComputable && aoProgredir(evento.loaded / evento.total);
    pedido.onload = () => (pedido.status === 200 ? resolver() : rejeitar(new Error(pedido.response?.erro ?? "Não foi possível enviar.")));
    pedido.onerror = () => rejeitar(new Error("A conexão caiu durante o envio."));
    pedido.send(dados);
  });
}

export function Envio({ temas }: { temas: Tema[] }) {
  const router = useRouter();
  const [tema, setTema] = useState(temas[0]?.slug ?? "");
  const [itens, setItens] = useState<Item[]>([]);
  const [ocupado, setOcupado] = useState(false);
  const proximoId = useRef(0);

  function atualizar(id: number, mudanca: Partial<Item>) {
    setItens((lista) => lista.map((item) => (item.id === id ? { ...item, ...mudanca } : item)));
  }

  async function escolher(evento: ChangeEvent<HTMLInputElement>) {
    const arquivos = [...(evento.target.files ?? [])];
    evento.target.value = "";
    if (arquivos.length === 0) return;
    const novos = arquivos.map((arquivo) => ({ id: proximoId.current++, nome: arquivo.name, progresso: 0, estado: "fila" as const }));
    setItens((lista) => [...novos, ...lista.filter((item) => item.estado === "erro")]);
    setOcupado(true);
    for (const [indice, arquivo] of arquivos.entries()) {
      const { id } = novos[indice];
      atualizar(id, { estado: "enviando" });
      try {
        await enviar(arquivo, tema, (fracao) => atualizar(id, { progresso: fracao, estado: fracao >= 1 ? "processando" : "enviando" }));
        atualizar(id, { estado: "pronta", progresso: 1 });
        router.refresh();
      } catch (erro) {
        atualizar(id, { estado: "erro", erro: erro instanceof Error ? erro.message : "Não foi possível enviar." });
      }
    }
    setOcupado(false);
  }

  return (
    <section aria-label="Adicionar fotos">
      <div className="adm-envio">
        <label>
          Tema das fotos novas
          <select className="adm-campo" value={tema} onChange={(evento) => setTema(evento.target.value)} disabled={ocupado}>
            {temas.map((opcao) => (
              <option key={opcao.slug} value={opcao.slug}>
                {opcao.label}
              </option>
            ))}
          </select>
        </label>
        <label className="adm-botao adm-botao--principal" aria-disabled={ocupado}>
          {ocupado ? "Enviando…" : "Adicionar fotos"}
          <input type="file" accept="image/jpeg,image/png,image/webp" multiple hidden disabled={ocupado || !tema} onChange={escolher} />
        </label>
      </div>
      {itens.length > 0 && (
        <ul className="adm-fila" aria-live="polite">
          {itens.map((item) => (
            <li key={item.id}>
              <span className="adm-fila__nome">{item.nome}</span>
              <span>{ROTULO[item.estado]}</span>
              {(item.estado === "enviando" || item.estado === "processando") && (
                <progress value={item.estado === "processando" ? undefined : item.progresso} max={1} />
              )}
              {item.erro && <span className="adm-fila__erro">{item.erro}</span>}
            </li>
          ))}
        </ul>
      )}
    </section>
  );
}
