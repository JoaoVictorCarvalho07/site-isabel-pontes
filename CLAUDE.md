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

- **O Next roda num contêiner Docker** (`node:22-slim`, nome `site-next`), com `--restart unless-stopped`, 1 GB de limite e porta só em loopback (`127.0.0.1:8082`), atrás do Caddy que já existe. Um serviço systemd foi descartado: a VM não tem Node instalado e o `sudo` pede senha, então nada disso seria automatizável.
- **Vercel descartada:** o plano gratuito proíbe uso comercial, o corpo de request é limitado (o briefing aceita 21 MB) e não há disco para SQLite nem para os anexos.
- **O build nunca roda na VM.** O workflow `pacote` do GitHub Actions compila num runner linux/arm64, dentro da mesma imagem `node:22-slim`, e publica o pacote standalone como artefato. O `better-sqlite3` não tem binário pronto e é compilado ali.
- **`npm run implantar`** baixa o pacote do commit atual, envia à VM e roda `deploy/next/subir.sh`, que extrai em `~/apps/site-next/versoes/`, recria o contêiner, espera responder e guarda as três versões mais recentes.
- **Na VM, fora do git:** `~/apps/site-next/.env` (segredos, modo 600) e `~/apps/site-next/data/` (banco e anexos, montado em `/dados`).
- **Backup diário:** o `backup.timer` que já existia chama `~/apps/site-test/backup.sh`, que agora só repassa para `deploy/next/backup.sh`. Esse copia o banco pelo próprio contêiner (API de backup do SQLite) e envia banco e anexos ao bucket `isabel-site-backup` com rclone. Restauração testada em 2026-10-13.
- **O binário do Go nunca pode abrir o banco do Next:** ao abrir, ele regrava `schema_version = 4`, e o Next tentaria reaplicar as migrações 5 e 6 e não subiria.
- **O que exige `sudo` fica com o dono:** editar `/etc/caddy/Caddyfile` e recarregar o Caddy.
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

### Galeria

- **O filtro navega pelo servidor.** O botão chama `router.push("/galeria?tema=…", { scroll: false })`; o Next renderiza a grade no servidor e troca só ela, sem recarregar o documento e sem rolar. É o equivalente do swap do HTMX, e a URL com `?tema=` continua sendo a página completa.
- **O catálogo fica em memória** (`lib/fotos.ts`) e é remontado quando o dia vira em `America/Sao_Paulo` ou quando `invalidarCatalogo()` é chamado.
- **A ordem do dia é função pura da data**: fotos com `sort_order` no topo, o resto embaralhado com a semente AAAAMMDD. No Go a ordem dependia de quando o processo tinha subido.
- **Lightbox e fade-in das fotos:** o lightbox continua no `app.js`; o fade-in é do componente `ImagemGrid`, porque o `app.js` só o refazia em evento do HTMX.
- **ETag e 304 da galeria não foram portados.** O Next não os oferece em página dinâmica e não há CDN na frente; cada visita baixa o HTML de novo.
- **O HTML da galeria ficou maior:** 130 KB comprimido contra 59 KB no Go, porque o App Router embute os dados da página uma segunda vez para hidratar.
- `aria-label` do botão de foto e textos do lightbox seguem só em português nas duas línguas, como no Go.

### Mídia kit

- **O conteúdo fica em `lib/midiakit/conteudo.ts`** (pacotes, preços, processo, métricas, vídeos), convertido do Go. Cada texto é um par `[pt, en]`; `resolver()` cai no português quando o inglês está vazio.
- **O filtro de vídeos é estado no navegador** (`components/midiakit/Portfolio.tsx`): os 26 vídeos já vêm no HTML e a URL não muda, como no Go. Não existe mais a rota `/midiakit/videos`.
- **O player e a impressão continuam no `app.js`.** Trocar o filtro remonta os grupos, o que descarta um vídeo aberto, igual ao swap do HTMX.
- **Os formulários são lidos do banco a cada requisição** (`lib/forms.ts`), sem cache: a edição no admin aparece na hora.
- **Texto colado a um elemento vai no mesmo nó.** O rótulo do campo e o espaço antes de "(opcional)" são uma string só; separados, o React cria dois nós de texto e um glifo se desloca.

### Briefing

