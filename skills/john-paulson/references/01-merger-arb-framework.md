# Merger Arbitrage Framework

> **When to read this file:** When analyzing an announced M&A deal for spread capture, when calculating deal probability and expected value, when scoring a deal across multiple risk dimensions, or when deciding whether to initiate or size a merger arb position.

Merger arbitrage is the oldest and most systematic form of event-driven investing. When a public company is acquired, its stock typically trades at a discount to the announced deal price — the "spread" — because the market is not certain the deal will close. The arbitrageur buys the target and, if cash deal, waits for close. The art is in correctly pricing the risk that the deal does NOT close.

> "Merger arbitrage is not a coin flip. It is a careful estimation of probabilities using all available information. The edge comes from doing that work better than others." — Paulson

---

## The Core Concept: What Creates the Spread

When Deal A is announced at $50 per share and the target trades at $48, the $2 (4%) spread exists because:

1. **Time value**: Even a certain deal takes months to close — investors demand compensation for tying up capital
2. **Deal risk**: Some probability exists that the deal breaks
3. **Opportunity cost**: The capital could be deployed elsewhere
4. **Liquidity premium**: Arb positions are concentrated and hard to unwind quickly

The arbitrageur's task is to determine whether the spread **overcompensates** for these risks (buy) or **undercompensates** (avoid/short the spread).

---

## Spread Calculation: The Mechanics

### Step 1: Identify the Terms

For a **cash deal**:
```
Deal Price = $50.00 (per target share)
Current Target Price = $47.50
Gross Spread = $2.50 (5.26% of current price)
```

For a **stock deal** (acquirer pays 0.75 shares for each target share):
```
Acquirer Current Price = $80.00
Implied Deal Value = 0.75 × $80.00 = $60.00
Current Target Price = $57.00
Gross Spread = $3.00 (5.26%)
Note: stock deals have floating value — spread must be re-marked daily
```

For **mixed consideration** (60% cash, 40% stock):
```
Cash component: 0.60 × $50 = $30.00 (fixed)
Stock component: 0.40 × (0.5 × Acquirer Price) = variable
Total deal value changes with acquirer stock price
```

### Step 2: Determine Downside (Break Price)

The break price is where the target would trade if the deal collapses:
```
Unaffected Price = Pre-announcement price = $35.00 (e.g., 30% below deal)
Gross Downside = Current Price − Unaffected Price = $47.50 − $35.00 = -$12.50
Downside % = -$12.50 / $47.50 = -26.3%
```

The spread/downside **asymmetry** is the central risk parameter. In this example:
- **Upside**: +$2.50 (+5.3%)
- **Downside**: -$12.50 (-26.3%)
- **Asymmetry ratio**: ~5:1 downside to upside

This means you need P(close) > ~83% to break even on EV. The spread is only attractive if you believe the deal closes with high probability.

### Step 3: Annualize the Return

```
Annualized Spread = (Gross Spread / Current Price) × (365 / Days to Close)

Example: 5.26% gross spread × (365 / 180 days) = 10.7% annualized
```

Always compare annualized spread to the **risk-free rate + Paulson spread premium**:
- Risk-free rate (10Y Treasury): 4.5%
- Paulson minimum premium: 300–500 bps
- Minimum hurdle: ~7.5–9.5%

In this example, 10.7% annualized > 9.5% hurdle → potentially attractive, pending deal analysis.

---

## The 7-Dimension Deal Scoring System

Every deal must be scored across all seven dimensions before initiating a position. No dimension can be skipped.

### Dimension 1: Antitrust / Competition Risk

**Green**: No material product or geographic overlap; no prior enforcement in sector; deal size below thresholds
**Yellow**: Moderate overlap in one product market; possible remedies sufficient to cure; sector under normal scrutiny
**Red**: Horizontal overlap creating dominant market share; current enforcement environment hostile; foreign jurisdictions adding complexity

