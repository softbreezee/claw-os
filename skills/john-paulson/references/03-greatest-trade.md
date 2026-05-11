# The Greatest Trade Ever: The 2007-2008 Subprime Short

> **When to read this file:** When analyzing systemic mispricings, when studying how to construct asymmetric macro trades via credit derivatives, when understanding CDS mechanics on asset-backed securities, or when the task involves identifying a widely-held false consensus in credit markets.

The 2007-2008 subprime short is the most important case study in Paulson's career — and arguably the most profitable single trade in financial history. Between 2007 and 2008, Paulson's funds earned approximately $15-20 billion in profit by betting against subprime mortgage securities. Understanding HOW he identified it, HOW he constructed the trade, and WHAT signs he used to time the entry and exit is essential to understanding Paulson's deepest intellectual contribution to investing.

> "Everyone could see the data. The delinquency rates were in public filings. The leverage ratios were disclosed. We just looked at it with different eyes." — Paulson

---

## Phase 1: Identifying the Mispricing (2005-2006)

### The Starting Observation

By 2005, the U.S. housing market showed unprecedented stress signals that most participants were ignoring or rationalizing:

**Observable facts (all publicly available)**:
1. Home prices had risen 50-100%+ in many markets (2000-2005) with no fundamental income growth
2. Mortgage underwriting standards had collapsed: no-doc loans ("NINJA" — no income, no job, no assets), stated income, interest-only, negative amortization
3. Loan-to-value ratios had risen from average 70-75% (2000) to 95-100%+ (2005-2006)
4. Adjustable-rate mortgages (ARMs) with 2-year teaser rates were being used for 30-year mortgages — ensuring payment shock when rates reset
5. Subprime origination volume had exploded: from $100bn/year (2000) to $600bn+/year (2005-2006)
6. Rating agencies were rating subprime MBS tranches as AAA despite loans that would never have been made by traditional banks

**The consensus view**: "Home prices nationally have never declined. Even if some regions soften, the diversification across geographies protects MBS portfolios."

**Paulson's counter-analysis**: The "national diversification" argument was circular — all markets had been inflated by the same deterioration in lending standards. A systemic cause (loose credit) would produce a systemic effect (correlated defaults), not the idiosyncratic geography-specific failures the rating models assumed.

> "The rating agencies assumed a 7% national home price decline as their stress case. We thought the real risk was 30-40%. Not because we were smarter — but because we looked at the actual loan files."

### The Analytical Work

Paulson's team (led by Paolo Pellegrini, his key analyst on this trade) dug into mortgage-backed security (MBS) loan-level data:

1. **Loan tapes**: Individual loan characteristics for specific securitized pools — LTV, FICO, documentation type, property type, originator
2. **FFELP vs. subprime comparison**: Compared historical default rates on properly underwritten loans vs. the current origination quality
3. **Payment reset modeling**: Modeled what would happen when 2-year ARMs reset to fully-indexed rates in 2007-2008 — payment increases of 30-50%+ for borrowers who already had no equity
4. **Geographic concentration**: Despite the "diversified" marketing of MBS, the underlying loans were heavily concentrated in bubble markets (CA, FL, NV, AZ)
5. **Originator quality analysis**: Identified which originators (New Century, IndyMac, AmeriQuest) were producing the worst loans — and which securitizations contained them

**The conclusion**: The worst subprime MBS tranches were priced as if they had 2-3% default rates. Paulson's models showed 30-50%+ default rates were likely under moderate stress scenarios. The gap between market pricing and realistic default rates was the mispricing.

---

## Phase 2: The Instrument — Credit Default Swaps on Subprime MBS

### Why Equity Shorts Wouldn't Work

The obvious trade was to short the stocks of mortgage originators (New Century, Countrywide) or investment banks with MBS exposure. **Problems with this approach**:
1. **Convexity**: Short stock positions have limited upside (stock can only go to zero) and theoretically unlimited downside
2. **Timing**: Could take years before the stocks fell; short-selling has carry costs (borrow fees, dividends)
3. **Scale**: Paulson needed to express a multi-billion-dollar view; large short positions in individual stocks are hard to establish without moving the market

### Credit Default Swaps (CDS) — The Instrument

A Credit Default Swap is an insurance-like derivative:
- **Protection buyer** (Paulson): Pays a periodic premium (like an insurance premium, measured in basis points per year)
- **Protection seller** (investment banks): Receives the premium; pays Paulson the face value of the bond minus recovery value if the reference bond defaults

**Example mechanics**:
```
Reference: $10M face value of "BBB-" rated subprime MBS tranche
CDS Premium: 50 bps/year (0.50% of notional)
Annual cost to Paulson: $50,000/year
If the tranche defaults/gets impaired: Paulson receives ~$8-9M (face minus recovery)
Profit: ~$8-9M on $50,000/year cost = 160-180x payoff if thesis is right
```

**Why CDS was the right instrument**:
1. **Defined, limited downside**: Maximum loss = premiums paid while thesis plays out
2. **Massive asymmetric upside**: If the securities went to zero, CDS paid the full notional
3. **No borrow issues**: No securities needed to be located and borrowed for short sale
4. **Scale**: The CDS market on subprime MBS was enormous (notional > $2 trillion by 2006-2007)

### The ABX Index

The ABX.HE index was created in 2006 as a standardized way to trade credit derivatives on subprime MBS. It had tranches corresponding to different rating levels (AAA, AA, A, BBB, BBB-) of subprime RMBS from recent vintages.

