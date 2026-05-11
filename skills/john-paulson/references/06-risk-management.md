# Risk Management for Event-Driven Portfolios

> **When to read this file:** When sizing a merger arb or event-driven position, when designing a hedging strategy against deal break risk, when managing a portfolio of multiple event-driven positions, or when assessing the maximum drawdown risk of a concentrated arb book.

Merger arbitrage and event-driven investing have a distinctive risk profile: most positions make modest, steady returns (capturing spreads), while occasional deal breaks produce large, sudden losses. Managing this asymmetric loss profile — avoiding the tail events that destroy annual P&L — is as important as deal selection. Paulson's risk management is the discipline that converts good deal selection into good portfolio performance.

> "In merger arb, you can be right 90% of the time and still have a terrible year if you're wrong in size on the 10%. Risk management is about making sure the 10% doesn't kill you."

---

## The Asymmetric Loss Profile of Merger Arb

Understanding the fundamental return distribution is the starting point:

**Typical position return profile**:
- Deal closes (P = 85-90%): +4-8% gross return (over 6-12 months)
- Deal breaks (P = 10-15%): -20-40% gross loss (immediate, on deal break news)

**Expected value is positive but variance is severe**:
```
EV = 0.87 × 5% + 0.13 × (-30%) = 4.35% − 3.90% = +0.45% per position

But the standard deviation of this distribution is:
σ = √[0.87×(5%-0.45%)² + 0.13×(-30%-0.45%)²] = ~11%

Sharpe-like ratio: 0.45%/11% ≈ 0.04 per position
```

This illustrates why **diversification and position sizing** are not optional in merger arb — they are the foundation of the strategy. With 20 uncorrelated deals, the portfolio EV stays positive while variance drops dramatically.

---

## Position Sizing Framework

### The Core Sizing Principle

Position size must be calibrated so that **the worst plausible outcome on a single deal** does not cause portfolio-level catastrophe.

**Paulson's sizing rules**:

**Rule 1: Maximum single-position loss constraint**
```
Max loss from any single position ≤ 1-2% of total portfolio value

If deal break causes -30% loss on the position:
  Max position size = 1-2% portfolio loss / 30% loss = 3-6% of portfolio

Example: $1B fund, 1.5% max loss, 30% break downside:
  Max position = $1B × 1.5% / 30% = $5M position = 0.5% of portfolio
```

**Rule 2: Conviction-adjusted sizing**
Start with base size from Rule 1, then adjust:
- 7-Dimension score: 7 Green → 100% of base size
- 6 Green, 1 Yellow → 75% of base size
- 5 Green, 2 Yellow → 50% of base size
- Any Red dimension → 25-35% of base size OR avoid

**Rule 3: Liquidity constraint**
```
Position must be liquidatable within 3-5 trading days at normal market conditions.
Daily volume limit: Position ≤ 10% of target company's average daily trading volume
```

### Maximum Positions and Concentration Limits

```
Maximum portfolio positions: 20-30 active deals (ensure diversification)
Maximum single-deal concentration: 5-8% of portfolio (high conviction, clean deal)
Maximum sector concentration: 25% in one sector (telecom, healthcare, tech)
Maximum regulatory correlation: 30% requiring approval from same regulator
  (e.g., max 30% of portfolio requiring DOJ approval — correlated break risk)
```

---

## Deal Break Risk Quantification

### The Break Scenario Analysis

Every position entry requires modeling the break scenario in detail:

**Step 1: Determine unaffected price**
- Pre-announcement price (most reliable baseline)
- Adjusted for market movement since announcement (if market rallied 10%, unaffected price may also be higher)
- Conservative: use the lower of pre-announcement price and current market equivalent for peers

**Step 2: Estimate break discount**
Not all breaks are equal:
- Cash deal breaks with no reverse break fee: target → unaffected price (full downside)
- Cash deal breaks with large reverse break fee: target → unaffected price + reverse break fee per share
- Stock deal breaks: target → unaffected price; deal's acquirer stock may also fall
- Deal breaks due to improved standalone: target may hold near deal price if standalone value has improved

**Step 3: Time to unaffected price**
- Immediate gap-down: most deal breaks are announced pre-market or at open
- Rarely gradual: when spreads widen significantly before a formal break, partial losses are possible to manage

**Full break scenario model**:
```
Target current price: $47.50
Deal price: $50.00
Unaffected price: $34.00
Reverse break fee: $2.50/share (5%)

Scenario A (deal closes): $50.00 → +$2.50 (+5.3%)
Scenario B (deal breaks, reverse break fee paid): $34.00 + $2.50 = $36.50 → -$11.00 (-23.2%)
Scenario C (deal breaks, no reverse break fee): $34.00 → -$13.50 (-28.4%)

P(A) = 85%, P(B) = 10%, P(C) = 5%
EV = 0.85×5.3% + 0.10×(-23.2%) + 0.05×(-28.4%)
EV = 4.5% - 2.3% - 1.4% = +0.8% gross
```

---

## Options Hedging Strategies

Options provide the most efficient hedge against deal break risk:

### Strategy 1: Protective Puts on Target

**Structure**: Buy put options on target at or below the unaffected price strike

**When to use**: High-conviction deal with one specific tail risk (regulatory ruling)

