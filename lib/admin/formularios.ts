import { randomBytes } from "node:crypto";
import { banco } from "../banco";
import { casar, opcoesDoTexto, type Opcao } from "./texto";

export const TIPOS = [
  { valor: "text", rotulo: "Texto curto" },
  { valor: "textarea", rotulo: "Texto longo (parágrafo)" },
  { valor: "select", rotulo: "Lista de opções" },
  { valor: "contato", rotulo: "WhatsApp ou e-mail" },
  { valor: "email", rotulo: "E-mail" },
  { valor: "date", rotulo: "Data" },
  { valor: "file", rotulo: "Anexar imagens" },
] as const;

export type PerguntaEdicao = {
  id: number;
  chave: string;
  label: string;
  ajuda: string;
  tipo: string;
  obrigatoria: boolean;
  opcoes: string;
  oculta: boolean;
  ingles: boolean;
  respostas: number;
};

export type FormularioEdicao = {
  slug: string;
  titulo: string;
  subtitulo: string;
  enviar: string;
  sucesso: string;
  visivel: boolean;
  perguntas: PerguntaEdicao[];
};

export class EdicaoRecusada extends Error {}

type LinhaPergunta = {
  id: number;
  chave: string;
  label_pt: string;
  label_en: string;
  ajuda_pt: string;
  ajuda_en: string;
  tipo: string;
  obrigatoria: number;
  opcoes_json: string;
  sort_order: number;
  oculta: number;
  en_stale: number;
};

type LinhaForm = Record<"slug" | `${"titulo" | "subtitulo" | "enviar" | "sucesso"}_${"pt" | "en"}`, string> & { visivel: number };

function lerOpcoes(json: string): Opcao[] {
  return json ? (JSON.parse(json) as Opcao[]) : [];
}

function tipoValido(tipo: string): boolean {
  return TIPOS.some((item) => item.valor === tipo);
}

function contarRespostas(slug: string): Map<string, number> {
  const contagem = new Map<string, number>();
  const linhas = banco().prepare("SELECT payload FROM submissions WHERE form = ?").all(slug) as { payload: string }[];
  for (const { payload } of linhas) {
    try {
      for (const [chave, valor] of Object.entries(JSON.parse(payload) as Record<string, string>)) {
        if (valor !== "") contagem.set(chave, (contagem.get(chave) ?? 0) + 1);
      }
    } catch {
      continue;
    }
  }
  return contagem;
}

export function listarFormularios(): { slug: string; titulo: string; perguntas: number; visivel: boolean }[] {
  return (
    banco()
      .prepare(
        `SELECT f.slug, f.titulo_pt AS titulo, f.visivel,
                (SELECT count(*) FROM form_questions q WHERE q.form_slug = f.slug) AS perguntas
           FROM forms f ORDER BY f.sort_order, f.slug`,
      )
      .all() as { slug: string; titulo: string; perguntas: number; visivel: number }[]
  ).map((form) => ({ ...form, visivel: form.visivel === 1 }));
}

function linhasDe(slug: string): LinhaPergunta[] {
  return banco().prepare("SELECT * FROM form_questions WHERE form_slug = ? ORDER BY sort_order, id").all(slug) as LinhaPergunta[];
}

export function carregarFormulario(slug: string): FormularioEdicao | null {
  const form = banco().prepare("SELECT * FROM forms WHERE slug = ?").get(slug) as LinhaForm | undefined;
  if (!form) return null;
  const respostas = contarRespostas(slug);
  return {
    slug,
    titulo: form.titulo_pt,
    subtitulo: form.subtitulo_pt,
    enviar: form.enviar_pt,
    sucesso: form.sucesso_pt,
    visivel: form.visivel === 1,
    perguntas: linhasDe(slug).map((linha) => ({
      id: linha.id,
      chave: linha.chave,
      label: linha.label_pt,
      ajuda: linha.ajuda_pt,
      tipo: linha.tipo,
      obrigatoria: linha.obrigatoria === 1,
      opcoes: lerOpcoes(linha.opcoes_json)
        .map((opcao) => opcao.pt)
        .join("\n"),
      oculta: linha.oculta === 1,
      ingles: linha.en_stale === 1,
      respostas: respostas.get(linha.chave) ?? 0,
    })),
  };
}

function pergunta(slug: string, id: number): LinhaPergunta {
  const linha = banco().prepare("SELECT * FROM form_questions WHERE id = ? AND form_slug = ?").get(id, slug) as LinhaPergunta | undefined;
  if (!linha) throw new EdicaoRecusada("Essa pergunta não existe mais.");
  return linha;
}

export function salvarTextos(slug: string, novo: { titulo: string; subtitulo: string; enviar: string; sucesso: string }): void {
  const form = banco().prepare("SELECT * FROM forms WHERE slug = ?").get(slug) as LinhaForm | undefined;
  if (!form) throw new EdicaoRecusada("Esse formulário não existe.");
  const limpo = { titulo: novo.titulo.trim(), subtitulo: novo.subtitulo.trim(), enviar: novo.enviar.trim(), sucesso: novo.sucesso.trim() };
  if (!limpo.titulo || !limpo.enviar) throw new EdicaoRecusada("O título e o texto do botão não podem ficar vazios.");
  const [tituloPT, tituloEN] = casar(form.titulo_pt, form.titulo_en, limpo.titulo);
  const [subtituloPT, subtituloEN] = casar(form.subtitulo_pt, form.subtitulo_en, limpo.subtitulo);
  const [enviarPT, enviarEN] = casar(form.enviar_pt, form.enviar_en, limpo.enviar);
  const [sucessoPT, sucessoEN] = casar(form.sucesso_pt, form.sucesso_en, limpo.sucesso);
  banco()
    .prepare(
      `UPDATE forms SET titulo_pt=?, titulo_en=?, subtitulo_pt=?, subtitulo_en=?, enviar_pt=?, enviar_en=?,
                        sucesso_pt=?, sucesso_en=?, updated_at=datetime('now') WHERE slug=?`,
    )
    .run(tituloPT, tituloEN, subtituloPT, subtituloEN, enviarPT, enviarEN, sucessoPT, sucessoEN, slug);
}

