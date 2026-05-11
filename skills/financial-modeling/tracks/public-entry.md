# integrated-modeling-public

上市公司综合建模入口。面向外部投资人视角的 Investor Model，自动路由到 Hedge Fund Track 或 Buyout Track。含 Layer 1-4 完整流程（研究→确认→建模指令→执行），年报/季报完整性扫描，模块化引用 core/hf/bt 子 skill。

**触发词**: 综合建模、建模[公司名]、Quick-Build、Four-Layer

---

## 模块引用声明

本 Skill 依赖以下模块，按场景动态加载：

| 模块 | 加载时机 | 角色 |
|------|---------|------|
| `integrated-modeling-core` | 始终加载 | 共享基础设施（数据管理、DCF 引擎、IRR 计算、Layer 1-4 框架） |
| `integrated-modeling-hf` | Track A 或 C | Hedge Fund Track（空头/多头投资、持股 <50%、短中期循环） |
| `integrated-modeling-bt` | Track B 或 C | Buyout Track（LBO、持股 ≥50%、长期控制权、杠杆） |
| `financial-modeling-dcf` | 始终可用（Engine 模式） | DCF 估值引擎（WACC、Terminal Value、Sensitivity） |

---

## 唤醒词与路由

| 触发词 | 行为 | 流程 |
|--------|------|------|
| "综合建模 [company]" | 自动诊断 | Layer 1 → Layer 2 → Layer 3 → Layer 4 |
| "建模 [company]" | 自动诊断 | Layer 1 → Layer 2 → Layer 3 → Layer 4 |
| "Quick-Build [company]" | 跳过诊断 | 直接 Layer 3（简化版），确认 Granularity |
| "Four-Layer [company]" | 跳过诊断 | 先确认 Granularity，再 Layer 3 |

---

## Layer 1: 研究员（自动执行，无需用户输入）

### 核心原则

**"分析结论反问，而非用问题采集信息"** — 研究员应主动形成观点，而后用 Targeted Questions 验证。

### 必扫描的 8 个维度

#### 1. 组织设计与架构（Organization & Design）
- 调用 `org-x-ray` module，提取 3-5 个关键要点
- 覆盖：事业部结构、权力分配、创始人话语权、管理层激励机制
- **Word limit**: 500 字以内

#### 2. 业务表现（Business Performance）
- 收入结构与增长（按地域/产品线/客户类型）
- 用户 KPI（DAU/MAU/使用时长/ARPU）
- 毛利率/运营杠杆趋势
- 新产品或业务线的贡献

#### 3. 财务健康（Financial Health）
- **净现金位置**（严格按 core module 定义）
  - 公式：现金及等价物 − 短期债 − 当年到期长期债
- 经营 CF、自由 CF（FCF = Operating CF − Capex）
- 杠杆比率（Net Debt/EBITDA，目标 <3x）
- 资本开支率与维护 vs 增长拆分

#### 4. 估值（Valuation）
- 当前市值、EV（Enterprise Value）
- 历史 PE、EV/EBITDA 范围（3Y、5Y）
- 相比同行中位数的溢价/折价
- 隐含增长与 Terminal Margin 假设

#### 5. 管理层与激励（Management & Incentives）
- 创始人/CEO 持股比例与承诺
- 股权激励结构（Vesting、Strike Price）
- 近期战略声明与资本配置政策
- 历史执行率（预算 vs 实际）

#### 6. 激进投资者与大股东（Activist & Major Shareholders）
- 前 10 大股东动态（过去 6 个月）
- 机构投资者的进出信号
- 私有化压力/战略审视
- 高管减持 vs 增持

#### 7. 战略事件（Strategic Events）
- 近期 M&A、剥离、合作
- 新产品或技术突破（AI、新平台等）
- 监管变化（反垄断、隐私法、行业牌照）
- 供应链重构、定价权变化

#### 8. 行业动态（Industry Dynamics）
- 竞争格局变化与市场份额
- 监管趋势（新法规、牌照难度）
- 宏观影响（利率、汇率、周期）
- 替代品威胁与颠覆风险

### Layer 1 子流程：收入与投资周期诊断

**必填项**（来自 core module）：
1. **Revenue Cycle vs Investment Cycle** - 当前阶段判断（高速增长/成熟平稳/衰退修复）
2. **Transition Signals** - 周期切换的领先指标（DAU 增速、客户留存率、定价权、产品创新节奏）
3. **Leading Indicator Dashboard** - 5-7 个关键动量驱动因子

### Layer 1 子流程：Targeted Questions 格式

每个 Targeted Question 必须包含三个成分：
1. **数据事实** — 当前披露数据与观察
2. **市场观点** — 牛/熊/分歧的主流论点
3. **3-Scenario 表格** — Bull/Base/Bear 假设与影响

### Layer 1 标准输出结构

