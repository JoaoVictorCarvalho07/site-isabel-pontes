<#
.SYNOPSIS
  Task runner do projeto.

  Existe em vez de um Makefile porque nao ha make nesta maquina (Windows), e
  instalar um so para rodar quatro comandos ia contra o objetivo do projeto de
  reduzir dependencias. As tarefas sao curtas de proposito -- se alguma crescer,
  ela vira um subcomando do proprio binario Go, nao mais linha aqui.

.EXAMPLE
  .\make.ps1 tools     # baixa o dart-sass standalone
  .\make.ps1 css       # compila assets/scss -> static/css/app.css
  .\make.ps1 watch     # recompila o css ao salvar
  .\make.ps1 run       # sobe o servidor local
  .\make.ps1 sync      # R2 -> SQLite
  .\make.ps1 build     # binario local
  .\make.ps1 deploy    # cross-compila ARM64 p/ a VM da Oracle
#>
param(
  [Parameter(Position = 0)]
  [ValidateSet('tools', 'css', 'watch', 'run', 'sync', 'build', 'deploy', 'test')]
  [string]$Task = 'run'
)

$ErrorActionPreference = 'Stop'
Set-Location $PSScriptRoot

$SASS = "tools\dart-sass\sass.bat"
$SASS_VERSION = "1.101.0"

# --- Configuracao de PRODUCAO (embutida no binario pelo `deploy`) -------------
# Estes valores viram os PADROES do binario (ver internal/config/config.go), para
# a VM nao precisar de arquivo .env. Uma variavel de ambiente na VM, se existir,
# ainda tem prioridade sobre eles.
#
# Só URL/porta aqui. NUNCA segredo: o valor fica legivel dentro do binario.
#
# BASE_URL = o endereco PUBLICO (o que o visitante ve). Entra no canonical,
# og:url, og:image, hreflang e sitemap -- e o que faz o card do WhatsApp
# funcionar. Ao apontar o dominio, troque para 'https://isabelpontes.com.br'.
$PROD_BASE_URL = 'https://isabelpontes.com.br'
# ADDR = onde o Go ESCUTA. Com o Caddy na frente (desde 2026-07-23), escuta so
# no loopback: a 8080 deixa de ser acessivel pela internet e ninguem pula o
# HTTPS -- todo trafego passa obrigatoriamente pelo Caddy na 443.
$PROD_ADDR = '127.0.0.1:8080'
# Arquitetura da VM: Ampere A1 = arm64; VM.Standard.E2.1.Micro (AMD) = amd64.
# Confira na VM com:  uname -m     (aarch64 -> arm64, x86_64 -> amd64)
$PROD_ARCH = 'arm64'
# Onde o binario vive na VM. Precisa bater com o WorkingDirectory/ExecStart do
# systemd unit instalado (ver deploy/site-simples.service).
$PROD_DIR = '/home/joao/apps/site-test'

function Need-Sass {
  if (-not (Test-Path $SASS)) {
    throw "dart-sass nao encontrado. Rode: .\make.ps1 tools"
  }
}

switch ($Task) {
  'tools' {
    $url = "https://github.com/sass/dart-sass/releases/download/$SASS_VERSION/dart-sass-$SASS_VERSION-windows-x64.zip"
    $zip = "$env:TEMP\dart-sass.zip"
    New-Item -ItemType Directory -Force -Path tools | Out-Null
    Invoke-WebRequest -Uri $url -OutFile $zip -UseBasicParsing
    Expand-Archive -Path $zip -DestinationPath tools -Force
    Remove-Item $zip
    & $SASS --version
  }

  'css' {
    Need-Sass
    # --style=compressed: o CSS e commitado, entao ele vai versionado ja minificado.
    & $SASS assets/scss/app.scss static/css/app.css --style=compressed --no-source-map
    $kb = [math]::Round((Get-Item static/css/app.css).Length / 1KB, 1)
    Write-Host "static/css/app.css  $kb KB"
  }

  'watch' {
    Need-Sass
    & $SASS --watch assets/scss/app.scss static/css/app.css --no-source-map
  }

  'run' { go run ./cmd/site serve }

  'sync' { go run ./cmd/site sync }

  'test' { go test ./... }

  'build' { go build -o site.exe ./cmd/site; Write-Host "site.exe" }

  'deploy' {
    # Cross-compila o binario linux/arm64 e, se SITE_HOST estiver setado, faz o
    # scp + restart. ARM64 porque o free tier da Oracle e Ampere A1. CGO_ENABLED=0
    # so e possivel porque o driver SQLite e modernc (Go puro) -- com
    # mattn/go-sqlite3 isto exigiria um cross-toolchain C para ARM.
    #
    #   $env:SITE_HOST='oracle'; .\make.ps1 deploy
    #
    # Use o ALIAS do ~/.ssh/config, nao 'joao@IP': a chave privada esta declarada
    # via IdentityFile no bloco `Host oracle`, e essa linha so vale quando o
    # destino casa com o alias. Com 'joao@IP' o ssh ignora o bloco inteiro e
    # falha com "Permission denied (publickey)".
    #
    # Sem SITE_HOST, so compila (voce faz o scp na mao). Primeira instalacao:
    # siga deploy/README.md.
    # -X injeta a config de producao (topo deste arquivo) como padrao do binario,
    # para a VM nao precisar de .env. Ver internal/config/config.go.
    $cfgPkg = 'github.com/JoaoVictorCarvalho07/site-da-bebel-htmx/internal/config'
    $ldflags = "-s -w -X $cfgPkg.baseURLPadrao=$PROD_BASE_URL -X $cfgPkg.addrPadrao=$PROD_ADDR"

    $env:GOOS = 'linux'; $env:GOARCH = $PROD_ARCH; $env:CGO_ENABLED = '0'
    go build -ldflags $ldflags -o site ./cmd/site
    Remove-Item Env:GOOS, Env:GOARCH, Env:CGO_ENABLED
    $mb = [math]::Round((Get-Item site).Length / 1MB, 1)
    Write-Host "site (linux/$PROD_ARCH, $mb MB) compilado"
    Write-Host "  BASE_URL embutido: $PROD_BASE_URL"
    Write-Host "  ADDR     embutido: $PROD_ADDR"

    if ($env:SITE_HOST) {
      Write-Host "enviando para $env:SITE_HOST ..."
      scp site "$($env:SITE_HOST):/tmp/site"
      # `mv` (nao `cp`): substituir um binario EM EXECUCAO com cp da "Text file
      # busy". O mv so troca a entrada do diretorio -- o processo velho segue no
      # inode antigo ate o restart. Requer o systemd unit instalado.
      $remoto = "mv /tmp/site $PROD_DIR/site && chmod +x $PROD_DIR/site && sudo systemctl restart site && echo reiniciado"
      ssh $env:SITE_HOST $remoto
    }
    else {
      Write-Host "SITE_HOST nao setado -- so compilei. Veja deploy/README.md."
    }
  }
}
