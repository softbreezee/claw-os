---
name: investment-data-layer
type: skill
version: 1.0.0
description: 独立数据获取层（Standalone Data Acquisition Layer）。为投资分析提供统一、标准化的数据供给引擎。按优先级从 SEC EDGAR、MCP 连接器、网络搜索等来源获取财务数据，支持自动填充、数据验证和标准化输出。可单独使用，也可作为 financial-modeling、MAVI 等技能的数据底座。
author: Investment Intelligence Suite
standalone: true
---

# investment-data-layer — 独立数据获取技能

## ⚠️ 铁律：训练数据禁止用于财务数字

> **任何财务数字（收入、利润、资产、负债、股价等）均不得来自训练数据。**
> 所有数字必须实时从外部来源获取，并标注来源与时间戳。
> 无法获取的字段必须标记 `MISSING` 并提示用户手工输入。

---

## 1. 技能定位与触发条件

### 1.1 独立使用场景

```
触发示例（任何涉及获取真实财务数据的请求）：
  "帮我获取 AAPL 最近三年的财务数据"
  "拉一下 [公司名/股票代码] 的财报"
  "给我 [公司] 的市值、EV、倍数"
  "Quick-Build [公司名]" → 自动触发数据层
  "帮我看看 [公司] 的资产负债表"
```

### 1.2 作为其他技能的数据底座

```
调用方：financial-modeling, MAVI, investment-docs, gp-org-analysis
调用方式：在自身技能中声明 "依赖 investment-data-layer 获取财务数据"
输出格式：Schema 9 标准 JSON（见第 5 节）
```

### 1.3 不触发本技能的情况

```
- 用户要求解释财务概念（无需获取数据）
- 讨论投资逻辑/策略（无需真实数据）
- 已有用户提供的完整数据集
```

---

## 2. 数据源优先级体系

### 2.1 四级数据源

```
P1 — PRIMARY（最高优先级，官方披露，应优先使用）
  • SEC EDGAR XBRL（美股：10-K, 10-Q, 8-K 结构化数据）
  • 港交所披露易（港股年报/中报）
  • 上交所/深交所（A股年报）
  • 公司官方 IR 网站直接下载

P2 — SECONDARY（机构级数据，高可信度）
  • Capital IQ（MCP 连接器）
  • Bloomberg（MCP 连接器）
  • FactSet（MCP 连接器）
  • Refinitiv / LSEG

P3 — TERTIARY（网络公开信息，需交叉验证）
  • Macrotrends / StockAnalysis / SimplyWallSt
  • 财经媒体数据（Bloomberg News, Reuters）
  • 公司 Investor Presentation / Earnings Call

P4 — INFERRED（推算，需标注 [INFERRED]）
  • 根据已知数据推算（如 Revenue × 行业均值 Margin）
  • 用于私有公司或数据缺口较大场景
  • 置信度标注为 LOW，需用户确认
```

### 2.2 强制规则

```
✅ P1 可用时必须使用 P1，不得用 P2 替代
✅ 使用 P3/P4 时必须显式标注来源和置信度
❌ 训练数据（模型内部知识）禁止作为财务数字来源
❌ 不得将"约"/"大约"/"估计"的未标注数字视为数据
```

---

## 3. MCP 连接器协议

### 3.1 连接器自动检测

在执行数据拉取前，依次检测可用连接器：

```
Step 0: 连接器可用性检测
  检测顺序: sec_edgar → capital_iq → bloomberg → factset → web_search
  
  IF sec_edgar.enabled:
    → 用于 P1 历史财务数据（IS/BS/CF）
  IF capital_iq.enabled:
    → 用于市场数据、NTM 共识预期、债务明细
  IF bloomberg.enabled:
    → 用于实时价格、Beta、期权数据
  IF factset.enabled:
    → 用于共识预期、行业分类、Comps 批量数据
  IF web_search.enabled:
    → 用于 P3 补充（新闻、IR 页面、Investor Day）
  ELSE:
    → 进入手工模式：逐项提示用户输入
```

### 3.2 数据类型与连接器路由

| 数据类型 | 首选连接器 | 备选连接器 | 最终手工输入 |
|---------|----------|----------|----------|
| 历史 IS/BS/CF（年报） | sec_edgar (P1) | capital_iq | 用户输入 |
| LTM 财务数据 | capital_iq | factset | 用户输入 |
| 实时股价 / 市值 | bloomberg | capital_iq | 用户输入 |
| EV 计算 | capital_iq | 自动计算 | 自动计算 |
| NTM 共识预期 | factset | bloomberg | 用户输入 |
| Beta（Blume-adjusted）| bloomberg | capital_iq | 用户输入 |
| 债务明细 | sec_edgar (10-K Notes) | capital_iq | 用户输入 |
| 行业分类 (GICS) | capital_iq | factset | 用户确认 |
| Comps 批量数据 | factset | capital_iq | 逐家输入 |

