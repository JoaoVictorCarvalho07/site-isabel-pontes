import { Topo } from "@/components/admin/Topo";
import { listarFormularios } from "@/lib/admin/formularios";
import { exigirAdmin } from "@/lib/admin/sessao";

export default async function Formularios() {
  await exigirAdmin();
  return (
    <>
      <Topo atual="/admin/formularios" />
      <main className="adm-corpo">
        <h1 className="adm-titulo">Formulários</h1>
        <ul className="adm-temas">
          {listarFormularios().map((form) => (
            <li key={form.slug}>
              <a href={`/admin/formularios/${form.slug}`}>{form.titulo}</a>
              <small>{`${form.perguntas} perguntas${form.visivel ? "" : " · desligado"}`}</small>
            </li>
          ))}
        </ul>
      </main>
    </>
  );
}
