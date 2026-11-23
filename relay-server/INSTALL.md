# CordCode Relay 自建部署手册（支持有域名 / 公网 IP 直连）

> 本手册面向「想在自己的 VPS 上部署 CordCode 中继」的用户，并按**「可直接交给任何 coding agent（Claude Code / Codex / Antigravity 等）自动执行」**的标准编写：所有命令、配置文件全文与验收门均自包含于本文，无需额外上下文。
>
> 覆盖两种标准形态（二选一，均采用标准 TCP 443 TLS + Nginx 反代 + systemd 常驻，**不部署任何网站**）：
> - **模式 A（有域名，推荐）**：`wss://<你的域名>`，Let's Encrypt 常规 90 天证书 + 自动续期（附可选 Cloudflare 代理要点）。
> - **模式 B（纯公网 IP，无域名）**：`wss://<你的公网IP>`，Let's Encrypt `shortlived` IP 证书 + 每日两次自动续期检查。

## 0. 执行 Agent 硬性约定（给拿到本文档的 agent）

1. **先锁定模式和变量再动手**：开工前只向用户确认 §1 的必要参数，先在 VPS shell 中导出 `DEPLOY_MODE`（`domain` 或 `ip`）与 `RELAY_HOST`（域名或 IPv4），后续命令全部引用 `${RELAY_HOST}`，严禁手工漏改或混跑两条分支。
2. **按顺序执行，每步先过验收门再进下一步**：任一门不通过 → 立即停止、保留现场、报告具体差异，禁止跳过或「先继续后面再说」。
3. **禁止任何形式的降级**：不用自签名证书、不加 `--insecure`、不关闭证书校验、不改成 `ws://` 明文。宁可失败上报，不可降级通过。
4. **秘密不出现在输出里**：provisioning token、证书私钥、API 凭据不打印、不写进任何会提交的文件；`relay.env` 只存 token 的 SHA-256 摘要。
5. **长命令必须有超时**：所有可能等待网络的命令（下载 / 签发 / 续期）都要包在 `timeout --signal=TERM --kill-after=5s <秒数>` 中；单条命令 5 分钟无有效进展即终止并排查，不无限等待。
6. **只做本手册内的事**：不重启或修改与部署无关的服务，不清理无关文件；遇到 80/443/8780 被未知服务占用（见步骤 1）→ 停止并向用户报告。
7. **完成后如实汇报**：列出通过的验收门与未执行项（尤其「步骤 8：真实 Mac/iPhone 接入」由用户人工完成，agent 不得宣称代替）。

## 1. 参数收集与分支决策（Agent 开工第一步）

执行前只需向用户确认以下 4 项（已提供则直接沿用，不重复追问）：

| 参数 | 变量名 | 示例 / 取值 | 说明 |
| --- | --- | --- | --- |
| 部署模式 | `DEPLOY_MODE` | `domain`（推荐）或 `ip` | 有可控域名选 `domain`；没有域名只用公网 IPv4 选 `ip` |
| 公网主机标识 | `RELAY_HOST` | `relay.example.com` 或 `203.0.113.10` | **纯主机名或纯 IPv4**，不带 `wss://`、端口或路径 |
| VPS 登录方式 | — | `ssh root@<vps>` | 需具备 root / sudo 权限 |
| 二进制获取方式 | — | GitHub Releases（推荐）或源码编译 | 预编译包位于 `https://github.com/openAgi2/cordcode-link/releases/latest` |

分支路由规则：
- 先按顺序执行 **§3 公共安装步骤（步骤 1–6）**；
- 到 **§4 TLS 与 Nginx 配置（步骤 7–11）**时：
  - 步骤 7：仅 `DEPLOY_MODE=domain` 额外先过 DNS 预检门，其余共用；
  - 步骤 9：**严格二选一**——`DEPLOY_MODE=domain` 只跑 **9A**，`DEPLOY_MODE=ip` 只跑 **9B**；
  - 步骤 8 / 10 / 11：两种模式共用；
- 最后按顺序执行 **§5 公网与真实接入验收（步骤 12–13）**。

## 2. 准备清单与部署形态

| 项 | 要求 |
| --- | --- |
| VPS | 公网 Linux，推荐 Debian 11+ / Ubuntu 20.04+（apt 系；其他发行版替换等价包名，验收门不变），root 可登录，内存 ≥ 512MB，磁盘 ≥ 5GB |
| 公网入口 | 模式 A：已将域名 `A` 记录指向 VPS 公网 IPv4；模式 B：静态公网 IPv4 |
| 安全组 / 防火墙 | 入方向放开 **TCP 22 / 80 / 443**（80 为 Let's Encrypt HTTP-01 签发与续期必需，443 为 Relay 入口） |
| `relay-server` 二进制 | `linux/amd64` 或 `linux/arm64`（与 VPS `uname -m` 一致），配对 `.sha256` 校验文件 |
| 一台 Mac | 已安装 CordCode Link（MacBridge），用于步骤 8 真实接入验收 |

