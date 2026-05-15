# Memory 系统优化路线图

> 起源：用户提问 "agent 有自己的 memory 吗、什么时候更新、每个 session 都会改吗"，进而引出对 [`internal/agent/memory.go`](../internal/agent/memory.go) 在引入 PostgreSQL/pgvector 之后是否还合理的 review。本文记录 review 中识别出的可改进项。

## 现状基线

每个 Agent 拥有独立 Memory，采用**双层存储**：

| 层 | 位置 | 数据形态 | 写入 | 读取 |
|----|------|---------|------|------|
| 文件 | `{workspace}/MEMORY.md` | 自由格式 markdown | ✅ AutoPersist + Heartbeat | ✅ 注入 system prompt |
| 文件 | `{workspace}/USER.md` | 自由格式 markdown | ✅ AutoPersist | ⚠️ 手工 / 偶尔 |
| 文件 | `{workspace}/HISTORY.md` | append-only 时间戳行 | ✅ AppendHistory | ✅ Heartbeat 扫描 |
| DB | `memories` 表 | (kind, content, embedding, tags) | ✅ AutoPersist 双写 | ❌ **生产链路零调用** |

更新触发：

| 触发器 | 频率 | 抽取方式 | 写文件 | 写 DB |
|--------|------|---------|--------|-------|
| AutoPersistMemory | 每 5 个对话回合 | LLM JSON 抽取 | ✅ | ✅ |
| Heartbeat → ReviewAndUpdateMemory | 每 30 分钟 | 关键词 grep | ✅ | ❌ |

> 详细函数索引见附录。

---

## 优化项

### 🔴 P0: 让 DB 真正参与读路径（兑现 pgvector 价值）

**问题**：
- [`MemoryStore.SearchSemantic`](../internal/store/pg/memory_store.go:63-84) / [`SearchKeyword`](../internal/store/pg/memory_store.go:87-125) / [`LoadAll`](../internal/store/pg/memory_store.go:128-141) 全工程零调用。
- [`BuildSystemPromptSections`](../internal/agent/context.go:118-125) 注入 Memory 时只读 `MEMORY.md` 文件。
- 结果：DB 是"只进不出的墓地"，pgvector 索引白建。

**方案**：把 system prompt 中的 Memory 注入分成两段：

```
# Long-term Memory (核心，固定注入)
- 来自 MEMORY.md 的精简核心档案，限 1-2KB

# Relevant Memory (动态检索，按当前用户消息检索)
- pgStore.SearchSemantic(agentID, embed(userMsg), top_k=5)
```

实现要点：
- 在 [`HandleMessage`](../internal/agent/loop.go:564-842) 拿到 user message 后、构 system prompt 前，调用 `SearchSemantic` 拿 top-k；拼到 prompt 的额外一节
- top_k 默认 5，加 token budget cap（例如 1KB）
- 当 pgStore 为空 / embedding 缺失时无缝退化为只读文件

### 🔴 P0: 接入 embedding pipeline

**问题**：[`AutoPersistMemory`](../internal/agent/memory.go:331) 调用 `Insert` 时 embedding 永远传 `nil`，`memories.embedding` 字段从未被填充。`SearchSemantic` 在 embedding 为 nil 时回退到 ILIKE，这退化成了"带 DB 的 grep"。

**方案**：
- 在 `provider.Provider` 接口或新建 `EmbeddingProvider` 接口加 `Embed(text) ([]float32, error)`
- 至少支持 OpenAI `text-embedding-3-small`、本地 Ollama `nomic-embed-text` 两个起点
- AutoPersist 落库前先 embed，失败不阻塞（记录 warn 后落库 embedding=nil）

### 🟡 P1: MEMORY.md 加去重 + 压缩

**问题**：
- [`AutoPersistMemory`](../internal/agent/memory.go:307-326) 每次都是纯追加，没有任何去重逻辑
- [`ReviewAndUpdateMemory`](../internal/agent/memory.go:163) 只有弱 `strings.Contains` 判断
- 整个 MEMORY.md 全量注入 system prompt → 用得越久 token 成本线性增长

**方案**：
- AutoPersist 抽取后调用 `MEMORY_FACT_DEDUPE` 提示词，让 LLM 比对 currentMemory 决定是否丢弃
- 文件超过阈值（推荐 4KB）时触发一次 LLM "rewrite & compress"：保留摘要核心档案，明细事实下沉到 DB
- 提供 CLI/UI 入口让用户手动 trigger 压缩

