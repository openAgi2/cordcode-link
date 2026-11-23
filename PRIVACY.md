# 隐私说明 / Privacy Policy

## 中文核心说明

**CordCode Link**（通过本仓库 Releases 分发的 Mac 伴侣应用）和开源的 **`relay-server`** 均遵循最小数据原则：**不含任何广告、行为统计或第三方遥测 SDK**。

1. **端到端加密，中继看不到明文**：手机与 Mac 之间的消息、会话历史、代码与离线信箱内容全程采用 **HPKE 端到端加密**。无论是官方默认中继还是你自己搭建的 `relay-server`，都只能看到路由编号、数据包大小和时间等必要转发信息，**无法解密聊天和代码内容**。
2. **本地数据只留在你的 Mac 上**：配对设备列表、本机中继凭据与运行日志保存在 `~/Library/Application Support/CordCode Link/`（敏感凭据文件权限为 `0600`，仅当前 Mac 用户可读）。
3. **不越权访问你的项目文件**：CordCode Link 只负责在手机与 Mac 之间传递指令和结果；真正读取或修改工作区文件的，是你在 Mac 上自己启用并授权的 AI 工具（如 Claude Code、Codex Desktop、Grok Build、DeepSeek Harness、OpenCode、Antigravity），它们各自按你原有的设置和账号运行。
4. **完全由你掌控**：你可以随时在 CordCode Link 里撤销任意已配对的手机、改用自己搭建的 `relay-server`，或在卸载时彻底删除本地数据目录。

如发现隐私或安全问题，请参阅 [SECURITY.md](SECURITY.md)。

---

## English Version

**CordCode Link** (the macOS companion application distributed via this repository's Releases) bridges local coding-agent backends on your Mac to your paired CordCode mobile client. Neither CordCode Link nor `relay-server` includes analytics, advertising, or telemetry SDKs.

### Local Data on Your Mac

CordCode Link stores runtime state locally under `~/Library/Application Support/CordCode Link/` (sensitive credential files use `0600` file permissions):

- **Core bridge state**: device pairing records, trusted device list, local Management API token, backend selections, and rotating local runtime logs (`logs/go-bridge.log` plus `.1`/`.2`/`.3` rotations).
- **Relay credentials**: activation identity, route ID, and route credentials under `relay-secrets/`.
- **Direct TLS & Web Push state**: Tailscale / direct TLS certificate and SPKI pin material, plus Web Push VAPID keys and client subscriptions when enabled.
- **Per-backend local authentication state** (stored or read only for backends you enable): local CLI/desktop sessions or OAuth tokens required by the selected coding agent.

Coding backends access workspace files and execute tools strictly according to their own configuration and the permissions you grant.

### Network Data & Encrypted Relay

CordCode Link supports multiple transport paths depending on your configuration:

1. **Encrypted Relay (default, official or self-hosted `relay-server`)**: the Relay processes routing identifiers, connection metadata, frame sizes, and timing, but **cannot decrypt business payloads** — all Bridge RPCs, events, and offline mailbox frames are end-to-end encrypted with HPKE between paired devices. Offline mailbox storage holds only bounded encrypted milestone envelopes until delivery, acknowledgement, TTL expiry (default `24h`), or device revocation.
2. **LAN Direct WebSocket (optional)**: same-LAN direct connection (`ws://...:8777`) when enabled.
3. **Tailscale Direct TLS (optional)**: `wss://...:8778` authenticated with SPKI certificate pinning.

Each coding backend (Claude Code, Codex Desktop, Grok Build, DeepSeek Harness, OpenCode, Antigravity) communicates directly with its own model provider or cloud service according to the provider you chose.

### Your Control

You can revoke any paired device at any time in CordCode Link, switch to your own self-hosted `relay-server` from `relay-server/`, disable optional transport paths, or remove `~/Library/Application Support/CordCode Link/` when uninstalling.

Security issues should be reported privately according to [SECURITY.md](SECURITY.md).
