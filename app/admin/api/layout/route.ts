import { FotoRecusada } from "@/lib/admin/fotos";
import { adicionarAoCarrossel, trocarImagem } from "@/lib/admin/layout";
import { emailAdmin } from "@/lib/admin/sessao";
import { ehSlot } from "@/lib/layout/slots";

export async function POST(requisicao: Request) {
  if (!(await emailAdmin())) return Response.json({ erro: "Sua sessão expirou. Entre de novo." }, { status: 401 });
  const dados = await requisicao.formData();
  const enviado = dados.get("arquivo");
  const destino = String(dados.get("destino") ?? "");
  if (!(enviado instanceof File)) return Response.json({ erro: "Nenhuma foto enviada." }, { status: 400 });
  if (destino !== "carrossel" && !ehSlot(destino)) return Response.json({ erro: "Lugar desconhecido." }, { status: 400 });
  try {
    const bytes = new Uint8Array(await enviado.arrayBuffer());
    if (destino === "carrossel") await adicionarAoCarrossel(bytes);
    else await trocarImagem(destino, bytes);
    return Response.json({});
  } catch (erro) {
    if (erro instanceof FotoRecusada) return Response.json({ erro: erro.message }, { status: 422 });
    console.error("admin: troca de imagem falhou", erro);
    return Response.json({ erro: "Algo deu errado ao salvar a foto. Tente de novo." }, { status: 500 });
  }
}