### 🟡 P1: 统一两条更新路径

**问题**：AutoPersist 走 LLM + 双写，Heartbeat 走 关键词 + 仅文件，方法论不一致 + 写入位置不一致 → 两边都会写 MEMORY.md，但 DB 只看到一半数据。

**方案**（二选一）：
1. **砍掉关键词路径**：直接删 [`ReviewAndUpdateMemory`](../internal/agent/memory.go:134-192)；Heartbeat 改为复用 AutoPersist 的 LLM 抽取流程
2. **保留但统一接口**：让 Heartbeat 也走 `AutoPersistMemory` 同款 prompt，并双写 DB

推荐方案 1，关键词匹配在 LLM 时代价值有限。

### 🟢 P2: agentID 改为稳定 ID

**问题**：[`SetPGStore(store, a.name)`](../internal/agent/memory.go:441-442) 用 agent name 当 DB 主键。改名 / 同名跨 workspace 都会破坏 DB 一致性。

**方案**：在 workspace 元数据（如 `agent.json` 或新建 `.agent-id`）持久化一个稳定 UUID，`SetPGStore` 用它当 agentID。

### 🟢 P2: 隐私扫描升级为可拦截

**问题**：[`SaveMemoryWithScan`](../internal/agent/memory.go:205-216) 检测到 PII 只 log warn，照样写。LLM 抽取出的 PII 同时落盘和入库，DB 里删除起来比文件麻烦得多。

**方案**：加配置 `memory.privacy.action: warn | redact | block`，redact 模式下用 [`privacy.Scrub`](../internal/privacy/scrub.go) 脱敏后再写。

### 🟢 P3: 清理 dead code

[`PGMemoryStore` interface](../internal/agent/memory.go:20-23) 全工程零引用，且 `LoadAll` 返回类型 `[]interface{ GetContent() string }` 与 [`MemoryStore.LoadAll`](../internal/store/pg/memory_store.go:128-141) 真正返回的 `[]MemoryRecord` 根本不兼容。删除或修正后再保留。

---

## 实现优先级建议

```
Phase 1 (打通读路径)：P0-1 + P0-2
  └ 收益：pgvector 真正起作用，MEMORY.md 不再线性膨胀
Phase 2 (运维质量)：P1-1 (去重压缩) + P1-2 (统一更新)
  └ 收益：长期运行下的成本与一致性
Phase 3 (打磨)：P2 + P3
```

---

## 附录：关键函数索引

| 函数 | 文件 | 功能 |
|------|------|------|
| [`LoadMemory`](../internal/agent/memory.go:91-97) | memory.go | 读 MEMORY.md |
| [`SaveMemory`](../internal/agent/memory.go:100-105) | memory.go | 写 MEMORY.md |
| [`AppendHistory`](../internal/agent/memory.go:108-121) | memory.go | 追加 HISTORY.md |
| [`ReviewAndUpdateMemory`](../internal/agent/memory.go:134-192) | memory.go | Heartbeat 扫描更新 |
| [`AutoPersistMemory`](../internal/agent/memory.go:246-369) | memory.go | LLM 驱动持久化 |
| [`runPostTurn`](../internal/agent/loop.go:946-985) | loop.go | 触发 AutoPersist 的 hook |
| [`Memory.SetPGStore`](../internal/agent/memory.go:37-42) | memory.go | 双写 DB 入口 |
| [`MemoryStore.Insert`](../internal/store/pg/memory_store.go:33-59) | pg/memory_store.go | DB 写入 |
| [`MemoryStore.SearchSemantic`](../internal/store/pg/memory_store.go:63-84) | pg/memory_store.go | **未被生产链路调用** |
| [`BuildSystemPromptSections`](../internal/agent/context.go:81-224) | context.go | system prompt 拼装入口 |

## 配置参考

| 字段 | 默认 | 位置 |
|------|------|------|
| `memory.autoPersist.enabled` | true | [`MemoryCfg`](../internal/config/config.go:118-121) |
| `memory.autoPersist.everyNTurns` | 5 | [`AutoPersistCfg`](../internal/config/config.go:124-128) |
| `memory.autoPersist.model` | (使用 agent 默认) | [`AutoPersistCfg`](../internal/config/config.go:124-128) |
| `heartbeat.intervalMinutes` | 30 | [`HeartbeatCfg`](../internal/config/config.go:32-35) |
