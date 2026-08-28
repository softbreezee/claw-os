# Pawnix 项目状态总览

> 更新于 **2026-08-28**。本文是"现在到底做了什么、什么没做、目标在哪里变了"的**单一入口**。
> 结论基于对代码的逐条核对(而非只看 git log 或旧计划文档),关键项附 `文件:行` 证据。
> 各专题文档的时效性见文末 [§6 文档地图](#6-文档地图)。

---

## 1. 一句话现状

Pawnix 已从"又一个 agent 运行时"重定位为 **memory-first**:核心是一个跨会话、跨工具共享的持久记忆后端(`pawnix mcp`),同一个二进制里仍带着完整的 agent 运行时。

- **v0.3 "Steering"(可打断 / 可围观 / 可托管)已交付**,三大特性都能用,但有若干欠账(见 [§4](#4-已知欠账--技术债))。
- **记忆策略发生转向**:从"把 in-agent 的 AutoPersist/Heartbeat 那套自动记忆做好"转向"对外暴露一个共享 MCP 记忆池,读积极写保守、判断权交回调用模型"。两条写入路径目前**并存并共写同一张 `memories` 表**。
- **当前最大的开放战线是"记忆治理"**:该不该写、去重、PII 拦截目前**只有一层软引导**(工具描述),服务端零强制。设计与 TODO 见 [docs/memory-governance.md](memory-governance.md)。

---

## 2. 里程碑状态

| 里程碑 | 主题 | 状态 |
|--------|------|------|
| v0.1 | Foundation(多 LLM / 多 agent / channels / skills / plugins / cron / 看板 / daemon) | ✅ 已交付 |
| v0.2 | Rebrand + OS 层(FastClaw→Pawnix、cron 账本统一、Inbox+通知、`notify`、`MyChatID`) | ✅ 已交付 |
| v0.3 | **Steering**(可打断 / 可围观 / 可托管) | ✅ 已交付(有欠账) |
| v0.4 | Multi-modal & Voice / `delegate_task` | 🟡 部分(`delegate_task` 已落地;图片附件 Stage-1 已落地;语音/屏幕理解未做) |
| v0.5 | Memory Graph & Marketplace | 🟡 部分(跨工具共享记忆池已落地;记忆图谱 / 时序检索 / 权限 / 浏览器 UI 未做) |

> 注:RELEASE_NOTES_v0.2.0 里当时预告的路线("v0.3 = Memory OS")**已与现实分叉**——v0.3 实际做成了 Steering,记忆能力则以 MCP 后端形态在 v0.5 线上提前落地。以 README 与本文为准。

---

## 3. 逐条实现核对

图例:✅ 已交付 · 🟡 部分 · ⬜ 未做 · ⚙️ 刻意不做

### 3.1 v0.3 Steering

| # | 目标 | 状态 | 证据 / 备注 |
|---|------|------|-------------|
| 1 | `bus.InboundMessage.Origin` 扩展枚举(heartbeat/subagent/goal_context/user_steer) | ✅ | `internal/bus/bus.go:24-27` |
| 2 | Skills 按需加载 + section token 预算 | ✅ | `load_skill` 工具 `tools/load_skill.go`;预算 `SkillsSummaryBudgetTokens=2000` `skills.go` |
| 3 | 可打断 · mid-run steering | ✅ | `drainSteerIntoMessages` 每轮 ReAct 头部 drain(`loop.go`);路由是 **`POST /api/chat/tasks/{id}/steer`**(非文档旧写的 `/api/chat/steer`) |
| 4 | 可围观 · `/plan` | 🟡 | `/plan` 出可审方案已通;但**不自己写 `todo.md`**,只是 nudge 模型下一步去写 |
| 5 | 可围观 · todo 进度面板 | 🟡 | Web `TodoPanel` 轮询 `todo.md` 已通;但**无专门 `todo_write/todo_check` 工具**,靠通用 `write_file`;plan nudge 里引用的 `update_todo` **是不存在的工具** |
| 6 | 可托管 · `/goal` MVP | 🟡 | goal 包 / `agent_goals` 表 / `/goal` 命令族 / PostTurn 续跑 continuation 均已通;**但 token 预算守护未接线**——`FoldUsage`/`BudgetLimitPrompt` 无生产调用方,创建 goal 从不设预算(永远无上限);`/goal` 无 `start` 子命令 |
| 7 | `delegate_task` 子 agent(v0.4 储备) | ✅ | `tools/delegate.go` + `delegate.go`;agent 内串行 mutex;子任务 toolset 过滤掉 `delegate_task` 防嵌套 |

### 3.2 In-agent 记忆(docs/memory-improvements.md 的 P 项)

| P 项 | 目标 | 状态 | 证据 / 备注 |
|------|------|------|-------------|
| P0-1 | 语义读路径注入 system prompt | ✅ | 实际实现在 **`BuildSystemPromptWithMemory`**(`context.go`)→ `searchRelevantMemory`;文档所指的 `BuildSystemPromptSections` 仍传 nil,核对时易误判 |
| P0-2 | AutoPersist 接 embedding pipeline | ✅ | `AutoPersistMemory` 落库前 `embedFactSafe`(`memory.go`) |
| P1-1 | MEMORY.md 去重 + 压缩 | ⬜ | 仍纯追加;迁移到 [memory-governance.md](memory-governance.md) |
| P1-2 | 统一两条更新路径 | ⬜ | Heartbeat(关键词+仅文件)与 AutoPersist(LLM+双写)仍分叉;迁移到治理文档 |
| P2 | 稳定 agentID(UUID 而非 agent name) | ⬜ | `SetPGStore(store, a.name)` 仍用 name 当 DB 主键 |
| P2 | 隐私扫描可拦截(warn/redact/block) | ⬜ | 写入点仍只 `Scan`+warn;`privacy.Scrub` 存在但只用在送 LLM 的消息上,没用在记忆写入点;迁移到治理文档 |
| P3 | 清理 dead code(PGMemoryStore) | ✅ | 已被实际引用,不兼容签名已移除 |

### 3.3 MCP 记忆后端(docs/memory-mcp.md)

| # | 目标 | 状态 | 证据 / 备注 |
|---|------|------|-------------|
| 1 | `pawnix mcp` + memory_search/stats/write | ✅ | `cmd/pawnix/cmd_mcp.go`;`--source`/`--agent` flag |
| 2 | `mcp_events` 观测表 + EventStore | ✅ | `store/pg/event_store.go` + DDL |
| 3 | `/api/memory/usage` + 看板前端 | ✅ | `handlers_memevents.go` + `web/src/app/memory/page.tsx` |
| 4 | 写入服务端护栏(去重 / PII / 校验) | ⚙️ | **刻意不做**:只校验 content 非空 + kind 枚举,其余判断全交调用模型(读积极写保守)。是否补护栏见 [memory-governance.md](memory-governance.md) |

---

## 4. 已知欠账 / 技术债

按优先级粗排,详情散见各专题文档:

1. **Goal token 预算守护是"文档承诺但代码缺席"**。v0.3-plan 把它列为 MVP 必带项(默认 50k),但代码里预算永远为 nil、`FoldUsage`/`BudgetLimitPrompt` 无人调用。目前防跑飞只靠"continuation 不再触发新一轮"的 tight-loop 闸(`slash_goal.go`)。**长目标烧 token 的风险实际未被拦。**
2. **记忆写入零服务端治理**,且现在有两个写入方(in-agent AutoPersist + MCP `memory_write`)共写 `memories` 表,去重/PII/一致性都没兜底。见 [memory-governance.md](memory-governance.md)。
3. **todo 能力名不副实**:没有专门的 todo 工具,plan nudge 还引用了不存在的 `update_todo`;进度面板靠约定的文件格式,模型跑偏就断。
4. **两条 in-agent 记忆更新路径分叉**(Heartbeat 仅写文件、AutoPersist 双写),DB 只看到一半数据。
5. **agentID 用 agent name 当主键**,改名/同名跨 workspace 会破坏 DB 一致性。
6. **文档与实现的命名/路由漂移**:steer 路由异名、P0-1 实现函数改名、`/goal` 无 `start`——已在各文档就地更正。

---

## 5. 目标变了的地方(给未来的自己)

- **记忆:从"自动"转向"受控共享"。** 早期方向是把 AutoPersist 的自动抽取做强(memory-improvements.md);踩过"自主判断写出一堆垃圾记忆"的坑后,转向 MCP 共享池 + 读积极写保守,把"该不该写"的判断权交回调用模型。memory-improvements.md 的 P1/P2 治理项没有消失,而是被重新定位成"服务两条写入路径的服务端护栏",收敛进 [memory-governance.md](memory-governance.md)。
- **v0.3 的主题从"Memory OS"改成了"Steering"。** 记忆能力被拆出去单独走 MCP 后端线;v0.3 专注把 agent 变成"能陪你跑一天的搭子"。
- **Codex/ChatGPT 的记忆实现被引为对照物**:后台自动提炼 + 跨会话归并(extract→consolidate 两模型)、低配额跳过、带外部工具结果的会话不入库——这些成为治理文档"可选自动层"的设计输入。

---

## 6. 文档地图

| 文档 | 用途 | 时效 |
|------|------|------|
| [README.md](../README.md) | 项目定位 / 安装 / 路线图 / 特性 | ✅ 已随本轮更新 |
| [docs/STATUS.md](STATUS.md) | **本文** · 状态单一入口 | ✅ 新增 |
| [docs/memory-mcp.md](memory-mcp.md) | MCP 记忆后端接入指南(4 个宿主) | ✅ 当前准确 |
| [docs/memory-governance.md](memory-governance.md) | **记忆治理设计 + TODO**(软引导→服务端护栏→自动层) | ✅ 新增 |
| [docs/memory-improvements.md](memory-improvements.md) | in-agent 记忆优化路线(部分已被 MCP 方向取代) | ✅ 已加 superseded 标注 |
| [docs/memory-verification.md](memory-verification.md) | in-agent pgvector 记忆的端到端验证流程 | ✅ 当前准确(限 in-agent 路径) |
| [docs/v0.3-plan.md](v0.3-plan.md) | v0.3 Steering 计划 → 交付记录 | ✅ 已改为交付归档 + 欠账 |
| [docs/v0.3-test-guide.md](v0.3-test-guide.md) | v0.3 特性手测指南 | ✅ 当前准确(全套最准) |
| [docs/system-prompt-improvements.md](system-prompt-improvements.md) | system prompt token 预算优化 | ✅ 已重写(skills 预算已发布) |
| [docs/multimodal.md](multimodal.md) | 图片附件 Stage-1 管线 | ✅ 已修正路径命名 |
| [docs/onboarding-issues.md](onboarding-issues.md) | onboarding 向导缺陷评审 + 修复台账 | ✅ 保留(P0 均已修) |
| [RELEASE_NOTES_v0.2.0.md](../RELEASE_NOTES_v0.2.0.md) | v0.2 rebrand 发版记录 | ✅ 历史存档,不改 |
| [deploy/browser/README.md](../deploy/browser/README.md) · [SOP.md](../deploy/browser/SOP.md) | 反检测浏览器容器部署 | ✅ 已修正 skill 路径 |
| [web/README.md](../web/README.md) | 看板前端说明 | ✅ 已重写(原为脚手架) |
