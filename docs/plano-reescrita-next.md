# Plano: reescrita em Next.js + admin de fotos

As decisões e os porquês estão no `CLAUDE.md` da raiz. Aqui fica só o que falta fazer.
Cada fase termina com o site Go ainda no ar e intocado; a virada é a Fase 8.

## Antes de começar (depende do João)

- Não apagar o bucket `bebels-portifolio-site` até a Fase 6 copiar os originais.
- Criar uma chave R2 com escrita no bucket `isabel-portfolio-gallery` e rotacionar as chaves antigas que vazaram no histórico do repositório React.
- Criar o cliente OAuth no Google Cloud e listar os e-mails que entram no admin.
- Confirmar o que está em produção: versão do binário, `ADMIN_*`, Resend, `backup.timer`.
- Criar o registro DNS do subdomínio de homologação no registro.br.

## Antes da Fase 2

- Corrigir `paridade/capturar.mjs` para esperar a pintura das imagens `loading="lazy"` e refazer a referência: a captura `light/pt/home-375` saiu com um card sem foto.

## Fase 1 — Esqueleto

- Projeto Next (App Router, TypeScript) na branch `reescrita-next`.
- `sass` com o `assets/scss/app.scss` atual como folha global; `static/` movido para `public/static/`.
- Prisma apontando para o `site.db` existente, com o schema obtido por introspecção.
- Layout raiz: script de tema antes da pintura, skip link, navbar, rodapé, `<head>` com canonical, hreflang, Open Graph e JSON-LD.
- Cabeçalhos: `immutable` de 1 ano em `/static/`, e os de segurança que hoje estão no Caddy.
- i18n por prefixo `/en`, com o dicionário de `i18n.go` portado.
- Pipeline do GitHub Actions gerando o pacote standalone para linux/arm64.

## Fase 2 — Páginas estáticas

- Home, Sobre, Contato e 404, com DOM e classes idênticos aos templates.
- Componentes cliente no lugar do `app.js`: menu mobile, navbar que some no scroll, alternador de tema, carrossel.
- Diff de screenshot contra a linha de base até zerar.

## Fase 3 — Galeria

- Leitura das fotos publicadas com cache invalidado por tag.
- Shuffle diário com semente pela data em `America/Sao_Paulo`, fotos com ordem fixa no topo.
- `/galeria?tema=` renderizado no servidor; filtro sem recarregar a página e sem rolar ao topo.
- Grid com LQIP e fade-in, lightbox em `<dialog>` com teclado, contador e pré-carga dos vizinhos.
- ETag e `Cache-Control` equivalentes aos atuais.

## Fase 4 — Mídia kit

- Página única com as seções-âncora, métricas, pacotes, processo e donuts.
- Portfólio de vídeos com filtro por categoria e player criado no clique.
- CSS de impressão e botão de PDF, abrindo os `<details>` e forçando as imagens antes de imprimir.

## Fase 5 — Briefing e painel atual

- Formulários renderizados a partir de `forms` e `form_questions`.
- Envio por Route Handler multipart: mesma validação, honeypot, rate limit de 30 s por IP, anexos conferidos pelos bytes, limites de 5 MB × 4.
- Funcionar sem JavaScript, como hoje.
- Aviso por e-mail via Resend.
- Login por Google OAuth com lista de permitidos.
- Painel de briefings e editor de formulários com o comportamento coberto pelos testes Go.
- `robots.txt` e `sitemap.xml`.

## Fase 6 — Admin de fotos

- Copiar os 190 originais do bucket `bebels-portifolio-site` para o prefixo privado e ligar cada um à sua foto pela `key`.
- Migração do schema: estado da foto (processando, publicada, oculta, na lixeira), chave do original, e fim da dependência do `sync`.
- Upload de várias fotos pelo celular: original direto ao R2 por URL assinada, fila de processamento com `sharp`, progresso visível.
- Por foto: tema, texto alternativo, publicar ou ocultar, fixar no topo, mover para a lixeira.
- Reordenar as fixadas arrastando.
- Temas: criar, renomear, ordenar, ocultar.
- Lixeira com restauração e exclusão definitiva.
- Toda alteração invalida o cache da galeria na hora.
- Teste de uso com a Isabel no celular antes de dar por pronto.

## Fase 7 — Imagens fixas do layout

- Tirar do pacote as imagens que hoje exigem redeploy: hero da home, carrossel do Sobre, card do mídia kit.
- Tela no admin para trocar cada uma, gerando as mesmas larguras e formatos usados hoje.

## Fase 8 — Virada

- Subir o Next na VM em `127.0.0.1` numa porta livre, sob o subdomínio de homologação.
- Conferir o card de preview no WhatsApp e no depurador do Facebook.
- Adaptar `backup.sh` ao novo caminho e testar uma restauração.
- Copiar o banco e os anexos mais recentes, trocar o `reverse_proxy` do Caddy e validar.
- Manter o binário Go parado na VM por duas semanas como retorno.
- Depois disso: remover o código Go, o Dockerfile, o `fly.toml` e os arquivos de Litestream; marcar a tag `go-final` antes.

## Depois da virada

- Corrigir os defeitos visuais herdados, um por mudança: navbar branca sobre creme, folga sob a navbar em `/briefing`, 404 e admin, impressão no tema escuro, pesos 500 e 600 da Jost.
- Pendências antigas que a reescrita não resolve sozinha: 301 do domínio antigo do mídia kit (o punycode do `Caddyfile` está errado; o certo é `xn--isabelpontesportflio-r8b`), número do WhatsApp com 12 dígitos, dois handles de Instagram, revisão do texto em inglês, HSTS de 5 minutos, analytics.
