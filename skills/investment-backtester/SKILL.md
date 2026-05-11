---
name: investment-backtester
description: |
  Universal backtesting engine for any investment analysis skill. Simulates historical investment decisions by invoking any compatible analysis skill at historical time points, then tracks portfolio performance. Works with any skill that outputs the Standard Analysis Signal format — berkshire-hathaway suite, munger, buffett, howard-marks, or any future investment skill. Trigger scenarios: any request to backtest an investment strategy, test historical performance, validate an investment framework against real outcomes, compare two investment approaches side-by-side, or run a historical simulation. This skill is the scheduling and evaluation layer — it does not contain investment knowledge itself.
---

# Investment Backtester

A universal backtesting engine that invokes **any investment analysis skill** at historical time points to simulate decisions and measure portfolio performance.

> **Compatible with any skill that outputs the Standard Analysis Signal.** This includes berkshire-hathaway, munger, buffett, and any future investment skill.
> **This skill is the scheduling and evaluation layer** — it does not contain investment knowledge itself.

> **Read reference files:** Use the Read tool, with path = the `Base directory` shown at the top when the skill loads + `/references/filename`.

---

## What This Skill Does

```
Input:  List of tickers + date range + analysis skill(s) to use + portfolio strategy
Output: Simulated portfolio performance + decision log + performance attribution
```

1. At each decision point, collects historical data available as of that date
2. Invokes the specified analysis skill in BACKTEST MODE
3. Collects Standard Analysis Signals from the skill
4. Simulates portfolio changes based on the configured portfolio strategy
5. Tracks performance over time
6. Compares against benchmark (S&P 500 or user-specified)
7. Generates performance report with attribution analysis
8. (Optional) Compares two skills side-by-side on the same data

---

## Standard Analysis Signal Protocol

**Any investment analysis skill can be backtested if it outputs this format.**

```json
{
  "ticker": "AAPL",
  "date": "2024-01-15",
  "signal": "strong_buy | buy | hold | sell | strong_sell",
  "confidence": 82,
  "target_allocation_pct": 15.0,
  "exit_trigger": "Description of what would cause an exit",
  "recheck_date": "2024-04-15",
  "source_skill": "berkshire-hathaway",
  "reasoning_summary": "One-sentence rationale"
}
```

**Required fields:** `ticker`, `date`, `signal`, `confidence`
**Optional fields:** `target_allocation_pct`, `exit_trigger`, `recheck_date`, `source_skill`, `reasoning_summary`

**Signal mapping for different skills:**

| Skill | Skill's Native Output | → Standard Signal |
|-------|----------------------|-------------------|
| **berkshire-hathaway** | Strong Buy / Buy / Watch / Pass / Strong Pass | strong_buy / buy / hold / sell / strong_sell |
| **munger** | Buy / Don't Buy / Keep Watching / Hold / Sell | buy / strong_sell / hold / hold / sell |
| **buffett** | Buy / Don't Buy / Keep Watching / Hold / Sell | buy / strong_sell / hold / hold / sell |
| **Any future skill** | Must map to the 5 standard signals | — |

When invoking an analysis skill in BACKTEST MODE, append this instruction to the analysis prompt:

```
[BACKTEST OUTPUT REQUIRED]
After your analysis, output a Standard Analysis Signal in this exact JSON format:
{"ticker":"...","date":"...","signal":"...","confidence":...,"target_allocation_pct":...,"exit_trigger":"...","recheck_date":"...","source_skill":"...","reasoning_summary":"..."}
```

---

## Backtest Configuration

### Required Parameters

```
tickers:           List of stock tickers to analyze
start_date:        Backtest start date (YYYY-MM-DD)
end_date:          Backtest end date (YYYY-MM-DD)
analysis_skill:    Which skill to invoke for analysis decisions
```

### Optional Parameters

```
initial_capital:     Starting portfolio value in USD (default: $1,000,000)

analysis_frequency:  How often to run analysis
                     - "quarterly" (default)
                     - "semi-annual"
                     - "annual"
                     - "monthly"

analysis_depth:      Tier to use (if the skill supports tiered analysis)
                     - "L1" — Quick scan
                     - "L2" — Intermediate
                     - "L3" — Full depth (default)

portfolio_strategy:  Predefined portfolio management style (see below)
                     - "concentrated" (default)
                     - "diversified"
                     - "hedged"
                     - "cyclical"
                     - "custom"

max_position_pct:    Maximum allocation per position (default varies by strategy)

benchmark:           Benchmark ticker for comparison (default: "SPY")

compare_skill:       Second skill for side-by-side comparison (optional)
```

---

## Portfolio Strategy Templates

Different investment philosophies demand different portfolio management rules. Select the strategy that matches the analysis skill being used.

