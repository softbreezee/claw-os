---
name: full-spectrum-modeling-comps
type: module
description: 可比公司分析（Comps）模块。Peer Selection 系统化框架、多维度估值倍数、统计基准、异常值处理、与 DCF 交叉验证。当需要相对估值或市场定价参考时加载。
author: Full-Spectrum Modeling Framework
version: 4.0.0
---

# 可比公司分析模块 (Comparable Company Analysis Module)

## 1. 模块概述 (Module Overview)

### 什么是 Comps

可比公司分析（Comparable Company Analysis, "Comps"）是通过选取一组业务模式、规模、增速相似的上市公司，利用其市场估值倍数来推断目标公司的合理估值范围。它是相对估值法（Relative Valuation）的核心工具。

### 与 DCF 的互补关系

| 维度 | DCF | Comps |
|------|-----|-------|
| **方法论** | 内在价值 (Intrinsic Value) | 市场定价 (Market Pricing) |
| **核心输入** | 现金流预测 + WACC | 可比公司倍数 |
| **优势** | 不受市场情绪影响 | 反映当前市场定价 |
| **劣势** | 对终端假设敏感 | 受市场周期影响 |
| **最佳场景** | 稳定现金流、可预测 | 有充足可比公司 |

**核心原则**: DCF 告诉你"应该值多少"，Comps 告诉你"市场认为值多少"。两者偏差过大时，必须深入分析原因。

### 适用场景

| 场景 | 适用度 | 说明 |
|------|--------|------|
| 上市公司估值 | ✅ 强适用 | 充足的公开数据和可比公司 |
| M&A 定价参考 | ✅ 强适用 | 买方/卖方都需要市场基准 |
| IPO 定价 | ✅ 强适用 | 发行价需要同业参考 |
| 融资估值谈判 | ✅ 适用 | 投资人和创始人的共同语言 |
| Pre-revenue 公司 | ⚠️ 弱适用 | 缺乏可比的财务指标 |
| 高度独特商业模式 | ⚠️ 弱适用 | 难以找到真正可比的公司 |
| 破产/困境公司 | ❌ 不适用 | 正常倍数不反映困境价值 |

---

## 2. Peer Selection 框架 (Peer Selection Framework)

### 2.1 五步筛选流程

**Step 1: 初始候选池 (Universe Generation)**

从以下维度生成初始候选清单：

```
候选来源:
├─ 行业分类: GICS / BICS / NAICS 同行业代码
├─ 竞争对手: 年报 / 10-K 中提及的竞争对手
├─ 分析师覆盖: 同一分析师覆盖的公司群
├─ 投资者重叠: 相似的机构投资者持仓
└─ 供应链: 上下游关系密切的公司

目标: 生成 20-40 家初始候选
```

**Step 2: 业务模式筛选 (Business Model Filter)**

```
必须匹配:
  ☐ 核心收入模式相同 (订阅 vs 交易 vs 广告 vs 硬件)
  ☐ 客户类型相似 (B2B vs B2C vs B2B2C)
  ☐ 地理市场重叠度 > 50%

软性匹配:
  ☐ 产品线组合相似
  ☐ 技术栈/平台类型相似
  ☐ 监管环境相似

筛选后目标: 15-25 家
```

**Step 3: 财务特征筛选 (Financial Profile Filter)**

```
关键财务筛选维度:
├─ 收入规模: 目标公司 0.3x ~ 3.0x 范围内
│  (如目标 $1B revenue, 筛选 $300M ~ $3B)
├─ 增速匹配: Revenue CAGR 偏差 < 10 ppts
│  (如目标 15% CAGR, 筛选 5% ~ 25% CAGR)
├─ 盈利阶段: 盈利 vs 亏损阶段需匹配
│  (盈利公司不应与亏损公司对标)
└─ 资本结构: 杠杆水平大致相当
   (高杠杆 vs 轻资产差异太大会扭曲 EV 倍数)

筛选后目标: 10-18 家
```

**Step 4: 流动性与数据质量筛选 (Liquidity & Data Quality)**

