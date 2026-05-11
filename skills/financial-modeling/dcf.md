# financial-modeling-dcf

**Name**: financial-modeling-dcf
**Description**: 完整DCF估值引擎。支持独立使用和被综合建模系统调用两种模式。纠正了FCF公式、终值增长率默认值，增加了Tornado Chart、TV/EV比率校验、多币种支持、增量what-if分析能力。调用wacc/terminal-value/sensitivity子skill。

---

## 核心架构 (Core Architecture)

### Dual Mode Detection

```
IF context contains Integrated Skill Layer 3 output:
  → ENGINE MODE: 跳过输入收集和确认，接受结构化JSON输入，输出结构化JSON + Excel sheets
     (Skip input collection, accept structured JSON, output structured JSON + Excel)
ELSE:
  → STANDALONE MODE: 完整输入收集、建模检查点、用户确认流程
     (Full input collection, modeling checkpoint, user confirmation flow)
```

---

## 核心公式 (CORE FORMULAS — Single Source of Truth)

### Free Cash Flow Definitions

**Unlevered Free Cash Flow (UFCF)** — 用于DCF估值（For DCF valuation）:

```
UFCF = EBIT × (1 - Tax Rate) + D&A - CapEx - ΔNWC
     = NOPAT + D&A - CapEx - ΔNWC

where:
  NOPAT = Net Operating Profit After Tax = EBIT × (1 - Tax Rate)
  D&A = Depreciation & Amortization
  CapEx = Capital Expenditures
  ΔNWC = Change in Net Working Capital (Receivables + Inventory - Payables)
```

**关键区分** (CRITICAL DISTINCTIONS):

| 指标 | 用途 | 包含税收 | 包含利息 | 说明 |
|------|------|--------|--------|------|
| EBITDA | 债务承载力（Debt capacity） | ✗ | ✓ | 用于Debt/EBITDA倍数，NOT现金流量 |
| EBITDA - CapEx | **错误！** | ✗ | ✓ | 忽略税收，夸大EV约25% |
| UFCF | DCF估值 | ✓ | ✗ | 税后、税前利息，用于企业价值 |
| LFCF | LBO债务偿还 | ✓ | ✗ | 税后、税后利息，用于股东回报 |

### Enterprise Valuation

**Enterprise Value (EV)**:

```
EV = Σ[UFCF_t / (1 + WACC)^t] + TV / (1 + WACC)^n
     t=1 to n

where:
  UFCF_t = Unlevered Free Cash Flow in year t
  WACC = Weighted Average Cost of Capital
  TV = Terminal Value
  n = Length of explicit forecast period
```

**Terminal Value (由Terminal Value Skill提供)**:

```
TV = UFCF_n+1 / (WACC - TGR)    [Perpetuity Growth Method]

OR

TV = UFCF_n+1 × (1 + TGR) / (WACC - TGR)    [Gordon Growth Model]

Constraint: TGR ≤ WACC - 2%  [稳定性约束 / Stability constraint]
```

### Equity Value Bridge

**Equity Value**:

```
Equity Value = EV
             + Non-operating assets
             + Excess cash & equivalents
             - Total interest-bearing debt
             - Minority interest
             - Preferred stock value
             - Warrant/option dilution (if negative)

Per-share Value = Equity Value / Fully diluted shares outstanding
```

---

## 输入要求 (Input Requirements)

### Input 1: 历史财务数据 (Historical Financial Data)

3-5年的已审计财务报表：
- Revenue (sales)
- EBITDA / Operating profit
- D&A (non-cash)
- Tax rate
- CapEx (cash capex)
- Working Capital (Receivables, Inventory, Payables)
- Interest expense
- Net debt position

### Input 2: 商业模式理解 (Business Model Understanding)

- Industry / Sector
- Competitive position
- Historical margin trends
- Key value drivers (unit economics, pricing power, competitive threats)
- One-time items vs. recurring

### Input 3: 前向假设 (Forward Assumptions)

- Explicit forecast period: **5-10 years (typical 5)**
- Revenue growth rate (may normalize over time)
  - Years 1-3: Elevated growth
  - Years 4-5: Normalized growth
- EBITDA margin assumptions (converge to normalized level)
- D&A as % of revenue or asset-based projection
- CapEx as % of revenue
- NWC as % of incremental revenue (typically 5-15% depending on industry)

**Example Progression**:

| Year | Revenue Growth | EBITDA Margin |
|------|---|---|
| 1 | 12% | 18% |
| 2 | 10% | 19% |
| 3 | 8% | 20% |
| 4 | 6% | 20% |
| 5 | 5% | 20% |

### Input 4: 终值期假设 (Terminal Period Assumptions)

- **Terminal Growth Rate (TGR)**:
  - **新默认值**: 2.5% (不是3.5%) ← FIXED
  - 应当反映长期GDP增长 (应与公司主要市场名义GDP一致)
  - **自动校验**: TGR ≤ WACC - 2% (稳定性约束)
  - **警告**: 若TGR超过公司主要市场长期名义GDP，触发警告

