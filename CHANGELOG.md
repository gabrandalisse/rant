# Changelog

All notable changes to this project will be documented in this file.

The format is based on [Keep a Changelog](https://keepachangelog.com/en/1.1.0/),
and this project adheres to [Semantic Versioning](https://semver.org/spec/v2.0.0.html).

## [1.0.0] - 2026-11-20

### Added
- **CLI Core (`rant`):** Record work frustrations via inline argument or securely through `$EDITOR` (`notepad` fallback on Windows, `vi` on Unix) with zero-wipe cleanup of temporary files.
- **Tag & Severity Metadata:** Support for automatic tag normalization (`#` stripping, lowercase, deduplication) and frustration severity ratings (`1` to `5`).
- **AES-256-GCM Encryption:** Authenticated encryption at rest using per-record random nonces stored in a local pure-Go SQLite vault (`~/.local/share/rant/vault.db`).
- **OS Keyring Integration:** Automatic generation and retrieval of the 256-bit master key via macOS Keychain, Windows Credential Manager, or Linux Secret Service.
- **Burnout Analytics (`rant retro`):** Terminal summary of total entries, average severity, peak weekday, peak hour, and top trigger tags over configurable day windows without exposing raw text.
- **Ephemeral Local Web Dashboard (`rant ui`):** Embedded single-file web UI served on `127.0.0.1` with an in-memory session token, 26-week activity heatmap, tag impact rankings, day/hour matrix, **Blur/Censor Mode**, and automatic server shutdown on tab close.
- **Vault Inspection & Panic Wipe:** `rant list` to decrypt and print entries in the terminal, `rant shell-init` to prevent shell history leaks, and `rant nuke` to overwrite the database file with zeros and remove the master key from the OS keyring.
- **Distribution & Docs:** Cross-platform GoReleaser workflow (`linux`, `darwin`, `windows` for `amd64`/`arm64`), one-line install scripts (`install.sh` and `install.ps1`), and GitHub Pages landing site.