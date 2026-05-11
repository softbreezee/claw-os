# Validation Ranges

Complete parameter validation ranges:

| Parameter | Valid Range | Default | Alert Condition |
|-----------|-----------|---------|----------------|
| Revenue CAGR (5yr) | 0% - 100% | 15% | > 100% → confirm no M&A |
| EBITDA Margin | -20% - 60% | 25% | Outside range → flag unrealistic |
| Tax Rate (ETR) | 0% - 40% | 25% | > statutory rate → investigate |
| CapEx % Revenue | 0% - 25% | 5% | > 25% → major capex program? |
| NWC % Revenue | 0% - 40% | 10% | Negative → confirm business model |
| WACC | 5% - 15% | 8% | Outside range → review inputs |
| TGR | 0% - (WACC-2%) | 2.5% | > WACC-2% → stability violation |
| TV / EV Ratio | - | - | > 70% → extend explicit forecast |
| IRR | - | - | < 5% or > 30% → mandatory review |
| SOM / SAM | - | - | > 30% → limited growth space |
| TAM CAGR vs Industry | - | - | Variance > 5pp → explain |
| TAM-derived Revenue vs Model | - | - | Variance > 15% → reconcile |
| Revenue line concentration | - | - | Single line > 30% → must split |
| Supplier concentration | - | - | > 50% → add "loss" scenario |