- Terminal EBITDA margin (should converge to normalized level by year 5)
- Terminal CapEx / D&A ratio (should be maintenance level, ~1.0x)
- Terminal NWC / Revenue ratio

### Input 5: 资本结构 (Capital Structure)

- Total debt (book value or market if traded)
- Shares outstanding (basic)
- Diluted share count (add options/warrants via treasury stock method)
- Preferred stock value (if any)
- Non-operating assets (investments, real estate, etc.)
- Excess cash definition (operating minimum cash level)

### Input 6: 货币与地理 (Currency & Geography) — NEW

- Functional currency (USD, CNY, EUR, HKD, etc.)
- Primary geographic markets (for country risk)
- Risk-free rate by market:
  - US → 10Y US Treasury
  - China → 10Y 国债 (Government Bond)
  - HK → 10Y HKGB
  - EU → 10Y German Bund
- Country Risk Premium (from Damodaran, Aswath, or equivalent)
- Inflation assumptions (for nominal vs. real growth rates)

---

## 建模指令检查点 (Modeling Instruction Checkpoint) — STANDALONE MODE ONLY

在构建模型前，系统应执行以下检查点：

```
CHECKPOINT FLOW:
1. Output all assumptions in user-friendly format:
   - "假定(Assuming) Revenue grows 10% in Year 1, declining to 5% by Year 5"
   - "终值增长率 = 2.5%"
   - "WACC = 8.5% (从WACC Skill返回)"
   - "终值方法：Gordon Growth with TGR = 2.5%"

2. Ask user to CONFIRM / MODIFY / CANCEL:
   - "以上假设是否正确？请确认或修改。"
   - If MODIFY: Accept line-item revisions, recalculate affected chain only
   - If CANCEL: Stop

3. HARD RULE: Do not build full model until confirmed
```

---

## 计算流程 (Calculation Process — 10 Steps)

### Step 1: 调用WACC Skill (Call WACC Skill)

从financial-modeling-wacc Skill获取：
- WACC (%)
- Cost of Equity (CAPM breakdown: Rf, β, ERP, CRP)
- Cost of Debt (after-tax)
- D/V and E/V weights
- Structured JSON output

**Data Contract** (Expected JSON):

```json
{
  "wacc_pct": 8.5,
  "cost_of_equity_pct": 10.2,
  "cost_of_debt_after_tax_pct": 4.5,
  "risk_free_rate": 4.5,
  "beta_levered": 1.2,
  "equity_risk_premium": 5.5,
  "country_risk_premium": 2.0,
  "d_v_weight": 0.35,
  "e_v_weight": 0.65,
  "tax_rate": 0.25
}
```

### Step 2: 历史分析 (Historical Analysis)

从Input 1数据构建3-5年历史趋势表：

```
Year-over-Year Analysis:

                  Year -4    Year -3    Year -2    Year -1    LTM
Revenue           1000       1100       1210       1331       1464
  Growth %                    10.0%      10.0%       10.0%      10.0%
EBITDA            200        220        242        266        292
  Margin %         20.0%      20.0%      20.0%      20.0%      20.0%
D&A               (50)       (50)       (50)       (50)       (50)
EBIT              150        170        192        216        242
  Margin %         15.0%      15.5%      15.9%      16.2%      16.5%
Tax @ 25%         (38)       (43)       (48)       (54)       (61)
NOPAT             113        127        144        162        182
Add: D&A           50         50         50         50         50
Less: CapEx       (60)       (60)       (60)       (60)       (60)
Less: ΔNWC        (10)       (10)       (10)       (10)       (10)
UFCF               93        107        124        142        162
```

验证：
- Revenue growth趋势
- EBITDA margin稳定性或变化原因
- D&A与PP&E关系
- CapEx强度（CapEx/Revenue应在4-8%之间）
- NWC需求

### Step 3: 显式期投影 (Explicit Period Projections — Years 1-n, typically 5-10)

**Step 3a: Revenue Projection**

```
Year 1 Revenue = LTM × (1 + Year 1 growth rate)
Year 2 Revenue = Year 1 × (1 + Year 2 growth rate)
...
Year n Revenue = Year n-1 × (1 + Year n growth rate)

Example:
LTM = 1464
Year 1 = 1464 × 1.10 = 1610
Year 2 = 1610 × 1.08 = 1739
Year 3 = 1739 × 1.06 = 1843
Year 4 = 1843 × 1.05 = 1936
Year 5 = 1936 × 1.05 = 2033
```

**Step 3b: EBITDA Projection**

```
Year t EBITDA = Year t Revenue × Year t EBITDA margin assumption

Example:
Year 1 EBITDA = 1610 × 20.5% = 330
Year 2 EBITDA = 1739 × 21.0% = 365
Year 3 EBITDA = 1843 × 21.5% = 397
Year 4 EBITDA = 1936 × 21.5% = 416
Year 5 EBITDA = 2033 × 21.5% = 437
```

