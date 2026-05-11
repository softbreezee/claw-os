# Excel Color Encoding Standard

Investment banking standard color scheme for financial models.

## Color Matrix

| Cell Type | Font Color | Background | Hex Codes | Meaning |
|-----------|-----------|------------|-----------|---------|
| Hard-coded Input | Blue | None | Font: `#0000FF` | Analyst manual input, no formula |
| Key Assumption | Blue | Light Yellow | Font: `#0000FF`, BG: `#FFFF99` | Most sensitive assumptions, requires partner review |
| Formula | Black | None | Font: `#000000` | Auto-calculated, do not manually edit |
| Cross-sheet Reference | Green | None | Font: `#008000` | References another tab's data |
| Key Output | Black Bold | Light Gray | Font: `#000000` bold, BG: `#F2F2F2` | Key return metrics (IRR, MoIC) |
| Bear Scenario | Red | None | Font: `#FF0000` | Stress scenario / risk identifier |
| Header Row | White | Dark Blue | Font: `#FFFFFF`, BG: `#003366` | Sheet structure headers |
| Subtotal | Black Bold | Light Tan | Font: `#000000` bold, BG: `#FFFFCC` | Sub-totals (distinct from key outputs) |

## Enforcement Rules

1. **No mixing**: Key assumptions must have BOTH blue font AND yellow background. Yellow background alone is insufficient.
2. **Consistency audit**: Before delivery, run a color audit. Any cell that "looks like an input but has wrong color" must be corrected.
3. **Tool support**: Templates include built-in Conditional Formatting rules for automatic coloring of new data.

## Application Example

```
Sheet: "Assumptions"
  (Blue, no BG)     2024 Revenue           [100.5]  ← User input
  (Blue, yellow BG) 5-year CAGR            [15.0%]  ← Key assumption
  (Black, none)     Terminal Growth Rate    [2.5%]   ← Formula: =WACC-2%
  (Green, none)     WACC                   [=CoE!B5] ← Cross-sheet ref
  (Black bold, gray) IRR (Base Case)       [18.3%]  ← Key output
  (Red, none)       IRR (Bear Case)        [8.2%]   ← Stress scenario
```