```
硬性门槛:
  ☐ 日均交易量 > $1M (确保价格反映公允价值)
  ☐ 上市超过 12 个月 (避免 IPO 溢价扭曲)
  ☐ 有至少 2 家卖方分析师覆盖 (确保共识估计可用)
  ☐ 最近一期财报发布不超过 6 个月

软性考量:
  ☐ 无重大pending诉讼或监管调查
  ☐ 无 M&A speculation (避免收购溢价)
  ☐ 无近期重大一次性事项 (重组、减值等)

筛选后目标: 8-15 家
```

**Step 5: 最终精选与分组 (Final Selection & Tiering)**

```
最终选择: 6-12 家

分组策略:
├─ Tier 1 — 核心可比 (3-5 家)
│  业务模式、规模、增速高度匹配
│  权重: 60-70% in implied valuation
│
├─ Tier 2 — 扩展可比 (3-5 家)
│  部分维度匹配，提供额外参考
│  权重: 20-30%
│
└─ Tier 3 — 参考可比 (1-3 家)
   不同行业但相似商业模式 (cross-sector comps)
   权重: 5-10%
   例: Uber (ride-sharing) vs DoorDash (food delivery) — 双边平台逻辑相似
```

### 2.2 Peer Selection 常见陷阱

| 陷阱 | 描述 | 修复 |
|------|------|------|
| **规模偏差** | 选了市值 10x 差距的公司 | 严格执行 0.3x-3.0x 收入规模 |
| **增速错配** | 高增长 vs 成熟期混在一起 | 按增速分组 (>20%, 10-20%, <10%) |
| **地理偏差** | 新兴市场 vs 发达市场混合 | 注明地理溢价/折扣 |
| **幸存者偏差** | 只选表现好的公司 | 包含行业中位数和下四分位 |
| **时点偏差** | 使用过时的倍数 | 统一使用最新季度数据 |
| **收购目标** | 含被传收购的公司 | 剔除或单独标注收购溢价 |

> 详细的 Peer Selection 指南含行业案例和中国市场特殊考虑，见 `references/peer-selection-guide.md`。

---

## 3. 估值指标矩阵 (Valuation Metrics Matrix)

### 3.1 通用指标 (Universal Metrics)

| 指标 | 公式 | 适用场景 | 注意事项 |
|------|------|---------|---------|
| **EV/Revenue** | Enterprise Value / Revenue | 亏损公司、高增长公司 | 不反映盈利能力 |
| **EV/EBITDA** | Enterprise Value / EBITDA | 最通用的估值倍数 | 排除资本结构和税的影响 |
| **EV/EBIT** | Enterprise Value / EBIT | 资本密集型行业 | 包含 D&A, 反映真实运营利润 |
| **P/E** | Price / Earnings per Share | 盈利稳定的成熟公司 | 受资本结构影响 |
| **P/B** | Price / Book Value | 金融机构、重资产行业 | 不适用轻资产公司 |
| **EV/FCF** | Enterprise Value / Free Cash Flow | 现金流驱动的估值 | FCF 定义需统一 |

### 3.2 行业专用指标 (Industry-Specific Metrics)

#### SaaS / 订阅型
| 指标 | 公式 | 基准范围 |
|------|------|---------|
| **EV/ARR** | Enterprise Value / Annual Recurring Revenue | 10-30x (高增长), 5-15x (成熟) |
| **EV/NRR-Adjusted Revenue** | EV / (Revenue × Net Revenue Retention) | 强调留存质量 |
| **Rule of 40 Score** | Revenue Growth % + FCF Margin % | >40 = 优秀, 25-40 = 良好 |
| **CAC Payback** | CAC / (ARPU × Gross Margin) | <18mo = 优秀, 18-36mo = 可接受 |

#### 电商 / 平台型
| 指标 | 公式 | 基准范围 |
|------|------|---------|
| **EV/GMV** | Enterprise Value / Gross Merchandise Value | 0.3-1.5x (platform dependent) |
| **EV/Take-Rate Adjusted Revenue** | EV / (GMV × Take Rate) | 更准确反映平台收入 |
| **Price/Monthly Active Users** | Market Cap / MAU | $50-500 per user (varies widely) |
| **EV/Gross Profit** | Enterprise Value / Gross Profit | 10-25x (反映真实经济性) |

