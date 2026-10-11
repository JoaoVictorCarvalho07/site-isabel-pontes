import type { Metadata, Viewport } from "next";
import type { ReactNode } from "react";
import "@/assets/scss/app.scss";
import { contexto } from "@/lib/requisicao";

const SCRIPT_TEMA = `(function(){try{var t=localStorage.getItem("tema");if(t!=="dark"&&t!=="light"){t=matchMedia("(prefers-color-scheme: dark)").matches?"dark":"light"}document.documentElement.dataset.theme=t}catch(e){}})();`;

const PAGINAS_HERO_ESCURO = new Set(["/", "/midiakit"]);

export const viewport: Viewport = { themeColor: "#3a2a22" };

export async function generateMetadata(): Promise<Metadata> {
  const { idioma } = await contexto();
  return { title: `${idioma === "en" ? "Page not found" : "Página não encontrada"} — Isabel Pontes` };
}

export default async function Layout({ children }: { children: ReactNode }) {
  const { idioma, caminhoPT } = await contexto();
  return (
    <html lang={idioma === "en" ? "en" : "pt-br"} suppressHydrationWarning>
      <head>
        <script dangerouslySetInnerHTML={{ __html: SCRIPT_TEMA }} />
      </head>
      <body className={PAGINAS_HERO_ESCURO.has(caminhoPT) ? "hero-escuro" : undefined} suppressHydrationWarning>
        {children}
      </body>
    </html>
  );
}
