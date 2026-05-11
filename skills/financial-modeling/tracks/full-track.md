# Full Track — Complete Scope Modeling

> Loaded when Pre-Start Check Q2 = Full.
> The Full track runs **all frameworks from HF and BT tracks** plus
> the complete Core pipeline with every quality gate active.
> Use for: deep due diligence, M&A, board IC memos, investor roadshows,
> lender presentations, and any analysis where rigor > speed.

---

## When to Use Full Track

| Trigger | Rationale |
|---------|-----------|
| M&A due diligence | Buyer needs complete 3-statement model + all scenario analysis |
| IC / board presentation | Decision-makers require maximum analytical rigor |
| Lender / credit committee | Debt covenants require full cash flow and leverage modeling |
| Investor roadshow | External audience; no tolerance for model gaps |
| Dual track (IPO + M&A) | Both HF and BT frameworks must run in parallel |
| First-time modeling on a company | Full track establishes baseline; subsequent updates use HF or BT |

---

## Full Track = Run Everything

The Full track does not define its own unique frameworks. It is an **execution
protocol**: run all relevant modules from Core + HF + BT, with all four
quality gates active and all granularity rules at maximum depth.

```
Full Track Execution Order:
  1. Pre-Start Check (§0 of SKILL.md)
  2. 8D Diagnostic (full, not Quick-Build) [§1]
  3. Layer 1→4 Pipeline [§2]
  4. HF frameworks (if public company):
     ├─ Investment thesis (Layer 0)
     ├─ Consensus divergence map
     ├─ Falsification thresholds
     ├─ Expected value calculation
     └─ Position sizing (if applicable)
  5. BT frameworks (if transaction involves leverage):
     ├─ Management investability assessment
     ├─ PE-specific risk matrix
     ├─ Complex capital structure + Sources & Uses
     ├─ LBO return decomposition
     ├─ 100-day plan
     ├─ Exit strategy matrix
     └─ Recovery analysis
  6. DCF (full)                 → dcf.md
  7. Comps (mandatory)          → comps.md
  8. Football Field             → §6 of SKILL.md
  9. All four quality gates     → checklists/gate-[1-4]-*.md
 10. JSON Data Package + Signal → §8 of SKILL.md
```

---

## Sheet Architecture — Full Track (18+ Sheets)

The exact sheet count varies by business model and track combination.
Minimum required sheets for Full track:

| # | Sheet Name | Purpose |
|---|-----------|---------|
| 1 | Cover | Model metadata, company overview, key outputs summary |
| 2 | Assumptions | All inputs (blue/yellow cells); scenario toggle |
| 3 | Revenue Build-up | Detailed revenue by segment, geography, channel |
| 4 | P&L Forecast | Full income statement (historical + 5–8 year forecast) |
| 5 | Balance Sheet | Full BS forecast (linked to P&L and CF) |
| 6 | Cash Flow Statement | Direct or indirect method; FCF derivation |
| 7 | DCF Valuation | UFCF → TV → EV → Equity value → Implied price |
| 8 | Comps | Peer trading multiples; implied valuation range |
| 9 | Football Field | Visual summary: all methods on one chart |
| 10 | Scenarios | Bull / Base / Bear side-by-side summary |
| 11 | Sensitivity | WACC × growth (2D); additional key variable sensitivities |
| 12 | CapEx & D&A | Detailed CapEx schedule; PP&E roll-forward; D&A bridge |
| 13 | Working Capital | DSO / DIO / DPO analysis; NWC % revenue trend |
| 14 | Debt Schedule | All debt layers; amortization; interest; covenant tests |
| 15 | LBO / Returns | IRR / MoIC decomposition (BT only) |
| 16 | Sources & Uses | Transaction financing structure (BT only) |
| 17 | 100-Day Plan | Initiative tracker; Year 1 EBITDA build (BT only) |
| 18 | Covenants | 12-quarter headroom trajectory; pinch points (BT only) |
| 19+ | Additional | Revenue by cohort; customer dimension; TAM build; audit log |

---

## Quality Gates — All Active

Full track runs all four gates without exception. No shortcuts.

