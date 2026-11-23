# Privacy

**CordCode Link** (the macOS companion application distributed via this repository's Releases) bridges local coding-agent backends on your Mac to your paired CordCode mobile client. Neither CordCode Link nor `relay-server` includes analytics, advertising, or telemetry SDKs.

## Local Data on Your Mac

CordCode Link stores runtime state locally under `~/Library/Application Support/CordCode Link/` (sensitive credential files use `0600` file permissions):

- **Core bridge state**: device pairing records, trusted device list, local Management API token, backend selections, and rotating local runtime logs (`logs/go-bridge.log` plus `.1`/`.2`/`.3` rotations).
- **Relay credentials**: activation identity, route ID, and route credentials under `relay-secrets/`.
- **Direct TLS & Web Push state**: Tailscale / direct TLS certificate and SPKI pin material, plus Web Push VAPID keys and client subscriptions when enabled.
- **Per-backend local authentication state** (stored or read only for backends you enable): local CLI/desktop sessions or OAuth tokens required by the selected coding agent.

Coding backends access workspace files and execute tools strictly according to their own configuration and the permissions you grant.

## Network Data & Encrypted Relay

CordCode Link supports multiple transport paths depending on your configuration:

1. **Encrypted Relay (default, official or self-hosted `relay-server`)**: the Relay processes routing identifiers, connection metadata, frame sizes, and timing, but **cannot decrypt business payloads** — all Bridge RPCs, events, and offline mailbox frames are end-to-end encrypted with HPKE between paired devices. Offline mailbox storage holds only bounded encrypted milestone envelopes until delivery, acknowledgement, TTL expiry (default `24h`), or device revocation.
2. **LAN Direct WebSocket (optional)**: same-LAN direct connection (`ws://...:8777`) when enabled.
3. **Tailscale Direct TLS (optional)**: `wss://...:8778` authenticated with SPKI certificate pinning.

Each coding backend (Claude Code, Codex Desktop, Grok Build, DeepSeek Harness, OpenCode, Antigravity) communicates directly with its own model provider or cloud service according to the provider you chose.

## Your Control

You can revoke any paired device at any time in CordCode Link, switch to your own self-hosted `relay-server` from `relay-server/`, disable optional transport paths, or remove `~/Library/Application Support/CordCode Link/` when uninstalling.

Security issues should be reported privately according to [SECURITY.md](SECURITY.md).
