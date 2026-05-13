<div align="center">

<img src="web/public/logo.svg" width="120" alt="Pawnix logo" />

# Pawnix

**A self-hosted, AI-Native personal OS — in a single Go binary.**

[![Go](https://img.shields.io/badge/Go-1.25+-00ADD8?style=flat&logo=go)](https://go.dev)
[![License](https://img.shields.io/badge/License-MIT-blue.svg)](LICENSE)
[![Status](https://img.shields.io/badge/status-v0.1%20foundation-f59e0b)](#-roadmap)

Multi-agent · Multi-channel · Long-term memory · Skills that learn · Plugins · MCP · Cron · Web dashboard

[Install](#-install) · [Quick Start](#-quick-start) · [Roadmap](#-roadmap) · [Architecture](#-architecture) · [Configuration](#-configuration)

</div>

---

## What is Pawnix?

Pawnix is **not just an agent runtime** — it's a long-running, self-hosted layer between you and any LLM, designed to behave the way an operating system does for your AI life:

- **Always on.** Daemon supervisor keeps the gateway alive across crashes and config changes.
- **Multi-agent.** Run a personal team — coder, analyst, scheduler — each with its own personality, memory, and skills, talking via @mentions.
- **Multi-channel by default.** Same agents, reachable from your browser, Telegram, Discord, Slack, or any custom plugin channel.
- **Persistent memory.** `MEMORY.md` + searchable conversation logs + pgvector semantic memory. Your agent actually remembers you.
- **Skills that grow.** Drop a `SKILL.md` and your agent can do new things; agents can also learn new skills from interaction patterns.
- **Local-first, cloud-optional.** Defaults to plain JSON files in `~/.fastclaw/`. Switch to PostgreSQL or SQLite when you outgrow files.
- **Single binary.** No Docker required, no Python venv, no Node runtime. Cross-compiles for macOS / Linux / Windows.

> Pawnix evolved from the FastClaw codebase. The binary is still `fastclaw` and the config dir is still `~/.fastclaw/` for v0.1 — see [Roadmap](#-roadmap) for the deep rename plan.

```bash
curl -fsSL https://raw.githubusercontent.com/fastclaw-ai/fastclaw/main/install.sh | bash
fastclaw    # Opens the setup wizard in your browser
```

---

## 🗺 Roadmap

Pawnix is built in five layers, each adding a capability that an "AI OS" needs.

### **v0.1 — Foundation** &nbsp;`← you are here`

> *The runtime. Every later milestone builds on this.*

- [x] Multi-LLM provider routing with per-call model override (OpenAI, Anthropic, DeepSeek, Gemini, GLM, Kimi, Groq, Ollama, OpenRouter)
- [x] Multi-agent + team `@mention` routing
- [x] Channels CRUD UI (Telegram / Discord / Slack) with hot agent bindings
- [x] Skills + Plugins (JSON-RPC subprocess) + MCP (HTTP & stdio)
- [x] Dual-layer memory (`MEMORY.md` + FTS / pgvector)
- [x] Cron jobs + per-agent heartbeat
- [x] Web dashboard at `:18953` for everything
- [x] Daemon supervisor with crash auto-restart and `restart-aware` exit codes

### **v0.2 — The Memory OS** &nbsp;*next*

> *Make the agent remember you, not just facts.*

- [ ] Memory graph: extract entities and relations from conversations
- [ ] Time-aware retrieval (recency-weighted scoring)
- [ ] Cross-agent shared memory pool with permissions
- [ ] Memory browser UI: visualize, edit, and prune what each agent knows
- [ ] Skill auto-induction: turn frequent prompt patterns into callable skills

### **v0.3 — Multi-modal & Voice**

> *Stop requiring a keyboard.*

- [ ] Voice in/out (Whisper STT + pluggable TTS)
- [ ] Screen understanding (screenshot → multimodal LLM → action)
- [ ] First-class file ingestion (PDF / video / Excel → memory)
- [ ] Native voice messages and image streams in Telegram channel

### **v0.4 — Agent Marketplace**

> *From "my agent" to "an ecosystem."*

- [ ] Skills Hub: a "GitHub for skills" with one-click install
- [ ] Agent export/import: full bundle (persona + skills + memory schema)
- [ ] Community catalog and rating
- [ ] Marketplaced plugins (Obsidian-style)

### **v0.5 — Distributed Mesh**

> *From "on my laptop" to "across all my devices."*

- [ ] Multi-device sync (laptop + phone + server share state)
- [ ] P2P session handover (start at home, finish on the road)
- [ ] End-to-end encrypted remote access — without renting a VPS

### **v1.0 — Production Persona**

> *An AI you can trust with important things.*

- [ ] Full audit log (every tool call, traceable & rollback-able)
- [ ] Permission system v2 (fine-grained, dangerous-op confirmation)
- [ ] Backup, disaster recovery, and high-availability mode
- [ ] Multi-user / team mode

---

## 📦 Install

### One-liner (macOS / Linux)

```bash
curl -fsSL https://raw.githubusercontent.com/fastclaw-ai/fastclaw/main/install.sh | bash
```

### Windows

Download the `.zip` from [Releases](https://github.com/fastclaw-ai/fastclaw/releases), extract, and double-click `fastclaw.exe`.

### From source

```bash
git clone https://github.com/fastclaw-ai/fastclaw.git
cd fastclaw && make build
```

### Quick environment setup (development)

```bash
git clone https://github.com/fastclaw-ai/fastclaw.git
cd fastclaw
./setup.sh          # checks Go, Node, PostgreSQL — installs only if you opt in
```

### Upgrade

```bash
fastclaw upgrade
```

---

## 🚀 Quick Start

1. Run `fastclaw` — your browser opens the setup wizard at `http://localhost:18953`.
2. Pick an LLM provider (OpenAI, Anthropic, or any OpenAI-compatible endpoint).
3. Click **Launch** — start chatting in the browser.
4. Open **Channels** in the sidebar, click **Add channel**, paste a Telegram bot token and bind it to an agent. Hit **Save & Restart**. Done.

---

## ✨ Features

### Core agent

| Feature | Description |
|---------|-------------|
| **ReAct loop** | Multi-turn reasoning + tool calling with configurable max iterations |
| **Any LLM** | OpenAI-compatible API (OpenAI, Anthropic, DeepSeek, Gemini, GLM, Kimi, Groq, Ollama, OpenRouter) |
| **Per-call model override** | Pick a different model on a per-message basis without restarting |
| **Multi-agent** | Independent personality, memory, skills per agent. Team support with `@mention` routing |
| **Context engineering** | Auto-pruning + LLM-driven compression for long conversations |
| **Dual-layer memory** | `MEMORY.md` (persistent facts) + FTS-searchable conversation logs |
| **Skill system** | On-demand `SKILL.md` loading. Agents can learn new skills from interaction patterns |
| **Hook system** | Before / After hooks on prompts, model calls, and tool calls |
| **Hot reload** | Edit config or `SOUL.md` → takes effect immediately, no restart |

### Storage

| Backend | Notes |
|---------|-------|
| **File** | Default. JSON in `~/.fastclaw/`. Zero setup. |
| **PostgreSQL** | Multi-tenant + `pgvector` HNSW for semantic memory |
| **SQLite** | Lightweight, with FTS5 full-text search |

### Channels

| Channel | Status |
|---------|--------|
| Web Chat (built-in async task system) | ✅ |
| Telegram (multi-account, multi-bot) | ✅ |
| Discord | ✅ |
| Slack (Socket Mode + Web API) | ✅ |
| Anything else | ✅ via JSON-RPC plugin |

### Built-in tools

`exec` (with optional Docker sandbox) · `read_file` / `write_file` / `list_dir` · `web_fetch` · `web_search` (Brave) · `memory_search` · `message` · `spawn_subagent` · `create_cron_job` / `list_cron_jobs` / `delete_cron_job` · `load_skill` · `db_query` / `db_create_table` · all MCP tools

### Automation

| Feature | Description |
|---------|-------------|
| **Cron** | Cron expressions, intervals, or one-shot timers |
| **Heartbeat** | Each agent wakes every N minutes to consult `HEARTBEAT.md` |
| **Webhooks** | `POST /hooks` triggers from external systems |
| **Distributed locks** | Safe cron in multi-instance deployments |

### Security

Sandbox exec · YAML policy engine (filesystem / network / tool / resource limits) · AES-256-GCM credential vault · optional PII scrubbing · tool-loop detection

### Platform

Web dashboard · OpenAI-compatible REST API (`/v1/chat/completions` SSE) · WebSocket gateway · MCP client (HTTP + stdio) · multi-tenant · daemon mode with auto-restart

---

## 🏗 Architecture

```
                   ┌────────────────────────────────────────────────────────────┐
                   │                       Pawnix Gateway                       │
                   │                                                            │
   Web UI ───────▶ │   ┌──────────┐               ┌─────────────────────────┐  │
   Channels ────▶ │   │ Message  │               │      Agent Manager      │  │
   Plugins ─────▶ │   │   Bus    │──────────────▶│  Agent A · Agent B · …  │  │
   Webhooks ────▶ │   │          │◀──────────────│  (skills · hooks · mem) │  │
   API ─────────▶ │   └──────────┘               └────────────┬────────────┘  │
                   │                                           │                │
                   │     ┌────────────┬───────────┬────────────┼──────────┐    │
                   │     ▼            ▼           ▼            ▼          ▼    │
                   │  ┌──────┐  ┌──────────┐  ┌──────────┐  ┌────────┐ ┌────┐  │
                   │  │Tools │  │  Memory  │  │ Sessions │  │ Policy │ │Cron│  │
                   │  └──────┘  └──────────┘  └──────────┘  └────────┘ └────┘  │
                   │                                                            │
                   │  ┌────────────────────────────────────────────────────┐    │
                   │  │  REST · SSE · WebSocket · Webhooks · Web UI :18953 │    │
                   │  └────────────────────────────────────────────────────┘    │
                   └────────────────────────────────────────────────────────────┘

                   ┌─────── daemon supervisor (crash + restart aware) ───────┐
```

The daemon supervisor (`fastclaw daemon __run`) wraps the gateway in a restart loop and recognises **exit code 75** as "restart-on-purpose" (used by the **Save & Restart** button), so config changes apply atomically without manual restarts.

---

## 🔧 Configuration

Pawnix reads `~/.fastclaw/fastclaw.json` on startup.

### Storage

```json
{
  "storage": {
    "type": "postgres",
    "dsn": "postgres://user:pass@localhost:5432/fastclaw?sslmode=disable",
    "autoMigrate": true
  }
}
```

Supported types: `"file"` (default), `"postgres"`, `"sqlite"`.

### LLM providers

```json
{
  "providers": {
    "openai":   { "apiKey": "${OPENAI_API_KEY}",   "apiBase": "https://api.openai.com/v1",      "apiType": "openai" },
    "deepseek": { "apiKey": "${DEEPSEEK_API_KEY}", "apiBase": "https://api.deepseek.com/v1",   "apiType": "openai" }
  }
}
```

### Multi-agent + bindings

```json
{
  "agents": {
    "defaults": { "model": "deepseek/deepseek-v4-pro", "maxTokens": 8192, "temperature": 0.7, "thinking": "medium" },
    "list": [
      { "id": "coder",   "model": "anthropic/claude-sonnet-4.5", "skills": ["debugging", "tdd"],            "thinking": "high"   },
      { "id": "analyst", "model": "deepseek/deepseek-v4-flash",  "skills": ["financial-modeling"],          "thinking": "medium" }
    ]
  },
  "bindings": [
    { "agentId": "coder", "match": { "channel": "telegram", "accountId": "main" } }
  ]
}
```

### Channels (managed via UI)

```json
{
  "channels": {
    "telegram": {
      "enabled": true,
      "accounts": {
        "main": { "botToken": "123456:ABC..." }
      }
    }
  }
}
```

> Edit channels through the **Channels** page in the dashboard — it handles token masking, agent binding, and triggers the daemon to restart cleanly via the supervisor.

### MCP servers

```json
{
  "mcpServers": {
    "brave-search": { "type": "http",  "url": "https://api.search.brave.com/mcp", "headers": { "Authorization": "Bearer ${BRAVE_API_KEY}" } },
    "local-tool":   { "type": "stdio", "command": "python", "args": ["-m", "my_mcp_server"], "env": { "API_KEY": "${MY_API_KEY}" } }
  }
}
```

---

## 🔌 Plugins

Pawnix plugins are subprocesses that speak JSON-RPC 2.0 over stdin/stdout — write them in Python, Node, Go, anything.

**Plugin types:** `channel` · `tool` · `provider` · `hook`

```bash
fastclaw plugins install telegram                 # from Pawnix Hub
fastclaw plugins install github.com/user/repo     # from a GitHub repo
fastclaw plugins install ./my-plugin              # from a local directory
```

Official plugins live in [`plugins/`](plugins/). Community plugins are indexed at [Pawnix Hub](https://github.com/fastclaw-ai/fastclaw-hub).

---

## 🖥 Web Dashboard

Available at `http://localhost:18953`:

| Page | What it does |
|------|--------------|
| **Overview** | Gateway status, stats, quick actions, channel CTA |
| **Chat** | Talk to your agents in the browser, async task-based |
| **Agents** | Create / edit / delete agents, edit `SOUL.md`, `MEMORY.md` |
| **Models** | Manage providers and the default model |
| **Skills** | Browse and manage installed skills |
| **Plugins** | Enable / disable / configure plugins |
| **Channels** | Add Telegram / Discord / Slack accounts and bind to agents |
| **Cron Jobs** | Create and manage scheduled tasks |
| **Apps** | Quick-launch companion dashboards |
| **Settings** | Storage backend, webhook, gateway config |

---

## 🔗 API

OpenAI-compatible:

```bash
curl -X POST http://localhost:18953/v1/chat/completions \
  -H "Authorization: Bearer $TOKEN" \
  -H "Content-Type: application/json" \
  -d '{"model":"auto","messages":[{"role":"user","content":"hello"}],"stream":true}'
```

WebSocket (OpenClaw-compatible):

```javascript
const ws = new WebSocket('ws://localhost:18953/ws');
ws.send(JSON.stringify({ type: 'chat', message: 'hello' }));
```

---

## 🛠 CLI Reference

```bash
# Core
fastclaw                      # Start (setup wizard or gateway)
fastclaw gateway              # Start gateway explicitly
fastclaw version              # Version info
fastclaw doctor               # Check config health
fastclaw upgrade              # Update to latest

# Daemon
fastclaw daemon start         # Start as background daemon
fastclaw daemon stop          # Stop
fastclaw daemon restart       # Restart
fastclaw daemon status        # Status check

# Plugins
fastclaw plugins install NAME
fastclaw plugins list
fastclaw plugins remove ID

# Agents
fastclaw agent create <name>
fastclaw agent list

# Other
fastclaw migrate              # Migrate JSONL sessions → PostgreSQL
fastclaw session ...
fastclaw provider ...
fastclaw policy ...
fastclaw backup
```

---

## 🛠 Development

```bash
git clone https://github.com/fastclaw-ai/fastclaw.git
cd fastclaw

./setup.sh           # check (or install) Go, Node, PostgreSQL

make build           # binary + embedded web UI
make build-web       # web UI only
make dev             # dev mode with hot reload
make release-local   # all platforms
make test
```

### Workspace structure

```
~/.fastclaw/
├── fastclaw.json           # Main config
├── fastclaw.pid            # Daemon PID (single source of truth)
├── logs/gateway.log        # Daemon log
├── agents/<id>/agent/      # SOUL / MEMORY / IDENTITY / sessions / skills / ...
└── plugins/                # Installed plugins
```

### PostgreSQL schema (when `storage.type = "postgres"`)

`configs` · `agents` · `workspace_files` · `sessions` · `memory_logs` · `cron_jobs` · `chat_tasks` · `memories` · `research_data` · `schema_registry`

---

## Contributing

Contributions welcome. Pawnix's strength is **simplicity** — keep it that way.

- **Core framework & official plugins** — contribute to this repo
- **Community plugins** — open your own repo, then submit it to [Pawnix Hub](https://github.com/fastclaw-ai/fastclaw-hub)

## License

[MIT](LICENSE)

---

<div align="center">
  <sub>Built with 🐾 by the Pawnix community</sub>
</div>
