import { Casca } from "@/components/Casca";
import { FormBriefing, type RetornoSemJS } from "@/components/FormBriefing";
import { Picture } from "@/components/Picture";
import { Portfolio } from "@/components/midiakit/Portfolio";
import { asset } from "@/lib/asset";
import { EMAIL, INSTAGRAM, WHATSAPP } from "@/lib/contato";
import { formularioPublico } from "@/lib/forms";
import { CATEGORIAS_EN, resolver, t, type Idioma } from "@/lib/i18n";
import { metadados } from "@/lib/metadados";
import { INVESTIMENTOS, METRICAS, PROCESSO, TAGS, TEMPO_MEDIO, VIDEOS, type Fatia } from "@/lib/midiakit/conteudo";
import { svg } from "@/lib/svg";
import { codificar } from "@/lib/url";

function whatsapp(mensagem: string): string {
  return `https://wa.me/${WHATSAPP}?text=${codificar(mensagem)}`;
}

function donut(fatias: Fatia[]): string {
  let acumulado = 0;
  const partes = fatias.map((fatia) => {
    const inicio = acumulado;
    acumulado += fatia.pct;
    return `${fatia.cor} ${inicio.toFixed(1)}% ${acumulado.toFixed(1)}%`;
  });
  return `conic-gradient(${partes.join(", ")})`;
}

export function metadadosMidiakit(idioma: Idioma) {
  return metadados({
    idioma,
    caminhoPT: "/midiakit",
    titulo: idioma === "en" ? "Media Kit — Isabel Pontes | UGC & Creator" : "Mídia Kit — Isabel Pontes | UGC & Creator",
    descricao:
      idioma === "en"
        ? "Authentic content with elevated aesthetics: UGC and IGC for brands, by Isabel Pontes, creator and creative director in Curitiba."
        : "Conteúdo autêntico com estética elevada: UGC e IGC para marcas, por Isabel Pontes, criadora e diretora criativa em Curitiba.",
  });
}

function Legenda({ idioma, fatias }: { idioma: Idioma; fatias: Fatia[] }) {
  return (
    <div className="mk-donut-linha">
      <span className="mk-donut" style={{ background: donut(fatias) }}></span>
      <ul className="mk-legenda">
        {fatias.map((fatia) => (
          <li key={fatia.rotulo[0]}>
            <span className="mk-ponto" style={{ background: fatia.cor }}></span>
            <strong>{`${fatia.pct}%`}</strong>
            {` ${resolver(idioma, fatia.rotulo)}`}
          </li>
        ))}
      </ul>
    </div>
  );
}

export type BuscaMidiakit = Promise<{ briefing?: string | string[] }>;

