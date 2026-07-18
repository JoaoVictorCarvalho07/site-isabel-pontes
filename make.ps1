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
    #   $env:SITE_HOST='opc@1.2.3.4'; .\make.ps1 deploy
    #
    # Sem SITE_HOST, so compila (voce faz o scp na mao). Primeira instalacao:
    # siga deploy/README.md.
    $env:GOOS = 'linux'; $env:GOARCH = 'arm64'; $env:CGO_ENABLED = '0'
    go build -ldflags '-s -w' -o site ./cmd/site
    Remove-Item Env:GOOS, Env:GOARCH, Env:CGO_ENABLED
    $mb = [math]::Round((Get-Item site).Length / 1MB, 1)
    Write-Host "site (linux/arm64, $mb MB) compilado"

    if ($env:SITE_HOST) {
      Write-Host "enviando para $env:SITE_HOST ..."
      scp site "$($env:SITE_HOST):/tmp/site"
      # move para o lugar e reinicia (systemd). Requer o setup do README feito.
      ssh $env:SITE_HOST 'sudo mv /tmp/site /opt/site/site && sudo chown site:site /opt/site/site && sudo chmod +x /opt/site/site && sudo systemctl restart site && echo reiniciado'
    }
    else {
      Write-Host "SITE_HOST nao setado -- so compilei. Veja deploy/README.md."
    }
  }
}