- **Toda a regra mora em `processarBriefing`** (`lib/briefing/processar.ts`): honeypot, validação, limite de 30 s por IP, gravação, anexos e aviso por e-mail.
- **Dois caminhos chegam nela.** Com JavaScript, o formulário chama a server action `enviarBriefing` e troca o próprio conteúdo. Sem JavaScript, o `<form>` posta em `/briefing/{slug}`, que responde 303 para a página com `?briefing=enviado|corrija|espere|erro`.
- **O redirecionamento usa caminho relativo no `Location`.** Atrás do Caddy, `request.nextUrl.origin` é o endereço interno do contêiner (`0.0.0.0:3000`), não o do site.
- **`useActionState` com `<form action={fn}>` foi descartado:** no envio sem JavaScript o Next 16.4 entra em laço infinito ao renderizar a resposta e trava o processo inteiro.
- **Sem JavaScript, um erro não devolve os valores digitados** nem aponta o campo; só a mensagem geral. O Go devolvia os dois. Os atributos `required` e `type` já barram o caso comum no navegador.
- **Anexos:** só os quatro primeiros arquivos, 5 MB cada, tipo decidido pelos bytes (`lib/briefing/tipos.ts`), nome em disco gerado. O que não passa é descartado sem derrubar o envio.
- **O `reply_to` do aviso é o e-mail do visitante** quando ele informou um; no Go era sempre o da Isabel.
- **Quebras de linha das respostas são gravadas como LF**; o navegador envia CRLF.
- O limite do corpo é 22 MB em `next.config.ts` (`serverActions.bodySizeLimit` e `proxyClientMaxBodySize`); o padrão do proxy é 10 MB.

### Página de privacidade

- `/privacidade` e `/en/privacidade` descrevem o que o site realmente faz com dados; mudou a coleta, muda o texto em `components/paginas/Privacidade.tsx`.
- Ela está no sitemap, mas nenhuma página aponta para ela ainda: um link no rodapé mudaria todas as capturas de paridade, então fica para depois da virada.
- O cliente OAuth do Google está em modo de teste, com os e-mails permitidos também cadastrados lá como usuários de teste. Publicar o app exige essa página no ar.

### Testes

- `npm run teste`: unitários com o executor do Node, em `testes/*.test.ts`. Módulo testado não pode importar `better-sqlite3` nem usar import sem extensão.
- `npm run teste:servidor` sobe o pacote compilado na porta 8098 com `data/teste.db` (cópia de `site.db`), `data/teste-uploads` e `data/teste-r2`, recriados a cada subida. `npm run paridade:servidor` faz o mesmo a partir de `paridade.db`, para as capturas.
- `npm run teste:login` sobe o mesmo pacote em `http://localhost:3000` com as credenciais de login do `.env.local`, para testar o Google de verdade. Banco e armazenamento continuam descartáveis; nada vai para o R2.
- `npm run teste:e2e`: galeria, mídia kit, briefing, admin de fotos e painel com Playwright contra esse servidor. Briefing e admin gravam no banco, então exigem o servidor recém-subido. O teste do admin forja o cookie de sessão com o `AUTH_SECRET` de teste.

### Admin de fotos

- **Público:** a Isabel, não técnica, só PT-BR, provavelmente no celular. Interface mobile-first em `/admin`, com folha própria (`app/admin/admin.scss`) sobre os tokens do site.
- **Login:** Google OAuth pelo Auth.js (`auth.ts`), sessão em cookie JWT de 30 dias, sem tabela de usuários. Só entra e-mail verificado que esteja em `ADMIN_EMAILS`. Substitui o Basic Auth do Go.
- **Toda página, ação e rota do admin chama `exigirAdmin()` ou `emailAdmin()`.** O `proxy.ts` não protege nada; a checagem é em cada ponto de entrada.
- **O upload passa pelo servidor**, uma foto por requisição em `POST /admin/api/fotos`, e só responde depois de gerar as variantes e gravar no banco. O navegador envia em fila e mostra o progresso. A URL assinada direto ao R2 foi descartada: exigiria CORS no bucket e uma fila de processamento à parte, e uma foto por vez já respeita os 2 núcleos da VM.
- **Variantes iguais às de hoje:** AVIF de 400, 800 e 1600 px e LQIP WebP de 24 px (`lib/admin/imagem.ts`), com a orientação EXIF aplicada. Limite de 40 MB por foto; HEIC é recusado porque o `sharp` não o decodifica.
- **Foto nova entra publicada**, no tema escolhido antes do envio.
- **Os originais são guardados** em `originais/{tema}/` no bucket de `R2_BUCKET_ORIGINAIS`. Sem essa variável eles caem no bucket público e ficam acessíveis a quem souber o endereço. Os das 190 fotos antigas estão no bucket `bebels-portifolio-site` (97 MB, nas chaves da coluna `photos.key`).
- **Sem credenciais do R2, o armazenamento é em disco** (`data/r2-local`, servido em `/dev-r2`), para desenvolver e testar. Com credenciais, `/dev-r2` responde 404.
- **Fixar no topo usa `sort_order`**; a ordem entre as fixadas se troca por botões de subir e descer, não por arrastar.
- **Apagar é em dois passos:** a lixeira marca `excluida_em` e tira a foto do site; "apagar de vez" remove as variantes e o original do R2 e a linha do banco.
- **Tema não se apaga**, só se oculta. O slug nasce do nome e não muda ao renomear.
- **O que precisa ser visto por rotas diferentes fica em `globalThis`** (catálogo de fotos, limitador de envios): cada rota do Next tem a própria cópia dos módulos, então uma variável de módulo invalidada pelo upload não era a mesma lida pela galeria.
- **Código que monta caminho em disco a partir de variável leva `/* turbopackIgnore: true */`**; sem isso o build empacota o repositório inteiro, inclusive `data/`.
- **Migrações de esquema** ficam em `lib/migracoes.ts` e rodam ao abrir o banco, pela `schema_version` da tabela `meta`. A versão 5 criou `photos.original` e `photos.excluida_em`; a 6, `layout_imagens` e `sobre_carrossel`. Depois dela o `site sync` do Go não pode mais rodar nesse banco.

