import { banco } from "../banco";
import { trechosComLinks, type Trecho } from "./texto";

export type Anexo = { id: number; nome: string; kb: number };

export type Briefing = {
  id: number;
  form: string;
  lida: boolean;
  quando: string;
  destaque: string;
  campos: { rotulo: string; trechos: Trecho[] }[];
  anexos: Anexo[];
};

type Linha = { id: number; form: string; payload: string; lida: number; created_at: string };

type Pergunta = { form_slug: string; chave: string; label_pt: string; tipo: string; obrigatoria: number };

const dataHora = new Intl.DateTimeFormat("pt-BR", { timeZone: "America/Sao_Paulo", dateStyle: "short", timeStyle: "short" });

function respostas(payload: string): Record<string, string> {
  try {
    return JSON.parse(payload) as Record<string, string>;
  } catch {
    return { _erro: `payload ilegível: ${payload}` };
  }
}

export function formulariosConhecidos(): { slug: string; titulo: string }[] {
  return banco().prepare("SELECT slug, titulo_pt AS titulo FROM forms ORDER BY sort_order, slug").all() as { slug: string; titulo: string }[];
}

export function contarNaoLidas(): number {
  return (banco().prepare("SELECT count(*) AS n FROM submissions WHERE lida = 0").get() as { n: number }).n;
}

export function listarBriefings(filtro: string): Briefing[] {
  const linhas = (
    filtro
      ? banco().prepare("SELECT id, form, payload, lida, created_at FROM submissions WHERE form = ? ORDER BY created_at DESC, id DESC LIMIT 200").all(filtro)
      : banco().prepare("SELECT id, form, payload, lida, created_at FROM submissions ORDER BY created_at DESC, id DESC LIMIT 200").all()
  ) as Linha[];
  const perguntas = banco()
    .prepare("SELECT form_slug, chave, label_pt, tipo, obrigatoria FROM form_questions ORDER BY form_slug, sort_order, id")
    .all() as Pergunta[];
  const anexos = banco().prepare("SELECT id, submission_id, nome_original, tamanho FROM submission_files ORDER BY id").all() as {
    id: number;
    submission_id: number;
    nome_original: string;
    tamanho: number;
  }[];

  return linhas.map((linha) => {
    const dadas = respostas(linha.payload);
    const doForm = perguntas.filter((pergunta) => pergunta.form_slug === linha.form);
    const pares =
      doForm.length > 0
        ? doForm.filter((pergunta) => dadas[pergunta.chave]).map((pergunta) => ({ pergunta, valor: dadas[pergunta.chave] }))
        : Object.entries(dadas).map(([chave, valor]) => ({ pergunta: { chave, label_pt: chave, tipo: "", obrigatoria: 0 } as Pergunta, valor }));
    const destaque = pares.find(({ pergunta }) => pergunta.obrigatoria === 1 && pergunta.tipo === "text")?.valor ?? "";
    return {
      id: linha.id,
      form: linha.form,
      lida: linha.lida === 1,
      quando: dataHora.format(new Date(`${linha.created_at.replace(" ", "T")}Z`)),
      destaque,
      campos: pares.map(({ pergunta, valor }) => ({ rotulo: pergunta.label_pt, trechos: trechosComLinks(valor) })),
      anexos: anexos
        .filter((anexo) => anexo.submission_id === linha.id)
        .map((anexo) => ({ id: anexo.id, nome: anexo.nome_original, kb: Math.ceil(anexo.tamanho / 1024) })),
    };
  });
}

export function marcarLida(id: number, lida: boolean): void {
  banco().prepare("UPDATE submissions SET lida = ? WHERE id = ?").run(lida ? 1 : 0, id);
}

export function anexoPorId(id: number): { caminho: string; nome: string; mime: string } | undefined {
  return banco().prepare("SELECT caminho, nome_original AS nome, mime FROM submission_files WHERE id = ?").get(id) as
    | { caminho: string; nome: string; mime: string }
    | undefined;
}
