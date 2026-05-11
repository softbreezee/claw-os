---
name: integrated-modeling-private
description: 非上市公司综合建模入口。面向PE/VC投资视角的Investor Model，适配数据匮乏场景。含置信度标注体系、Adjusted EBITDA标准化、流动性折扣框架、非上市公司专属Revenue/Cost拆分策略。模块化引用core/hf/bt子skill。触发词：综合建模[非上市公司]、建模[非上市公司]。
---

# Integrated Modeling for Private Companies (非上市公司综合建模)

## Overview
This is the ENTRY POINT for **PRIVATE (unlisted) company modeling**. Same modular architecture as public version but with **critical adaptations** for:
- Data scarcity and partial disclosure
- PE/VC investment decision frameworks
- Adjusted EBITDA normalization (Singer standard)
- Liquidity discount & control premium frameworks
- Revenue/Cost split strategies adapted to data availability

**Core Philosophy**: Honesty about what we don't know. Every data point carries a confidence tag.

---

## 1. Module References (Modular Architecture)

### Core Dependencies
```
├── core/financial-model-builder
│   ├── 3-statement modeling (P&L, BS, CF)
│   ├── Working capital dynamics
│   └── Debt schedule construction
│
├── hf/comparable-company-analysis (HF-specific)
│   ├── EV/Revenue, EV/EBITDA multiples
│   ├── Private company adjustments
│   └── Minority-majority valuation bridges
│
└── bt/dcf-valuation
    ├── Terminal value estimation
    ├── WACC calibration for private firms
    └── Sensitivity analysis frameworks
```

### Private-Specific Extensions
```
├── normalized-ebitda-bridge
│   ├── Owner compensation adjustments
│   ├── Related-party transaction normalization
│   ├── One-time expense treatment
│   └── SBC (Stock-Based Compensation) estimation
│
├── liquidity-discount-framework (DLOM/Control Premium)
│   ├── Lack of Marketability Discount (15-35%)
│   ├── Control Premium (20-40%)
│   └── Minority Discount (20-30%)
│
├── private-company-research-adapter
│   ├── Data source credibility tracking
│   ├── Comparable company inference
│   └── Industry benchmark cross-checks
│
└── due-diligence-gap-list
    ├── Critical missing data identification
    └── Materiality threshold assessment
```

---

## 2. Trigger Conditions

### Activation Keywords (Chinese & English)
```
Primary triggers:
  - "综合建模[非上市公司]"
  - "建模[非上市公司]"
  - "非上市公司 建模"
  - "Private company modeling"
  - "PE modeling [private]"
  - "VC投资 建模"

Secondary triggers (specific use cases):
  - "收购建模 + [公司名] + [非上市]"
  - "估值 + [公司名] + [私企/非上市]"
  - "EBITDA归一化"
  - "Adjusted EBITDA"
  - "PE investor model"
```

### Pre-Conditions Check
1. User has identified a specific private company (or provide company profile)
2. User clarifies investment thesis: minority stake / majority acquisition / full buyout
3. User confirms available data: audited financials / partial data / no data

If pre-conditions not met → **Mandatory First Question** (see Layer 1)

---

## 3. NEW — Confidence Level System (Addresses Data Transparency)

### Confidence Framework: Every Data Point Tagged

This system directly addresses the problem of data scarcity in private company modeling. Each assumption, data point, and calculated metric receives a **confidence level** with color coding and sensitivity treatment guidelines.

| **Data Source** | **Confidence Level** | **Tag** | **Sensitivity Treatment** |
|---|---|---|---|
| User-provided audited financials (Full audit trail) | ■■■■■ High | `[HIGH]` Blue | Standard assumption range ±5% |
| User verbal guidance or unaudited data (Provided but not verified) | ■■■■□ Medium-High | `[MED-HIGH]` Blue + light yellow BG | ±10% range on base case |
| Comparable company inference (From public comp multiples or private databases) | ■■■□□ Medium | `[MED]` Orange font | ±20% range, sensitivity testing required |
| Industry average / rough estimate (From industry reports, benchmarks) | ■■□□□ Low | `[LOW]` Red font | ±30% range, independent scenario analysis |
| Pure guess / no data support (Placeholder, must be replaced) | ■□□□□ Very Low | `[VERY-LOW]` Red font + red BG | **MUST do independent sensitivity**; flag for due diligence |

### Practical Application
- Every revenue line item tagged with data source
- Every cost assumption carries confidence level
- Adjusted EBITDA bridge shows each adjustment's confidence
- Model output displays sensitivity ranges aligned to confidence levels
- "Very Low" confidence items trigger automatic due diligence flag

### Example Tags in Practice
```
Revenue (2024):
  ├─ Direct Sales: $10M [HIGH] (User provided, audited)
  ├─ Distribution Channel: $5M [MED] (Estimated from comparable company mix)
  └─ Online Channel: $2M [VERY-LOW] (Rough estimate, needs DD confirmation)

Cost of Goods Sold:
  ├─ COGS % of Revenue: 35% [MED-HIGH] (User guidance + partial year data)
  ├─ Manufacturing overhead: 8% [MED] (Industry benchmark range 7-10%)
  └─ Logistics cost: 3% [LOW] (Estimated from similar-sized competitors)

Owner Compensation:
  ├─ Actual paid: $1.5M [HIGH] (Tax returns provided)
  ├─ Market-rate equivalent: $800K [MED] (Salary comp database: 50th percentile)
  ├─ Adjustment for normalization: +$700K [MED] (Market-rate approach)
```