### Imagens fixas do site

- **A Isabel troca pelo painel, em `/admin/site`:** capa da home (uma foto em pé para celular e uma deitada para computador), cartão do mídia kit, as duas fotos do texto do Sobre e o carrossel do Sobre.
- **Cada lugar é um "slot"** descrito em `lib/layout/slots.ts`, com as larguras e formatos que o HTML atual espera. Sem troca, o slot aponta para o arquivo de `public/static/`, e o site fica idêntico ao Go.
- **A troca gera os mesmos tamanhos de hoje** e grava em `layout/{slot}/` no R2; a imagem anterior enviada pelo painel é apagada. "Voltar à original" apaga a linha e volta ao arquivo do pacote.
- **O carrossel mora na tabela `sobre_carrossel`**, semeada com as oito fotos originais. Remover uma original só tira a linha; o arquivo continua no pacote.
- **Foto menor que a largura pedida não é ampliada.**
- **O enquadramento da capa continua fixo no CSS** (`object-position`); o painel não tem ajuste de corte.
- Ficaram de fora, ainda no pacote: o vídeo e o pôster da abertura do mídia kit, os prints de celular das métricas, os pôsteres dos vídeos e a imagem de preview de link (`og-image.jpg`).

### Painel de briefings e editor de formulários

- Ficam em `/admin/briefings` e `/admin/formularios`, no mesmo visual do admin de fotos. O comportamento veio dos testes Go; o layout é novo.
- **Resposta de visitante nunca vira HTML.** `trechosComLinks` quebra o texto em pedaços e só os que começam com `http://` ou `https://` viram `<a>`; o React escapa o resto.
- **Anexo só sai por `/admin/arquivo/{id}`**, com login, `nosniff`, CSP `sandbox` e caminho conferido contra a pasta de uploads. Sem login responde 404, não 401, para não revelar que existe.
- **Editar o português limpa o inglês daquele campo** e marca "EN a revisar"; o site em inglês mostra o português até alguém traduzir. Mexer só em "obrigatória" não toca o inglês.
- **Opção de lista que não mudou mantém a tradução.** No Go, salvar qualquer coisa numa pergunta de lista apagava o inglês de todas as opções.
- **A `chave` da pergunta nunca muda**, nem ao trocar o tipo; é ela que liga a pergunta às respostas antigas. Pergunta nova recebe `q_` + 8 hex.
- **Pergunta com resposta não se apaga**, só se oculta; a resposta antiga continua aparecendo no painel com o rótulo.
- **O formulário de cada pergunta não é remontado depois de salvar.** Remontar apagava o que a pessoa já estivesse digitando no mesmo cartão.
- Criar formulário novo e ligar ou desligar um formulário continuam sem tela, como no Go.

### Repositório e virada

- Mesmo repositório, branch `reescrita-next`. O Go segue na `main` e em produção até a virada.
- O Next sobe na VM em outra porta, sob um subdomínio de homologação. A virada é trocar o `reverse_proxy` do Caddy; voltar atrás é desfazer essa linha.