```
## [公司名] 研究备忘录 — Layer 1 输出

【Track Label】🔵 Hedge Fund / 🟠 Buyout / 🟣 Full Model

### 执行摘要（3-5 句话）
核心投资逻辑 + 周期诊断 + 当前最大分歧点

### 1. 组织洞察（来自 org-x-ray，3-5 bullet）
- [洞察 1]
- [洞察 2]
- ...

### 1b. 管理层投资性评分 🟠
*仅限 Buyout Track，来自 bt module*
- 运营能力：评分
- 财务纪律：评分
- 文化粘性：评分

### 2. 核心财务数据表
| 指标 | 2023A | 2024E | 2025E | YoY Growth | 数据源 |
|------|-------|-------|-------|-----------|--------|
| 收入 | XXX | XXX | XXX | X% | [来源] |
| Operating CF | XXX | XXX | XXX | X% | [来源] |
| FCF | XXX | XXX | XXX | X% | [来源] |
| Net Cash | XXX | XXX | XXX | - | [来源] |
| Net Debt/EBITDA | Xx | Xx | Xx | - | [计算] |

### 3. 业务结构分析
- 收入驱动分解（产品线/地域/客户类型）
- 毛利率与经营杠杆趋势
- **Leading Indicator Identification** — 5-7 个关键动量驱动因子及其走势

### 4. 收入周期 vs 投资周期诊断
- **当前阶段**：高速增长 / 成熟平稳 / 衰退修复
- **周期切换信号**：确认 transition 的 3-5 个关键指标
- **Revenue Cycle 链条**：销售→确认→收现 的时间跨度
- **投资周期领先性**：市场何时会定价周期转换

### 5. 资本结构与 LBO 可行性评估
*仅限 Full 或 Buyout Track*
- 当前杠杆能力（Net Debt/EBITDA，目标范围）
- FCF 覆盖比例（Debt/FCF）
- 利率敏感性与续期风险
- 优化空间（税收盾、资产轻型化）

### 6. SOTP 框架
*仅限 Full Model，在研究备忘录末尾*
| 业务线 | 2024E EBITDA | Multiple | 隐含价值 | 占比 |
|--------|-------------|----------|--------|------|
| 核心业务 | XXX | Xx.x | XXX | X% |
| 高增长线 | XXX | Xx.x | XXX | X% |
| 其他 | XXX | Xx.x | XXX | X% |
| **合计** | **XXX** | **-** | **XXX** | **100%** |
| 加：现金 | XXX | - | - | - |
| 减：债务 | (XXX) | - | - | - |
| **股权价值** | **-** | **-** | **XXX** | **-** |
| **目标价格** | **-** | **-** | **X.XX/股** | **-** |

### 7. 关键风险
- 竞争加剧 / 市场份额丧失
- 监管不确定性
- 定价权侵蚀
- 成本压力 / 毛利率恶化
- 技术迭代风险
- 宏观周期敏感性
- 流动性风险（并购、破产）

### 7b. PE 特定风险 🟠
*仅限 Buyout Track，来自 bt module*
- 杠杆风险（利率上升、违约风险）
- 市场退出时机风险
- 运营改善滞后风险
- 管理层流失风险

### 8. Targeted Questions Q1-QN
每个问题格式：

**Q[N]：[问题标题]**

**数据事实**：
- 当前披露数据…
- 观察数据…

**市场观点**：
- 🐂 Bull Case：…
- 🐻 Bear Case：…
- 分歧点：…

**3-Scenario 分析**：
| 假设 | Bull | Base | Bear |
|------|------|------|------|
| [关键变量 1] | X | Y | Z |
| [关键变量 2] | X | Y | Z |
| 暗示 IRR/Multiple | X% | Y% | Z% |

### 9. 共识分歧图谱 🔵
*仅限 Hedge Fund Track，来自 hf module*
- 当前共识估值（市场隐含的 PE/EV）
- 研究员分歧的前 3 个维度
- 不同派系的关键假设对比

### Thesis Validation 声明 🔵
*仅限 Hedge Fund Track，来自 hf module*
"本投资逻辑依赖于以下核心假设验证…" — 列示 3-5 个关键验证点及其时间表

### → 确认继续进入 Layer 2？
如确认，进入 Layer 2 Partner 阶段。

```

---

## Layer 2: 合伙人（交互式，HARD RULE：必须等待用户）

### 核心规则

**禁止自动填充答案或跳过本层** — 用户意见是基线，不可被 AI 假设覆盖。

### 问题清单结构

**同时展示所有问题**（不逐个等待）。用户一次性回答，格式紧凑。

#### 标准 10 个问题

1. **Track Selection**（必填）
   - [ ] Track A：Hedge Fund（空头/多头，短中期持股，<50%）
   - [ ] Track B：Buyout（LBO，控制权，≥50%）
   - [ ] Track C：Full Model（两条线合并）

2. **Granularity Level**（必填）
   - [ ] L1：Quarterly（按季披露拆分）
   - [ ] L2：Annual + Key Driver（年度+关键驱动）
   - [ ] L3：Full Build（每个产品线/地域/成本中心）

3. **Scenario Probability Distribution**（可调）
   - 默认：Bull 25% / Base 50% / Bear 25%
   - 用户可微调（必须合计 100%）

4. **Exit Timeline**（仅 BT 必填）
   - [ ] 3 年（PE 标准）
   - [ ] 5 年（战略买家）
   - [ ] 7 年+（长期控制）

5. **Leverage Target**（仅 BT 必填）
   - [ ] Conservative：1.5-2.0x Net Debt/EBITDA
   - [ ] Moderate：2.0-3.0x
   - [ ] Aggressive：3.0-4.0x

6. **Revenue Growth Assumption**（必填）
   - Base Case 3 年平均增速（%）
   - 增长来源：有机 / M&A / 定价

7. **Margin Expansion Opportunity**（必填）
   - EBITDA Margin 目标（3 年末）
   - 利用杠杆实现途径

8. **Working Capital Normalization**（必填）
   - DPO / DSO / Inventory Normalization 是否需要建模调整
   - 现金循环周期改善潜力

9. **Synergy / Value Creation Levers**（仅 BT 或 M&A）
   - 成本协同（削减重复成本、供应链优化）
   - 收入协同（交叉销售、定价权）
   - 财务工程（税收盾、资本结构优化）
   - 定价假设（% of acquisition price）