**Key questions:**
- What is the combined market share in the relevant product/geography?
- What is the HHI (Herfindahl-Hirschman Index) pre and post-merger?
- Has FTC or DOJ challenged similar deals in the past 3 years?
- Are EU, China, or other foreign approvals required?
- Is this a vertical deal (potentially less sensitive) or horizontal?

> "Antitrust is the most dangerous dimension because it is hardest to predict and its failure mode is binary — the deal either gets clearance or it doesn't."

### Dimension 2: Financing Risk

**Green**: Deal funded with acquirer cash or committed bank financing with no MAC clause allowing withdrawal
**Yellow**: Bridge financing committed but will be syndicated; some MAC language present; acquirer needs debt markets
**Red**: Financing contingent on market conditions; private equity buyer with LBO financing; significant MAC provisions allowing exit

**Key questions:**
- Is there a "financing out" in the merger agreement?
- Is the financing committed by investment-grade banks, or merely "highly confident" letters?
- What is the debt-to-EBITDA ratio for the combined entity? Can it bear the leverage?
- If debt markets close, does the acquirer still have resources to close?

### Dimension 3: Shareholder Vote

**Green**: No vote required; or vote required but major holders publicly supportive; < 30% shareholder approval threshold
**Yellow**: Vote required; some activist holders with uncertain positions; tight margin for approval
**Red**: Vote required; significant holder opposition (>10% against); activist running competing proxy; ISS recommends against

**Key questions:**
- Does the target need a shareholder vote? What threshold?
- What do the top 10 shareholders (by %) intend to do?
- Is ISS or Glass Lewis opinion expected, and when?
- Has any holder publicly opposed the deal?

### Dimension 4: Closing Conditions / MAC Risk

**Green**: Standard closing conditions; no unusual representations; MAC clause narrowly defined
**Yellow**: Broad MAC language; material adverse change in target's business would allow buyer exit
**Red**: Deal has unusual "outs"; target business has deteriorated since announcement; specific performance metrics as conditions

**Key questions:**
- What specifically constitutes a Material Adverse Change under the merger agreement?
- Has anything happened to the target since announcement that might trigger MAC?
- Are there specific regulatory approvals listed as conditions (beyond standard)?
- Is there a "specific performance" remedy requiring buyer to close even if it wants to walk?

### Dimension 5: Break Fee / Reverse Break Fee

