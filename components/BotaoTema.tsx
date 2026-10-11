import { ICONES_TEMA } from "@/lib/icones";
import { t, type Idioma } from "@/lib/i18n";

export function BotaoTema({ idioma }: { idioma: Idioma }) {
  return (
    <button
      className="botao-tema"
      data-tema-toggle
      type="button"
      aria-label={t(idioma, "a11y.tema")}
      dangerouslySetInnerHTML={{ __html: ICONES_TEMA }}
    />
  );
}
