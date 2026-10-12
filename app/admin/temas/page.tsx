import { ListaTemas } from "@/components/admin/ListaTemas";
import { Topo } from "@/components/admin/Topo";
import { listarTemas } from "@/lib/admin/fotos";
import { exigirAdmin } from "@/lib/admin/sessao";

export default async function Temas() {
  await exigirAdmin();
  return (
    <>
      <Topo atual="/admin/temas" />
      <main className="adm-corpo">
        <h1 className="adm-titulo">Temas da galeria</h1>
        <ListaTemas temas={listarTemas()} />
      </main>
    </>
  );
}
