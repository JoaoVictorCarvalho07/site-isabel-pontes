import { redirect } from "next/navigation";
import { entrar } from "@/lib/admin/acoes";
import { emailAdmin } from "@/lib/admin/sessao";

export default async function Entrar({ searchParams }: { searchParams: Promise<{ error?: string }> }) {
  if (await emailAdmin()) redirect("/admin/fotos");
  const { error } = await searchParams;
  return (
    <main className="adm-entrar">
      <div>
        <h1 className="adm-titulo">Painel do site</h1>
        <p>Entre com a conta Google autorizada para gerenciar as fotos.</p>
        {error && <p className="adm-aviso">Essa conta não tem acesso ao painel.</p>}
        <form action={entrar}>
          <button className="adm-botao adm-botao--principal" type="submit">
            Entrar com Google
          </button>
        </form>
      </div>
    </main>
  );
}