10. **Sensitivity & Stress Test Focus**（必填）
    - 最敏感的变量（EBITDA Multiple / Revenue Growth / WACC）
    - Stress Test 范围（±X% 或绝对值）

#### 针对性问题（来自 Layer 1 Targeted Questions）
用户可选择性回答，标记 Q1, Q2, Q3... 对应 Layer 1 中的关键分歧点。

#### 组织变量（仅 BT，来自 bt module）
- **O1：管理层保留与激励** — CEO/CFO 持股比例、新增 Equity Rollover
- **O2：文化与组织架构调整** — 拆分成本中心、新 KPI 制度
- **O3：供应链与成本中心优化** — 关键供应商、内部采购价格
- **O4：资本支出计划** — Maintenance CapEx vs Growth CapEx
- **O5：债务与利率假设** — 融资成本、期限结构

#### Track 特定问题
- **🔵 Hedge Fund Track**：短期催化剂周期、空头研究关键点、轮换触发条件
- **🟠 Buyout Track**：100-Day Plan 重点、运营改善指标、出口战略

### Layer 2 用户输入与输出

**用户回答格式**（紧凑）：
```
1. Track C
2. L2
3. Bull 30% / Base 45% / Bear 25%
4. 5 年
5. Moderate（2.5x）
6. 12% 有机 + 3% 并购
7. EBITDA Margin 从 X% → Y%（成本削减 + 定价）
8. DSO 从 45 天 → 40 天（应收加速）
9. Synergy：成本 $XX | 收入 $XX | 财务工程 $XX
10. 敏感变量：EBITDA Multiple 和 Revenue CAGR
补充：Q3 关于定价权的回答 → Bear Case 假设 -5% 价格
O1：CEO 继续 100% 持股，新任 CFO Equity Rollover 5%
```

**Layer 2 JSON 确认单**（自动生成）：
```json
{
  "company": "[Company Name]",
  "track": "C (HF + BT)",
  "granularity": "L2 (Annual + Key Driver)",
  "scenarios": {
    "bull": 0.30,
    "base": 0.45,
    "bear": 0.25
  },
  "bt_exit_timeline": "5 years",
  "bt_leverage_target": "2.5x Net Debt/EBITDA",
  "revenue_growth": "12% organic + 3% M&A",
  "margin_target": "Y%",
  "working_capital_delta": "DSO 45d → 40d",
  "value_creation_levers": {
    "cost_synergy": "$XX",
    "revenue_synergy": "$XX",
    "financial_engineering": "$XX"
  },
  "key_sensitivities": ["EBITDA Multiple", "Revenue CAGR"],
  "management_assumptions": {
    "ceo_retention": "100%",
    "cfo_rollover": "5%"
  },
  "status": "Ready for Layer 3"
}
```

**HARD STOP**：等待用户确认此 JSON 后，方可进入 Layer 3。

---

## Layer 3: 投资组合经理（自动执行，基于 Layer 2 反馈）

### 标准输出结构

#### 第 1 节：交易概览（仅 LBO/Buyout）

```
## Transaction Overview

**交易类型**：LBO / Strategic Add-On / Dividend Recapitalization / Merger of Equals
**交易规模**：
- Enterprise Value (EV)：$XXX
- Equity Value（含 Roll-over）：$XXX
- 估值倍数：X.Xx EV/EBITDA

**融资结构**：
- 股权投入：$XXX（X% 股权，含 Sponsor + Management）
- Senior Debt：$XXX（X.Xx Net Debt/EBITDA）
- Sub Debt（if any）：$XXX
- Cash Retained：$XXX

**目标指标**（3 年末）**：
- 目标 IRR：X%
- 目标 MoIC（Multiple on Invested Capital）：X.Xx
- 目标 Exit Equity Value：$XXX

```

#### 第 2 节：来源与运用（Source & Uses）

```
## Sources & Uses

### Sources（融资来源）
| 来源 | 金额 | % of Total |
|------|------|-----------|
| Senior Debt | $XXX | X% |
| Sub Debt | $XXX | X% |
| Sponsor Equity | $XXX | X% |
| Management Roll-over | $XXX | X% |
| Cash on Hand | $XXX | X% |
| **Total Sources** | **$XXX** | **100%** |

### Uses（资金运用）
| 用途 | 金额 | % of Total |
|------|------|-----------|
| Acquisition | $XXX | X% |
| Refinance Existing Debt | $XXX | X% |
| Working Capital Adjustment | $XXX | X% |
| Transaction Fees | $XXX | X% |
| **Total Uses** | **$XXX** | **100%** |

```

#### 第 3 节：价值创造驱动因素（Value Creation Levers）

```
## Value Creation Levers

### 杠杆 1：收入增长（Revenue Growth）
- **Base Case 3 年 CAGR**：X%
- **有机增长**：X%（市场份额 / 定价 / 新产品）
- **M&A 贡献**：X%（目标金额与倍数）
- **驱动因子**：[KPI 列表]

### 杠杆 2：边际改善（Margin Expansion）
- **基准 EBITDA Margin**：X%
- **目标 EBITDA Margin**（3 年末）：Y%
- **改善空间**：Y% - X% = Z ppt
- **具体举措**：
  - SG&A 削减：$XX → $XX（% of Revenue）
  - COGS 优化：供应商整合、定价权
  - 运营效率：产能利用率提升、流程自动化

### 杠杆 3：财务工程（Financial Engineering）
- **杠杆降低**：初始 X.Xx → 3 年末 Y.Yy（通过 EBITDA 增长）
- **税收盾**（如适用）：年均 $XX
- **资本支出优化**：Maintenance CapEx $XX vs Growth CapEx $XX
- **营运资本释放**：DSO/DPO 优化释放 $XX

### 杠杆 4：出口倍数扩张（Multiple Expansion）
- **入口倍数**：X.Xx EV/EBITDA
- **假设出口倍数**：Y.Yy EV/EBITDA（+ Z ppt 因质量改善）
- **倍数扩张的驱动**：增长率加快、风险降低、市场重评

```