| Gate | Trigger | Key Addition vs HF/BT |
|------|---------|----------------------|
| **Gate 1** | After Layer 1 research | All 8D dimensions; no Quick-Build substitution |
| **Gate 2** | After Layer 2 lock | Full scenario binning; JSON slip complete; MAVI integration if used |
| **Gate 3** | Before Layer 4 | Complete MID with 18+ sheet architecture reviewed and confirmed by user |
| **Gate 4** | Before delivery | Full 3-statement link verified; all scenarios; IRR sanity; color audit |

Checklist files:
- `checklists/gate-1-research.md`
- `checklists/gate-2-partner.md`
- `checklists/gate-3-pm-instructions.md`
- `checklists/gate-4-delivery.md`

---

## Granularity — Maximum Depth

Full track enforces the deepest granularity tier from Core:

| Dimension | Full Track Minimum |
|-----------|-------------------|
| Revenue | By cohort / segment + channel + geography |
| Cost | Full P&L module per business line; COGS split; R&D vs SG&A; interest split |
| CapEx | By asset category (leasehold / equipment / software) |
| NWC | By account (AR / inventory / AP / accruals) |
| Tax | ETR assumption (not fixed $); deferred tax if material |
| Scenarios | Three full independent scenarios; no shared rows between scenarios |

The 30% rule applies at maximum strictness: any sub-line > 30% of parent
**must** be broken out, regardless of model complexity it creates.

---

## Deliverables

Full track produces all three output types:

| # | Deliverable | Format | Contents |
|---|-----------|--------|---------|
| 1 | **Excel Workbook** | `.xlsx` | 18+ sheet linked model; all scenarios; all sensitivities |
| 2 | **JSON Data Package** | `.json` | All assumptions, results, gate statuses — see `data-contracts/schemas.md` |
| 3 | **Analysis Signal** | `.json` | Standard signal output for backtesting — see §8 of SKILL.md |
| 4 | **Modeling Instruction Doc** | `.md` / `.pdf` | Layer 3 MID; serves as audit trail for all modeling decisions |

---

## Template and Reference Index (Full Track)

All templates apply. Priority order for Full track:

```
templates/
  ├─ modeling-instruction-doc.md   ← Layer 3 MID template
  ├─ three-scenario-table.md       ← Scenario binning
  ├─ sources-and-uses.md           ← BT capital structure
  ├─ debt-schedule.md              ← Debt amortization
  ├─ 100-day-plan.md               ← BT operating plan
  ├─ hf-investment-thesis.md       ← HF thesis structure
  ├─ hf-expected-value.md          ← EV calculation
  ├─ comps-table.md                ← Peer analysis
  ├─ dcf-historical-analysis.md    ← Historical calibration
  ├─ backtest-report.md            ← Post-investment review
  └─ ic-memo.md                    ← Full IC presentation format

references/
  ├─ trap-atlas.md                 ← Ten modeling traps; detection + fixes
  ├─ pattern-library.md            ← Eight recurring signal patterns
  ├─ decision-philosophy.md        ← Precision vs accuracy; trust the model?
  ├─ tam-methodology.md            ← TAM five-layer bottom-up framework
  ├─ customer-dimension-revenue.md ← Multi-supplier / distributor modeling
  ├─ excel-engineering-guide.md    ← Advanced Excel structure guidance
  └─ validation-ranges.md          ← Acceptable ranges for all key assumptions
```

---

## Time Estimates

| Mode | Wall-Clock Time | Output |
|------|----------------|--------|
| Quick-Build (not recommended for Full) | 20 min | 6 sheets; limited rigor |
| Standard (HF or BT only) | 60 min | 11 sheets |
| **Full Track** | **2–4 hours** | **18+ sheets; all gates; all frameworks** |
| Deep Dive (Full + comps + add-ons) | Half day | 20+ sheets; precedent transactions; sensitivity pages |

Full track is not a speed tool. If the user needs results in < 60 minutes,
redirect to HF or BT track with Standard mode.

---

*Full Track v1.0 | Investment Intelligence Suite | References FSM v5.0 integrated-modeling-core + hf + bt*
