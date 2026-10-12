import { notFound } from "next/navigation";
import { EditorFormulario } from "@/components/admin/EditorFormulario";
import { Topo } from "@/components/admin/Topo";
import { carregarFormulario, TIPOS } from "@/lib/admin/formularios";
import { exigirAdmin } from "@/lib/admin/sessao";

const PREVIA: Record<string, string> = { fotografia: "/briefing", ugc: "/midiakit#briefing" };

export default async function Formulario({ params }: { params: Promise<{ slug: string }> }) {
  await exigirAdmin();
  const form = carregarFormulario((await params).slug);
  if (!form) notFound();
  return (
    <>
      <Topo atual="/admin/formularios" />
      <main className="adm-corpo">
        <h1 className="adm-titulo">{`Editar: ${form.titulo}`}</h1>
        <EditorFormulario form={form} tipos={[...TIPOS]} previa={PREVIA[form.slug] ?? "/briefing"} />
      </main>
    </>
  );
}
