import { EMAIL, INSTAGRAM, WHATSAPP } from "@/lib/contato";
import { LOGO } from "@/lib/icones";
import { prefixo, t, type Idioma } from "@/lib/i18n";
import { svg } from "@/lib/svg";
import { codificar } from "@/lib/url";

const TIKTOK = "https://www.tiktok.com/@isabelpontesfoto";

export function Rodape({ idioma, colado }: { idioma: Idioma; colado?: boolean }) {
  const en = idioma === "en";
  const p = prefixo(idioma);
  const whatsapp = `https://wa.me/${WHATSAPP}?text=${codificar(en ? "Hi! I came from your website ✨" : "Oi! Vim pelo seu site ✨")}`;
  return (
    <footer className={colado ? "rodape rodape--colado" : "rodape"}>
      <div className="rodape__corpo">
        <div>
          <a href={en ? "/en" : "/"} className="rodape__marca" aria-label="Isabel Pontes" dangerouslySetInnerHTML={{ __html: LOGO }} />
          <p className="rodape__frase">
            {en
              ? "Fine art photography, conceptual shoots and creative direction in Curitiba."
              : "Fotografia artística, ensaios conceituais e direção criativa em Curitiba."}
          </p>
        </div>
        <nav className="rodape__links" aria-label={en ? "Footer" : "Rodapé"}>
          <a href={`${p}/galeria`}>{t(idioma, "nav.galeria")}</a>
          <a href={`${p}/sobre`}>{t(idioma, "nav.sobre")}</a>
          <a href={`${p}/midiakit`}>{t(idioma, "nav.midiakit")}</a>
          <a href={`${p}/contato`}>{t(idioma, "nav.contato")}</a>
          <a href={`${p}/briefing`}>Briefing</a>
        </nav>
        <div className="rodape__redes">
          <a href={INSTAGRAM} target="_blank" rel="noopener" aria-label="Instagram" dangerouslySetInnerHTML={{ __html: svg("img/social/instagram.svg") }} />
          <a href={TIKTOK} target="_blank" rel="noopener" aria-label="TikTok" dangerouslySetInnerHTML={{ __html: svg("img/social/tiktok.svg") }} />
          <a href={whatsapp} target="_blank" rel="noopener" aria-label="WhatsApp" dangerouslySetInnerHTML={{ __html: svg("img/social/whatsapp.svg") }} />
          <a href={`mailto:${EMAIL}`} aria-label={en ? "Email" : "E-mail"} dangerouslySetInnerHTML={{ __html: svg("img/social/email.svg") }} />
        </div>
      </div>
      <div className="rodape__base">
        <p>{`© ${new Date().getFullYear()} Isabel Pontes · ${t(idioma, "rodape")}`}</p>
        <a href={`${p}/privacidade`}>{en ? "Privacy policy" : "Política de privacidade"}</a>
      </div>
    </footer>
  );
}