export async function Midiakit({ idioma, busca }: { idioma: Idioma; busca: BuscaMidiakit }) {
  const en = idioma === "en";
  const formUGC = formularioPublico("ugc");
  const { briefing } = await busca;
  const grupos = VIDEOS.map((categoria) => ({
    slug: categoria.slug,
    nome: (en && CATEGORIAS_EN[categoria.slug]) || categoria.nome,
    conteudo: (
      <div className="mk-vgrid">
        {categoria.videos.map((video) => (
          <figure className="mk-video" key={video.src}>
            <div className="mk-video__frame">
              <button type="button" className="mk-video__mock" data-video-src={video.src} aria-label={`${t(idioma, "mk.assistir")}: ${video.titulo}`}>
                <Picture base={video.poster} w={400} h={711} alt={`${t(idioma, "mk.previa")} ${video.titulo}`} className="mk-video__poster" />
                <span className="mk-video__play" aria-hidden="true"></span>
              </button>
            </div>
            <figcaption className="mk-video__titulo">{video.titulo}</figcaption>
          </figure>
        ))}
      </div>
    ),
  }));

  return (
    <Casca mostrarPdf ocultarRodape>
      <div className="midiakit">
        <section className="mk-hero">
          <video
            className="mk-hero__video"
            autoPlay
            muted
            loop
            playsInline
            preload="metadata"
            poster={asset("/static/img/midiakit/hero-poster.webp")}
            aria-hidden="true"
            tabIndex={-1}
          >
            <source src={asset("/static/video/midiakit-hero.webm")} type="video/webm" />
            <source src={asset("/static/video/midiakit-hero.mp4")} type="video/mp4" />
          </video>
          <div className="mk-hero__overlay"></div>
          <div className="mk-hero__texto">
            <h1>Isa Pontes</h1>
            <p className="mk-hero__sub">UGC Creator &amp; Diretora Criativa</p>
            <p className="mk-hero__tag">
              {en
                ? "Organic advertising with elevated aesthetics — storytelling that connects brand and audience."
                : "Publicidade orgânica com estética elevada — storytelling que conecta marca e audiência."}
            </p>
            <div className="mk-hero__ctas">
              <a href="#portfolio" className="mk-btn mk-btn--claro">
                {t(idioma, "mk.verTrabalhos")}
              </a>{" "}
              <a href="#investimento" className="mk-btn mk-btn--cor-claro">
                {t(idioma, "mk.investimento")}
              </a>
            </div>
          </div>
        </section>

        <section id="bio" className="mk-secao mk-bio">
          <h2>
            {en ? "Hi! I'm Isa " : "Oi! Eu sou a Isa "}
            <span aria-hidden="true">✨</span>
          </h2>
          <p>
            {en
              ? "I'm a Creative Director, Fine Art Photographer and Creator. I work as an Influencer and also create content for brands."
              : "Sou Diretora Criativa, Fotógrafa Artística e Creator. Atuo como Influencer e também crio conteúdo para marcas."}
          </p>
          {en ? (
            <p>
              {"My strength is "}
              <strong>authenticity and creativity</strong>
              {". I bring my background in cinema, art direction and visual arts to turn products into stories with aesthetics and intention. You take one look and see there's storytelling — it's not just showing the product, it's building a narrative that connects."}
            </p>
          ) : (
            <p>
              {"Meu ponto forte é "}
              <strong>autenticidade e criatividade</strong>
              {". Trago meu repertório de cinema, direção de arte e artes visuais para transformar produtos em histórias com estética e intenção. Você bate o olho e percebe que tem storytelling — não é só mostrar o produto, é construir uma narrativa que conecta."}
            </p>
          )}
          <ul className="mk-tags">
            {TAGS.map((tag) => (
              <li key={tag[0]}>{resolver(idioma, tag)}</li>
            ))}
          </ul>
        </section>

        <section id="ugc" className="mk-secao mk-dic">
          <article className="mk-verbete">
            <h3>UGC</h3>
            <span className="mk-verbete__fon">{`/juː dʒiː siː/ • ${en ? "noun" : "substantivo"}`}</span>
            {en ? (
              <>
                <p>
                  <em>User Generated Content</em>
                  {". Authentic material created to spark immediate identification. In my method, UGC stops being a homemade clip and becomes a piece of "}
                  <strong>organic advertising with elevated aesthetics</strong>
                  {"."}
                </p>
                <p className="mk-verbete__por">
                  <strong>Why hire it?</strong>
                  {" Humanization, objection-breaking and conversion."}
                </p>
              </>
            ) : (
              <>
                <p>
                  <em>User Generated Content</em>
                  {" (Conteúdo Gerado pelo Usuário). Material autêntico criado para gerar identificação imediata. No meu método, o UGC deixa de ser um vídeo caseiro e vira uma peça de "}
                  <strong>publicidade orgânica com estética elevada</strong>
                  {"."}
                </p>
                <p className="mk-verbete__por">
                  <strong>Por que contratar?</strong>
                  {" Humanização, quebra de objeção e conversão."}
                </p>
              </>
            )}
          </article>
          <article className="mk-verbete">
            <h3>IGC</h3>
            <span className="mk-verbete__fon">{`/aɪ dʒiː siː/ • ${en ? "noun" : "substantivo"}`}</span>
            {en ? (
              <>
                <p>
                  <em>Influencer Generated Content</em>
                  {". Content by someone who already has presence and connection with an audience. In my method, IGC combines "}
                  <strong>authority, narrative and elevated aesthetics</strong>
                  {" — designed to build desire and trust, without feeling like an ad."}
                </p>
                <p className="mk-verbete__por">
                  <strong>Why hire it?</strong>
                  {" Reach, credibility and conversion."}
                </p>
              </>
            ) : (
              <>
                <p>
                  <em>Influencer Generated Content</em>
                  {" (Conteúdo Gerado por Influenciadora). Conteúdo de quem já tem presença e conexão com uma audiência. No meu método, o IGC une "}
                  <strong>autoridade, narrativa e estética elevada</strong>
                  {" — pensado para gerar desejo e confiança, sem cara de anúncio."}
                </p>
                <p className="mk-verbete__por">
                  <strong>Por que contratar?</strong>
                  {" Alcance, credibilidade e conversão."}
                </p>
              </>
            )}
          </article>
        </section>

        <section id="portfolio" className="mk-secao mk-portfolio">
          <h2>{t(idioma, "mk.trabalhos")}</h2>
          <p>{t(idioma, "mk.curadoria")}</p>
          <Portfolio grupos={grupos} rotuloTodos={t(idioma, "comum.todos")} rotuloNav={t(idioma, "mk.filtrarCat")} />
          <a href="https://www.instagram.com/isabelpontesfotografia/" className="mk-btn" target="_blank" rel="noopener">
            {t(idioma, "mk.verInstagram")}
          </a>
        </section>

        <section id="metricas" className="mk-secao mk-metricas">
          <h2>{t(idioma, "mk.metricas")}</h2>
          <div className="mk-metricas__grid">
            {METRICAS.map((rede) => (
              <article className={`mk-rede mk-rede--${rede.lado}`} key={rede.rede}>
                <img
                  className="mk-rede__phone"
                  src={asset(rede.foto)}
                  alt={en ? `${rede.rede} profile metrics — Isabel Pontes` : `Prints das métricas de ${rede.rede} da Isabel Pontes`}
                  width={398}
                  height={799}
                  loading="lazy"
                  decoding="async"
                />
                <div className="mk-rede__conteudo">
                  <h3>{rede.rede}</h3>
                  <div className="mk-kpis">
                    <div className="mk-kpi">
                      <strong>{rede.seguidores}</strong>
                      <span>{t(idioma, "mk.seguidores")}</span>
                    </div>
                    <div className="mk-kpi">
                      <strong>{rede.views}</strong>
                      <span>{t(idioma, "mk.visualizacoes")}</span>
                    </div>
                  </div>
                  <h4>{t(idioma, "mk.publico")}</h4>
                  <div className="mk-barras">
                    {rede.locais.map((barra) => (
                      <div className="mk-barra" key={barra.rotulo[0]}>
                        <div className="mk-barra__topo">
                          <span>{resolver(idioma, barra.rotulo)}</span>
                          <strong>{`${barra.pct}%`}</strong>
                        </div>
                        <div className="mk-barra__trilha">
                          <span className="mk-barra__preenche" style={{ width: `${barra.pct}%` }}></span>
                        </div>
                      </div>
                    ))}
                  </div>
                  <div className="mk-donuts">
                    <Legenda idioma={idioma} fatias={rede.genero} />
                    <Legenda idioma={idioma} fatias={rede.idades} />
                  </div>
                </div>
              </article>
            ))}
          </div>
        </section>

        <section id="processo" className="mk-secao mk-processo">
          <h2>{t(idioma, "mk.comoFunciona")}</h2>
          <p className="mk-processo__pill">{resolver(idioma, TEMPO_MEDIO)}</p>
          <ol className="mk-passos">
            {PROCESSO.map((passo) => (
              <li className="mk-passo" key={passo.num}>
                <span className="mk-passo__num" aria-hidden="true">
                  {passo.num}
                </span>
                <div className="mk-passo__corpo">
                  <h3>{resolver(idioma, passo.titulo)}</h3>
                  <p>{resolver(idioma, passo.resumo)}</p>
                  <details className="mk-passo__mais">
                    <summary>
                      {`${t(idioma, "mk.entenda")} `}
                      <span className="mk-passo__prazo">{resolver(idioma, passo.prazo)}</span>
                    </summary>
                    <ul>
                      {passo.detalhes.map((detalhe) => (
                        <li key={detalhe[0]}>{resolver(idioma, detalhe)}</li>
                      ))}
                    </ul>
                  </details>
                </div>
              </li>
            ))}
          </ol>
        </section>

        {INVESTIMENTOS.map((bloco) => {
          const subtitulo = bloco.subtitulo ? resolver(idioma, bloco.subtitulo) : "";
          return (
            <section id={bloco.tag === "UGC" ? "investimento" : undefined} className="mk-secao mk-invest" key={bloco.tag}>
              <h2>{resolver(idioma, bloco.titulo)}</h2>
              {subtitulo && <p className="mk-invest__sub">{subtitulo}</p>}
              <div className="mk-pacotes">
                {bloco.pacotes.map((pacote) => {
                  const nome = resolver(idioma, pacote.nome);
                  const resumo = resolver(idioma, pacote.resumo);
                  const mensagem = t(idioma, "mk.wppPacote").replace("%s", nome).replace("%s", bloco.tag).replace("%s", resumo);
                  return (
                    <article className={pacote.destaque ? "mk-pacote mk-pacote--destaque" : "mk-pacote"} key={pacote.nome[0]}>
                      {pacote.destaque && <span className="mk-pacote__selo">{t(idioma, "mk.maisPopular")}</span>}
                      <h3>{nome}</h3>
                      <p className="mk-pacote__resumo">{resumo}</p>
                      <ul className="mk-pacote__itens">
                        {pacote.itens.map((item) => (
                          <li key={item[0]}>{resolver(idioma, item)}</li>
                        ))}
                      </ul>
                      <p className="mk-pacote__preco">{pacote.preco}</p>
                      <a className="mk-btn mk-pacote__cta" href={whatsapp(mensagem)} target="_blank" rel="noopener">
                        {t(idioma, "mk.solicitar")}
                      </a>
                    </article>
                  );
                })}
              </div>
            </section>
          );
        })}

        {formUGC && (
          <section className="mk-briefing no-print" id="briefing" aria-labelledby="mk-briefing-titulo">
            <div className="mk-briefing__texto">
              <h2 id="mk-briefing-titulo">{resolver(idioma, formUGC.titulo)}</h2>
              <p>{resolver(idioma, formUGC.subtitulo)}</p>
            </div>
            <FormBriefing idioma={idioma} form={formUGC} retorno={briefing as RetornoSemJS} />
          </section>
        )}

        <footer className="mk-contato">
          <h2>{t(idioma, "mk.contatoTitulo")}</h2>
          <p>{t(idioma, "mk.contatoSub")}</p>
          <div className="mk-contato__redes">
            <a href={whatsapp(t(idioma, "mk.wppGeral"))} target="_blank" rel="noopener" aria-label="WhatsApp" dangerouslySetInnerHTML={{ __html: svg("img/social/whatsapp.svg") }} />{" "}
            <a href={`mailto:${EMAIL}`} aria-label="E-mail" dangerouslySetInnerHTML={{ __html: svg("img/social/email.svg") }} />{" "}
            <a href={INSTAGRAM} target="_blank" rel="noopener" aria-label="Instagram" dangerouslySetInnerHTML={{ __html: svg("img/social/instagram.svg") }} />{" "}
            <a href="https://www.tiktok.com/@isabelpontesfoto" target="_blank" rel="noopener" aria-label="TikTok" dangerouslySetInnerHTML={{ __html: svg("img/social/tiktok.svg") }} />
          </div>
          <p className="mk-contato__print">{`${EMAIL} · instagram.com/isapontesfoto · wa.me/${WHATSAPP}`}</p>
          <p className="mk-contato__copy">{`© ${new Date().getFullYear()} Isa Pontes`}</p>
        </footer>
      </div>
    </Casca>
  );
}
