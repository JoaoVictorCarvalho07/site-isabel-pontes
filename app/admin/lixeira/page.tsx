import { ListaFotos } from "@/components/admin/ListaFotos";
import { Topo } from "@/components/admin/Topo";
import { listarFotos, listarTemas } from "@/lib/admin/fotos";
import { exigirAdmin } from "@/lib/admin/sessao";

export default async function Lixeira() {
  await exigirAdmin();
  const fotos = listarFotos().filter((foto) => foto.naLixeira);
  return (
    <>
      <Topo atual="/admin/lixeira" />
      <main className="adm-corpo">
        <h1 className="adm-titulo">{`Lixeira · ${fotos.length}`}</h1>
        <ListaFotos fotos={fotos} temas={listarTemas()} lixeira />
      </main>
    </>
  );
}