#### 第 4 节：情景定义（Scenario Definitions）

```
## Scenario Definitions

### 概率分布（来自 Layer 2）
- Bull Case：X%
- Base Case：Y%
- Bear Case：Z%
  **→ 用户可在 Layer 2 中调整；一旦确认，Layer 3-4 锁定**

### Bull Case（+X ppt 概率权重）
| 假设 | Base | Bull |
|------|------|------|
| Revenue CAGR（3Y） | X% | +Y ppt |
| EBITDA Margin（Year 3） | X% | +Y ppt |
| Exit Multiple | X.Xx | Y.Yy |
| Exit Value | $XXX | $XXX |
| **IRR** | **-** | **X%** |
| **MoIC** | **-** | **X.Xx** |

### Base Case
[同上结构]

### Bear Case（-X ppt 概率权重）
[同上结构]

```

#### 第 5 节：Sheet 结构（Modeling Architecture）

```
## Modeling Sheet Structure

### Core Sheets（所有 Track）
1. **Assumptions Hub**
   - Input 变量集中管理（Revenue, Margin, CapEx, Tax Rate 等）
   - Scenario Toggle（可切换 Bull/Base/Bear）
   - Leading Indicator Drivers（KPI 仪表板）

2. **Revenue Build**
   - Granularity per Layer 2（产品线 / 地域 / 客户类型）
   - YoY Growth Rate & Margin %
   - 检查点：vs historical, vs market growth, vs guidance

3. **P&L（Income Statement）**
   - Gross Margin, Operating Margin, EBITDA Margin
   - SBC, D&A, Interest Expense, Tax
   - NOPAT for WACC calculation

4. **Cash Flow**
   - Operating CF（从 EBITDA 开始）
   - CapEx（Maintenance vs Growth）
   - Working Capital Changes
   - FCF & Unlevered FCF

5. **Balance Sheet**
   - Assets（PP&E, Intangibles, Goodwill, WC）
   - Debt Schedule（Senior / Sub, 偿还时间表）
   - Equity Rollover & Dilution
   - 检查点：Net Debt, Leverage Ratio

6. **Valuation & Return**
   - DCF Calculation（from core module）
   - Exit Valuation（Multiple or DCF）
   - Cash on Hand 与 Equity Value
   - IRR & MoIC Calculation

### Track-Specific Sheets（来自 hf/bt modules）
- **🔵 Hedge Fund Track**：
  - Hedge Position Drivers（短期催化、股价敏感性）
  - Market Consensus vs Model Gap（价差分析）
  - Short Thesis Stress Test（空头攻击场景）

- **🟠 Buyout Track**：
  - Sources & Uses（融资架构）
  - 100-Day Plan（运营改善时间表）
  - Synergy Waterfall（成本协同、收入协同）
  - Exit Scenarios（IPO, Sale, Dividend Recap）

```

#### 第 6 节：建模原则（Modeling Principles）

*来自 core module，强制遵守*：

```
## Modeling Principles (Per Core Module)

1. **数据溯源**：每个数据点（历史或预测）必须关联来源（年报脚注 / 管理层指引 / 行业数据）

2. **守恒原理**：
   - Operating CF = Net Income + D&A − Working Capital Delta − Other Adjustments
   - FCF = Operating CF − Capex（Maintenance + Growth）
   - Net Debt = Total Debt − Cash（严格定义）

3. **收入预测阶梯**：
   - Base Year（最近公布年）→ Transition Year（明年）→ Normalized Year（第 3 年）
   - 每一步增速递进需合理说明（市场饱和、竞争压力、产品成熟度）

4. **边际假设一致性**：
   - 不同产品线的 Margin 假设需参考历史波动 & 管理层指引
   - Margin 改善需有具体成本削减或定价权实现的支撑

5. **CapEx 合理性**：
   - Maintenance CapEx ≥ D&A（资产维持）
   - Growth CapEx ≤ Sales Growth（边际投入产出）
   - CapEx/Sales 需vs历史平均 & 行业标准

6. **税率假设**：
   - 有效税率（Effective Tax Rate）需参考历史 & 当地税法
   - 税损结转（Tax Loss Carryforward）如有，需单独建模

7. **WACC 构成**：
   - Cost of Equity = Risk Free Rate + Beta × Market Risk Premium
   - Cost of Debt = Weighted Avg Interest Rate + Credit Spread
   - WACC = (E/V) × Re + (D/V) × Rd × (1 - Tax Rate)
   - Terminal WACC 需与长期增长率一致（不能超过 GDP 增速）

8. **Terminal Value**：
   - 采用 Perpetuity Growth Method：TV = FCF(Year N) × (1 + g) / (WACC − g)
   - g （Terminal Growth Rate）不超过 2-2.5%（全球 GDP 增速）

9. **Sensitivity Range**：
   - 至少 2 个变量双向 Sensitivity（EBITDA Multiple vs Exit Year 现金流）
   - 范围：±20% 或 ±200 bps（相对重要性调整）

10. **Scenario Waterfall**：
    - 每个 Scenario（Bull/Base/Bear）的关键假设差异需清晰列示
    - 不能只改变一个变量，需整体逻辑一致

```

