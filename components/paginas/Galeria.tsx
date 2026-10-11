import { notFound } from "next/navigation";
import { Casca } from "@/components/Casca";
import { Filtros } from "@/components/galeria/Filtros";
import { ImagemGrid } from "@/components/galeria/ImagemGrid";
import { Lightbox } from "@/components/galeria/Lightbox";
import { catalogo, fotosDoTema, temaValido } from "@/lib/fotos";
import { prefixo, t, TEMAS_EN, type Idioma } from "@/lib/i18n";
import { metadados } from "@/lib/metadados";

export type Busca = Promise<{ tema?: string | string[] }>;

async function temaPedido(busca: Busca): Promise<string> {
  const { tema } = await busca;
  return typeof tema === "string" ? tema : "";
}

function rotulos(idioma: Idioma): Map<string, string> {
  return new Map(catalogo().temas.map((tema) => [tema.slug, (idioma === "en" && TEMAS_EN[tema.slug]) || tema.label]));
}

export async function metadadosGaleria(idioma: Idioma, busca: Busca) {
  const tema = await temaPedido(busca);
  const en = idioma === "en";
  const valido = tema !== "" && temaValido(tema);
  const base = en ? "Gallery — Isabel Pontes" : "Galeria — Isabel Pontes";
  return metadados({
    idioma,
    caminhoPT: valido ? `/galeria?tema=${tema}` : "/galeria",
    titulo: valido ? `${rotulos(idioma).get(tema)} — ${base}` : base,
    descricao: en
      ? "Fine art photography shoots by Isabel Pontes, in Curitiba."
      : "Ensaios de fotografia artística por Isabel Pontes, em Curitiba.",
  });
}

export async function Galeria({ idioma, busca }: { idioma: Idioma; busca: Busca }) {
  const tema = await temaPedido(busca);
  if (tema && !temaValido(tema)) notFound();

  const en = idioma === "en";
  const rotulo = rotulos(idioma);
  const filtros = [
    { slug: "", label: t(idioma, "comum.todos"), ativo: tema === "" },
    ...catalogo().temas.map(({ slug }) => ({ slug, label: rotulo.get(slug) ?? "", ativo: slug === tema })),
  ];
  const fotos = fotosDoTema(tema);

  return (
    <Casca>
      <section className="galeria">
        <div id="galeria-body">
          <Filtros filtros={filtros} caminho={`${prefixo(idioma)}/galeria`} rotulo={t(idioma, "galeria.filtrar")} />
          {fotos.length > 0 ? (
            <div className="masonry" key={tema}>
              {fotos.map((foto, indice) => {
                const temaLabel = rotulo.get(foto.tema) ?? "";
                const alt = foto.alt || (en ? `${temaLabel} shoot — Isabel Pontes` : `Ensaio de ${temaLabel} — Isabel Pontes`);
                return (
                  <button
                    key={foto.id}
                    type="button"
                    className="masonry__item"
                    data-idx={indice}
                    data-full={foto.full}
                    data-alt={alt}
                    data-tema-label={temaLabel}
                    style={{ aspectRatio: `${foto.w} / ${foto.h}`, animationDelay: `${(indice % 12) * 45}ms` }}
                    aria-label={`Abrir ${alt}`}
                  >
                    <img className="masonry__lqip" src={foto.lqip} alt="" aria-hidden="true" />
                    <ImagemGrid grid={foto.grid} grid2x={foto.grid2x} w={foto.w} h={foto.h} alt={alt} />
                    <span className="masonry__tema">{temaLabel}</span>
                  </button>
                );
              })}
            </div>
          ) : (
            <p className="galeria-vazia">{t(idioma, "galeria.vazia")}</p>
          )}
        </div>
      </section>
      <Lightbox />
    </Casca>
  );
}
