import { Casca } from "@/components/Casca";
import { contexto } from "@/lib/requisicao";

export default async function NaoEncontrada() {
  const { idioma } = await contexto();
  const en = idioma === "en";
  return (
    <Casca>
      <section className="erro">
        <h1>{en ? "Page not found" : "Página não encontrada"}</h1>
        <p>{en ? "The page you tried to reach doesn't exist or was moved." : "O endereço que você acessou não existe ou foi movido."}</p>
        <p>
          <a href={en ? "/en" : "/"}>{en ? "Back to home" : "Voltar para a página inicial"}</a>
        </p>
      </section>
    </Casca>
  );
}