#### 第 6b 节：Track-Specific 分析（来自 hf/bt modules）

**🔵 Hedge Fund Track Extra Sections**（来自 integrated-modeling-hf）：

```
### Hedge Fund-Specific Analysis

#### 市场共识 vs 模型缺口（Consensus Gap Analysis）
| 假设维度 | 市场隐含 | 我们假设 | 缺口 | 机制 |
|---------|--------|--------|------|------|
| 3Y Revenue CAGR | X% | Y% | +/-Z ppt | [催化或风险] |
| Year 3 EBITDA Margin | X% | Y% | +/-Z ppt | [杠杆或竞争] |
| Exit Multiple | X.Xx | Y.Yy | +/-Z ppt | [质量 / 周期] |
| **隐含 IRR** | **X%** | **Y%** | **+/-Z ppt** | **—** |

#### 短期催化图表（12-24 个月）
- Q1-Q4：[关键财务公布、战略声明、并购动作]
- 预期反应：股价上升 X%、重评倍数至 Y.Yy

#### 空头攻击防守（如果存在空头论点）
- 空头主张：[论点]
- 我们反驳：[数据 + 逻辑]
- Stress Test：即使空头论点发生，底部价格仍为 $X

#### Thesis Validation Checkpoints（验证关键点）
- Catalyst 1：[预期确认时间表]
- Catalyst 2：[预期确认时间表]
- Catalyst 3：[预期确认时间表]

```

**🟠 Buyout Track Extra Sections**（来自 integrated-modeling-bt）：

```
### Buyout-Specific Analysis

#### 管理层评估与保留激励（Management Assessment & Retention）
| 关键人员 | 角色 | 当前持股 | Rollover | 新激励 | 风险 |
|--------|------|--------|---------|--------|------|
| CEO | 首席执行官 | X% | Y% | Equity/Options | [评估] |
| CFO | 财务官 | X% | Y% | Equity/Options | [评估] |
| COO | 运营官 | X% | Y% | Equity/Options | [评估] |

#### 100-Day Plan（第一个季度运营改善蓝图）
```
**第 1 周-2 周：诊断与团队建设**
- 完成财务审查（Cash Flow、成本中心、客户利润率）
- 组建核心运营小组（与现任管理层）
- 识别 Quick Win（30-60 天可实现）

**第 3 周-4 周：成本削减启动**
- 供应商招标（目标削减 % of COGS）
- SG&A 冻结（暂停非关键招聘、差旅）
- 流程梳理（识别自动化机会）

**第 5-8 周：收入加速**
- 客户拜访（定价权、跨销售机会）
- 产品优化（去除低 margin SKU）
- 定价审查（按客户/产品线）

**第 9-12 周：财务规范化**
- 建立 KPI Dashboard
- 月度董事会报告格式化
- 2024 年完整预算与目标设定
```

#### 出口蓝图（Exit Blueprint）
| 出口类型 | 目标 EV | 目标年份 | 先决条件 | 风险 |
|---------|--------|---------|---------|------|
| IPO | $XXX | Year 4-5 | EBITDA $XX, Margin X% | 市场周期、流动性 |
| Strategic Sale | $XXX | Year 3-4 | 增长 X% YoY, Market Share | 行业整合、价格竞争 |
| Secondary (LBO) | $XXX | Year 5+ | Margin X%, Growth X% | 融资成本上升 |

#### 杠杆路径（Deleveraging Path）
```
Debt Schedule:
| 年份 | Beginning Debt | Cash Flow 用途 | Principal Paydown | Ending Debt | Leverage |
|------|----------------|----------------|------------------|------------|----------|
| Yr 0 | $XXX | — | — | $XXX | X.Xx |
| Yr 1 | $XXX | $XXX FCF | $XX | $XXX | X.Xx |
| Yr 2 | $XXX | $XXX FCF | $XX | $XXX | X.Xx |
| Yr 3 | $XXX | $XXX FCF + Dividend | $XX | $XXX | X.Xx |
```

#### PE 特定风险（来自 bt module）
- 杠杆与利率风险：利率上升 X bps → 年度利息增加 $XX
- 退出时机风险：市场倍数在 X.Xx - Y.Yy 范围波动
- 运营改善滞后：如成本削减延迟，IRR 下降至 X%
- 管理层流失：关键人员离职概率 X%，替代成本 $XX

```

#### 第 7 节：100-Day Plan（仅 Buyout Track）

*来自 bt module，强制包含*

（详见上方第 6b 节的完整展开）

#### 第 8 节：Exit Blueprint（仅 Buyout Track）

*来自 bt module，强制包含*

（详见上方第 6b 节的完整展开）

### HARD STOP

**本 Layer 3 生成完毕后，必须等待用户确认建模指令后，方可进入 Layer 4 执行**。确认格式：

```
用户确认：
"Layer 3 已审阅，确认进入 Layer 4 执行"

→ 若未收到确认，停止，不自动执行 Layer 4
```

---

## Layer 4: 分析员（执行，Layer 3 确认后）

### 执行原则

**严格按照 Layer 3 指令执行**，不自行修改假设。

### 初稿完成后：强制年报/季报完整性扫描

执行 6 类完整性检查，生成结构化扫描报告：

#### 1. 财务报表主表再验证（Financial Statements Primary Tables）

