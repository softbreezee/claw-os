# Backtest Methodology

> When to read this file: when executing a backtest, calculating performance metrics, managing portfolio simulation, or debugging backtest results.

---

## Decision Date Generation

### Quarterly (Default)

Generate decision dates on the first trading day of each quarter:
```
Q1: First trading day of January
Q2: First trading day of April
Q3: First trading day of July
Q4: First trading day of October
```

### Semi-Annual

First trading day of January and July.

### Annual

First trading day of January.

### On-Earnings

After each earnings release date for the target companies. Earnings dates can be obtained from:
- SEC EDGAR filing dates (10-Q/10-K)
- Historical earnings calendar data

**Practical note:** When exact historical earnings dates are unavailable, use the end of the fiscal quarter + 45 days (typical 10-Q filing deadline) as a proxy.

---

## Historical Data Collection

### SEC EDGAR (Primary Free Source)

**10-K Annual Reports:**
```
URL: https://www.sec.gov/cgi-bin/browse-edgar?action=getcompany&CIK={ticker}&type=10-K&dateb={YYYYMMDD}&owner=include&count=10
```
The `dateb` parameter restricts to filings before the specified date — critical for time isolation.

**10-Q Quarterly Reports:**
```
URL: https://www.sec.gov/cgi-bin/browse-edgar?action=getcompany&CIK={ticker}&type=10-Q&dateb={YYYYMMDD}&owner=include&count=10
```

**Form 4 (Insider Transactions):**
```
URL: https://www.sec.gov/cgi-bin/browse-edgar?action=getcompany&CIK={ticker}&type=4&dateb={YYYYMMDD}&owner=include&count=40
```

### Historical Price Data

For historical stock prices at specific dates, use:
```
web_fetch → stockanalysis.com/stocks/{ticker}/history/
```
Or Yahoo Finance historical data.

**Important:** Use the closing price on the decision date (or the prior trading day if the decision date falls on a weekend/holiday).

---

## Position Sizing Algorithms

### Universal Position Sizing

Position sizing is determined by the portfolio strategy template selected in the backtest configuration. The backtester applies these rules regardless of which analysis skill is invoked.

```
Position sizing rules (universal):

1. strong_buy signal (confidence ≥ 85%):
   Allocate min(target_allocation_pct, strategy.max_per_position) of portfolio value

2. buy signal (confidence 65-84%):
   Allocate min(target_allocation_pct × 0.7, strategy.max_per_position) of portfolio value

3. hold signal:
   No new allocation. Existing positions held.

4. sell signal:
   Exit position fully.
   
5. strong_sell signal:
   Exit position fully. Flag for do-not-rebuy for 2 analysis cycles.

Override: If the analysis skill provides a target_allocation_pct in the Standard Signal,
use that value (capped by strategy.max_per_position). If not provided, use the defaults above.
```

### Cash Management

Governed by the selected portfolio strategy template:
- **Minimum cash reserve**: Defined per strategy (20% concentrated, 10% diversified, 25% hedged, 15% cyclical)
- **Cash earns risk-free rate**: 3-month T-bill yield for the period
- **Crisis mode**: During market crashes (benchmark drawdown > 20%), strategies with cash reserves can deploy opportunistically

### Rebalancing Rules

```
IF strategy.rebalance_trigger == "on-signal":
    Only adjust when the analysis produces a changed signal
    Unchanged signals → No action

IF strategy.rebalance_trigger == "quarterly" or "monthly":
    At each decision date, recalculate target weights
    If actual weight deviates from target by > 5 percentage points → Rebalance
    If deviation < 5pp → No action (avoid churning)
```

---

## Performance Metric Calculations

### Total Return
```
Total Return = (Final Portfolio Value - Initial Capital) / Initial Capital
```

### CAGR (Compound Annual Growth Rate)
```
Years = (End Date - Start Date) / 365.25
CAGR = (Final Value / Initial Capital)^(1/Years) - 1
```

### Maximum Drawdown
```
Track running maximum portfolio value
At each point: Drawdown = (Current Value - Running Max) / Running Max
Maximum Drawdown = min(all Drawdowns)  — this will be negative
```

### Sharpe Ratio
```
Periodic Returns = [(V_t - V_{t-1}) / V_{t-1} for each period]
Excess Returns = [R - R_f for each period]  (R_f = risk-free rate for the period)
Sharpe = mean(Excess Returns) / std(Excess Returns) × sqrt(periods_per_year)

For quarterly analysis: periods_per_year = 4
For monthly tracking: periods_per_year = 12
```