#### 金融服务
| 指标 | 公式 | 基准范围 |
|------|------|---------|
| **P/TBV** | Price / Tangible Book Value | 0.8-2.5x (banks), 1.5-4.0x (fintech) |
| **P/E (Normalized)** | Price / Normalized Earnings | 使用 mid-cycle earnings |
| **P/AUM** | Price / Assets Under Management | 1-5% (asset managers) |

#### 硬件 / 制造
| 指标 | 公式 | 基准范围 |
|------|------|---------|
| **EV/Units Shipped** | Enterprise Value / Annual Unit Shipments | 行业差异极大 |
| **EV/EBITDA (Normalized)** | 使用中周期 EBITDA | 消除周期性波动 |
| **EV/Installed Base** | Enterprise Value / Cumulative Installed Units | 反映 aftermarket 潜力 |

### 3.3 LTM vs NTM 选择

| 方法 | 定义 | 适用 | 不适用 |
|------|------|------|--------|
| **LTM** (Last Twelve Months) | 过去12个月实际数据 | 稳定业务、验证历史 | 高增长、转型期 |
| **NTM** (Next Twelve Months) | 未来12个月共识预期 | 增长型公司、前瞻估值 | 共识覆盖不足时 |
| **NTM+1** | 未来第二年预期 | 高增长公司、长期视角 | 预测不确定性大 |

**选择规则**:
```
IF revenue growth > 20%
  THEN 使用 NTM (前瞻倍数更能反映增长价值)
ELSE IF revenue growth 5-20%
  THEN 同时展示 LTM 和 NTM (提供全景)
ELSE
  THEN 使用 LTM (历史数据更可靠)
```

---

## 4. 统计分析 (Statistical Analysis)

### 4.1 基本统计量

对每个估值指标计算以下统计量:

```
Comps Summary Statistics:
├─ Mean (均值) — 简单平均和加权平均 (by market cap)
├─ Median (中位数) — 推荐作为核心参考
├─ P25 (25th percentile) — 下限参考
├─ P75 (75th percentile) — 上限参考
├─ Min / Max — 极值参考
└─ Standard Deviation (标准差) — 离散程度
```

**为什么中位数优于均值**: 均值容易被极端值拉偏。例如一组 EV/EBITDA: [8x, 10x, 12x, 14x, 45x]，均值 17.8x 被异常值严重拉高，中位数 12x 更具代表性。

### 4.2 异常值处理 (Outlier Treatment)

**标准方法: 2σ 规则**

```
Step 1: 计算均值 (μ) 和标准差 (σ)
Step 2: 标记 |x - μ| > 2σ 的数据点为异常值
Step 3: 分析异常值成因:

  IF 异常值因 M&A 溢价 → 剔除并注明
  IF 异常值因一次性事件 → 使用 Normalized 数据
  IF 异常值因商业模式差异 → 移至 Tier 3 参考
  IF 异常值反映真实市场定价 → 保留但单独注明
```

**替代方法: Winsorization**

```
将超过 P5/P95 的值截断到 P5/P95
适用场景: 样本量 > 15 且分布近似正态
```

### 4.3 Implied Valuation 计算

```
Step 1: 选择核心指标 (通常 2-3 个)
  例: EV/EBITDA + EV/Revenue + P/E

Step 2: 确定参考点
  ├─ 中位数 (核心参考)
  ├─ P25-P75 范围 (合理区间)
  └─ Tier 1 中位数 (精选参考)

Step 3: 计算 Implied Value
  Implied EV = Target_Metric × Peer_Multiple
  Implied Equity = Implied EV - Net Debt + Cash
  Implied Price per Share = Implied Equity / Diluted Shares

Step 4: 综合多指标
  Weighted Average = Σ(Implied Value_i × Weight_i)

  推荐权重:
  ├─ EV/EBITDA: 40-50% (最通用)
  ├─ EV/Revenue: 20-30% (增长公司)
  └─ P/E or EV/FCF: 20-30% (成熟公司)
```

### 4.4 置信度评估

