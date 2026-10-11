import { Casca } from "@/components/Casca";
import { CIDADE, EMAIL, INSTAGRAM, WHATSAPP } from "@/lib/contato";
import { metadados } from "@/lib/metadados";
import { prefixo, t, type Idioma } from "@/lib/i18n";
import { codificar } from "@/lib/url";

const MENSAGEM = {
  pt: {
    whatsapp: "Oi! Vim pelo seu site ✨\n\nQuero falar sobre: (ensaio / evento / UGC)\nData/cidade: \nReferências (se tiver): \n",
    assunto: "Contato pelo site",
    corpo: "Oi! Vim pelo seu site ✨\n\nQuero falar sobre:\n",
  },
  en: {
    whatsapp: "Hi! I came from your website ✨\n\nI'd like to talk about: (photoshoot / event / UGC)\nDate/city: \nReferences (if any): \n",
    assunto: "Contact from the website",
    corpo: "Hi! I came from your website ✨\n\nI'd like to talk about:\n",
  },
};

export function metadadosContato(idioma: Idioma) {
  return metadados({
    idioma,
    caminhoPT: "/contato",
    titulo: idioma === "en" ? "Contact — Isabel Pontes" : "Contato — Isabel Pontes",
    descricao:
      idioma === "en"
        ? "Reach Isabel Pontes via WhatsApp, Instagram or email. Based in Curitiba."
        : "Fale com Isabel Pontes por WhatsApp, Instagram ou e-mail. Base em Curitiba.",
  });
}

type CardProps = { href: string; externo?: boolean; destaque?: boolean; titulo: string; texto: string; cta: string };

function Card({ href, externo, destaque, titulo, texto, cta }: CardProps) {
  return (
    <a
      className={destaque ? "contato-card contato-card--destaque" : "contato-card"}
      href={href}
      target={externo ? "_blank" : undefined}
      rel={externo ? "noopener noreferrer" : undefined}
    >
      <div className="contato-card__topo">
        <div>
          <h2>{titulo}</h2>
          <p>{texto}</p>
        </div>
        <span className="contato-card__seta" aria-hidden="true">
          →
        </span>
      </div>
      <span className="contato-card__cta">{cta}</span>
    </a>
  );
}

export function Contato({ idioma }: { idioma: Idioma }) {
  const en = idioma === "en";
  const mensagem = MENSAGEM[idioma];
  const whatsapp = `https://wa.me/${WHATSAPP}?text=${codificar(mensagem.whatsapp)}`;
  const mailto = `mailto:${EMAIL}?subject=${codificar(mensagem.assunto)}&body=${codificar(mensagem.corpo)}`;
  return (
    <Casca>
      <section className="contato">
        <header className="contato__head">
          <h1>{en ? "Contact" : "Contato"}</h1>
          <p>{en ? "Reach out however works best — I reply within 24 business hours. 🤍" : "Me chama por onde for melhor — eu respondo em até 24h úteis. 🤍"}</p>
        </header>
        <div className="contato__cards">
          <Card
            href={whatsapp}
            externo
            destaque
            titulo="WhatsApp"
            texto={en ? "Fastest for quotes and availability." : "Mais rápido para orçamento e disponibilidade."}
            cta={en ? "Message on WhatsApp" : "Chamar no WhatsApp"}
          />
          <Card
            href={INSTAGRAM}
            externo
            titulo="Instagram"
            texto={en ? "Chat, references and behind the scenes." : "Bate-papo, referências e bastidores."}
            cta={en ? "Open Instagram" : "Abrir Instagram"}
          />
          <Card
            href={mailto}
            titulo={en ? "Email" : "E-mail"}
            texto={en ? "For proposals, brands and formal requests." : "Para propostas, marcas e demandas formais."}
            cta={en ? "Send email" : "Enviar e-mail"}
          />
        </div>
        <div className="contato__info">
          <div className="contato-info-box">
            <h3>{en ? "How to reach me" : "Como falar comigo"}</h3>
            <ul>
              <li>{en ? "Tell me what you're looking for (shoot / event / UGC)." : "Me diga o que você procura (ensaio / evento / UGC)."}</li>
              <li>{en ? "If you have them, send theme + references (Pinterest/screenshots)." : "Se tiver, mande tema + referências (Pinterest/prints)."}</li>
              <li>{en ? "Date and city help me reply faster." : "Data e cidade ajudam a eu te responder mais rápido."}</li>
            </ul>
          </div>
          <div className="contato-info-box">
            <h3>{en ? "Availability" : "Atendimento"}</h3>
            <p>
              {en ? "Based in " : "Base em "}
              <strong>{CIDADE}</strong>
              {en ? ". Other locations on request." : ". Outros locais sob consulta."}
            </p>
            <div className="contato-tags">
              <span>{en ? "Artistic shoots" : "Ensaios artísticos"}</span>
              <span>{en ? "Art direction" : "Direção de arte"}</span>
              <span>{en ? "Content / UGC" : "Conteúdo / UGC"}</span>
            </div>
          </div>
        </div>
        <div className="contato__briefing">
          <h2>{en ? "Prefer to tell me everything at once?" : "Prefere já me contar tudo?"}</h2>
          <p>
            {en
              ? "Fill out the briefing and I'll come back with a proposal ready — including the snack I'll prepare for you 🤍"
              : "Preenche o briefing e eu já volto com uma proposta pronta — inclusive com o lanchinho que vou preparar pra você 🤍"}
          </p>
          <a className="contato__briefing-cta" href={`${prefixo(idioma)}/briefing`}>
            {t(idioma, "form.abrirBriefing")}
          </a>
        </div>
        <p className="contato__dica">
          {en
            ? "Tip: if you prefer, just say “I came from the website” and I'll guide you with the questions ✨"
            : "Dica: se preferir, pode mandar só “vim pelo site” que eu te guio com as perguntas ✨"}
        </p>
      </section>
    </Casca>
  );
}
