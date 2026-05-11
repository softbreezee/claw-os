# TGR Benchmarks

Terminal Growth Rate reference data:

## Why 2.5% (not 3.5%)

The old industry standard TGR of 3.5% has been corrected to 2.5% based on:

| Region/Market | Long-term Growth Range | Source |
|--------------|----------------------|--------|
| Global GDP | 2.3% - 2.7% | IMF long-term projection |
| Developed nations | 2.0% - 3.0% | OECD average |
| China (new normal) | 4.0% - 5.0% | Slowed from 10% historical |
| US | 2.0% - 2.5% | CBO long-term forecast |
| Europe | 1.5% - 2.0% | Aging demographics |

## Impact of TGR Error (Gordon Growth Model)

```
TV = FCF_terminal / (WACC - TGR)

Example: FCF = $100M, WACC = 8%

TGR = 3.5%: TV = $100M / 4.5% = $2,222M
TGR = 2.5%: TV = $100M / 5.5% = $1,818M

Difference = $404M (18% overvaluation)
Each +0.5pp TGR → typically +2-3pp IRR
```

## Validation

TGR must be < (WACC - 2%). If sensitivity to TGR is steep (>3pp IRR per 0.5pp TGR), the model depends too heavily on terminal value — extend explicit forecast period.
