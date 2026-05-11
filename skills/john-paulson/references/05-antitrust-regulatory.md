# Antitrust and Regulatory Risk Framework

> **When to read this file:** When assessing antitrust and regulatory risk in announced M&A deals, when evaluating HSR filing timelines, when understanding how DOJ and FTC analyze deals, when modeling probability of regulatory clearance, or when assessing the impact of remedy requirements on deal economics.

Antitrust risk is the most complex and consequential regulatory hurdle in modern merger arbitrage. Unlike financing risk (bounded by deal documents) or shareholder vote risk (bounded by known holder positions), antitrust risk requires understanding legal doctrine, economic analysis, political environment, and precedent simultaneously. Getting this analysis right — or wrong — often determines the success or failure of an arb position.

> "The question isn't just whether the deal violates antitrust law. The question is what the specific regulator in the specific political environment at the specific moment in history believes about it — and whether they'll fight it in court."

---

## The U.S. Antitrust Framework

### Two Agencies, One Process (Usually)

U.S. antitrust review of mergers is conducted by either the **Department of Justice Antitrust Division** or the **Federal Trade Commission**. The two agencies divide jurisdiction by industry:

**DOJ Antitrust Division typical sectors**: Technology, media/telecom, defense, transportation, financial services, agriculture
**FTC typical sectors**: Healthcare, pharmaceuticals, consumer products, retail, energy

The split is not always clean — clearance decisions between agencies happen informally. What matters for arb:
- DOJ sues in federal court to block a deal
- FTC can issue administrative complaints OR seek injunctions in federal court
- Either can negotiate consent decrees (remedies)

### The Hart-Scott-Rodino (HSR) Filing Process

The HSR Act requires parties to file pre-merger notification and wait for regulatory review before closing. Understanding this timeline is essential for arb.

**HSR Thresholds (2024 levels)**:
- Size of transaction threshold: >$119.5M (adjusted annually for inflation)
- Size of person threshold: acquirer/target > $23.9M in assets/sales

**Phase 1 — Initial Review**:
- Filing triggers 30-day waiting period (15 days for cash deals)
- Agency reviews filing to decide if deeper review needed
- If no action: deal can close after waiting period expires
- Probability of receiving Second Request: ~3-5% of all HSR filings; much higher for large horizontal deals

**Phase 2 — Second Request**:
- Agency issues "Second Request" for additional documents and data
- Parties must substantially comply before restart of 30-day period
- Average time from Second Request to resolution: 6-12 months additional
- During Second Request: agency is building a case; probability of challenge higher
- ~40-50% of deals receiving Second Request are challenged or require significant remedies

**Timeline model for arb**:
```
HSR Filing + 30 days → Phase 1 clear (best case)
HSR Filing + 12-18 months → Phase 2 extensive review
HSR Filing + 18-30 months → Litigation (DOJ sues to block)
```

---

## The Legal Standard: What Regulators Must Prove

### Section 7 of the Clayton Act

The legal standard: a merger is unlawful if its effect "may be substantially to lessen competition, or to tend to create a monopoly."

Key operative words: **"may be"** — the standard is forward-looking probability, not certainty. Regulators don't need to prove competition will be harmed; they need to show it's probable.

### The Two-Step Analytical Framework

**Step 1: Market Definition**
- Regulators must first define the relevant market (product + geography)
- Narrow market definition → higher concentration → easier to challenge
- Wide market definition → lower concentration → harder to challenge
- Example: Is Staples/Office Depot in the "office supply superstore" market (narrow, high concentration) or "office supplies" (wide, includes Amazon/Walmart, lower concentration)?

**Step 2: Competitive Effects Analysis**
Two theories of harm:
1. **Unilateral effects**: Combined entity can profitably raise prices without needing competitor cooperation (most common in modern antitrust)
2. **Coordinated effects**: Merger makes tacit or explicit coordination more likely (less common)

### The HHI Calculation

Herfindahl-Hirschman Index is the standard concentration measure:

```
HHI = Sum of (market share)² for all competitors
  e.g., 4-player market: 40%, 30%, 20%, 10%
  HHI = 40² + 30² + 20² + 10² = 1600 + 900 + 400 + 100 = 3000

Post-merger (if 40% + 30% combine):
  Combined: 70%, 20%, 10%
  HHI = 70² + 20² + 10² = 4900 + 400 + 100 = 5400
  ΔHHI = 5400 - 3000 = 2400 (very significant)
```

**DOJ/FTC Horizontal Merger Guidelines thresholds**:
- Post-merger HHI < 1500: Unconcentrated — unlikely to challenge
- Post-merger HHI 1500-2500: Moderately concentrated — challenge if ΔHHI > 100
- Post-merger HHI > 2500: Highly concentrated — challenge if ΔHHI > 200
- ΔHHI > 200 in highly concentrated market: "presumed to be likely to enhance market power"

---

## Types of Regulatory Outcomes

### Outcome 1: Unconditional Clearance ("Phase 1 Clear")

Deal cleared without any remedies or conditions. Happens when:
- No meaningful market overlap
- HHI below thresholds
- Competitive effects analysis shows no plausible harm

**Arb impact**: Spread collapses to time value only (cost of carry) — best outcome