---

## 4. Layer 1: Researcher (Private Company Adaptation)

### 4.1 Mandatory First Question
Before proceeding with any analysis:

> **"What non-public financial information can you provide about [Company Name]?"**
>
> Please select your data availability level:
> - **[A] Full Financials**: Audited P&L, BS, CF for 3+ years + detailed breakdown of revenue/cost
> - **[B] Partial Data**: Some financials provided (e.g., revenue only, or recent year only) + qualitative details
> - **[C] No Financial Data**: Only business description, no numbers; sourcing from public announcements or industry inference
> - **[D] Confidential / NDA**: Company confidential; will work with limited disclosed metrics

**Credibility Warning Protocol**:
If [C] or [D] selected → **DISPLAY PROMINENTLY AT TOP OF RESEARCH MEMO**:
```
⚠️  CREDIBILITY WARNING
This analysis is based on limited/inferred data.
Confidence levels on all key assumptions are [LOW] or [VERY-LOW].
This is a PRELIMINARY FRAMEWORK ONLY and must be validated through formal due diligence.
Not suitable for final investment decision without data confirmation.
```

---

### 4.2 Research Dimensions (Same 8 as Public + Data Source Tracking)

For each dimension, track:
- Primary source: User provided / Annual report / News / Regulatory filing / Inferred from comparables
- Confidence level: [HIGH] / [MED-HIGH] / [MED] / [LOW] / [VERY-LOW]
- Alternative data points (if available)

#### Research Dimension 1: Business Model & Revenue Architecture
```
Questions:
├─ What are the primary revenue streams? (Count and % contribution)
├─ Which channels generate revenue? (Direct/Distribution/Online/Offline/Other)
├─ What is the customer concentration? (Top 5 customers as % of revenue)
├─ Are there long-term contracts or recurring revenue? (%)
└─ Growth trajectory (historical CAGR, forward guidance if available)

Data Source Template:
├─ Revenue breakdown: ______ [Confidence: ___]
├─ Customer concentration: ______ [Confidence: ___]
└─ Contract terms: ______ [Confidence: ___]
```

#### Research Dimension 2: Competitive Position & Market Share
```
Questions:
├─ What is the company's market share estimate in primary segment?
├─ Key competitors (private/public)?
├─ What is the competitive advantage (product/cost/brand)?
├─ Barriers to entry in this market?
└─ Industry growth rate (vs. company growth)?

Data Source Template:
├─ Market share estimate: ______ [Confidence: ___]
├─ Industry growth rate: ______ [Confidence: ___]
└─ Competitive moat assessment: ______ [Confidence: ___]
```

#### Research Dimension 3: Cost Structure & Unit Economics
```
Questions:
├─ Gross margin (reported / normalized)?
├─ Operating expense breakdown: Sales / G&A / R&D / Other?
├─ What drives COGS: materials / labor / manufacturing overhead?
├─ Are there one-time expenses or seasonality?
└─ Unit economics: revenue per employee, CAC, LTV (if B2B/B2C)?

Data Source Template:
├─ Gross margin: ______ [Confidence: ___]
├─ OpEx as % of revenue: ______ [Confidence: ___]
└─ Unit economics available: Yes / No [Confidence: ___]
```

#### Research Dimension 4: Management & Organization
```
Questions:
├─ Founder/CEO background and track record?
├─ Key management team (tenure, experience)?
├─ Employee count and growth trajectory?
├─ Organizational gaps or succession risks?
└─ Owner compensation vs. market rate?

Data Source Template:
├─ Owner compensation (paid): ______ [Confidence: ___]
├─ Market-rate equivalent: ______ [Confidence: ___]
├─ Key personnel risks: ______ [Confidence: ___]
```

#### Research Dimension 5: Financial Health & Working Capital
```
Questions:
├─ Cash position and burn rate (if applicable)?
├─ Days Sales Outstanding (DSO) for receivables?
├─ Days Inventory Outstanding (DIO)?
├─ Payment terms to suppliers (DPO)?
└─ Seasonal working capital swings?

Data Source Template:
├─ Cash position: ______ [Confidence: ___]
├─ Operating cycle (DSO/DIO/DPO): ______ [Confidence: ___]
└─ Working capital needs: ______ [Confidence: ___]
```

#### Research Dimension 6: Capital Expenditure & Asset Base
```
Questions:
├─ Annual CapEx requirement (as % of revenue)?
├─ Asset-light vs. asset-heavy business model?
├─ Depreciation & amortization profile?
├─ Any significant capital investments planned?
└─ Real estate ownership or leasing?

Data Source Template:
├─ CapEx % of revenue: ______ [Confidence: ___]
├─ D&A run-rate: ______ [Confidence: ___]
└─ Asset-light indicator: ______ [Confidence: ___]
```

#### Research Dimension 7: Debt & Capital Structure
```
Questions:
├─ Existing debt: amount, terms, covenants?
├─ Shareholder loans or related-party financing?
├─ Preferred stock or complex cap structure?
├─ Debt capacity based on cash flow?
└─ Covenant compliance status?

Data Source Template:
├─ Existing debt: ______ [Confidence: ___]
├─ Debt-to-EBITDA ratio: ______ [Confidence: ___]
├─ Available debt capacity: ______ [Confidence: ___]
```

