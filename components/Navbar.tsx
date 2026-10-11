import { ICONE_PDF, LOGO } from "@/lib/icones";
import { caminho, prefixo, t, type Idioma } from "@/lib/i18n";
import { BotaoTema } from "./BotaoTema";

type Props = { idioma: Idioma; caminhoPT: string; mostrarPdf?: boolean };

function Links({ idioma }: { idioma: Idioma }) {
  const p = prefixo(idioma);
  return (
    <>
      <a href={`${p}/galeria`}>{t(idioma, "nav.galeria")}</a>
      <a href={`${p}/sobre`}>{t(idioma, "nav.sobre")}</a>
      <a href={`${p}/midiakit`}>{t(idioma, "nav.midiakit")}</a>
      <a href={`${p}/contato`}>{t(idioma, "nav.contato")}</a>
    </>
  );
}

export function Navbar({ idioma, caminhoPT, mostrarPdf }: Props) {
  const en = idioma === "en";
  const outro = caminho(en ? "pt" : "en", caminhoPT);
  const hrefLang = en ? "pt-BR" : "en";
  return (
    <>
      <header className="navbar" data-navbar>
        <a
          href={en ? "/en" : "/"}
          className="navbar__logo logo-anim-focus"
          aria-label="Isabel Pontes"
          dangerouslySetInnerHTML={{ __html: LOGO }}
        />
        <nav className="navbar__links" aria-label={t(idioma, "a11y.principal")}>
          <Links idioma={idioma} />
          {mostrarPdf && (
            <button
              className="navbar__pdf"
              data-imprimir
              type="button"
              aria-label={t(idioma, "mk.baixarPdf")}
              dangerouslySetInnerHTML={{ __html: `${ICONE_PDF}\n      PDF\n    ` }}
            />
          )}
          <a className="navbar__lang" href={outro} hrefLang={hrefLang} aria-label={en ? "Ver em português" : "View in English"}>
            {en ? "PT" : "EN"}
          </a>
          <BotaoTema idioma={idioma} />
        </nav>
        <button className="navbar__hamburguer" data-menu-abrir aria-label={t(idioma, "a11y.abrir")} aria-haspopup="dialog">
          <span></span>
          <span></span>
          <span></span>
        </button>
      </header>
      <dialog id="menu" className="menu" aria-label={t(idioma, "a11y.menu")}>
        <button className="menu__fechar" data-menu-fechar aria-label={t(idioma, "a11y.fechar")}>
          ✕
        </button>
        <nav className="menu__links" aria-label={t(idioma, "a11y.principal")}>
          <Links idioma={idioma} />
          {mostrarPdf && (
            <button className="menu__pdf" data-imprimir type="button">
              {t(idioma, "mk.baixarPdf")}
            </button>
          )}
          <div className="menu__acoes">
            <a className="menu__lang" href={outro} hrefLang={hrefLang}>
              {en ? "Português" : "English"}
            </a>
            <BotaoTema idioma={idioma} />
          </div>
        </nav>
      </dialog>
    </>
  );
}
