# Plano: reescrita em Next.js + admin de fotos

As decisões e os porquês estão no `CLAUDE.md` da raiz. Aqui fica só o que falta fazer.
Cada fase termina com o site Go ainda no ar e intocado; a virada é a Fase 8.

## Antes de começar (depende do João)

- Criar o cliente OAuth no Google Cloud e definir `AUTH_GOOGLE_ID`, `AUTH_GOOGLE_SECRET`, `AUTH_SECRET` e `ADMIN_EMAILS`.
- Criar um bucket privado para os originais e definir `R2_BUCKET_ORIGINAIS`; hoje eles iriam para o bucket público.
- Rotacionar as chaves antigas do R2 que vazaram no histórico do repositório React.
- Não apagar o bucket `bebels-portifolio-site` até os originais das 190 fotos serem copiados.
- Confirmar o que está em produção: versão do binário, `ADMIN_*`, Resend, `backup.timer`.
- Opcional: registro DNS de um subdomínio de teste, para a Isabel experimentar pelo celular antes da virada.

## Fase 5 — Painel de briefings

- Conferir um envio real de e-mail pelo Resend, com a chave de produção.

## Fase 6 — Admin de fotos

- Testar o login de verdade com o cliente OAuth do Google.
- Copiar os 190 originais do bucket `bebels-portifolio-site` para o bucket dos originais e preencher `photos.original`.
- Teste de uso com a Isabel no celular.

## Fase 8 — Virada

- Trocar o `reverse_proxy` do Caddy de `127.0.0.1:8080` para `127.0.0.1:8082` (exige `sudo`) e validar o site público, o login e o card de preview no WhatsApp.
- Manter o serviço `site` do Go parado na VM por duas semanas como retorno.
- Depois disso: remover o código Go, o Dockerfile, o `fly.toml` e os arquivos de Litestream; marcar a tag `go-final` antes.

## Depois da virada

- Link para `/privacidade` no rodapé e junto ao botão de enviar do briefing.
- Tela para apagar um briefing, para atender pedido de exclusão de dados; hoje só direto no banco.

- Adaptar `scripts/gen_video_posters.py` para escrever `lib/midiakit/conteudo.ts`; hoje ele reescreve o arquivo Go.

- Corrigir os defeitos visuais herdados, um por mudança: navbar branca sobre creme, folga sob a navbar em `/briefing`, 404 e admin, impressão no tema escuro, pesos 500 e 600 da Jost.
- Pendências antigas que a reescrita não resolve sozinha: 301 do domínio antigo do mídia kit (o punycode do `Caddyfile` está errado; o certo é `xn--isabelpontesportflio-r8b`), número do WhatsApp com 12 dígitos, dois handles de Instagram, revisão do texto em inglês, HSTS de 5 minutos, analytics.
