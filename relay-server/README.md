# CordCode Relay Server（自建加密中继）

**这是什么？**

你可以把 Relay（中继）理解成一个**「只管传密信的邮局」**：当你带着 iPhone 出门、和家里的 Mac 不在同一个 Wi-Fi 时，由它帮两边把加密数据转发过去。因为内容全程端到端加密（HPKE），**服务器本身看不到你的聊天记录、代码或文件**。

> **免折腾提示**：CordCode Link 官方已经默认自带免费加密中继，**绝大多数用户不需要自己部署**。只有当你希望使用自己的 Linux 服务器、审计中继源码或自选线路时，才需要继续往下看。

---

## 推荐方式：让 AI 帮你自动部署（不懂 Linux 也能搞定）

完整的部署手册 [INSTALL.md](INSTALL.md) 是专门按 AI 代理一步步检查验收的规范写的。你不需要自己看懂那些 Linux 命令，只要做好下面 3 件准备，剩下的交给 AI 跑就行：

### 你只需要自己准备这 3 件事
1. **一台 Linux 云服务器（VPS）**：阿里云、腾讯云、Vultr、DigitalOcean 等都可以，最便宜的 512MB 内存（Debian 11+ / Ubuntu 20.04+）就够用；记下它的**公网 IP** 和 **`root` 登录方式**。
2. **去云服务商网页后台放行端口**：在云厂商控制台的「安全组 / 防火墙」里，允许外部访问 **TCP 80 和 443 端口**（这一步需要你在网页后台手动点一下，AI 替不了你）。
3. **（可选）一个自己的域名**：把域名 `A` 记录解析到这台服务器的公网 IP（最推荐、最稳定）。**没有域名也没关系**，直接用纯公网 IP 也能部署。

### 把这段话直接复制给你的 AI 工具

> 请把下面这段话发给**电脑上能执行命令的 AI 工具**（例如 **Claude Code、Codex、Antigravity**），**不要**发给网页版 ChatGPT 或 Claude.ai（网页聊天框连不上你的服务器）。

```text
请严格按以下部署手册，帮我在我的 Linux VPS 上自建 CordCode Relay：
https://raw.githubusercontent.com/openAgi2/cordcode-link/main/relay-server/INSTALL.md

请先向我确认部署模式（有域名还是纯公网 IP）、域名/公网 IP 以及 VPS SSH 信息，再严格按手册中的验收门逐步执行，不要跳过检查或降级为自签名证书。
```

### 部署完成后，填回 CordCode Link（最后一步）

等 AI 告诉你部署并验收通过、给你一个 `wss://<你的域名或公网IP>` 地址后：
1. 在 Mac 上打开 **CordCode Link** →「连接状态」→「Relay」→「中继服务器」→「更改…」→ 切到 **「自定义中继」**。
2. 把 `wss://<你的域名或公网IP>` 粘贴进去并点击保存，等待状态变成 **「已接入中继网 / 已连接」**。
3. 切换新服务器后，在主界面点击 **「扫码配对新设备」** 显示 **iOS 二维码**，用手机重新扫一次并在 Mac 上点同意即可。

遇到问题欢迎进 [Telegram 交流群](https://t.me/+4wBkquLet2JhMjZk) 或在 [GitHub Issues](https://github.com/openAgi2/cordcode-link/issues) 提问。

---

## 手动部署与安装包下载

- **完整部署手册（支持有域名 / 纯公网 IP）**：[INSTALL.md](INSTALL.md)
- **预编译 Linux 二进制安装包（`amd64` / `arm64` + `.sha256`）**：[GitHub Releases](https://github.com/openAgi2/cordcode-link/releases/latest)

---

## 以下内容仅供开发者与运维参考（普通用户无需阅读）

`relay-server/` 是独立 Go module（`module cordcode-relay`，Go 1.25+，零 CGO，SQLite WAL 持久化），遵循 **AGPL-3.0-only** 开源协议（见仓库根目录 `LICENSE`）。

<details>
<summary>核心接口与安全机制</summary>

- `POST /v1/activations/routes`：CordCode Link 使用本机 Ed25519 激活身份自动创建或幂等恢复 route，普通用户无需配置部署级密钥。
- `POST /v1/routes/register`：保留给运维验收的部署级 provisioning token 入口（服务端只保存 SHA-256 摘要，不存明文）。
- `POST /v1/routes/{route}/devices/register`：Mac 批准配对后登记已信任设备。
- `GET /v1/routes/{route}/bridge` 与 `GET /v1/routes/{route}/devices/{device}`：在线 WebSocket 密文透传。
- `GET .../mailbox` 与 `POST .../mailbox/ack`：仅对显式标记 `"mailbox:"` key epoch 的离线里程碑密文做有界持久化（默认 TTL `24h`、单设备上限 `50 MiB`）。
- `POST .../revoke`：即时撤销设备连接并清空其未确认离线信箱。

</details>

<details>
<summary>启动参数与环境变量表（详见 cmd/relay-server/main.go）</summary>

所有参数同时支持 CLI flag 与环境变量（flag 优先）：

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

</details>

<details>
<summary>从源码测试、交叉编译与公网验收测试</summary>

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

需要额外验证 VPS 重启恢复时，额外指定 `RELAY_LIVE_RESTART_GATE=/tmp/relay-restart.ready`，测试会在密文入库后暂停等待服务重启并删除 gate 文件后再校验 fetch/ack。

</details>
