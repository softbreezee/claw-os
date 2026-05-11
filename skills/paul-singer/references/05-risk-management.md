# Elliott Risk Management: Aggression Without Blowup

> **When to read this file**: when sizing an activist position, designing the hedging strategy for a campaign, evaluating how to manage campaign risk (timeline extension, management response, legal risk), assessing liquidity management during a proxy fight, or understanding why Elliott has such remarkably low drawdowns despite aggressive tactics.

---

## The Paradox: Aggressive Tactics, Conservative Risk Profile

Elliott has generated ~13-14% net annual returns since 1977 with extremely few negative years (the fund has never lost money in a calendar year for most of its history through at least the early 2020s). This performance combines aggressive campaign tactics with rigorous risk management — a combination most observers find contradictory.

> "We are extremely risk-conscious. We think about all the ways we can lose money before we think about the ways we can make money." — Paul Singer

> "Our goal is to make money in good markets and not lose money in bad markets. We are always thinking about what can go wrong." — Elliott Management investor letter

**The key insight**: Activist positions are not normal equity positions. They carry four distinct risk types that must each be managed separately:
1. **Market/sector risk** (the stock goes down with the market)
2. **Campaign risk** (the campaign fails, thesis doesn't materialize)
3. **Legal risk** (litigation backfires, regulatory intervention)
4. **Liquidity risk** (position is too large to exit cleanly if needed)

---

## Module 1: Position Sizing for Activist Positions

### The Core Sizing Rule

Elliott's typical activist position is **3-6% of total fund capital** at initiation, scaling toward 5-8% if the campaign is proceeding well. Maximum is ~10% for highest-conviction situations.

**Rationale**:
- 3-6% allows meaningful impact without catastrophic loss if wrong
- Campaign outcomes are binary-ish (win or don't win) — sizing for binary outcomes requires smaller positions than for continuous outcomes
- Portfolio diversification across 20-30 campaigns means any single failure doesn't define the year
- Legal/reputation costs of a failed campaign can be $30-100M+ — position must be large enough to absorb this overhead and still generate adequate return

### The Sizing Framework

```
Input 1: Value gap analysis
  If value gap is 40%+, position can be at high end (5-8% of fund)
  If value gap is 25-35%, position should be at low end (3-5% of fund)
  Below 25% value gap → insufficient for an Elliott campaign

Input 2: Campaign success probability
  High probability (board seat + supportive ISS): 5-8% size
  Medium probability (complex multi-year campaign): 3-5% size
  Low probability (founder controlled, government involved): 1-3% size (speculative)

Input 3: Liquidity profile
  Highly liquid target (>$10B market cap, daily volume >$100M): full size
  Mid-cap target ($2-10B market cap): reduce by 20%
  Small-cap target (<$2B market cap): reduce by 40%, or avoid (can't exit)

Input 4: Correlation to existing portfolio
  If already running 3 campaigns in similar sector: reduce new position by 30%
  Key correlation risk: multiple tech sector campaigns in a tech drawdown

Final size = min(Value gap size, Probability size, Liquidity size) × Correlation adjustment
```

### Position Building Mechanics

**Pre-disclosure phase** (before 5% ownership triggers 13D filing):
- Build position slowly across multiple execution desks (Barclays, Goldman, JPMorgan, DB) to avoid detection
- Use combination of: common stock, call options, total return swaps (TRS on stock)
- TRS allows economic exposure without voting rights (useful for positions above 5% threshold if jurisdiction doesn't count TRS as "beneficial ownership" — US TRS often does count)
- Timeline: 45-90 days for a large position in a liquid stock

**Post-disclosure phase** (after 13D):
- Stock typically jumps 5-15% on Elliott 13D announcement — this is partially a sign of Elliott's credibility/reputation
- Continue buying on any weakness — the announcement itself validated the thesis for many other holders

---

## Module 2: Hedging Activist Positions

### The Four Hedging Layers

**Layer 1: Sector Hedge (Always Required)**
```
Purpose: Protect against sector-wide drawdown that has nothing to do with the campaign
Mechanism: Short the sector ETF or a basket of sector peers
Sizing: Beta-adjusted short equal to ~60-80% of the activist position's market value
Example: Long $500M of XYZ Corp (tech hardware) → Short $350M of SMH (semiconductor ETF)
Cost: 50-150bps/year in carry + short rebate
Benefit: Neutralizes ~70% of sector beta; campaign alpha is "pure"
```

**Layer 2: Market Hedge (Situational)**
```
Purpose: Protect against broad market selloff during campaign (campaigns often take 2+ years)
Mechanism: S&P 500 put options (3-6 month, 10-15% OTM puts)
When to use: When campaign will take >18 months AND market valuations are stretched
Cost: 100-200bps/year depending on VIX environment
When to NOT use: Short campaigns (<12 months) where cost isn't worth it
```

**Layer 3: Campaign-Specific Hedge**
```
Purpose: Protect against specific thesis risks
Mechanism: Varies by campaign type:

Capital return campaign: Long company puts (if campaign fails, stock re-rates down)
  Sizing: Buy 3-month 10% OTM puts covering 25% of position at each campaign milestone

Strategic review campaign: Buy peers as hedge
  If Company A is seeking a merger, Company B (same sector, also undervalued) is a hedge
  If A deal falls through, B may get bid instead

Sovereign/distressed: CDS on the sovereign/issuer as partial hedge
  Long the bonds + short CDS = pure recovery/legal beta hedge
```

**Layer 4: Options Overlay (Monetization)**
```
Purpose: Generate income from optionality + reduce cost basis during long campaigns
Mechanism: Sell covered calls against long stock position
When to use: When campaign is in "waiting" phase (management engaging, not yet at crisis point)
Constraint: Don't sell calls below the campaign's price target — that caps your upside

Example: Long XYZ at $65, campaign target $85
  Sell $90 calls for 18-month duration at $3.50 (5.4% premium)
  If campaign succeeds at $85: keep all gain + call premium
  If campaign fails: break-even at $61.50 instead of $65
```

---

## Module 3: Legal Risk Management

### The Four Categories of Legal Risk in Activism

**1. Regulatory / Antitrust Risk**
```
Risk: Campaign forces a transaction that triggers antitrust review → deal killed
Assessment: Before demanding a sale or merger, model the antitrust landscape
  - HHI concentration in the relevant market post-merger
  - CFIUS risk for technology/defense companies
  - EU/UK merger control for cross-border combinations
Mitigation: Demand spin-offs (no antitrust issue) before M&A (antitrust risk)
```

**2. Securities Law Risk**
```
Risk: Accusations of market manipulation, trading on material non-public information, or 13D/13G filing violations
Assessment: All Elliott campaigns are run through securities counsel (Wachtell, Lipton; or Paul Weiss; or Freshfields)
Key issues:
  - 13D vs. 13G: 13D required if "purpose" is to influence management (almost always for Elliott)
  - HSR filing: Required if position crosses $100M+ threshold in certain securities
  - Group formation: If Elliott coordinated with other investors before 13D, they must file as a "group"
Mitigation: Rigorous legal compliance; file timely and comprehensively
```

**3. Litigation Counter-Risk**
```
Risk: Target company sues Elliott for defamation, tortious interference, or breach of confidentiality
Historical examples: Companies have sought injunctions against Elliott's proxy solicitations
Mitigation:
  - Every public statement reviewed by counsel before release
  - All factual claims in letters verified against public record
  - Never make projections that could be construed as materially misleading
  - Confidentiality agreements honored strictly (don't use NDA-protected info in public campaigns)
```

**4. Regulatory Investigation Risk**
```
Risk: SEC or DOJ investigation into trading patterns surrounding an activist campaign
Historical context: SEC has investigated several activist campaigns for potential front-running (by information shared before 13D filing)
Mitigation:
  - No sharing of position details with anyone outside the deal team before 13D
  - "Wall Street Journal Rule": If you'd be embarrassed to see it on the front page, don't do it
  - Legal firewall between public equity team and credit team at Elliott
```

---

## Module 4: Liquidity Management

### Campaign Liquidity Rules

**Rule 1: Never be forced to sell**
```
The most dangerous situation: Campaign has momentum but market selloff forces liquidation at the worst time
Prevention:
  - Use only "permanent capital" (long-term LP money, not hot money) for activist positions
  - Activist fund has 1-3 year lock-up for LP capital → prevents forced redemption during campaigns
  - 20% of fund kept in liquid positions (could be sold in 30 days) as buffer
```

**Rule 2: Size to the bid**
```
Never build a position larger than what can be liquidated in 20 trading days at 30% of average daily volume
Formula: Max position = 20 × (ADV × 30%) = 6× ADV
Example: Stock trades $50M/day → Max position = $300M (before liquidity becomes a constraint)
For large campaigns ($500M+): Requires $80M+/day average volume → large-cap targets only
```

**Rule 3: Options as liquidity buffer**
```
Using deep-in-the-money calls instead of stock for part of the position:
  - Provides economic exposure with lower cash outlay
  - If campaign fails early, let options expire rather than selling stock at a loss
  - Reduces the "stuck" problem in illiquid positions
Constraint: Options expire — must roll carefully if campaign extends longer than expected
```

---

## Module 5: Drawdown Control in Practice

### Elliott's Drawdown Philosophy

> "We think about the potential for capital loss before we think about potential gain. We don't use a traditional Value at Risk model — we use scenario analysis for the specific risks of each position."

**The Scenario Analysis Template**:
```
For each major campaign position, model three scenarios:

Scenario A — Full Success (campaign achieves all demands):
  Probability: 40%
  Outcome: Full value gap closed + market re-rating
  Return: +40-80% on position

Scenario B — Partial Success (campaign achieves primary demand only):
  Probability: 35%
  Outcome: 50% of value gap closed
  Return: +15-30% on position

Scenario C — Campaign Failure (management rebuffs all demands):
  Probability: 25%
  Outcome: Stock re-rates back to "no activist" level
  Return: -10-20% on position (partially offset by sector hedge)

Expected Return = (40% × 60%) + (35% × 22.5%) + (25% × -15%)
               = 24% + 7.9% - 3.75% = 28.2% expected return

Compare to: Campaign overhead (legal, PR, proxy) = 100-300bps drag
Net expected return = 25-26% gross position return
```

### Portfolio-Level Drawdown Management

**Correlation limits**:
- No more than 30% of portfolio in the same sector simultaneously
- No more than 3 simultaneous campaigns with the same management (some CEOs end up at multiple companies)
- Sovereign debt exposure capped at 15% of total portfolio (duration and political risk concentration)

**Stress test quarterly**:
- What if markets drop 20%? (Sector hedges cover 70% of market beta)
- What if 3 campaigns fail simultaneously? (3 × 3% position × 15% loss = 1.35% portfolio impact — manageable)
- What if a liquidity event forces selling at worst time? (Permanent capital structure prevents this — but stress test anyway)

---

## Module 6: Reputation as a Risk Management Tool

### Why Reputation Is Elliott's Most Valuable Risk Asset

> "Our reputation for following through on our threats is itself a risk management tool. When we say we will run a full proxy fight, management knows we will. That credibility prevents more conflicts than it creates."

**The credibility loop**:
1. Elliott makes a credible threat
2. Management settles to avoid the cost of the fight
3. Settlement validates the threat as credible
4. Next target settles faster, reducing Elliott's cost

**Reputation maintenance rules**:
- Never make a threat Elliott won't follow through on
- Never misstate facts in a public letter (one factual error = permanent credibility damage)
- Honor every commitment made to a target company during settlement
- Never trade on information received under a confidentiality agreement

**The "constructive" framing as risk management**:
Framing every campaign as "helping management" rather than "attacking management":
- Reduces the personal stakes for the CEO (less likely to "fight to the death")
- Maintains relationships with management who may be at different companies later
- Preserves Elliott's ability to do non-adversarial private equity transactions
- Lowers the political/media cost of the campaign

This is not just PR — it is active risk mitigation.
