# 记忆治理设计(Memory Governance)

> **状态:设计 + TODO,尚未实现。** 更新于 2026-08-28。
>
> 回答一个问题:**Pawnix 怎么决定"什么该记、什么不该记、要不要去重、要不要拦 PII"?**
> 现状是——只有一层"软引导"(工具描述),服务端零强制。本文把缺口讲清楚,给出三层设计和分阶段落地方案。
>
> 相关:[memory-mcp.md](memory-mcp.md)(MCP 后端)· [memory-improvements.md](memory-improvements.md)(本文收编了它的 P1-1/P1-2/P2)· [STATUS.md](STATUS.md)

---

## 1. 现状:只有"软引导"这一层

判断"该不该写记忆"的权力,现在**完全交给调用模型**,靠 `memory_write` 的工具描述(`cmd/pawnix/cmd_mcp.go` 里 `memoryWriteDesc`)去引导。判据本身写得不差:

**应当写**:用户明确说"记住 / 以后都这样 / 下次注意";稳定的用户偏好、长期项目背景、已确认的关键决策与事实;用户纠正过你、且这个纠正对以后仍适用。

**不要写**:本轮任务的临时上下文、中间结果、可从文件/工具重新查到的信息;尚未确认、还在讨论中的想法;敏感个人信息(身份证件/银行卡/健康/家庭住址),除非用户明确要求。

核心启发式:"写之前先自问——下次新开一个会话,这条信息还值得被记住吗?拿不准就先问用户。"

这一层是**软的**:模型读了描述**自愿遵守**,`memory_write` 收到什么就写什么。

---

## 2. 缺口:服务端零把关,而且现在有两个写入方

`memory_write` 只校验 `content` 非空 + `kind` 枚举合法,随后自动打上 `source:*` 标签、能嵌入就嵌入,然后**原样 INSERT**(`internal/store/pg/memory_store.go` 的 `Insert` 是一条没有 `ON CONFLICT` 的普通插入)。**没有去重、没有 PII 拦截、没有内容质量校验。**

更要紧的是:经过记忆策略转向,现在有**两个写入方共写同一张 `memories` 表**:

| 写入方 | 路径 | 现有约束 |
|--------|------|----------|
| in-agent AutoPersist | `internal/agent/memory.go` `AutoPersistMemory`(每 5 轮 LLM 抽取 + 双写 + 现已嵌入) | `SaveMemoryWithScan` 只 `Scan`+warn,不拦;纯追加无去重 |
| MCP `memory_write` | `cmd/pawnix/cmd_mcp.go` `registerMemoryWrite`(外部工具调用) | 仅 content/kind 校验,原样插入 |

两边都没有去重、PII 拦截、一致性保证。memory-improvements.md 当年为 in-agent 路径设计的 P1-1(去重压缩)、P2(隐私可拦截)从没实现——它们不该只服务 in-agent,而应下沉成**服务两条写入路径的共享护栏**。这正是本文要收编的东西。

---

## 3. 三层治理设计

```
Layer 0  软引导(已存在)      —— 工具描述,模型自愿遵守;成本 0,但不可靠
Layer 1  服务端护栏(建议先做)—— store 层,确定性强,两个写入方都过一遍
Layer 2  自动层(可选,后置)  —— 借鉴 Codex 的后台提炼/归并,重、需另配模型
```

### Layer 0 — 软引导(保留 + 调优)

保留现有工具描述的正例/反例。可持续调优,但不作为唯一防线。这层的价值是"让模型少写垃圾",不是"保证不写"。

### Layer 1 — 服务端护栏(确定性,建议优先)

不依赖模型,在 store 写入点统一施加。**两个写入方都调同一个 `WriteGuard`,一处实现两处受益。**

1. **近似去重(收编 P1-1)。** 写入前用新记忆的 embedding 在同 `agent_id`(+可选同 `source`)里做相似度检索,超阈值(如 cosine ≥ 0.92)则**更新/合并**已有条目而非新增,或直接跳过。落点:`memory_store.go` 加 `InsertOrMerge`,或在其上包一层 `WriteGuard`。无 embedding 时降级为 content 归一化后的精确/前缀去重。

2. **PII 拦截 / 脱敏(收编 P2)。** 新增配置 `memory.privacy.action: warn | redact | block`。`redact` 用现成的 `privacy.Scrub`(`internal/privacy/scrub.go`,目前只用在送 LLM 的消息上、**没用在记忆写入点**)脱敏后再写;`block` 直接拒写并回报原因。落点:`AutoPersistMemory` 的 `SaveMemoryWithScan` 与 MCP `registerMemoryWrite` 都接同一个前置钩子。

3. **内容/结构校验。** 长度上限(超长要么拒、要么摘要下沉);`kind` 合法性(已有);可选:拒绝明显是临时上下文的模式(如纯 URL、纯数字串)。

> Layer 1 是"硬"保证:即使模型判断失误,服务端兜底。工程量小、收益直接,应最先做。

### Layer 2 — 自动层(可选,借鉴 Codex,后置)

