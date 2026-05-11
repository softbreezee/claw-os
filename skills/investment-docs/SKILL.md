---
name: investment-docs
type: skill
version: 1.0.0
description: 独立投资文档生成技能（Standalone Investment Document Generator）。从建模数据、MAVI 变量锁定表、原始分析笔记等多种输入，生成五类标准化投资文档：IC Memo（~10页）、Investment Brief（~3页）、One-Pager（~1页）、Backtest Report、MAVI 综合报告。可单独使用，也可作为 financial-modeling 或 MAVI 的下游输出层。
author: Investment Intelligence Suite
standalone: true
---

# investment-docs — 独立投资文档生成技能

## 1. 技能定位

本技能是 Investment Intelligence Suite 的**文档输出引擎**，负责将投资数据、模型结果或分析笔记转化为专业格式的投资文档。

### 支持的输入类型

```
输入类型 A: financial-modeling 输出（完整模型数据 Schema 9/10）
输入类型 B: MAVI 变量锁定表（Multi-Agent Validation Interface 输出）
输入类型 C: 原始分析笔记（用户提供的文字/要点）
输入类型 D: investment-data-layer 输出（Schema 9 市场数据 + 历史财务）
输入类型 E: 以上任意组合
```

### 五种输出文档

| 文档类型 | 适用场景 | 推荐页数 | 模板文件 |
|---------|---------|---------|---------|
| **IC Memo** | PE/VC/并购投委会决策 | ~10页 | `templates/ic-memo.md` |
| **Investment Brief** | HF 仓位推荐、晨会汇报 | ~3页 | `templates/investment-brief.md` |
| **One-Pager** | 快速分享、初筛、Teaser | ~1页 | `templates/one-pager.md` |
| **Backtest Report** | 已投项目复盘、假设标定 | ~5页 | `templates/backtest-report.md` |
| **MAVI Report** | 多智能体验证综合报告 | ~4页 | 见第 6 节内嵌模板 |

### 触发条件

```
用户主动请求:
  "帮我生成 IC Memo"
  "出一份 Investment Brief"
  "给我写个 One-Pager"
  "做一个 Backtest Report"
  "生成 MAVI 报告"

自动建议（当检测到上游技能输出时）:
  financial-modeling 完成 → 建议生成 IC Memo 或 Investment Brief
  MAVI 完成 → 建议生成 MAVI Report
  已有分析笔记 → 建议生成 One-Pager
```

---

## 2. 文档生成工作流

```
Step 1: 识别文档类型
  → 用户请求哪种文档？
  → 未指定时：按场景推荐（见表格）

Step 2: 评估输入完整性
  → 检查哪些输入类型可用（A/B/C/D/E）
  → 识别缺失的关键数据节（见各文档必需字段）
  → 缺失时提示用户补充，或标注 [DATA MISSING]

Step 3: 加载对应模板
  → 从 templates/ 目录加载对应 .md 文件
  → 将 [占位符] 替换为实际数据

Step 4: 填充内容（按数据来源优先级）
  优先级: financial-modeling 输出 > MAVI 锁定表 > 原始笔记 > 手工输入

Step 5: 数据质量标注
  → 来自 LOW 置信度数据的字段标注 [INFERRED]
  → 缺失字段标注 [DATA MISSING — 请提供: X]
  → 草稿状态标注 [DRAFT]

Step 6: 用户审阅
  → 展示完整文档草稿
  → 接受: "确认" → 最终输出
  → 修改: "修改 Section X" → 仅重做该节
  → 重做: "全部重做" → 返回 Step 3

Step 7: 最终输出
  → Markdown 格式（默认）
  → 可转 DOCX / PDF（需工具链支持）
```

---

## 3. IC Memo — 投委会备忘录

### 3.1 适用场景与输入需求

```
适用: PE/VC 投资决策（少数股权/成长期/控股）、并购审批、LP 报告
推荐输入: financial-modeling 完整输出 + investment-data-layer Schema 9
最低输入: 公司名称 + 行业 + 基本财务数据 + 用户提供的分析要点
```

### 3.2 Section 结构（10 个必需节）

