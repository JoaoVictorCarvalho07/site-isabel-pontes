import { Casca } from "@/components/Casca";
import { catalogo } from "@/lib/fotos";
import { endereco, imagemFixa } from "@/lib/layout/imagens";
import { arquivo, SLOTS, type NomeSlot } from "@/lib/layout/slots";
import { metadados } from "@/lib/metadados";
import { prefixo, type Idioma } from "@/lib/i18n";

function srcset(nome: NomeSlot, formato: "avif" | "webp") {
  const slot = SLOTS[nome];
  const { base } = imagemFixa(nome);
  return slot.larguras.map((largura) => `${endereco(arquivo(base, slot, largura, formato))} ${largura}w`).join(", ");
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
  const fotoDoDia = catalogo().todas[0];
  const retrato = imagemFixa("hero-retrato");
  const cartao = imagemFixa("card-midiakit");
  return (
    <Casca>
      <section className="hero">
        <picture>
          <source media="(min-width:768px)" type="image/avif" sizes="100vw" srcSet={srcset("hero-paisagem", "avif")} />
          <source media="(min-width:768px)" type="image/webp" sizes="100vw" srcSet={srcset("hero-paisagem", "webp")} />
          <source type="image/avif" sizes="100vw" srcSet={srcset("hero-retrato", "avif")} />
          <source type="image/webp" sizes="100vw" srcSet={srcset("hero-retrato", "webp")} />
          <img
            src={endereco(arquivo(retrato.base, SLOTS["hero-retrato"], 1280, "webp"))}
            alt={en ? "Isabel Pontes, photographer and creative director, in an artistic shoot" : "Isabel Pontes, fotógrafa e diretora criativa, em ensaio artístico"}
            width={retrato.w}
            height={retrato.h}
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
            src={endereco(arquivo(cartao.base, SLOTS["card-midiakit"], 1080, "jpg"))}
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
