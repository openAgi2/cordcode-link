# CordCode Link & Relay Server

本仓库是 **CordCode Link**（macOS 伴侣应用安装包）与 **CordCode Relay Server**（开源端到端加密中继服务端）的官方公开仓库。

- **下载 CordCode Link（macOS）与 Relay 安装包**：[GitHub Releases](https://github.com/openAgi2/cordcode-link/releases/latest)
- **查看 Relay 开源源码与自建部署手册**：[relay-server/](relay-server/) · [VPS 自建部署手册（有域名 / 纯公网 IP）](relay-server/INSTALL.md)

---

## 1. 安装 CordCode Link（macOS）

CordCode Link 运行在你的 Mac 上，用于将本地已安装和登录的编程代理安全桥接到 iPhone / iPad 端的 CordCode App。

> **系统要求**：macOS 14.0+（Apple Silicon）。  
> **发布形态**：当前提供的是**未签名预览版（Unsigned Preview）**，尚未经过 Apple Developer ID 公证。

### 安装步骤

1. 前往 [GitHub Releases](https://github.com/openAgi2/cordcode-link/releases/latest)，下载最新的 `CordCodeLink-<版本>-macos-arm64-unsigned.zip` 和同名 `.sha256` 校验文件。
2. 校验压缩包完整性（与 `.sha256` 一致再继续）：

   ```bash
   shasum -a 256 -c CordCodeLink-*-macos-arm64-unsigned.zip.sha256
   ```

3. 解压压缩包，将 `CordCodeLink.app` 拖入 `/Applications` 目录。
4. 由于预览版未公证，macOS 会为下载的文件附加隔离属性（quarantine）。在终端执行以下命令**仅清除该应用自身的隔离属性**（无需全局关闭系统 Gatekeeper）：

   ```bash
   xattr -cr /Applications/CordCodeLink.app
   ```

   （也可以在 `/Applications` 中对 `CordCodeLink.app` 右键 → 选择「打开」→ 再次确认打开。）
5. 正常启动 **CordCode Link**，即可通过二维码与手机端配对。

### 支持的编程代理后端

在 Mac 上按需安装并登录以下任意后端后，即可在 CordCode Link 中一键启用：

- **Claude Code**
- **Codex Desktop**
- **Grok Build**
- **DeepSeek Harness**
- **OpenCode**
- **Antigravity**

### 卸载

退出 CordCode Link 并删除 `/Applications/CordCodeLink.app`。如需一并清理本机配对记录、中继凭据与运行日志，可执行：

```bash
rm -rf "$HOME/Library/Application Support/CordCode Link"
defaults delete org.openagi.cordcode.link 2>/dev/null || true
```

---

## 2. 自建加密中继（Relay Server）

CordCode 默认提供开箱即用的端到端加密中继；如果你希望使用自己的服务器、审计中继代码或优化线路，可以在自己的 Linux VPS 上自建 `relay-server`。

- **源码目录**：[relay-server/](relay-server/)（独立 Go module，零 CGO，SQLite WAL，可随时审计与自行编译）
- **预编译二进制**：[GitHub Releases](https://github.com/openAgi2/cordcode-link/releases/latest) 提供 `linux-amd64` 与 `linux-arm64` 静态二进制及 `.sha256`
- **VPS 自建部署手册**：[relay-server/INSTALL.md](relay-server/INSTALL.md)
  - 同时覆盖 **有域名（推荐）** 与 **纯公网 IP 直连（无域名）** 两种形态
  - 按 coding agent 验收门规范编写，可直接交给 **Claude Code / Codex** 等代理按步骤帮你部署

部署完成后，在 **CordCode Link** →「连接状态」→「Relay」→「中继服务器」→「更改…」→「自定义中继」填入 `wss://<你的域名或公网IP>` 并重新用 iOS 二维码配对即可切换。

---

## 3. 隐私与安全

- 所有业务消息、会话历史与离线信箱均采用 **HPKE 端到端加密**，Relay 服务器仅处理路由元数据与密文转发，无法查看明文内容。
- 详细隐私说明见 [PRIVACY.md](PRIVACY.md)，安全漏洞私密反馈方式见 [SECURITY.md](SECURITY.md)。

---

## 4. License（开源协议与分发边界）

- **`relay-server/` 目录下的所有源码**遵循 **[GNU Affero General Public License v3.0 (AGPL-3.0-only)](LICENSE)** 开源。
- **GitHub Releases 中发布的预编译 `CordCode Link` macOS 客户端安装包（`CordCodeLink.app`）为闭源免费分发软件**，不适用 AGPL-3.0 协议；客户端源码不在此仓库公开，未经授权不得反编译、二次打包或作为商业产品重新分发。
