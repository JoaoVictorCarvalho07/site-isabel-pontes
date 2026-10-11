import { banco } from "./banco";
import type { Par } from "./midiakit/conteudo";

export type TipoCampo = "text" | "textarea" | "select" | "contato" | "email" | "tel" | "date" | "file";

export type Pergunta = {
  id: number;
  chave: string;
  label: Par;
  ajuda: Par;
  tipo: TipoCampo;
  obrigatoria: boolean;
  opcoes: Par[];
};

export type Formulario = {
  slug: string;
  titulo: Par;
  subtitulo: Par;
  enviar: Par;
  sucesso: Par;
  perguntas: Pergunta[];
};

type LinhaForm = Record<"slug" | `${"titulo" | "subtitulo" | "enviar" | "sucesso"}_${"pt" | "en"}`, string>;

type LinhaPergunta = {
  id: number;
  chave: string;
  label_pt: string;
  label_en: string;
  ajuda_pt: string;
  ajuda_en: string;
  tipo: TipoCampo;
  obrigatoria: number;
  opcoes_json: string;
};

function opcoes(json: string): Par[] {
  if (!json) return [];
  return (JSON.parse(json) as { pt: string; en: string }[]).map(({ pt, en }) => [pt, en] as const);
}

export function formularioPublico(slug: string): Formulario | null {
  const form = banco().prepare("SELECT * FROM forms WHERE slug = ? AND visivel = 1").get(slug) as LinhaForm | undefined;
  if (!form) return null;
  const perguntas = banco()
    .prepare(
      `SELECT id, chave, label_pt, label_en, ajuda_pt, ajuda_en, tipo, obrigatoria, opcoes_json
         FROM form_questions
        WHERE form_slug = ? AND oculta = 0
        ORDER BY sort_order, id`,
    )
    .all(slug) as LinhaPergunta[];
  return {
    slug: form.slug,
    titulo: [form.titulo_pt, form.titulo_en],
    subtitulo: [form.subtitulo_pt, form.subtitulo_en],
    enviar: [form.enviar_pt, form.enviar_en],
    sucesso: [form.sucesso_pt, form.sucesso_en],
    perguntas: perguntas.map((p) => ({
      id: p.id,
      chave: p.chave,
      label: [p.label_pt, p.label_en],
      ajuda: [p.ajuda_pt, p.ajuda_en],
      tipo: p.tipo,
      obrigatoria: p.obrigatoria === 1,
      opcoes: opcoes(p.opcoes_json),
    })),
  };
}