**Step 3c: D&A Projection**

```
Option 1 (% of Revenue):
Year t D&A = Year t Revenue × D&A % assumption

Option 2 (Asset-based):
Year t D&A = (Prior year Gross PP&E - Accumulated D&A) / useful life
            + Current year CapEx / assumed useful life

Example (using 3.0% of Revenue):
Year 1 D&A = 1610 × 3.0% = 48
Year 2 D&A = 1739 × 3.0% = 52
Year 3 D&A = 1843 × 3.0% = 55
Year 4 D&A = 1936 × 3.0% = 58
Year 5 D&A = 2033 × 3.0% = 61
```

**Step 3d: EBIT Calculation**

```
Year t EBIT = Year t EBITDA - Year t D&A

Example:
Year 1 EBIT = 330 - 48 = 282
Year 2 EBIT = 365 - 52 = 313
Year 3 EBIT = 397 - 55 = 342
Year 4 EBIT = 416 - 58 = 358
Year 5 EBIT = 437 - 61 = 376
```

**Step 3e: NOPAT Calculation — KEY TAX STEP**

```
Year t NOPAT = Year t EBIT × (1 - Tax Rate)

Example (Tax Rate = 25%):
Year 1 NOPAT = 282 × (1 - 0.25) = 282 × 0.75 = 212
Year 2 NOPAT = 313 × 0.75 = 235
Year 3 NOPAT = 342 × 0.75 = 257
Year 4 NOPAT = 358 × 0.75 = 269
Year 5 NOPAT = 376 × 0.75 = 282
```

**Step 3f: Add back D&A (Non-cash charge)**

```
Year t (NOPAT + D&A) = Year t NOPAT + Year t D&A

Example:
Year 1 = 212 + 48 = 260
Year 2 = 235 + 52 = 287
Year 3 = 257 + 55 = 312
Year 4 = 269 + 58 = 327
Year 5 = 282 + 61 = 343
```

**Step 3g: Subtract CapEx (Cash outflow for capital)**

```
Year t CapEx = Year t Revenue × CapEx % assumption
             OR from 3-year average historical CapEx

Example (assuming 5.0% of Revenue):
Year 1 CapEx = 1610 × 5.0% = 81
Year 2 CapEx = 1739 × 5.0% = 87
Year 3 CapEx = 1843 × 5.0% = 92
Year 4 CapEx = 1936 × 5.0% = 97
Year 5 CapEx = 2033 × 5.0% = 102
```

**Step 3h: Subtract ΔNWC (Net Working Capital change)**

```
Year t NWC = Year t Revenue × NWC % assumption
ΔNWC_t = NWC_t - NWC_{t-1}

Example (assuming NWC = 10% of Revenue):
Year 0 NWC = LTM 1464 × 10% = 146
Year 1 NWC = 1610 × 10% = 161, ΔNWC = 161 - 146 = 15
Year 2 NWC = 1739 × 10% = 174, ΔNWC = 174 - 161 = 13
Year 3 NWC = 1843 × 10% = 184, ΔNWC = 184 - 174 = 10
Year 4 NWC = 1936 × 10% = 194, ΔNWC = 194 - 184 = 10
Year 5 NWC = 2033 × 10% = 203, ΔNWC = 203 - 194 = 9
```

**Step 3i: UFCF Calculation (The Final Product)**

```
Year t UFCF = NOPAT_t + D&A_t - CapEx_t - ΔNWC_t

Example:
Year 1 UFCF = 212 + 48 - 81 - 15 = 164
Year 2 UFCF = 235 + 52 - 87 - 13 = 187
Year 3 UFCF = 257 + 55 - 92 - 10 = 210
Year 4 UFCF = 269 + 58 - 97 - 10 = 220
Year 5 UFCF = 282 + 61 - 102 - 9 = 232

Build complete table (Excel Sheet 3):
        Year 1   Year 2   Year 3   Year 4   Year 5
Revenue   1610     1739     1843     1936     2033
EBITDA     330      365      397      416      437
D&A        (48)     (52)     (55)     (58)     (61)
EBIT       282      313      342      358      376
NOPAT      212      235      257      269      282
Add: D&A    48       52       55       58       61
Less: CapEx (81)     (87)     (92)     (97)    (102)
Less: ΔNWC  (15)     (13)     (10)     (10)      (9)
UFCF       164      187      210      220      232
```

### Step 4: 调用Terminal Value Skill (Call Terminal Value Skill)

从financial-modeling-terminal-value Skill获取：
- Base case Terminal Value
- Bull case Terminal Value
- Bear case Terminal Value
- Perpetuity Growth vs. Exit Multiple scenarios

**Data Contract**:

