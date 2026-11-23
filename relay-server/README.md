# CordCode Relay Server

`relay-server` 是 CordCode 的独立公网密文路由服务（Go 编写，无 CGO，SQLite WAL 持久化）。它不启动 coding agent、不访问 Mac 本地文件，也无法解密 iOS ↔ Mac 之间的端到端 HPKE 业务密文。

## 快速入口

- **下载预编译安装包（Linux amd64 / arm64 + `.sha256`）**：[GitHub Releases](https://github.com/openAgi2/cordcode-link/releases/latest)
- **VPS 自建部署手册（有域名 / 纯公网 IP 二选一）**：[INSTALL.md](INSTALL.md)
- **在 Mac 客户端接入**：打开 **CordCode Link** →「连接状态」→「Relay」→「中继服务器」→「更改…」→「自定义中继」，填入 `wss://<你的域名或公网IP>`。

### 让 Claude Code / Codex 直接帮你部署

[INSTALL.md](INSTALL.md) 按 coding agent 逐步验收门格式编写，你可以直接把下面这段话发给 Claude Code、Codex 或其他编程代理：

```text
请严格按以下部署手册，帮我在我的 Linux VPS 上自建 CordCode Relay：
https://raw.githubusercontent.com/openAgi2/cordcode-link/main/relay-server/INSTALL.md

请先向我确认部署模式（有域名还是纯公网 IP）、域名/公网 IP 以及 VPS SSH 信息，再严格按手册中的验收门逐步执行，不要跳过检查或降级为自签名证书。
```

## 核心能力与安全边界

- **`POST /v1/activations/routes`**：MacBridge 使用本机 Ed25519 激活身份自动创建或幂等恢复 route，普通用户无需配置部署级密钥。
- **`POST /v1/routes/register`**：保留给运维验收的部署级 provisioning token 入口（服务端只保存 SHA-256 摘要，不存明文）。
- **`POST /v1/routes/{route}/devices/register`**：Mac 批准配对后登记已信任设备。
- **`GET /v1/routes/{route}/bridge` 与 `GET /v1/routes/{route}/devices/{device}`**：在线 WebSocket 密文透传。
- **`GET .../mailbox` 与 `POST .../mailbox/ack`**：仅对显式标记 `"mailbox:"` key epoch 的离线里程碑密文做有界持久化（默认 TTL `24h`、单设备上限 `50 MiB`）。
- **`POST .../revoke`**：即时撤销设备连接并清空其未确认离线信箱。

## 关键运行参数

所有参数同时支持 CLI flag 与环境变量（flag 优先，详见 `cmd/relay-server/main.go`）：

| 环境变量 | CLI Flag | 默认值 | 说明 |
| --- | --- | --- | --- |
| `RELAY_LISTEN_ADDR` | `-listen` | `127.0.0.1:8780` | 监听地址（生产环境仅绑定回环，由 Nginx 提供 TLS） |
| `RELAY_DB_PATH` | `-db` | `/var/lib/cordcode-relay/relay.db` | SQLite WAL 数据库路径 |
| `RELAY_PUBLIC_ENDPOINT` | `-public-endpoint` | `ws://127.0.0.1:8780` | 对外通告的根地址（部署必填 `wss://<域名或IP>`） |
| `RELAY_PROVISION_TOKEN_SHA256` | `-provision-token-sha256` | *(必填)* | 运维验收 token 的 64 位 hex SHA-256，缺省拒绝启动 |
| `RELAY_MAILBOX_TTL` | `-mailbox-ttl` | `24h` | 离线信箱过期时间 |
| `RELAY_MAX_MAILBOX_BYTES` | `-max-mailbox-bytes` | `52428800` (50 MiB) | 单设备未确认离线信箱上限 |
| `RELAY_MAX_FRAME_BYTES` | `-max-frame-bytes` | `33554432` (32 MiB) | 单条 WebSocket 信封上限（勿压低，避免大历史帧触发 close 1009） |
| `RELAY_RATE_LIMIT_PER_MINUTE` | `-rate-limit-per-minute` | `120` | 单 IP 每分钟路由注册限流（[INSTALL.md](INSTALL.md) 建议收紧至 `30`） |
| `RELAY_ACTIVATION_RATE_LIMIT_PER_MINUTE` | `-activation-rate-limit-per-minute` | `6` | 单 IP 每分钟自助激活限流 |
| `RELAY_TRUSTED_PROXY_CIDRS` | `-trusted-proxy-cidrs` | `127.0.0.0/8,::1/128` | 允许提供 `CF-Connecting-IP` 的反代网段 |

## 从源码测试与编译

本目录是独立 Go module（`module cordcode-relay`，要求 Go 1.25+）：

```bash
go test ./... -count=1 -race

# 交叉编译 Linux 静态二进制（amd64 / arm64）
CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -trimpath -ldflags='-s -w' -o cordcode-relay-linux-amd64 ./cmd/relay-server
CGO_ENABLED=0 GOOS=linux GOARCH=arm64 go build -trimpath -ldflags='-s -w' -o cordcode-relay-linux-arm64 ./cmd/relay-server
```

部署完成后，可使用仅保存在当前进程环境变量中的 provisioning token 执行真实公网 API/WebSocket 验收：

```bash
RELAY_LIVE_BASE_URL="https://<你的域名或公网IP>" \
RELAY_LIVE_PROVISION_TOKEN="<token>" \
go test ./internal/relay -run TestLiveRelayDeployment -count=1
```

需要额外验证 VPS 重启恢复时，指定 `RELAY_LIVE_RESTART_GATE=/tmp/relay-restart.ready`，测试会在密文入库后暂停等待服务重启并删除 gate 文件后再校验 fetch/ack。

## License

本目录（`relay-server/`）源码基于 **AGPL-3.0-only** 开源（见仓库根目录 `LICENSE`）。