```
Section 1: Executive Summary（1页）
  必需字段: 公司名、行业、交易类型、投资规模、推荐评级、
            IRR（三场景）、MoIC、Entry EV、关键风险
  来源: financial-modeling 估值摘要

Section 2: Investment Highlights（3-5个亮点）
  必需字段: 每个亮点含标题 + ≥2个量化数据点 + 来源
  来源: 用户分析笔记 + financial-modeling 财务数据

Section 3: Company Overview
  必需字段: 业务模式、收入来源结构、市场地位、竞争格局、里程碑
  来源: 用户研究笔记 + investment-data-layer 行业分类

Section 4: Market Opportunity
  必需字段: TAM/SAM/SOM（含方法论）、市场驱动因子
  来源: 用户研究笔记

Section 5: Financial Analysis（IS/BS/CF 汇总）
  必需字段: 3年历史 + 3年预测（Revenue、Margin、FCF、Net Debt）
  来源: financial-modeling IS/BS/CF 输出 + investment-data-layer 历史数据

Section 6: Valuation（多方法汇总 + Football Field）
  必需字段: DCF（三场景）、Comps 倍数（EV/Rev、EV/EBITDA）、Blended
  来源: financial-modeling DCF 模块 + Comps 模块

Section 7: Key Risks & Mitigants（≥5个风险）
  必需字段: 每个风险含 Severity/Probability/IRR 影响/缓解措施
  来源: 用户研究 + financial-modeling 敏感性分析

Section 8: Management Assessment（仅 PE/VC）
  必需字段: CEO 5维评分（见下表）、执行折扣
  来源: 用户评估 + Org X-Ray（若可用）

Section 9: Returns Analysis（IRR/MoIC 三场景 + 敏感性矩阵）
  必需字段: Bear/Base/Bull 三场景完整参数、期望值（概率加权）
  来源: financial-modeling BT 模型

Section 10: Recommendation（决策 + 下一步）
  必需字段: Go/Pass/Conditional Go、条件（如适用）、行动清单
  来源: 综合判断
```

### 3.3 格式规范

```
财务数据表格: 左对齐标签，右对齐数字，单位在表头注明（$M / $B）
场景对比: 始终呈现 Bear / Base / Bull 三列
估值汇总: 必须包含 Comps vs DCF 偏差（< 30% ✅，> 30% ⚠️）
风险矩阵: Severity × Probability 热力图格式
管理层评分:
  | 维度 | 权重 | 评分 |
  CEO 战略能力    25%   [X]/10
  执行力与交付    25%   [X]/10
  管理团队深度    20%   [X]/10
  资本配置能力    20%   [X]/10
  文化与激励对齐  10%   [X]/10
  综合得分        ---   [X.X]/10 → 执行折扣 -[X%]
```

### 3.4 草稿 vs 正式版

```
正式版条件: 所有 Section 的关键字段均已填充，无 [DATA MISSING]
草稿标注: 在文档顶部显示:
  [DRAFT — 以下字段待补充: Section X (字段名), Section Y (字段名)]
```

---

## 4. Investment Brief — HF 投资简报

### 4.1 适用场景与输入需求

```
适用: HF 仓位推荐（多头/空头）、晨会汇报、月度股票观点更新
推荐输入: HF 投资论文笔记 + financial-modeling 估值输出
最低输入: 核心押注一句话 + 共识差异分析 + 目标价（三场景）
```

### 4.2 Section 结构（6个节，约 2-3 页）

```
Section 1: Thesis Statement（核心押注，1句话）
  格式: "[买入/做空] [公司] — [核心理由]"
  必须: 包含核心变量、预期变化方向、时间窗口

Section 2: Consensus vs My View（差异地图）
  必须: ≥3 个量化差异维度（Revenue/Margin/EPS），标注 Evidence Grade (1-5)
  核心分歧说明: 2-3行，解释"为什么市场错了"

Section 3: Expected Value（概率加权收益）
  必须: Bull/Base/Bear 三场景目标价、概率、回报
  必须: 盈亏比 > 1.5x 才推荐入仓

Section 4: Valuation（简化版）
  格式: 当前价格 → DCF 隐含价格、Comps 隐含价格、与同行倍数对比
  备注: 不需要完整 Football Field，仅关键倍数

Section 5: Catalyst & Timeline
  必须: ≥2 个具体催化剂（日期 + 可观测指标 + 触发阈值）
  必须: 论文证伪条件（Thesis DISPROVEN if...）

Section 6: Position Sizing & Risk
  必须: 推荐仓位占比、仓位类型（Screening/Tactical/Conviction/Core）
  必须: Hard Stop / Soft Stop / Time Stop 三个止损条件
```

