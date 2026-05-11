# Reference 02 · Credit Analysis: Metrics, Covenants, and Capital Structure

> "The credit agreement is the constitution of the company. It tells you exactly what the company can and can't do, what happens when it fails, and who has the power to do something about it."
> — Distressed investing principle

---

## Core Credit Metrics

### Yield Metrics

**Yield to Worst (YTW)**
The lowest possible yield an investor can receive on a bond without the issuer defaulting. YTW accounts for all call provisions, put provisions, and sinking fund features — it is the most conservative yield measure and the standard for comparing bonds.

*Formula*: Calculate yield assuming the issuer calls or redeems the bond at the earliest possible date, then take the lowest resulting yield across all scenarios.

*Interpretation table*:
| YTW | Signal |
|-----|--------|
| < 6% | Investment grade or near-IG; minimal distress risk |
| 6–9% | High yield, elevated risk; monitor covenant headroom |
| 9–12% | Stressed credit; covenant breach risk within 12 months |
| 12–18% | Distressed; market pricing significant default probability |
| >18% | Deeply distressed; market implying high default probability; recovery analysis dominates |

**Yield to Maturity (YTM)**
The total return assuming the bond is held to maturity and all coupon and principal payments are made. YTM > YTW when the bond is trading at a premium to the call price (unlikely in distressed). In distressed analysis, focus on YTW not YTM.

**Spread to Treasuries / Spread to LIBOR (OAS)**
The excess yield over the risk-free rate. Option-adjusted spread (OAS) removes the value of embedded options. Spread is the credit premium for taking on default risk.

*Benchmark spreads*:
- IG corporate bonds: 50–200 bps over Treasuries
- BB-rated HY: 200–400 bps
- B-rated HY: 400–600 bps
- CCC-rated / stressed: 600–1,000+ bps
- Distressed (>10% YTW): effectively priced on dollar price and recovery, not spread

---

## Leverage Metrics

### Debt / EBITDA

The most common leverage metric in leveraged finance. Measures how many years of operating cash flow are needed to repay total debt.

*Calculation*: Total Debt (or Net Debt) ÷ LTM EBITDA

*Benchmark by credit quality*:
| Debt/EBITDA | Credit Signal |
|-------------|--------------|
| < 2.0x | Conservative; investment grade territory |
| 2.0–3.5x | Moderate leverage; standard for many IG issuers |
| 3.5–5.0x | Leveraged; typical LBO range at issuance |
| 5.0–6.5x | Elevated; stress starts here, especially in cyclical downturns |
| 6.5–8.0x | High stress; covenant headroom eroding |
| > 8.0x | Distressed zone; refinancing extremely difficult |

**Adjustments that matter**:
- Use **LTM adjusted EBITDA** — company-adjusted EBITDA often adds back items that are genuinely recurring costs (restructuring, stock-based comp, one-time items that happen every year). Apply a haircut to management EBITDA.
- Use **gross debt** for bankruptcy analysis (no credit for cash that could be consumed), **net debt** for going-concern analysis
- **EBITDA-CapEx** (sometimes called "maintenance CapEx EBITDA") is more conservative and appropriate for capital-intensive industries

### Net Leverage vs. Gross Leverage

- **Gross leverage**: Total debt ÷ EBITDA (use in covenant calculations, bankruptcy analysis)
- **Net leverage**: (Total debt − cash) ÷ EBITDA (use in going-concern refinancing analysis)

For companies approaching distress, cash burn is often rapid — do not over-credit the cash balance. Model cash runway: at the current burn rate, how many months of liquidity does the company have?

---

## Coverage Metrics

### Interest Coverage Ratio (ICR)

*Formula*: EBITDA ÷ Interest Expense (cash interest only, not PIK)

*Benchmarks*:
| ICR | Signal |
|-----|--------|
| > 4.0x | Comfortable; ample headroom |
| 2.5–4.0x | Adequate; watch for EBITDA deterioration |
| 1.5–2.5x | Stressed; covenant breach possible if EBITDA falls 20%+ |
| 1.0–1.5x | Critical; company barely covering interest; refinancing will be difficult |
| < 1.0x | Burning cash to pay interest; default or PIK toggle likely |

### Fixed Charge Coverage Ratio (FCCR)

More comprehensive than ICR — includes all fixed charges: interest, required amortization, capital leases, and sometimes capex and preferred dividends.

*Formula*: (EBITDA − CapEx) ÷ (Interest + Required Amortization + Capital Lease Payments)