```
Example:
  Target at $47.50; deal price $50.00; unaffected $34.00
  Buy put at $35 strike expiring at deal close date
  Put premium: $1.50 (3.2% of stock price)
  
If deal closes: lose $1.50 put premium; net gain = $2.50 - $1.50 = $1.00 (2.1%)
If deal breaks: put pays ~$12.50 (from $47.50 to $35.00); net = -$13.50 + $12.50 = -$1.00 (2.1% loss)
```

This converts a potentially -28% position into a defined -2% max loss position.

**Cost**: Protective puts are expensive (high implied volatility in arb situations). Cost must be weighed against risk reduction.

### Strategy 2: Short Acquirer Shares (Stock Deals)

In stock-for-stock deals, standard arb technique:
- Long target shares
- Short acquirer shares (in the ratio specified by the exchange ratio)

This eliminates acquirer stock price risk from the position — the position profits from spread convergence regardless of where acquirer stock trades.

```
Exchange ratio: 0.75 acquirer shares per target share
Hedge: Short 0.75 acquirer shares for each target share owned
Position is now long the "spread" between target and 0.75× acquirer price
Position profits if spread narrows (deal closes) regardless of market direction
```

**Important**: The short acquirer hedge does NOT protect against deal break. It only eliminates the floating value risk.

### Strategy 3: CDS on Acquirer (Financing Risk Hedge)

For PE-sponsored deals or deals with levered acquirers, buy CDS protection on the acquirer:
- If acquirer credit deteriorates → debt markets close → deal financing fails
- CDS on acquirer gains value as acquirer creditworthiness declines
- This hedges the specific risk of "financing market closes before deal closes"

### Strategy 4: Correlation Baskets

For portfolios with multiple deals facing similar regulatory risk (e.g., all tech deals under FTC scrutiny), buy basket-level protection:
- Buy puts on an ETF tracking the sector
- If FTC embarks on sweeping challenge of multiple deals, sector ETF falls, puts profit
- Imperfect hedge but cheap relative to deal-specific protection

---

## Portfolio-Level Risk Management

### Correlation Monitoring

The most dangerous scenario for an arb portfolio is **correlated deal breaks** — multiple deals failing for the same reason simultaneously.

**Sources of correlation**:
1. **Regulatory**: Multiple deals requiring DOJ approval in same sector; if DOJ challenges one aggressively, it signals they'll challenge others
2. **Financing**: In credit-sensitive environments, multiple PE deals may lose financing simultaneously (2008 experience)
3. **Market**: In a severe market downturn, acquirers may invoke MAC clauses across multiple deals
4. **Sector**: Industry-specific shocks (pandemic hits airlines → multiple airline deals break)

**Correlation management rules**:
- Limit exposure to any single regulatory decision-maker (no more than 30% requiring same agency approval)
- Monitor "regulatory regime change" signals — personnel changes at agencies, Congressional pressure, election risk
- Stress test portfolio assuming top 3 deals break simultaneously

### Drawdown Limits and Stop-Loss Disciplines

**Maximum monthly drawdown** before mandatory position review: -3 to -5% on portfolio
**Maximum annual drawdown** before strategy reassessment: -10 to -15%

**Stop-loss triggers at position level**:
- Spread widens beyond 2× initial spread without new deal information → reduce position 50%
- Deal receives DOJ Second Request → recalculate probability, may reduce position
- Acquirer stock falls >20% (for stock deals) → recalculate deal economics, consider exiting

**"Adding to losers" discipline**:
In arb, spreads widening can mean the market is pricing higher break risk — OR it can mean the market is wrong and a buying opportunity exists. Paulson's discipline:
- ADD if: spread widened due to market irrationality, deal fundamentals unchanged, 7-dimension analysis still intact
- DO NOT ADD if: spread widened because new material information emerged (regulatory challenge, MAC concern, etc.)

---

## Leverage and Margin

Merger arbitrage is frequently run with modest leverage (1.5-2× gross for institutional arb funds) because:
- Spreads are small (5-10%) and must be magnified to generate meaningful portfolio returns
- Leverage is appropriate because individual positions are short-duration and cash-settled

**Paulson leverage discipline**:
- Never leverage beyond 2× gross in a single-strategy arb portfolio
- Reduce leverage in high-correlation periods (credit stress, regulatory wave)
- Maintain sufficient liquidity to absorb 2-3 simultaneous deal breaks without forced selling

> "Leverage in merger arb is a tool, not a necessity. We use it modestly, and we reduce it before we need to. The worst time to deleverage is when you're being forced to."

---

## Key Risk Management Metrics (Track Weekly)

```
Portfolio Level:
  Gross exposure:       [$ and %]
  Net exposure:         [$ and %]  
  Leverage ratio:       [x]
  # Active positions:   [count]
  
Regulatory Concentration:
  % requiring DOJ approval:    [%]
  % requiring FTC approval:    [%]
  % requiring China MOFCOM:    [%]
  % requiring EU EC:           [%]
  
Downside Scenarios:
  Max single-deal loss:        [$]
  Correlated break scenario:   [Top 3 deals break → portfolio loss %]
  Market stress scenario:      [If credit markets seize → portfolio loss %]
  
Position Level (per deal):
  Gross spread:                [%]
  Annualized spread:           [%]
  Deal close probability:      [%]
  EV per position:             [%]
  Days to close:               [days]
  Break downside:              [%]
```
