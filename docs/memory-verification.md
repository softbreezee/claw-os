# Memory 读写路径端到端验证

> 用途：在每次 release 前 / 改动 memory 相关代码后，确认 pgvector 语义检索真的进入了生产路径，而不是"代码看起来对、运行起来 embedding 列全是 NULL"。
>
> 上次验证：⬜ 待填日期 / 验证人

## 为什么需要这份文档

代码层面 P0-1（读路径接入）和 P0-2（embedding pipeline）已经实现，但有 5 种 silent-failure 模式会让 agent 表现得"还是不记得事"：

1. agent 没配 `embedModel` —— `searchRelevantMemory` 返回 nil，无任何报错
2. embed provider 没在 registry 里注册 —— 同上
3. embed API key 错 —— 写入和检索全失败，但日志只是 warn
4. PG 没启 pgvector extension —— `CREATE TABLE memories` 失败，整个 PG 不工作
5. agent name 改过 —— 旧 memories 检索不到（agent_id 不匹配），看起来像没记住

这份验证流程会逐个排除这些失败模式。

> v0.3 已经落地的自动加固（不再完全依赖人工跑这份验证才能发现）：
>
> - **启动期 pgvector readiness probe**（`pg.MemoryStore.VerifyVectorReady`，gateway 启动时跑一次）—— 直接把失败模式 4 在启动日志里大写出来，并附 `CREATE EXTENSION vector;` 提示。
> - **Heartbeat 每 30 分钟跑 memory health 巡检**（`pg.MemoryStore.HealthStats`）—— 近 24h embedding 覆盖率 < 50% 且样本数 ≥ 3 时 warn-level 告警，覆盖失败模式 1 / 2 / 3 / 5。
> - **memory health 单测**（`internal/agent/memory_health_test.go`）—— 锁 RecentCoveragePct 的边界（空窗 / 全覆盖 / 部分截断 / 旧数据）。
>
> 因此这份手动验证流程主要在两种情况下还需要跑：(a) 端到端 demo 前的最后兜底；(b) 改动了 embedding 路径核心代码（provider.Embed / AutoPersistMemory / SearchSemantic）时。

---

## 前置条件