function opcoesObrigatorias(texto: string, anteriores: Opcao[] = []): string {
  const opcoes = opcoesDoTexto(texto, anteriores);
  if (opcoes.length === 0) throw new EdicaoRecusada("Uma lista precisa de pelo menos uma opção.");
  return JSON.stringify(opcoes);
}

export function salvarPergunta(slug: string, id: number, novo: { label: string; ajuda: string; obrigatoria: boolean; opcoes: string }): void {
  const atual = pergunta(slug, id);
  const label = novo.label.trim();
  const ajuda = novo.ajuda.trim();
  if (!label) throw new EdicaoRecusada("A pergunta não pode ficar sem texto.");
  let revisar = atual.en_stale === 1;
  const [labelPT, labelEN] = casar(atual.label_pt, atual.label_en, label);
  const [ajudaPT, ajudaEN] = casar(atual.ajuda_pt, atual.ajuda_en, ajuda);
  if (labelPT !== atual.label_pt || ajudaPT !== atual.ajuda_pt) revisar = true;
  let opcoes = atual.opcoes_json;
  if (atual.tipo === "select") {
    opcoes = opcoesObrigatorias(novo.opcoes, lerOpcoes(atual.opcoes_json));
    if (opcoes !== atual.opcoes_json) revisar = true;
  }
  banco()
    .prepare(
      `UPDATE form_questions SET label_pt=?, label_en=?, ajuda_pt=?, ajuda_en=?, obrigatoria=?, opcoes_json=?, en_stale=?,
                                 updated_at=datetime('now') WHERE id=?`,
    )
    .run(labelPT, labelEN, ajudaPT, ajudaEN, novo.obrigatoria ? 1 : 0, opcoes, revisar ? 1 : 0, id);
}

export function adicionarPergunta(slug: string, nova: { label: string; ajuda: string; tipo: string; obrigatoria: boolean; opcoes: string }): void {
  if (!banco().prepare("SELECT 1 FROM forms WHERE slug = ?").get(slug)) throw new EdicaoRecusada("Esse formulário não existe.");
  const label = nova.label.trim();
  if (!label || !tipoValido(nova.tipo)) throw new EdicaoRecusada("Escreva a pergunta e escolha o tipo.");
  const opcoes = nova.tipo === "select" ? opcoesObrigatorias(nova.opcoes) : "";
  const { proxima } = banco().prepare("SELECT coalesce(max(sort_order), 0) + 10 AS proxima FROM form_questions WHERE form_slug = ?").get(slug) as {
    proxima: number;
  };
  banco()
    .prepare(
      `INSERT INTO form_questions (form_slug, chave, label_pt, label_en, ajuda_pt, ajuda_en, tipo, obrigatoria, opcoes_json, sort_order, oculta, en_stale)
       VALUES (?, ?, ?, '', ?, '', ?, ?, ?, ?, 0, 1)`,
    )
    .run(slug, `q_${randomBytes(4).toString("hex")}`, label, nova.ajuda.trim(), nova.tipo, nova.obrigatoria ? 1 : 0, opcoes, proxima);
}

export function moverPergunta(slug: string, id: number, direcao: "antes" | "depois"): void {
  const linhas = linhasDe(slug);
  const posicao = linhas.findIndex((linha) => linha.id === id);
  const vizinha = linhas[posicao + (direcao === "antes" ? -1 : 1)];
  if (posicao < 0 || !vizinha) return;
  banco().transaction(() => {
    banco().prepare("UPDATE form_questions SET sort_order = ? WHERE id = ?").run(vizinha.sort_order, id);
    banco().prepare("UPDATE form_questions SET sort_order = ? WHERE id = ?").run(linhas[posicao].sort_order, vizinha.id);
  })();
}

export function ocultarPergunta(slug: string, id: number, oculta: boolean): void {
  pergunta(slug, id);
  banco().prepare("UPDATE form_questions SET oculta = ?, updated_at = datetime('now') WHERE id = ?").run(oculta ? 1 : 0, id);
}

export function apagarPergunta(slug: string, id: number): void {
  const atual = pergunta(slug, id);
  if ((contarRespostas(slug).get(atual.chave) ?? 0) > 0) {
    throw new EdicaoRecusada("Essa pergunta já tem respostas. Oculte em vez de apagar, para não perder o histórico.");
  }
  banco().prepare("DELETE FROM form_questions WHERE id = ?").run(id);
}

export function trocarTipo(slug: string, id: number, tipo: string, opcoes: string): void {
  pergunta(slug, id);
  if (!tipoValido(tipo)) throw new EdicaoRecusada("Escolha um tipo da lista.");
  const json = tipo === "select" ? opcoesObrigatorias(opcoes) : "";
  banco().prepare("UPDATE form_questions SET tipo = ?, opcoes_json = ?, en_stale = 1, updated_at = datetime('now') WHERE id = ?").run(tipo, json, id);
}
