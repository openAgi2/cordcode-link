# Security Policy

## Reporting a Vulnerability

Please report security issues privately to the repository owner (for example via GitHub Private Vulnerability Reporting if enabled, or by contacting the maintainer directly). **Do not open public GitHub issues** for security vulnerabilities, credential leaks, pairing bypasses, relay abuse, or device trust failures.

## Scope & Sensitive Areas

- `relay-server` route authentication, device registration, device revocation, rate limiting, and SQLite mailbox isolation.
- End-to-end HPKE encryption boundaries between CordCode Link and paired mobile clients.
- TLS certificate validation (self-hosted Relay deployments must use valid public CA certificates; never bypass TLS verification or downgrade to `ws://`).
- Local credential storage permissions (`0600`) for CordCode Link and self-hosted `relay.env` / TLS private keys.

## Self-Hosting Security Checklist

When deploying `relay-server` on your own VPS (see [relay-server/INSTALL.md](relay-server/INSTALL.md)):

- Verify the `.sha256` checksum of downloaded binaries before installing.
- Bind `relay-server` to loopback (`127.0.0.1:8780`) behind a TLS-terminating reverse proxy (`https://` / `wss://` with a trusted CA certificate).
- Store only the SHA-256 digest of your provisioning token in `/etc/cordcode-relay/relay.env` with `0600` permissions.
- Keep `RELAY_TRUSTED_PROXY_CIDRS` restricted to loopback (or explicitly validated proxy networks) so clients cannot spoof `CF-Connecting-IP` rate-limit keys.