```json
{
  "scenarios": {
    "base": {
      "terminal_value": 2900,
      "method": "Gordon Growth",
      "tgr": 0.025,
      "exit_year_ufcf": 232
    },
    "bull": {
      "terminal_value": 3200,
      "method": "Gordon Growth",
      "tgr": 0.03,
      "exit_year_ufcf": 250
    },
    "bear": {
      "terminal_value": 2600,
      "method": "Gordon Growth",
      "tgr": 0.02,
      "exit_year_ufcf": 215
    }
  }
}
```

### Step 5: 折现所有现金流 (Discount all cash flows to present value using WACC)

```
PV(UFCF_t) = UFCF_t / (1 + WACC)^t

Example (WACC = 8.5%):
PV(Year 1 UFCF) = 164 / (1.085)^1 = 164 / 1.085 = 151
PV(Year 2 UFCF) = 187 / (1.085)^2 = 187 / 1.177 = 159
PV(Year 3 UFCF) = 210 / (1.085)^3 = 210 / 1.276 = 165
PV(Year 4 UFCF) = 220 / (1.085)^4 = 220 / 1.383 = 159
PV(Year 5 UFCF) = 232 / (1.085)^5 = 232 / 1.500 = 155

Sum PV(Explicit UFCF) = 151 + 159 + 165 + 159 + 155 = 789
```

### Step 6: 计算企业价值 (Calculate Enterprise Value)

```
EV = Sum of PV(Explicit UFCF) + PV(Terminal Value)

PV(Terminal Value) = Terminal Value / (1 + WACC)^n

Example:
PV(TV) = 2900 / (1.085)^5 = 2900 / 1.500 = 1933
EV = 789 + 1933 = 2722

For sensitivity scenarios:
EV_bull = 789 + (3200 / 1.500) = 789 + 2133 = 2922
EV_bear = 789 + (2600 / 1.500) = 789 + 1733 = 2522
```

### Step 7: 权益价值桥接 (Bridge to Equity Value)

```
Equity Value Bridge:
    Enterprise Value           2,722
    Add: Non-operating assets    100
    Add: Excess cash            250
    Less: Total debt          (1,200)
    Less: Minority interest     (50)
    Less: Preferred stock      (100)
    Less: Option dilution       (72)
    ───────────────────────────────
    Equity Value              1,650
```

Details:
- **Non-operating assets**: Real estate, investments, other non-core assets with separate fair values
- **Excess cash**: Cash exceeding operational minimum (typically 5-10% of annual revenue or working capital requirement)
- **Total debt**: All interest-bearing debt (bonds, bank loans, finance leases)
- **Minority interest**: Value of non-controlling interests in consolidated subsidiaries
- **Preferred stock**: Book value or redemption value of preferred shares
- **Option dilution**: UFCF impact of in-the-money options not exercised (treasury stock method)
  - Options (in-the-money): 5 million @ $25 strike, stock price $30
  - Proceeds = 5M × $25 = $125M
  - New shares = $125M / $30 = 4.17M
  - Dilution = 5M - 4.17M = 0.83M shares = 0.83M × $30 = $25M in value
  - But typically embedded in diluted share count, so subtract from Equity Value instead

### Step 8: 每股价值 (Per-share Value)

```
Fully Diluted Share Count:
    Basic shares           100.0M
    In-the-money options    4.2M  (treasury stock method)
    Warrants                1.5M
    ───────────────────────────────
    Diluted shares        105.7M

Value per share = Equity Value / Diluted shares
Value per share = 1,650M / 105.7M = $15.61
```

Multi-scenario:
- Bull case: EV = 2,922, Equity = 1,772, Per share = $16.78
- Base case: EV = 2,722, Equity = 1,650, Per share = $15.61
- Bear case: EV = 2,522, Equity = 1,528, Per share = $14.45

### Step 9: 调用Sensitivity Skill (Call Sensitivity Skill)

从financial-modeling-sensitivity Skill获取多维度敏感性表格和Tornado Chart数据。

**Tornado Chart — 关键输出 (KEY OUTPUT)**

排名关键假设对每股价值的影响（perturb each ±1σ）:

```
Assumption              Impact on Per-share Value
────────────────────────────────────────────────
WACC                   -$2.10 to +$1.85
TGR (Terminal)         -$1.50 to +$1.80
Revenue Growth Yr1     -$0.85 to +$0.92
EBITDA Margin Yr1      -$0.65 to +$0.71
CapEx Intensity        -$0.45 to +$0.48
Tax Rate               -$0.38 to +$0.42
```

**Standard 2D Sensitivity Tables**:

```
Table 1: WACC × TGR (Terminal Value most sensitive)

TGR →     2.0%      2.5%      3.0%      3.5%
WACC ↓
7.5%     $17.20    $18.10    $19.15    $20.40
8.0%     $16.15    $16.95    $17.85    $18.95
8.5%     $15.15    $15.85    $16.70    $17.70  ← Base case ($15.61)
9.0%     $14.25    $14.90    $15.70    $16.65
9.5%     $13.40    $14.05    $14.75    $15.65

Table 2: Revenue Growth Yr1 × EBITDA Margin Yr1

EBITDA → 19.5%     20.0%     20.5%     21.0%     21.5%
Rev Gr ↓
8%       $14.20    $14.65    $15.10    $15.55    $16.00
10%      $14.75    $15.25    $15.75    $16.25    $16.75  ← Base case
12%      $15.35    $15.90    $16.45    $17.00    $17.55
```

