import type { ReactNode } from "react";
import { asset } from "@/lib/asset";
import { t } from "@/lib/i18n";
import { contexto } from "@/lib/requisicao";
import { JsonLd } from "./JsonLd";
import { Navbar } from "./Navbar";

type Props = { mostrarPdf?: boolean; ocultarRodape?: boolean; children: ReactNode };

export async function Casca({ mostrarPdf, ocultarRodape, children }: Props) {
  const { idioma, caminhoPT } = await contexto();
  return (
    <>
      <JsonLd idioma={idioma} />
      <a className="skip-link" href="#conteudo">
        {t(idioma, "a11y.pular")}
      </a>
      <Navbar idioma={idioma} caminhoPT={caminhoPT} mostrarPdf={mostrarPdf} />
      <main id="conteudo">{children}</main>
      {!ocultarRodape && (
        <footer className="rodape">
          <p>{`© ${new Date().getFullYear()} Isabel Pontes · ${t(idioma, "rodape")}`}</p>
        </footer>
      )}
      <script src={asset("/static/js/app.js")} defer />
    </>
  );
}
