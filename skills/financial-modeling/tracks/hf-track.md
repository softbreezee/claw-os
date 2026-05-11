# HF Track — Hedge Fund / Public Equity

> Loaded when Pre-Start Check Q2 = HF or Full.
> This track is **thesis-driven**. Without a falsifiable investment hypothesis
> with positive expected value, research does not begin.

---

## Layer 0: Investment Thesis (Required Before Any Research)

All five fields below must be populated. An incomplete thesis = research
cannot start. If the user cannot fill Layer 0, the goal of Layer 1 shifts to
**helping the user find a tradeable thesis**.

### Field 1 — Core Conviction

**Template:**
> I believe [Company X]'s [Core Variable Y] will [Change A, divergent from
> consensus] within [Time Z].

**Requirements:**
- Core Variable Y must be a financial model **leading indicator** that drives price
- Change A must be explicitly divergent from consensus (not vague)
- Time Z must be specific (a quarter, not "long-term")

**Examples:**
- "I believe Nvidia's GPU utilization will fall from 85% to 65% within 12 months as custom silicon from hyperscalers scales."
- "I believe Alibaba's take rate will expand from 2.1% to 2.4% by end of FY2026 due to merchant services adoption."

### Field 2 — Consensus vs My View

| Dimension | Market Consensus | My View | Delta | Evidence |
|-----------|-----------------|---------|-------|---------|
| Core Variable | [What market assumes] | [What I believe] | [Gap] | [Data/logic] |
| Secondary 1 | [Consensus] | [My view] | [Gap] | [Evidence] |
| Secondary 2 | [Consensus] | [My view] | [Gap] | [Evidence] |

**Quality check:** Delta on Core Variable must be > 5pp or > 5% to be worth
trading. Smaller deltas get absorbed by transaction costs and price volatility.

### Field 3 — Catalyst

```
What event/data will PROVE or DISPROVE my thesis?

Prove date:       [Q report date / event date — must be a calendar date]
Key metric:       [Specific number to watch — e.g., gross margin ≥ 48%]
Trigger:          [If metric > X → thesis strengthened; < Y → falsified]
Re-pricing window:[How long after catalyst for market to reprice?]
Disprove trigger: [e.g., Management cuts guidance by > 5% → forced exit]
```

**Rule:** Catalyst must be **observable** and **datable**. "Long term" or
"eventually" are not catalysts. If catalyst is > 12 months out, the carrying
cost of the position may exceed the expected gain — flag this.

### Field 4 — Expected Value (Asymmetry Calculation)

| Scenario | Probability | Target Price | Return vs Current | EV Contribution |
|----------|------------|-------------|------------------|----------------|
| Bull (thesis correct) | P% | $X | +Y% | P% × Y% |
| Base | P% | $X | ±Y% | P% × Y% |
| Bear (thesis wrong) | P% | $X | −Y% | P% × (−Y%) |
| **Expected Value** | 100% | — | — | **Σ contributions** |

**Quality thresholds:**
- EV < 0% → Do not trade
- EV 0–5% → Too marginal; pass unless asymmetry ratio > 3×
- EV 5–8% → Worth 1–2% portfolio weight
- EV 8–12% → Worth 2–3% portfolio weight
- EV > 12% → Core position candidate (> 3%)
- Asymmetry ratio = Upside % ÷ Downside % → must be > 1.5× minimum

**Example:**
```
40% chance correct → +25% = +10.0% contribution
50% base case      → +5%  = +2.5% contribution
10% chance wrong   → −15% = −1.5% contribution
Expected Value = +11.0% → TRADE
Asymmetry = 25% / 15% = 1.67× → Acceptable
```

### Field 5 — Stop-Loss Triggers

```
Hard stop:  [Specific metric crosses specific threshold → exit within 24h]
Soft stop:  [Thesis deteriorating → review position, reduce by 30–50%]
Time stop:  [If thesis not proven by [date], exit regardless]
```

---

## Layer 1: HF-Specific Research

### 1.1 Consensus Divergence Map

Data sources: Bloomberg consensus / FactSet / Yahoo Finance target prices,
sell-side reports, and price-implied growth reverse-engineering.

| Dimension | Market Consensus | My Divergence View | Delta | Evidence Quality |
|-----------|-----------------|-------------------|-------|----------------|
| FY+1 Revenue Growth | [%] | [%] | [pp] | [source] |
| EBITDA Margin | [%] | [%] | [pp] | [source] |
| Long-term ROIC | [%] | [%] | [pp] | [source] |
| Terminal Growth Rate | [%] | [%] | [pp] | [source] |
| Valuation Multiple | [x] | [x] | [x] | [source] |

