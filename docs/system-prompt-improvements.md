# System Prompt 优化路线图

> 起源：用户在 chat header 的 Context Usage 圆环上看到一个新会话就有 5.2k token 占用，进而对 FastClaw 的 system prompt 设计提出疑问。本文记录架构 review 中识别出的可改进项。

## 现状基线

一个装了 ~28 个 skill 的 agent，新 session 的 system prompt 约 **5.2k tokens**，分布大致如下：

| 部分 | 估算 token | 来源 |
|------|-----------|------|
| Identity（硬编码） | ~50 | `internal/agent/context.go: BuildSystemPrompt()` |
| `workspace/*.md`（4 个 bootstrap 文件） | ~300 | `loadFile()` |
| **Skills summary** | **~4500** | `skills.go: BuildSkillsSummary()` |
| Workspace Self-Update guidance | ~120 | 始终拼入 |
| 其他（thinking 等） | ~200 | |

横向对比：

| 工具 | 系统提示词基线 |
|------|---------------|
| Aider | ~2-3k |
| Claude Code | ~3-5k |
| Cursor Composer | ~4-6k |
| **FastClaw（28 skills）** | **~5.2k** |
| Cline / Roo Code | ~5-8k |
| Cursor Agent | ~8-12k |

5.2k 在合理区间，但**架构上还有几处可优化**。

---

## 优化项

### 🟡 P1: Skills Summary 加 Token Budget

**问题**：当前逻辑是"装多少 skill 拼多少描述"。装 100 个 skill 直接 20k+，吃掉模型 context。

**方案**（任选其一或组合）：
1. 每个 skill description hard limit（建议 200 chars）
2. 全局 skills section budget（建议 4k tokens），超了截断 + 提示
3. 基于 user query 的 RAG 召回（top-K 相关 skill 才拼入完整描述，其他只给 name）
4. 二级摘要：始终注入 skill name + 一句话描述；详细描述按需 `load_skill` 拉取（已有 `load_skill` tool）

**推荐**：方案 4 + 方案 2 组合。

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

### 🟢 P2: Identity 国际化

**问题**：`"You are FastClaw, a lightweight AI Agent."` 是英文硬编码。中文用户的 agent 内心独白还是英文，体验不沉浸。

**改动**：
- 根据 user locale / agent config 切换语言
- 或允许用户在 `IDENTITY.md` 完全覆盖默认 identity

---

### 🟢 P3: Context Usage Tooltip 增加分段细节

**问题**：现在 tooltip 只显示总体 used / window。看不到 "system prompt 占多少 / 历史消息占多少 / 工具结果占多少"。

**改动**：扩展 `SessionContextInfo` 返回 breakdown，tooltip 加堆叠 bar。

---

## 不在本次范围

- 多模态 prompt 优化
- Tool calling schema 的 token 优化
- Memory 长期压缩策略