#### Research Dimension 8: Growth Drivers & Future Outlook
```
Questions:
├─ What is management's revenue projection (next 3-5 years)?
├─ New products, market expansions, or M&A planned?
├─ Regulatory or macro tailwinds/headwinds?
├─ Exit timeline and exit assumptions?
└─ Key value creation levers for investor thesis?

Data Source Template:
├─ Management revenue projection: ______ [Confidence: ___]
├─ Growth drivers: ______ [Confidence: ___]
├─ Exit timeline & assumptions: ______ [Confidence: ___]
```

---

### 4.3 Private Company Revenue Split Strategy

**Rule**: Even if data is unavailable, NEVER eliminate the split row structure. Keep row, tag with confidence level, estimate from benchmark.

| **Data Availability** | **Approach** | **Confidence Tag** |
|---|---|---|
| User provides complete breakdown | Build P&L rows from actual dimensions; each row sourced as [HIGH] or [MED-HIGH] | [HIGH] |
| User provides partial (e.g., only total revenue) | Known dimensions precise, unknown dimensions use industry benchmark | [MED] for known, [LOW] for estimated |
| No financial data at all | Infer entire split from comparable company mix or industry report | [LOW] or [VERY-LOW] on all rows |

#### Template: Private Company Revenue Split
```
FY2024 Revenue Breakdown:

1. Direct Sales (Direct-to-Customer)
   └─ Amount: $____M [Confidence: ___]
   └─ % of Total: ___%
   └─ Data Source: ____________

2. Distribution / Wholesale Channel
   └─ Amount: $____M [Confidence: ___]
   └─ % of Total: ___%
   └─ Data Source: ____________

3. Online / E-commerce Channel
   └─ Amount: $____M [Confidence: ___]
   └─ % of Total: ___%
   └─ Data Source: ____________

4. Licensing / Royalties / Other
   └─ Amount: $____M [Confidence: ___]
   └─ % of Total: ___%
   └─ Data Source: ____________

TOTAL REVENUE: $____M [Blended Confidence: ___]

Example with data scarcity:
1. Direct Sales: $8M [HIGH] (User provided from accounting records)
2. Distribution: $4M [LOW] (Estimated as 33% of revenue based on industry median)
3. Online: $2M [VERY-LOW] (Placeholder, requires DD confirmation)
4. Licensing: $1M [MED] (User estimate from partnership agreements)
TOTAL: $15M [Blended: MED]
```

#### Template: Private Company Cost Split Strategy
```
FY2024 Cost of Revenue:

1. Direct Materials / COGS
   └─ Amount: $____M [Confidence: ___]
   └─ % of Revenue: ___%
   └─ Data Source: ____________

2. Direct Labor / Manufacturing
   └─ Amount: $____M [Confidence: ___]
   └─ % of Revenue: ___%
   └─ Data Source: ____________

3. Manufacturing Overhead / Facility Costs
   └─ Amount: $____M [Confidence: ___]
   └─ % of Revenue: ___%
   └─ Data Source: ____________

4. Logistics / Distribution Costs (if applicable)
   └─ Amount: $____M [Confidence: ___]
   └─ % of Revenue: ___%
   └─ Data Source: ____________

GROSS PROFIT: $____M
Gross Margin: ___% [Confidence: ___]
```

---

## 5. NEW — Adjusted EBITDA / Normalization (Singer Standard for Private Companies)

### 5.1 Overview: Why Normalization Matters

Private companies often have items that materially distort financial statements:
- Founder paying themselves above/below market rates
- Related-party transactions at non-arm's length prices
- One-time events (litigation, restructuring, M&A integration)
- Stock-based compensation often understated
- Seasonal or cyclical effects

**Objective**: Bridge from **Reported EBITDA** → **Adjusted EBITDA** to create a normalized earnings base for valuation.

---

### 5.2 Normalization Adjustment Framework

#### Adjustment Type 1: Owner Compensation Above Market
```
Scenario: Founder pays themselves $5M salary; market rate for similar CEO role is $500K

Adjustment:
├─ Reported owner compensation: $5.0M
├─ Market-rate compensation: $0.5M
├─ Excess compensation: $4.5M
└─ Treatment: **ADD BACK** $4.5M to Reported EBITDA

Justification: Excess comp is discretionary; normalized business can operate at market rate.
Confidence: [MED] (Based on salary database lookup; CEO-specific market rates may vary by region/industry)

Formula: Adjusted EBITDA = Reported EBITDA + $4.5M
```

#### Adjustment Type 2: Owner Compensation Below Market
```
Scenario: Founder is underpaid (working for free or minimal salary) to preserve cash

Adjustment:
├─ Actual owner compensation: $100K
├─ Market-rate compensation: $1.5M
├─ Undercompensation: $1.4M
└─ Treatment: **SUBTRACT** $1.4M from Reported EBITDA

Justification: Normalized business must account for sustainable labor costs.
Confidence: [MED-HIGH] (Market rate from salary comp databases; founder might work more hours)

Formula: Adjusted EBITDA = Reported EBITDA - $1.4M
```

#### Adjustment Type 3: Related-Party Transactions
```
Example 1 - Founder's Real Estate (Above Market Rent):
├─ Reported rent expense: $500K annually
├─ Market-rate rent for similar space: $250K annually
├─ Excess rent: $250K
└─ Treatment: **ADD BACK** $250K
└─ Rationale: Founder using business to overpay for personal real estate

Example 2 - Services from Affiliate (Below Market Rate):
├─ Affiliate IT services (reported): $50K annually
├─ Market-rate IT services: $200K annually
├─ Underpriced services: $150K
└─ Treatment: **SUBTRACT** $150K
└─ Rationale: Company benefiting from subsidized services; normalized cost is market rate

Confidence: [MED-HIGH] on market rates; [HIGH] on actual reported amounts
```