### Strategy: "concentrated" (Default — Buffett/Munger/Berkshire Style)

```
Max positions:        5-8
Max per position:     25% of portfolio
Min cash reserve:     20% of portfolio
Rebalance trigger:    On signal change only (not calendar-based)
Position sizing:      Confidence-weighted (higher confidence → larger position)
Sell discipline:      Only on explicit sell signal or exit trigger
Turnover target:      <30% annually
```

Best paired with: `berkshire-hathaway`, `buffett`, `munger`

### Strategy: "diversified" (Lynch/Fisher Style)

```
Max positions:        15-25
Max per position:     8% of portfolio
Min cash reserve:     10% of portfolio
Rebalance trigger:    Quarterly calendar rebalance
Position sizing:      Equal-weight with confidence tilt (+/-2%)
Sell discipline:      Sell on explicit signal OR when thesis no longer applies
Turnover target:      <50% annually
```

Best paired with: `peter-lynch`, `phil-fisher`, future growth-oriented skills

### Strategy: "hedged" (Marks/Taleb Style)

```
Max positions:        10-15 long, up to 5 short/hedge positions
Max per position:     10% of portfolio (long), 5% (short)
Min cash reserve:     25% of portfolio (dry powder for distressed opportunities)
Rebalance trigger:    On signal change + quarterly risk review
Position sizing:      Risk-adjusted (smaller positions for higher-risk names)
Sell discipline:      Sell when price exceeds intrinsic value (no holding forever)
Turnover target:      <60% annually
Special rule:         Increase cash reserve when market-wide overvaluation detected
```

Best paired with: `howard-marks`, `nassim-taleb`, credit/distressed investing skills

### Strategy: "cyclical" (Druckenmiller/Dalio Style)

```
Max positions:        5-10 (concentrated macro bets)
Max per position:     30% of portfolio
Min cash reserve:     15% of portfolio
Rebalance trigger:    Monthly (macro conditions can shift fast)
Position sizing:      Conviction-weighted (top idea gets largest weight)
Sell discipline:      Sell on regime change signal or loss exceeds -15%
Turnover target:      <100% annually (higher turnover acceptable for macro)
Special rule:         Can go to 50%+ cash if no high-conviction ideas
```

Best paired with: `stanley-druckenmiller`, `ray-dalio`, macro-oriented skills

### Strategy: "custom"

User specifies all parameters manually:
```
max_positions, max_per_position, min_cash_reserve,
rebalance_trigger, position_sizing_method,
sell_discipline, turnover_target, special_rules
```

---

## Execution Protocol

### Step 1: Initialize

```python
portfolio = {
    "cash": initial_capital,
    "positions": {},          # ticker → {shares, cost_basis, entry_date, entry_signal}
    "history": [],            # chronological list of all actions
    "performance": [],        # periodic portfolio value snapshots
}

decision_dates = generate_decision_dates(start_date, end_date, analysis_frequency)
strategy = load_strategy(portfolio_strategy)
```

### Step 2: For Each Decision Date

```
For each date in decision_dates:

  2a. COLLECT HISTORICAL DATA
      Read references/01-backtest-methodology.md for detailed data collection protocol.
      
      Key principle: Data package MUST NOT contain information from after the decision date.
      Use date-restricted queries (SEC EDGAR dateb parameter, historical price lookups).

  2b. INVOKE ANALYSIS SKILL IN BACKTEST MODE
      Construct prompt:
      
      "[BACKTEST MODE: Analysis date = {date}. Use ONLY the provided data below. 
       Do not use any knowledge of events, prices, or outcomes after {date}.
       You are making this decision on {date} with only the information available then.]
       
       Analyze {ticker} using the [{analysis_skill}] framework.
       
       Data Package:
       {historical_data_package}
       
       Current price as of {date}: ${price}
       
       [BACKTEST OUTPUT REQUIRED]
       After your analysis, output a Standard Analysis Signal JSON."

  2c. PARSE STANDARD ANALYSIS SIGNAL
      Extract: signal, confidence, target_allocation_pct, exit_trigger

  2d. EXECUTE PORTFOLIO ACTION (per strategy rules)
      Apply the selected portfolio_strategy's rules for:
      - Position sizing
      - Cash reserve enforcement
      - Rebalance triggers
      - Sell discipline

  2e. RECORD STATE
      Log portfolio value, positions, cash balance, decision rationale
```

### Step 3: Calculate Performance Metrics

See `references/01-backtest-methodology.md` for detailed calculation formulas.

---

## Skill Comparison Mode

When `compare_skill` is specified, run two parallel portfolios:

