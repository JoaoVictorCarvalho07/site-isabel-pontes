// Package site existe apenas para embutir templates/ e static/ no binario.
//
// Mora na raiz do modulo porque os padroes de //go:embed sao relativos ao
// diretorio do PACOTE, nao ao modulo: um arquivo em cmd/site/ nao consegue
// embutir ../../templates. Este e o motivo de o pacote parecer vazio.
package site

import "embed"

// Assets carrega templates/ e static/ dentro do binario. E o que permite o
// deploy ser um scp de um arquivo so -- sem rsync de diretorio, sem volume de
// asset, sem imagem de container carregando arquivo solto.
//
// O prefixo `all:` faz o embed incluir arquivos iniciados por "_" e ".", que
// ele ignora por padrao.
//
//go:embed all:templates all:static
var Assets embed.FS
