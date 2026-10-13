import { ImagensDoSite } from "@/components/admin/ImagensDoSite";
import { Topo } from "@/components/admin/Topo";
import { exigirAdmin } from "@/lib/admin/sessao";
import { carrossel, endereco, imagemFixa } from "@/lib/layout/imagens";
import { arquivo, CARROSSEL, SLOTS, type NomeSlot } from "@/lib/layout/slots";

export default async function Site() {
  await exigirAdmin();
  const slots = (Object.keys(SLOTS) as NomeSlot[]).map((nome) => {
    const slot = SLOTS[nome];
    const atual = imagemFixa(nome);
    return {
      nome,
      rotulo: slot.rotulo,
      dica: slot.dica,
      miniatura: endereco(arquivo(atual.base, slot, slot.larguras[0], slot.formatos.at(-1) ?? "webp")),
      original: atual.base === slot.padrao.base,
    };
  });
  const itens = carrossel().map((item) => ({ id: item.id, miniatura: endereco(arquivo(item.base, CARROSSEL, 800, "webp")) }));
  return (
    <>
      <Topo atual="/admin/site" />
      <main className="adm-corpo">
        <h1 className="adm-titulo">Imagens fixas do site</h1>
        <ImagensDoSite slots={slots} carrossel={itens} />
      </main>
    </>
  );
}
