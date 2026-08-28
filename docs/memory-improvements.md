# Memory 系统优化路线图（部分已被 MCP 方向取代）

> **状态:部分交付,治理项已迁移。** 更新于 2026-08-28。
>
> 本文是早期对 **in-agent 自动记忆**(`internal/agent/memory.go` 的 AutoPersist/Heartbeat)的 review。此后记忆策略发生转向 —— 见 [STATUS.md §5](STATUS.md#5-目标变了的地方给未来的自己):从"把 in-agent 自动抽取做强"转向"对外暴露共享 MCP 记忆池,读积极写保守"。本文各项的现状:
> - **P0-1 / P0-2 已交付**(读路径已接通、embedding pipeline 已接),下方逐项标注。
> - **P1-1 / P1-2 / P2(去重压缩、统一更新、PII 拦截)未做,且已重新定位**:它们不该只服务 in-agent 路径,而应下沉成"服务两条写入路径的服务端护栏"。设计与 TODO 收敛进 [memory-governance.md](memory-governance.md)。
> - 关键变化:现在 **in-agent AutoPersist 与 MCP `memory_write` 两个写入方共写同一张 `memories` 表**,下方"现状基线"的双层存储图已不完整,以 [memory-mcp.md](memory-mcp.md) + [STATUS.md](STATUS.md) 为准。
>
> 起源:用户提问 "agent 有自己的 memory 吗、什么时候更新、每个 session 都会改吗",进而引出对 [`internal/agent/memory.go`](../internal/agent/memory.go) 在引入 PostgreSQL/pgvector 之后是否还合理的 review。本文记录 review 中识别出的可改进项。

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

### 🔴 P0: 让 DB 真正参与读路径（兑现 pgvector 价值） — ✅ 已交付

> 实际落点是 **`BuildSystemPromptWithMemory`**(`context.go`)→ `searchRelevantMemory`,而非下方设想的 `BuildSystemPromptSections`(它仍传 nil,按旧函数名核对易误判"没做")。

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

### 🔴 P0: 接入 embedding pipeline — ✅ 已交付

> `AutoPersistMemory` 落库前已调 `embedFactSafe`,失败不阻塞(warn 后 embedding=nil 落库)。

**问题**(已解决,存档):[`AutoPersistMemory`](../internal/agent/memory.go:331) 调用 `Insert` 时 embedding 永远传 `nil`，`memories.embedding` 字段从未被填充。`SearchSemantic` 在 embedding 为 nil 时回退到 ILIKE，这退化成了"带 DB 的 grep"。

**方案**：
- 在 `provider.Provider` 接口或新建 `EmbeddingProvider` 接口加 `Embed(text) ([]float32, error)`
- 至少支持 OpenAI `text-embedding-3-small`、本地 Ollama `nomic-embed-text` 两个起点
- AutoPersist 落库前先 embed，失败不阻塞（记录 warn 后落库 embedding=nil）

### 🟡 P1: MEMORY.md 加去重 + 压缩 — ⬜ 未做 · 已迁移至 [memory-governance.md](memory-governance.md)

> 重新定位:去重不该只服务 in-agent 文件路径,而应作为 **Layer 1 服务端护栏**(近似去重 `InsertOrMerge`)服务两个写入方。"rewrite & compress" 归入 Layer 2 自动层的跨会话归并。下方为原始分析,存档。

**问题**：
- [`AutoPersistMemory`](../internal/agent/memory.go:307-326) 每次都是纯追加，没有任何去重逻辑
- [`ReviewAndUpdateMemory`](../internal/agent/memory.go:163) 只有弱 `strings.Contains` 判断
- 整个 MEMORY.md 全量注入 system prompt → 用得越久 token 成本线性增长

**方案**：
- AutoPersist 抽取后调用 `MEMORY_FACT_DEDUPE` 提示词，让 LLM 比对 currentMemory 决定是否丢弃
- 文件超过阈值（推荐 4KB）时触发一次 LLM "rewrite & compress"：保留摘要核心档案，明细事实下沉到 DB
- 提供 CLI/UI 入口让用户手动 trigger 压缩

### 🟡 P1: 统一两条更新路径 — ⬜ 未做 · 已迁移至 [memory-governance.md](memory-governance.md)

> 注意语境已变:本项指的是 **in-agent 内部** AutoPersist 与 Heartbeat 两条路径的分叉;而现在还多了 **MCP `memory_write`** 这条外部写入路径。治理文档 Phase 2 收编了本项(统一走 LLM 抽取 + 双写)。下方为原始分析,存档。

**问题**：AutoPersist 走 LLM + 双写，Heartbeat 走 关键词 + 仅文件，方法论不一致 + 写入位置不一致 → 两边都会写 MEMORY.md，但 DB 只看到一半数据。

**方案**（二选一）：
1. **砍掉关键词路径**：直接删 [`ReviewAndUpdateMemory`](../internal/agent/memory.go:134-192)；Heartbeat 改为复用 AutoPersist 的 LLM 抽取流程
2. **保留但统一接口**：让 Heartbeat 也走 `AutoPersistMemory` 同款 prompt，并双写 DB

推荐方案 1，关键词匹配在 LLM 时代价值有限。

### 🟢 P2: agentID 改为稳定 ID — ⬜ 未做 · 已迁移至 [memory-governance.md](memory-governance.md)（Phase 2）

**问题**：[`SetPGStore(store, a.name)`](../internal/agent/memory.go:441-442) 用 agent name 当 DB 主键。改名 / 同名跨 workspace 都会破坏 DB 一致性。

**方案**：在 workspace 元数据（如 `agent.json` 或新建 `.agent-id`）持久化一个稳定 UUID，`SetPGStore` 用它当 agentID。

### 🟢 P2: 隐私扫描升级为可拦截 — ⬜ 未做 · 已迁移至 [memory-governance.md](memory-governance.md)（Layer 1）

> 同样重新定位为服务端护栏:`memory.privacy.action: warn|redact|block`,`AutoPersistMemory` 的 `SaveMemoryWithScan` 与 MCP `registerMemoryWrite` 接同一前置钩子。`privacy.Scrub` 现只用在送 LLM 的消息上、**没用在记忆写入点**。

**问题**：[`SaveMemoryWithScan`](../internal/agent/memory.go:205-216) 检测到 PII 只 log warn，照样写。LLM 抽取出的 PII 同时落盘和入库，DB 里删除起来比文件麻烦得多。

**方案**：加配置 `memory.privacy.action: warn | redact | block`，redact 模式下用 [`privacy.Scrub`](../internal/privacy/scrub.go) 脱敏后再写。

### 🟢 P3: 清理 dead code — ✅ 已交付

> `PGMemoryStore` 已被实际引用,不兼容的签名已移除。下方为原始分析,存档。

[`PGMemoryStore` interface](../internal/agent/memory.go:20-23) 全工程零引用，且 `LoadAll` 返回类型 `[]interface{ GetContent() string }` 与 [`MemoryStore.LoadAll`](../internal/store/pg/memory_store.go:128-141) 真正返回的 `[]MemoryRecord` 根本不兼容。删除或修正后再保留。

---

## 实现优先级建议(已部分兑现,存档)

```
Phase 1 (打通读路径)：P0-1 + P0-2   —— ✅ 已交付
  └ 收益：pgvector 真正起作用，MEMORY.md 不再线性膨胀
Phase 2 (运维质量)：P1-1 (去重压缩) + P1-2 (统一更新)   —— ⬜ 未做,迁移至 memory-governance.md
  └ 收益：长期运行下的成本与一致性
Phase 3 (打磨)：P2 + P3   —— P3 ✅ 已交付;P2(稳定 UUID / PII 拦截)⬜ 迁移至 memory-governance.md
```

> 后续治理路线不再按本文的 P 编号推进,改以 [memory-governance.md](memory-governance.md) 的三层(Layer 0 软引导 / Layer 1 服务端护栏 / Layer 2 自动层)+ 分阶段(Phase 1 护栏 / Phase 2 一致性 / Phase 3 自动层)为准。

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
