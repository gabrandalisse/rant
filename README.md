# >_ rant

**A local-first, encrypted CLI vault to vent work frustrations and measure burnout objectively.**

`rant` lets you offload daily workplace friction directly from your terminal without leaving traces in Slack, Jira, or third-party cloud services. Every entry is encrypted at rest with AES-256-GCM using a master key stored in your operating system's native keyring, paired with local observability tools to spot patterns before burnout hits.

**[Live Landing Page & Docs](https://gabrandalisse.github.io/rant/)**

---

## Features & Security Model

- **Zero-Cloud & Offline-First:** Runs as a single static binary with embedded pure-Go SQLite (`modernc.org/sqlite`, `CGO_ENABLED=0`). No external servers, no telemetry, no accounts.
- **Authenticated Encryption (AES-256-GCM):** Every entry is encrypted with a fresh cryptographic nonce before touching disk (`~/.local/share/rant/vault.db`).
- **OS-Native Key Management:** Your 256-bit master key never sits in a plaintext config file—it is generated and retrieved directly from **macOS Keychain**, **Windows Credential Manager**, or **Linux Secret Service**.
- **Objective Burnout Metrics (`rant retro`):** Aggregates severity scores, peak hours, critical weekdays, and recurring trigger tags directly from unencrypted metadata without exposing raw text.
- **Ephemeral Web Dashboard (`rant ui`):** Spins up a local-only HTTP server on a random free port secured with an in-memory session token. Features a 26-week heatmap, tag impact rankings, a day/hour matrix, and a **Blur/Censor Mode** for screen-sharing safety. Automatically shuts down when you close the browser tab.
- **Panic Wipe (`rant nuke`):** Overwrites the SQLite database file with zeros, deletes it from disk, and purges the master key from your OS keyring.

---

## Installation

### macOS & Linux
```sh
curl -fsSL [https://raw.githubusercontent.com/gabrandalisse/rant/main/install.sh](https://raw.githubusercontent.com/gabrandalisse/rant/main/install.sh) | sh