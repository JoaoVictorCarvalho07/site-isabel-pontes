# Deploy

Alvo primário: **VM Oracle Cloud (ARM64) + systemd + Caddy + Litestream**.
Fallback: **Fly.io** (Docker), caso a Oracle demore a liberar máquina.

```
Internet ──▶ Cloudflare (DNS) ──▶ Caddy :443 (TLS) ──▶ Go :8080 (127.0.0.1)
                                                          │
                                                    SQLite (/opt/site/data)
                                                          │
                                                    Litestream ──▶ R2 (backup)
```

O runtime é **um binário Go** (templates, CSS, JS e imagens embutidos). O deploy
é `scp` do binário + `systemctl restart`. O SQLite vive no disco da VM; o
Litestream replica para o R2 continuamente (backup/restore).

---

## Pré-requisitos (uma vez)

1. **Domínio** apontado (ou a apontar) para o IP da VM: `isabelpontes.com.br`.
2. **Bucket de backup R2 PRIVADO** — ex.: `isabel-litestream`. **Não** reuse o
   bucket de mídia (`isabel-portfolio-gallery`), que é de leitura pública: o
   backup do banco ficaria baixável por qualquer um.
3. **Chave R2 com escrita** nesse bucket de backup (pode ser a mesma da mídia se
   ela tiver escrita, mas o ideal é uma chave dedicada).
4. **Go instalado na sua máquina** (para cross-compilar o binário ARM64). Já está.

---

## Parte A — VM Oracle (primário)

### 1. Criar a instância

- Shape: **Ampere A1 (ARM64)**, Ubuntu 22.04. Se der *"Out of host capacity"*
  (comum no free tier — ver Troubleshooting), o **VM.Standard.E2.1.Micro** (AMD,
  1 GB) roda este site sem suar; nesse caso o binário é `GOARCH=amd64`.
- Região: `sa-saopaulo-1` ou `sa-vinhedo-1`.
- Guarde o IP público e a chave SSH.

### 2. Firewall DUPLO ⚠️ (a pegadinha nº 1 da Oracle)

Abrir a porta na **Security List da VCN** NÃO basta — a imagem Ubuntu da Oracle
vem com `iptables` local bloqueando tudo além do 22. **Os dois** precisam liberar
80 e 443:

**a) Security List da VCN** (painel Oracle → Networking → VCN → Subnet → Security
List → Add Ingress Rules): libere `0.0.0.0/0` TCP **80** e **443**.

**b) Firewall local da VM** (por SSH):
```sh
sudo iptables -I INPUT 6 -m state --state NEW -p tcp --dport 80  -j ACCEPT
sudo iptables -I INPUT 6 -m state --state NEW -p tcp --dport 443 -j ACCEPT
sudo netfilter-persistent save     # persiste o reboot
```
> Se passar horas achando que "é DNS", quase sempre é este firewall local.

### 3. Instalar Caddy + Litestream

```sh
# Caddy
sudo apt install -y debian-keyring debian-archive-keyring apt-transport-https
curl -1sLf 'https://dl.cloudsmith.io/public/caddy/stable/gpg.key' | sudo gpg --dearmor -o /usr/share/keyrings/caddy-stable-archive-keyring.gpg
curl -1sLf 'https://dl.cloudsmith.io/public/caddy/stable/debian.deb.txt' | sudo tee /etc/apt/sources.list.d/caddy-stable.list
sudo apt update && sudo apt install -y caddy

# Litestream (ARM64; troque por -amd64 se for a VM AMD)
curl -L https://github.com/benbjohnson/litestream/releases/latest/download/litestream-linux-arm64.tar.gz | sudo tar -xz -C /usr/local/bin litestream
```

### 4. Usuário, diretórios e o binário

Na VM:
```sh
sudo useradd --system --home /opt/site --shell /usr/sbin/nologin site
sudo mkdir -p /opt/site/data /var/log/caddy
sudo chown -R site:site /opt/site
```

Na sua máquina (Windows), compila o ARM64 e envia binário + manifest:
```powershell
.\make.ps1 deploy   # cross-compila linux/arm64 e faz scp (ver make.ps1)
```
Ou manualmente:
```powershell
$env:GOOS='linux'; $env:GOARCH='arm64'; $env:CGO_ENABLED='0'
go build -ldflags '-s -w' -o site ./cmd/site
scp site opc@SEU_IP:/tmp/site
scp data/manifest.json opc@SEU_IP:/tmp/manifest.json
```
Na VM:
```sh
sudo mv /tmp/site /opt/site/site && sudo chmod +x /opt/site/site
sudo mv /tmp/manifest.json /opt/site/data/manifest.json
sudo chown -R site:site /opt/site
```

### 5. Arquivos de configuração

