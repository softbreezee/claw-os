<div align="center">

# ⚡ FastClaw

A lightweight, self-hosted AI Agent framework written in Go.

[![Go](https://img.shields.io/badge/Go-1.25+-00ADD8?style=flat&logo=go)](https://go.dev)
[![License](https://img.shields.io/badge/License-MIT-blue.svg)](LICENSE)
[![Release](https://img.shields.io/github/v/release/fastclaw-ai/fastclaw?include_prereleases)](https://github.com/fastclaw-ai/fastclaw/releases)

**Single binary · Any LLM · Multi-channel · Plugin system · MCP · Web dashboard · PostgreSQL**

[Install](#-install) · [Quick Start](#-quick-start) · [Features](#-features) · [Architecture](#-architecture) · [Configuration](#-configuration)

</div>

---

## What is FastClaw?

FastClaw is a self-hosted AI agent runtime. It connects your LLM to chat platforms, executes tools, manages memory, runs scheduled tasks, and supports multi-agent teams — all from a single Go binary.

```bash
curl -fsSL https://raw.githubusercontent.com/fastclaw-ai/fastclaw/main/install.sh | bash
fastclaw    # Opens setup wizard in browser
```

## 📦 Install

### One-liner (macOS / Linux)

```bash
curl -fsSL https://raw.githubusercontent.com/fastclaw-ai/fastclaw/main/install.sh | bash
```

### Windows

Download `.zip` from [Releases](https://github.com/fastclaw-ai/fastclaw/releases), extract, double-click `fastclaw.exe`.

### Upgrade

```bash
fastclaw upgrade
```

### Quick Setup (database + dependencies)

For development environments, use the all-in-one setup script:

```bash
# Clone the repo
git clone https://github.com/fastclaw-ai/fastclaw.git
cd fastclaw

# Install all dependencies: PostgreSQL, Go, Node.js, build & configure
./setup.sh
```

### From source

```bash
git clone https://github.com/fastclaw-ai/fastclaw.git
cd fastclaw && make build
```

## 🚀 Quick Start

1. Run `fastclaw` — browser opens the setup wizard at `http://localhost:18953`
2. Pick your LLM provider (OpenAI, Anthropic, or custom OpenAI-compatible)
3. Click Launch — start chatting in the browser

That's it. Your agent is live. Connect messaging channels (Telegram, Discord, Slack) via built-in or plugin channels.

## ✨ Features

### Core Agent

| Feature | Description |
|---------|-------------|
| **ReAct Agent Loop** | Multi-turn reasoning + tool calling with configurable max iterations |
| **Any LLM** | OpenAI-compatible API (OpenAI, Anthropic Claude, DeepSeek, Gemini, GLM, Kimi, Groq, Ollama, OpenRouter) |
| **Multi-Agent** | Multiple agents with independent personalities, memory, and skills. Team support with @mentions |
| **Context Engineering** | Auto-pruning & LLM compression for long conversations, session key management |
| **Dual-Layer Memory** | `MEMORY.md` (persistent facts) + searchable conversation logs with FTS (full-text search) |
| **Skill System** | On-demand skill loading. Agents can learn new skills from interaction patterns |
| **Hook System** | Before/After hooks on prompts, model calls, and tool calls |
| **Hot Reload** | Edit config or SOUL.md → takes effect immediately, no restart |

### Storage & Database

| Feature | Description |
|---------|-------------|
| **File Storage** | Default: all data in `~/.fastclaw/` as JSON files |
| **PostgreSQL** | Full multi-tenant support with pgvector for semantic search |
| **SQLite** | Lightweight alternative with FTS5 full-text search |
| **Auto Migration** | Automatic table creation on startup (`storage.autoMigrate: true`) |
| **Session Persistence** | Per-channel session history with JSONL local + PostgreSQL remote |
| **Vector Search** | pgvector HNSW indexing for embedding-based semantic memory search |
| **DB Tools** | Agents can `db_query` and `db_create_table` on the research database |
| **Schema Registry** | Tracks all agent-created tables for future reference |

### Channels

| Channel | Status | Description |
|---------|--------|-------------|
| Web Chat | ✅ Built-in | Full chat UI at `/chat` with async task system |
| Telegram | ✅ Built-in | Native bot support, multi-account |
| Discord | ✅ Built-in | Bot integration via discordgo |
| Slack | ✅ Built-in | Socket Mode + Web API support |
| Any platform | ✅ Via plugin | JSON-RPC plugin protocol |

### Agent Tools

| Tool | Description |
|------|-------------|
| `exec` | Shell commands (with optional Docker sandbox) |
| `read_file` / `write_file` / `list_dir` | File system operations |
| `web_fetch` | Fetch web pages → markdown |
| `web_search` | Brave Search API integration |
| `memory_search` | Search conversation history and memory logs |
| `message` | Send messages to any channel |
| `spawn_subagent` | Delegate tasks to other agents |
| `create_cron_job` / `list_cron_jobs` / `delete_cron_job` | Manage scheduled tasks |
| `load_skill` | Dynamically load skill instructions |
| `db_query` | Execute SQL queries (SELECT/INSERT/UPDATE/DELETE) |
| `db_create_table` | Create research tables with schema registry |
| MCP tools | Connect external tools via Model Context Protocol |

### Automation & Scheduling

| Feature | Description |
|---------|-------------|
| **Cron Scheduler** | Three modes: cron expressions, intervals, one-time execution |
| **Heartbeat** | Agent wakes every N minutes to check `HEARTBEAT.md` |
| **Webhooks** | `POST /hooks` to trigger agent actions from external systems |
| **Distributed Locks** | Prevents duplicate cron job execution in multi-instance deployments |

### Security

| Feature | Description |
|---------|-------------|
| **Sandbox Exec** | Docker-based isolated command execution |
| **Policy Engine** | YAML policies for filesystem, network, tools, resources |
| **Credential Manager** | AES-256-GCM encrypted key storage, env auto-discovery |
| **PII Scrubbing** | Optional PII detection and scrubbing before LLM calls |
| **Tool Loop Detection** | Breaks after 3 identical consecutive calls |

### Platform & Extensibility

| Feature | Description |
|---------|-------------|
| **Web Dashboard** | Full management UI at `localhost:18953` |
| **Plugin System** | JSON-RPC 2.0 subprocess plugins in any language (Python, Node.js, Go, etc.) |
| **OpenAI-Compatible API** | `POST /v1/chat/completions` with SSE streaming |
| **WebSocket Gateway** | OpenClaw-compatible protocol |
| **MCP Support** | HTTP + stdio MCP server integration |
| **Multi-Tenant** | Built-in tenant isolation for cloud deployments |
| **Daemon Mode** | Background service with PID file, log management |

## 🏗 Architecture

```
                    ┌─────────────────────────────────────────────────────────┐
                    │                    FastClaw Gateway                      │
                    │                                                         │
  Web UI ────────▶ │  ┌──────────┐              ┌──────────────────────────┐ │
  Plugins ───────▶ │  │ Message  │              │     Agent Manager        │ │
  Webhooks ──────▶ │  │   Bus    │─────────────▶│                          │ │
  API ───────────▶ │  │          │◀─────────────│  Agent 1 (with skills)   │ │
  Channels ──────▶ │  └──────────┘              │  Agent 2 (with skills)   │ │
                    │                            │  Agent N ...              │ │
                    │                            └──────────────────────────┘ │
                    │                                      │                  │
                    │        ┌─────────────────────────────┼─────────────┐   │
                    │        ▼               ▼             ▼             ▼   │
                    │  ┌──────────┐  ┌──────────┐  ┌──────────┐  ┌──────────┐ │
                    │  │  Tools   │  │  Memory  │  │Sessions  │  │  Policy  │ │
                    │  │          │  │          │  │          │  │          │ │
                    │  │ exec     │  │MEMORY.md │  │ JSONL    │  │ rules    │ │
                    │  │ files    │  │ logs/    │  │ per-chat │  │ sandbox  │ │
                    │  │ web      │  │ FTS      │  │ compress │  │ creds    │ │
                    │  │ db       │  │ search   │  │          │  │ PII      │ │
                    │  │ MCP      │  │          │  │          │  │          │ │
                    │  └──────────┘  └──────────┘  └──────────┘  └──────────┘ │
                    │                                                         │
                    │  ┌──────────┐  ┌──────────┐  ┌──────────────────────┐  │
                    │  │  Cron    │  │ Plugins  │  │       Storage        │  │
                    │  │Schedule  │  │ JSON-RPC │  │  File | PG | SQLite  │  │
                    │  │Heartbeat │  │ channels │  │  pgvector            │  │
                    │  └──────────┘  └──────────┘  └──────────────────────┘  │
                    │                                                         │
                    │  ┌──────────────────────────────────────────────────┐  │
                    │  │  /v1/chat/completions (SSE)                      │  │
                    │  │  /ws (WebSocket)                                 │  │
                    │  │  /api/* (REST)                                   │  │
                    │  │  /hooks (Webhook ingress)                        │  │
                    │  │  Web Dashboard (:18953)                          │  │
                    │  └──────────────────────────────────────────────────┘  │
                    └─────────────────────────────────────────────────────────┘
```

### Data Flow

1. **Message Ingress**: Web UI, API, WebSocket, Channels, Plugins, or Webhooks send messages through the Message Bus
2. **Agent Routing**: Bindings map channel+account+peer patterns to specific agents
3. **ReAct Loop**: Agent reasons with LLM → calls tools → integrates results → continues until complete
4. **State Persistence**: Sessions, memory, and workspace files are stored to the configured backend
5. **Event Streaming**: SSE for chat completions, WebSocket for real-time updates

## 🔧 Configuration

FastClaw reads `~/.fastclaw/fastclaw.json` on startup. Key configuration blocks:

### Storage Backend

```json
{
  "storage": {
    "type": "postgres",
    "dsn": "postgres://user:pass@localhost:5432/fastclaw?sslmode=disable",
    "autoMigrate": true
  }
}
```

Supported types: `"file"` (default), `"postgres"`, `"sqlite"`

### LLM Providers

```json
{
  "providers": {
    "openai": {
      "apiKey": "${OPENAI_API_KEY}",
      "apiBase": "https://api.openai.com/v1",
      "apiType": "openai"
    },
    "deepseek": {
      "apiKey": "${DEEPSEEK_API_KEY}",
      "apiBase": "https://api.deepseek.com/v1",
      "apiType": "openai"
    }
  }
}
```

### Multi-Agent Example

```json
{
  "agents": {
    "defaults": {
      "model": "deepseek-v4-pro",
      "maxTokens": 8192,
      "temperature": 0.7,
      "maxToolIterations": 20,
      "thinking": "medium"
    },
    "list": [
      {
        "id": "coder",
        "model": "deepseek-v4-pro",
        "skills": ["debugging", "tdd"],
        "tools": ["exec", "read_file", "write_file", "db_query"],
        "thinking": "high"
      },
      {
        "id": "analyst",
        "model": "deepseek-v4-flash",
        "skills": ["financial-modeling"],
        "thinking": "medium"
      }
    ]
  },
  "bindings": [
    { "agentId": "coder", "match": { "channel": "telegram", "peer": { "kind": "dm" } } }
  ]
}
```

### Team Configuration

```json
{
  "teams": {
    "research-team": {
      "agents": ["analyst", "coder"],
      "defaultAgent": "analyst",
      "groupBehavior": "mention-only"
    }
  }
}
```

### MCP Servers

```json
{
  "mcpServers": {
    "brave-search": {
      "type": "http",
      "url": "https://api.search.brave.com/mcp",
      "headers": { "Authorization": "Bearer ${BRAVE_API_KEY}" }
    },
    "local-tool": {
      "type": "stdio",
      "command": "python",
      "args": ["-m", "my_mcp_server"],
      "env": { "API_KEY": "${MY_API_KEY}" }
    }
  }
}
```

## 🔌 Plugin System

Extend FastClaw with plugins in any language. Plugins communicate via JSON-RPC 2.0 over stdin/stdout, running as isolated subprocesses.

**Plugin types:** `channel` · `tool` · `provider` · `hook`

```bash
# Install from FastClaw Hub
fastclaw plugins install telegram

# Install from GitHub
fastclaw plugins install github.com/user/fastclaw-plugin

# Bridge an OpenClaw tool plugin
fastclaw plugins install @ollama/openclaw-web-search
```

Official plugins are in the [`plugins/`](plugins/) directory. Community plugins are indexed at [FastClaw Hub](https://github.com/fastclaw-ai/fastclaw-hub).

### Plugin Manifest (`plugin.json`)

```json
{
  "id": "my-plugin",
  "name": "My Plugin",
  "version": "1.0.0",
  "type": "tool",
  "command": "python plugin.py",
  "capabilities": ["tool"],
  "config": {
    "apiKey": { "type": "string", "required": true }
  }
}
```

### Community Plugins

| Plugin | Type | Description |
|--------|------|-------------|
| [fastclaw-plugin-weixin](https://github.com/videGavin/fastclaw-plugin-weixin) | Channel | WeChat messaging via ilink bot API (Node.js) |
| [fastclaw-mattermost-plugin](https://github.com/cornking2020/fastclaw-mattermost-plugin) | Channel | Mattermost messaging via WebSocket API (Go) |

## 🖥 Web Dashboard

Full management UI at `http://localhost:18953`:

| Page | What you can do |
|------|----------------|
| **Overview** | Gateway status, stats, quick actions |
| **Chat** | Talk to your agents in the browser |
| **Agents** | Create, edit, delete agents; edit SOUL.md, MEMORY.md |
| **Models** | Manage LLM providers and default model |
| **Skills** | View and manage installed skills |
| **Plugins** | Enable/disable plugins, edit config |
| **Channels** | Channel status and configuration |
| **Cron Jobs** | Create and manage scheduled tasks |
| **Apps** | Quick-launch companion dashboards |
| **Settings** | Storage backend, webhook, gateway config |

## 🔗 API

FastClaw exposes an OpenAI-compatible API for programmatic access:

```bash
# Chat with an agent (SSE streaming)
curl -X POST http://localhost:18953/v1/chat/completions \
  -H "Authorization: Bearer $TOKEN" \
  -H "Content-Type: application/json" \
  -d '{"model":"auto","messages":[{"role":"user","content":"hello"}],"stream":true}'

# List agents
curl http://localhost:18953/v1/agents -H "Authorization: Bearer $TOKEN"
```

### WebSocket (OpenClaw Compatible)

```javascript
const ws = new WebSocket('ws://localhost:18953/ws');
ws.send(JSON.stringify({ type: 'chat', message: 'hello' }));
```

## 🛠 CLI Reference

```bash
# Core
fastclaw                    # Start (setup wizard or gateway)
fastclaw gateway            # Start gateway explicitly
fastclaw version            # Version info
fastclaw doctor             # Check config health
fastclaw upgrade            # Update to latest

# Plugins
fastclaw plugins install NAME    # Install from Hub / GitHub / npm
fastclaw plugins list            # List installed plugins
fastclaw plugins remove ID       # Remove a plugin

# Agents
fastclaw agent create <name> # Create new agent with workspace files
fastclaw agent list           # List agents

# Skills
fastclaw skill list           # List installed skills

# Database
fastclaw migrate              # Migrate local JSONL sessions → PostgreSQL

# Daemon
fastclaw daemon start --port 18953   # Start as daemon
fastclaw daemon stop                 # Stop daemon
fastclaw daemon restart              # Restart daemon
fastclaw daemon status               # Check daemon status

# Administration
fastclaw session ...          # Session management
fastclaw provider ...         # Provider management
fastclaw sandbox ...          # Sandbox configuration
fastclaw policy ...           # Policy management
fastclaw backup               # Backup data
```

## 🛠 Development

```bash
git clone https://github.com/fastclaw-ai/fastclaw.git
cd fastclaw

# Install dependencies
./setup.sh          # Install PostgreSQL, Go toolchain, build, configure
# Or manually:
make build          # Build binary
make build-web      # Build web UI

# Run
make dev            # Dev mode with hot reload
make release-local  # Build all platforms
make test           # Run tests

# Management
./fastclaw-manager.sh deploy   # Build + install + restart
./fastclaw-manager.sh status   # Service status + logs
./fastclaw-manager.sh logs     # Tail live logs
```

### Workspace Structure

```
~/.fastclaw/
├── fastclaw.json           # Main configuration
├── fastclaw.pid            # Daemon PID file
├── logs/
│   └── gateway.log         # Daemon log
├── agents/
│   └── {agent-id}/
│       └── agent/
│           ├── agent.json          # Per-agent config
│           ├── SOUL.md             # Personality & behavior
│           ├── AGENTS.md           # Capabilities description
│           ├── TOOLS.md            # Tool usage instructions
│           ├── USER.md             # User information
│           ├── BOOTSTRAP.md        # Startup instructions
│           ├── HEARTBEAT.md        # Periodic task instructions
│           ├── MEMORY.md           # Long-term memory
│           ├── IDENTITY.md         # Identity information
│           ├── memory/             # Memory log directory
│           ├── sessions/           # Conversation history (JSONL)
│           └── skills/             # Agent-specific skills
└── plugins/                # Installed plugins
```

### PostgreSQL Schema

When `storage.type = "postgres"` is configured, the following tables are created:

| Table | Purpose |
|-------|---------|
| `configs` | Global configuration (per-tenant) |
| `agents` | Agent definitions and config |
| `workspace_files` | SOUL.md, MEMORY.md, etc. |
| `sessions` | Conversation history (JSONB) |
| `memory_logs` | Searchable memory entries |
| `cron_jobs` | Scheduled tasks with distributed locking |
| `chat_tasks` | Async web chat task tracking |
| `memories` | Fact store with pgvector embeddings |
| `research_data` | Structured research data with vector search |
| `schema_registry` | Metadata for agent-created tables |

## Contributing

Contributions welcome. FastClaw's strength is simplicity — keep it that way.

- **Core framework & official plugins** — contribute to this repo
- **Community plugins** — create your own repo, submit to [FastClaw Hub](https://github.com/fastclaw-ai/fastclaw-hub) index

## License

[MIT](LICENSE)

---

<div align="center">
  <sub>Built with ⚡ by the FastClaw community</sub>
</div>