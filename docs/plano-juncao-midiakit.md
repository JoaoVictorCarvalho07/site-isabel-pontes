# Junção Mídia Kit → site Go+HTMX — Decisões & Plano

> Documento da fase de decisão. Consolida a análise dos dois projetos da Isabel
> Pontes e o plano para fundir o **Mídia Kit** (UGC/IGC creator) dentro deste
> app (fotografia, Go+HTMX). Última atualização: 2026-07-20.
>
> **Nenhum código foi escrito ainda.** Este doc é o alvo a aprovar antes de codar.

---

## 1. Contexto

Existem hoje dois sites da mesma pessoa:

| | Fotografia | Mídia Kit |
|---|---|---|
| Repo | `site-da-bebel-htmx` (este) | `Midiakit_isabelPontes` |
| Stack | **Go + HTMX**, SQLite, R2, binário único | HTML/CSS/JS estático puro |
| Domínio | `isabelpontes.com.br` | `isabelpontesportfólio.com.br` (punycode `xn--isabelpontesportflio-r8b`) |
| Público | Clientes de ensaio/evento | Marcas (UGC/IGC) |
| Maturidade | Produção, testado, documentado | Protótipo |

**Achado decisivo:** a home deste site (`templates/pages/home.gohtml`) **já** trata o
mídia kit como um dos dois pilares do "universo da Isabel" — hoje só linka pra
fora (`card-ed` → domínio do mídia kit). A junção termina algo que o código já
começou.

**A identidade visual é praticamente a mesma marca** (ver §5): creme, café,
dourado, vinho, verde; serif editorial + sans. Não são dois produtos — são duas
metades de uma presença só.

---

## 2. Decisões consolidadas

### Macro (junção)
| Tema | Decisão |
|---|---|
| Arquitetura | **Fundir no Go+HTMX.** Mídia kit vira rotas neste binário. Plano Astro/Vercel do mídia kit fica **aposentado**. |
| Domínio | **Tudo em `isabelpontes.com.br`.** Mídia kit vira seção/rota (ex.: `/midiakit`). `isabelpontesportfólio.com.br` → **301** pro novo caminho. |
| Deploy | O que já existe: Oracle VM + systemd + Caddy + Litestream→R2 (fallback Fly.io). Um binário, um deploy. |