Copie os três do repo (`deploy/`) para a VM e preencha o `.env`:
```sh
sudo cp deploy/litestream.yml /etc/litestream.yml
sudo cp deploy/site.service   /etc/systemd/system/site.service
sudo cp deploy/Caddyfile      /etc/caddy/Caddyfile
sudo cp deploy/env.example    /opt/site/.env
sudo nano /opt/site/.env      # preencha R2_* e confirme BASE_URL
sudo chown site:site /opt/site/.env && sudo chmod 600 /opt/site/.env
```

### 6. Popular o banco (uma vez)

```sh
sudo -u site env $(cat /opt/site/.env | xargs) /opt/site/site sync
# deve logar: fotos=190 temas=7
```

### 7. Subir os serviços

```sh
sudo systemctl daemon-reload
sudo systemctl enable --now site
sudo systemctl reload caddy      # ou restart, na primeira vez
sudo systemctl status site --no-pager
```

### 8. DNS + TLS

Aponte o `A` de `isabelpontes.com.br` (e `www`) para o IP da VM.

⚠️ **Cloudflare**: se o registro estiver **proxiado** (nuvem laranja), o desafio
do Caddy para emitir o certificado falha. Deixe **"DNS only" (nuvem cinza)** até
o Caddy emitir (segundos), depois pode religar o proxy em modo **Full (strict)**.