#### Adjustment Type 4: One-Time Expenses
```
One-time items to add back (if truly non-recurring):
├─ Litigation settlement: $500K (vs. typical $0K)
├─ M&A / integration costs: $200K (one-time; not ongoing)
├─ Restructuring charges: $300K (facility closure in past year)
├─ Acquisition earnout payment: $100K (non-recurring)

Confidence Approach:
├─ High confidence if documented as non-recurring
├─ Medium confidence if "appeared once" in historical record
├─ Low confidence if unclear whether truly one-time

Rule: Add back only if VERY likely to not recur. When in doubt, don't add back.
Formula: Adjusted EBITDA = Reported EBITDA + $1.1M (sum of one-time items)
```

#### Adjustment Type 5: Pro Forma Adjustments (Full-Run-Rate Impact)
```
Scenario: Company acquired a division mid-year (July 2024) that contributed $3M revenue, $500K EBITDA

Adjustment:
├─ Acquisition date: July 1, 2024
├─ Contribution FY2024: 6 months = $250K EBITDA
├─ Full-year equivalent: 12 months = $500K EBITDA
├─ Pro forma adjustment: +$250K
└─ Treatment: **ADD** $250K to Reported EBITDA

Rationale: For valuation purposes, assume division operated full year at acquired run-rate.
Confidence: [MED-HIGH] (Depends on validation that run-rate is sustainable)

Formula: Adjusted EBITDA = Reported EBITDA + $250K
```

#### Adjustment Type 6: Stock-Based Compensation (SBC)
```
Scenario: Private company awarded options/restricted stock to key employees
├─ Reported SBC expense: $100K (conservative; often below market)
├─ Estimated fair-value SBC: $300K (based on peer company SBC benchmarks)
├─ Incremental SBC adjustment: $200K
└─ Treatment: **SUBTRACT** $200K from Reported EBITDA

Rationale: If company goes public or is acquired, SBC will be actual cash equivalent.
Confidence: [MED] (Market rate SBC varies; depends on stage, growth rate, risk profile)

Formula: Adjusted EBITDA = Reported EBITDA - $200K
```

---

### 5.3 Adjusted EBITDA Bridge Template

```
═════════════════════════════════════════════════════════════════
Adjusted EBITDA Bridge: [Company Name] FY2024
═════════════════════════════════════════════════════════════════

Starting Point: Reported Net Income
├─ Net Income (FY2024): $2.0M [HIGH]
├─ Add: Interest Expense: $0.5M [HIGH]
├─ Add: Tax Expense: $0.8M [HIGH]
├─ Add: Depreciation & Amortization: $1.2M [HIGH]
├─ EBITDA (Reported): $4.5M [HIGH]

Normalization Adjustments:
├─ (1) Owner compensation above market: +$3.5M [MED]
├─ (2) Related-party rent adjustment: +$0.2M [MED-HIGH]
├─ (3) One-time litigation settlement: +$0.5M [MED]
├─ (4) Pro forma for mid-year acquisition: +$0.3M [MED-HIGH]
├─ (5) SBC fair-value adjustment: -$0.2M [MED]
├─ (6) Other adjustments: $0.0M

─────────────────────────────────────────────────────────────────
Adjusted EBITDA (FY2024): $8.8M [Blended Confidence: MED]
═════════════════════════════════════════════════════════════════

Adjusted EBITDA Margin: 8.8M / [Total Revenue] = ___% [MED]

Notes:
├─ Adjusted EBITDA is the basis for all subsequent valuation (multiples, DCF, debt capacity)
├─ Confidence reflects data source quality; bridge to be validated in formal due diligence
├─ High-confidence items: reported financials; low-confidence items: market-rate assumptions
└─ DD Gap List: Confirm sustainability of all add-back adjustments (especially owner comp normalization)
```

---

## 6. NEW — Liquidity Discount & Control Premium Framework

### 6.1 Discount for Lack of Marketability (DLOM)

**Definition**: The percentage reduction applied to a public company valuation (or DCF value) to account for the fact that private company shares cannot be freely traded.

**DLOM Range**: 15-35% (Based on Restricted Stock Studies, Equity Research Developing Losses / ERDL, Option Pricing Method)

#### DLOM Estimation Factors
```
Key Driver 1: Company Stage & Maturity
├─ Early-stage (Pre-revenue, Series A): 30-35% DLOM (High illiquidity)
├─ Growth-stage (Revenue growing, Series B/C): 20-30% DLOM
├─ Mature private (Stable revenue/EBITDA, Series D+): 15-20% DLOM
└─ Pre-IPO (12 months to likely exit): 10-15% DLOM

Key Driver 2: Information Quality & Transparency
├─ Audited financials + quarterly reporting: Lower DLOM (10-15%)
├─ Unaudited financials + annual reporting: Mid DLOM (15-25%)
├─ Limited financial disclosure: Higher DLOM (25-35%)
└─ No reliable financials: Maximum DLOM (30-35%)

Key Driver 3: Revenue Predictability & Stability
├─ Recurring/SaaS revenue, stable margin: Lower DLOM (12-18%)
├─ Project-based / cyclical revenue: Higher DLOM (20-28%)
├─ Highly volatile or unpredictable: Very High DLOM (25-35%)

Key Driver 4: Expected Time to Liquidity Event
├─ Expected exit in <2 years: 12-18% DLOM
├─ Expected exit in 3-5 years: 18-25% DLOM
├─ No clear exit timeline: 25-35% DLOM
```

