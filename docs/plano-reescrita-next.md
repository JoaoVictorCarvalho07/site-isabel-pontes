# Plano: reescrita em Next.js + admin de fotos

As decisões e os porquês estão no `CLAUDE.md` da raiz. Aqui fica só o que falta fazer.
Cada fase termina com o site Go ainda no ar e intocado; a virada é a Fase 8.

## Antes de começar (depende do João)

- Criar o cliente OAuth no Google Cloud e definir `AUTH_GOOGLE_ID`, `AUTH_GOOGLE_SECRET`, `AUTH_SECRET` e `ADMIN_EMAILS`.
- Criar um bucket privado para os originais e definir `R2_BUCKET_ORIGINAIS`; hoje eles iriam para o bucket público.
- Rotacionar as chaves antigas do R2 que vazaram no histórico do repositório React.
- Não apagar o bucket `bebels-portifolio-site` até os originais das 190 fotos serem copiados.
- Confirmar o que está em produção: versão do binário, `ADMIN_*`, Resend, `backup.timer`.
- Criar o registro DNS do subdomínio de homologação no registro.br.

## Fase 5 — Painel de briefings

- Conferir um envio real de e-mail pelo Resend, com a chave de produção.

## Fase 6 — Admin de fotos

- Testar o login de verdade com o cliente OAuth do Google.
- Copiar os 190 originais do bucket `bebels-portifolio-site` para o bucket dos originais e preencher `photos.original`.
- Teste de uso com a Isabel no celular.

## Fase 7 — Imagens fixas do layout

- Tirar do pacote as imagens que hoje exigem redeploy: hero da home, carrossel do Sobre, card do mídia kit.
- Tela no admin para trocar cada uma, gerando as mesmas larguras e formatos usados hoje.

## Fase 8 — Virada

- Enviar a branch ao GitHub e conferir que o workflow `pacote` gera o pacote linux/arm64.
- Subir o Next na VM em `127.0.0.1` numa porta livre, sob o subdomínio de homologação.
- Conferir o card de preview no WhatsApp e no depurador do Facebook.
- Adaptar `backup.sh` ao novo caminho e testar uma restauração.
- Copiar o banco e os anexos mais recentes, trocar o `reverse_proxy` do Caddy e validar.
- Manter o binário Go parado na VM por duas semanas como retorno.
- Depois disso: remover o código Go, o Dockerfile, o `fly.toml` e os arquivos de Litestream; marcar a tag `go-final` antes.

## Depois da virada

- Adaptar `scripts/gen_video_posters.py` para escrever `lib/midiakit/conteudo.ts`; hoje ele reescreve o arquivo Go.

- Corrigir os defeitos visuais herdados, um por mudança: navbar branca sobre creme, folga sob a navbar em `/briefing`, 404 e admin, impressão no tema escuro, pesos 500 e 600 da Jost.
- Pendências antigas que a reescrita não resolve sozinha: 301 do domínio antigo do mídia kit (o punycode do `Caddyfile` está errado; o certo é `xn--isabelpontesportflio-r8b`), número do WhatsApp com 12 dígitos, dois handles de Instagram, revisão do texto em inglês, HSTS de 5 minutos, analytics.
