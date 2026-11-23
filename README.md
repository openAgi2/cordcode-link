# CordCode Link

**在 iPhone / iPad 上，随时远程使用你 Mac 里的 Claude Code、Codex、Antigravity 等 AI 助手。**

> **先看这条，少走弯路：**
> - **绝大多数用户**：只需按下面「第 1 步」安装 Mac 软件，配合手机端扫码即可直接用。**官方已默认提供免费加密网络连接，不需要自己买服务器**，不用管仓库里的 `relay-server/` 文件夹。
> - **进阶玩家（可选）**：如果你想用自己的 Linux 服务器转发加密连接，再看底部的「[5. 进阶可选：自建加密中继](#5-进阶可选自建加密中继普通用户可跳过)」。

---

## 1. 快速上手

### 开始前先确认 3 件事

1. **Mac 电脑要求**：macOS 14（Sonoma）及以上，仅支持 **M1 / M2 / M3 / M4 等苹果芯片（Apple Silicon）**，暂不支持 Intel 老款 Mac。（可点击屏幕左上角 `` →「关于本机」查看芯片与系统版本。）
2. **Mac 上已装好至少 1 个支持的 AI 工具**：CordCode Link 本身不替代这些 AI 工具，而是帮你把它们安全连到手机上。目前支持（装好并登录任意 1 个即可）：
   - **Claude Code**
   - **Codex Desktop**
   - **Grok Build**
   - **DeepSeek Harness**
   - **OpenCode**
   - **Antigravity**
3. **iPhone / iPad 端（CordCode App）**：**即将上线 App Store（敬请期待，链接待补充）**。上线后会第一时间更新在此处并在交流群通知。

### 安装 Mac 端 CordCode Link

1. **下载安装包**：前往 [GitHub Releases 最新版本页面](https://github.com/openAgi2/cordcode-link/releases/latest)，**只需要下载名字里带 `macos-arm64` 的 `.zip` 文件**（例如 `CordCodeLink-x.x.x-macos-arm64-unsigned.zip`）。
   > 页面里名字带 `linux` 或 `.sha256` 的文件，普通用户不用下载。
2. **解压并放进应用程序**：双击刚下载的 `.zip` 文件解压，把得到的 **`CordCodeLink.app`** 拖进 **「应用程序（Applications）」** 文件夹。
3. **首次打开放行（重要）**：
   目前提供的是早期预览版，暂时还没走苹果官方的付费公证流程，所以直接双击打开时，macOS 可能会弹出「无法验证开发者」或「已损坏，应移到废纸篓」（**不是病毒**）。
   按 `⌘ + 空格` 搜 **「终端」** 打开，复制下面这一整行粘进去并按回车，**只对这一个 App 放行**（不会降低电脑系统安全，请不要全局关闭 Gatekeeper）：

   ```bash
   xattr -cr /Applications/CordCodeLink.app && echo "已放行，现在可以直接双击打开了"
   ```

   看到终端打印出 **`已放行，现在可以直接双击打开了`** 即表示完成。
4. **启动并扫码配对**：
   在「应用程序」里双击打开 **CordCode Link**，勾选启用你 Mac 上已装好的 AI 工具，点击主界面里的 **「扫码配对新设备」** 显示二维码；再用手机端 CordCode 扫码并在 Mac 上点同意，即可在手机上远程使用。

<details>
<summary>可选：验证安装包完整性（SHA-256 校验，普通用户可跳过）</summary>

如果你同时下载了同名的 `.sha256` 文件，想确认安装包在下载过程中未被篡改，可在终端执行：

```bash
cd ~/Downloads
shasum -a 256 -c CordCodeLink-*-macos-arm64-unsigned.zip.sha256
```

如果终端末尾显示 **`OK`**，说明安装包完整无误。（如之前下载过多个旧版本，请把命令里的 `*` 替换为你刚下载的具体版本号。）

</details>

---

## 2. 隐私、安全与开源说明

很多朋友在看到「未签名」「客户端闭源」时会有顾虑，这里把核心原则直接讲透明：

- **全程端到端加密，官方也看不到你的内容**：手机和 Mac 之间的代码、对话与离线消息均采用端到端加密（HPKE），中继服务器只能看到无法解密的密文，无法窥探明文。
- **零广告、零统计埋点**：CordCode Link 不含任何广告、行为统计或第三方遥测 SDK。
- **不乱碰你的电脑文件**：CordCode Link 只负责传话，真正读取或修改项目的只是你在 Mac 上自己启用并授权的 AI 工具。
- **开源协议与闭源边界**：
  - 负责公网密文转发的 **`relay-server/` 服务端代码全部按 [AGPL-3.0-only](LICENSE) 开源**——欢迎任何人审计源码或自己搭建中继。
  - GitHub Releases 中分发的 **`CordCode Link` Mac 客户端（`CordCodeLink.app`）为免费提供的闭源软件**，不适用 AGPL-3.0 协议；客户端源码不在此仓库公开，未经授权不得反编译、二次打包或作为商业产品重新分发。

更详细的说明请参阅中英双语版 [隐私说明（PRIVACY.md）](PRIVACY.md) 与 [安全反馈策略（SECURITY.md）](SECURITY.md)。

---

## 3. 交流群、反馈与常见问题

- **日常求助 / 聊天 / 提建议**：欢迎加入 **[Telegram 交流群](https://t.me/+4wBkquLet2JhMjZk)**，或在 [GitHub Issues](https://github.com/openAgi2/cordcode-link/issues) 留言反馈。
- **安全漏洞私密上报**：请勿在公开群或公开 Issues 贴出漏洞利用细节，详见 [SECURITY.md](SECURITY.md)。

<details>
<summary>常见问题（点击展开）</summary>

- **Q：没有自己的服务器，或者手机出门用 4G/5G 流量能连家里/公司的 Mac 吗？**
  可以。CordCode 默认自带免费的加密中继，只要 Mac 开着且联网，手机出门在外也能直接连，不需要自己买服务器。
- **Q：打开 CordCode Link 提示「已损坏，应移到废纸篓」或「无法验证开发者」怎么办？**
  因为预览版暂未经过苹果付费公证。先确保已经把 `CordCodeLink.app` 拖进「应用程序」文件夹，再在终端执行上方第 3 步的 `xattr -cr` 放行命令即可。
- **Q：Intel 芯片的老款 Mac 能用吗？**
  暂不支持，当前只提供 M1 / M2 / M3 / M4 等 Apple Silicon 芯片版本。
- **Q：需要单独在 CordCode 买 AI 额度吗？**
  不需要。CordCode Link 本身免费，连接的是你 Mac 上原本已经安装并登录好的 AI 工具账号。

</details>

---

## 4. 卸载

- **普通卸载**：先退出 CordCode Link，然后把「应用程序」文件夹里的 `CordCodeLink` 拖进废纸篓即可。

<details>
<summary>彻底清空本地配对记录、中继凭据与运行日志（可选）</summary>

如果希望把本机保存的配对信息和配置一并清空，退出软件后在终端复制执行（请原样粘贴，不要改动引号内的路径）：

```bash
rm -rf "$HOME/Library/Application Support/CordCode Link"
defaults delete org.openagi.cordcode.link 2>/dev/null || true
```

</details>

---

## 5. 进阶可选：自建加密中继（普通用户可跳过）

**什么是中继（Relay）？**

你可以把它理解成一个**「只负责传密信的邮局」**：当你带着手机出门、和家里的 Mac 不在同一个 Wi-Fi 时，帮两边隔空把加密消息送到对方手里；因为信件全程加密，邮局本身拆不开信。

**官方已经默认配好了中继，绝大多数用户完全不需要自己搭。** 如果你出于隐私审计、内网偏好或想自选 VPS 线路，希望把中继搭在自己的 Linux 服务器上：

- **不懂 Linux 也能搭**：直接把部署手册交给电脑里的 **Claude Code / Codex** 等 AI 工具自动执行即可。
- **入口**：进入 [relay-server/](relay-server/) 目录（置顶有可直接复制给 AI 的提示词、预编译包下载和完整开源源码），或直接阅读 [VPS 自建部署手册（支持有域名 / 纯公网 IP）](relay-server/INSTALL.md)。