#### DLOM Calculation & Application
```
Base Valuation (from DCF or comparable company multiples): $100M [Public-equivalent basis]

DLOM Adjustment:
├─ Stage: Growth-stage private
├─ Information quality: Unaudited financials
├─ Revenue predictability: Cyclical
├─ Exit timeline: 4-5 years
├─ Recommended DLOM: 22% (midpoint of 20-25% range)

Private Company Fair Value:
├─ $100M × (1 - 0.22) = $78M [Fair value, applying DLOM]

Sensitivity Analysis:
├─ DLOM at 15% → Valuation: $85M
├─ DLOM at 22% → Valuation: $78M (Base case)
├─ DLOM at 30% → Valuation: $70M
```

---

### 6.2 Control Premium (If Acquiring Majority / 100%)

**Definition**: The percentage premium an acquirer pays above the per-share value of minority shares to acquire control of the company.

**Control Premium Range**: 20-40% (Based on Mergerstat / Dealogic data)

#### Control Premium Estimation Factors
```
Key Driver 1: Quality of Existing Management
├─ Stable, experienced management team in place: Lower premium (15-20%)
├─ Key person dependency / management gaps: Higher premium (30-40%)
├─ Founder-dependent; successor planning unclear: Very high premium (35-45%)

Key Driver 2: Synergy Potential
├─ High strategic synergies (revenue synergies + cost cuts): Higher premium (35-45%)
├─ Moderate synergies (cost cuts mainly): Mid premium (25-35%)
├─ Limited synergies (financial buyer, portfolio co): Lower premium (15-25%)

Key Driver 3: Competitive Bidding Environment
├─ Single buyer (no competitive tension): Lower premium (15-25%)
├─ Multiple bidders / auction process: Higher premium (30-45%)

Key Driver 4: Current Minority Holder Expectations
├─ Minority holders have low expectations: Lower premium needed (15-25%)
├─ Minority holders expecting exit premium: Higher premium (30-40%)
```

#### Control Premium Calculation & Application
```
Minority Share Value (after applying DLOM): $78M (for 100 shares = $780K per share minority)

Control Premium Adjustment (Majority/Full Buyout):
├─ Existing management quality: Key person dependent (management gap)
├─ Synergy potential: Moderate (cost cuts, some revenue synergies)
├─ Bidding environment: Single buyer (no competitive auction)
├─ Recommended control premium: 28% (midpoint of 25-30% range)

Transaction Price (100% ownership):
├─ Base value per share (minority): $780K
├─ Control premium: +28%
├─ Acquirer pays per share: $780K × 1.28 = $998K
├─ Total deal value: $998K × 100 = $99.8M ≈ $100M

Alternative: Starting from DLOM-adjusted minority value, add control premium to get full buyout price.
```

---

### 6.3 Minority Discount (If Acquiring Minority Stake, e.g., Series D Investment)

**Definition**: The reduction in per-share value when acquiring a non-controlling minority stake (typically <20% ownership).

**Minority Discount Range**: 20-30%

#### Minority Discount Factors
```
Key Drivers:
├─ No board seat / voting power: Higher discount (25-30%)
├─ Board observer + information rights: Lower discount (15-20%)
├─ Protective provisions (vetoes on major decisions): Lower discount (12-18%)
├─ Tag-along / drag-along rights: Medium discount (18-25%)

Application:
├─ Base valuation (implied per-share value if majority): $100M
├─ Minority discount (20%): $100M × (1 - 0.20) = $80M
├─ Investment into minority stake: $80M valuation for minority investment position
```

---

### 6.4 Putting It All Together: Multi-Layer Valuation Bridge

```
═════════════════════════════════════════════════════════════════
Valuation Bridge: [Company Name]
═════════════════════════════════════════════════════════════════

Step 1: Base Valuation Method (DCF or Comparable Companies)
├─ DCF Enterprise Value (public-equivalent): $100M [HIGH confidence in method, pending assumptions]
├─ OR Comparable Company Multiple Approach: $100M [MED-HIGH confidence]
└─ Source basis: [Describe: market multiples / assumptions]

Step 2: Apply Lack of Marketability Discount (DLOM)
├─ Selected DLOM: 22% (growth-stage, unaudited financials, 4-5 year exit timeline)
├─ Adjusted value: $100M × (1 - 0.22) = $78M [Minority fair value]

Step 3: Apply Control Premium (if majority/full acquisition)
├─ Selected control premium: 28% (key person dependent, moderate synergies)
├─ Transaction price (100%): $78M / (1 - 0.28) = $108M [OR: $78M × 1.28 = $100M for full buyout]
│
│  *Note: Control premium calculation can be done as:
│   Method A) Add premium to minority: $78M × 1.28 = $100M
│   Method B) Reverse from majority: $78M / (1 - Control%) [if starting with minority discount]*

OR Step 3 Alternative: Apply Minority Discount (if minority investment)
├─ Selected minority discount: 20% (observer rights, info access, tag-along provision)
├─ Investment valuation: $100M × (1 - 0.20) = $80M [Minority investment price]

═════════════════════════════════════════════════════════════════
Final Valuation:
├─ Minority stake (passive): $78M - $80M [DLOM-adjusted, minority discount variant]
├─ Majority acquisition (50%+): $98M - $108M [Control premium applied]
├─ Full buyout (100%): $100M - $110M [Full control, post-normalization]

Sensitivity Analysis Required:
├─ DLOM low (15%) / Control high (35%) → $108M
├─ DLOM high (30%) / Control low (20%) → $90M
└─ Most likely case (22% DLOM / 28% control) → $100M
═════════════════════════════════════════════════════════════════
```

