# LBO Framework: Entry to Exit Math

> "Private equity is an exercise in buying cash flows at the right price, engineering a capital structure, and selling at a premium. The math is simple. The discipline is hard."

---

## 1. The LBO Equation

Every buyout starts with the same fundamental equation:

```
Enterprise Value (Entry) = Debt + Equity Check

Where:
  Enterprise Value = Entry Multiple × LTM EBITDA
  Debt             = Maximum sustainable leverage (typically 4–6x EBITDA)
  Equity Check     = EV − Debt (what the PE fund must write)
```

**Example:**
```
LTM EBITDA:      $100M
Entry Multiple:   9.0x
Enterprise Value: $900M
Debt (5.0x):      $500M  → interest at ~7% = $35M/year
Equity Check:     $400M
```

The equity check is the fund's capital at risk. The returns are driven entirely by what happens to that $400M over the hold period.

---

## 2. Debt Structure: How PE Leverages the Capital Stack

### Typical Debt Tranches (2020s market conditions)

| Tranche | Sizing | Rate | Features |
|---------|--------|------|----------|
| Revolving Credit Facility | 0.5–1.0x EBITDA | SOFR + 200–350bps | Working capital; drawn as needed |
| Term Loan B (TLB) | 3.0–4.5x EBITDA | SOFR + 300–500bps | Institutional; minimal amortization (1%/yr) |
| Second Lien / Unitranche | 0.5–1.0x EBITDA | SOFR + 600–800bps | Higher yield; structural subordination |
| Mezzanine / Preferred | 0.5–1.0x EBITDA | 12–15% PIK | Used in large/complex deals |

**Typical total leverage:** 4.0–6.0x EBITDA for sponsored buyouts
**Asian market context:** Korean and Japanese lenders more conservative (3.5–5.0x); Chinese onshore leverage often constrained by regulatory capital rules

### Debt Capacity Analysis

Minimum debt service coverage ratios that banks require:
- **Interest Coverage:** EBITDA / Interest ≥ 2.0x (hard floor); ≥ 2.5x preferred
- **Fixed Charge Coverage:** (EBITDA − CapEx) / (Interest + Mandatory Amort) ≥ 1.2x
- **Leverage Ceiling:** Net Debt/EBITDA ≤ 6.5x at close (covenant); lenders step in at breach

**Debt capacity formula (quick screen):**
```
Maximum Debt = Min(
  6.0x × LTM EBITDA,                   ← leverage multiple ceiling
  EBITDA × (1 − tax rate) / (Rate × 2.0)  ← 2x coverage floor
)
```

---

## 3. The 5-Year EBITDA Build

The investment thesis is expressed as a 5-year EBITDA trajectory. Every number must be justified by specific operational initiatives.

### Build Structure

```
Year 0 (Entry)     EBITDA:  $100M   ← LTM actual, normalized
Year 1             EBITDA:  $112M   ← Quick wins: pricing +5%, SG&A cut
Year 2             EBITDA:  $126M   ← Revenue growth: new geographies/products
Year 3             EBITDA:  $138M   ← Add-on acquisition integration
Year 4             EBITDA:  $152M   ← Margin expansion: procurement savings
Year 5 (Exit)      EBITDA:  $165M   ← Steady-state + add-on contribution
CAGR (5Y):         10.5%
```

**Red flags in EBITDA build:**
- CAGR > 15% without specific, named initiatives → hockey stick fantasy
- Margin improvement > 5 points without structural cost actions → wishful thinking
- Revenue growth faster than market growth without market share thesis → unsupported
- Large add-on contributions without deal pipeline identified → speculative

### EBITDA Normalization Adjustments (common)

| Adjustment | Direction | Typical Size |
|------------|-----------|-------------|
| Owner compensation above market | Add back | $2–15M |
| Non-recurring legal/restructuring | Add back | $1–10M |
| Earnout payments to former owners | Add back | Variable |
| Stock-based compensation (cash method) | Add back | $1–5M |
| Pro-forma for completed acquisitions | Add | Trailing 12M of acquired EBITDA |
| Revenue recognition pull-forward | Deduct | Deal-specific |
| Channel stuffing / one-time orders | Deduct | Deal-specific |
| Synergies not yet achieved | Deduct (if unconfirmed) | Deal-specific |

---

## 4. Debt Paydown Waterfall

Over the hold period, FCF pays down debt, creating equity value even if EBITDA doesn't grow.

```
Year 0 Debt:       $500M
Annual FCF:        ~$65M (EBITDA $100M × 65% conversion)
Mandatory Amort:   $5M/yr (TLB at 1% of original balance)
Debt Paydown/yr:   $60–70M (discretionary sweep after mandatory)

Year 5 Debt:       $500M − ($65M × 5) = ~$175M (if all FCF sweeps debt)
Net Debt reduction: $325M over 5 years
```

**Note:** In practice, not all FCF goes to debt paydown. CapEx for growth, add-on acquisitions, and dividend recaps all compete for the same cash.