```text
iPhone / Mac（CordCode 客户端）
    │ wss://${RELAY_HOST}（TCP 443，系统受信 CA TLS）
    ▼
nginx :443（TLS 终结 + WebSocket 反代）
    ▼ 127.0.0.1:8780
relay-server（仅监听回环：路由 / 离线密文信箱 / 设备撤销）
    └─ SQLite WAL：/var/lib/cordcode-relay/relay.db

nginx :80（仅 /.well-known/acme-challenge/，其余 404）→ 供 Let's Encrypt 签发与续期
systemd：cordcode-relay.service（常驻）+ cordcode-cert-renew.timer（自动续期）
```

Relay 转发的是 iOS ↔ Mac 的端到端 HPKE 密文，服务器看不到代码和消息明文；TLS 只保证传输层可信。本节点不部署网页客户端，`/` 与 `/web/` 返回 `404` 属**预期行为**。

## 3. 公共安装步骤（两种模式共用）

> 以下所有命令在 VPS 上以 root 执行。先在当前 shell 导出变量（替换为真实值）：
>
> ```bash
> export DEPLOY_MODE="domain"            # 有域名填 domain；纯 IP 填 ip
> export RELAY_HOST="relay.example.com"  # 域名或公网 IPv4，不带协议和斜杠
> ```

### 步骤 1 服务器盘点与前置检查

```bash
uname -m                                  # x86_64 -> linux-amd64；aarch64 -> linux-arm64
cat /etc/os-release | head -2             # 记录发行版
ss -lntp                                  # 检查 80 / 443 / 8780 是否已被占用
```

- 若 `80` / `443` / `8780` 已被**非本手册部署**的服务监听 → **停止并报告**（宿主机可能运行着其他网站或反代，须由用户决定，agent 不得静默顶替）。
- 记录当前已运行的重要服务（`systemctl list-units --state=running --type=service`），部署过程中不触碰它们。

### 步骤 2 安装系统依赖

```bash
apt-get update
apt-get install -y curl openssl ca-certificates nginx
nginx -v && curl --version | head -1 && openssl version
```

> Debian 11 注意：若 `apt-get update` 报 404（bullseye 已进归档），备份原源配置后把 security 源切换为 `https://archive.debian.org/debian-security/` 再重试；仍失败 → 停止报告。

### 步骤 3 创建专用用户与目录

```bash
id -u cccode-relay >/dev/null 2>&1 || \
  useradd --system --user-group --home-dir /var/lib/cordcode-relay --shell /usr/sbin/nologin cccode-relay
mkdir -p /opt/cordcode-relay/bin /opt/cordcode-relay/backups /etc/cordcode-relay/tls \
         /var/lib/cordcode-relay /var/lib/cordcode-acme/.well-known/acme-challenge
chown cccode-relay:cccode-relay /var/lib/cordcode-relay
chmod 0750 /var/lib/cordcode-relay
chmod 0755 /var/lib/cordcode-acme /var/lib/cordcode-acme/.well-known /var/lib/cordcode-acme/.well-known/acme-challenge
```

### 步骤 4 获取并安装 `relay-server` 二进制（SHA-256 校验，fail-closed）

按 VPS 架构二选一（`x86_64` → `amd64`，`aarch64` → `arm64`）：

#### 方式 4A：从 GitHub Releases 下载预编译包（推荐）

