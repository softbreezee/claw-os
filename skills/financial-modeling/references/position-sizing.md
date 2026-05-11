# Position Sizing

HF Track position management rules:

## EV-to-Position Mapping

| Expected Value | Asymmetry Ratio | Weight | Type | Max Loss |
|---------------|----------------|--------|------|---------|
| < 2% | < 1.5x | PASS | - | - |
| 2-5% | 1.5-2.0x | 0-1% | Screening | -5% |
| 5-8% | 2.0-3.0x | 1-2% | Tactical | -10% |
| 8-12% | 3.0-4.0x | 2-3% | Conviction | -15% |
| > 12% | > 4.0x | > 3% | Core | -20% |

## Exit Rules

1. Falsification trigger → exit within 24 hours
2. EV drops to 0% → sell 25%; drops to -2% → full exit
3. Catalyst not triggered by deadline → gradual reduction
4. Upside/downside ratio drops below 1.3x → cut 30%
5. Single position daily loss > 1% of portfolio or cumulative floating loss > 3% → cut 50%