Mandatory question set for Layer 1:

**S1:** What is sell-side consensus (price target, buy/hold/sell split)? Where
do the bulls and bears disagree?

**S2:** How large is my Delta vs consensus? Is it large enough to overcome
transaction costs and market noise? (> 5pp or > 5% threshold)

**S3:** What specific data point / event will cause the market to reprice my
thesis? What is the timeline?

**S4:** What is my target position size? (drive from Expected Value in Layer 0)

### 1.2 Value Bridge

Map the delta between consensus target price and my fair value estimate:

```
Consensus Target Price:  $120  (+20% vs current $100)

My Valuation Bridge:
  Start: Consensus Fair Value         $120
  ± Revenue growth delta (−4pp)       −$6
  ± EBITDA margin delta (+2pp)        +$8
  ± Terminal growth delta (−0.5pp)    −$2
  ─────────────────────────────────
  My Fair Value                       $120  (similar in this example)

  Risk-adjusted (60% prob):  $120 × 0.6 + $85 × 0.4 = $106
  Expected return at $100:   +6% → Marginal; review asymmetry ratio
```

### 1.3 Falsification Thresholds

Set before research begins. Non-negotiable exit triggers:

```
My thesis is DISPROVEN if next quarter shows:
☐ Revenue < $[X]B     → implies market share loss, not temporary
☐ Gross margin < [Y]% → implies pricing power weaker than assumed
☐ FY guidance cut > [Z]% → implies demand deterioration
☐ Competitive market share down > [N]pp

IF ANY above occurs → MANDATORY EXIT within 2 trading days.
No exceptions. No "wait and see."
```

### 1.4 Thesis Validation Statement (Mandatory After Layer 1)

```
[INVESTMENT THESIS VALIDATION]
Pre-research hypothesis: [restate Layer 0 Core Conviction]

Layer 1 research findings:
  1. Does data support my view?            ☐ Yes  ☐ No  ☐ Partial
  2. Does consensus divergence still hold? ☐ Yes  ☐ No  ☐ Narrowed
  3. Is the catalyst still valid?          ☐ Yes  ☐ No  ☐ Delayed

Conclusion:
  [✓] Supported → Proceed to Layer 2 (position sizing)
  [Δ] Partial  → Revise Core Variable / thesis; re-run EV calc
  [✗] Rejected → Abandon position OR flip to short

Revised thesis (if needed): [new Core Conviction]
```

Revision triggers:
- Consensus divergence now < 5pp (too small to trade)
- Catalyst now > 12 months (carrying cost too high)
- Core Variable correlation with price < 0.65 (too weak)
- EV turned negative

---

## Layer 2: Single-Factor Sensitivity Analysis

**Principle:** Lock all assumptions at Base Case. Vary only Core Variable Y
across its P10–P90 historical or industry range. Quantify the asymmetry.

### Template

**Core Variable: FY+1 Gross Margin (range: 26% → 32%)**

| Margin % | Scenario | EBITDA Impact | Target Price | Return vs Current | Weight |
|---------|---------|--------------|-------------|-----------------|--------|
| 26% | Bear (P10) | −$200M | $82 | −18% | 10% |
| 27% | Low | −$160M | $88 | −12% | 25% |
| 28% | **Base (P50)** | — | $100 | 0% | 50% |
| 30% | High | +$160M | $115 | +15% | 75% |
| 32% | Bull (P90) | +$320M | $130 | +30% | 90% |

**Asymmetry conclusion:**
```
If thesis is WRONG (margin at P10–P25):  avg return −15%
If thesis is CORRECT (margin at P75–P90): avg return +22.5%
Asymmetry Ratio: 22.5% / 15% = 1.5× → Acceptable (minimum threshold)
```

**2D Sensitivity Table (when Base Case depends on two key variables):**
```
                    Revenue Growth: 6%  | 8%  | 10% | 12%
EBITDA Margin: 25%                 $80  | $90  | $100 | $110
               28%                 $95  | $110 | $125 | $140
               30%                 $110 | $130 | $150 | $170
```

**Constraints:**
- P10–P90 range must be anchored to historical distribution or industry data
- All non-Core Variables must stay locked at Base Case
- Both 3-scenario EV (from Layer 0) AND single-factor sensitivity are required

---

## Layer 2 Supplemental: Adversarial Scenario (Management vs My View)

When management guidance is available, generate a structured challenge version.

