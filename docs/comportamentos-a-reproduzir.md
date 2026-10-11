# Comportamentos a reproduzir no Next

Tirados dos 67 testes Go (`internal/**/*_test.go`). Cada item vira um teste no Next e sai desta lista quando passar.

## Páginas e render

- `/`, `/sobre`, `/contato` e `/midiakit` respondem 200 com seus marcadores: "R$ 550", "Mais popular", `wa.me/554191977011`, `conic-gradient`.
- Toda página tem `<title>`, `description`, canonical absoluto e `<h1>` onde hoje existe.
- Rota inexistente responde a página 404 com status 404, no idioma do prefixo.
- Asset ausente não derruba a página.

## Idioma

- `/en` tem `<html lang="en">`, "Welcome to my universe" e hreflang `pt-BR`, `en` e `x-default`.
- O seletor de idioma usa caminho relativo; o canonical é absoluto.

## SEO

- Open Graph, Twitter Card e JSON-LD `ProfessionalService` em `/`, `/galeria`, `/sobre` e `/contato`, com `og:image` absoluta.
- `robots.txt` tem a linha `Sitemap:` e bloqueia `/admin`.
- `sitemap.xml` é XML válido, inclui as páginas e um `/galeria?tema=<slug>` por tema visível, e não inclui o admin.

## Mídia kit

- Com o formulário UGC invisível, a página renderiza sem a seção de briefing.

## Briefing

- Envio válido grava a resposta com o idioma do visitante.
- Erro de validação preserva os valores digitados e não grava.
- Honeypot preenchido responde sucesso sem gravar.
- Select com valor fora da lista é rejeitado.
- Segundo envio do mesmo IP em menos de 30 s é barrado com "Espera um minutinho".
- Sem JavaScript: responde a página inteira, com 422 em erro.
- Slug de formulário desconhecido responde 404; formulário invisível sai das rotas.
- Em inglês: `action` em `/en/briefing/…` e mensagem "Got your briefing".
- Contato aceita e-mail ou telefone com 10 ou 11 dígitos, ou 12 ou 13 começando por 55; contato inválido bloqueia o envio.

## Anexos

- A imagem é gravada com nome gerado, nunca o nome enviado.
- HTML, shell, PDF e SVG disfarçados de imagem são recusados.
- JPEG, GIF, HEIC e AVIF são aceitos.
- Nome com `../` não escapa da pasta de anexos.
- No máximo 4 anexos por campo.
- O anexo só é servido com login, com `nosniff` e o `Content-Type` correto.

## Painel de briefings

- Sem login, o painel não é acessível.
- A resposta aparece com o rótulo da pergunta, inclusive de pergunta que foi ocultada depois.
- Link em resposta só vira clicável se for `http(s)://`; o resto é escapado.
- O anexo aparece no painel sem expor o caminho em disco.

## Editor de formulários

- O banco vazio é semeado com os dois formulários, uma vez só.
- Salvar título, subtítulo, botão e mensagem de sucesso; título vazio responde 422.
- Editar o texto PT limpa o EN daquele campo e marca "EN a revisar"; mexer só em "obrigatória" não.
- Pergunta nova recebe chave `q_…` e entra no fim.
- Select sem opção responde 422, na criação e na troca de tipo.
- Mover pergunta para cima e para baixo.
- Ocultar e mostrar pergunta: some do formulário público, continua no painel.
- Apagar pergunta com resposta responde 422; sem resposta, apaga.
- Trocar o tipo preserva a chave e as respostas antigas.
- A tela mostra quantas respostas cada pergunta tem.
- A edição aparece no site público na hora.
- POST sem `Origin` ou com `Origin` de outro site responde 403.

## Banco

- A cópia de backup feita com escrita concorrente sai íntegra.