### 4.3 仓位类型分级

```
| EV 区间  | 盈亏比  | 推荐仓位 | 类型       |
|---------|--------|---------|-----------|
| < 2%    | < 1.5x | PASS    | —         |
| 2-5%    | 1.5-2x | 0-1%    | Screening |
| 5-8%    | 2-3x   | 1-2%    | Tactical  |
| 8-12%   | 3-4x   | 2-3%    | Conviction|
| > 12%   | > 4x   | > 3%    | Core      |
```

---

## 5. One-Pager — 概览页

### 5.1 适用场景与输入需求

```
适用: 快速分享、项目初筛、卖方 Teaser、投前快速呈现
推荐输入: investment-data-layer 市场数据 + 基本估值 + 用户分析要点
最低输入: 公司基本信息 + 3个投资亮点 + 关键财务指标
```

### 5.2 四区块布局（严格单页）

```
区块 1（左上）: Company Snapshot
  公司名、股票代码、行业、总部、成立年、员工数、CEO、商业模式、发展阶段

区块 2（右上）: Key Financials
  LTM vs NTM 对比: Revenue、Gross Margin、EBITDA、EBITDA Margin、FCF、Net Debt

区块 3（左下）: Valuation Summary
  EV / Equity Value / 流通股
  LTM vs NTM 倍数（EV/Revenue、EV/EBITDA、P/E）vs 同行中位数
  DCF 区间 + Comps 区间

区块 4（右下）: Investment Thesis
  3个要点，每个1-2行（含关键数字）
  + Key Risk（一句话 + 量化影响）
```

### 5.3 格式规范

```
严格控制在单页（Markdown 中约 60-80 行）
避免长段落；全部使用表格或要点形式
数字必须标注单位（$M/$B/%）
关键数字加粗
```

---

## 6. Backtest Report — 回测报告

### 6.1 适用场景与输入需求

```
适用: 已投项目复盘、模型假设标定、LP 汇报、投资团队经验积累
推荐输入: 原始建模数据（历史预测值）+ 已实现的实际财务数据
最低输入: 至少 1 年的预测 vs 实际对比数据
```

### 6.2 Section 结构（5个节）

```
Section 1: Overview
  原始预测 IRR vs 更新 IRR、IRR 偏差、MoIC、持有期、投资状态

Section 2: Forecast vs Actual（逐行对比）
  指标: Revenue、EBITDA Margin、FCF、CapEx、NWC 变动
  偏差分级:
    ✅ 小偏差: < ±10%
    ⚠️ 中偏差: ±10% - ±25%
    🔴 大偏差: > ±25%

Section 3: Assumption Deviation Ranking（IRR 归因）
  排名最大偏差假设（≥5个），附 IRR 影响量化
  IRR Waterfall（瀑布图文字版）:
    原始预测 IRR → 各假设偏差 ppts → 实际实现 IRR

Section 4: Calibration Recommendations
  为未来同类型/行业模型提供校正系数
  格式: 假设类别 | 原始做法 | 建议修正 | 适用场景

Section 5: Lessons Learned
  分三个维度: 模型层面 / 行业分析层面 / 投资流程层面
  每条教训必须对应具体的"未来操作改变"
```

---

## 7. MAVI Report — 多智能体验证综合报告（新增）

### 7.1 适用场景与输入需求

```
适用: MAVI（Multi-Agent Validation Interface）分析完成后的综合输出
推荐输入: MAVI 变量锁定表（Variable Lock Sheet）+ 各智能体分析摘要
最低输入: MAVI 锁定表 + 至少 2 个智能体的观点

MAVI Report 是 investment-docs 对 MAVI 技能的专属输出适配器。
```

### 7.2 Report 结构（4个主节）

