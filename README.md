<div align="center">

<img src="web/public/logo.svg" width="120" alt="Pawnix logo" />

# Pawnix

**A persistent, shared memory backend for your AI tools — in a single Go binary.**

[![Go](https://img.shields.io/badge/Go-1.25+-00ADD8?style=flat&logo=go)](https://go.dev)
[![License](https://img.shields.io/badge/License-MIT-blue.svg)](LICENSE)
[![Status](https://img.shields.io/badge/status-v0.3.x-5eead4)](#-roadmap)

Cross-session memory that persists forever · retrieved every turn · shared across hermes, Claude Code & Codex over MCP

[Install](#-install) · [Memory Backend](#-memory-backend-pawnix-mcp) · [Quick Start](#-quick-start) · [Architecture](#-architecture) · [Roadmap](#-roadmap)

</div>

---

## What is Pawnix?

**Your AI tools forget everything between sessions. Pawnix is the memory they share.**

Every chat with hermes, Claude Code, or Codex starts from zero — the same preferences re-explained, the same decisions re-litigated, context that mattered yesterday gone today. Pawnix is a self-hosted **memory layer** that sits underneath all of them: one persistent, searchable pool of long-term memory, exposed over the Model Context Protocol so any MCP client can read and write it.

- **Persists forever.** Memory lives in PostgreSQL + pgvector, not in a process or a session. Close the tool, come back a month later — it's still there.
- **Retrieved every turn.** Clients query the pool at the start of each turn (read-aggressive tool descriptions steer this), so the model walks in already knowing your preferences, past decisions, and project background.
- **Shared across tools.** hermes, Claude Code, and Codex each spawn a `pawnix mcp` subprocess against the **same shared pool** — what one learns, the others can recall. Per-memory `source:` tags keep provenance.
- **Read-aggressive, write-conservative.** Reads are cheap and encouraged; writes are deliberately restrained (only cross-session-important facts) so the pool stays signal, not noise.
- **Single binary.** No Docker, no Python venv, no Node runtime. `pawnix mcp` is a stdio subprocess any MCP client can launch. Cross-compiles for macOS / Linux / Windows.

```bash
curl -fsSL https://raw.githubusercontent.com/softbreezee/claw-os/main/install.sh | bash
pawnix mcp --source claude-code     # run the memory backend as a stdio MCP server
```

> **Pawnix is also a full agent runtime.** The same binary can run standalone multi-agent teams over Telegram / Discord / Slack, with cron, an inbox, and a web dashboard — see [Agent Runtime](#-agent-runtime-optional) below. But its center of gravity is the memory layer: **lightweight, and good at remembering.**

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

## 🧠 Memory Backend (`pawnix mcp`)

The heart of Pawnix. `pawnix mcp` runs a stdio MCP server that exposes one PostgreSQL + pgvector memory pool to any MCP client, so hermes, Claude Code, and Codex all read and write the **same** long-term memory.

```
hermes ─┐
codex  ─┼─ spawn ─→  pawnix mcp  (stdio, JSON-RPC 2.0)
claude ─┘                 │
                          ▼
                 PostgreSQL + pgvector
                 (memories table, shared pool agent_id=shared)
```

**Design principle — read-aggressive, write-conservative.** The decision of *what's worth remembering* stays with the calling model, steered by the tool descriptions. Reads are encouraged every turn; writes are held to cross-session-important facts only — so the pool never fills with the disposable context that plagues naive auto-persist.

### Tools

| Tool | Role | When |
|---|---|---|
| `memory_search` | Semantic search over the pool (degrades to keyword + recency without embeddings) | **Every turn** that might touch past context — read aggressively |
| `memory_write` | Write one long-term memory (`content` + `kind`: fact / user_note / report) | **Only cross-session-important info**; when unsure, ask the user first |
| `memory_stats` | Pool health (total / embedded / last 24h) | Diagnostics only |

`--source <tool>` bakes a `source:<tool>` provenance tag into every written memory at spawn time — the origin is fixed by the launch flag, not self-reported by the model. `--agent <id>` overrides the default shared pool to read/write a single agent's memory.

### Wire it into your tools

```jsonc
// hermes / claw-os — config "mcpServers"
{ "mcpServers": { "memory": { "type": "stdio", "command": "pawnix", "args": ["mcp", "--source", "hermes"] } } }

// Claude Code — ~/.claude.json or project .mcp.json
{ "mcpServers": { "memory": { "command": "pawnix", "args": ["mcp", "--source", "claude-code"] } } }
```

```toml
# Codex — ~/.codex/config.toml
[mcp_servers.memory]
command = "pawnix"
args = ["mcp", "--source", "codex"]
```

**Prerequisite:** `~/.pawnix/pawnix.json` must set `storage.type = "postgres"` with a valid DSN (the memory backend needs pgvector). If no embedder is configured, search degrades gracefully to keyword + recency. Full guide: [docs/memory-mcp.md](docs/memory-mcp.md). Smoke-test the JSON-RPC loop with [`scripts/mcp-smoke-test.sh`](scripts/mcp-smoke-test.sh).

---

## 🤖 Agent Runtime (optional)

Beyond the memory backend, the same binary is a full standalone agent runtime — a self-hosted layer between you and any LLM: multi-agent teams, multi-channel messaging (Telegram / Discord / Slack), cron jobs, an inbox, and a web dashboard. You don't need any of it to use `pawnix mcp` as a memory backend — it's here if you want a batteries-included agent host too. Everything from here down describes that runtime.

## 🚀 Quick Start

1. Run `pawnix` — the setup wizard opens at `http://localhost:18953`.
2. Pick an LLM provider (OpenAI, Anthropic, or any OpenAI-compatible endpoint).
3. (Recommended) Pick PostgreSQL as the storage backend in step 2 — the shared memory pool, `/goal`, semantic memory, and the heartbeat health probe all need it.
4. Click **Launch** — start chatting in the browser.
5. Optional: open **Channels** in the sidebar to connect Telegram / Discord / Slack (see below).

### v0.3 demo path (5 minutes)

- **`/plan 复盘今天的两个账户`** → review the plan → click **▶ 继续执行**;watch the `todo.md` panel tick through steps in real time.
- While the agent is running, type a follow-up in the input box and click **Steer** — it folds into the next ReAct iteration without interrupting the current tool.
- **`/goal 盯着恩捷股份突破 85 提醒我`** → keep chatting; the agent will keep auditing the goal across turns until you `/goal clear`. See [docs/v0.3-test-guide.md](docs/v0.3-test-guide.md) for caveats (token usage, etc.).

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

### **v0.2 — Rebrand & The OS Layer**

> *From FastClaw to Pawnix — and from "agent runtime" to "personal OS".*

- [x] Binary `pawnix` · config dir `~/.pawnix/` · config file `pawnix.json` · module `github.com/softbreezee/claw-os`
- [x] launchd / systemd labels updated; new logo + favicon; full UI/CLI rename
- [x] **Cron OS-ification** — single store-backed ledger; UI / agent tool / scheduler all read the same source; cron tools auto-inherit current chat origin (Web → Inbox, Telegram → Telegram)
- [x] **Inbox + Notifications subsystem** — `store.NotificationRecord` + `/api/notifications` + Sidebar badge + browser-native toast
- [x] **`notify(text, channel?)` tool** — any agent can push to the user; Inbox by default, IM channels when `MyChatID` is configured
- [x] **`MyChatID` per channel** — decouples "agent that handles incoming messages" (binding) from "where to push outgoing notifications"

### **v0.3 — Steering** &nbsp;`← you are here`

> *Turn the agent from a one-shot chatbot into a partner you can sit beside all day.* See [docs/v0.3-plan.md](docs/v0.3-plan.md) and [docs/v0.3-test-guide.md](docs/v0.3-test-guide.md). For a code-verified ledger of what shipped vs. what carries delivery debt, see [docs/STATUS.md](docs/STATUS.md).

**Foundation (Week 1)**
- [x] **`bus.InboundMessage.Origin` 枚举常量化** — `OriginUser/Cron/Webhook/Internal/Heartbeat/SubAgent/GoalContext/UserSteer` + `IsRuntimeInjected` helper
- [x] **Skills section token budget** — system prompt skills 段限到 ~2k token;`load_skill` 工具按需读 body + IP guard wrapper(防 chatter 套出 prompt template)
- [x] **Memory pgvector readiness** — 启动期 `VerifyVectorReady` 探针 + Heartbeat 30min embedding 覆盖率巡检;silent failure 不再隐形
- [x] **Onboarding 收尾** — `handleLaunch` `waitForGateway` 轮询代替硬编码 3s 跳转;Telegram bot token 字段加 UI

**可打断 — Mid-run Steering (Week 2)**
- [x] **`SteerBuffer`** + `ContextWithSteerBuffer` + agent loop 在每轮 ReAct iteration 头部 drain
- [x] **`POST /api/chat/tasks/{id}/steer`** + taskrunner.Steer + `ErrTaskNotRunning` (409 fallback to /submit)
- [x] **Web UI**: turn 进行中输入框走 steer 路径,新增 **Steer** 按钮 + 虚线 steer 气泡 + `queued` / `applied` badge

**可围观 — `/plan` slash + Todo 进度面板 (Week 3)**
- [x] **`/plan <task>`** slash:tools=nil 单轮 LLM 调用 + plan-mode nudge + tool catalog 注入
- [x] **PlanBubble**:streaming 时 `drafting…` badge,完成后底部出现 **▶ 继续执行** / **✎ 调整方案** 内联按钮
- [x] **TodoPanel**:每 3s 轮询 `workspace/todo.md`,markdown checkbox 渲染成进度卡,塌缩 + N/M badge

**可托管 — `/goal` MVP (Week 4)**
- [x] **`internal/agent/goal/`** 包:Goal domain + Status 状态机 + `FoldUsage` 计费 + `ContinuationPrompt` / `BudgetLimitPrompt` 模板
- [x] **`agent_goals` PG 表** + `GoalStore` (Create/Get/Update/Delete) + UNIQUE (agent_id, session_key)
- [x] **`/goal` slash 命令族**:create / show / pause / resume / clear,tight-loop 防御(goal_context origin 不再触发新一轮)
- [x] **PostTurn 自动 fire continuation**:每个用户 turn 结束都 audit goal,直到 model 标 complete 或 `/goal clear`
- ⚠️ **Token budget guard 未接线** — `FoldUsage`/`BudgetLimitPrompt` 代码就位但无生产调用方,goal 从不设预算;防跑飞目前只靠 continuation tight-loop 闸。详见 [docs/STATUS.md §4](docs/STATUS.md)

**v0.4 储备**
- [x] **`delegate_task` 工具** — sub-task 独立 ReAct loop + 独立 iteration budget,parent context 不污染;per-Agent serial mutex 防 sandbox/workspace 写竞争;sub-task toolset 过滤 `delegate_task` 防嵌套

### **v0.4 — Multi-modal & Voice**

- [ ] Voice in/out (Whisper STT + pluggable TTS)
- [ ] Screen understanding (screenshot → multimodal LLM → action)
- [ ] First-class file ingestion (PDF / video / Excel → memory)

### **v0.5 — Memory Graph & Agent Marketplace**

- [x] **Cross-tool shared memory pool** — `pawnix mcp` stdio server, one pgvector pool shared by hermes / Claude Code / Codex, per-row `source:` provenance
- [ ] Memory graph: extract entities and relations from conversations
- [ ] Time-aware retrieval (recency-weighted scoring)
- [ ] Per-source read/write permissions on the shared pool
- [ ] Memory browser UI: visualize, edit, prune what each agent knows
- [ ] Skill auto-induction: turn frequent prompt patterns into callable skills
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
| **Memory backend** &nbsp;`core` | `pawnix mcp` stdio server exposes one pgvector pool to hermes / Claude Code / Codex; shared `agent_id`, per-row `source:` provenance; read-aggressive, write-conservative |
| **Memory** | `MEMORY.md` + FTS / pgvector semantic retrieval injected each turn; AutoPersist LLM extraction every N turns + embedding; readiness probe + heartbeat coverage巡检. (Dedup / compression / PII-blocking are designed but not yet shipped — see [docs/memory-governance.md](docs/memory-governance.md)) |
| **Skills** | Per-section token budget (~2k);on-demand `SKILL.md` loading via `load_skill` + IP guard; agents can learn skills from interaction patterns |
| **Channels** | Web · Telegram · Discord · Slack · custom (JSON-RPC plugin) |
| **Cron** | Cron expressions / intervals / one-shot — single store-backed ledger |
| **Inbox** | Cron / webhook / `notify` results land in the dashboard Inbox + browser toast |
| **`notify(text, channel?)`** | Any agent can push to the user — Inbox by default, IM when `MyChatID` is set |
| **Mid-run steering** &nbsp;`v0.3` | Add a "supplemental instruction" while the agent is running — folded in at the next ReAct iteration boundary |
| **`/plan` slash** &nbsp;`v0.3` | One-shot plan-only turn (tools disabled), with **继续执行 / 调整方案** inline buttons + live `todo.md` progress panel |
| **`/goal` slash** &nbsp;`v0.3` | Persistent multi-turn objective; agent auto-audits across turns until complete or `/goal clear`; PG-backed (one row per session) |
| **`delegate_task` tool** &nbsp;`v0.4` | Sub-task with own context + own iteration budget; SERIAL within an agent; no nesting |
| **Hooks** | Before / After hooks on prompts, model calls, tool calls |
| **Hot reload** | Edit config or `SOUL.md` → takes effect immediately |
| **Storage** | File (default) · PostgreSQL + pgvector · SQLite + FTS5 |
| **Security** | Sandbox exec · YAML policy engine · AES-256-GCM credential vault · PII scrubbing · tool-loop detection |
| **Platform** | Web dashboard · OpenAI-compatible REST · WebSocket · MCP client · daemon mode |

### Built-in tools

`exec` · `read_file` / `write_file` / `list_dir` · `web_fetch` · `web_search` · `memory_search` / `memory_write` / `memory_stats` · `message` · `notify` · `spawn_subagent` · `delegate_task` · `create_cron_job` / `list_cron_jobs` / `delete_cron_job` · `load_skill` · `db_query` / `db_create_table` · all MCP tools

The three `memory_*` tools are also exposed standalone over stdio via [`pawnix mcp`](#-memory-backend-pawnix-mcp), so external MCP clients (hermes / Claude Code / Codex) share the same pool.

### Slash commands

`/help` · `/status` · `/usage` · `/insights [N]` · `/new` `/reset` · `/retry` · `/undo` · `/compact` · `/personality [<name>]` · `/model [<name>]` · `/version` · **`/plan <task>`** · **`/goal [<task>|pause|resume|clear]`**

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
pawnix mcp                  # run the memory MCP server (stdio) — add --source <tool>
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

PostgreSQL tables (when `storage.type = "postgres"`): `configs` · `agents` · `workspace_files` · `sessions` · `memory_logs` · `cron_jobs` · `notifications` · `chat_tasks` · `memories` · `mcp_events` · `research_data` · `schema_registry`

---

## Contributing

Contributions welcome. Pawnix's strength is **simplicity** — keep it that way.

## License

[MIT](LICENSE)

## 🙏 Acknowledgements

Pawnix stands on the shoulders of [**fastclaw-ai/fastclaw**](https://github.com/fastclaw-ai/fastclaw).

This project began as a fork of FastClaw and has since refocused into a lightweight, self-hosted **memory backend** for AI tools — with a full agent runtime still riding along in the same binary. The entire v0.1 Foundation — multi-LLM provider routing, multi-agent gateway, channels (Telegram / Discord / Slack), the skill system, plugin protocol, MCP client, dual-layer memory, daemon supervisor, web dashboard scaffolding — all originated in or grew directly out of FastClaw's design and code.

Huge thanks to the FastClaw authors for building such a clean, hackable foundation, and for releasing it under MIT so projects like this one are even possible. If you like Pawnix, please go give the [original repo](https://github.com/fastclaw-ai/fastclaw) a star.

---

<div align="center">
  <sub>🐾 The memory your AI tools share · Built by the Pawnix community · Forked with gratitude from <a href="https://github.com/fastclaw-ai/fastclaw">FastClaw</a></sub>
</div>