从 [openAgi2/cordcode-link Releases](https://github.com/openAgi2/cordcode-link/releases/latest) 获取目标版本 `<VERSION>`（如 `0.2.0`）的二进制与 `.sha256` 文件，或在 VPS 上直接下载并校验：

```bash
ARCH="$(uname -m)"
case "$ARCH" in
  x86_64)  RELAY_ARCH="amd64" ;;
  aarch64) RELAY_ARCH="arm64" ;;
  *) echo "Unsupported arch: $ARCH" >&2; exit 1 ;;
esac

# 将 <VERSION> 替换为 Releases 页面上的实际版本号（不带 v 前缀，例如 0.2.0）
VERSION="<VERSION>"
ASSET="cordcode-relay-${VERSION}-linux-${RELAY_ARCH}"
BASE_URL="https://github.com/openAgi2/cordcode-link/releases/download/v${VERSION}"

timeout --signal=TERM --kill-after=5s 120s curl -fsSL -o "/tmp/${ASSET}" "${BASE_URL}/${ASSET}"
timeout --signal=TERM --kill-after=5s 60s  curl -fsSL -o "/tmp/${ASSET}.sha256" "${BASE_URL}/${ASSET}.sha256"
(cd /tmp && sha256sum -c "${ASSET}.sha256")   # 验收门：输出 OK；失败立即停止
install -o root -g root -m 0755 "/tmp/${ASSET}" /opt/cordcode-relay/bin/relay-server
```

> 若 VPS 访问 GitHub 不稳定，可在你本机下载后用 `scp` 传至 VPS `/tmp/`，同样执行 `sha256sum -c` 再 `install`。

#### 方式 4B：从源码编译（备选，零 CGO 静态二进制）

若你希望直接审计源码并自行编译（需 Go 1.25+）：

```bash
git clone --depth 1 https://github.com/openAgi2/cordcode-link.git /tmp/cordcode-link-src
cd /tmp/cordcode-link-src/relay-server
go test ./... -count=1
CGO_ENABLED=0 GOOS=linux GOARCH="${RELAY_ARCH}" \
  go build -trimpath -ldflags='-s -w' -o /tmp/relay-server ./cmd/relay-server
sha256sum /tmp/relay-server
install -o root -g root -m 0755 /tmp/relay-server /opt/cordcode-relay/bin/relay-server
```

安装完成后统一验收：

```bash
file /opt/cordcode-relay/bin/relay-server   # 应显示 ELF 64-bit ... statically linked
/opt/cordcode-relay/bin/relay-server --help 2>&1 | head -5 || true
```

### 步骤 5 provisioning token 与环境文件

provisioning token 仅用于步骤 7 的运维 API 验收（日常 MacBridge 走 Ed25519 自助签名激活，不需要此 token）。**明文仅存一份在 `/root` 下，环境文件只保存 SHA-256 摘要**：

```bash
TOKEN="$(openssl rand -base64 48)"
umask 077
printf '%s\n' "$TOKEN" > /root/cordcode-relay-provision-token
DIGEST="$(printf '%s' "$TOKEN" | sha256sum | awk '{print $1}')"   # 注意 printf 不带换行
cat > /etc/cordcode-relay/relay.env <<EOF
RELAY_LISTEN_ADDR=127.0.0.1:8780
RELAY_DB_PATH=/var/lib/cordcode-relay/relay.db
RELAY_PUBLIC_ENDPOINT=wss://${RELAY_HOST}
RELAY_PROVISION_TOKEN_SHA256=${DIGEST}
RELAY_MAILBOX_TTL=24h
RELAY_MAX_MAILBOX_BYTES=52428800
RELAY_MAX_FRAME_BYTES=33554432
RELAY_RATE_LIMIT_PER_MINUTE=30
RELAY_ACTIVATION_RATE_LIMIT_PER_MINUTE=6
RELAY_TRUSTED_PROXY_CIDRS=127.0.0.0/8,::1/128
EOF
chmod 0600 /etc/cordcode-relay/relay.env
```

> 关键约束：
> - `RELAY_PUBLIC_ENDPOINT` 必须与将来在 MacBridge 中填写的自定义中继地址完全一致（根地址，不带路径、不带 `:443`、不带尾斜杠）。
> - `RELAY_MAX_FRAME_BYTES` 保持 `33554432`（32 MiB），**不要调低至 2 MiB**，否则长会话全量同步会触发 WebSocket close 1009 断连循环。

### 步骤 6 systemd 服务（全文照抄）

`/etc/systemd/system/cordcode-relay.service`：

```ini
[Unit]
Description=CordCode encrypted relay
After=network-online.target
Wants=network-online.target
StartLimitIntervalSec=60
StartLimitBurst=5

[Service]
Type=simple
User=cccode-relay
Group=cccode-relay
EnvironmentFile=/etc/cordcode-relay/relay.env
WorkingDirectory=/var/lib/cordcode-relay
ExecStart=/opt/cordcode-relay/bin/relay-server
Restart=on-failure
RestartSec=3
TimeoutStopSec=20
UMask=0027
NoNewPrivileges=true
PrivateTmp=true
ProtectHome=true
ProtectSystem=strict
ReadWritePaths=/var/lib/cordcode-relay
RestrictAddressFamilies=AF_UNIX AF_INET AF_INET6
LimitNOFILE=65536
StandardOutput=journal
StandardError=journal

[Install]
WantedBy=multi-user.target
```

```bash
systemd-analyze verify /etc/systemd/system/cordcode-relay.service   # 无 error 再继续
systemctl daemon-reload
systemctl enable --now cordcode-relay
curl --fail --retry 5 --retry-delay 1 --retry-connrefused -sS http://127.0.0.1:8780/healthz
# 验收门 ①：输出 {"status":"ok"}
curl --fail --retry 5 --retry-delay 1 --retry-connrefused -sS http://127.0.0.1:8780/readyz
# 验收门 ②：输出 {"status":"ready"}
ss -lntp | grep 8780     # 验收门 ③：仅监听 127.0.0.1:8780，且属 relay-server 进程
```

## 4. TLS 与 Nginx 配置（按 `DEPLOY_MODE` 执行）

### 步骤 7 DNS 预检（仅模式 A）与 nginx 80 端口（证书验证专用站）

**仅当 `DEPLOY_MODE=domain` 时先过 DNS 预检门**（纯 IP 模式跳过此命令）：

```bash
getent ahosts "${RELAY_HOST}" | awk '{print $1}' | sort -u
# 验收门（domain）：输出包含本 VPS 公网 IPv4（若已开启 Cloudflare 橙云代理，先切回灰云直连完成首次 HTTP-01 签发，或改按步骤 9A 末尾的 Cloudflare Origin CA 方式配置）
```

两种模式共用 `/etc/nginx/sites-available/cordcode-relay-acme`（把 `${RELAY_HOST}` 替换为实际域名或 IPv4）：

```nginx
server {
    listen 80 default_server;
    server_name ${RELAY_HOST};
    server_tokens off;
    access_log off;
    error_log /dev/null crit;
    location ^~ /.well-known/acme-challenge/ {
        root /var/lib/cordcode-acme;
        default_type text/plain;
        try_files $uri =404;
    }
    location / { return 404 ""; }
}
```

```bash
# 若系统自带默认站点占用 80：确认它只是发行版样例（/etc/nginx/sites-available/default）
# 才允许删除 sites-enabled/default 的软链，否则停止报告
ln -sf /etc/nginx/sites-available/cordcode-relay-acme /etc/nginx/sites-enabled/cordcode-relay-acme
nginx -t && systemctl reload nginx
printf 'install-check\n' > /var/lib/cordcode-acme/.well-known/acme-challenge/probe
curl -sS -H "Host: ${RELAY_HOST}" http://127.0.0.1/.well-known/acme-challenge/probe   # 验收门：输出 install-check
curl -sS -o /dev/null -w '%{http_code}\n' -H "Host: ${RELAY_HOST}" http://127.0.0.1/  # 验收门：404
rm /var/lib/cordcode-acme/.well-known/acme-challenge/probe
```

再从**你自己的电脑**（公网侧）验证 TCP 80 可达（确认云安全组已放开 80）：

```bash
curl -sS -o /dev/null -w '%{http_code}\n' "http://${RELAY_HOST}/"   # 验收门：404（能连上即可）
```

连不上 → 去云厂商控制台放行 TCP 80 入方向，再继续。**此门不过，后续 Let's Encrypt HTTP-01 签发必失败。**

### 步骤 8 安装 acme.sh（固定版本 + 哈希核验）与有界重试

两种模式统一使用固定版本 `acme.sh 3.1.6`，**先验哈希再使用，绝不 `curl | sh` 拉最新版**：

```bash
mkdir -p /opt/cordcode-acme
timeout --signal=TERM --kill-after=5s 90s \
  curl -fsSL --connect-timeout 15 --max-time 60 \
    -o /opt/cordcode-acme/acme.sh \
    https://raw.githubusercontent.com/acmesh-official/acme.sh/3.1.6/acme.sh
sha256sum /opt/cordcode-acme/acme.sh
# 验收门：c7d68b021cfd6380ea83a82962abde5b484779fee0b97d38681dfa1396bbc8d7
# 不一致 → 停止报告（上游变动需人工确认，不要自动换版本）
chmod 0755 /opt/cordcode-acme/acme.sh
bash /opt/cordcode-acme/acme.sh --version    # v3.1.6
```

给 ACME 流量配一个**专用的**有界超时 `curl` wrapper（防止弱网无限挂起；不改全局 curl、不绕过 CA 校验）：

```bash
mkdir -p /opt/cordcode-acme/bin
cat > /opt/cordcode-acme/bin/curl <<'EOF'
#!/bin/sh
exec /usr/bin/curl --connect-timeout 8 --max-time 20 --retry 2 --retry-delay 2 --retry-max-time 50 "$@"
EOF
chmod 0755 /opt/cordcode-acme/bin/curl
/opt/cordcode-acme/bin/curl -sS -o /dev/null -w '%{http_code}\n' https://acme-v02.api.letsencrypt.org/directory
# 验收门：200。若连续多次 000/超时 → VPS 到 LE 网络不通，停止报告（不要换 --insecure）
```

### 步骤 9 签发 TLS 证书（严格二选一）

> Agent 必须根据 `DEPLOY_MODE` **只执行 9A 或 9B 其中之一**，不可混跑。

#### 分支 9A：`DEPLOY_MODE=domain`（有域名 · 常规 90 天证书，推荐）

```bash
export PATH=/opt/cordcode-acme/bin:/usr/local/sbin:/usr/local/bin:/usr/sbin:/usr/bin:/sbin:/bin
timeout --signal=TERM --kill-after=8s 240s \
  bash /opt/cordcode-acme/acme.sh --issue \
    --server letsencrypt \
    --domain "${RELAY_HOST}" \
    --webroot /var/lib/cordcode-acme \
    --keylength ec-256 \
    --home /opt/cordcode-acme \
    --config-home /etc/cordcode-acme \
    --cert-home /etc/cordcode-acme/certs \
    --fullchain-file /etc/cordcode-relay/tls/fullchain.pem \
    --key-file /etc/cordcode-relay/tls/privkey.pem \
    --no-color
# 验收门 ①：日志出现 Cert success，且两个文件存在
openssl x509 -in /etc/cordcode-relay/tls/fullchain.pem -noout -checkhost "${RELAY_HOST}"   # 验收门 ②：does match
chmod 0600 /etc/cordcode-relay/tls/privkey.pem
```

> **可选：若域名前面使用 Cloudflare 代理（橙云）**
> 1. Cloudflare SSL/TLS 模式设为 **`Full (strict)`**，并确保 **WebSockets** 已开启；
> 2. 源站证书可使用上面的 Let's Encrypt 证书，也可改用 Cloudflare Origin CA 证书（放至 `/etc/cordcode-relay/tls/fullchain.pem` 与 `privkey.pem`，权限 `0600`；使用长期 Origin CA 证书时可跳过步骤 11 的自动续期 timer）；
> 3. 在 Nginx 中配置 Cloudflare 官方 IP 段的 `set_real_ip_from` + `real_ip_header CF-Connecting-IP;`，使 `$remote_addr` 还原为真实客户端 IP（防止源站直连伪造限速键），切勿放开任意来源的 `CF-Connecting-IP`。

#### 分支 9B：`DEPLOY_MODE=ip`（纯公网 IP · Let's Encrypt `shortlived` IP 证书）

```bash
export PATH=/opt/cordcode-acme/bin:/usr/local/sbin:/usr/local/bin:/usr/sbin:/usr/bin:/sbin:/bin
timeout --signal=TERM --kill-after=8s 240s \
  bash /opt/cordcode-acme/acme.sh --issue \
    --server letsencrypt \
    --cert-profile shortlived \
    --domain "${RELAY_HOST}" \
    --webroot /var/lib/cordcode-acme \
    --keylength ec-256 \
    --days 2 \
    --home /opt/cordcode-acme \
    --config-home /etc/cordcode-acme \
    --cert-home /etc/cordcode-acme/certs \
    --fullchain-file /etc/cordcode-relay/tls/fullchain.pem \
    --key-file /etc/cordcode-relay/tls/privkey.pem \
    --no-color
# 验收门 ①：日志出现 Cert success，且两个文件存在
openssl x509 -in /etc/cordcode-relay/tls/fullchain.pem -noout -checkip "${RELAY_HOST}"   # 验收门 ②：does match
chmod 0600 /etc/cordcode-relay/tls/privkey.pem
```

签发失败处理（真实遇过的场景）：

| 日志现象 | 原因 | 处理 |
|---|---|---|
| 一直在 Pending 后超时（HTTP-01 未过） | 公网 80 不可达或 DNS 尚未指向本机 | 回步骤 7 复查 DNS 与公网 80 自检 |
| `curl error 28`（下载阶段超时） | VPS → LE 网络间歇抖动 | order 往往已 valid：**不要立即换参数重签**。重跑一次相同 `--issue` 命令（acme.sh 会复用已完成验证），仍失败 → 保留 `--config-home` 下状态并报告 |
| rate limit / too many certificates | 短时间内重复签发 | 等待窗口恢复（IP 证书限速更严格），或改用模式 A（域名） |

**禁止的处理方式**：换自签名证书、加 `--insecure`、跳过哈希校验换 acme.sh 版本。

### 步骤 10 nginx 443 站（两种模式共用，全文照抄）

`/etc/nginx/sites-available/cordcode-relay`（将 `${RELAY_HOST}` 替换为实际域名或 IPv4）：

```nginx
map $http_upgrade $cordcode_connection_upgrade {
    default upgrade;
    '' close;
}
server {
    listen 443 ssl default_server;
    server_name ${RELAY_HOST};
    server_tokens off;
    ssl_certificate /etc/cordcode-relay/tls/fullchain.pem;
    ssl_certificate_key /etc/cordcode-relay/tls/privkey.pem;
    ssl_protocols TLSv1.2 TLSv1.3;
    ssl_session_cache shared:CordCodeTLS:10m;
    ssl_session_timeout 1d;
    access_log off;
    error_log /dev/null crit;
    client_max_body_size 1m;
    location = /healthz {
        proxy_pass http://127.0.0.1:8780;
        proxy_set_header Host $host;
        proxy_set_header CF-Connecting-IP $remote_addr;
    }
    location = /readyz {
        proxy_pass http://127.0.0.1:8780;
        proxy_set_header Host $host;
        proxy_set_header CF-Connecting-IP $remote_addr;
    }
    location /v1/ {
        proxy_pass http://127.0.0.1:8780;
        proxy_http_version 1.1;
        proxy_set_header Host $host;
        proxy_set_header CF-Connecting-IP $remote_addr;
        proxy_set_header Upgrade $http_upgrade;
        proxy_set_header Connection $cordcode_connection_upgrade;
        proxy_buffering off;
        proxy_read_timeout 3600s;
        proxy_send_timeout 3600s;
    }
    location / { return 404 ""; }
}
```

> 说明：
> - `CF-Connecting-IP` 是 `relay-server` 约定的「可信回环代理透传客户端 IP」头名，此处由 Nginx 用 `$remote_addr` 覆写，与是否使用 Cloudflare 无关。
> - `client_max_body_size 1m` 仅限制普通 HTTP 请求体；WebSocket 密文帧大小由 `RELAY_MAX_FRAME_BYTES=33554432`（32 MiB）控制。

```bash
ln -sf /etc/nginx/sites-available/cordcode-relay /etc/nginx/sites-enabled/cordcode-relay
nginx -t && systemctl reload nginx
```

### 步骤 11 证书自动续期（两种模式共用）

创建 nginx reload 钩子 `/etc/cordcode-acme/reload-nginx.sh`：

```bash
cat > /etc/cordcode-acme/reload-nginx.sh <<'EOF'
#!/bin/sh
set -eu
/usr/sbin/nginx -t
/bin/systemctl reload nginx
EOF
chmod 0750 /etc/cordcode-acme/reload-nginx.sh
```

把落盘路径与 reload 钩子注册进 acme.sh（以后每次续期自动更新文件并 reload nginx）：

```bash
export PATH=/opt/cordcode-acme/bin:/usr/local/sbin:/usr/local/bin:/usr/sbin:/usr/bin:/sbin:/bin
bash /opt/cordcode-acme/acme.sh --install-cert --domain "${RELAY_HOST}" --ecc \
  --home /opt/cordcode-acme --config-home /etc/cordcode-acme --cert-home /etc/cordcode-acme/certs \
  --fullchain-file /etc/cordcode-relay/tls/fullchain.pem \
  --key-file /etc/cordcode-relay/tls/privkey.pem \
  --reloadcmd /etc/cordcode-acme/reload-nginx.sh \
  --no-color
# 验收门：出现 Running reload cmd ... Reload successful
```

`/etc/systemd/system/cordcode-cert-renew.service`：

```ini
[Unit]
Description=Renew CordCode Relay TLS certificate
Wants=network-online.target
After=network-online.target nginx.service

[Service]
Type=oneshot
Environment=PATH=/opt/cordcode-acme/bin:/usr/local/sbin:/usr/local/bin:/usr/sbin:/usr/bin:/sbin:/bin
UMask=0077
ExecStart=/bin/bash /opt/cordcode-acme/acme.sh --cron --home /opt/cordcode-acme --config-home /etc/cordcode-acme --cert-home /etc/cordcode-acme/certs --server letsencrypt --no-color
TimeoutStartSec=240
TimeoutStopSec=10
KillMode=control-group
```

`/etc/systemd/system/cordcode-cert-renew.timer`（每日两次检查：未到续期窗口会毫秒级跳过，同时兼容域名 90 天证书与 IP 短期证书）：

```ini
[Unit]
Description=Check CordCode Relay certificate renewal twice daily

[Timer]
OnCalendar=*-*-* 02,14:17:00
RandomizedDelaySec=20m
Persistent=true

[Install]
WantedBy=timers.target
```

```bash
systemctl daemon-reload
systemctl enable --now cordcode-cert-renew.timer
systemctl start cordcode-cert-renew.service
systemctl show cordcode-cert-renew.service -p Result -p ExecMainStatus   # 验收门 ①：success / 0
grep Le_NextRenewTimeStr "/etc/cordcode-acme/certs/${RELAY_HOST}_ecc/${RELAY_HOST}.conf"
# 验收门 ②：存在且为未来时间（以该字段为准，勿按 --days 字面推断）
systemctl list-timers cordcode-cert-renew.timer                          # 验收门 ③：有下一次触发时间
```

> 首次部署**必须**强制续签一次以证明整条续期 + reload 链真实可用（只配 timer 不等于续期可用）：

```bash
export PATH=/opt/cordcode-acme/bin:/usr/local/sbin:/usr/local/bin:/usr/sbin:/usr/bin:/sbin:/bin
timeout --signal=TERM --kill-after=8s 240s \
  bash /opt/cordcode-acme/acme.sh --renew --domain "${RELAY_HOST}" --ecc --force \
  --server letsencrypt --home /opt/cordcode-acme --config-home /etc/cordcode-acme \
  --cert-home /etc/cordcode-acme/certs --no-color
# 验收门：Cert success + Running reload cmd + Reload successful
```

## 5. 公网与真实接入验收

### 步骤 12 公网验收（公网探测在**你自己的电脑**执行，API 验收在 VPS 执行）

在你自己的电脑上执行（把 `RELAY_HOST` 替换为你的域名或公网 IPv4）：

```bash
RELAY_HOST="relay.example.com"   # 或你的公网 IPv4，如 203.0.113.10

curl -fsS "https://${RELAY_HOST}/healthz"                                # 验收门：① {"status":"ok"}
curl -fsS "https://${RELAY_HOST}/readyz"                                 # 验收门：② {"status":"ready"}
curl -sS -o /dev/null -w '%{http_code}\n' "https://${RELAY_HOST}/"       # 验收门：③ 404（无网站）
curl -sS -o /dev/null -w '%{http_code}\n' "https://${RELAY_HOST}/web/"   # 验收门：④ 404（无网页客户端）

# 证书校验（域名模式传 -servername，且按模式二选一检查 SAN）
echo | openssl s_client -connect "${RELAY_HOST}:443" -servername "${RELAY_HOST}" 2>/dev/null \
  | openssl x509 -noout -subject -issuer -dates -ext subjectAltName
# 验收门 ⑤：issuer 为受信公共 CA（如 Let's Encrypt）；notAfter 在未来；
#   - 模式 A（domain）：SAN 含 DNS:<RELAY_HOST>
#   - 模式 B（ip）：SAN 含 IP Address:<RELAY_HOST>
```

API 全链路验收（在 VPS 上执行，使用步骤 5 生成的 provisioning token；凭据只留在 shell 变量里，不回显）：

```bash
TOKEN="$(cat /root/cordcode-relay-provision-token)"
ROUTE_JSON="$(curl -fsS -X POST "https://${RELAY_HOST}/v1/routes/register" \
  -H "Authorization: Bearer ${TOKEN}")"
ROUTE_ID="$(printf '%s' "${ROUTE_JSON}" | python3 -c 'import sys,json;print(json.load(sys.stdin)["routeId"])')"
BRIDGE_AUTH="$(printf '%s' "${ROUTE_JSON}" | python3 -c 'import sys,json;print(json.load(sys.stdin)["bridgeAuth"])')"
test -n "${ROUTE_ID}" && test -n "${BRIDGE_AUTH}" && echo "gate1: route register OK"

DEV_JSON="$(curl -fsS -X POST "https://${RELAY_HOST}/v1/routes/${ROUTE_ID}/devices/register" \
  -H "Authorization: Bearer ${BRIDGE_AUTH}" -H 'Content-Type: application/json' \
  -d '{"deviceId":"install-check"}')"
DEVICE_AUTH="$(printf '%s' "${DEV_JSON}" | python3 -c 'import sys,json;print(json.load(sys.stdin)["deviceAuth"])')"
test -n "${DEVICE_AUTH}" && echo "gate2: device register OK"

curl -fsS "https://${RELAY_HOST}/v1/routes/${ROUTE_ID}/devices/install-check/mailbox" \
  -H "Authorization: Bearer ${DEVICE_AUTH}"   # 验收门 ③：200，frames 为空数组
curl -fsS -X POST "https://${RELAY_HOST}/v1/routes/${ROUTE_ID}/devices/install-check/revoke" \
  -H "Authorization: Bearer ${BRIDGE_AUTH}"   # 验收门 ④：200（撤销验收设备，防遗留）
```

服务器侧验收通过判定（缺一不可）：

- 公网 TLS 正常 CA 校验 + 域名或 IP SAN 匹配；
- `/healthz` 返回 `ok`，`/readyz` 返回 `ready`；
- route 注册 / 设备注册 / 信箱 fetch / 撤销全部成功；
- 同期 `journalctl -u cordcode-relay --since "10 minutes ago"` 无 ERROR 级日志。

**尚未证明项**（由步骤 13 真实完成）：WebSocket 双向转发、离线信箱跨服务重启、真实 HPKE 端到端加密。若本机有 `relay-server/` 源码，也可额外运行：

```bash
RELAY_LIVE_BASE_URL="https://${RELAY_HOST}" \
RELAY_LIVE_PROVISION_TOKEN="<token>" \
go test ./internal/relay -run TestLiveRelayDeployment -count=1
```

只有预编译二进制的用户以步骤 13 的真实 App 连接作为最终验收。

### 步骤 13 真实接入（由用户在 Mac / iPhone 上人工完成）

1. Mac 打开 **CordCode Link** →「连接状态」→「Relay」→「中继服务器」→「更改…」→「自定义中继」。
2. 填入 `wss://${RELAY_HOST}`（不带端口、路径或尾斜杠），点击保存。
3. 等待 Relay 状态显示**已接入中继网 / 已连接**（「配置已应用」只代表配置保存完成，不是已连接证据）。
4. 切换新节点后旧节点的配对凭据不会自动迁移：打开 **iOS 二维码**，在 iPhone 的 CordCode App 扫码并在 Mac 上批准配对。
5. iPhone 断开同一 WiFi（切蜂窝网络）验证远程可用：能看到会话列表、发消息能收到流式回复。
6. 本节点未部署 `/web/`，**Web 二维码不可用**（预期行为，非故障）。

## 6. 日常维护与排障

```bash
systemctl status cordcode-relay --no-pager
journalctl -u cordcode-relay -n 100 --no-pager          # 业务日志（不含密钥和消息明文）
systemctl list-timers cordcode-cert-renew.timer         # 下次续期检查时间
grep Le_NextRenewTimeStr /etc/cordcode-acme/certs/*_ecc/*.conf
sha256sum /opt/cordcode-relay/bin/relay-server          # 当前运行的二进制哈希
sqlite3 'file:/var/lib/cordcode-relay/relay.db?mode=ro' 'PRAGMA quick_check;'
```

| 现象 | 已证明 | 下一跳 |
|---|---|---|
| `healthz` 公网不通但 VPS 本机 `127.0.0.1:8780` 返回 200 | relay 正常，公网 / 防火墙 / TLS 层问题 | 查云安全组 TCP 443、DNS 记录、`nginx -t`、证书是否过期（`openssl x509 -in /etc/cordcode-relay/tls/fullchain.pem -checkend 604800`） |
| 证书到期或 `does not match` | 续期链断了或主机名/IP 填错 | `systemctl start cordcode-cert-renew.service` 查日志；复查步骤 8 到 LE 连通性 |
| App 提示「配置已应用」但未连上 | 配置保存成功 ≠ 握手成功 | 看 Relay 真实状态与 `journalctl -u cordcode-relay` 中 `auth failed` / `close` 记录；核对 `RELAY_PUBLIC_ENDPOINT` 与 App 填写地址是否完全一致 |
| 纯 IP 模式续签频繁失败 | VPS → LE 网络抖动或 IP 证书限速较严 | 换用**模式 A（有域名）**或更换 VPS 区域，**严禁**降级为自签名证书 |

## 7. 回滚与清理

```bash
systemctl disable --now cordcode-relay cordcode-cert-renew.timer
rm -f /etc/nginx/sites-enabled/cordcode-relay /etc/nginx/sites-enabled/cordcode-relay-acme
nginx -t && systemctl reload nginx
# 数据与配置保留在 /var/lib/cordcode-relay /etc/cordcode-relay /etc/cordcode-acme；
# 确认不再需要时再手动删除。删除 relay.db 会使该节点上所有已配对设备失效，需重新配对。
```

## 8. 安全边界（这份部署能承诺什么）

- Relay 仅转发 iOS ↔ Mac 的 HPKE 端到端密文；服务器（包括拥有 root 权限的运维者）看不到代码与聊天明文，只能看到路由元数据、密文大小与时序。
- TLS 严格使用系统受信公共 CA，无任何证书校验豁免；私钥与 provisioning token 权限 `0600`，仅 root 可读。
- `relay-server` 仅监听 `127.0.0.1:8780`，以专用非登录系统用户沙箱运行（`ProtectSystem=strict`、`NoNewPrivileges=true`）。
- **威胁模型不包含「云厂商或中间网络掐断服务」**：这属于可用性问题而非加密问题，只能通过更换线路或 VPS 区域解决。
