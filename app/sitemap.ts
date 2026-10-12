import type { MetadataRoute } from "next";
import { BASE_URL } from "@/lib/config";
import { formularioPublico } from "@/lib/forms";
import { catalogo } from "@/lib/fotos";
import { caminho } from "@/lib/i18n";

export const dynamic = "force-dynamic";

export default function sitemap(): MetadataRoute.Sitemap {
  const caminhos = [
    "/",
    "/galeria",
    ...catalogo().temas.map((tema) => `/galeria?tema=${tema.slug}`),
    "/sobre",
    "/midiakit",
    "/contato",
    ...(formularioPublico("fotografia") ? ["/briefing"] : []),
  ];
  return caminhos.flatMap((caminhoPT) => [{ url: BASE_URL + caminhoPT }, { url: BASE_URL + caminho("en", caminhoPT) }]);
}