```
【完整性检查 1】Financial Statements Reconciliation

□ Income Statement:
  □ Revenue（当年、去年、YoY%）
  □ Gross Profit / Gross Margin %
  □ EBITDA / EBIT（核实公式 = 自下而上或自上而下）
  □ Net Income（包括特殊项、税率）
  ✓ 检查：是否与脚注一致

□ Balance Sheet:
  □ Current Assets（现金、应收、存货、其他）
  □ PP&E, Net（总额 - 累计折旧）
  □ Intangibles & Goodwill（收购价差、减损）
  □ Current Liabilities（应付、短期债、其他）
  □ Long-term Debt（利率、到期日、担保）
  □ Stockholders' Equity（普通股、保留盈余、回购）
  ✓ 检查：Assets = Liabilities + Equity？

□ Cash Flow:
  □ Operating CF = Net Income + D&A ± WC Changes
  □ Investing CF（CapEx, M&A, Investments）
  □ Financing CF（债务、股权融资、分红、回购）
  □ Net Change in Cash
  ✓ 检查：Beginning Cash + Net Change = Ending Cash？

```

#### 2. 收入脚注深度扫描（Revenue Footnotes）

```
【完整性检查 2】Revenue Recognition & Deferred Items

□ 收入按产品线/地域拆分（年度数据 + 同比增速）
  例：Product A 占 X% of Total, YoY +Y%

□ Deferred Revenue（递延收入）
  □ 当年期末余额 vs 当年期初余额（变化 $XXX）
  □ 与 Implied Sales 是否一致

□ Contract Assets / Unbilled Revenue
  □ 金额与趋势（持续增长表示强劲签约）

□ RPO（Remaining Performance Obligations）
  □ 合同等级/客户类型拆分
  □ 多长时间内确认（签约周期风险）

□ Geographic Split
  □ 主要地域占比（国内/海外）
  □ 增长率差异（高增地域 vs 成熟地域）

```

#### 3. 成本与费用脚注深度扫描（Cost & Expense Footnotes）

```
【完整性检查 3】Cost Structure & Operating Expenses

□ COGS Breakdown（如果披露）
  □ 原材料 / 劳工 / 制造费用 占比
  □ 毛利率变化的驱动（产品 mix, 定价, 成本上升）

□ SBC（Stock-Based Compensation）
  □ 年度费用 $XXX（% of Revenue）
  □ 未来年份摊销义务（如有）
  □ vs Diluted Shares 增长的一致性

□ Depreciation & Amortization（D&A）
  □ D&A as % of Revenue（历史均值 X%）
  □ Useful Life Assumptions（PP&E 折旧年限）
  □ Intangible Assets 摊销期（注意减损风险）

□ 其他 Operating Expenses
  □ R&D Spend（% of Revenue, 投资强度）
  □ Sales & Marketing（% of Revenue, CAC 趋势）
  □ G&A（固定 vs 变动成分）

□ Restructuring / One-time Items
  □ 是否存在非经常项（分离、关闭、遣散费）
  □ 调整后的 Normalized EBITDA

```

#### 4. 资产负债表脚注深度扫描（Balance Sheet Footnotes）

```
【完整性检查 4】Balance Sheet Quality & Hidden Liabilities

□ Receivables（应收账款）
  □ 总额与 DSO（Days Sales Outstanding）
  □ Allowance for Doubtful Accounts（%）
  □ 老龄分析（>90 days 占比）

□ Inventory
  □ 存货类型（原材料/在产/产成品）占比
  □ Inventory Turnover（销售成本 / 平均存货）
  □ 减值迹象（过时、滞销）

□ PP&E & Intangibles
  □ 资本化 vs 费用化 决策（一致性）
  □ 减值测试结果（是否有减值迹象）
  □ Goodwill 占比（> Total Assets 的 20% 时警惕）

□ Deferred Taxes
  □ DTAs（递延所得税资产）可实现性
  □ DTLs（递延所得税负债）
  □ Valuation Allowance（是否全额备抵）

□ Off-Balance Sheet Items
  □ 租赁债务（IFRS 16/ASC 842）
  □ 保证、或有负债
  □ 特殊目的实体（SPE）关联

□ Related Party Transactions
  □ 与关联方的交易定价合理性
  □ 担保、贷款的条款

```

#### 5. 资本配置脚注深度扫描（Capital Allocation Footnotes）

```
【完整性检查 5】Capital Allocation & Debt Structure

□ CapEx 拆分
  □ Maintenance CapEx（维持现有产能）vs Growth CapEx（扩张）
  □ CapEx % of Sales（历史 X%, 目标 Y%）
  □ 地域/产品线 CapEx 预算

□ Capitalized vs Expensed R&D
  □ 内部研发资本化比例
  □ Software/IP Capitalization Policy（SOP）

□ Share Buybacks
  □ 年度回购金额与股数
  □ 平均回购价格 vs 当前股价
  □ 剩余授权额度

□ Dividends（如有）
  □ 每股派息 / Payout Ratio
  □ 可持续性（Dividend Coverage Ratio = FCF / Dividends）

□ 债务结构与期限
  □ 按到期日的债务分布（1Y/2-5Y/>5Y）
  □ Fixed vs Floating 利率
  □ 借贷协议限制条款（Covenants）

□ Interest Coverage
  □ EBIT / Interest Expense（>2.0x 为健康）
  □ 利率上升敏感性

```

#### 6. 最新季报增量披露（Latest Quarterly Incremental）