- Pawnix 源码 + `make build` 成功
- Docker（用来跑 PG，最省事）
- 一个能用的 OpenAI API key（或本地 ollama，见 [备选方案](#备选本地-ollama-embed)）

## 验证流程

### Step 1：起 PG with pgvector

```bash
docker run -d --name pawnix-pg-verify \
  -e POSTGRES_PASSWORD=pawnix \
  -e POSTGRES_DB=pawnix \
  -p 5433:5432 \
  pgvector/pgvector:pg16

# 等 5 秒让 PG 起来
sleep 5

# 确认 pgvector extension 装好
docker exec pawnix-pg-verify psql -U postgres -d pawnix \
  -c "CREATE EXTENSION IF NOT EXISTS vector;"
# 期望输出: CREATE EXTENSION
```

> **失败模式 4 检查点**：如果这步报错，pgvector 没装好，后面全部白费。务必看到 `CREATE EXTENSION` 才往下走。

### Step 2：配置 Pawnix 用这个 PG

编辑 `~/.pawnix/pawnix.json`：

```json
{
  "providers": {
    "openai": {
      "apiKey": "sk-...",
      "apiBase": "https://api.openai.com/v1",
      "apiType": "openai"
    }
  },
  "agents": {
    "defaults": {
      "model": "openai/gpt-4o-mini",
      "embedModel": "openai/text-embedding-3-small"
    }
  },
  "memory": {
    "embedModel": "openai/text-embedding-3-small",
    "semanticTopK": 5,
    "semanticByteCap": 1024,
    "autoPersist": {
      "enabled": true,
      "everyNTurns": 3
    }
  },
  "storage": {
    "type": "postgres",
    "dsn": "postgres://postgres:pawnix@localhost:5433/pawnix?sslmode=disable",
    "autoMigrate": true
  }
}
```

> **关键配置**：
> - `everyNTurns: 3` —— 验证时调小，3 轮就触发 AutoPersist，不用聊半小时
> - `embedModel` 必须包含 `provider/` 前缀，registry 才能路由到 OpenAI

### Step 3：启 gateway 并观察启动日志

```bash
./build/pawnix gateway 2>&1 | tee /tmp/pawnix-verify.log
```

启动日志里搜这几行（必须都出现）：

```bash
grep -E "pg.*memories|pg.*connected|autoMigrate" /tmp/pawnix-verify.log
```

期望看到：
- `pg: connected` 或类似 PG 连接成功
- `memories` 表建好（看 autoMigrate 日志）

> **失败模式 4 兜底**：如果这里看到 `relation "memories" does not exist`，说明 autoMigrate 没跑成功 —— 可能是 vector extension 没装。回到 Step 1。

### Step 4：触发 AutoPersist（写路径）

打开 web UI（默认 `http://localhost:18953/chat`），发给 agent **至少 6 轮**对话，每轮都说点"值得记住"的事实。例：

```
轮 1: 我叫 Alex，是一个独立开发者
轮 2: 我现在用 macOS，工作偏好是 minimalism、不喜欢花哨的 UI
轮 3: 我手头在做一个叫 Pawnix 的项目，定位是 AI-Native personal OS
轮 4: 我最讨厌的反模式是 "demo 跑通就上线"，永远应该先写测试
轮 5: 联系我可以发邮件到 alex@example.com（注：这条用来验证 PII 扫描日志）
轮 6: 今天先聊到这吧
```

`everyNTurns=3` 意味着第 3、6 轮会触发 AutoPersist。

### Step 5：跑验证脚本

```bash
./scripts/memory-verify.sh
```

脚本会输出 7 个关键指标。**每一项都必须 ✅**，任何一个 ❌ 都意味着读写路径有问题，下文有对应排查表。

### Step 6：检索路径手动验证

开一个**新的 chat session**（很重要，不能复用上面那个，否则命中的是 session history 而不是 memory），问一个跟前面相关的问题：

```
我之前跟你说过我做的什么项目？它的定位是什么？
```

期望：
- agent 答出 "Pawnix" + "AI-Native personal OS"（即使你开了新 session）
- `/tmp/pawnix-verify.log` 里出现一行 `memory search: hits` 且 `count >= 1`

```bash
grep "memory search: hits" /tmp/pawnix-verify.log | tail -5
```

### Step 7：UI 上看 Relevant Memory section

在 chat UI 里点 system prompt preview（具体入口取决于当前 web 实现，通常在 chat header 的 Context Usage 圆环旁边），确认能看到一节标题为 `# Relevant Memory` 的内容，里面列着前面对话沉淀下来的事实。

如果 UI 没暴露这个入口，最直接的方式：
```bash
grep -A 20 "Relevant Memory" /tmp/pawnix-verify.log
```

应该能在 `system prompt` debug 日志里看到这一节。

---

## 验证脚本期望输出

`./scripts/memory-verify.sh` 会逐项检查并输出类似：

```
============================================
 Pawnix Memory Pipeline Verification
============================================

[1/7] PG container reachable ......................... ✅
[2/7] vector extension installed ..................... ✅
[3/7] memories table exists .......................... ✅
[4/7] memories rows count ............................ ✅ (8 rows)
[5/7] embedding column populated ..................... ✅ (8/8 non-NULL)
[6/7] semantic search returns results ................ ✅ (top hit: cosine 0.23)
[7/7] log shows "memory search: hits" ................ ✅ (3 occurrences)

PIPELINE STATUS: HEALTHY ✨
```

任何一项 ❌ 时脚本会输出对应的修复建议。

---

## 排查表（每个失败模式 → 怎么修）

| 现象 | 根因 | 修法 |
|------|------|------|
| Step 5 [4/7] `0 rows` | AutoPersist 没触发 | 看 `everyNTurns` 配置；轮数够了吗；日志里搜 `auto-persist:` |
| Step 5 [5/7] `0/N non-NULL` | embedding pipeline 写路径断了 | 日志搜 `auto-persist: embed failed`；常见是 `embedModel` 路由不到 OpenAI provider |
| Step 5 [6/7] `0 results` | 数据进去了但检索查不出 | agent_id 不一致？看 `SELECT DISTINCT agent_id FROM memories` 和当前 agent name |
| Step 5 [7/7] `0 occurrences` | 读路径没接通 | `pg.PGStore() == nil`？看 `gateway/memstore_adapter.go` 是否真的注入到 agent.Memory |
| Step 6 答非所问 | 检索质量差 | 看 cosine distance（脚本会打印 top hit 的距离），>0.5 说明 query 和 facts 语义远 |

---

## 备选：本地 Ollama embed

不想用 OpenAI 的话：

```bash
ollama pull nomic-embed-text
```

config 改：
```json
"providers": {
  "ollama": { "apiBase": "http://localhost:11434/v1", "apiType": "openai", "apiKey": "ollama" }
},
"memory": { "embedModel": "ollama/nomic-embed-text" }
```

⚠️ **注意维度**：`nomic-embed-text` 是 768 维，但当前 schema 是 `vector(1536)`。需要改 schema 或换模型。建议验证阶段先用 OpenAI，跑通了再考虑本地化。

---

## 清理

```bash
docker rm -f pawnix-pg-verify
rm /tmp/pawnix-verify.log
```

---

## Changelog

- (待填) 首次验证通过 / 失败原因 / 修了什么