### Sortino Ratio
```
Same as Sharpe but denominator uses only downside deviation:
Downside Returns = [R for R in Excess Returns if R < 0]
Sortino = mean(Excess Returns) / std(Downside Returns) × sqrt(periods_per_year)
```

### Win Rate
```
Closed Positions = positions that were both entered and exited during the backtest
Winners = positions with positive realized return
Win Rate = Winners / Closed Positions
```

### Alpha (vs Benchmark)
```
Alpha = Portfolio CAGR - Benchmark CAGR
```

For a more sophisticated alpha calculation:
```
Jensen's Alpha = Portfolio Return - [R_f + Beta × (Benchmark Return - R_f)]
```
Where Beta is the regression slope of portfolio returns vs benchmark returns.

---

## Portfolio Simulation Mechanics

### Trade Execution

```
On BUY:
  1. Calculate dollar amount = portfolio_value × target_allocation_pct
  2. Calculate shares = floor(dollar_amount / price)  — no fractional shares
  3. Actual cost = shares × price
  4. Cash -= actual_cost
  5. Record position: {ticker, shares, cost_basis=price, entry_date}

On SELL:
  1. Proceeds = shares × price
  2. Cash += proceeds
  3. Realized gain/loss = proceeds - (shares × cost_basis)
  4. Record: exit_date, exit_price, holding_period, gain/loss

On PARTIAL SELL (confidence-based reduction):
  1. Shares to sell = floor(current_shares × reduction_pct)
  2. Same as full sell but with partial shares
```

### Portfolio Valuation (at each tracking point)

```
Portfolio Value = Cash + sum(shares_i × current_price_i for all positions)
```

Track portfolio value at minimum monthly intervals (even if analysis is quarterly) to compute accurate drawdown and volatility metrics.

### Dividend Handling

```
Option 1 (Default): Dividends reinvested into the paying stock
  → shares += floor(dividend_amount / current_price)

Option 2: Dividends added to cash
  → cash += dividend_amount
```

---

## Benchmark Construction

### S&P 500 (SPY) Benchmark

```
Benchmark portfolio:
  At start_date: Buy SPY with 100% of initial_capital
  Hold throughout entire period
  Dividends reinvested
  Final value = shares × final_SPY_price + accumulated_dividends
```

### Historical SPY prices can be obtained from:
```
web_fetch → stockanalysis.com/etf/spy/history/
```

---

## Decision Quality Attribution

After backtest completion, analyze each decision to identify systematic strengths and weaknesses:

### Per-Decision Analysis

For each buy decision that was later closed:
```
1. What was the five-voice vote? (or L1/L2 score)
2. What was the confidence level?
3. What was the actual return?
4. Which voice's thesis proved most accurate?
5. Which voice's concern proved most prescient?
6. Was there a failure pattern from 04-berkshire-mistakes.md that applied?
```

### Pattern Analysis

Look for systematic patterns:
```
- Does high confidence correlate with better outcomes?
- Does the Munger veto prevent losses? (Compare vetoed vs non-vetoed decisions)
- Does the Jain risk assessment predict drawdowns?
- Are L1 quick scans effective at filtering? (Compare L1 pass rate vs actual outcomes)
- Which industries/sectors perform best under this framework?
- Does the framework work better in bull markets or bear markets?
```

### Framework Calibration

Based on backtest results, suggest adjustments:
```
- If Munger veto catches real risks → System is working; maintain veto power
- If Munger veto blocks winners → Veto threshold may be too sensitive
- If tail risks materialize that Jain missed → Expand tail risk scenarios
- If technology disruptions are missed → Weight Combs/Weschler lens higher
- If valuation discipline is too strict → Missing growth opportunities
- If valuation discipline is too loose → Overpaying systematically
```

---

## Practical Execution Tips

### Token Management

A full L3 backtest across 5 tickers × 20 quarterly decisions = 100 full analyses. This is extremely token-intensive. Practical options:

1. **L1 backtest first**: Run all tickers through L1 to validate the screening function. This is cheap and fast.
2. **L2 for most, L3 for key dates**: Use L2 for routine quarterly decisions, L3 only for dates around major events (earnings surprises, market crashes, management changes).
3. **Sub-agent parallelism**: If the platform supports sub-agents, run multiple ticker analyses in parallel.

### Avoiding Hallucination in Historical Data

When collecting historical data via web scraping:
- **Cross-validate**: If possible, check numbers against two sources
- **Flag discrepancies**: If sources disagree by >5%, note the discrepancy
- **Use conservative estimates**: When uncertain, use the less favorable number

---