FCCR is the metric most commonly used in **maintenance covenants** in credit agreements. When FCCR covenant = 1.1x, the company must maintain FCCR above 1.1x at all times.

### Free Cash Flow (FCF) Metrics

**FCF Conversion** = FCF ÷ EBITDA. High-quality businesses convert 50–80% of EBITDA to FCF. Companies with high CapEx, working capital needs, or interest burdens convert much less.

**Cash Interest Coverage** = FCF (before debt service) ÷ Cash Interest Expense. More conservative than ICR because it uses actual FCF rather than EBITDA.

**Debt Paydown Capacity** = FCF after interest ÷ Total Debt. How many years at current FCF generation would it take to repay all debt? <10 years = manageable; >20 years = structural problem.

---

## Maturity Wall Analysis

### Building the Maturity Profile

Create a table showing debt maturities by year:

| Year | Tranche | Amount ($M) | Rate | Notes |
|------|---------|------------|------|-------|
| 2025 | Revolver | 250 | SOFR+300 | Drawn 60% |
| 2026 | TLB | 800 | SOFR+400 | Amortizing 1%/yr |
| 2027 | Second lien | 400 | 8.5% fixed | No amortization |
| 2028 | Senior unsecured notes | 600 | 9.0% fixed | Bullet maturity |

### Maturity Wall Risk Assessment

A "maturity wall" exists when:
- Multiple large tranches mature within the same 12–18 month window
- The company's liquidity (revolver capacity + FCF) is insufficient to repay maturities without refinancing
- Refinancing conditions in the relevant market (leveraged loan, HY bond) are uncertain or expensive

**The refinancing cliff**: If the TLB matures in 2026, the company must refinance it in 2025 or earlier (lenders will reclassify it as current if unresolved). If the company cannot refinance, it faces a liquidity crisis even if operations are fine.

**Springing maturity**: Watch for provisions in revolvers where the revolver maturity date "springs" to 91 days before the maturity of any large tranche. This can accelerate a liquidity crisis dramatically.

> "A company can have excellent operations and still go bankrupt because of a maturity wall. The Wall Street Journal calls it a 'cash flow problem.' It's actually a capital structure timing problem — a very different thing."

### Refinancing Risk Model

For each tranche approaching maturity (within 24 months):
1. What market would need to accept this refinancing? (leveraged loan / HY bond / private credit)
2. At what spread/yield would this company refinance today?
3. How much does annual interest expense increase at refinancing rates vs. current rate?
4. Does the increased interest expense push coverage ratios below covenant thresholds?

---

## Covenant Analysis

### Covenant Types: Maintenance vs. Incurrence

**Maintenance Covenants** (also called "financial maintenance covenants")
- **Tested at regular intervals** (typically quarterly)
- **Proactive**: the company must maintain the ratio above/below the threshold regardless of what it is doing
- **Example**: "Borrower shall maintain Total Net Leverage below 6.5x as of the last day of each fiscal quarter"
- **Found in**: Most bank credit agreements (revolvers, traditional term loans), European leveraged loans
- **Implication**: A maintenance covenant gives lenders an early warning system and the ability to accelerate or renegotiate before default

**Incurrence Covenants** (also called "high-yield style covenants")
- **Tested only when the company takes a specific action** (incurring new debt, making a restricted payment, etc.)
- **Reactive**: the company has no obligation unless it is doing something
- **Example**: "Borrower may incur additional debt only if, on a pro forma basis, Total Leverage would not exceed 5.5x"
- **Found in**: High-yield bonds, "covenant-lite" leveraged loans (increasingly common since 2012)
- **Implication**: Incurrence covenants provide much weaker protection for lenders; they do not trigger events of default for operational underperformance alone

**Why this matters for distressed investors**:
In a "cov-lite" capital structure (incurrence covenants only), the company can deteriorate significantly without triggering a covenant default. This is good for equity (more time to recover) but potentially bad for senior lenders (they cannot accelerate and protect themselves). It creates a "slow default" problem where the company bleeds for years before restructuring.

Conversely, a maintenance covenant in a bank revolver creates an early catalyst: the moment leverage exceeds the threshold, the bank can demand repayment or renegotiate. This often forces a restructuring earlier — before value has been consumed — which is better for recovery.

### Covenant Headroom Calculation

**Step 1**: Identify the covenant threshold and metric tested.
- Example: Total Net Leverage covenant of 6.0x, tested quarterly