### Outcome 2: Consent Decree with Behavioral Remedies

Deal cleared conditional on ongoing behavioral commitments. Examples:
- Firewall between combined entity's data and competitive businesses
- Commitment to continue licensing intellectual property on fair terms
- Prohibition on certain sales practices
- Open access requirements

**Arb impact**: Deal closes; spread collapses; behavioral remedies rarely affect deal economics materially

**Trend**: DOJ has moved AWAY from behavioral remedies under recent administrations — prefers structural remedies

### Outcome 3: Consent Decree with Structural Remedies (Divestitures)

Deal cleared conditional on divesting overlapping businesses. Examples:
- T-Mobile/Sprint: Must divest to create DISH as 4th carrier
- AT&T/DirecTV: Must divest regional sports networks
- Albertsons/Safeway: Must divest hundreds of stores in overlapping geographies

**Arb impact**: Deal closes, but:
- Divestiture process takes time → extended timeline → reduced annualized return
- Divestiture buyer quality matters — weak buyer may not satisfy regulator
- Divestitures sometimes reduce deal economics (buyer negotiates hard)

**Evaluating remedy sufficiency**:
- Is the divested business "viable standalone"? Regulators require this.
- Is the divestiture buyer strong enough to compete effectively? (DISH-as-remedy was criticized)
- Does the remaining combined entity still have an antitrust problem after divestiture?

### Outcome 4: Challenge and Litigation

Agency sues to block deal in federal court. Now requires parties to:
1. Fight in court (18-24 months minimum)
2. Accept remedies more extensive than originally offered
3. Walk away from the deal (if merger agreement permits)

**Probability of DOJ/FTC winning in court**:
- Historical win rate for DOJ/FTC: ~55-65% on injunctive relief (preliminary injunction)
- But: even losing at PI stage often ends deal (parties don't want extended uncertainty)
- Recent trend: courts have become somewhat more skeptical of government theories

### Outcome 5: Deal Abandoned Before Litigation

Parties abandon deal after receiving Second Request or after DOJ signals intent to sue:
- Faster resolution than litigation
- No reverse break fee in many cases (deal was terminated mutually)
- Target stock falls sharply to unaffected price

---

## International Regulatory Complexity

### EU Merger Control

European Commission reviews mergers meeting EU thresholds (combined worldwide turnover > €5B, EU-wide turnover >€250M each).

**Key differences from US**:
- EC has more power to block deals unilaterally (no need for court approval)
- "Significant impediment to effective competition" standard (subtly different from US)
- EC has historically been more willing to impose behavioral remedies than recent US DOJ
- Timeline: Phase 1 (25 days) → Phase 2 (90 days, extendable to 125 days)

**Intel/McAfee, GE/Honeywell examples**: EU blocked deals that US cleared — different analytical framework

### China MOFCOM

MOFCOM (Ministry of Commerce) review has become increasingly significant:
- Required for any deal where either party has significant China revenues
- Semiconductors, technology, and strategic sectors subject to heightened scrutiny
- MOFCOM can impose conditions designed to benefit Chinese competitors
- Political/trade war dimension: MOFCOM approval is partly a geopolitical tool
- Average timeline: 30-180 days; can extend indefinitely

**Arb framework for China risk**:
- If deal requires MOFCOM approval: add 10-20% to timeline; add 5-15% break risk
- Semiconductor deals: add 15-25% break risk regardless of market share analysis
- During US-China tension: treat MOFCOM as unpredictable, not analytical

### CFIUS (Foreign Investment)

Committee on Foreign Investment in the United States reviews foreign acquirers of US businesses.
- Covers national security concerns (technology, infrastructure, defense supply chain, personal data)
- Can require divestitures, security agreements, or block transactions outright
- Mandatory review for certain critical technology and infrastructure deals
- Timeline: 30-day review → 45-day investigation → Presidential decision

**When CFIUS matters in arb**: When acquirer is foreign (especially Chinese, Russian, or from adversarial nations) acquiring US technology, defense, critical infrastructure, or personal data companies.

---

## Enforcement Environment: Reading the Political Tea Leaves

Antitrust enforcement philosophy changes with administrations and agency leadership. This has direct impact on deal completion probability.

**Obama era (2009-2017)**: Active enforcement, particularly on healthcare and financial services; willing to accept behavioral remedies
**Trump era (2017-2021)**: Generally permissive on economic mergers, but DOJ challenged AT&T/TW (political?); FTC quiet
**Biden era (2021-2025)**: Most aggressive since the 1970s; Khan FTC and Kanter DOJ challenged unprecedented deals; courts pushed back on FTC theories
**Current environment**: Assess current chair's stated philosophy; recent win/loss record; which deals they've challenged

**Signals that a deal faces hostile environment**:
- Chair has publicly commented negatively on sector consolidation
- Agency has recently challenged similar deals (even if they lost)
- HSR second request issued quickly (within days, not weeks)
- Agency hires/fires personnel associated with particular enforcement philosophy
- Political environment: populist/anti-big-business sentiment in Congress

> "Antitrust is law + economics + politics. You have to read all three simultaneously. A deal that would have been easily cleared in 2019 might face a year of litigation in 2023 with a different FTC chair."
