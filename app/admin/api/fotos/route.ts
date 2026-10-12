import { adicionarFoto, FotoRecusada, TAMANHO_MAXIMO_FOTO } from "@/lib/admin/fotos";
import { emailAdmin } from "@/lib/admin/sessao";

export async function POST(requisicao: Request) {
  if (!(await emailAdmin())) return Response.json({ erro: "Sua sessão expirou. Entre de novo." }, { status: 401 });
  if (Number(requisicao.headers.get("content-length") ?? 0) > TAMANHO_MAXIMO_FOTO + 1024 * 1024) {
    return Response.json({ erro: "A foto passa do limite de 40 MB." }, { status: 413 });
  }
  const dados = await requisicao.formData();
  const arquivo = dados.get("arquivo");
  if (!(arquivo instanceof File)) return Response.json({ erro: "Nenhuma foto enviada." }, { status: 400 });
  try {
    const foto = await adicionarFoto(arquivo.name, new Uint8Array(await arquivo.arrayBuffer()), String(dados.get("tema") ?? ""));
    return Response.json({ foto });
  } catch (erro) {
    if (erro instanceof FotoRecusada) return Response.json({ erro: erro.message }, { status: 422 });
    console.error("admin: upload falhou", erro);
    return Response.json({ erro: "Algo deu errado ao salvar a foto. Tente de novo." }, { status: 500 });
  }
}
