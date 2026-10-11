import type { Metadata } from "next";
import { asset } from "./asset";
import { BASE_URL, NOME_SITE } from "./config";
import { caminho, type Idioma } from "./i18n";

type Pagina = {
  idioma: Idioma;
  titulo: string;
  descricao?: string;
  caminhoPT: string;
};

export function metadados({ idioma, titulo, descricao, caminhoPT }: Pagina): Metadata {
  const hrefPT = BASE_URL + caminhoPT;
  const hrefEN = BASE_URL + caminho("en", caminhoPT);
  const canonical = idioma === "en" ? hrefEN : hrefPT;
  const imagem = `${BASE_URL}/static/img/og-image.jpg`;
  return {
    title: titulo,
    description: descricao,
    alternates: {
      canonical,
      languages: { "pt-BR": hrefPT, "x-default": hrefPT, en: hrefEN },
    },
    icons: {
      icon: { url: asset("/static/img/favicon-32.png"), type: "image/png", sizes: "32x32" },
      apple: asset("/static/img/apple-touch-icon.png"),
    },
    openGraph: {
      type: "website",
      siteName: NOME_SITE,
      locale: idioma === "en" ? "en_US" : "pt_BR",
      url: canonical,
      title: titulo,
      description: descricao,
      images: [
        {
          url: imagem,
          width: 1200,
          height: 630,
          alt: "Retrato artístico com coroa dourada, ao lado de um busto clássico",
        },
      ],
    },
    twitter: { card: "summary_large_image", title: titulo, description: descricao, images: [imagem] },
  };
}
