# Pawnix v0.2.0 — The Pawnix Rebrand 🐾

> The deep rename release. Everything that was FastClaw is now Pawnix — binary, paths, module path, service labels, scripts, JSON tags, even the launchd label. No backwards-compat shims; the new layout is the only layout.

---

## ⚠️ Breaking changes (this is a clean rename — no auto-migration)

If you used FastClaw v0.1.x, you must move your data manually before starting `pawnix`:

```bash
mv ~/.fastclaw                       ~/.pawnix
mv ~/.pawnix/fastclaw.json           ~/.pawnix/pawnix.json
mv ~/.pawnix/fastclaw.gateway.json   ~/.pawnix/pawnix.gateway.json
rm -f ~/.pawnix/fastclaw.pid         # stale PID, regenerated on startup
```

If you installed the launchd / systemd service:

```bash
# macOS — old service
launchctl unload ~/Library/LaunchAgents/ai.fastclaw.gateway.plist
rm        ~/Library/LaunchAgents/ai.fastclaw.gateway.plist
# Then re-install with the new binary
pawnix daemon install

# Linux — old service
sudo systemctl stop fastclaw-gateway.service
sudo systemctl disable fastclaw-gateway.service
sudo rm /etc/systemd/system/fastclaw-gateway.service
pawnix daemon install
```

If you used `fastclaw-manager.sh`, it is now `pawnix-manager.sh` — same commands, same flags, just renamed.

---

## What changed

### Identity

| Concept | Before (v0.1) | After (v0.2) |
|---|---|---|
| Binary name | `fastclaw` | **`pawnix`** |
| Config dir | `~/.fastclaw/` | **`~/.pawnix/`** |
| Main config file | `fastclaw.json` | **`pawnix.json`** |
| Daemon PID file | `fastclaw.pid` | **`pawnix.pid`** |
| Gateway info file | `fastclaw.gateway.json` | **`pawnix.gateway.json`** |
| Go module path | `github.com/fastclaw-ai/fastclaw` | **`github.com/softbreezee/claw-os`** |
| `cmd/` package | `cmd/fastclaw/` | **`cmd/pawnix/`** |
| launchd label | `ai.fastclaw.gateway` | **`ai.pawnix.gateway`** |
| systemd unit | `fastclaw-gateway.service` | **`pawnix-gateway.service`** |
| GitHub repo | `fastclaw-ai/fastclaw` | **`softbreezee/claw-os`** |
| Manager script | `fastclaw-manager.sh` | **`pawnix-manager.sh`** |
| Web fetch User-Agent | `FastClaw/1.0` | **`Pawnix/1.0`** |
| HTTP override headers | `x-fastclaw-agent-id`, `x-fastclaw-session-key` | **`x-pawnix-agent-id`, `x-pawnix-session-key`** |
| SkillMetadata JSON tag | `fastclaw` | **`pawnix`** *(falls back to `openclaw`)* |

### Visuals

- New paw-print-in-signal-ring logo at [`web/public/logo.svg`](web/public/logo.svg) and the matching favicon at [`web/src/app/icon.svg`](web/src/app/icon.svg) — wifi-style signal arcs over a deep navy rounded square.

### Documentation

- [README.md](README.md) fully rewritten for the new identity and updated 7-milestone roadmap (v0.1 Foundation · v0.2 Rebrand · v0.3 Memory OS · v0.4 Multi-modal & Voice · v0.5 Agent Marketplace · v0.6 Distributed Mesh · v1.0 Production Persona).
- [`pawnix-manager.sh`](pawnix-manager.sh) renamed from `fastclaw-manager.sh`.
- [`install.sh`](install.sh) now installs the `pawnix` binary from the `softbreezee/claw-os` repo, honouring `PAWNIX_INSTALL_DIR`.

---

## Highlights inherited from v0.1 Foundation

For completeness — these are the capabilities the runtime shipped in v0.1 that v0.2 carries over unchanged:

- Multi-LLM (OpenAI, Anthropic, DeepSeek, Gemini, GLM, Kimi, Groq, Ollama, OpenRouter) with per-call model override
- Multi-agent + team `@mention` routing
- Channels: Web, Telegram (multi-account), Discord, Slack, plus JSON-RPC plugin channels
- Channels CRUD UI with **Save & Restart** (uses exit code 75 to ride the daemon supervisor)
- Dual-layer memory (`MEMORY.md` + FTS / pgvector)
- Skill system + skill auto-induction primitives + plugins + MCP (HTTP / stdio)
- Cron jobs + per-agent heartbeat + webhooks
- Async chat tasks that survive page reloads (resumable SSE)
- Daemon supervisor with crash auto-restart + restart-aware exit codes
- Docker-sandboxed `exec` + YAML policy engine + AES-256-GCM credential vault + PII scrubbing

---

## Install

```bash
curl -fsSL https://raw.githubusercontent.com/softbreezee/claw-os/main/install.sh | bash
pawnix    # opens the setup wizard
```

From source:

```bash
git clone https://github.com/softbreezee/claw-os.git
cd claw-os && make build
./bin/pawnix
```

---

## 🙏 Acknowledgements

Pawnix is forked from and stands on the shoulders of [**fastclaw-ai/fastclaw**](https://github.com/fastclaw-ai/fastclaw). The entire v0.1 Foundation surface — multi-provider routing, channels, skills, plugins, MCP, daemon supervisor, web dashboard — all originated there. This rebrand is the surface change; the credit for the runtime belongs to FastClaw's authors. Thank you 🙏

---

🐾