### 3.3 Fallback 链

```python
# 伪代码: Fallback 调度逻辑
def fetch_data_point(data_type: str, entity: str) -> DataPoint:
    chain = get_fallback_chain(data_type)  # 按优先级排列的连接器列表
    
    for connector in chain:
        if not connector.is_enabled():
            continue
        try:
            raw = connector.fetch(entity, data_type)
            if validate_raw(raw):
                return DataPoint(
                    value=raw.value,
                    source=connector.id,
                    source_priority=connector.priority,  # P1/P2/P3/P4
                    timestamp=now(),
                    confidence=map_confidence(connector.priority)
                )
        except ConnectorError as e:
            log_failure(connector.id, data_type, e)
            continue
    
    # 所有连接器失败
    return DataPoint(
        value=None,
        source="MANUAL",
        confidence="LOW",
        flag="MANUAL_ENTRY_REQUIRED",
        prompt=f"请手工输入 {entity} 的 {data_type}"
    )
```

---

## 4. Quick-Build 自动填充

### 4.1 触发流程

用户提供公司名称或股票代码后，自动执行：

```
Step 1: 实体识别 (Entity Resolution)
  → 确认股票代码、交易所、公司全称
  → 解决多重上市问题（ADR/H股/A股）
  → 获取 GICS 行业分类

Step 2: 拉取 3 年历史财务数据（IS/BS/CF）
  → 来源: P1 (SEC EDGAR / 官方年报)
  → 字段: 见 4.2 的 40+ 字段清单

Step 3: 拉取市场数据
  → 来源: P2 (Bloomberg / Capital IQ)
  → 字段: 股价、市值、EV、Beta、52周区间

Step 4: 拉取 NTM 共识预期
  → 来源: P2 (FactSet / Bloomberg BEST)
  → 字段: NTM Revenue, EBITDA, EPS，分析师人数

Step 5: 数据验证（见第 5 节）

Step 6: 输出标准化 JSON（Schema 9）+ 展示摘要
```

### 4.2 自动填充字段清单（40+ 字段）

**损益表 (Income Statement)**：
```
revenue, revenue_growth_yoy
gross_profit, gross_margin
sga_expense, sga_pct_revenue
rd_expense, rd_pct_revenue
ebitda, ebitda_margin
da, ebit, ebit_margin
interest_expense, tax_expense, tax_rate (ETR)
net_income, net_margin
diluted_eps, diluted_shares
```

**资产负债表 (Balance Sheet)**：
```
cash_and_equivalents, short_term_investments
accounts_receivable, inventory
total_current_assets, total_assets
accounts_payable, accrued_liabilities
short_term_debt, long_term_debt
total_interest_bearing_debt
total_current_liabilities, total_liabilities
total_equity, retained_earnings
net_debt [计算: total_debt - cash - st_investments]
```

**现金流量表 (Cash Flow Statement)**：
```
cfo (operating cash flow)
capex (capital expenditures), capex_pct_revenue
fcf (unlevered free cash flow) [计算: ebit*(1-t) + da - capex - nwc_change]
nwc_change (change in net working capital)
dividends_paid, share_repurchases
net_debt_change
```

**市场数据**：
```
stock_price, market_cap, enterprise_value
ev_to_ltm_revenue, ev_to_ltm_ebitda, ev_to_ntm_revenue, ev_to_ntm_ebitda
pe_ntm, peg_ratio
beta_adjusted (Blume 5yr weekly)
week_52_low, week_52_high
ntm_revenue_consensus, ntm_ebitda_consensus, ntm_eps_consensus
analyst_count
```

### 4.3 交互展示模板

```
🔍 已为 [公司名 (TICKER)] 自动获取数据（截至 [日期]）

━━━ 损益表摘要 (单位: $M) ━━━
  Revenue:           FY22A=[X]  FY23A=[X]  FY24A=[X]  [来源: 10-K, 置信度 ✅]
  EBITDA Margin:     FY22=[X%]  FY23=[X%]  FY24=[X%]  [来源: 10-K, 置信度 ✅]
  Net Income:        FY22=[X]   FY23=[X]   FY24=[X]   [来源: 10-K, 置信度 ✅]

━━━ 市场数据 ━━━
  Stock Price:       $[X]       [来源: Bloomberg, [日期]]    ✅
  Market Cap:        $[X]M      [计算: 股价 × 稀释股数]        ✅
  Enterprise Value:  $[X]M      [计算: 市值 + 净债务]          ✅
  EV/NTM Revenue:    [X]x       [来源: FactSet 共识]           🟡
  EV/NTM EBITDA:     [X]x       [来源: FactSet 共识]           🟡
  分析师数量:         [X] 位                                   🟡

━━━ 数据质量 ━━━
  数据置信度: [X] 个 HIGH, [X] 个 MEDIUM, [X] 个 LOW
  ⚠️ 低置信度字段: [字段名] — [原因]
  ❌ 缺失字段: [字段名] — 请手工输入

回复 "确认" 使用以上数据，或 "修改 [字段] 为 [值]" 进行调整。
```