**Paulson's approach**: Focus on the lower tranches (BBB and BBB-) of the ABX — these were the "first loss" tranches that would be wiped out first when defaults mounted. They also had the most mispriced risk relative to premiums.

**The ABX trade**:
- Buy protection on BBB and BBB- tranches of 2005-2006 vintage subprime MBS
- Pay ~50-150 bps/year for protection on bonds that were actually at risk of 70-100% loss
- The implied loss in the market price: ~5-10%
- Paulson's estimated actual loss: 50-70%

---

## Phase 3: Construction and Sizing

### Position Building (2006-early 2007)

**Timeline**:
- **Mid-2005**: Paulson begins internal research, skeptical of housing market
- **Late 2005**: Pellegrini develops the analytical framework; confirms the mispricing
- **Early 2006**: Paulson Credit Opportunities Fund launched specifically to execute this trade
- **2006**: Gradual accumulation of CDS protection across multiple instruments and reference bonds
- **Q4 2006**: Position fully sized; fund heavily concentrated in the short thesis

**The sizing rationale**:
- CDS premiums were low (50-150 bps) relative to the potential recovery
- Even at maximum premium cost, the carry was manageable (< 1-2% of AUM per year)
- If the thesis was right, the payoff would be 50-100x the annual carry cost
- Paulson was willing to pay the carry for multiple years while waiting for the thesis to manifest

### The Crucial Insight: Asymmetric Risk/Reward

```
Worst case (thesis completely wrong): 
  Pay 150 bps/year × 3 years = 4.5% of notional in premiums
  
Best case (thesis right, securities go to zero):
  Receive ~90-95% of notional face value
  
Risk/reward: risking 4.5% to make 90-95% = 20:1+ asymmetry
```

This asymmetry is what made the position "large" in expected value terms even if the probability of being right was uncertain.

---

## Phase 4: The Trade Plays Out (2007-2008)

### The Bear Market Begins (Mid-2007)

**February 2007**: HSBC announces $10.6B in write-offs on U.S. subprime mortgages — first major public acknowledgment
**June 2007**: Two Bear Stearns hedge funds that had sold CDS protection to Paulson (and others) collapse
**July-August 2007**: BNP Paribas freezes three money market funds exposed to U.S. subprime; European credit markets seize
**August 2007**: ABX index collapses; the CDS positions Paulson held suddenly mark to market at enormous gains

**Paulson's Q1-Q3 2007 returns**: Paulson Credit Opportunities Fund I: +490%; Fund II (launched 2007): +350%

### The Financial Crisis Deepens (2008)

By early 2008, Paulson had largely converted the subprime CDS profits into a new thesis: **short the equity of major financial institutions** that had written protection on CDOs containing subprime MBS (Citigroup, Bank of America, UBS, Lehman Brothers, Washington Mutual, Wachovia).

**The logic**:
- Banks had enormous hidden losses in "Level 3" assets (illiquid, marked to model)
- When losses were recognized, bank equity would be severely diluted or wiped out
- Short equity of financial institutions was the cleanest expression of the "banks are insolvent" thesis

**2008 performance**: Paulson Advantage Fund +37%; additional profits from financial shorts

---

## Lessons for Systemic Mispricing Analysis

### Lesson 1: Consensus Is Not Evidence

The "housing prices have never declined nationally" argument was historical fact — but it was the result of lending standards that had always existed. Those standards were gone by 2005-2006. Past performance of the asset class was not predictive when the input quality had fundamentally changed.

> "When everyone agrees something is safe, ask: what has changed that would make the historical safety record no longer applicable?"

### Lesson 2: Look at Primary Data, Not Ratings

Rating agencies modeled subprime defaults using historical data from 1998-2002 (a period of normal lending). The actual loan files for 2005-2006 vintages showed categorically different characteristics. Primary research beat consensus research.

**Framework**: When evaluating credit risk, go to loan-level data, not ratings summaries. Ratings are opinions; loan tapes are facts.

### Lesson 3: Find the Asymmetric Instrument

Even if you identify a mispricing, the instrument matters. Paulson could have shorted homebuilder stocks or mortgage originator stocks. Instead, he found an instrument (CDS) where:
- The cost of being wrong was bounded
- The profit from being right was uncapped
- The sizing could be large without market impact

**Paulson's instrument selection framework**:
1. What is the maximum loss if I'm wrong?
2. What is the gain if I'm right?
3. Can I size this position appropriately given my conviction and fund size?
4. What is the carry cost while waiting for the thesis to resolve?

### Lesson 4: The "Greatest Trade" Checklist

For any potential systemic mispricing, ask:
- [ ] Is there a widely-held consensus belief that is built on an empirically outdated assumption?
- [ ] Is there primary data (loan tapes, flow data, credit metrics) that contradicts the consensus?
- [ ] Is there a derivative or structured instrument that allows asymmetric exposure?
- [ ] What is the carry cost per year to maintain the position?
- [ ] What is the expected payoff if correct? Is the risk/reward > 10:1?
- [ ] What would cause you to admit you're wrong and exit?

> "The hardest part wasn't finding the trade. The hardest part was being willing to be alone on the other side of a consensus held by virtually every bank, rating agency, and government regulator in the world — for over a year — while paying premium carry and watching the market disagree with you every day."