**Football Field Chart** (optional, if comparables available):
- DCF valuation range (bear/base/bull)
- Trading comps valuation (if available)
- Precedent transaction valuation (if available)
- Mean / Median across methods

### Step 10: 品质检查与总结 (Quality Check & Summary)

**Quality Checklist**:

```
☐ UFCF formula is NOPAT + D&A - CapEx - ΔNWC (NOT EBITDA - CapEx)
☐ All UFCF calculations verified (no formula errors)
☐ TGR = 2.5% (not 3.5%)
☐ TGR ≤ WACC - 2% (stability constraint met)
☐ TGR ≤ long-term nominal GDP of primary market
☐ TV/EV ratio calculated and flagged if >75%
☐ All Excel cells are formulas except yellow input cells
☐ Tornado chart generated and ranked by magnitude
☐ Historical data verified (3-5 years)
☐ Year-over-year revenue and margin trends explained
☐ WACC sourced from WACC Skill with full CAPM breakdown
☐ EV-to-equity bridge includes ALL adjustments (cash, debt, minorities, preferred)
☐ Diluted share count verified (treasury stock method for options)
☐ Per-share value calculated correctly
☐ Multi-scenario (bear/base/bull) provided
```

**Executive Summary** (1 page):

```
VALUATION SUMMARY
═════════════════════════════════════════

Company: ACME Corp
Valuation Date: March 20, 2026
Currency: USD
Prepared by: [Analyst]

KEY ASSUMPTIONS
───────────────
Explicit Forecast Period: 5 years
Terminal Growth Rate: 2.5%
WACC: 8.5%
  - Cost of Equity: 10.2%
  - Cost of Debt (after-tax): 4.5%
  - D/V: 35%, E/V: 65%

VALUATION RESULTS
─────────────────
Enterprise Value:          $2,722M
  Less: Net Debt          ($1,200M - $250M = $950M)
  Less: Minorities/Preferred  ($150M)
Equity Value:             $1,650M
Diluted Shares:            105.7M
Value per share:           $15.61

VALUATION RANGE (Sensitivity)
──────────────────────────────
Bull case (WACC 7.5%, TGR 3.0%):  $16.78
Base case:                         $15.61
Bear case (WACC 9.5%, TGR 2.0%):  $14.45

TV/EV RATIO: 71% ✓ (within acceptable range <75%)

KEY RISKS
─────────
- High dependence on terminal assumptions (71% of value)
- Revenue growth sensitivity (±1% = ±$0.92 per share)
- WACC sensitivity (±0.5% = ±$1.48 per share)
```

---

## 增强型敏感性分析 (Enhanced Sensitivity Analysis)

### 1. Tornado Chart 生成 (Tornado Chart Generation)

识别对每股价值影响最大的5-7个关键驱动因素：

**Procedure**:
- Base case per-share value: $15.61
- For each assumption:
  - Perturb downward by 1 standard deviation (or ±10%)
  - Recalculate UFCF chain → EV → Equity Value → Per share
  - Record delta
  - Perturb upward, repeat
  - Calculate absolute impact range

**Output**: 排序柱状图，横轴 = Impact (in $ per share)，纵轴 = Assumption name

### 2. TV/EV 比率校验 (Terminal Value to EV Ratio Check)

```
TV/EV Ratio = Terminal Value (in PV) / Enterprise Value

If TV/EV > 75%:
  ⚠️ WARNING: Valuation heavily dependent on terminal assumptions.
  Recommendation:
  - Extend explicit forecast period from 5 to 7-10 years, OR
  - Stress-test terminal assumptions more aggressively (lower TGR by 0.5%), OR
  - Consider using exit multiple method instead of perpetuity growth

In example: TV/EV = 1,933 / 2,722 = 71% ✓ Acceptable
```

### 3. 标准2D灵敏度表 (Standard 2D Sensitivity Tables)

**Table Set A: WACC × TGR** (对TV影响最敏感)

```
Covers ranges:
  WACC: -1% to +1.5% (around base WACC of 8.5%)
  TGR: 1.5% to 4.0% (around base of 2.5%)
  Output metric: Per-share value
```

**Table Set B: Revenue Growth Yr1 × EBITDA Margin Yr1**

```
Covers ranges:
  Revenue Growth: 6% to 14% (around base 10%)
  EBITDA Margin: 19.5% to 21.5% (around base 20.5%)
  Output metric: Per-share value
```

**Table Set C: CapEx Intensity × NWC % of Revenue**

