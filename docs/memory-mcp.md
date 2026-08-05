# Pawnix Memory MCP (`pawnix mcp`)

把 claw-os 变成一个**跨会话、跨工具的持久记忆后端**：通过 MCP (stdio) 暴露
PostgreSQL + pgvector 记忆池，供 Claude Code / ChatGPT(Codex CLI) / hermes / codewiz 协作使用。

- **长 session 记忆永久保留** —— 记忆存在 Postgres，不随进程/会话销毁。
- **每轮检索** —— 调用模型每轮开对话先 `memory_search`（靠工具描述强约束，读积极）。
- **多方协作** —— 每个工具各自 spawn 一个 `pawnix mcp` 子进程，共享同一个记忆池
  （固定 `agent_id = shared`），provenance 靠每条记忆的 `source:*` 标签区分。
- **只吃 stdio** —— `pawnix mcp` 是本地 stdio 子进程，任何支持 stdio MCP server 的
  CLI/编辑器都能接（Claude Code、Codex CLI、hermes、codewiz…）。只支持远程
  HTTP/SSE 的宿主（如 ChatGPT 桌面/网页版连接器）需另加 HTTP 包装层，暂不覆盖。

## 架构

```
Claude Code ─┐
Codex CLI   ─┤
hermes      ─┼─ spawn ─→ `pawnix mcp` (stdio, JSON-RPC 2.0)
codewiz     ─┘                │
                              ↓
                     PostgreSQL + pgvector
                     (memories 表, 共享池 agent_id=shared)
```

设计原则：**读积极、写保守**。判断"该不该写"的权力留在调用模型侧（靠工具描述约束），
claw-os 不自作主张——避开当年 AutoPersist 自主判断写出"46 条垃圾记忆"的坑。

## 前提

`~/.pawnix/config`（JSON）需要：

```json
{
  "storage": { "type": "postgres", "dsn": "postgres://user:pass@localhost:5432/pawnix?sslmode=disable" },
  "memory":  { "embedModel": "openai/text-embedding-3-small" },
  "providers": {
    "openai": { "apiKey": "sk-...", "apiBase": "https://api.openai.com/v1" }
  }
}
```

- `storage.type` 必须是 `postgres` 且 DSN 有效（记忆后端依赖 pgvector）。
- `memory.embedModel` 未设时默认 `openai/text-embedding-3-small`（对应 `vector(1536)` 列）。
- embedder 构不出来（provider 没配等）时，检索自动降级为关键词 + 时间排序，不会报错崩溃。

## 构建

```bash
make dev-build      # 快速本地循环（复用已构建的前端），产出 bin/pawnix
# 或
make build          # 完整发布构建（含前端 embed）
```

## 手测 JSON-RPC 回路

```bash
./scripts/mcp-smoke-test.sh
```

它把 4 条请求喂给 `pawnix mcp` 并打印响应：`initialize` → `tools/list` →
`memory_stats` → `memory_search`。也可手动：

```bash
printf '%s\n' \
  '{"jsonrpc":"2.0","id":1,"method":"initialize","params":{}}' \
  '{"jsonrpc":"2.0","id":2,"method":"tools/list","params":{}}' \
  '{"jsonrpc":"2.0","id":3,"method":"tools/call","params":{"name":"memory_search","arguments":{"query":"用户偏好","limit":5}}}' \
  | ./bin/pawnix mcp
```

## 接入配置

前提：`command` 用 `pawnix` 需要它在 PATH 上（`make install` 或 `pawnix-manager.sh`
一键部署会装到全局）。**没装到全局就用 `bin/pawnix` 的绝对路径**，例如：

```
/Users/liuyang42/workspace/personal/claw-os/bin/pawnix
```

下面片段统一用 `pawnix` 占位；未 `make install` 时把它整段替换成上面的绝对路径。

`--source` 让每个工具写入的记忆带上自己的来源标签（`source:<tool>`），配置时一并加上，
四个工具各用一个：`claude-code` / `codex` / `hermes` / `codewiz`。

### Claude Code（`~/.claude.json` 或项目 `.mcp.json`）

```json
{
  "mcpServers": {
    "memory": { "command": "pawnix", "args": ["mcp", "--source", "claude-code"] }
  }
}
```

或用 CLI 一行加：

```bash
claude mcp add memory -- pawnix mcp --source claude-code
```

### ChatGPT = Codex CLI（`~/.codex/config.toml`）

> 说明：ChatGPT 桌面/网页版的连接器只支持远程 HTTP/SSE，接不了 stdio 的 `pawnix mcp`。
> 这里对接的是 OpenAI 的 **Codex CLI**（原生支持 stdio MCP server）。

```toml
[mcp_servers.memory]
command = "pawnix"
args = ["mcp", "--source", "codex"]
```

### hermes / claw-os（config `mcpServers`）

```json
{
  "mcpServers": {
    "memory": { "type": "stdio", "command": "pawnix", "args": ["mcp", "--source", "hermes"] }
  }
}
```

### codewiz（codex 类 CLI，`config.toml`）

codewiz 与 Codex CLI 同属 stdio MCP 宿主，配置格式一致，只是来源标签换成 `codewiz`：

```toml
[mcp_servers.memory]
command = "pawnix"
args = ["mcp", "--source", "codewiz"]
```

> 若 codewiz 用 JSON 而非 TOML 配置，改成与 Claude Code 相同的 `mcpServers` 结构、
> `--source codewiz` 即可。

## 工具

| 工具 | 作用 | 何时调用 |
|------|------|----------|
| `memory_search` | 检索记忆池（语义 + 降级关键词） | **每轮开对话就查**，只要可能涉及过往上下文（读积极） |
| `memory_write`  | 写入一条长期记忆 | **只存跨会话重要信息**，拿不准先问用户（写保守） |
| `memory_stats`  | 记忆库健康状况（总数/嵌入覆盖/近24h） | 排查问题时才用 |

`--agent <id>` 可覆盖默认共享池，只读/只写某一个 agent 的记忆。

`--source <tool>` 把来源标签（`source:<tool>`）自动打进每条**写入**的记忆，用于区分
provenance。来源在启动时定死（模型无法可靠自知运行在哪个工具里），不由模型自填。各工具各自配：

```
claude-code  → pawnix mcp --source claude-code
codex        → pawnix mcp --source codex     # ChatGPT/Codex CLI
hermes       → pawnix mcp --source hermes
codewiz      → pawnix mcp --source codewiz
```

### `memory_write` 参数

| 参数 | 必填 | 说明 |
|------|------|------|
| `content` | 是 | 记忆正文，写清楚"是什么/为什么"，让未来会话能独立看懂 |
| `kind` | 否 | `fact`=事实 / `user_note`=用户偏好或叮嘱 / `report`=阶段结论，默认 `user_note` |

写入时若 embedder 可用会自动生成向量嵌入；构不出来则降级为无嵌入（后续靠关键词检索）。
description 里对"何时该写"做了强约束（保守写、拿不准先问用户），避免污染记忆池。