---

## 5. 数据验证规则

### 5.1 财务三表一致性检验

```
RULE F1: IS ↔ CF 勾稽
  CF Statement 顶行 Net Income ≈ IS Net Income（允许差异 < 2%）
  
RULE F2: BS ↔ CF 勾稽
  期末现金 (BS) = 期初现金 + 经营/投资/融资现金流净额
  留存收益变动 = Net Income - Dividends（允许差异 < 2%）
  
RULE F3: EV 计算验证
  EV = Market Cap + 计息债务 - 现金 - 短期投资
  IF ∣报告 EV - 计算 EV∣ / EV > 5%: 触发 ALERT

RULE F4: 利率合理性
  隐含利率 = Interest Expense / Avg 计息债务
  IF 隐含利率 > 20%: ALERT（需用户确认）
  IF 隐含利率 < 0.5%: ALERT（可能遗漏债务）
```

### 5.2 增长率与比率合理性检验

```
RULE M1: 增速合理性
  Revenue CAGR < 200% (超过则标记异常)
  EBITDA Margin ∈ [-50%, 85%]
  Gross Margin ∈ [-10%, 100%]
  CapEx % Revenue ∈ [0%, 50%]
  
RULE M2: NWC 合理性
  NWC % Revenue ∈ [-30%, 60%]
  IF NWC > 60% Revenue: 标记为异常，需检查行业特性

RULE M3: 数据内部一致性
  LTM = 过去 12 个月（需标注截止日期）
  NTM 来源需 ≥ 2 位分析师，否则降级为 LOW 置信度
```

### 5.3 时效性检验

```
RULE T1: 市场数据不超过 1 个交易日（否则标注 STALE）
RULE T2: 共识预期不超过 30 天（否则标注 STALE）
RULE T3: 年报财务数据不超过 18 个月（否则请求用户确认最新披露）
RULE T4: Comps 数据必须与目标公司数据同一 as_of_date
```

### 5.4 置信度标注体系

```
✅ HIGH    — P1 来源（官方审计年报），误差 < 2%
🟡 MEDIUM  — P2 来源（机构数据库，未审计），误差 < 10%
⚠️  LOW     — P3/P4 来源（推算或网络数据），误差可能 > 20%
❌ MISSING  — 数据缺失，需用户手工输入

[INFERRED] 标签 — 任何推算数据，模型中必须显示此标注
[STALE] 标签   — 超过时效性规则的数据
```

---

## 6. 标准化输出 Schema（Schema 9）

所有数据输出遵循以下 JSON 结构，供其他技能消费：

```json
{
  "schema_version": "9",
  "generated_at": "ISO-8601 timestamp",
  "entity": {
    "name": "公司全称",
    "ticker": "股票代码",
    "exchange": "NYSE / NASDAQ / HKEX / SSE / ...",
    "gics_sector": "GICS 一级行业",
    "gics_industry": "GICS 细分行业",
    "fiscal_year_end": "月份（如 December）",
    "currency": "USD / HKD / CNY / ..."
  },
  "historical_financials": {
    "unit": "millions",
    "years": ["FY2022", "FY2023", "FY2024"],
    "revenue":          [X, X, X],
    "revenue_growth":   [null, X, X],
    "gross_profit":     [X, X, X],
    "gross_margin":     [X, X, X],
    "sga":              [X, X, X],
    "rd":               [X, X, X],
    "ebitda":           [X, X, X],
    "ebitda_margin":    [X, X, X],
    "da":               [X, X, X],
    "ebit":             [X, X, X],
    "ebit_margin":      [X, X, X],
    "interest_expense": [X, X, X],
    "tax_expense":      [X, X, X],
    "etr":              [X, X, X],
    "net_income":       [X, X, X],
    "net_margin":       [X, X, X],
    "diluted_eps":      [X, X, X],
    "diluted_shares":   [X, X, X],
    "capex":            [X, X, X],
    "capex_pct_rev":    [X, X, X],
    "cfo":              [X, X, X],
    "fcf":              [X, X, X],
    "nwc_change":       [X, X, X],
    "cash":             [X, X, X],
    "total_debt":       [X, X, X],
    "net_debt":         [X, X, X],
    "total_assets":     [X, X, X],
    "total_equity":     [X, X, X],
    "source": "10-K FY2024 (filed YYYY-MM-DD) via SEC EDGAR",
    "confidence": "HIGH"
  },
  "market_data": {
    "as_of_date": "YYYY-MM-DD",
    "stock_price": X,
    "market_cap": X,
    "enterprise_value": X,
    "ev_ltm_revenue": X,
    "ev_ltm_ebitda": X,
    "ev_ntm_revenue": X,
    "ev_ntm_ebitda": X,
    "pe_ntm": X,
    "beta_adjusted": X,
    "week_52_low": X,
    "week_52_high": X,
    "consensus_estimates": {
      "ntm_revenue": X,
      "ntm_ebitda": X,
      "ntm_eps": X,
      "analyst_count": X,
      "source": "FactSet consensus YYYY-MM-DD",
      "confidence": "MEDIUM"
    }
  },
  "debt_breakdown": {
    "total_interest_bearing_debt": X,
    "items": [
      {
        "type": "Senior Notes / Term Loan / ...",
        "amount": X,
        "maturity": "YYYY-MM",
        "rate_type": "fixed / floating",
        "rate": X
      }
    ],
    "maturity_profile": {"2025": X, "2026": X, "2027+": X},
    "source": "10-K FY2024, Note X — Long-term Debt",
    "confidence": "HIGH"
  },
  "data_quality_log": [
    {
      "field": "字段名",
      "confidence": "HIGH / MEDIUM / LOW / MISSING",
      "flag": "STALE / INFERRED / MANUAL_ENTRY_REQUIRED / null",
      "note": "说明"
    }
  ]
}
```

