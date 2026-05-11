# Debt Schedule Template

## Per-Tranche Documentation

For each debt instrument, record:

```
[Tranche Name]
├── Rate Type: Fixed / Floating (index + spread) / PIK / Hybrid
├── Outstanding Amount: $___M
├── Maturity Date: YYYY-MM-DD
├── Annual Interest: $___M (or % of principal)
├── Amortization Schedule:
│   ├── Year 1-2: __% amortization
│   ├── Year 3-5: __% amortization
│   └── Year 6+: Bullet at maturity
├── Collateral: [asset description]
├── Call Protection:
│   ├── No-call period: Year 0-__
│   ├── Make-whole provision: ___bps
│   └── Par call: Year __+
├── Covenant Package:
│   ├── Type: Maintenance / Incurrence
│   ├── Financial: Max Debt/EBITDA = __x, Min ICR = __x
│   ├── Operational: [restrictions]
│   └── Reporting: [frequency]
├── Waterfall Priority: [Rank 1-8]
└── Cross-default Trigger: >$___M default in any tranche
```

## 5-Year Debt Summary

| Metric | Year 1 | Year 2 | Year 3 | Year 4 | Year 5 |
|--------|--------|--------|--------|--------|--------|
| Total Debt (BOP) | | | | | |
| Debt Reduction | | | | | |
| Total Debt (EOP) | | | | | |
| Interest Expense | | | | | |
| Avg Interest Rate | | | | | |
| Debt / EBITDA | | | | | |
| ICR (EBITDA/Interest) | | | | | |
| Covenant Test: Max Leverage | Pass/Fail | | | | |
| Covenant Test: Min ICR | Pass/Fail | | | | |