---

## 7. Layer 2: Partner (Private Company Version)

### 7.1 Core Questions (Adapted for PE/VC Context)

**Question S1: Strategic Thesis**
```
What is your investment thesis for [Company]?
├─ [A] Operational improvement / margin expansion
├─ [B] Revenue growth acceleration (organic or add-on acquisitions)
├─ [C] Platform / roll-up consolidation
├─ [D] Financial engineering / leverage capture
├─ [E] Market/segment disruption or new category
└─ [F] Multiple arbitrage (buy private at X, sell public at higher multiple)
```

**Question S2: Ownership Structure (Replaces Control %)**
```
What is the target ownership structure post-investment?
├─ [A] Minority stake (<50%), passive or observership
├─ [B] Majority stake (50-99%), operational control
├─ [C] Full buyout (100%), clean acquisition
└─ [D] Staged acquisition (Series financing + future buyout path)
```

**Question S3: Investment Size & Fund Capacity**
```
What is the target investment size for this deal?
├─ [A] <$50M (Growth equity, lower mid-market buyout)
├─ [B] $50M-$200M (Mid-market buyout)
├─ [C] $200M-$500M (Large/upper mid-market)
├─ [D] >$500M (Mega-deal, platform acquisition)
└─ Rationale: Determines leverage capacity, hold period, operational intensity
```

**Question S4: Time Horizon & Exit Assumptions**
```
What is your hold period target?
├─ [A] 3-5 years (typical PE hold)
├─ [B] 5-7 years (longer hold, transformation play)
├─ [C] 7-10+ years (platform / buy-and-build strategy)
└─ Exit assumptions: [Secondary sale / IPO / strategic buyer / dividend recap]
```

---

### 7.2 DD Gap List (Critical Private-Company Addition)

**Mandatory Follow-Up Question**:
```
What critical data is still missing that we need from formal due diligence?

For each gap, rank by materiality:
├─ 🔴 CRITICAL (Deal-breaker if not resolved positively)
├─ 🟠 IMPORTANT (Materially impacts valuation / thesis)
├─ 🟡 NICE-TO-HAVE (Useful for modeling but not deal-defining)

Examples of typical gaps:
├─ [CRITICAL] Audited financial statements (if not provided)
├─ [CRITICAL] Customer contract terms and renewal rates
├─ [CRITICAL] Key person employment agreements and non-competes
├─ [IMPORTANT] Management team depth / succession plan
├─ [IMPORTANT] Related-party transaction true economics
├─ [IMPORTANT] Environmental / regulatory compliance status
├─ [NICE-TO-HAVE] Detailed product roadmap / R&D pipeline
├─ [NICE-TO-HAVE] Market size and TAM validation

This DD Gap List feeds into quality gate checklist (Section 10).
```

---

## 8. Layer 1 Output Structure (Private Company Version)

### 8.1 Research Memo Template

```
═════════════════════════════════════════════════════════════════
RESEARCH MEMO: [Company Name]
Private Company Investor Model
Date: [Date]
═════════════════════════════════════════════════════════════════

⚠️  [CREDIBILITY WARNING - if applicable]
[If data availability is [C] or [D], display warning that analysis is preliminary]

═════════════════════════════════════════════════════════════════
I. EXECUTIVE SUMMARY
═════════════════════════════════════════════════════════════════
├─ Company name: [Name]
├─ Founded: [Year]
├─ Headquarters: [Location]
├─ Industry / Segment: [Description]
├─ Key business model: [Summary]
└─ Investment thesis (short): [1-2 sentences]

Key Financials Summary (with confidence levels):
├─ FY2024 Revenue: $[X]M [Confidence: ___]
├─ FY2024 Adjusted EBITDA: $[Y]M [Confidence: ___]
├─ Adjusted EBITDA Margin: ___% [Confidence: ___]
├─ Revenue CAGR (3-year): ___% [Confidence: ___]
└─ Estimated Enterprise Value (base case): $[Z]M [Confidence: ___]

═════════════════════════════════════════════════════════════════
II. BUSINESS OVERVIEW (Research Dimensions 1-8)
═════════════════════════════════════════════════════════════════
[Include detailed write-up of each research dimension with confidence tags]

═════════════════════════════════════════════════════════════════
III. ADJUSTED EBITDA BRIDGE
═════════════════════════════════════════════════════════════════
[Insert detailed bridge from Reported EBITDA to Adjusted EBITDA]
[Each adjustment tagged with confidence and data source]

═════════════════════════════════════════════════════════════════
IV. VALUATION SUMMARY
═════════════════════════════════════════════════════════════════
[Insert valuation bridge with DLOM, control premium]
[Sensitivity analysis on key drivers]

═════════════════════════════════════════════════════════════════
V. DUE DILIGENCE GAP LIST
═════════════════════════════════════════════════════════════════
[Critical items needed to validate thesis]

═════════════════════════════════════════════════════════════════
VI. CONFIDENCE DASHBOARD
═════════════════════════════════════════════════════════════════
[Summary table of all key assumptions and their confidence levels]
```

