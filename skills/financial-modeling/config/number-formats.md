# Excel Number Format Standard

## Format Codes

| Data Type | Format Code | Example | Notes |
|-----------|-----------|---------|-------|
| Revenue / Costs | `$#,##0.0` | `$1,234.5M` | One decimal (millions); use `$#,##0` for thousands |
| Percentages | `0.0%` | `15.0%`, `2.5%` | One decimal; percent sign required |
| Multiples | `0.00"x"` | `2.50x`, `8.75x` | Two decimals; "x" suffix |
| Share Price | `$#,##0.00` | `$125.50` | Two decimals (cent precision) |
| Integer Counts | `#,##0` | `1,500,000` | No decimals |
| Interest Rates (WACC) | `0.0%` | `8.0%`, `12.5%` | One decimal |
| Years | `0` | `5`, `10` | No decimals |

## Currency Symbol Handling

**Principle**: Currency symbols in independent cells, not embedded in number cells.

```
Correct:
  A1: "$" (currency header)
  A2: 100.5 (formatted as $#,##0.0 → displays as $100.5)

Wrong:
  A2: ="$" & 100.5  (embeds symbol in formula; breaks data portability)
```

## Context Examples

```
Sheet: "P&L Forecast"
  Year 1 Revenue:        $1,234.5M    [Revenue format]
  EBITDA Margin:         32.5%        [Percentage format]

Sheet: "Valuation"
  Entry EV/EBITDA:       12.50x       [Multiples format]
  Exit Price per Share:  $42.75       [Share price format]
  Equity Multiple:       2.50x        [Multiples format]
```
