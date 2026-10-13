import { Casca } from "@/components/Casca";
import { Picture } from "@/components/Picture";
import { carrossel, imagemFixa } from "@/lib/layout/imagens";
import { metadados } from "@/lib/metadados";
import type { Idioma } from "@/lib/i18n";

const TEXTO = {
  pt: {
    inicio:
      "Meu nome é Isabel Pontes, sou fotógrafa artística e criadora visual atuando em Curitiba. Minha relação com a arte começou muito antes da fotografia entrar oficialmente na minha vida. Desde criança sempre fui movida por imaginação, movimento e criação. Participei de companhia de dança e sempre tive uma curiosidade natural sobre expressão artística. Gostava de montar cenários, inventar histórias e transformar ideias em pequenas encenações — muitas vezes sem perceber que esse impulso criativo acabaria se tornando parte fundamental da minha profissão no futuro.",
    inspiracao:
      "Desde cedo também desenvolvi o hábito de buscar inspiração em livros, filmes, pinturas e diferentes formas de arte. Gosto de observar como a imaginação humana se manifesta através das imagens, das histórias e das simbologias que atravessam diferentes culturas. Essa curiosidade constante sobre o imaginário e sobre as formas de expressão artística acabou moldando profundamente o meu olhar criativo.",
    manaus:
      "Nasci em Manaus, cercada por natureza, e esse ambiente sempre teve uma influência muito forte na forma como eu enxergo o mundo. A presença da floresta, dos rios e da vida natural despertou em mim uma sensibilidade especial para observar detalhes, atmosferas e símbolos presentes na natureza. Ao mesmo tempo, sempre me senti atraída por temas que exploram o imaginário, o simbólico e o espiritual. O universo das mitologias, do místico e das narrativas arquetípicas sempre despertou minha curiosidade.",
    fecho:
      "Fora das telas me considero uma pessoa bastante reservada, mas que valoriza muito as conexões profundas e significativas. Gosto de ocupar meu tempo com atividades que me mantenham ativa tanto na mente quanto no corpo. Sou apaixonada por dança mas atualmente tenho me dedicado ao Crossfit o que me ajuda a manter o equilíbrio físico e mental. Gosto de assistir animes no meu tempo livre e atualmente junto com meu parceiro estou entrando aos poucos no mundo dos games, o que tem sido uma experiência divertida e desafiadora. Acredito que essas atividades complementam minha prática artística, pois me permitem explorar diferentes formas de expressão e conexão com o mundo ao meu redor.",
  },
  en: {
    inicio:
      "My name is Isabel Pontes, I'm a fine art photographer and visual creator based in Curitiba. My relationship with art began long before photography officially entered my life. Since childhood I've been driven by imagination, movement and creation. I was part of a dance company and always had a natural curiosity about artistic expression. I loved building sets, inventing stories and turning ideas into little scenes — often without realizing this creative impulse would become a fundamental part of my profession in the future.",
    inspiracao:
      "Early on I also developed the habit of seeking inspiration in books, films, paintings and different art forms. I like observing how human imagination expresses itself through images, stories and the symbolism that runs across different cultures. This constant curiosity about the imaginary and about forms of artistic expression deeply shaped my creative eye.",
    manaus:
      "I was born in Manaus, surrounded by nature, and that environment always had a strong influence on how I see the world. The presence of the forest, the rivers and natural life awakened in me a special sensitivity for observing details, atmospheres and symbols in nature. At the same time, I've always been drawn to themes that explore the imaginary, the symbolic and the spiritual. The world of mythologies, the mystical and archetypal narratives has always sparked my curiosity.",
    fecho:
      "Off screen I consider myself a fairly reserved person, but one who deeply values profound and meaningful connections. I like filling my time with activities that keep me active in both mind and body. I'm passionate about dance, but lately I've been dedicating myself to CrossFit, which helps me keep physical and mental balance. I enjoy watching anime in my free time, and lately, together with my partner, I've been slowly getting into the world of games, which has been a fun and challenging experience. I believe these activities complement my artistic practice, as they let me explore different forms of expression and connection with the world around me.",
  },
};

export function metadadosSobre(idioma: Idioma) {
  return metadados({
    idioma,
    caminhoPT: "/sobre",
    titulo: idioma === "en" ? "About — Isabel Pontes" : "Sobre — Isabel Pontes",
    descricao:
      idioma === "en"
        ? "Fine art photographer and visual creator in Curitiba. Get to know Isabel Pontes's journey."
        : "Fotógrafa artística e criadora visual em Curitiba. Conheça a trajetória de Isabel Pontes.",
  });
}

export function Sobre({ idioma }: { idioma: Idioma }) {
  const en = idioma === "en";
  const texto = TEXTO[idioma];
  const primeira = imagemFixa("sobre-bio-1");
  const segunda = imagemFixa("sobre-bio-2");
  return (
    <Casca>
      <div className="sobre">
        <div className="sobre__bloco">
          <Picture
            base={primeira.base}
            w={primeira.w}
            h={primeira.h}
            alt={`Isabel Pontes, ${en ? "photographer" : "fotógrafa"}`}
            className="sobre__float sobre__float--esq"
          />
          <p>{texto.inicio}</p>
        </div>
        <div className="sobre__bloco">
          <Picture
            base={segunda.base}
            w={segunda.w}
            h={segunda.h}
            alt={`${en ? "Photo shoot by" : "Ensaio de"} Isabel Pontes`}
            className="sobre__float sobre__float--dir"
          />
          <p>{texto.inspiracao}</p>
          <p className="sobre__so-desktop">{texto.manaus}</p>
        </div>
        <div className="carrossel-wrap">
          <button className="carrossel-seta carrossel-seta--prev" data-carrossel-prev aria-label={en ? "Previous" : "Anterior"}>
            ‹
          </button>
          <div className="carrossel" data-carrossel>
            {carrossel().map((imagem) => (
              <div className="carrossel__item" key={imagem.id}>
                <Picture base={imagem.base} w={imagem.w} h={imagem.h} alt={en ? "Photography by Isabel Pontes" : "Fotografia de Isabel Pontes"} className="carrossel__img" />
              </div>
            ))}
          </div>
          <button className="carrossel-seta carrossel-seta--next" data-carrossel-next aria-label={en ? "Next" : "Próxima"}>
            ›
          </button>
        </div>
        <p className="sobre__fecho">{texto.fecho}</p>
      </div>
    </Casca>
  );
}