---

## 9. Layers 3-4: Financial Model & Valuation (Same Framework as Public, with Private Adaptations)

### 9.1 Key Adaptations

#### Revenue Modeling
```
Principle: All revenue line items preserved even with missing data (red tagged)

├─ If user provides detailed breakdown → Each line [HIGH] or [MED-HIGH]
├─ If user provides total only → Split by industry benchmarks [LOW] / [VERY-LOW]
├─ If no data → Infer from comparable company mix [LOW] / [VERY-LOW]

Pro-forma Growth Assumptions:
├─ Years 1-2: Historical CAGR (if available) or management guidance [MED-HIGH]
├─ Years 3-5: Industry growth rate or normalization to mature rate [MED]
├─ Terminal: Long-term GDP growth + market growth premium [MED] / [LOW]

Confidence adjustment: Historical actuals > Management projections > Industry benchmarks > Guesses
```

#### Cost Modeling
```
COGS Assumptions:
├─ If audited: Use reported / normalized [HIGH]
├─ If unaudited: Use reported ± margin of error [MED-HIGH]
├─ If not available: Use industry benchmark [LOW] / [VERY-LOW]

OpEx Assumptions:
├─ Fixed vs. variable cost split (affects operating leverage)
├─ Sales & marketing efficiency (Revenue / S&M spend trend)
├─ G&A scalability (G&A as % of revenue decreasing with scale)
├─ R&D intensity (% of revenue, particularly for tech companies)

Confidence Tags:
├─ Reported categories: [HIGH] / [MED-HIGH]
├─ Benchmarked categories: [MED] / [LOW]
└─ Fully estimated categories: [LOW] / [VERY-LOW]
```

#### Tax Rate Consideration (China-Specific)
```
Standard Corporate Tax Rate: 25%
High-Tech Enterprise (HTE) Status: 15% [if applicable]

Private company considerations:
├─ Does company have HTE certification? → 15% rate [HIGH confidence if yes]
├─ Is HTE status at risk? → Model dual scenarios (25% / 15%)
├─ Tax planning opportunities? → Model after-tax impact [MED confidence]

Formula:
├─ Normalized tax rate: [25% standard] or [15% if HTE eligible] [MED] confidence
├─ Sensitivity: Run scenarios at 15% / 20% / 25% tax rates
```

#### Working Capital Modeling
```
Private companies often have material working capital swings:

├─ Days Sales Outstanding (DSO): Receivables balance / (Revenue / 365)
├─ Days Inventory Outstanding (DIO): Inventory balance / (COGS / 365)
├─ Days Payables Outstanding (DPO): Accounts payable / (COGS / 365)
├─ Operating cycle: DSO + DIO - DPO

Model assumptions:
├─ Changes in working capital as driver of cash flow
├─ Seasonal patterns (often pronounced in private companies)
├─ Cash conversion cycle improvement as value creation lever

Confidence:
├─ Actual reported balances: [HIGH]
├─ Estimated trend from available data: [MED-HIGH] / [MED]
├─ Industry benchmark: [LOW] / [VERY-LOW]
```

#### Capital Expenditure
```
Private companies often under-invest in CapEx (to preserve cash) or over-invest (founder whim)

CapEx Modeling:
├─ Maintenance CapEx: [% of revenue or $ amount based on historical]
├─ Growth CapEx: [Additional for stated expansion plans]
├─ Total CapEx: [Maintenance + Growth]

Confidence:
├─ Historical actual CapEx: [HIGH] (if available)
├─ Management plan for growth CapEx: [MED] / [MED-HIGH]
├─ Maintenance CapEx inferred from industry: [MED] / [LOW]

Adjustment: Ensure CapEx is sufficient for stated growth; flag if unrealistic.
```

---

### 9.2 Normalization Adjustments in Model

All normalization adjustments from the EBITDA bridge (Section 5) flow through:
- Owner compensation normalization
- Related-party transaction adjustments
- One-time expense elimination
- SBC fair-value normalization
- Pro forma full-year adjustments

These are modeled as separate line items in the P&L to make them transparent and adjustable for scenario analysis.

---

## 10. Private-Specific Quality Gate Additions

### Final Checklist Before Outputting Model

