<div align="center">

<img src="web/public/logo.svg" width="120" alt="Pawnix logo" />

# Pawnix

**A self-hosted, AI-Native personal OS — in a single Go binary.**

[![Go](https://img.shields.io/badge/Go-1.25+-00ADD8?style=flat&logo=go)](https://go.dev)
[![License](https://img.shields.io/badge/License-MIT-blue.svg)](LICENSE)
[![Status](https://img.shields.io/badge/status-v0.2.x-5eead4)](#-roadmap)

Multi-agent · Multi-channel · Long-term memory · Cron + Inbox notifications · MCP · Plugins · Web dashboard

[Install](#-install) · [Quick Start](#-quick-start) · [Channels](#-channels-setup) · [Architecture](#-architecture) · [Roadmap](#-roadmap)

</div>

---

## What is Pawnix?

Pawnix is **not just an agent runtime** — it's a long-running, self-hosted layer between you and any LLM, designed to behave the way an operating system does for your AI life:

- **Always on.** Daemon supervisor keeps the gateway alive across crashes and config changes.
- **Multi-agent.** Run a personal team — coder, analyst, scheduler — each with its own personality, memory, and skills.
- **Multi-channel by default.** Same agents, reachable from your browser, Telegram, Discord, Slack, or any custom plugin channel.
- **Notifications as a first-class primitive.** Cron jobs, watchers, and any agent can `notify` you — through the in-app Inbox, browser toasts, or back through your IM channels.
- **Persistent memory.** `MEMORY.md` + searchable conversation logs + pgvector semantic memory.
- **Skills that grow.** Drop a `SKILL.md` and your agent can do new things; agents can learn skills from interaction patterns.
- **Local-first, cloud-optional.** Defaults to plain JSON in `~/.pawnix/`. Switch to PostgreSQL / SQLite when you outgrow files.
- **Single binary.** No Docker, no Python venv, no Node runtime. Cross-compiles for macOS / Linux / Windows.

```bash
curl -fsSL https://raw.githubusercontent.com/softbreezee/claw-os/main/install.sh | bash
pawnix    # opens the setup wizard at http://localhost:18953
```

---

## 📦 Install

```bash
# macOS / Linux
curl -fsSL https://raw.githubusercontent.com/softbreezee/claw-os/main/install.sh | bash

# Windows: download the .zip from Releases and double-click pawnix.exe

# From source
git clone https://github.com/softbreezee/claw-os.git
cd claw-os && make build

# Upgrade
pawnix upgrade
```

---

## 🚀 Quick Start

1. Run `pawnix` — the setup wizard opens at `http://localhost:18953`.
2. Pick an LLM provider (OpenAI, Anthropic, or any OpenAI-compatible endpoint).
3. Click **Launch** — start chatting in the browser.
4. Optional: open **Channels** in the sidebar to connect Telegram / Discord / Slack (see below).

---

## 💬 Channels Setup

Connect a chat platform once and the same agents become reachable from anywhere — Web, Telegram, Discord, Slack. Each bot binds to one agent (who handles incoming messages) and carries a **My chat ID** (where push notifications come back to you).

### Telegram

1. **Create a bot.** Talk to [@BotFather](https://t.me/BotFather) → `/newbot` → name it → copy the bot token (`123456789:ABC…`).
2. **Find your chat ID.** Talk to [@userinfobot](https://t.me/userinfobot) → it replies with your numeric ID (e.g. `8175643861`).
3. **Add the channel.** Pawnix dashboard → **Channels** → **Add channel** → Telegram → paste the token → bind to an agent → **Save & Restart**.
4. **Set "My chat ID".** In the Telegram card, paste the ID from step 2 into the **My chat ID** field → **Save & Restart**.
5. **Done.** Open Telegram, search for your bot, send a message — your agent replies. Cron jobs and `notify(..., channel='telegram')` calls now also push back here.

### Discord

1. **Create a bot.** [Discord Developer Portal](https://discord.com/developers/applications) → New Application → **Bot** tab → reset & copy the token. Enable **MESSAGE CONTENT INTENT**.
2. **Invite it to your server.** OAuth2 URL Generator → scopes: `bot` → permissions: `Send Messages`, `Read Message History` → open the URL → invite to your server.
3. **Find your user ID.** Discord → Settings → Advanced → enable Developer Mode → right-click your name → **Copy User ID**.
4. **Add the channel.** Pawnix dashboard → **Channels** → **Add channel** → Discord → paste the token → bind agent → fill **My user ID** → **Save & Restart**.
5. **Done.** DM the bot from Discord to talk; cron / `notify` deliver back to your DMs.

### Slack

1. **Create a Slack app** at <https://api.slack.com/apps> with **Socket Mode** enabled.
2. Generate a **Bot User OAuth Token** (`xoxb-…`) and an **App-Level Token** with `connections:write` (`xapp-…`).
3. **Find your member ID.** Profile → **⋯** → **Copy member ID** (`U0XXXXXXX`).
4. **Add the channel.** Channels → Add → Slack → paste both tokens → bind agent → fill **My user ID** → **Save & Restart**.

> The "My chat/user ID" you set is what unlocks **proactive notifications** — any agent can call `notify(text, channel='telegram')` to ping you back here, with no manual chat-ID juggling.

---

## 🗺 Roadmap

### **v0.1 — Foundation**

> *The runtime. Every later milestone builds on this.*

- [x] Multi-LLM provider routing with per-call model override (OpenAI, Anthropic, DeepSeek, Gemini, GLM, Kimi, Groq, Ollama, OpenRouter)
- [x] Multi-agent + team `@mention` routing
- [x] Channels CRUD UI (Telegram / Discord / Slack) with hot agent bindings
- [x] Skills + Plugins (JSON-RPC subprocess) + MCP (HTTP & stdio)
- [x] Dual-layer memory (`MEMORY.md` + FTS / pgvector)
- [x] Cron jobs + per-agent heartbeat
- [x] Web dashboard at `:18953`
- [x] Daemon supervisor with crash auto-restart and restart-aware exit codes

### **v0.2 — Rebrand & The OS Layer** &nbsp;`← you are here`

> *From FastClaw to Pawnix — and from "agent runtime" to "personal OS".*

- [x] Binary `pawnix` · config dir `~/.pawnix/` · config file `pawnix.json` · module `github.com/softbreezee/claw-os`
- [x] launchd / systemd labels updated; new logo + favicon; full UI/CLI rename
- [x] **Cron OS-ification** — single store-backed ledger; UI / agent tool / scheduler all read the same source; cron tools auto-inherit current chat origin (Web → Inbox, Telegram → Telegram)
- [x] **Inbox + Notifications subsystem** — `store.NotificationRecord` + `/api/notifications` + Sidebar badge + browser-native toast
- [x] **`notify(text, channel?)` tool** — any agent can push to the user; Inbox by default, IM channels when `MyChatID` is configured
- [x] **`MyChatID` per channel** — decouples "agent that handles incoming messages" (binding) from "where to push outgoing notifications"

### **v0.3 — Memory OS** &nbsp;*next*

> *Make the agent remember you, not just facts.*

- [ ] Memory graph: extract entities and relations from conversations
- [ ] Time-aware retrieval (recency-weighted scoring)
- [ ] Cross-agent shared memory pool with permissions
- [ ] Memory browser UI: visualize, edit, prune what each agent knows
- [ ] Skill auto-induction: turn frequent prompt patterns into callable skills

### **v0.4 — Multi-modal & Voice**

- [ ] Voice in/out (Whisper STT + pluggable TTS)
- [ ] Screen understanding (screenshot → multimodal LLM → action)
- [ ] First-class file ingestion (PDF / video / Excel → memory)

### **v0.5 — Agent Marketplace**

- [ ] Skills Hub: a "GitHub for skills" with one-click install
- [ ] Agent export/import: full bundle (persona + skills + memory schema)

### **v0.6 — Distributed Mesh**

- [ ] Multi-device sync (laptop + phone + server share state)
- [ ] P2P session handover (start at home, finish on the road)
- [ ] End-to-end encrypted remote access — without renting a VPS

### **v1.0 — Production Persona**

- [ ] Full audit log (every tool call, traceable & rollback-able)
- [ ] Permission system v2 (fine-grained, dangerous-op confirmation)
- [ ] Backup, disaster recovery, HA mode
- [ ] Multi-user / team mode

---

## ✨ Features

| Capability | Notes |
|---|---|
| **ReAct loop** | Multi-turn reasoning + tool calling, configurable max iterations |
| **Any LLM** | Any OpenAI-compatible API; per-call model override from the chat UI |
| **Multi-agent** | Independent persona / memory / skills per agent; team `@mention` routing |
| **Memory** | `MEMORY.md` + FTS / pgvector; auto-pruning + LLM-driven compression |
| **Skills** | On-demand `SKILL.md` loading; agents can learn skills from interaction patterns |
| **Channels** | Web · Telegram · Discord · Slack · custom (JSON-RPC plugin) |
| **Cron** | Cron expressions / intervals / one-shot — single store-backed ledger |
| **Inbox** | Cron / webhook / `notify` results land in the dashboard Inbox + browser toast |
| **`notify(text, channel?)`** | Any agent can push to the user — Inbox by default, IM when `MyChatID` is set |
| **Hooks** | Before / After hooks on prompts, model calls, tool calls |
| **Hot reload** | Edit config or `SOUL.md` → takes effect immediately |
| **Storage** | File (default) · PostgreSQL + pgvector · SQLite + FTS5 |
| **Security** | Sandbox exec · YAML policy engine · AES-256-GCM credential vault · PII scrubbing · tool-loop detection |
| **Platform** | Web dashboard · OpenAI-compatible REST · WebSocket · MCP client · daemon mode |

### Built-in tools

`exec` · `read_file` / `write_file` / `list_dir` · `web_fetch` · `web_search` · `memory_search` · `message` · `notify` · `spawn_subagent` · `create_cron_job` / `list_cron_jobs` / `delete_cron_job` · `load_skill` · `db_query` / `db_create_table` · all MCP tools

---

## 🏗 Architecture

```
                ┌────────────────────────────────────────────────────────────┐
                │                       Pawnix Gateway                       │
                │                                                            │
   Web UI ────▶ │   ┌──────────┐               ┌─────────────────────────┐  │
   Channels ──▶ │   │ Message  │               │      Agent Manager      │  │
   Plugins ───▶ │   │   Bus    │──────────────▶│  Agent A · Agent B · …  │  │
   Webhooks ──▶ │   │          │◀──────────────│  (skills · hooks · mem) │  │
   API ───────▶ │   └──────────┘               └────────────┬────────────┘  │
                │                                           │                │
                │   ┌─────────┬───────────┬─────────────┬───┴─────┬──────┐  │
                │   ▼         ▼           ▼             ▼         ▼      ▼  │
                │ Tools     Memory     Sessions       Policy    Cron  Inbox │
                │                                                            │
                │   ┌────────────────────────────────────────────────────┐  │
                │   │  REST · SSE · WebSocket · Webhooks · Web UI :18953 │  │
                │   └────────────────────────────────────────────────────┘  │
                └────────────────────────────────────────────────────────────┘

                ┌────── daemon supervisor (crash + restart aware) ──────┐
```

The daemon supervisor (`pawnix daemon __run`) wraps the gateway in a restart loop and recognises **exit code 75** as "restart-on-purpose" (used by **Save & Restart** in the UI), so config changes apply atomically.

### Notifications & cron data flow

```
   ┌─── single store ledger (cron_jobs, notifications) ───┐
   │                                                       │
   ▼                                                       │
[scheduler.pollStore]   ◄── store.SaveCronJob ◄──── UI / agent's create_cron_job
   │
   │ fire (msg.AgentID + msg.Origin='cron')
   ▼
[routing.routeDM] ──► [agent.HandleMessage] ──► reply
                                                  │
                ┌─────────────────────────────────┴────────────────────┐
                │                                                       │
                ▼                                                       ▼
   Channel='' or 'web'                                  Channel='telegram' / 'slack' / …
        │                                                       │
        ▼                                                       ▼
   store.SaveNotification                              bus.Outbound → chanMgr → IM bot
        │                                                       │
        ▼                                                       ▼
   Sidebar badge + browser toast                         Push lands in Telegram / Slack
```

The same `notify(text, channel?)` tool — usable by any agent, not just the one bound to the channel — lets watchers, finished long-running tasks, and proactive alerts share this pipeline.

---

## 🔧 Configuration

Pawnix reads `~/.pawnix/pawnix.json` on startup. The dashboard manages most of this for you; the snippets below are for power users.

```json
{
  "providers": {
    "openai":   { "apiKey": "${OPENAI_API_KEY}",   "apiBase": "https://api.openai.com/v1",      "apiType": "openai" },
    "deepseek": { "apiKey": "${DEEPSEEK_API_KEY}", "apiBase": "https://api.deepseek.com/v1",   "apiType": "openai" }
  },

  "agents": {
    "defaults": { "model": "deepseek/deepseek-v4-pro", "maxTokens": 8192, "temperature": 0.7, "thinking": "medium" },
    "list": [
      { "id": "coder",   "model": "anthropic/claude-sonnet-4.5", "skills": ["debugging", "tdd"] },
      { "id": "analyst", "model": "deepseek/deepseek-v4-flash",  "skills": ["financial-modeling"] }
    ]
  },

  "channels": {
    "telegram": {
      "enabled": true,
      "accounts": {
        "main": {
          "botToken": "123456:ABC...",
          "myChatId": "8175643861"
        }
      }
    }
  },

  "bindings": [
    { "agentId": "coder", "match": { "channel": "telegram", "accountId": "main" } }
  ],

  "storage": {
    "type": "postgres",
    "dsn": "postgres://user:pass@localhost:5432/pawnix?sslmode=disable",
    "autoMigrate": true
  },

  "mcpServers": {
    "brave-search": { "type": "http",  "url": "https://api.search.brave.com/mcp", "headers": { "Authorization": "Bearer ${BRAVE_API_KEY}" } },
    "local-tool":   { "type": "stdio", "command": "python", "args": ["-m", "my_mcp_server"] }
  }
}
```

Storage backends: `"file"` (default), `"postgres"`, `"sqlite"`.

---

## 🔌 Plugins

Pawnix plugins are subprocesses speaking JSON-RPC 2.0 over stdin/stdout — write them in Python, Node, Go, anything.

**Plugin types:** `channel` · `tool` · `provider` · `hook`

```bash
pawnix plugins install telegram                 # from Pawnix Hub
pawnix plugins install github.com/user/repo     # from a GitHub repo
pawnix plugins install ./my-plugin              # from a local directory
```

Official plugins live in [`plugins/`](plugins/).

---

## 🖥 Web Dashboard

`http://localhost:18953`

| Page | What it does |
|---|---|
| **Overview** | Gateway status, stats, quick actions |
| **Chat** | Talk to your agents in the browser, async task-based |
| **Inbox** | Cron / webhook / agent-initiated notifications + browser toasts |
| **Agents** | Create / edit / delete agents, edit `SOUL.md`, `MEMORY.md` |
| **Models** | Manage providers and the default model |
| **Skills** | Browse and manage installed skills |
| **Plugins** | Enable / disable / configure plugins |
| **Channels** | Add Telegram / Discord / Slack accounts; bind agents; set "My chat ID" for push |
| **Cron Jobs** | Create and manage scheduled tasks |
| **Apps** | Quick-launch companion dashboards |
| **Settings** | Storage backend, webhook, gateway config |

---

## 🔗 API

```bash
# OpenAI-compatible chat
curl -X POST http://localhost:18953/v1/chat/completions \
  -H "Authorization: Bearer $TOKEN" \
  -H "Content-Type: application/json" \
  -d '{"model":"auto","messages":[{"role":"user","content":"hello"}],"stream":true}'
```

```javascript
// WebSocket (OpenClaw-compatible)
const ws = new WebSocket('ws://localhost:18953/ws');
ws.send(JSON.stringify({ type: 'chat', message: 'hello' }));
```

Internal endpoints (Pawnix-native): `/api/cron`, `/api/notifications`, `/api/channels`, `/api/agents`, `/api/chat/...`, `/api/daemon/restart`.

---

## 🛠 CLI

```bash
pawnix                      # start (setup wizard or gateway)
pawnix gateway              # start gateway explicitly
pawnix doctor               # check config health
pawnix upgrade              # update to latest

pawnix daemon start|stop|restart|status

pawnix plugins install <name|repo|path>
pawnix plugins list|remove

pawnix agent create|list
pawnix migrate              # JSONL sessions → PostgreSQL
pawnix backup
```

---

## 🛠 Development

```bash
git clone https://github.com/softbreezee/claw-os.git
cd claw-os
./setup.sh           # checks (and offers to install) Go, Node, PostgreSQL

make build           # binary + embedded web UI (production)
make dev-build       # fast local rebuild after `cd web && pnpm build`
make dev             # hot reload via air
make test
make release-local   # all platforms
```

### Workspace layout

```
~/.pawnix/
├── pawnix.json             # main config
├── pawnix.pid              # daemon PID
├── cron_jobs.json          # cron ledger (file backend)
├── notifications.json      # inbox ledger (file backend)
├── logs/gateway.log
├── agents/<id>/agent/      # SOUL / MEMORY / IDENTITY / sessions / skills / ...
└── plugins/
```

PostgreSQL tables (when `storage.type = "postgres"`): `configs` · `agents` · `workspace_files` · `sessions` · `memory_logs` · `cron_jobs` · `notifications` · `chat_tasks` · `memories` · `research_data` · `schema_registry`

---

## Contributing

Contributions welcome. Pawnix's strength is **simplicity** — keep it that way.

## License

[MIT](LICENSE)

## 🙏 Acknowledgements

Pawnix stands on the shoulders of [**fastclaw-ai/fastclaw**](https://github.com/fastclaw-ai/fastclaw).

This project began as a fork of FastClaw and evolved into a self-hosted AI-Native personal OS. The entire v0.1 Foundation — multi-LLM provider routing, multi-agent gateway, channels (Telegram / Discord / Slack), the skill system, plugin protocol, MCP client, dual-layer memory, daemon supervisor, web dashboard scaffolding — all originated in or grew directly out of FastClaw's design and code.

Huge thanks to the FastClaw authors for building such a clean, hackable foundation, and for releasing it under MIT so projects like this one are even possible. If you like Pawnix, please go give the [original repo](https://github.com/fastclaw-ai/fastclaw) a star.

---

<div align="center">
  <sub>Built with 🐾 by the Pawnix community · Forked with gratitude from <a href="https://github.com/fastclaw-ai/fastclaw">FastClaw</a></sub>
</div>
