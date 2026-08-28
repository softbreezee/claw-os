# System Prompt 优化路线图

> **状态:P1(skills 预算)已交付,余项部分未做。** 更新于 2026-08-28(项目已 FastClaw→Pawnix 重命名)。
> - ✅ **P1 skills token budget 已上线**:`SkillsSummaryBudgetTokens = 2000`(`skills.go:246`),skills 段硬性截到 ~2k,`load_skill` 按需拉 body。这是本文最主要的成果,已把"装越多 skill prompt 越膨胀"的线性增长掐断。
> - ⬜ bootstrap 文件缓存(P1)、Self-Update 条件注入(P2)、Context Usage 分段 tooltip(P3)未做。
> - ✅ identity 已换成 Pawnix 文案(见下方 P2)。
>
> 起源:用户在 chat header 的 Context Usage 圆环上看到一个新会话就有 5.2k token 占用,进而对 system prompt 设计提出疑问。本文记录架构 review 中识别出的可改进项。

## 现状基线

> ⚠️ 下表是**优化前**的基线(skills 段约 4.5k)。P1 上线后 skills 段被截到 ~2k 预算,现总量显著低于当时的 5.2k。保留此表作为优化前后的对照。

一个装了 ~28 个 skill 的 agent,优化前新 session 的 system prompt 约 **5.2k tokens**,分布大致如下:

| 部分 | 估算 token | 来源 |
|------|-----------|------|
| Identity（硬编码） | ~50 | `internal/agent/context.go: BuildSystemPrompt()` / `BuildSystemPromptWithMemory()` |
| `workspace/*.md`（4 个 bootstrap 文件） | ~300 | `loadFile()` |
| **Skills summary** | **~4500 → 现截到 ~2000** | `skills.go: BuildSkillsSummary()`（受 `SkillsSummaryBudgetTokens` 约束） |
| Workspace Self-Update guidance | ~120 | 始终拼入 |
| 其他（thinking 等） | ~200 | |

横向对比：

| 工具 | 系统提示词基线 |
|------|---------------|
| Aider | ~2-3k |
| Claude Code | ~3-5k |
| Cursor Composer | ~4-6k |
| **Pawnix（28 skills，优化前）** | **~5.2k** |
| Cline / Roo Code | ~5-8k |
| Cursor Agent | ~8-12k |

5.2k 在合理区间，但**架构上还有几处可优化**。

---

## 优化项

### ✅ P1: Skills Summary 加 Token Budget — 已交付

> 按推荐的"方案 4 + 方案 2 组合"落地:`SkillsSummaryBudgetTokens = 2000`(`internal/agent/skills.go:246`)硬性约束 skills 段;`BuildSkillsSummary` 始终注入 name + 一句话描述,超预算截断并提示,完整 body 走 `load_skill` 工具按需拉取(外加 IP guard 防 chatter 套出 prompt template)。下方为原始分析,存档。

**问题**：当前逻辑是"装多少 skill 拼多少描述"。装 100 个 skill 直接 20k+，吃掉模型 context。

**方案**（任选其一或组合）：
1. 每个 skill description hard limit（建议 200 chars）
2. 全局 skills section budget（建议 4k tokens），超了截断 + 提示
3. 基于 user query 的 RAG 召回（top-K 相关 skill 才拼入完整描述，其他只给 name）
4. 二级摘要：始终注入 skill name + 一句话描述；详细描述按需 `load_skill` 拉取（已有 `load_skill` tool）

**推荐**：方案 4 + 方案 2 组合。 **← 已按此落地,预算取 2000 token。**

---

### 🟡 P1: Bootstrap 文件 / Memory 的内存缓存

**问题**：`context.go: loadFile()` 和 `memory.go: LoadMemory()` 每次 `BuildSystemPrompt()` 都直接 `os.ReadFile`。在高频对话场景下，4-7 个 `.md` 文件每次读盘，不优雅。

**改动**：
- 在 `ContextBuilder` 内加 `map[string]cachedFile`，记录 mtime
- 调用时先 `os.Stat` 比 mtime，未变就用缓存
- 可选：用 fsnotify 主动失效

**状态**：未实施

---

### 🟢 P2: Workspace Self-Update 改为条件性注入

**问题**：那段 120 token 的 "Workspace Self-Update" 提示词每次都拼入，但绝大多数对话不需要它。

**改动**：
- 检测到 agent 最近 N 轮调用了 `write_file` 修改 MEMORY.md/USER.md 才注入
- 或每隔 M 轮注入一次
- 或加配置项 `agent.workspaceSelfUpdate: bool`

---

### 🟢 P2: Identity 国际化 — 🟡 部分

> 默认 identity 现为 `You are Pawnix, a self-hosted AI-Native personal OS.`(`internal/agent/context.go:111`),已随重命名更新,但**仍是英文硬编码**。用户可用 `IDENTITY.md` 覆盖(该路径已支持),但按 locale 自动切换语言未做。

**问题(原文,存档)**:`"You are FastClaw, a lightweight AI Agent."` 是英文硬编码。中文用户的 agent 内心独白还是英文,体验不沉浸。

**改动**：
- 根据 user locale / agent config 切换语言 — ⬜ 未做
- 或允许用户在 `IDENTITY.md` 完全覆盖默认 identity — ✅ 已支持

---

### 🟢 P3: Context Usage Tooltip 增加分段细节

**问题**：现在 tooltip 只显示总体 used / window。看不到 "system prompt 占多少 / 历史消息占多少 / 工具结果占多少"。

**改动**：扩展 `SessionContextInfo` 返回 breakdown，tooltip 加堆叠 bar。

---

## 不在本次范围

- 多模态 prompt 优化
- Tool calling schema 的 token 优化
- Memory 长期压缩策略