```
【完整性检查 6】Latest Quarterly & Forward Guidance

□ 最新季度财务数据
  □ Revenue, EBITDA Margin, FCF
  □ vs 同期去年（YoY）& vs 前一季度（QoQ）

□ 新披露的指引（Guidance）
  □ 全年 Revenue Range（上下限）
  □ Margin Outlook（保守/基础/乐观）
  □ CapEx 计划
  □ Free Cash Flow 预期

□ 新的战略公告
  □ 并购、剥离、合作框架
  □ 产品发布、市场扩展
  □ 监管风险、诉讼进展

□ 管理层评论（Outlook）
  □ CEO 对市场前景的论述
  □ 成本压力、定价能力的信号
  □ 资本配置意向

```

### 完整性扫描输出格式

**Table：Annual Report Completeness Matrix**

```
| 检查项 | 确认 ✓ | 发现的缺口 | 模型调整 | 优先级 |
|--------|-------|----------|--------|--------|
| 1. Financial Statements | ✓ | — | — | — |
| 2a. Revenue by Segment | ✓ | Deferred Revenue trend not disclosed | Estimate based on historical ratio | High |
| 2b. Contract Terms | ✓ | RPO 未按客户分类 | 按行业平均周期假设 | Med |
| 3a. SBC | ✓ | — | — | — |
| 3b. D&A Policy | ✓ | Intangible life 未披露 | 采用同行 5 年 | Med |
| 4a. AR/DSO | ✓ | 地域 DSO 拆分缺失 | 假设国内 30d, 海外 45d | Med |
| 4b. Goodwill Impairment | ✓ | 未进行年度减值测试披露 | 监控并在下季更新 | High |
| 5a. CapEx | ✓ | 维护 vs 增长拆分不清 | 按历史 D&A 推算维护 CapEx | Med |
| 5b. Debt Maturity | ✓ | — | — | — |
| 6. 最新季度指引 | ✓ | Q2 guidance range 宽泛 | 采用中点值 | Low |

**扫描结论**：
- 高优先级缺口：2 项（财报未披露，需要估算）
  - 补救：[具体调整]

- 模型已调整：所有关键假设已验证或合理估算
- **建议**：下个季度重点关注 Goodwill 减值风险

```

### 模型再校验（Model Reconciliation After Scan）

如发现缺口，执行以下重新校验：

1. **Rerun DCF**（如果关键假设改变）
2. **Rerun IRR & MoIC Sensitivity**（基于新的 CapEx / Working Capital 假设）
3. **Update Scenarios**（如果 Bear Case 的风险因子增加）

### Layer 4 交付总结（Delivery Summary）

```
## Model Delivery Summary

### Base Case Return Profile
| 指标 | 数值 |
|------|------|
| Entry EBITDA Multiple | X.Xx |
| Exit EBITDA Multiple | Y.Yy |
| Year 3 Revenue | $XXX |
| Year 3 EBITDA | $XXX |
| **IRR** | **X%** |
| **MoIC** | **X.Xx** |
| Exit Equity Value | $XXX |

### 概率加权 IRR
**Probability-Weighted IRR = Bull IRR × P(Bull) + Base IRR × P(Base) + Bear IRR × P(Bear)**
= X% × 25% + Y% × 50% + Z% × 25%
= **W%**

### 估值总结
| 估值方法 | 目标价格 | Upside/Downside |
|---------|--------|----------------|
| SOTP Framework | $X.XX/share | +/-Y% |
| DCF (Terminal g=2.0%) | $X.XX/share | +/-Y% |
| EV/EBITDA Multiple (Exit) | $X.XX/share | +/-Y% |
| **加权平均目标价** | **$X.XX/share** | **—** |

### 关键假设敏感性
**对 Base Case IRR 影响最大的 3 个变量**：
1. Exit Multiple（±X%）→ IRR ±Z ppt
2. Revenue CAGR（±X%）→ IRR ±Z ppt
3. WACC（±50 bps）→ IRR ±Z ppt

### 风险评分
| 风险因子 | 评分（1-5）| 缓释措施 |
|---------|----------|--------|
| 竞争加剧 | 3 | 产品差异化，定价权 |
| 宏观衰退 | 3 | FCF 防守，低杠杆 |
| 管理层变动 | 2 | 核心人员激励保留 |
| 杠杆风险（BT） | 2 | 逐年降低，利率对冲 |

### 模型置信度（Confidence Rating）
**综合得分：X/10**
- 数据完整性：X/10
- 假设合理性：X/10
- 市场验证：X/10

### 后续行动项
- [ ] 等待 [日期] 季报确认...
- [ ] 完成 [特定假设] 验证
- [ ] 跟进 [战略事件] 进展

```

---

## 第 7 节：IRR 合理性检验（Per Core Module Gate 4）

来自 core module 的强制 Gate：

```
## IRR Reasonableness Check (Gate 4)

所有模型在交付前必须通过以下检验：

### 检验 1：IRR vs Revenue Growth 一致性
- 如果 Revenue CAGR = X%, IRR 应在 [Y% - Z%] 范围内
- 公式：IRR ≥ Revenue CAGR（如果 Margin 稳定或扩张）
- 如果 IRR < Revenue CAGR → 需解释资本效率差原因

### 检验 2：IRR vs Margin Expansion
- 如果 EBITDA Margin 从 X% → Y%（+Z ppt），IRR 加成应为 ±A ppt
- 公式：Margin ppt × EBITDA Leverage ≈ IRR 贡献

### 检验 3：IRR vs Entry/Exit Multiple
- Entry Multiple X.Xx, Exit Multiple Y.Yy
- 如果 Exit Multiple < Entry Multiple，则需要超出 EBITDA Growth 的收益来达成目标 IRR
- Multiple Arbitrage 不应是主要驱动（<30% of IRR）

### 检验 4：Probability-Weighted IRR vs Target
- 加权 IRR 应高于 Fund 的目标回报（如：12% for PE, 18% for HF）
- 如果未达到，需重新评估假设或放弃投资

### 检验 5：Stress Test Coverage
- Bear Case IRR ≥ 8-10%（PE 的最低底线）
- 如果 Bear Case IRR < 8%，风险溢价不足

**检验通过标准**：
✓ 所有 5 项检验均通过，或
✓ 未通过项有明确商业原因支持（且已文档化）

```