| 条件 | 置信度 | 说明 |
|------|--------|------|
| Tier 1 可比 ≥ 5 家, 倍数离散度 < 30% | 🟢 高 | 结论可靠 |
| Tier 1 可比 3-4 家, 离散度 < 50% | 🟡 中 | 需结合 DCF 验证 |
| Tier 1 可比 < 3 家, 或离散度 > 50% | 🔴 低 | 仅作参考，不可单独依赖 |

---

## 5. DCF 交叉验证 (DCF Cross-Validation)

### 5.1 偏差分析框架

将 Comps Implied Value 与 DCF Intrinsic Value 对比:

```
Deviation = (Comps_Value - DCF_Value) / DCF_Value × 100%
```

| 偏差范围 | 判断 | 行动 |
|---------|------|------|
| **< 15%** | 🟢 一致 | Comps 和 DCF 互相确认，结论稳健 |
| **15-30%** | 🟡 需调查 | 分析偏差来源：市场情绪? 增长预期差异? 行业周期? |
| **> 30%** | 🔴 重大分歧 | 必须深入分析原因，不可忽略 |

### 5.2 常见偏差原因

```
Comps > DCF (市场比模型乐观):
├─ 市场处于乐观周期 (bubble risk)
├─ 可比公司含 M&A 溢价
├─ DCF 的 WACC 过高
└─ DCF 的 Terminal Growth 过低

Comps < DCF (市场比模型悲观):
├─ 市场处于恐慌/低迷期
├─ 可比公司基本面恶化
├─ DCF 的增速假设过于乐观
└─ 市场尚未认识到公司的价值
```

### 5.3 Football Field 图

将多种估值方法的结果并排展示:

```
估值方法              低值    中值    高值
─────────────────────────────────────────
DCF (Bear/Base/Bull)  ████████████████████
Comps (P25/Med/P75)     ██████████████████
Precedent Transactions    ████████████████
52-Week Range         ███████████████████████
Analyst Targets           ████████████████████
─────────────────────────────────────────
                      $20   $30   $40   $50   $60
```

**Football Field 构建规则**:
- 至少包含 3 种估值方法
- 每种方法显示低/中/高三个值
- 突出显示重叠区间 (consensus range)
- 标注当前市场价格

---

## 6. Excel 产出规格 (Excel Output Specification)

### Sheet 1: Comps Table (核心表)

**列结构**:

| 列 | 字段 | 格式 | 说明 |
|----|------|------|------|
| A | Company Name | Text | 公司全名 |
| B | Ticker | Text | 股票代码 |
| C | Tier | 1/2/3 | 可比分组 |
| D | Market Cap ($M) | `$#,##0` | 最新市值 |
| E | Enterprise Value ($M) | `$#,##0` | EV = Mkt Cap + Net Debt |
| F | LTM Revenue ($M) | `$#,##0.0` | 过去12个月收入 |
| G | NTM Revenue ($M) | `$#,##0.0` | 未来12个月共识 |
| H | LTM EBITDA ($M) | `$#,##0.0` | 过去12个月 EBITDA |
| I | NTM EBITDA ($M) | `$#,##0.0` | 未来12个月共识 |
| J | Revenue Growth (LTM) | `0.0%` | 收入增速 |
| K | EBITDA Margin (LTM) | `0.0%` | EBITDA 利润率 |
| L | EV/Revenue (LTM) | `0.0x` | 收入倍数 |
| M | EV/Revenue (NTM) | `0.0x` | 前瞻收入倍数 |
| N | EV/EBITDA (LTM) | `0.0x` | EBITDA 倍数 |
| O | EV/EBITDA (NTM) | `0.0x` | 前瞻 EBITDA 倍数 |
| P | P/E (NTM) | `0.0x` | 前瞻市盈率 |
| Q-Z | 行业专用指标 | varies | 按行业添加 |

**表尾统计行**: Mean, Median, P25, P75（分 Tier 1 / All Peers 两组）

### Sheet 2: Valuation Summary (估值汇总)

**结构**:

```
Section 1: Target Company Profile
  ├─ Key financials (Revenue, EBITDA, Net Income, FCF)
  ├─ Growth metrics
  └─ Current market pricing (if public)

Section 2: Implied Valuation
  ├─ By EV/Revenue: Low | Mid | High
  ├─ By EV/EBITDA: Low | Mid | High
  ├─ By P/E: Low | Mid | High
  └─ Weighted Average: Low | Mid | High

Section 3: Football Field Chart
  ├─ DCF range
  ├─ Comps range
  ├─ Precedent Txn range (if available)
  └─ Consensus range (overlap zone)

Section 4: Sensitivity
  ├─ Multiple sensitivity (±1x)
  └─ Growth sensitivity (±5 ppts)
```

**颜色规范**: 遵循 `config/color-spec.md` 统一标准。

---

## 7. 与 IM 架构的集成 (Integration with IM Architecture)

### 7.1 Layer 集成点

| Layer | Comps 集成方式 |
|-------|---------------|
| **Layer 1 (研究)** | Peer Selection 在 8D 诊断的"Competitive Position"维度中完成。Comps 分析结果为 Layer 1 提供市场定价背景。 |
| **Layer 2 (参数化)** | Comps 的 implied valuation 作为 DCF 假设的 sanity check。如 Comps 暗示 EV/EBITDA = 12x 但 DCF 推出 18x，需要审视假设。 |
| **Layer 4 (交付)** | Comps Table 和 Valuation Summary 作为 Excel Workbook 的独立 Sheet 交付。Football Field 图是最终 Presentation 的核心可视化。 |

### 7.2 Gate 4 新增检查项 (Comps)

```
☐ Peer group 满足最低 6 家要求
☐ Tier 1 可比公司至少 3 家
☐ 所有倍数数据时间戳一致 (同一季度)
☐ 异常值已处理并注明理由
☐ Implied valuation 与 DCF 偏差已分析
  ├─ 偏差 < 15%: 无需额外操作
  ├─ 偏差 15-30%: 偏差原因已记录
  └─ 偏差 > 30%: 深度分析报告已附
☐ Football Field 图至少包含 3 种估值方法
☐ 统计量计算正确 (Mean/Median/P25/P75)
```

### 7.3 数据契约

Comps 模块输出遵循 `data-contracts/schemas.md` 中的 Schema 8 格式:

```
Comps Module → Valuation Summary (Schema 8)
├─ peer_group: 可比公司列表及分组
├─ multiples: 各指标的 LTM/NTM 倍数
├─ statistics: Mean/Median/P25/P75
└─ implied_valuation: Weighted implied value
```

### 7.4 触发条件

```
IF track = HF (Hedge Fund)
  THEN Comps 是 Layer 1 的标准步骤 (快速估值参考)

IF track = BT (Buyout)
  THEN Comps 在 Layer 1 用于 Entry pricing
       在 Layer 4 用于 Exit multiple assumption

IF track = Full
  THEN Comps 贯穿 Layer 1-4
       Layer 1: Peer identification
       Layer 2: Assumption validation
       Layer 4: Final valuation triangulation
```

---

## 附录: 快速参考

### Comps 分析流程速查

```
1. Peer Selection (5步筛选) → 6-12 家可比
2. 数据收集 (FactSet/Bloomberg/Capital IQ)
3. 计算倍数 (LTM + NTM)
4. 统计分析 (Mean/Median/P25/P75)
5. 异常值处理 (2σ rule)
6. Implied Valuation (加权平均)
7. DCF 交叉验证 (偏差分析)
8. Football Field (可视化)
9. Gate 4 检查 → 交付
```

### 常见问题

**Q: 最少需要几家可比公司?**
A: 最低 6 家，其中 Tier 1 至少 3 家。如果可比公司不足 6 家，Comps 分析的置信度标为"低"，仅作辅助参考。

**Q: LTM 和 NTM 应该用哪个?**
A: 高增长公司 (>20%) 用 NTM，成熟公司用 LTM，中间地带两个都展示。

**Q: 如何处理不同财年结束日的公司?**
A: 使用 Calendar Year (CY) 统一口径，或明确标注各公司的 Fiscal Year End。

**Q: 中国公司有什么特殊考虑?**
A: 见 `references/peer-selection-guide.md` 中的中国市场专节，覆盖 A/H 股溢价、VIE 结构折扣、行业监管溢价/折扣等。

---

**版本**: 4.0.0 | **所属**: Full-Spectrum Modeling Framework | **最后更新**: 2026-04-12