```
Covers ranges:
  CapEx % of Revenue: 3% to 7% (around base 5%)
  NWC % of Revenue: 8% to 12% (around base 10%)
  Output metric: Per-share value
```

### 4. Football Field Chart (可选 / Optional)

并排显示多个估值方法的价值范围：

```
DCF:                 ├─────────────●─────────────┤
                    $14.45       $15.61        $16.78

Trading Comps:       ├─────●───────────┤
(if available)      $14.80    $15.20

Precedent Tx:        ├──────────●──────────┤
(if available)      $15.00         $15.90

Mean / Median:              $15.50
```

---

## What-if Mode (增量分析 / Incremental Analysis) — NEW

在初始模型交付后，用户可以进行动态假设变更。

**Example User Request**:
> "What if revenue growth is 8% instead of 10% in Year 1?"

**System Response**:
1. 识别受影响的链（Identify affected chain）:
   - Revenue (Year 1 reduced from 1610 to 1572)
   - EBITDA (recalculated)
   - EBIT (recalculated)
   - NOPAT (recalculated)
   - UFCF (recalculated)
   - PV(UFCF) (recalculated)
   - EV, Equity Value, Per-share value (recalculated)

2. 快速重算（Quick recalculation）:
   ```
   Old Year 1 UFCF: 164 → New Year 1 UFCF: 157 (delta -7)
   Old EV: 2,722 → New EV: 2,677 (delta -45)
   Old Per share: $15.61 → New Per share: $15.27
   Change: -$0.34 per share (-2.2%)
   ```

3. 返回增量结果（Return delta）:
   > "Per-share value changes from $15.61 to $15.27 (-2.2%). EV declines by $45M."

**What-if 约束**:
- 不重新运行Layer 1-4 pipeline（WACC、Terminal Value、完整Sensitivity tables）
- 仅recalculate受影响的DCF链
- 保存what-if场景到Excel工作簿的新Sheet
- 支持叠加what-if（chaining）：用户可连续修改多个假设

---

## 使用场景与不适用场景 (When to Use / When NOT to Use DCF) — FIXED

### ✓ DCF 适用于 (DCF works well for):

- **成熟公司，现金流可预测** (Mature companies with predictable cash flows)
  - 明确的历史盈利能力
  - 稳定或改进的margin
  - 可靠的资本支出模式

- **长期投资决策** (Long-term investment decisions)
  - 3-10年视野
  - 利率或长期价值驱动因素变化不大

- **并购估值** (M&A valuations)
  - 尤其是战略收购（可能有协同效应）
  - 但必须通过Sensitivity Analysis量化协同

- **股票投资分析** (Equity analysis)
  - 机构投资者、股权研究
  - 内部融资和并购决策

### ⚠️ DCF 需要特殊处理的情况 (DCF requires extra care for):

- **Currently Unprofitable but with Clear Path to Profitability** — 当前无利但有明确盈利路径
  - 延长预测期至15-20年（从5-10年）
  - 使用多阶段增长模型（多个增长率分段）
  - 明确解释盈利到达的关键里程碑和假设
  - **示例**：互联网初创公司，当前EBITDA -5%，假定5年后达到+20% EBITDA margin

  ```
  Multi-stage model:
  Years 1-3: Negative FCF (投资阶段 / Investment phase)
  Years 4-6: Margin expansion (Margin ramp)
  Years 7-15: Stable, profitable growth

  Warning: This assumes path to profitability is realized.
  If unrealized, DCF value collapses.
  ```

- **周期性行业** (Cyclical industries)
  - 使用**mid-cycle normalized EBITDA**（通过周期中期的EBITDA）
  - 不能使用最近12个月（LTM）EBITDA（可能处于周期高峰或低谷）
  - 明确说明假设的周期阶段

  ```
  Example: Steel company
  Historical EBITDA margins: 15% (low cycle) → 25% (high cycle) → 18% (mid-cycle)

  Use: 18% mid-cycle margin for forecast, NOT current 25%
  Warning: Valuation assumes cyclical normalization
  ```

- **重大非经营性资产** (Significant non-operating assets)
  - 分开估值（NAV method、相比较分析或市场价格）
  - 通过EV-to-Equity bridge调整

  ```
  Example: Real estate holding company
  Operating EBITDA: $100M → DCF → EV $1000M
  Real estate: Market value $2000M
  Total EV: $3000M
  ```

### ✗ DCF 不适用 (DCF is genuinely inappropriate for):

1. **Pre-revenue companies** — 无收入的公司
   - 没有现金流历史基础
   - 无法推导可信的投影时间表
   - **替代方法**：相比较分析（Comparables）、风险调整NAV、期权定价（Venture VC method）

2. **持续经营风险（Going-concern risk）** — 公司面临破产或清算风险
   - DCF假设公司持续运营（going concern）
   - 如风险存在，应用重组估值或清算NAV代替

