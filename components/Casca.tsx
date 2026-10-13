import type { ReactNode } from "react";
import { asset } from "@/lib/asset";
import { t } from "@/lib/i18n";
import { contexto } from "@/lib/requisicao";
import { JsonLd } from "./JsonLd";
import { Navbar } from "./Navbar";
import { Rodape } from "./Rodape";

type Props = { mostrarPdf?: boolean; rodapeColado?: boolean; children: ReactNode };

export async function Casca({ mostrarPdf, rodapeColado, children }: Props) {
  const { idioma, caminhoPT } = await contexto();
  return (
    <>
      <JsonLd idioma={idioma} />
      <a className="skip-link" href="#conteudo">
        {t(idioma, "a11y.pular")}
      </a>
      <Navbar idioma={idioma} caminhoPT={caminhoPT} mostrarPdf={mostrarPdf} />
      <main id="conteudo">{children}</main>
      <Rodape idioma={idioma} colado={rodapeColado} />
      <script src={asset("/static/js/app.js")} defer />
    </>
  );
}
