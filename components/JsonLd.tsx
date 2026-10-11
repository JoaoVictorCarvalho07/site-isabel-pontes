import { BASE_URL, NOME_SITE } from "@/lib/config";
import { EMAIL, INSTAGRAM } from "@/lib/contato";
import type { Idioma } from "@/lib/i18n";

export function JsonLd({ idioma }: { idioma: Idioma }) {
  const dados = {
    "@context": "https://schema.org",
    "@type": "ProfessionalService",
    name: NOME_SITE,
    description:
      idioma === "en"
        ? "Fine art photography, conceptual shoots and creative direction."
        : "Fotografia artística, ensaios conceituais e direção criativa.",
    url: `${BASE_URL}/`,
    image: `${BASE_URL}/static/img/og-image.jpg`,
    email: EMAIL,
    areaServed: "Curitiba, PR",
    address: {
      "@type": "PostalAddress",
      addressLocality: "Curitiba",
      addressRegion: "PR",
      addressCountry: "BR",
    },
    sameAs: [INSTAGRAM],
  };
  const json = JSON.stringify(dados).replace(/</g, "\\u003c");
  return <script type="application/ld+json" dangerouslySetInnerHTML={{ __html: json }} />;
}
