import { Envio } from "@/components/admin/Envio";
import { ListaFotos } from "@/components/admin/ListaFotos";
import { Topo } from "@/components/admin/Topo";
import { listarFotos, listarTemas } from "@/lib/admin/fotos";
import { exigirAdmin } from "@/lib/admin/sessao";

export default async function Fotos() {
  await exigirAdmin();
  const temas = listarTemas();
  const fotos = listarFotos().filter((foto) => !foto.naLixeira);
  return (
    <>
      <Topo atual="/admin/fotos" />
      <main className="adm-corpo">
        <h1 className="adm-titulo">{`Fotos do site · ${fotos.length}`}</h1>
        <Envio temas={temas} />
        <ListaFotos fotos={fotos} temas={temas} />
      </main>
    </>
  );
}
