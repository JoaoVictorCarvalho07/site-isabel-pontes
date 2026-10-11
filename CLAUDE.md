# Site Isabel Pontes

Site da fotógrafa Isabel Pontes (fotografia + mídia kit), em https://isabelpontes.com.br.
Hoje em Go + HTMX + SQLite. Em reescrita para Next.js, com admin de fotos para a própria Isabel.
As tarefas estão em `docs/plano-reescrita-next.md`. Este arquivo guarda só as decisões.

## Decisões da reescrita (2026-10-08)

### Stack: Next.js sozinho, sem NestJS

- **Next.js (App Router) faz front e back.** Server Components leem o banco direto; Route Handlers recebem formulários e uploads.
- **Um layout raiz só, com idioma e classe do `<body>` vindos do caminho.** O `proxy.ts` repassa o caminho no cabeçalho `x-caminho` e `lib/requisicao.ts` o traduz em idioma e caminho PT. Vários layouts raiz (um por idioma) foram descartados: com eles o Next entrega o 404 como casca vazia renderizada no navegador. Consequência: toda página é renderizada a cada requisição.
- **Cada rota tem dois arquivos finos** (`app/x/page.tsx` e `app/en/x/page.tsx`) que chamam o mesmo componente em `components/paginas/`.
- **NestJS fica de fora.** O site tem um admin de uma usuária e cerca de 25 rotas. Um Nest separado significaria segundo processo Node na VM, segundo deploy e uma API HTTP entre dois pedaços que hoje são um só, sem nada no escopo que precise disso. Volta à mesa se um dia existir app mobile ou outro cliente consumindo a mesma API.
- **O motivo original de sair do React continua atendido:** o HTML chega pronto do servidor, então o card de preview do WhatsApp e do Instagram funciona.
- **O que a reescrita abandona:** "um binário, sem Node no servidor". Troca aceita em nome de manter o projeto na stack que o dono domina.

### Hospedagem: a mesma VM Oracle

- Next com `output: "standalone"`, serviço systemd, atrás do Caddy que já existe.
- **Vercel descartada:** o plano gratuito proíbe uso comercial, o corpo de request é limitado (o briefing aceita 21 MB) e não há disco para SQLite nem para os anexos.
- **O build nunca roda na VM.** O `sharp` e o driver do SQLite têm binário nativo, então o build sai de um runner linux/arm64 do GitHub Actions e chega à VM como pacote pronto.
- Uma instância só. Cache e rate limit em memória dependem disso.

### Dados: SQLite continua

- Mesmo arquivo `site.db`, lido com `better-sqlite3` direto, em funções tipadas em `lib/`. Prisma ficou de fora: a versão 7 exige Node 22.18+ e adaptador de driver, peso demais para sete tabelas. Postgres não entra: a VM não tem, e o backup atual (`site snapshot` + rclone) já cobre um arquivo.
- As tabelas `submissions`, `submission_files`, `forms` e `form_questions` migram intactas. As `chave` das perguntas são a identidade das respostas antigas e não mudam.
- **O banco vira a fonte de verdade das fotos.** O fluxo `manifest.json` + `site sync` é aposentado: ele apaga e recria `photos`, o que destruiria qualquer curadoria feita pelo admin.

### Visual: paridade por construção, conferida por screenshot

- O SCSS atual entra inteiro como folha global, na mesma ordem. Sem Tailwind, sem CSS Modules.
- Os componentes reproduzem o mesmo DOM, as mesmas classes e os mesmos `data-*` dos templates `.gohtml`.
- **Sem `next/image` e sem `next/font`.** Os dois mudam o HTML gerado (wrappers, `srcset`, preload, nome da família, pesos). Ficam `<img>`/`<picture>` escritos à mão e o `@font-face` atual.
- **Navegação pública por carga completa de documento** (`<a>`, não `next/link`). O site atual é MPA e várias animações e estados dependem disso (animação do logo a cada página, navbar e menu zerados).
- Os `url()` do CSS continuam apontando para `/static/...`, servido de `public/static/`.
- "Idêntico" é verificado por diff de screenshot do site Go contra o site Next, não a olho.
- Os defeitos visuais conhecidos do site atual são preservados na reescrita e corrigidos depois, em mudanças separadas.

### Viabilidade medida (2026-10-08)

Protótipo descartável de layout + home em Next 16, com o SCSS atual, rodado na VM num contêiner sem porta exposta.

- **Visual:** a home do Next saiu idêntica à do Go pixel a pixel em 11 de 12 capturas; a 12ª diverge por defeito da referência. Go contra Go dá zero pixel de diferença, então qualquer diferença é real.
- **Texto com variável vai numa string só.** `© {ano} Isabel` vira dois nós de texto no React e desloca os glifos; usar template string.
- **`app.js` funciona sem alteração** no Next, porque a navegação é por carga completa. Ele continua sendo o JavaScript do site público; só o que era HTMX vira componente.
- **Memória na VM:** 39 MB ocioso, 121 MB depois de 500 requisições. O Go usa 29 MB.
- **Resposta na VM:** mediana de 12 ms com leitura do SQLite a cada requisição.
- **`sharp` e `better-sqlite3`** instalam com binário pronto em linux/arm64.
- **Processar uma foto leva de 9 a 11 s** na VM (três AVIFs + LQIP), com pico de 250 MB. Vinte fotos levam cerca de 3 minutos, daí a fila com progresso.

### Linha de base da paridade

- A referência é o site Go rodando local com `data/paridade.db`: cópia do banco de produção com `sort_order = id` em todas as fotos. Ordem fixa dos dois lados, porque o embaralhamento do Go não é reproduzível em JavaScript.
- `npm run paridade:capturar` grava em `paridade/referencia` (fora do git): 8 páginas × PT/EN × claro/escuro × 6 larguras, a impressão do mídia kit e o HTML de cada página. Para capturar o Next: `-- --base <url> --saida paridade/atual`.
- As capturas são cortadas em 14.000 px de altura e os vídeos são mascarados.
- `paridade/comparar.mjs` compara `paridade/atual` com a referência; `paridade/regiao.mjs` mostra onde ficam os pixels diferentes.
- Não há Go instalado na máquina de desenvolvimento; o binário e os testes rodam pela imagem Docker `golang:1.26`.

### Admin de fotos

- **Público:** a Isabel, não técnica, só PT-BR, provavelmente no celular. Interface mobile-first.
- **Login:** Google OAuth com lista de e-mails permitidos, o mesmo modelo do imageColor. Substitui o Basic Auth.
- **Upload:** o navegador envia o original direto ao R2 por URL assinada; o servidor gera as variantes com `sharp` (AVIF 400, 800 e 1600 px + LQIP), nos mesmos tamanhos de hoje. Processamento em fila de um por vez, porque a VM tem 2 núcleos.
- **Os originais passam a ser guardados** num prefixo privado do R2. Os das 190 fotos atuais estão no bucket antigo `bebels-portifolio-site` (97 MB, nas mesmas chaves da coluna `photos.key`); o bucket `isabel-portfolio-gallery` só tem as variantes.
- **Apagar é reversível:** a foto vai para uma lixeira antes de sumir do R2.
- O painel de briefings e o editor de formulários são portados com o mesmo comportamento.

### Repositório e virada

- Mesmo repositório, branch `reescrita-next`. O Go segue na `main` e em produção até a virada.
- O Next sobe na VM em outra porta, sob um subdomínio de homologação. A virada é trocar o `reverse_proxy` do Caddy; voltar atrás é desfazer essa linha.