**Green**: Reverse break fee (buyer's penalty for walking) is large (>4% of deal value); buyer has no regulatory out
**Yellow**: Standard reverse break fee (~3%); regulatory out could allow buyer to walk without paying
**Red**: No reverse break fee; small break fee (<2%); "regulatory out" allowing buyer exit without penalty

**Key questions:**
- What is the break fee (target pays if it walks from the deal)?
- What is the reverse break fee (acquirer pays if it fails to close)?
- Is there a "hell-or-high-water" clause requiring buyer to accept any antitrust remedy?
- Under what conditions does the reverse break fee substitute for specific performance?

**Note**: A large reverse break fee effectively puts a floor under the target price if the deal breaks — it changes the downside calculation.

### Dimension 6: Strategic Rationale

**Green**: Clear industrial logic; deal announced at reasonable premium; both boards enthusiastic; no obvious better alternative
**Yellow**: Some strategic rationale but synergies are aggressive; premium is very high (>40%)
**Red**: Unclear strategic rationale; deal appears defensive; acquirer made impulsive public announcement; synergies strain credibility

**Key questions:**
- Why does this deal make strategic sense for the acquirer?
- Is the acquirer's management team under pressure to "do something"?
- Is the premium reasonable relative to comparable transactions?
- Would the acquirer prefer to walk away if it legally could?

> "If you can't explain why this deal makes sense for the acquirer in two sentences, that's a warning sign."

### Dimension 7: Interloper / Competing Bid Risk

**Green**: No obvious competing bidder; target company not broadly shopped; strategic premium already high
**Yellow**: A few potential competing bidders; target was shopped to multiple parties; fiduciary out allows consideration of other offers
**Red**: Multiple strategic buyers identified; deal announced with "go-shop" period; target in auction process

**Key questions:**
- Were other potential buyers approached before this deal was signed?
- Does the merger agreement have a "go-shop" period (allowing active solicitation of others)?
- Is there a "fiduciary out" allowing the target board to accept a superior proposal?
- What happens to the spread if a competing bid emerges? (Usually spreads to the new higher price)

**Note**: Interloper risk is usually POSITIVE for the arbitrageur — a higher bid means higher returns. But it can also extend timeline and introduce regulatory risk of a new buyer.

---

## The Expected Value Formula

This is the core calculation that drives every merger arb decision:

```
EV = P(close) × Gross Spread + P(break) × Gross Downside

Where:
  P(close) + P(break) = 100%
  Gross Spread = Deal Price − Current Price (positive)
  Gross Downside = Unaffected Price − Current Price (negative)
```

**Example calculation:**
```
Deal Price: $50.00
Current Price: $47.50
Gross Spread: +$2.50 (+5.26%)
Unaffected Price: $35.00
Gross Downside: -$12.50 (-26.3%)

P(close) = 85%
P(break) = 15%

EV = 0.85 × $2.50 + 0.15 × (-$12.50)
EV = $2.125 + (-$1.875)
EV = +$0.25 per share (+0.53% on $47.50)

Annualized EV = 0.53% × (365/180) = 1.1% annualized

Verdict: Does NOT clear hurdle → Do not initiate position
```

**Sensitivity analysis** — always run the EV at different P(close) assumptions:
```
P(close) = 90%: EV = $2.25 − $1.25 = +$1.00 → +4.3% annualized ← potentially attractive
P(close) = 85%: EV = +$0.25 → +1.1% annualized ← unattractive
P(close) = 80%: EV = $2.00 − $2.50 = -$0.50 → negative EV ← avoid
```

The sensitivity analysis tells you the **break-even P(close)** — the probability at which EV turns positive. If you believe actual P(close) > break-even, the trade has merit.

---

## Determining Deal Close Probability

P(close) is not arbitrary — it should be anchored to observable evidence:

**Base rate starting point**: Historical merger completion rates by category:
- All-cash deals, no regulatory issues: ~90–95%
- All-cash deals with moderate regulatory issues: ~75–85%
- Stock deals: ~80–88% (lower because acquirer share price risk adds complexity)
- PE-sponsored LBOs with financing: ~70–80%
- Deals requiring antitrust remedies: ~65–80% depending on complexity

**Adjustments from base rate:**
- Each "Red" dimension: −5 to −15 percentage points
- Each "Green" dimension: +2 to +5 percentage points
- Comparable precedent deal completion rates: anchor heavily to the closest historical precedent
- Market signals (acquirer stock price, CDS spreads on acquirer, options skew on target): incorporate implied market probabilities

**Final P(close)** = Base rate ± dimension adjustments, informed by precedent and market signals.

---

## Deal Selection Criteria: What Paulson Looks For

Not all deals are worth analyzing deeply. Apply these filters first:

1. **Spread size**: Gross spread > 3% for cash deals (below this, not enough compensation)
2. **Annualized return potential**: > risk-free + 300 bps after probability adjustment
3. **Understandable regulatory risk**: Can you bound it using precedent?
4. **Committed financing**: Equity-funded or committed bank financing strongly preferred
5. **Time to close**: < 18 months (longer timelines introduce too much macro uncertainty)
6. **Size**: Target market cap > $1B (sufficient liquidity for position sizing)
7. **Agreement quality**: Merger agreement reviewed for MACs, break fees, specific performance provisions

> "We probably look at 50 deals for every 1 we put on. Most deals don't clear the filter. The ones that do, we size appropriately. Discipline in deal selection is half the battle."
