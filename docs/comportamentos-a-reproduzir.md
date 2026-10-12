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

## Mídia kit

- Com o formulário UGC invisível, a página renderiza sem a seção de briefing.

## Banco

- A cópia de backup feita com escrita concorrente sai íntegra.