对标 ChatGPT/Codex 的本地记忆实现(见 [§5](#5-与-codex-的对照)),把"提炼"和"归并"自动化:

1. **后台空闲提炼**:不在每轮结束即写,等会话空闲足够久再在后台抽取,避免总结"还没干完"的工作。可复用 AutoPersist 的 LLM 抽取,改触发时机。
2. **跨会话归并(consolidation)**:定期把零散条目合并成稳定档案,明细下沉。对应 memory-improvements.md 想做的"rewrite & compress",也和现有 `consolidate-memory` skill 呼应。可为提炼/归并各配一个模型(`extract_model` / `consolidation_model`)。
3. **低配额跳过**:token/额度紧张时跳过后台生成,不与前台抢资源。
4. **外部上下文不入库**:带 MCP 工具调用 / web search 结果的会话,默认不进入自动记忆生成(这类内容噪声大、易过期)。Pawnix 的 `mcp_events` 已能标注来源与命中,可据此实现。

> Layer 2 重且需要另配模型与调度,放在护栏之后。**它是"自动多写一些",与 Layer 1 的"拦住不该写的"正交,顺序上先做拦截、再谈自动。**

---

## 4. 分阶段落地

| 阶段 | 内容 | 触及文件 |
|------|------|----------|
| **Phase 1 — 护栏** | Layer 1 全部:`WriteGuard`(去重 + PII action + 校验),两个写入方接同一钩子 | `internal/store/pg/memory_store.go`(InsertOrMerge/WriteGuard)· `internal/config/config.go`(`memory.privacy.action` 等)· `internal/privacy/scrub.go`(复用 Scrub)· 调用点 `internal/agent/memory.go` + `cmd/pawnix/cmd_mcp.go` |
| **Phase 2 — 一致性** | 收编 P1-2(统一 Heartbeat/AutoPersist 两条更新路径,都走 LLM 抽取 + 双写)+ P2(agentID 改稳定 UUID) | `internal/agent/memory.go`(ReviewAndUpdateMemory / SetPGStore)· workspace 元数据存 UUID |
| **Phase 3 — 自动层** | Layer 2:后台空闲提炼 + 跨会话归并 + 低配额跳过 + 外部上下文不入库 | 新增后台调度;`extract_model`/`consolidation_model` 配置;`mcp_events` 来源判定 |

建议只承诺 **Phase 1**,它把"两个写入方共写、零护栏"这个当前最痛的洞补上,且不依赖模型行为。Phase 2/3 视记忆池规模与运行成本再排期。

---

## 5. 与 Codex 的对照

| 维度 | Codex/ChatGPT 本地记忆 | Pawnix 现状 | Pawnix 治理后 |
|------|------------------------|-------------|---------------|
| 写入方式 | LLM 后台隐式提炼,用户无感 | MCP 工具显式写,模型自律 | 显式写 + 服务端护栏兜底(Layer 1);可选后台自动(Layer 2) |
| 存储/共享 | 按 host 隔离的本地文件 `~/.codex/memories/` | **共享 Postgres 池**,`source:*` 区分来源(差异化优势) | 不变 |
| 去重/归并 | extract→consolidate 两段式 + 独立模型 | 无 | Layer 1 去重 + Layer 2 归并 |
| PII | 生成时自动脱敏 | 写入点不拦(仅 warn) | `memory.privacy.action` 可 redact/block |
| 外部上下文 | `disable_on_external_context` 可排除 | 无区分 | Layer 2:带 MCP/web 的会话不自动入库 |
| 低配额行为 | 余量低于阈值跳过后台生成 | 无后台生成 | Layer 2:低配额跳过 |

一句话:**Pawnix 的差异化是"共享池 + 显式可控",Codex 的强项是"自动省心"。治理路线不是抄 Codex 变全自动,而是先用服务端护栏补齐"硬保证",再按需引入自动层。**

---

## 6. 参考:关键落点索引

| 关注点 | 文件 | 说明 |
|--------|------|------|
| MCP 写入入口 | `cmd/pawnix/cmd_mcp.go` `registerMemoryWrite` | 加 WriteGuard 前置钩子 |
| 工具描述(软引导) | `cmd/pawnix/cmd_mcp.go` `memoryWriteDesc` | Layer 0,持续调优 |
| in-agent 写入入口 | `internal/agent/memory.go` `AutoPersistMemory` / `SaveMemoryWithScan` | 加同一 WriteGuard |
| DB 写入 | `internal/store/pg/memory_store.go` `Insert` | 加 `InsertOrMerge` / 去重 |
| PII 扫描/脱敏 | `internal/privacy/scrub.go` `Scan` / `Scrub` | Scrub 现只用于 LLM 消息,复用到写入点 |
| 配置 | `internal/config/config.go` `MemoryCfg` | 加 `privacy.action` 等 |
| 观测(外部上下文判定) | `internal/store/pg/event_store.go` `mcp_events` | Layer 2 判断哪些会话带外部工具 |
