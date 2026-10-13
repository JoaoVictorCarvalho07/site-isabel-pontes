# Plano: reescrita em Next.js + admin de fotos

As decisões e os porquês estão no `CLAUDE.md` da raiz. Aqui fica só o que falta fazer.
O Next está em produção; o site Go segue na VM só como retorno.

## Depende do João e da Isabel

- Apagar os dois tokens antigos do R2 vazados no repositório público `site-da-bebel` e, se quiser, os buckets `bebels-portifolio-site` e `gallery`.
- Confirmar que o e-mail de aviso de briefing chega na caixa da Isabel.
- Teste de uso do painel pela Isabel no celular, incluindo foto tirada no iPhone.

## Fase 8 — Virada (no ar desde 2026-10-13)

- Colar um link do site no WhatsApp e conferir o card de preview.
- Testar o login do painel em produção.
- Até 2026-10-27: manter o serviço `site` do Go ativo na VM como retorno. Depois, parar o serviço, marcar a tag `go-final` e remover o código Go, o Dockerfile, o `fly.toml` e os arquivos de Litestream.
- Fazer o merge de `reescrita-next` na `main`.

## Depois da virada

- Link para `/privacidade` junto ao botão de enviar do briefing.
- Subir o HSTS do Caddy de 5 minutos para um ano, a partir de 2026-10-20.
- Tela para apagar um briefing, para atender pedido de exclusão de dados; hoje só direto no banco.
- Adaptar `scripts/gen_video_posters.py` para escrever `lib/midiakit/conteudo.ts`; hoje ele reescreve o arquivo Go.
- Corrigir os defeitos visuais herdados, um por mudança: navbar branca sobre creme, folga sob a navbar em `/briefing`, 404 e admin, impressão no tema escuro, pesos 500 e 600 da Jost.
- Pendências antigas que a reescrita não resolve sozinha: 301 do domínio antigo do mídia kit (o punycode do `Caddyfile` está errado; o certo é `xn--isabelpontesportflio-r8b`), número do WhatsApp com 12 dígitos, dois handles de Instagram, revisão do texto em inglês, HSTS de 5 minutos, analytics.