### Produto/conteúdo (herdadas da análise do mídia kit)
| Tema | Decisão |
|---|---|
| Escopo | Técnico + polir visual, com **redesign** das seções Métricas, Pacotes/Investimento, Portfólio+Galeria; **repensar** a seção Processo (hoje roxo/rosa destoante). |
| Identidade | **Evoluir/refinar** a marca existente (não repensar do zero). Consolidar num design system único (§5). Eu proponho o moodboard partindo da estética atual. |
| Edição de conteúdo | **Arquivos simples** (o padrão do projeto: struct Go / JSON / `manifest` + `sync` no SQLite). Sem CMS. |
| Conversão | **WhatsApp + dois formulários** de briefing (fotografia e UGC), backend em Go + SQLite + e-mail via Resend (ver §8). |
| Idioma | **PT-BR + Inglês.** Nenhum dos dois tem i18n hoje; será construído (§7). |
| Vídeo (hospedagem) | **Manter Cloudflare R2** por enquanto (custo zero). Reavaliar Stream/Mux/Bunny só se volume ou UX de player exigir (§6). |
| SEO / Analytics / A11y | **Alta prioridade.** A base de SEO/OG/JSON-LD deste projeto já cobre boa parte; estender pras rotas novas + analytics. |
| Seções novas | **Nenhuma.** "Marcas atendidas" já existe implícito no portfólio organizado por marca (L'Oréal, Natura, Dove, Mercado Livre...). |
| Prazo | Sem prazo rígido. Fasear por impacto. |

---

## 3. Arquitetura alvo

Um único app Go, mantendo todos os padrões atuais (`go:embed`, snapshot em RAM,
SQLite como verdade editável, ETag/cache, server-render pra SEO).

```
isabelpontes.com.br
├── /                     Home  (universo: fotografia + mídia kit)   [existe]
├── /sobre                Sobre                                       [existe]
├── /galeria              Galeria de fotos (shuffle diário)           [existe]
├── /galeria/fotos        Fragmento HTMX do filtro                    [existe]
├── /contato              Contato                                     [existe]
│
├── /midiakit             Mídia kit — ONE-PAGE de pitch (rolagem)     [novo]
│     └ seções por âncora: #portfolio #metricas #processo #investimento
├── /midiakit/videos      Fragmento HTMX (filtro de categoria)        [novo]
├── /midiakit.pdf         Download do mídia kit (PDF pré-renderizado) [novo]
│
├── /en/...               Espelho em inglês de tudo acima             [novo, §7]
├── /robots.txt /sitemap.xml /healthz                                 [existe]
```

**O mídia kit é UMA página, não um subsite** (ver §4b). Marca grande espera
receber um mídia kit como documento de pitch rolável — não como site com
navegação. As antigas "sub-rotas" viram âncoras (`#metricas` etc.) na mesma
página. O site de fotografia é que segue multi-página (navegação de verdade).

Rotas novas usam o mesmo `ServeMux` (Go 1.22+), o mesmo `Renderer`, o mesmo
`pageData`/ViewModel e os partials compartilhados. Zero dependência nova no
runtime — inclusive o PDF (ver §4b) é pré-gerado fora do servidor.

---

## 4. Mapeamento: seções do mídia kit → destino

Todas as seções vivem na **mesma página** `/midiakit` (rolagem); a coluna "Âncora"
é o `id` da seção pra deep-link e pro menu interno.

| Seção atual (mídia kit) | Âncora | Redesign? | Notas |
|---|---|---|---|
| Hero (vídeo de fundo) | topo | refinar | Trocar MP4 autoplay pesado por `<picture>` AVIF + `poster` (padrão que a home de fotografia já usa) |
| Bio "Oi! Eu sou a Isa" | `#bio` | refinar | Conteúdo estático em template |
| Dicionário UGC/IGC | `#ugc` | refinar | Estático |
| Portfólio de vídeos | `#portfolio` | **redesign** | Ver §6 — entra no padrão manifest+SQLite. Filtro via `/midiakit/videos` (HTMX) sem sair da página |
| Galeria fotografia (carrossel) | **remover daqui** | — | Já existe `/galeria`, muito superior. Elimina duplicação. |
| Métricas IG/TikTok | `#metricas` | **redesign** | Donuts hoje em `conic-gradient`; manter técnica, refazer layout. Dados em struct/JSON. |
| Processo (timeline roxa) | `#processo` | **repensar** | Descartar tema roxo/rosa; reconstruir na identidade da marca |
| Investimento UGC + Influencer | `#investimento` | **redesign** | Pacotes/preços em JSON editável à mão. CTA WhatsApp (padrão `urlEncode` já existe em `pages.go`) |
| Contato | link p/ `/contato` | — | Corrigir placeholders do mídia kit (`seuemail@exemplo.com`, `href="#"`) — o `/contato` do Go já tem os dados certos |

**Placeholders a corrigir na migração** (bugs vivos no mídia kit atual): e-mail
`seuemail@exemplo.com`, WhatsApp de contato `href="#"`, Instagram divergente
(`isapontesfoto` vs `isabelpontesfotografia`). Confirmar com a Isabel o número
E.164 correto — já há uma `NOTA` em `pages.go` sobre `554191977011` ter 12
dígitos (possível falta de 1).

---

## 4b. Formato: one-page de pitch + PDF (a página é o principal)

> **Diretriz:** a **página** `/midiakit` é o produto principal — bem construída,
> viva, interativa. O **PDF é necessário mas secundário**: não deve ditar
> arquitetura, virar dependência de runtime, nem consumir esforço desproporcional.
> Quando houver conflito entre "melhor página" e "melhor PDF", a página vence.

**Página (`/midiakit`) — one-pager híbrido:**
- Uma página, tudo por rolagem; seções como âncoras (§4). É o formato que marca
  grande espera receber.
- **Híbrido:** foco de pitch (nav enxuta — logo + "Baixar mídia kit"), com um link
  discreto no rodapé de volta pro "universo" (fotografia). Não carrega a navbar
  completa do site levando o decisor pra outros cantos.
- Interativa de verdade (vídeos, filtro HTMX, métricas animadas) — é o diferencial
  vs. um PDF morto.

**PDF (`/midiakit.pdf`) — espelho, pré-renderizado, baixa prioridade:**
- **Espelho fiel** da página, adaptado ao papel (frame estático/poster no lugar do
  vídeo autoplay; links clicáveis pros vídeos).
- **Pré-gerado por script** (Chrome headless *offline*, no mesmo espírito do
  `migrate_media.py`), guardado em `static/` ou R2. **O runtime Go não gera PDF** —
  nada de Chrome no container, o binário leve continua leve.
- Regenera quando o conteúdo do mídia kit muda (passo manual/CI, junto do `sync`).
- **É a última coisa a construir** (ver Fase 12). O site pode ir ao ar sem o PDF;
  ele entra depois sem bloquear nada. PT e EN.

---

## 5. Design system unificado

As duas paletas já são a mesma marca. Consolidar nos **tokens deste projeto**
(`assets/scss/_tokens.scss`), que já são HSL custom properties com dark mode.

| Papel | Mídia kit (hex) | Este projeto (HSL) | Ação |
|---|---|---|---|
| Fundo claro creme | `#f5f0e6` | `hsl(35 29% 92%)` | manter token atual |
| Base café escura | `#2e2420` | `hsl(22 26% 18%)` | manter |
| Dourado | `#d4af37` | `hsl(43 38% 63%)` | manter (`--gold`) |
| Vinho | `#58111a` | `hsl(346 50% 24%)` | manter (`--destructive`) |
| Verde | `#2f4f4f` | `hsl(166 30% 21%)` | manter (`--accent`) |
| Serif | Playfair Display | Cormorant Garamond | **manter Cormorant** (já self-hosted em `static/fonts`) |
| Sans | Montserrat | Jost | **manter Jost** |

**Regra:** o mídia kit adota os tokens/fontes deste projeto — não o contrário.
Nada de Google Fonts externo (o projeto já self-hospeda fontes; ver `_fonts.scss`).

**Grão de filme — decidido: portar como token opcional.** A textura de ruído
analógico do mídia kit (`body::before` com SVG `feTurbulence`, opacidade ~5%)
reforça a pegada "cinema/papel" da marca e vem pro site unificado. Implementar
como **token liga/desliga** (ex.: classe no `<body>` ou custom property), pra:
(a) poder ajustar intensidade num lugar só, e (b) **desativar sobre a galeria de
fotos** se o ruído sujar as imagens — testar no redesign.

**Formato de imagem — padrão AVIF, exceção WebP p/ transparência.** O padrão do
projeto é **AVIF** (gerado por `D:\projetos\scripts\to_avif.py`, ffmpeg
libaom-av1). Exceção: imagens que **precisam de canal alpha** (ex.: os mockups de
celular das métricas, que flutuam sobre o card com cantos e sombra) usam **WebP**,
porque o `to_avif.py` força `yuv420p` e não preserva transparência. WebP já é
cidadão de primeira no projeto (o `<picture>` serve AVIF + WebP). Se um dia
precisar de AVIF com alpha como padrão, trocar o encoder por `avifenc`/libavif.

---

## 6. Vídeos do portfólio UGC — modelo de dados

Hoje o mídia kit serve vídeos por `videos-index.json` + `labels.json` + URLs R2
(`videos.js`). Isso é **o mesmo problema** que a galeria de fotos deste projeto
já resolve melhor. Convergir:

- Estender o **schema** (`internal/store/schema.sql`) com uma tabela `videos`
  (ou generalizar `photos` → `media` com `kind`), espelhando o padrão: `key`,
  `category`, variantes, `poster`/`lqip`, `published`, `sort_order`, `label`.
- Estender `migrate_media.py` (ou script irmão) pra gerar o manifest dos vídeos:
  poster/thumbnail + o que a hospedagem escolhida exigir.
- `site sync` ingere; snapshot em RAM serve; HTMX filtra por categoria — igual à
  galeria de fotos.
- `labels.json`/`videos-index.json` atuais viram fonte do manifest inicial.

### Hospedagem de vídeo — **decidido: manter Cloudflare R2**

R2 continua servindo os vídeos por enquanto — **custo zero** e já integrado. O
trabalho da Fase 9 é só melhorar a *entrega*, não trocar de provedor:

- Gerar **poster/thumbnail** (AVIF/WebP) por vídeo no pipeline → sem tela preta
  no carregamento, sem CLS.
- `<video preload="metadata">` + `loading`/lazy fora da dobra; player enxuto
  (o `videos.js` atual tem lógica de mute/play duplicada e frágil — reescrever
  limpo no padrão do projeto).

Reavaliar migração pra **Cloudflare Stream / Bunny / Mux** só se aparecer
necessidade real (HLS adaptativo, volume grande, analytics de vídeo). Fica
registrado como caminho futuro, não como trabalho agora.

---

## 7. i18n (PT-BR + Inglês)

Nenhum dos dois tem hoje. Abordagem enxuta, sem framework:

- Prefixo de rota `/en/...` (PT é default sem prefixo). Handler resolve `lang`.
- Strings em mapa de tradução por chave (Go `map` ou arquivos JSON embutidos via
  `go:embed`), injetadas no ViewModel. Conteúdo longo (bio, pacotes) por locale
  no mesmo JSON de conteúdo.
- `<link rel="alternate" hreflang>` + `og:locale` por página (o layout já monta
  OG por rota — estender).
- `sitemap.xml` lista os dois idiomas.

Custo real: é mais manual que o i18n nativo do Astro, mas é o único item onde a
stack Go perde — e não é regressão (não existia).

---

## 8. Conversão — formulários de briefing

**Decidido.** Além do CTA de WhatsApp (mantém o `urlEncode`/`wa.me` já em
`internal/web/pages.go`), o site terá **formulários de briefing reais** — estilo
"Google Forms": a Isa define as perguntas, o visitante responde, e as respostas
ficam guardadas pra ela ler. Backend **no próprio Go** (o app já é um servidor;
não precisa de terceiro pro core).

### Dois formulários separados, um por frente

Sem misturar responsabilidades nem páginas:

| Form | Página | Público | Perguntas de exemplo (da Isa) |
|---|---|---|---|
| **Briefing Fotografia** | lado fotografia (`/contato` ou `/briefing`) | quem quer ensaio/evento | tipo (ensaio/evento), data/cidade, **motivação pra fazer o ensaio**, **o que gosta de comer** (ela faz lanche pra quem vai ao estúdio), referências |
| **Briefing UGC/Creator** | lado mídia kit (`/midiakit/briefing`) | marcas | marca, produto, objetivo (awareness/conversão), **por que o contato / por que escolheu a Isa como creator**, prazo, pacote de interesse |

As perguntas específicas dela (lanche, motivação, "por que me escolheu") são o
coração do pedido — o form não é genérico, é a personalidade da Isa. Cada form
vive só na sua página.

### Modelo técnico (casa com a filosofia do projeto)

- **Perguntas editáveis** em arquivo de conteúdo (JSON/struct por form, no padrão
  "arquivos simples" já adotado; futuramente migrável pro SQLite pra ela editar
  sem redeploy — ver Fase 2 abaixo). Cada pergunta: chave, label, tipo
  (texto/textarea/select/checkbox), obrigatória?, opções.
- **Handlers Go**: `POST /briefing/fotografia` e `POST /briefing/ugc`. Validação
  server-side, honeypot + rate-limit anti-spam (sem CAPTCHA de terceiro).
- **Persistência**: tabela `submissions` no SQLite (`form_type`, `payload` JSON,
  `created_at`). Fonte de verdade das respostas; sobrevive a restart; já entra no
  backup Litestream→R2 de graça.
- **HTMX** pro envio (POST + swap de "obrigada!" sem full reload), coerente com o
  resto do site.

### Como a Isa recebe/lê as respostas

O Go guarda; falta o canal de leitura. Duas camadas:

1. **Notificação por e-mail** a cada envio — **único ponto que usa serviço
   externo**: **Resend** (decidido). Free tier generoso, API HTTP simples (o Go
   faz um POST, sem SMTP cru na mão). Precisa: conta Resend, uma API key
   (`RESEND_API_KEY` no `.env`, no padrão do `config.go`) e verificar o domínio
   `isabelpontes.com.br` (registros DNS na Cloudflare) pra os e-mails não caírem
   em spam.
2. **Painel `/admin` protegido** (fase 2, opcional) — lista as `submissions`
   direto do SQLite, sem depender de e-mail. Simples: basic-auth + uma página.

> **Fase 2 (futuro, opcional):** mover as perguntas pro SQLite e dar à Isa uma
> tela pra editar perguntas e ver respostas — aí vira "Google Forms" completo, sem
> dev no meio. Não é bloqueante pro lançamento; v1 já resolve com perguntas em
> arquivo + e-mail.

---

## 9. SEO, domínio e redirects

- **301** de `isabelpontesportfólio.com.br/*` → `isabelpontes.com.br/midiakit`
  (preserva o pouco de SEO do domínio antigo e consolida autoridade num só).
- OG/Twitter/canonical/JSON-LD por rota nova (estender o que o `layout.gohtml` e
  `seo.go` já fazem). JSON-LD de `Service`/`Offer` pros pacotes.
- Analytics: escolher (Plausible/GoatCounter self-host casam com a pegada
  "privado, sem terceiro pesado" do projeto; ou GA4). — *Questão aberta (§11).*
- Manter a força do projeto: HTML server-rendered → card de preview no
  WhatsApp/Instagram funciona (o motivo original de sair da SPA).

---

## 10. Plano de fases

Seguindo a convenção de fases do repo (o histórico vai até "Fase 6").

- **Fase 7 — Fundação da junção.** Design system unificado (tokens/fontes já
  cobrem; validar contraste WCAG AA). Esqueleto de rotas `/midiakit*` servindo
  páginas estáticas com layout/navbar/footer compartilhados. Navbar ganha o
  ramo "Mídia kit".
- **Fase 8 — One-page do mídia kit.** Montar a página `/midiakit` (rolagem,
  âncoras, nav híbrida — §4b): hero, bio, dicionário UGC/IGC, processo (repensado,
  na identidade), investimento (pacotes em JSON editável + CTA WhatsApp). Métricas
  redesenhadas (dados em struct/JSON). **Sem PDF ainda** — ele é Fase 12.
- **Fase 9 — Portfólio de vídeos.** Manter R2 (§6). Estender schema +
  `migrate_media.py` (gerar poster/thumb) + `sync`. Reescrever o player frágil.
  Filtro por categoria via HTMX (espelha a galeria de fotos). Migrar
  `videos-index.json`/`labels.json`.
- **Fase 10 — i18n.** Rota `/en`, mapa de traduções, hreflang, sitemap bilíngue.
- **Fase 11 — Conversão (formulários de briefing).** Dois forms separados
  (fotografia + UGC, §8): perguntas em arquivo, handlers Go, tabela
  `submissions` no SQLite, envio HTMX. Integrar e-mail transacional (Resend/Brevo)
  pra notificar a Isa. `/admin` de leitura fica pra fase 2.
- **Fase 12 — SEO/deploy final + PDF.** 301 do domínio antigo, JSON-LD de
  serviço, analytics, OG por rota, auditoria A11y AA, atualizar `deploy/README.md`.
  Por último e de baixa prioridade (§4b): script de **pré-render do PDF** (Chrome
  headless offline) espelhando `/midiakit`, em PT e EN, servido de `static/`/R2.

Ordem é por impacto e independência — dá pra publicar incrementalmente (o mídia
kit no ar em PT já na Fase 8, antes de i18n/vídeo maduros).

---

## 11. Questões em aberto (decidir antes da fase relevante)

1. **As perguntas exatas** de cada form (§8) — a Isa fecha a lista final de
   perguntas dos dois briefings antes da Fase 11.
2. **Analytics** (§9) — Plausible/GoatCounter self-host vs GA4.
3. **Número de WhatsApp E.164** — confirmar dígitos com a Isabel (nota em
   `pages.go`).
4. **Slug da rota** — `/midiakit` vs `/mídia-kit` vs `/ugc`. (Recomendo
   `/midiakit`, ASCII, sem acento.)

**Fechadas:** arquitetura (fundir no Go+HTMX), domínio (tudo em
`isabelpontes.com.br`), hospedagem de vídeo (mantém R2, custo zero), backend do
formulário (Go + SQLite + dois forms separados), e-mail transacional (**Resend**),
**formato do mídia kit** (one-page híbrida de pitch — §4b) e **PDF** (espelho fiel,
pré-renderizado por script offline, baixa prioridade, a página vem primeiro).