---

## 7. 独立使用模式

本技能可在**不触发任何建模或分析**的情况下独立运行：

### 7.1 典型独立请求

```
用户: "给我 AAPL 过去三年的财务数据"
→ 执行: 实体识别 → P1 历史财务拉取 → 验证 → Schema 9 输出 → 用户友好摘要
→ 不触发: financial-modeling, MAVI, investment-docs

用户: "TSLA 的 EV 和倍数是多少？"
→ 执行: 市场数据拉取 → EV 计算 → 倍数计算 → 输出
→ 不触发: 任何下游分析技能
```

### 7.2 处理私有公司数据

```
P1/P2 数据不可用时的降级策略:
1. 检索公开的 Investor Presentation / 融资公告（P3）
2. 使用行业均值推算（P4，标注 [INFERRED]）
3. 提示用户提供数据：
   "无法从公开数据源获取 [公司] 的财务数据（私有公司）。
    请提供以下信息：[字段清单]"
```

### 7.3 Comps 批量拉取

```
输入: peer_list = ["AAPL", "MSFT", "GOOG", ...]
处理: 并发请求（max 5 连接器并发），失败自动重试（max 3次）
输出: 每家公司的 Schema 9 子集 + 统一 as_of_date 标注
注意: 跨公司财年不一致时，自动转换为 Calendar Year (CY) 口径
```

---

## 8. 其他技能的消费接口

### 8.1 声明依赖

其他技能在自身 SKILL.md 中声明：
```
data_dependency: investment-data-layer
data_schema: Schema 9
```

### 8.2 消费方式

```python
# 消费方技能的伪代码
def get_company_data(ticker: str) -> Schema9:
    # investment-data-layer 负责数据获取
    data = investment_data_layer.fetch(ticker)
    
    # 检查数据质量
    critical_missing = [f for f in data.quality_log 
                        if f.confidence == "MISSING" and f.is_critical]
    if critical_missing:
        raise DataGapError(f"关键数据缺失: {critical_missing}")
    
    return data
```

### 8.3 数据质量门控

```
消费方技能在使用数据前应检查 data_quality_log:
  - confidence=MISSING 且为关键字段 → 阻止继续，要求用户补充
  - confidence=LOW 且为非关键字段 → 标注 [INFERRED]，允许继续
  - flag=STALE → 警告用户数据时效性，允许继续（用户确认后）
```

---

## 9. 快速参考

### 触发词

```
"获取 / 拉取 / 查一下 [公司/股票代码] 的数据/财务"
"Quick-Build [公司名]"
"[公司] 的市值/EV/倍数是多少"
"给我 [公司] 的资产负债表/利润表/现金流"
"帮我准备 [公司] 的数据用于建模"
```

### 常见问题

**Q: 没有 Bloomberg/FactSet 怎么办？**
A: 启用 sec_edgar（免费）获取 P1 历史财务，市场数据和 NTM 预期需手工输入。NTM 数据标注 `MANUAL`，置信度 LOW。

**Q: 如何处理中国 A 股？**
A: 使用上交所/深交所官方公告（P1），单位通常为 CNY 百万元，自动转换并标注货币。

**Q: 数据版本控制？**
A: 每次拉取附带 `generated_at` 时间戳和 `as_of_date`。消费方技能缓存数据时应记录这两个字段。

---

**版本**: 1.0.0 | **所属**: Investment Intelligence Suite | **最后更新**: 2026-04-16