```
═════════════════════════════════════════════════════════════════
QUALITY GATE CHECKLIST: Private Company Modeling
═════════════════════════════════════════════════════════════════

Research Phase:
☐ Mandatory First Question answered (data availability confirmed)
☐ All 8 research dimensions addressed with data sources
☐ Credibility warning displayed (if data limited)
☐ Revenue and cost split strategies applied (even where data sparse)
☐ DD Gap List generated with materiality ranking

EBITDA Normalization Phase:
☐ Adjusted EBITDA bridge completed with all material adjustments
☐ Each adjustment sourced and justified (confidence level documented)
☐ Bridge reconciles: Reported EBITDA → Adjusted EBITDA
☐ Adjusted EBITDA is the basis for all downstream valuation work
☐ Normalization adjustments flagged for validation in formal DD

Confidence Level Tagging:
☐ EVERY revenue line item has confidence tag [HIGH] / [MED-HIGH] / [MED] / [LOW] / [VERY-LOW]
☐ EVERY cost assumption has confidence tag
☐ EBITDA bridge adjustments have confidence levels
☐ Terminal value and discount rate have confidence levels
☐ "Very Low" confidence items identified and flagged

Liquidity & Control Framework:
☐ DLOM explicitly stated and justified (or noted as N/A with reason)
☐ Control premium stated (if applicable to investment thesis)
☐ Minority discount stated (if applicable)
☐ Valuation bridge shows pre-DLOM and post-DLOM values
☐ Sensitivity analysis on DLOM / control premium completed

Financial Model:
☐ 3-statement model complete (P&L, Balance Sheet, Cash Flow)
☐ Working capital treated explicitly (DSO, DIO, DPO)
☐ CapEx modeled with supporting assumptions
☐ Debt schedule (if applicable) with covenant testing
☐ All assumptions documented in separate assumptions sheet

Valuation Methods:
☐ DCF valuation completed with terminal value approach
☐ Comparable company analysis (if sufficient data) with control adjustments
☐ Both methods reconciled or disparities explained
☐ Exit scenario analysis (strategic / financial buyer / IPO)

Sensitivity Analysis:
☐ Sensitivity tables on: Revenue growth / EBITDA margin / WACC / Terminal growth rate
☐ Scenario analysis (base / upside / downside cases)
☐ Focus on "Very Low" confidence variables with independent sensitivity

Due Diligence Gap List:
☐ Critical items ranked by materiality (🔴 / 🟠 / 🟡)
☐ DD gaps mapped to model impact (which assumptions depend on resolution)
☐ Timeline for DD closure identified
☐ Contingency plan for negative findings documented

Final Output:
☐ Research memo includes all 8 dimensions
☐ Adjusted EBITDA bridge visible and explained
☐ Valuation bridge with DLOM/control premium shown
☐ Sensitivity analysis provided
☐ DD Gap List prioritized and actionable
☐ Clear statement: "This is a preliminary model pending due diligence closure"

════════════════════════════════════════════════════════════════
If ANY checkbox is unchecked, return model to researcher for completion.
Do not release to Partner layer without full completion.
════════════════════════════════════════════════════════════════
```

---

## 11. Execution Flow & Trigger Routing

### 11.1 User Trigger → Layer Routing

```
User input: "综合建模[非上市公司] [Company Name]"

Step 1: Confirm Trigger
├─ ✓ Trigger matches [integrated-modeling-private]
├─ Route to Layer 1: Researcher
└─ Skip public company model; use private adaptations

Step 2: Layer 1 Execution
├─ Ask Mandatory First Question (data availability)
├─ Display credibility warning (if applicable)
├─ Execute 8-dimensional research with confidence tags
├─ Build EBITDA normalization bridge
├─ Generate DD Gap List
└─ Output Research Memo

Step 3: Layer 2 Execution
├─ Ask strategic thesis & ownership structure questions
├─ Clarify investment size, hold period, exit assumptions
├─ Route to Layer 3/4 based on decision
└─ Surface DD gaps that block model completion

Step 4: Layer 3-4 Execution
├─ Build normalized financial model (using Adjusted EBITDA)
├─ Model all adaptations: revenue splits, cost categories, tax rate, working capital, CapEx
├─ Calculate DLOM and control premium adjustments
├─ Run DCF & comparable company valuation
├─ Sensitivity analysis on all [LOW] / [VERY-LOW] assumptions

Step 5: Quality Gate
├─ Execute full checklist (Section 10)
├─ Flag any missing elements
├─ Return to researcher if gaps exist
└─ Release comprehensive model if all gates passed

Step 6: Output
├─ Complete model workbook with all 4 layers visible
├─ Emphasis on confidence levels and assumptions transparency
├─ Clear labeling: "PRELIMINARY MODEL — Due Diligence Validation Required"
└─ DD Gap List prioritized and actionable
```

---

## 12. Abbreviations & Glossary

```
DLOM            Discount for Lack of Marketability (15-35%)
SBC             Stock-Based Compensation
EBITDA          Earnings Before Interest, Tax, Depreciation, Amortization
Adjusted EBITDA EBITDA after normalization adjustments
DSO             Days Sales Outstanding (receivables metric)
DIO             Days Inventory Outstanding
DPO             Days Payables Outstanding
CapEx           Capital Expenditure
HTE             High-Tech Enterprise (China tax status, 15% rate)
WACC            Weighted Average Cost of Capital
DCF             Discounted Cash Flow
PE              Private Equity
VC              Venture Capital
LBO             Leveraged Buyout
NWC             Net Working Capital
FY              Fiscal Year
CAGR            Compound Annual Growth Rate
EV              Enterprise Value
TAM             Total Addressable Market
```

---

## 13. Additional Resources & External References

```
Normalization & EBITDA Guidance:
├─ Mergerstat / FactSet M&A reports (control premiums)
├─ Restricted Stock Studies (DLOM ranges)
├─ Option Pricing Models (DLOM estimation)

Comparable Company Data:
├─ PitchBook (private company benchmarks)
├─ Prequin (private equity multiples)
├─ S&P Capital IQ (public company multiples)

Valuation Standards:
├─ AICPA Valuation Guidance (DLOM frameworks)
├─ ASA (American Society of Appraisers) standards
├─ NACVA (National Association of Certified Valuators) publications

China-Specific:
├─ KPMG China M&A reports (control premiums in Chinese deals)
├─ Deloitte China valuations guidance
├─ CAC (China Association of Appraisers) standards
```

---

**Version**: 1.0
**Last Updated**: March 2026
**Author**: Integrated Modeling System - Private Company Module
**Status**: Production Ready
