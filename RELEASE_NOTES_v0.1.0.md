# Pawnix v0.1.0 — Foundation

> *The first release under the Pawnix name.* 🐾
>
> Pawnix is a self-hosted, AI-Native personal OS — a long-running runtime that connects any LLM to any channel, with multi-agent teams, durable memory, skills, plugins, and a full web dashboard, all in a single Go binary.

---

## Highlights

This is the **v0.1 Foundation** milestone — everything you need to run a real, always-on personal AI from your own machine.

### Multi-LLM, multi-agent

- Plug in **any OpenAI-compatible API**: OpenAI, Anthropic, DeepSeek, Gemini, GLM, Kimi, Groq, Ollama, OpenRouter, …
- **Per-call model override** — pick a different model per message without restarting the gateway
- **Multiple agents** with independent personas, memories, skills, and tool sets
- **Team `@mention` routing** so multiple agents can collaborate in the same channel

### Channels

- **Built-in:** Web chat, Telegram (multi-account, multi-bot), Discord, Slack (Socket Mode + Web API)
- **Plugin channel protocol:** drop in any other platform via JSON-RPC subprocess
- **New full CRUD UI:** Add / edit / delete channel accounts directly from the dashboard, bind them to agents, and Save & Restart in one click

### Memory & skills

- **Dual-layer memory:** `MEMORY.md` (persistent facts) + searchable conversation logs (FTS for SQLite, `pgvector` HNSW for PostgreSQL)
- **Skill system:** drop a `SKILL.md` and your agents can do new things; agents can also induce new skills from interaction patterns
- **Hot reload:** edit `SOUL.md` or config and changes take effect without restart

### Automation

- **Cron jobs** (cron expressions, intervals, one-shot timers) with distributed locking
- **Per-agent heartbeat** (each agent wakes every N minutes to consult `HEARTBEAT.md`)
- **Webhooks:** `POST /hooks` triggers from external systems
- **Async chat tasks:** chat continues in the background, survives page reloads, resumes via SSE

### Platform

- **Web dashboard at `:18953`:** Overview, Chat, Agents, Models, Skills, Plugins, Channels, Cron, Apps, Settings
- **OpenAI-compatible REST API** with SSE streaming
- **WebSocket gateway** (OpenClaw-compatible)
- **MCP client** (HTTP + stdio) for external tool servers
- **Multi-tenant** isolation built in
- **Single-binary daemon mode** with auto-restart

### Storage

- **File** (default, zero setup, JSON in `~/.fastclaw/`)
- **PostgreSQL** with `pgvector` for semantic memory + multi-tenant
- **SQLite** with FTS5
- **Auto-migration** on startup

### Daemon supervisor (new in this release)

- The daemon supervisor now distinguishes between three exit modes of the gateway:
  - **exit 0** → permanent stop
  - **exit 75** → "restart-on-purpose" (e.g. config saved via the **Channels** UI), supervisor relaunches immediately
  - **signal / non-zero** → crash, exponential backoff restart up to 10 attempts
- The **Save & Restart** button in the Channels page now works reliably, even though channels still don't support hot reload internally

### Security

- Sandbox `exec` (Docker)
- YAML policy engine for filesystem / network / tool / resource limits
- AES-256-GCM credential vault with env auto-discovery
- Optional PII scrubbing before LLM calls
- Tool-loop detection (breaks after 3 identical consecutive calls)

---

## Naming

This release renames the **product** from FastClaw to **Pawnix**. The rename is intentionally surface-only in v0.1:

- ✅ Web UI, CLI banners, agent identity prompt, README all say **Pawnix**
- ✅ New paw-print-in-kernel-ring logo and SVG favicon
- ⚠️ The **binary is still `fastclaw`**, the **config dir is still `~/.fastclaw/`**, the **Go module path is still `github.com/fastclaw-ai/fastclaw`**

This keeps existing installations working without any migration. A deeper rename (binary, config dir, module path, with a one-shot user-data migration script) is planned as a separate release.

---

## Roadmap

The next milestones are tracked in [README.md → Roadmap](./README.md#-roadmap):

- **v0.2 — The Memory OS** *(next)* — memory graph, time-aware retrieval, cross-agent memory pool, memory browser UI, skill auto-induction
- **v0.3 — Multi-modal & Voice** — voice in/out, screen understanding, file ingestion
- **v0.4 — Agent Marketplace** — Skills Hub, agent export/import, community catalog
- **v0.5 — Distributed Mesh** — multi-device sync, P2P session handover
- **v1.0 — Production Persona** — full audit, fine-grained permissions, HA, multi-user

---

## Install

```bash
curl -fsSL https://raw.githubusercontent.com/fastclaw-ai/fastclaw/main/install.sh | bash
fastclaw    # opens the setup wizard
```

Or from source:

```bash
git clone https://github.com/fastclaw-ai/fastclaw.git
cd fastclaw && make build
./bin/fastclaw
```

---

## Thanks

Pawnix evolved from the FastClaw codebase across dozens of commits adding multi-provider routing, prompt caching, model catalogs, the skills system, the chat task runtime, the channels CRUD UI, the daemon restart supervisor, and many more refinements. Thanks to everyone who contributed code, ideas, and bug reports.

🐾
