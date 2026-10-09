param([switch]$Forcar)

$ErrorActionPreference = 'Stop'

$alvo = if ($env:SITE_HOST) { $env:SITE_HOST } else { 'oracle' }
$dirRemoto = '/home/joao/apps/site-test'
$tempRemoto = '/tmp/site-puxar'
$destino = Join-Path (Split-Path -Parent $PSScriptRoot) 'data'

function Executar {
    param([string]$Descricao, [scriptblock]$Comando)
    Write-Host "==> $Descricao"
    & $Comando
    if ($LASTEXITCODE -ne 0) { throw "falhou: $Descricao" }
}

if ((Test-Path (Join-Path $destino 'site.db')) -and -not $Forcar) {
    throw "ja existe $destino\site.db; rode de novo com -Forcar para substituir"
}

New-Item -ItemType Directory -Force $destino | Out-Null

$preparar = @(
    'set -e'
    "rm -rf $tempRemoto"
    "mkdir -p $tempRemoto"
    "cd $dirRemoto"
    "./site snapshot $tempRemoto/site.db"
    "if [ -f data/manifest.json ]; then cp data/manifest.json $tempRemoto/; elif [ -f manifest.json ]; then cp manifest.json $tempRemoto/; fi"
    "if [ -d data/uploads ]; then tar czf $tempRemoto/uploads.tgz -C data uploads; fi"
    "ls -la $tempRemoto"
) -join '; '

Executar 'gerando copia consistente na VM' { ssh $alvo $preparar }
Executar 'baixando' { scp "${alvo}:$tempRemoto/*" $destino }
Executar 'limpando a VM' { ssh $alvo "rm -rf $tempRemoto" }

$pacote = Join-Path $destino 'uploads.tgz'
if (Test-Path $pacote) {
    $uploads = Join-Path $destino 'uploads'
    if (Test-Path $uploads) { Remove-Item -Recurse -Force $uploads }
    Executar 'extraindo anexos' { tar -xzf $pacote -C $destino }
    Remove-Item $pacote
}

if (-not (Test-Path (Join-Path $destino 'manifest.json'))) {
    Write-Warning 'manifest.json nao foi encontrado na VM'
}

Get-ChildItem $destino | Format-Table Name, Length, LastWriteTime