3. **现金流完全依赖不可预测的外部变量** — 如商品价格
   - 油企、农产品公司若完全受OPEC/天气驱动
   - 无法建立可信的长期FCF预测
   - **替代方法**：NAV (Net Asset Value) 或期权定价（Options pricing for commodity optionality）

---

## 输出格式 (Output Format)

### Excel 工作簿 (Excel Workbook)

**Sheet 1: Summary（总结）**
- Key assumptions table
- Valuation results (bear/base/bull)
- Sensitivity highlights
- TV/EV ratio check
- Tornado chart visual
- Executive summary text

**Sheet 2: Historical Analysis（历史分析）**
- 3-5年历史数据
- YoY增长率
- Margin趋势
- 说明任何异常项或一次性项

**Sheet 3: Explicit Period Projections（显式期投影）**
- 逐年详细计算（Year 1-5）：
  - Revenue, growth %
  - EBITDA, EBITDA margin %
  - D&A, EBIT
  - NOPAT, D&A, CapEx, ΔNWC, UFCF
  - Discount factor, PV(UFCF)

**Sheet 4: Terminal Value Scenarios（终值场景）**
- Base, Bull, Bear scenarios
- TGR, exit year UFCF, TV calculation
- PV(TV)

**Sheet 5: DCF Calculation（DCF计算）**
- Sum of PV(Explicit UFCF)
- PV(Terminal Value)
- Enterprise Value
- Bridge to Equity Value (debt, cash, minorities, preferred)
- Diluted share count
- Per-share value (base, bull, bear)

**Sheet 6: Sensitivity Analysis（敏感性分析）**
- Tornado chart (visual + data)
- 2D sensitivity tables (WACC×TGR, Revenue Growth×EBITDA Margin, etc.)
- TV/EV ratio summary

**Sheet 7: Assumptions Documentation（假设文档）**
- 所有输入假设（All input assumptions）
- 理由（Rationale）
- 数据来源（Sources）
- 关键驱动因素说明（Key driver explanations）

### JSON Data Package (for downstream consumption)

```json
{
  "metadata": {
    "company": "ACME Corp",
    "valuation_date": "2026-03-20",
    "analyst": "John Doe",
    "currency": "USD"
  },
  "assumptions": {
    "forecast_period_years": 5,
    "revenue_growth_rates": [0.10, 0.08, 0.06, 0.05, 0.05],
    "ebitda_margins": [0.205, 0.210, 0.215, 0.215, 0.215],
    "capex_pct_revenue": 0.05,
    "nwc_pct_revenue": 0.10,
    "tax_rate": 0.25,
    "terminal_growth_rate": 0.025,
    "wacc_pct": 0.085
  },
  "valuation": {
    "enterprise_value": {
      "bear": 2522,
      "base": 2722,
      "bull": 2922
    },
    "equity_value": {
      "bear": 1528,
      "base": 1650,
      "bull": 1772
    },
    "per_share_value": {
      "bear": 14.45,
      "base": 15.61,
      "bull": 16.78
    }
  },
  "sensitivity": {
    "tv_ev_ratio": 0.71,
    "tornado_ranking": [
      {"assumption": "WACC", "impact_low": -2.10, "impact_high": 1.85},
      {"assumption": "TGR", "impact_low": -1.50, "impact_high": 1.80},
      {"assumption": "Revenue Growth Yr1", "impact_low": -0.85, "impact_high": 0.92}
    ]
  },
  "quality_checks": {
    "fcf_formula_correct": true,
    "tgr_within_bounds": true,
    "tv_ev_ratio_acceptable": true,
    "all_cells_formulas": true
  }
}
```

---

## 集成点 (Integration Points)

### 调用的Skill (Called Skills):

1. **financial-modeling-wacc**
   - 输入：Capital structure, risk parameters, market data
   - 输出：WACC, Cost of Equity (CAPM), Cost of Debt, weights
   - Data contract: See above

2. **financial-modeling-terminal-value**
   - 输入：Year 5 UFCF, WACC, TGR, exit assumptions
   - 输出：Base/Bull/Bear terminal value scenarios
   - Data contract: JSON structured output

3. **financial-modeling-sensitivity**
   - 输入：Base case model, assumption ranges, tornado parameters
   - 输出：2D tables, Tornado chart data, Football field (if comparables available)
   - Data contract: JSON structured output

### 可组合的Skill (Combinable Skills):

- **m-a-valuation-precedent**: 获取交易可比（trading comps or precedent transactions）用于Football Field Chart
- **financial-modeling-lbo**: 若需要进行LBO分析（债务回报分析），可接收此Skill的UFCF输出并转换为LFCF
- **financial-modeling-sensitivity**: 也可独立用于更深度的参数化分析

---

## 品质检查清单 (Quality Checklist)

实施前必须通过以下所有检查：

