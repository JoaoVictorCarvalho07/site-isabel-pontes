import { Casca } from "@/components/Casca";
import { EMAIL } from "@/lib/contato";
import type { Idioma } from "@/lib/i18n";
import { metadados } from "@/lib/metadados";

type Secao = { titulo: string; paragrafos?: string[]; itens?: string[] };

const CONTEUDO: Record<Idioma, { titulo: string; atualizacao: string; abertura: string; secoes: Secao[]; contato: string }> = {
  pt: {
    titulo: "Política de privacidade",
    atualizacao: "Última atualização: 12 de outubro de 2026",
    abertura:
      "Este site é o portfólio e o mídia kit de Isabel Pontes, fotógrafa em Curitiba. Esta página explica quais dados pessoais o site recebe, para que servem e como pedir a exclusão deles.",
    secoes: [
      {
        titulo: "O que o site coleta",
        paragrafos: ["Navegar pelo site não exige cadastro. Dados pessoais só são recebidos quando você preenche um formulário de briefing:"],
        itens: [
          "o que você escrever nos campos, como nome, WhatsApp ou e-mail, e os detalhes do ensaio ou do projeto;",
          "as imagens de referência que você anexar;",
          "o endereço IP e a identificação do navegador no momento do envio, usados só para barrar envios automáticos.",
        ],
      },
      {
        titulo: "Para que servem",
        paragrafos: [
          "As respostas são usadas para entender o seu pedido, preparar uma proposta e entrar em contato com você. Elas não são vendidas, não são usadas para publicidade e não são compartilhadas com terceiros para fins próprios deles.",
        ],
      },
      {
        titulo: "Onde ficam",
        itens: [
          "As respostas e os anexos ficam em um servidor contratado pelo site, com cópias de segurança armazenadas na Cloudflare.",
          "Quando um briefing é enviado, um aviso com as respostas pode ser encaminhado por e-mail a Isabel pelo serviço Resend.",
          "As fotos da galeria e os vídeos do mídia kit são entregues pela Cloudflare, que recebe o endereço IP de quem os carrega, como qualquer serviço de hospedagem.",
        ],
      },
      {
        titulo: "Cookies e rastreamento",
        paragrafos: [
          "O site público não usa cookies, não tem ferramentas de análise de audiência e não carrega scripts de redes sociais ou de anúncios. A escolha entre tema claro e escuro fica guardada apenas no seu navegador.",
          "A área administrativa, restrita à equipe do site, usa um cookie de sessão e o login do Google para identificar quem está autorizado. Do Google são recebidos apenas o nome e o e-mail da conta.",
        ],
      },
      {
        titulo: "Por quanto tempo",
        paragrafos: [
          "Os briefings são mantidos enquanto forem úteis para o atendimento e para o histórico do trabalho realizado. Você pode pedir a exclusão a qualquer momento.",
        ],
      },
      {
        titulo: "Seus direitos",
        paragrafos: [
          "A Lei Geral de Proteção de Dados garante a você confirmar se há dados seus aqui, acessá-los, corrigi-los e pedir que sejam apagados. Para qualquer um desses pedidos, escreva para o e-mail abaixo.",
        ],
      },
    ],
    contato: "Responsável pelos dados: Isabel Pontes. Contato: ",
  },
  en: {
    titulo: "Privacy policy",
    atualizacao: "Last updated: October 12, 2026",
    abertura:
      "This website is the portfolio and media kit of Isabel Pontes, a photographer based in Curitiba, Brazil. This page explains which personal data the site receives, what it is used for and how to ask for its deletion.",
    secoes: [
      {
        titulo: "What the site collects",
        paragrafos: ["Browsing the site requires no account. Personal data is only received when you fill out a briefing form:"],
        itens: [
          "what you write in the fields, such as your name, WhatsApp or email, and the details of the shoot or project;",
          "the reference images you attach;",
          "the IP address and browser identification at the time of submission, used only to block automated submissions.",
        ],
      },
      {
        titulo: "What it is used for",
        paragrafos: [
          "Your answers are used to understand your request, prepare a proposal and get in touch with you. They are not sold, not used for advertising and not shared with third parties for their own purposes.",
        ],
      },
      {
        titulo: "Where it is stored",
        itens: [
          "Answers and attachments are kept on a server hired by the site, with backups stored at Cloudflare.",
          "When a briefing is submitted, a notice with the answers may be forwarded by email to Isabel through the Resend service.",
          "Gallery photos and media kit videos are delivered by Cloudflare, which receives the IP address of whoever loads them, as any hosting service does.",
        ],
      },
      {
        titulo: "Cookies and tracking",
        paragrafos: [
          "The public site uses no cookies, has no audience analytics and loads no social media or advertising scripts. Your choice between light and dark theme is stored only in your browser.",
          "The administrative area, restricted to the site's team, uses a session cookie and Google sign-in to identify who is authorized. Only the account's name and email are received from Google.",
        ],
      },
      {
        titulo: "For how long",
        paragrafos: ["Briefings are kept while they are useful for the service and for the record of the work done. You can ask for deletion at any time."],
      },
      {
        titulo: "Your rights",
        paragrafos: [
          "Brazil's General Data Protection Law (LGPD) entitles you to confirm whether your data is held here, access it, correct it and ask for it to be erased. For any of these requests, write to the email below.",
        ],
      },
    ],
    contato: "Data controller: Isabel Pontes. Contact: ",
  },
};

export function metadadosPrivacidade(idioma: Idioma) {
  return metadados({
    idioma,
    caminhoPT: "/privacidade",
    titulo: `${CONTEUDO[idioma].titulo} — Isabel Pontes`,
    descricao:
      idioma === "en"
        ? "Which personal data this website receives, what it is used for and how to ask for its deletion."
        : "Quais dados pessoais este site recebe, para que servem e como pedir a exclusão deles.",
  });
}

export function Privacidade({ idioma }: { idioma: Idioma }) {
  const conteudo = CONTEUDO[idioma];
  return (
    <Casca>
      <article className="legal">
        <h1>{conteudo.titulo}</h1>
        <p className="legal__data">{conteudo.atualizacao}</p>
        <p>{conteudo.abertura}</p>
        {conteudo.secoes.map((secao) => (
          <section key={secao.titulo}>
            <h2>{secao.titulo}</h2>
            {secao.paragrafos?.map((paragrafo) => <p key={paragrafo}>{paragrafo}</p>)}
            {secao.itens && (
              <ul>
                {secao.itens.map((item) => (
                  <li key={item}>{item}</li>
                ))}
              </ul>
            )}
          </section>
        ))}
        <p>
          {conteudo.contato}
          <a href={`mailto:${EMAIL}`}>{EMAIL}</a>
          {"."}
        </p>
      </article>
    </Casca>
  );
}