```
Portfolio A: Uses analysis_skill (e.g., "berkshire-hathaway")
Portfolio B: Uses compare_skill (e.g., "munger")

Both portfolios:
  - Start with same initial_capital
  - Analyze same tickers on same dates
  - Use same portfolio_strategy (for fair comparison)
  - Compare against same benchmark

Output includes:
  - Side-by-side performance chart
  - Decision agreement rate (how often both skills agree)
  - Divergence analysis (cases where they disagreed — who was right?)
  - Combined alpha (what if you only invested when both agreed?)
```

### Comparison Output Format

```
## Skill Comparison: [Skill A] vs [Skill B]

### Performance Summary
| Metric | [Skill A] | [Skill B] | Benchmark |
|--------|-----------|-----------|-----------|
| CAGR | XX.X% | XX.X% | XX.X% |
| Sharpe | X.XX | X.XX | — |
| Max DD | -XX.X% | -XX.X% | -XX.X% |
| Win Rate | XX% | XX% | — |

### Decision Agreement
- Agreed on XX% of decisions
- When both said Buy: avg return = +XX%
- When they disagreed: [Skill A] was right XX% of the time

### Notable Divergences
| Date | Ticker | [Skill A] | [Skill B] | Actual Outcome | Right Call |
|------|--------|-----------|-----------|----------------|-----------|
```

---

## Standard Output Format

```
# Investment Backtest Report

## Configuration
- Analysis Skill: [skill name]
- Portfolio Strategy: [strategy name]
- Tickers: [list]
- Period: [start] to [end] ([X years])
- Analysis Depth: [L1/L2/L3]
- Frequency: [quarterly/semi-annual/annual/monthly]
- Initial Capital: $X,XXX,XXX

## Executive Summary
- Portfolio Final Value: $X,XXX,XXX
- Total Return: XX.X% vs Benchmark XX.X%
- CAGR: XX.X% vs Benchmark XX.X%
- Alpha: +/-XX.X%
- Max Drawdown: -XX.X%
- Sharpe Ratio: X.XX

## Equity Curve
[ASCII chart or description of portfolio value vs benchmark over time]

## Decision Log
| Date | Ticker | Signal | Confidence | Action | Price | Outcome |
|------|--------|--------|------------|--------|-------|---------|
| [date] | [ticker] | [signal] | [XX%] | [Buy/Sell/Hold] | $XX | [+/-XX%] |

## Position Attribution
| Ticker | Entry | Exit | Holding | Return | Contribution |
|--------|-------|------|---------|--------|--------------|
| [ticker] | $XX @ [date] | $XX @ [date] | X mo | +/-XX% | +/-$XX,XXX |

## Decision Quality Analysis

### What Worked
[Patterns in successful decisions — which signal characteristics predicted winners?]

### What Didn't Work
[Patterns in failed decisions — what went wrong and why?]

### Skill Calibration
- Does higher confidence correlate with better outcomes? [Yes/No, r=X.XX]
- Average return on strong_buy signals: +/-XX%
- Average return on buy signals: +/-XX%
- Average return on hold signals: +/-XX%
- False positive rate (buy signal → negative return): XX%
- False negative rate (sell signal → stock went up): XX%

### Suggested Improvements
[Based on backtest results, what adjustments to the analysis framework would help?]

## Caveats
- LLM time isolation is imperfect — model training data may leak future knowledge
- Data availability may differ from what was actually available historically  
- Transaction costs and slippage not modeled
- Tax effects not modeled
- This is a simulation for educational purposes, not investment advice
```

---

## Time Isolation Protocol

The most critical (and most difficult) aspect of LLM-based backtesting.

### Five Layers of Protection

1. **Explicit prompt restriction**: `[BACKTEST MODE: Analysis date = YYYY-MM-DD]`
2. **Data-only analysis**: Provide specific data package, instruct model to use only that data
3. **Blind ticker option**: Replace tickers with anonymous labels (Company A, B, C) to reduce recall of specific historical events
4. **Confidence discounting**: Apply 10-15% systematic reduction to all confidence scores
5. **Contamination audit**: After each decision, check if the model references events it shouldn't know about

### Known Limitations

- Major events (2008 crisis, COVID, specific scandals) are nearly impossible to "forget"
- Well-known companies (AAPL, TSLA, AMZN) have highest contamination risk
- Older analysis dates have more contamination (more hindsight in training data)
- **Recommendation**: For cleanest results, use less-covered mid-cap companies and recent time periods

---

## Reference Files

| File | Contents |
|------|----------|
| `references/01-backtest-methodology.md` | Detailed methodology: decision date generation, historical data collection with time isolation, position sizing algorithms for each strategy template, portfolio simulation mechanics, performance metric calculations, benchmark construction, decision quality attribution |
