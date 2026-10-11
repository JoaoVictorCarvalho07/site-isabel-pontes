import { Casca } from "@/components/Casca";
import { asset } from "@/lib/asset";
import { fotosPublicadas } from "@/lib/fotos";
import { metadados } from "@/lib/metadados";
import { prefixo, type Idioma } from "@/lib/i18n";

const PAISAGEM = [960, 1280, 1920, 2560];
const RETRATO = [640, 960, 1280, 1920];

function srcset(base: string, extensao: string, larguras: number[]) {
  return larguras.map((largura) => `${asset(`/static/img/hero/${base}-${largura}.${extensao}`)} ${largura}w`).join(", ");
}

export function metadadosHome(idioma: Idioma) {
  return metadados({
    idioma,
    caminhoPT: "/",
    titulo:
      idioma === "en"
        ? "Isabel Pontes — Fine art photography and creative direction in Curitiba"
        : "Isabel Pontes — Fotografia artística e direção criativa em Curitiba",
    descricao:
      idioma === "en"
        ? "Conceptual shoots, fine art photography and creative direction in Curitiba."
        : "Ensaios conceituais, fotografia artística e direção criativa em Curitiba.",
  });
}

export function Home({ idioma }: { idioma: Idioma }) {
  const en = idioma === "en";
  const p = prefixo(idioma);
  const fotoDoDia = fotosPublicadas()[0];
  return (
    <Casca>
      <section className="hero">
        <picture>
          <source media="(min-width:768px)" type="image/avif" sizes="100vw" srcSet={srcset("hero-land", "avif", PAISAGEM)} />
          <source media="(min-width:768px)" type="image/webp" sizes="100vw" srcSet={srcset("hero-land", "webp", PAISAGEM)} />
          <source type="image/avif" sizes="100vw" srcSet={srcset("hero", "avif", RETRATO)} />
          <source type="image/webp" sizes="100vw" srcSet={srcset("hero", "webp", RETRATO)} />
          <img
            src={asset("/static/img/hero/hero-1280.webp")}
            alt={en ? "Isabel Pontes, photographer and creative director, in an artistic shoot" : "Isabel Pontes, fotógrafa e diretora criativa, em ensaio artístico"}
            width={4000}
            height={6000}
            fetchPriority="high"
            decoding="async"
            className="hero__img"
          />
        </picture>
        <div className="hero__overlay"></div>
        <div className="hero__texto">
          <h1>{en ? "Welcome to my universe" : "Bem-vindo(a) ao meu universo"}</h1>
          <p>{en ? "Creative direction, fine art photography and storytelling." : "Direção criativa, fotografia artística e storytelling."}</p>
        </div>
      </section>
      <section className="cards-editoriais">
        <a href={`${p}/galeria`} className="card-ed">
          {fotoDoDia && (
            <img
              className="card-ed__img"
              src={fotoDoDia.full}
              alt={en ? "Artistic shoot — Isabel Pontes" : "Ensaio artístico — Isabel Pontes"}
              loading="lazy"
              decoding="async"
            />
          )}
          <div className="card-ed__overlay"></div>
          <div className="card-ed__texto">
            <h2>{en ? "Fine art photography" : "Fotografia artística"}</h2>
            <p>{en ? "Conceptual shoots and directing people." : "Ensaios conceituais e direção de pessoas."}</p>
          </div>
        </a>
        <a href={`${p}/midiakit`} className="card-ed">
          <img
            className="card-ed__img"
            src={asset("/static/img/MidiaKit.jpg")}
            alt={en ? "Media kit and creator content" : "Mídia kit e conteúdo de creator"}
            loading="lazy"
            decoding="async"
          />
          <div className="card-ed__overlay"></div>
          <div className="card-ed__texto">
            <h2>{en ? "Media kit & creator" : "Mídia kit & creator"}</h2>
            <p>{en ? "Authentic content with elevated aesthetics." : "Conteúdo autêntico com estética elevada."}</p>
          </div>
        </a>
      </section>
    </Casca>
  );
}