---

## 第 8 节：数据完整性矩阵（New — Data Completeness Matrix）

### 设计原则

将年报完整性扫描**前置到 Layer 1**，作为数据尽调的一部分，而非 Layer 4 的事后检查。

### Layer 1 前置检查

在进入 Layer 2 前，自动生成以下矩阵：

```
## Data Completeness Matrix (Layer 1)

### 优先级 1：关键数据（必须有）
| 数据项 | 2024A | 2023A | 2022A | 可用性 | 缺口处理 |
|--------|-------|-------|-------|--------|---------|
| Revenue | ✓ | ✓ | ✓ | 100% | — |
| EBITDA / Operating CF | ✓ | ✓ | ✓ | 100% | — |
| CapEx | ✓ | ✓ | ✓ | 100% | — |
| Net Debt | ✓ | ✓ | ✓ | 100% | — |
| Tax Rate | ✓ | ✓ | ✓ | 100% | — |

### 优先级 2：中等重要（建模需要）
| 数据项 | 披露 | 是否估算 | 风险 |
|--------|------|--------|------|
| Deferred Revenue | ✓ | — | Low |
| COGS Breakdown | × | 需估算 | Med |
| D&A Policy | △（部分） | 需补充 | Med |
| CapEx by Type | × | 按历史比例推算 | High |

### 优先级 3：补充信息（加值但非必需）
| 数据项 | 披露 | 影响 |
|--------|------|------|
| 地域 DSO | × | Customer Mix Assumption |
| SBC 摊销 | ✓ | DCF 调整因子 |
| Lease Obligation | △ | Balance Sheet Quality |

### 整体数据质量评分
- **关键数据完整率**: X%（目标 ≥95%）
- **补充数据完整率**: Y%（目标 ≥70%）
- **综合评分**: Z/10

### 缺口风险等级
- 🟢 Green：所有关键数据可用，可放心建模
- 🟡 Yellow：1-2 项数据需合理估算，可进行，需标记敏感性
- 🔴 Red：>2 项关键数据缺失，需补充尽调后再建模

**本项目数据质量：[Green / Yellow / Red]**

```

### Layer 4 最终协调（Light Reconciliation Only）

Layer 4 只进行轻量级的"最终核对"，而非完整重扫：

```
## Layer 4 Data Reconciliation (Light)

### Final Spot Checks
- [ ] 财报数字输入无误（抽查 3-5 个关键行项）
- [ ] 新季度数据已整合（若有最新季报）
- [ ] DCF Terminal Value 假设无变化
- [ ] IRR Sensitivity 反映最新假设

### 无需重复扫描的项目
✓ 收入脚注结构（Layer 1 已确认）
✓ 成本结构（Layer 1 已确认）
✓ 资本配置意向（Layer 1 已确认）

**只需更新**：[如果有新披露的季度数据或管理层指引变化]

```

---

## 使用流程总结（Quick Reference）

### 用户触发 → 自动路由

| 用户输入 | 建议 Track | 建议 Granularity | 流程 |
|---------|-----------|-----------------|------|
| "综合建模 Apple" | 诊断 → (A/B/C) | 诊断 → (L1/L2/L3) | Layer 1 → ... |
| "建模 Apple, Hedge Fund 视角" | A (HF) | 诊断 | Layer 1 → ... |
| "建模 Apple, Buyout 分析" | B (BT) | 诊断 | Layer 1 → ... |
| "Quick-Build Apple" | 诊断 → (A/B/C) | L2 | Layer 3 直接（简版） |
| "Four-Layer Apple" | 诊断 → (A/B/C) | 确认 | 全 Layer 1-4 |

---

## 模块引用速查表

| 模块名 | 加载条件 | 提供的内容 |
|--------|---------|---------|
| `integrated-modeling-core` | 始终 | 框架、DCF 引擎、数据管理、Gate 4 |
| `integrated-modeling-hf` | Track A/C | 共识分歧、短期催化、Thesis Validation |
| `integrated-modeling-bt` | Track B/C | 管理层评估、100-Day Plan、Exit 蓝图、PE 风险 |
| `financial-modeling-dcf` | 始终（引擎） | WACC、Terminal Value、Sensitivity |

---

## 常见问题与快速答案

**Q1：如果用户在 Layer 2 中更改概率分布，是否需要重新计算 Layer 3？**
A：是的。更新 JSON 确认单，重新生成 Scenario Definitions（第 4 节），但建模结构不变。

**Q2：年报完整性扫描在哪一层执行？**
A：Layer 1 前置（6 类检查清单），Layer 4 仅进行轻量级协调。

**Q3：如果 Bear Case IRR < 8%，是否应该放弃投资？**
A：不一定。需要 Layer 2 中与用户确认风险承受度。如果用户明确接受，可以继续。

**Q4：SOTP 框架何时使用？**
A：仅限 Track C（Full Model）的 Layer 1 末尾，作为补充估值。

**Q5：100-Day Plan 具体由谁执行？**
A：Layer 3 生成框架，Layer 4 细化时间表与关键人员分工。

---

**版本**: 1.0
**最后更新**: 2026-03-20
**维护者**: Integrated Modeling Team
