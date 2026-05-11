> 本文件是 `full-spectrum-modeling` 的参考资料。
> 核心框架见 `../integrated-modeling-core/SKILL.md`。

# TAM自下而上定量方法论 (TAM Bottom-Up Quantification Methodology)

> **💀 曼巴注记 — 来源**: Project Stratus的TAM sheet (128行×29列) 从GDP出发，逐层渗透到可寻址市场。这不是"TAM = $100B"这种PPT数字，这是可审计、可回溯的完整推导链。

## 适用场景

Layer 1研究阶段，当需要量化市场机会时，**强制使用本方法论**。禁止直接引用第三方报告的TAM数字而不做独立推导。

## 五层TAM推导框架

```
Layer 1: 宏观经济基数 (Macro Foundation)
  GDP (名义) → 行业占GDP比重 → 行业总产出

Layer 2: IT/技术支出占比 (IT Spending Penetration)
  行业总产出 × IT支出占比(%) → 行业IT总支出

Layer 3: 目标技术渗透率 (Technology Penetration)
  行业IT总支出 × 目标技术渗透率(%) → 目标技术市场规模 (TAM)

Layer 4: 可寻址市场 (Serviceable Addressable Market, SAM)
  TAM × 地理限制 × 客户规模限制 × 技术限制 → SAM

Layer 5: 可获取市场 (Serviceable Obtainable Market, SOM)
  SAM × 公司市占率假设 → SOM ≈ Revenue Target验证
```

## 实操模板 (以Stratus云基础设施为例)

```
TAM CALCULATION WORKSHEET
─────────────────────────────────────────────────

                          FY2021    FY2022E   FY2025E
Layer 1: Macro
  Korea Nominal GDP (T KRW)  1,898     2,010     2,250

Layer 2: IT Penetration
  Enterprise IT Spend/GDP     3.2%      3.4%      3.8%
  Enterprise IT Spend (B KRW) 60.7      68.3      85.5

Layer 3: Cloud Penetration
  Cloud % of IT Spend         12%       15%       25%
  Cloud Market Size (B KRW)   7.3       10.2      21.4

  By Segment:
  ├─ IaaS/PaaS               4.4B      6.1B      12.8B
  ├─ Managed Services         1.5B      2.0B       4.3B
  ├─ Professional Services    0.9B      1.2B       2.6B
  └─ SaaS                    0.5B      0.9B       1.7B

Layer 4: SAM (Addressable)
  Geographic: Korea only       100%      100%      100%
  Client Size: >100 employees  70%       70%       75%
  SAM (B KRW)                 5.1B      7.1B      16.1B

Layer 5: SOM (Company Share)
  Current Market Share         3.2%      3.0%      4.0%
  SOM = Revenue Implied (B KRW) 165B    209B      640B

  Actual/Projected Revenue     165B     210B      650B
  Variance                     0%       +0.5%     +1.6%
  ← Revenue forecast与TAM推导的偏差在5%以内 ✓
```

## TAM QA检查表

```
TAM VALIDATION CHECKLIST
─────────────────────────
☐ Layer 1数据来自官方统计(世界银行/IMF/国家统计局)
☐ Layer 2 IT支出占比与Gartner/IDC数据交叉验证
☐ Layer 3渗透率有历史趋势支持(非拍脑袋)
☐ Layer 4的限制条件已明确说明
☐ Layer 5 SOM与公司实际Revenue偏差<10%
☐ TAM CAGR与行业报告CAGR交叉验证(偏差<5ppts)
☐ 数据来源和日期已标注(Source + Date)
☐ 竞争对手的SOM之和<100% TAM (否则TAM低估)
```

## 决策规则

```
IF SOM / SAM > 30%
  THEN 红灯: 增长空间有限，除非SAM扩大

IF TAM CAGR < Revenue CAGR
  THEN 红灯: 公司增速不可持续(在抢份额且市场不增长)

IF TAM推导Revenue vs 模型Revenue偏差 > 15%
  THEN 黄灯: 需要解释为什么模型假设偏离TAM支持的范围
```
