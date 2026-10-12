import { ListaBriefings } from "@/components/admin/ListaBriefings";
import { Topo } from "@/components/admin/Topo";
import { contarNaoLidas, formulariosConhecidos, listarBriefings } from "@/lib/admin/briefings";
import { exigirAdmin } from "@/lib/admin/sessao";

export default async function Briefings({ searchParams }: { searchParams: Promise<{ form?: string }> }) {
  await exigirAdmin();
  const formularios = formulariosConhecidos();
  const pedido = (await searchParams).form ?? "";
  const filtro = formularios.some((form) => form.slug === pedido) ? pedido : "";
  const briefings = listarBriefings(filtro);
  const naoLidas = contarNaoLidas();
  const emailAtivo = Boolean(process.env.RESEND_API_KEY && process.env.MAIL_PARA);
  return (
    <>
      <Topo atual="/admin/briefings" />
      <main className="adm-corpo">
        <h1 className="adm-titulo">
          {`Briefings · ${briefings.length} ${briefings.length === 1 ? "resposta" : "respostas"}${naoLidas ? ` · ${naoLidas} não ${naoLidas === 1 ? "lida" : "lidas"}` : ""}`}
        </h1>
        <nav className="adm-abas" aria-label="Filtrar por formulário">
          <a href="/admin/briefings" aria-current={filtro === "" ? "page" : undefined}>
            Todos
          </a>
          {formularios.map((form) => (
            <a key={form.slug} href={`/admin/briefings?form=${form.slug}`} aria-current={filtro === form.slug ? "page" : undefined}>
              {form.slug}
            </a>
          ))}
        </nav>
        {!emailAtivo && (
          <p className="adm-nota">O aviso por e-mail está desligado. As respostas continuam sendo salvas aqui; só não chega mensagem na caixa de entrada.</p>
        )}
        <ListaBriefings briefings={briefings} />
      </main>
    </>
  );
}