Verifique:
```sh
curl -I https://isabelpontes.com.br            # 200
curl -s https://isabelpontes.com.br/robots.txt # aponta o sitemap
```
E cole o link no WhatsApp / no [debugger do Facebook](https://developers.facebook.com/tools/debug/):
o card de preview tem que vir com imagem e título.

### 9. Atualizações (redeploy)

```powershell
.\make.ps1 deploy   # rebuilda + scp + restart
```
Trocaram as fotos? Rode o `migrate_media.py`, envie o novo `manifest.json`, e na
VM `sudo -u site ... /opt/site/site sync` + `sudo systemctl restart site`.

---

## Backup (banco + anexos dos briefings)

**Por que não é só Litestream.** O Litestream replica o SQLite continuamente,
mas não cobre **arquivo** — e as imagens de referência que o cliente anexa ao
briefing vivem em `data/uploads/`. Um backup só do banco deixaria a Isa com o
texto do briefing e sem as referências visuais. Este backup cobre os dois.

**RPO:** até 24 h de perda no pior caso. Para alguns briefings por semana é a
troca certa. Se o volume crescer, o Litestream entra **por cima** disto (RPO de
segundos para o banco), não no lugar.

### 1. Bucket dedicado, privado

Crie um bucket **novo** no R2 (ex.: `isabel-site-backup`) e **não** habilite
acesso público nele. Ele guarda dados pessoais de terceiros — nome, contato,
motivação e fotos de clientes. Nunca reaproveite o bucket da galeria, que é
público por definição.

**Lifecycle rule** (é ela que faz a retenção, não o script): no bucket →
*Settings* → *Object lifecycle rules* → prefixo `db/`, expirar em **30 dias**.
Deixe `uploads/` de fora — anexo não expira.

### 1b. A chave: escopo, onde mora, o que nunca fazer

O token é gerado em **R2 → Manage API Tokens → Create API Token**:

| Campo | Valor | Por quê |
|---|---|---|
| Permissão | **Object Read & Write** | Admin daria poder sobre a conta R2 inteira |
| Bucket | **apenas** `isabel-site-backup` | Um vazamento não pode alcançar o bucket da galeria |
| TTL | sem expiração (ou anote a data) | Token expirado = backup falha em silêncio |

> **Sobre exclusão:** os presets do R2 juntam escrita e exclusão — não dá para
> emitir um token append-only pelo painel. Por isso a retenção é lifecycle rule:
> o script nunca chama delete, então um bug nele não apaga histórico. Se quiser
> proteção contra a *chave* ser usada para apagar, ligue **object versioning**
> no bucket — aí um delete só cria um marcador e a versão anterior sobrevive.

**Onde a chave mora — e onde não pode morar:**

| Lugar | Vale? |
|---|---|
| `~/.config/rclone/rclone.conf`, na VM, modo `600` | ✅ é o certo |
| `-ldflags` do `make.ps1 deploy` | ❌ **nunca** — vira texto legível dentro do binário (`strings site` mostra) |
| `.env.example` | ❌ é versionado; é template vazio, por definição |
| Qualquer arquivo do repositório | ❌ ver o aviso no topo do `.env.example` |

O `rclone config` já grava com permissão restrita, mas confirme:

```sh
chmod 600 ~/.config/rclone/rclone.conf
ls -l ~/.config/rclone/rclone.conf   # deve ser -rw------- e do usuário joao
```

**Se uma chave vazar:** revogue no painel do R2 *antes* de qualquer outra coisa
— rotacionar depois não ajuda, porque a chave antiga continua válida até ser
revogada. Depois gere outra e rode `rclone config` de novo. O bucket em si não
precisa ser recriado (ele nunca foi público).

⚠️ Este repositório já tem histórico disso: as chaves R2 do projeto React estão
no histórico do git e continuam recuperáveis. **Antes do primeiro commit** de
qualquer arquivo de credencial novo, rode `git check-ignore -v <arquivo>`.

### 2. rclone na VM

```sh
sudo -v ; curl https://rclone.org/install.sh | sudo bash
rclone config
```
No wizard: `n` (novo) → nome **`r2backup`** → tipo **`s3`** → provider
**`Cloudflare`** → access key / secret do token → region `auto` → endpoint
`https://<ACCOUNT_ID>.r2.cloudflarestorage.com`.

Confira que ele enxerga o bucket:
```sh
rclone lsd r2backup:
```

### 3. Instalar o script e o timer

```sh
cp deploy/backup.sh ~/apps/site-test/backup.sh
chmod +x ~/apps/site-test/backup.sh
sudo cp deploy/backup.service deploy/backup.timer /etc/systemd/system/
sudo systemctl daemon-reload
sudo systemctl enable --now backup.timer
```

Habilite o **`.timer`**, não o `.service` — o service é o que o timer dispara.

Rode uma vez à mão para validar antes de confiar nele:
```sh
sudo systemctl start backup.service
journalctl -u backup.service -n 30 --no-pager
rclone ls r2backup:isabel-site-backup
```

Ver quando é o próximo disparo: `systemctl list-timers backup.timer`.

### 4. Restaurar

> Um backup que nunca foi restaurado não é um backup. Faça este teste **uma
> vez**, agora, com o site no ar — a restauração é para um diretório separado,
> então não encosta em produção.

```sh
# banco (pegue o snapshot mais recente)
rclone ls r2backup:isabel-site-backup/db
rclone copy r2backup:isabel-site-backup/db/site-AAAA-MM-DD.db /tmp/restore/

# confira que o arquivo é um SQLite íntegro ANTES de promovê-lo
cd ~/apps/site-test && ./site snapshot /tmp/restore/teste.db && echo "banco OK"

# anexos
rclone copy r2backup:isabel-site-backup/uploads /tmp/restore/uploads
```

Para promover de verdade (desastre real):
```sh
sudo systemctl stop site
cp /tmp/restore/site-AAAA-MM-DD.db ~/apps/site-test/data/site.db
rm -f ~/apps/site-test/data/site.db-wal ~/apps/site-test/data/site.db-shm
cp -r /tmp/restore/uploads ~/apps/site-test/data/
sudo systemctl start site
```

Apagar o `-wal`/`-shm` é obrigatório: eles pertencem ao banco **antigo**, e
deixá-los ao lado de um site.db restaurado corrompe o resultado.

### Decisões embutidas no script

| Escolha | Por quê |
|---|---|
| `site snapshot` (VACUUM INTO) em vez de `cp site.db` | Em modo WAL as escritas recentes estão no `-wal`; `cp` gera arquivo truncado **sem erro nenhum** |
| `rclone copy`, nunca `sync`, nos uploads | `sync` apaga no destino o que sumiu na origem — um disco corrompido propagaria a perda para o backup |
| Snapshot vai para `/tmp` e é apagado no fim | Não deixa uma segunda cópia dos dados pessoais parada no disco |
| Retenção só nos snapshots do banco | Anexo nunca é apagado: ele é a referência que o cliente mandou uma vez só |

---

## Parte B — Fly.io (fallback)

Com o `Dockerfile`, `deploy/docker-entrypoint.sh` e `fly.toml` já prontos:

```sh
fly launch --no-deploy         # ou `fly apps create isabel-pontes`
fly volumes create data -r gru -s 1
fly secrets set R2_ACCOUNT_ID=... R2_ACCESS_KEY_ID=... R2_SECRET_ACCESS_KEY=...
fly deploy
```
O entrypoint restaura o banco do R2 se o volume sumir, popula do manifest no
primeiro boot, e serve com replicação contínua. O manifest vai embutido na
imagem (`COPY data/manifest.json`), então ele precisa existir no contexto de
build (rode `python scripts/migrate_media.py` antes, ou copie o `out/manifest.json`
para `data/`).

---

## Troubleshooting

| Sintoma | Causa provável |
|---|---|
| Site não responde na 80/443 | Firewall **local** da VM (passo 2b), não só a VCN |
| `fly`/Oracle: "Out of host capacity" | Ampere A1 esgotado no free tier — retry (seu bot) ou use o AMD micro |
| Caddy não emite certificado | Domínio proxiado na Cloudflare — deixe DNS-only até emitir |
| Galeria vazia ("Nenhuma foto") | Faltou `site sync` (passo 6) |
| Backup do banco baixável publicamente | Bucket de backup não é privado — use um bucket separado |
| Oracle apagou a VM ociosa | Free tier recupera instância parada — o Litestream→R2 é justamente o seguro; `litestream restore` traz o banco de volta |