```
Section 1: Variable Lock Sheet 可视化
  展示所有锁定变量，按置信度分级显示
  格式:
    | 变量 | 锁定值 | 置信度 | 锁定依据 | 锁定时间 |
    | Revenue CAGR | 18% | HIGH ✅ | 3个智能体一致 | Step 3 |
    | Exit Multiple | 15x | MED 🟡 | 2:1 多数 | Step 4 |
    | CapEx % Rev | [CONTESTED] | LOW ⚠️ | 智能体分歧 | — |

Section 2: 智能体分析摘要（Agent-by-Agent Summary）
  每个智能体一个小节:
    [智能体名称] — [核心判断一句话]
    关键输入: [X, Y, Z]
    核心发现: [主要结论]
    关键假设: [最重要的假设]
    与共识的分歧: [如有]

Section 3: 宏观覆盖层摘要（Macro Overlay Summary）
  宏观变量对分析的影响:
    利率环境: [当前状态] → 对估值影响 [+/-X%]
    行业景气度: [上行/中性/下行] → 对倍数影响 [+/-X]x
    地缘政治/监管: [相关风险] → 概率 [X%]，IRR 影响 [-X ppts]

Section 4: 催化剂日历 + 分歧高亮（Catalyst Calendar & Disagreement Highlights）

  催化剂日历（按时间线排列）:
    | 日期 | 催化剂事件 | 关键指标 | 各智能体预判 | 触发阈值 |

  分歧高亮（Disagreement Highlights）:
    列出智能体间分歧最大的 Top 3 变量:
    ⚡ [变量名]: 智能体 A=[X] vs 智能体 B=[Y]，差异=[Z]
    → 分歧原因: [分析]
    → 当前处理: [采用哪个/如何加权/标记为 CONTESTED]
```

### 7.3 MAVI Report 格式规范

```
Variable Lock Sheet: 必须使用表格，置信度用 ✅🟡⚠️❌ 标注
CONTESTED 变量: 用 [CONTESTED] 标注，不得静默填入任何值
各智能体摘要: 统一格式（4行：核心判断/关键输入/核心发现/分歧）
分歧高亮: 最多展示 Top 5 分歧，过多分歧说明 MAVI 未收敛，提示用户
```

---

## 8. 格式标准与输出规范

### 8.1 通用格式规范

```
表格对齐:
  文字列: 左对齐
  数字列: 右对齐（Markdown 中用 :--- 和 ---: 控制）
  单位: 在列标题中注明（Revenue ($M) / EV ($B) / Margin (%)）

标题层级:
  文档标题: H1
  Section: H2（## 1. SECTION NAME）
  子节: H3（### 1.1 Sub-section）

数字规范:
  金额: $X.XM / $X.XB（精确到1位小数）
  百分比: X.X%（精确到1位小数）
  倍数: X.Xx（精确到1位小数）
  IRR: X.X%（精确到1位小数）
```

### 8.2 状态标注规范

```
[DRAFT]                  — 文档为草稿，关键字段未完整填充
[DATA MISSING — 请提供: X]  — 关键数据缺失，需用户补充
[INFERRED]               — 数据由推算得出，置信度 LOW
[STALE — 数据日期: X]    — 数据超过时效性规则
[CONTESTED]              — 多智能体存在分歧，未锁定
```

### 8.3 输出格式

```
默认输出: Markdown（内联到对话）
DOCX 转换: 可复制 Markdown 到文档编辑器格式化
PDF 导出: 通过工具链支持（用户需自行转换）
```

### 8.4 常见问题处理

| 问题 | 处理方式 |
|------|---------|
| 数据不完整 | 标注 `[DATA MISSING — 请提供: X]`，不留空白占位符 |
| 数据来自 LOW 置信度来源 | 标注 `[INFERRED]`，不静默使用 |
| 用户要求合并文档 | One-Pager 摘要在前，IC Memo 正文在后 |
| 需要多语言版本 | 先输出中英双语，再按需翻译 |
| MAVI 未收敛（分歧 > 5个）| 在 MAVI Report 顶部标注 ⚠️，建议用户先解决分歧 |

---

## 9. 模板文件索引

```
templates/ic-memo.md           — IC Memo 完整模板（10节）
templates/investment-brief.md  — Investment Brief 模板（6节）
templates/one-pager.md         — One-Pager 模板（4区块）
templates/backtest-report.md   — Backtest Report 模板（5节）
templates/hf-investment-thesis.md — HF 投资论文模板（Layer 0）
templates/hf-post-mortem.md    — HF 仓位事后复盘模板
templates/hf-expected-value.md — HF 期望值计算模板
```

---

**版本**: 1.0.0 | **所属**: Investment Intelligence Suite | **最后更新**: 2026-04-16
**相关技能**: investment-data-layer, financial-modeling, MAVI