---

## 5. Exit Equity Calculation

```
Exit Enterprise Value = Exit Multiple × Year 5 EBITDA
                      = 9.0x × $165M = $1,485M

Exit Equity Value     = Exit EV − Exit Net Debt
                      = $1,485M − $175M = $1,310M

Entry Equity Check:   $400M
Exit Equity Value:    $1,310M

MOIC:  $1,310M / $400M = 3.28x
IRR:   ~27% (using 5-year hold, approximate)
```

**IRR approximation formula (quick mental math):**
- 2.0x in 5 years ≈ 15% IRR
- 2.5x in 5 years ≈ 20% IRR
- 3.0x in 5 years ≈ 25% IRR
- 3.5x in 5 years ≈ 29% IRR
- 4.0x in 5 years ≈ 32% IRR

---

## 6. Sources of Return Decomposition

A rigorous LBO analysis always decomposes where the return comes from. MBK's discipline: EBITDA growth and debt paydown should account for > 80% of returns; multiple expansion is a gift, not a plan.

```
Total Return Decomposition (example):

Entry EV:    $900M  (9.0x × $100M)
Exit EV:     $1,485M (9.0x × $165M)
EV Change:   +$585M

  Of which:
  EBITDA Growth:      ($165M − $100M) × 9.0x = $585M EV impact
  Multiple Expansion: ($0 — same multiple)    = $0
  Debt Paydown:       $500M − $175M           = $325M equity impact
  FCF Yield:          Captured in debt paydown/EV calc

As % of equity value created ($910M total):
  EBITDA Growth:         64%
  Debt Paydown:          36%
  Multiple Expansion:     0%
```

**Warning signs in return decomposition:**
- Multiple expansion > 30% of return → fragile thesis (depends on market)
- Debt paydown > 60% of return with flat EBITDA → financial engineering, not value creation
- Financial engineering as sole driver → does not pass MBK's quality bar

---

## 7. Sensitivity Tables

### Table A: Entry Multiple vs. Exit Multiple (MOIC)

*Assumptions: $100M LTM EBITDA, $165M Exit EBITDA (5Y), $500M initial debt, $175M exit debt*

| Entry Multiple \ Exit Multiple | 7.0x | 8.0x | 9.0x | 10.0x | 11.0x |
|-------------------------------|------|------|------|-------|-------|
| **7.0x** | 2.9x | 3.5x | 4.1x | 4.7x | 5.3x |
| **8.0x** | 2.3x | 2.8x | 3.3x | 3.8x | 4.3x |
| **9.0x** | 1.9x | 2.3x | 2.8x | 3.2x | 3.6x |
| **10.0x** | 1.6x | 2.0x | 2.3x | 2.7x | 3.1x |
| **11.0x** | 1.3x | 1.7x | 2.0x | 2.3x | 2.6x |

*Green (≥ 2.5x MOIC) = meets hurdle. Yellow = borderline. Red = avoid.*

### Table B: Revenue CAGR vs. Initial Leverage (IRR)

*Assumptions: entry at 9.0x, exit at 9.0x, 5-year hold*

| Revenue CAGR \ Initial Leverage | 3.0x | 4.0x | 5.0x | 6.0x |
|----------------------------------|------|------|------|------|
| **5%** | 18% | 20% | 22% | 19% |
| **8%** | 22% | 25% | 27% | 24% |
| **10%** | 25% | 28% | 31% | 27% |
| **12%** | 28% | 32% | 35% | 30% |
| **15%** | 33% | 37% | 40% | 34% |

*Note: High leverage improves IRR until coverage ratios constrain; diminishing returns above 6x.*

---

## 8. Downside / Stress Case

MBK's discipline: always model a downside case. The downside must show capital preservation.

**Stress Case Assumptions:**
- Revenue flat (0% growth) for Years 1–2, then 5% recovery
- Margin compression 200bps from cost inflation
- Exit multiple contracts 1.5x from entry
- Debt paydown 50% of base case (FCF consumed by working capital stress)

**Downside minimum standard:**
- MOIC ≥ 1.0x (no capital loss)
- Interest coverage > 1.5x in worst year (not in default)
- If downside breaches these floors → deal requires re-pricing or is a pass

---

## Quick Reference: LBO Math Checklist

- [ ] LTM EBITDA normalized and defensible
- [ ] Debt capacity checked at 2x coverage minimum
- [ ] 5-year EBITDA build tied to specific initiatives
- [ ] Debt paydown schedule modeled (mandatory + discretionary)
- [ ] Base case: MOIC ≥ 2.5x, IRR ≥ 20%
- [ ] Downside case: capital preserved, no covenant breach
- [ ] Sources of return decomposed (growth vs. multiple vs. leverage)
- [ ] Sensitivity tables run for entry/exit multiple and growth rate
- [ ] Management case (what management claims) vs. sponsor case documented separately