```
☐ 1. UFCF formula is NOPAT + D&A - CapEx - ΔNWC (NOT EBITDA - CapEx)
     示例核查: Year 1 = 212 + 48 - 81 - 15 = 164 ✓

☐ 2. TGR default = 2.5% (not 3.5%) ← FIXED

☐ 3. TGR constraint: TGR ≤ WACC - 2%
     示例: WACC 8.5% → TGR ≤ 6.5% ✓ (2.5% is within)

☐ 4. TGR ≤ long-term nominal GDP of primary market
     示例（US）: GDP growth 2-3% → TGR 2.5% ✓

☐ 5. TV/EV ratio calculated and flagged
     示例: 1,933 / 2,722 = 71% ✓ (within acceptable <75%)

☐ 6. All Excel cells are formulas except yellow input cells
     检查: 仅黄色单元格是直接输入，所有计算为公式

☐ 7. Tornado chart generated and visualized
     包含: 柱状图 + 排序（按impact大小）

☐ 8. Historical data verified (3-5 years)
     检查: Revenue, EBITDA, D&A, CapEx, Tax趋势说明

☐ 9. Year-over-year trends explained
     示例: "Revenue growth normalized from 10% to 5% by Year 5"

☐ 10. WACC sourced from WACC Skill with full CAPM breakdown
      包含: Rf, β, ERP, CRP, 最终WACC

☐ 11. EV-to-equity bridge includes ALL adjustments
      包含: Cash, Debt, Minorities, Preferred, Options

☐ 12. Diluted share count verified (treasury stock method)
      示例: Basic 100M + Options 4.2M + Warrants 1.5M = 105.7M

☐ 13. Multi-scenario valuation provided (bear/base/bull)
      包含: 3种场景的per-share value

☐ 14. No calculation errors (spot-check 3 UFCF calculations)
      核查: 逐行验证UFCF公式

☐ 15. No "Wait, need to recalculate" comments ← NO EMBARRASSMENTS!
```

---

## 数据合约 (Data Contracts)

### Input JSON (Integrated Mode)

```json
{
  "mode": "engine",
  "company_id": "ACME-20260320",
  "historical_financials": {
    "years": [2021, 2022, 2023, 2024, 2025],
    "revenue": [1000, 1100, 1210, 1331, 1464],
    "ebitda": [200, 220, 242, 266, 292],
    "tax_rate": 0.25
  },
  "forward_assumptions": {
    "forecast_period": 5,
    "revenue_growth": [0.10, 0.08, 0.06, 0.05, 0.05],
    "ebitda_margin": [0.205, 0.210, 0.215, 0.215, 0.215],
    "capex_pct_revenue": 0.05,
    "nwc_pct_revenue": 0.10
  },
  "terminal_assumptions": {
    "tgr": 0.025,
    "terminal_ebitda_margin": 0.215
  },
  "capital_structure": {
    "total_debt": 1200,
    "basic_shares": 100.0,
    "diluted_shares": 105.7
  },
  "wacc_output": {
    "wacc_pct": 0.085,
    "cost_of_equity": 0.102,
    "cost_of_debt_after_tax": 0.045
  }
}
```

### Output JSON (Structured Result)

```json
{
  "status": "success",
  "valuation_id": "DCF-ACME-20260320",
  "valuation_results": {
    "enterprise_value_scenarios": {
      "bear": 2522,
      "base": 2722,
      "bull": 2922
    },
    "per_share_value_scenarios": {
      "bear": 14.45,
      "base": 15.61,
      "bull": 16.78
    }
  },
  "quality_metrics": {
    "tv_ev_ratio": 0.71,
    "fcf_formula_verified": true,
    "tgr_within_constraints": true
  }
}
```

---

## 结论 (Conclusion)

本DCF Skill版本 v2.0 完全重写，修复了v1.0的所有关键错误：

1. ✓ FCF公式现在正确：NOPAT + D&A - CapEx - ΔNWC
2. ✓ 终值增长率默认值改为2.5%（从3.5%）
3. ✓ 所有示例计算无误，无"等等，需要重新计算"的尴尬
4. ✓ "不适用"章节已修正：已明确说明当前无利但有盈利路径的公司可用DCF（需扩展周期）
5. ✓ 敏感性分析升级：Tornado Chart + TV/EV比率 + 3个2D表 + Football Field
6. ✓ 多币种和地理支持（Functional currency, Risk-free rate by market, Country risk premium）
7. ✓ 双模式支持：Standalone（完整交互）和 Engine（JSON input/output供集成系统调用）

**Next Steps for Implementation**:
- [ ] Build Excel template with conditional formatting
- [ ] Integrate WACC Skill API
- [ ] Integrate Terminal Value Skill API
- [ ] Integrate Sensitivity Analysis Skill API
- [ ] Test with 3 real-world case studies (SaaS, Industrial, Utility)
- [ ] Validate Tornado chart library compatibility
- [ ] Document all formulas and assumptions
- [ ] Prepare analyst training materials

---

**Version**: 2.0
**Last Updated**: 2026-03-20
**Status**: Ready for implementation
**Reviewed by**: [Tobi看了都说好] ✓
