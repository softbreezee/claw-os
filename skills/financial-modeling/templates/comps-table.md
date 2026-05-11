# Comps Table Template

Standard column structure for Comparable Company Analysis tables.

## Sheet: Comps Table

### Column Definitions

| Col | Header | Excel Format | Source | Notes |
|-----|--------|-------------|--------|-------|
| A | Company Name | Text | Manual | Full legal name |
| B | Ticker | Text | Manual | Exchange:Ticker (e.g. NASDAQ:AAPL) |
| C | Tier | 1 / 2 / 3 | Manual | Peer tier classification |
| D | Country | Text | Manual | Headquarter country |
| E | FY End | MMM | Manual | Fiscal year end month |
| F | Market Cap ($M) | `$#,##0` | Live | Latest closing price × diluted shares |
| G | Net Debt ($M) | `$#,##0` | Latest filing | Total debt − cash − ST investments |
| H | Enterprise Value ($M) | `=$F+$G` | Formula | Mkt Cap + Net Debt |
| I | LTM Revenue ($M) | `$#,##0.0` | Latest filing | Last 4 quarters |
| J | NTM Revenue ($M) | `$#,##0.0` | Consensus | Next 12-month FactSet/Bloomberg |
| K | LTM EBITDA ($M) | `$#,##0.0` | Latest filing | Adjusted where disclosed |
| L | NTM EBITDA ($M) | `$#,##0.0` | Consensus | Next 12-month consensus |
| M | LTM Net Income ($M) | `$#,##0.0` | Latest filing | GAAP net income |
| N | NTM EPS ($) | `$#,##0.00` | Consensus | Diluted EPS consensus |
| O | Rev Growth (LTM YoY) | `0.0%` | Formula | `=(I−prior_I)/prior_I` |
| P | Rev Growth (NTM) | `0.0%` | Consensus | Implied by NTM revenue |
| Q | EBITDA Margin (LTM) | `0.0%` | Formula | `=K/I` |
| R | EBITDA Margin (NTM) | `0.0%` | Formula | `=L/J` |
| S | EV/Rev (LTM) | `0.0"x"` | Formula | `=H/I` |
| T | EV/Rev (NTM) | `0.0"x"` | Formula | `=H/J` |
| U | EV/EBITDA (LTM) | `0.0"x"` | Formula | `=H/K` |
| V | EV/EBITDA (NTM) | `0.0"x"` | Formula | `=H/L` |
| W | P/E (NTM) | `0.0"x"` | Formula | `=F/(N×shares)` |
| X–Z | [Industry-specific] | varies | varies | Add per §3.2 of Comps SKILL.md |

### Summary Statistics Rows (append after last company row)

```
Row: Mean (All Peers)     =AVERAGE(S2:S{n})  ...repeat for each col
Row: Median (All Peers)   =MEDIAN(S2:S{n})
Row: P25 (All Peers)      =PERCENTILE(S2:S{n}, 0.25)
Row: P75 (All Peers)      =PERCENTILE(S2:S{n}, 0.75)
Row: Mean (Tier 1 only)   =AVERAGEIF(C2:C{n},1,S2:S{n})
Row: Median (Tier 1 only) array formula: {=MEDIAN(IF(C2:C{n}=1,S2:S{n}))}
```

### Formatting Rules

- Header row: White font `#FFFFFF`, Dark Blue background `#003366`
- Tier 1 rows: No special background
- Tier 2 rows: Light gray background `#F5F5F5`
- Tier 3 rows: Light yellow background `#FFFFEE`
- Outlier cells (>2σ): Orange font `#FF6600` with comment explaining treatment
- Statistics section: Dark divider line above, bold font for Median row
- All multiples: right-aligned, 1 decimal place

### Data Freshness

- Market Cap / EV: Updated as of [DATE] closing price — note in cell comment
- LTM financials: As of [COMPANY FISCAL QUARTER END] — note in cell comment
- NTM consensus: As of [DATE] — source noted (FactSet / Bloomberg / Capital IQ)
