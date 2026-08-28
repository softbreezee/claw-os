# Pawnix Dashboard (web/)

The Pawnix web dashboard — a Next.js app that is **statically exported and embedded into the Go binary**, then served by the gateway at `http://localhost:18953`. There is no separate frontend server in production: `pawnix` ships one binary with this UI baked in.

## How it fits together

```
web/  ──(pnpm build)──▶  web/out/  ──(copy)──▶  internal/setup/web/  ──(go:embed)──▶  pawnix binary
   Next.js source          static export           embed dir                     served at :18953
```

- `next.config.ts` sets `output: 'export'` (+ `trailingSlash`, unoptimized images) so `pnpm build` emits a fully static site into `web/out/`.
- `internal/setup/embed.go` embeds that tree via `//go:embed all:web`; the gateway serves it directly.
- Because the UI is embedded at compile time, **editing `web/` has no effect until you rebuild and re-embed** (see below).

## Develop

```bash
# from repo root
cd web
pnpm install
pnpm dev        # hot-reload dev server at http://localhost:3000 (talks to a running gateway on :18953)
```

For the dev server, API calls hit the gateway on `:18953`, so run `pawnix` (or `make dev`) alongside it.

## Build & embed into the binary

Do this from the **repo root** (the Makefile drives it):

```bash
make build-web   # pnpm install --frozen-lockfile + pnpm build, then sync web/out → internal/setup/web
make embed-web   # faster: just sync an already-built web/out into internal/setup/web
make dev-build   # embed-web + rebuild the Go binary (the usual loop after editing web/)
```

> Common gotcha: run `pnpm build` (or `make build-web`) after editing `web/`, otherwise the binary keeps serving a stale `internal/setup/web`. See the repo memory note on the build env.

## Pages (`src/app/`)

Each dashboard page maps to a route directory. Current pages:

`overview` · `chat` · `inbox` · `agents` · `models` · `skills` · `plugins` · `channels` · `cron` · `apps` · `settings` · `mcp` · `memory` · `onboard`

Notable ones:

- **memory** — the MCP memory observability dashboard: per-source / per-session rollups of `mcp_events` (which tool searched or wrote, turn counts, topics, read-hit status). Backed by `GET /api/memory/usage` (`internal/setup/handlers_memevents.go`).
- **chat** — async, task-based chat with mid-run **Steer**, `/plan` PlanBubble, and the live `todo.md` progress panel (v0.3 Steering features).
- **onboard** — the first-run setup wizard (provider, storage, launch).

## Stack

Next.js (App Router) · React · Tailwind CSS + `@tailwindcss/typography` · shadcn / `@base-ui/react` components · `lucide-react` icons · `react-markdown` + `remark-gfm` for rendering. Static export only — no server components at runtime, no browser storage APIs.
