# 安全与漏洞反馈策略 / Security Policy

## 中文说明

请区分**日常使用问题**与**安全漏洞**：

- **日常安装求助、普通 Bug、功能建议**：欢迎直接发到 [GitHub Issues](https://github.com/openAgi2/cordcode-link/issues)，或加入 **[Telegram 交流群](https://t.me/+4wBkquLet2JhMjZk)** 反馈讨论。
- **高危安全漏洞（请私密上报）**：如果你发现可能危害用户安全的漏洞（例如配对鉴权绕过、越权访问其他路由、凭据泄露或可被滥用的中继缺陷），**请不要在公开 Issues 或公开群里直接贴出复现细节**，请通过本仓库顶部的 **Security → Report a vulnerability**（GitHub Private Vulnerability Reporting）私密提交，只有仓库维护者可见。

### 自建 Relay 安全提醒

如果你在自己的 VPS 上部署 `relay-server`（详见 [relay-server/INSTALL.md](relay-server/INSTALL.md)），请务必遵守：

- 安装前用 `.sha256` 核对二进制完整性；
- `relay-server` 只监听本机回环 `127.0.0.1:8780`，公网入口必须使用受信 CA 签发的有效 HTTPS/WSS 证书，不要用自签名证书或降级为 `ws://`；
- `/etc/cordcode-relay/relay.env` 只保存 provisioning token 的 SHA-256 摘要，并保持文件权限为 `0600`；
- 保持 `RELAY_TRUSTED_PROXY_CIDRS` 仅限本机回环（或明确验证过的代理网段），防止客户端伪造限速 IP。

---

## English Version

### Reporting Issues vs. Security Vulnerabilities

- **General questions, installation help, and non-security bugs**: please open a [GitHub Issue](https://github.com/openAgi2/cordcode-link/issues) or join the [Telegram community group](https://t.me/+4wBkquLet2JhMjZk).
- **Security vulnerabilities**: **do not open public GitHub issues or post exploit details in public chat groups** for vulnerabilities, credential leaks, pairing bypasses, relay abuse, or device trust failures. Please submit a private report via **Security → Report a vulnerability** (GitHub Private Vulnerability Reporting) in this repository.

### Scope & Self-Hosting Security Checklist

When deploying `relay-server` on your own VPS (see [relay-server/INSTALL.md](relay-server/INSTALL.md)):

- Verify the `.sha256` checksum of downloaded binaries before installing.
- Bind `relay-server` to loopback (`127.0.0.1:8780`) behind a TLS-terminating reverse proxy (`https://` / `wss://` with a trusted CA certificate; never bypass TLS verification or downgrade to `ws://`).
- Store only the SHA-256 digest of your provisioning token in `/etc/cordcode-relay/relay.env` with `0600` permissions.
- Keep `RELAY_TRUSTED_PROXY_CIDRS` restricted to loopback (or explicitly validated proxy networks) so clients cannot spoof `CF-Connecting-IP` rate-limit keys.