**Step 2**: Calculate current covenant metric.
- Current Total Net Debt: $1,800M
- Current LTM EBITDA: $300M
- Current Total Net Leverage: 6.0x (exactly at threshold — zero headroom)

**Step 3**: Model forward trajectory.
- Projected EBITDA next quarter: $280M (declining)
- Projected Total Net Debt: $1,820M (revolver draws continuing)
- Projected Leverage: $1,820 ÷ $280 = 6.5x → **Covenant breach next quarter**

**Step 4**: Assess likelihood of waiver or amendment.
- Does the company have a relationship bank that can grant a waiver?
- What is the cost of the waiver (fees, pricing grid step-up, tighter terms)?
- Will lenders demand an amend-and-extend vs. refusing to waive?

**Covenant Headroom Formula**:
Headroom = (Covenant Threshold × Current EBITDA) − Current Total Debt

Example: If the covenant is Gross Leverage ≤ 6.5x and current EBITDA is $300M:
- Maximum allowable debt = 6.5x × $300M = $1,950M
- Current debt = $1,800M
- Headroom = $1,950M − $1,800M = $150M (company can absorb $150M more debt before breach)

Then stress-test: if EBITDA falls 15% to $255M:
- Maximum allowable debt = 6.5x × $255M = $1,657M
- But current debt is $1,800M → breach with -15% EBITDA

---

## Credit Agreement Structure

### The Typical Leveraged Loan Capital Structure

```
Priority        Tranche          Typical Rate          Notes
─────────────────────────────────────────────────────────────
1st Priority    Revolving Credit SOFR + 300-400 bps    Undrawn = unused fee
1st Priority    Term Loan A (TLA) SOFR + 300-350 bps   Fully amortizing
1st Priority    Term Loan B (TLB) SOFR + 400-500 bps   1% annual amortization, bullet
2nd Priority    Second Lien TL    SOFR + 600-800 bps   PIK option sometimes
3rd Priority    Senior Unsecured  HY rates (7-11%)     Public bonds usually
4th Priority    Subordinated      HY rates (9-13%)     Rare in modern structures
Last            Equity            Residual claim        No contractual return
```

### Key Credit Agreement Provisions

**Cross-default provisions**: A default under one debt instrument triggers a default under others. This is what makes maturity walls lethal: if the company cannot refinance its HY bonds, the cross-default clause triggers an event of default in the revolver, which accelerates all debt.

**Restricted payments basket**: Limits the company's ability to pay dividends, make acquisitions, or repurchase equity. Protects lenders from value extraction by the sponsor/management.

**Portability**: In some HY bonds, the indenture allows the bond to be assumed by an acquirer without triggering a change of control premium. Makes the company more easily sold (reduces drag on M&A).

**EBITDA add-backs and "Adjusted EBITDA"**: The credit agreement defines exactly what EBITDA means for covenant calculation purposes. Sponsors negotiate aggressively to include add-backs for one-time items, run-rate savings, and projected synergies. Read the definition carefully — "Adjusted EBITDA" in a credit agreement can be dramatically different from GAAP EBITDA.

**Builder/ratio basket**: Allows incremental debt issuance or restricted payments up to a defined amount (the "builder basket") plus any amounts that can be incurred without breaching the leverage ratio covenant (the "ratio basket"). Understanding the basket capacity tells you how much additional debt the company can legally incur.

---

## Putting It Together: Credit Scorecard

| Metric | Current | Threshold | Headroom | Trend | Score |
|--------|---------|-----------|----------|-------|-------|
| Debt/EBITDA | 6.2x | 7.0x (covenant) | $240M | Deteriorating | ⚠️ |
| Interest Coverage | 1.8x | 1.5x (covenant) | $45M EBITDA | Stable | ✅ |
| FCCR | 1.1x | 1.0x (covenant) | Minimal | Deteriorating | 🔴 |
| Nearest Maturity | 18 months | — | Must refinance | — | ⚠️ |
| YTW (bonds) | 14.5% | — | Distressed territory | — | 🔴 |
| Cash Runway | 9 months | — | <12 months = urgent | — | 🔴 |

**Reading this scorecard**: Three red flags out of six metrics, with FCCR barely above covenant at 1.1x and deteriorating. Distressed-zone YTW. Refinancing required within 18 months. This is a deep-dive situation — proceed to fulcrum analysis.

> "The credit scorecard is not a formula. It's a map. The map tells you where the pressure points are. Then you go figure out what's causing them and whether it's fixable."