**Step 1:** Decompose management guidance
```
Management FY2026 claims:
  Revenue:        $1.2B (+15% YoY)
  EBITDA Margin:  32%
  Implied EBITDA: $384M
  Valuation @18×: $6.9B ($92/share)
```

**Step 2:** Challenge version (cut 3 most optimistic assumptions 20–30%)
```
Revenue:        $1.14B (80% of guided) — competitive share loss risk
EBITDA Margin:  29% — cost inflation not fully offset
CapEx intensity: 6% (vs 5% guided) — growth reinvestment underestimated
Challenge valuation: $5.1B ($68/share)
```

**Step 3:** Valuation bridge (Management → My view)

| Assumption | Management | My Challenge | Price Impact | Cumulative |
|-----------|-----------|-------------|-------------|-----------|
| Revenue | $1.2B | $1.14B | −$6/share | −$6 |
| Margin | 32% | 29% | −$4/share | −$10 |
| CapEx | 5% | 6% | −$2/share | −$12 |
| TGR | 3% | 2.5% | −$3/share | −$15 |
| Multiple | 18× | 15× | −$8/share | −$23 |
| **Fair Value** | **$92** | **$69** | **−$23** | |

---

## Layer 3: Position Sizing Rules

### EV-to-Position Anchor Table

| Expected Value | Asymmetry Ratio | Recommended Weight | Max Drawdown Tolerance |
|--------------|----------------|-------------------|----------------------|
| EV < 2% | < 1.5× | Pass — no position | — |
| EV 2–5% | 1.5–2.0× | 0–1% | −5% |
| EV 5–8% | 2.0–3.0× | 1–2% | −10% |
| EV 8–12% | 3.0–4.0× | 2–3% | −15% |
| EV > 12% | > 4.0× | > 3% (Core) | −20% |

Full position sizing methodology: `references/position-sizing.md`

### Exit Rules

1. **Hard stop:** Any falsification metric hit → exit within 24 hours, no exceptions
2. **EV erosion:** EV drops to 0% → reduce 25%; EV drops to −2% → full exit
3. **Time stop:** Catalyst date passed with no confirmation → step down 25%/week
4. **Symmetry break:** Upside/Downside < 1.3× → reduce weight 30%
5. **Risk budget:** Single position daily loss > 1% of portfolio → reduce 50%

---

## Post-Mortem Template

Run after any position is closed (profit or loss).

```
[POSITION POST-MORTEM]
Thesis:          [Core Conviction stated at entry]
Hold period:     [Entry date] → [Exit date]
Realized return: [+X% / −X%]

Key variable outcomes (forecast vs actual):
  [Variable 1]: Forecast [X%] → Actual [Y%] → Δ [Z pp]
  [Variable 2]: Forecast [X%] → Actual [Y%] → Δ [Z pp]

Why did I win / lose?
  [Analysis of gap between expected and actual]

Thesis errors (if loss):
  ☐ Consensus divergence was smaller than I thought
  ☐ Catalyst timing was wrong
  ☐ Core Variable was not the actual price driver
  ☐ Position sizing was too large for the confidence level
  ☐ Falsification rules were not enforced

Improvements for next position:
  [3 specific changes to thesis process, sizing, or exit discipline]
```

Full template: `templates/hf-post-mortem.md`

---

## Quick Reference Checklists

**Layer 0:**
- [ ] Core conviction in one sentence; specific variable, direction, timeframe
- [ ] Consensus delta > 5pp on Core Variable
- [ ] Catalyst is observable and dated
- [ ] EV > 5% (minimum to trade)
- [ ] Stop-loss conditions defined (hard / soft / time)

**Layer 1:**
- [ ] Sell-side consensus confirmed (Bloomberg/FactSet)
- [ ] Divergence map complete (3–5 dimensions)
- [ ] 3–4 binary falsification triggers defined
- [ ] Thesis Validation Statement completed (yes/partial/rejected)
- [ ] EV and asymmetry ratio recalculated post-research

**Layer 2:**
- [ ] Core Variable single-factor sensitivity: P10–P90 range
- [ ] Worst-case (P10) and best-case (P90) return spread > 30pp
- [ ] Asymmetry ratio > 1.5×
- [ ] Position size mapped to EV range

**Adversarial (if management guidance available):**
- [ ] Challenge version: cut 3 most optimistic assumptions 20–30%
- [ ] Valuation bridge computed (management target → my fair value)
- [ ] Upgrade/downgrade triggers documented

---

*HF Track v1.0 | Investment Intelligence Suite | Condensed from FSM v5.0 integrated-modeling-hf*
