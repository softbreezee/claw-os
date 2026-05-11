---
name: financial-modeling
version: 1.0.0
type: skill
description: >
  Standalone financial modeling engine for investment analysis. Supports public
  and private companies across HF (hedge fund), BT (buyout), and Full tracks.
  Operates independently — does NOT require MAVI. Can optionally receive a
  Variable Lock Sheet from MAVI as a pre-built Layer 2 input, skipping the
  standard 8D diagnostic for assumptions already locked. Produces Excel
  workbooks, JSON data packages, and analysis signal output for backtesting.
author: Investment Intelligence Suite
dependencies: []   # standalone — no upstream skill required
references:
  - tracks/hf-track.md
  - tracks/bt-track.md
  - tracks/full-track.md
  - dcf.md
  - comps.md
  - templates/
  - checklists/
  - config/
  - references/
  - data-contracts/schemas.md
---

# Financial Modeling — Standalone Skill

This skill is the **modeling engine** of the Investment Intelligence Suite.
It is self-contained. It does not call MAVI, does not require market
intelligence upstream, and can be triggered directly by a user who wants to
build a financial model from scratch.

If the user already has a **Variable Lock Sheet** produced by the MAVI skill,
they may supply it and this skill will consume it as a pre-confirmed Layer 2
parameter set, bypassing the standard diagnostic for those variables.

---

## §0 — Pre-Start Check (Always Run First)

Ask the user **three questions** before proceeding. Do not skip any.

```
[PRE-START CHECK]

Q1. Company type?
    (A) Public company   → listed equity; comps available
    (B) Private company  → no listed price; may need implied valuation

Q2. Analysis track?
    (A) HF   — Hedge fund / public equity; thesis-driven; position sizing
    (B) BT   — Buyout / LBO; private equity; leverage & exit
    (C) Full — Complete scope; runs HF + BT frameworks + all quality gates

Q3. Do you have a Variable Lock Sheet from MAVI?
    (A) Yes  → Paste or attach the JSON. Skip 8D for locked variables.
               Proceed directly to §3 Layer Pipeline with L2 pre-filled.
    (B) No   → Proceed with standard 8D Diagnostic (§1) to build assumptions.
```

**Routing logic:**

```
IF Q3 = Yes (MAVI Variable Lock Sheet provided):
  → Load lock sheet (see §9: Interface: Receiving MAVI Output)
  → L2 parameters pre-filled from lock sheet
  → Ask user to confirm or override individual assumptions
  → Proceed to §3 (Layer 1→4 Pipeline) at Layer 2

IF Q3 = No (no MAVI input):
  → Run §1 (8D Diagnostic) to gather L1 research
  → Run §2 (Quick-Build 10Q or full 8D) to generate L2 assumptions
  → Proceed through full §3 pipeline
```

---

## §1 — 8D Diagnostic Framework

The 8D framework is the Layer 1 research protocol. It must be completed (or
partially satisfied via MAVI Variable Lock Sheet) before any model parameters
are locked in Layer 2.

### The Eight Diagnostic Dimensions

| # | Dimension | Core Questions | Minimum Data Required |
|---|-----------|---------------|----------------------|
| **D1** | **Market Size** | TAM/SAM/SOM? Source and date? | Quantified TAM with bottom-up or top-down support. Dated within 18 months. |
| **D2** | **Growth Rate** | Historical CAGR? Projected CAGR? Who is projecting? | 3-year historical CAGR + at least one independent forward estimate |
| **D3** | **Customer Economics** | CAC, LTV, payback period? Churn rate? | B2C: DAU/MAU, ARPU, retention cohort. B2B: contract value, churn, NRR |
| **D4** | **Competitive Position** | Market share? Key competitors? Moat type? | Market share % with source; top 3 competitor names and relative position |
| **D5** | **Regulatory / ESG Risks** | Material risks? Timeline? Cost impact? | At least 2 identified risks with probability and financial impact estimate |
| **D6** | **Technology / Product Roadmap** | Near-term milestones (12-18 months)? | 3+ named upcoming products/features with expected launch dates |
| **D7** | **Management Quality** | Team depth? Relevant experience? Track record? | CEO tenure + prior outcomes; CFO background; key person risk |
| **D8** | **Financial Health** | Net cash position? Leverage ratio? FCF quality? | Net cash = Cash + ST investments − ALL interest-bearing debt (see formula below) |

> ⚠️ **Net Cash Formula (Critical — never deviate):**
> ```
> Net Cash = Cash + Short-term Investments − (Current LT Debt + ST Loans + LT Debt)
>
> ✗ WRONG: Current Assets − Current Liabilities  (overstates by ignoring ST debt)
> ✗ WRONG: Cash only (ignores debt)
> ✓ CORRECT: All cash-like items minus ALL interest-bearing obligations
> ```
> Incorrect net cash inflates IRR 2–3× in entry/exit calculations.

### 8D Output Format

For each dimension, produce:

```
[D#: Dimension Name]
Historical data:  [quantified, sourced]
Market view:      [independent estimate / consensus]
3-Scenario table:
  Bear: [assumption + rationale]
  Base: [assumption + rationale]  ← used in Layer 2 by default
  Bull: [assumption + rationale]
```

### Quick-Build Mode — 10 Default Questions

When the user chooses speed over depth (< 20 min target), bypass full 8D and
collect these 10 inputs. Defaults apply if user presses Enter.

| # | Question | Default | Validation |
|---|----------|---------|-----------|
| 1 | Base year revenue (actual or estimated) | — | Must be > 0 |
| 2 | 5-year average revenue growth rate | 15% | 0% – 50% |
| 3 | Steady-state EBITDA margin | 25% | −20% – 60% |
| 4 | Effective tax rate (ETR) | 25% | 0% – 40% |
| 5 | CapEx as % of revenue (steady-state) | 5% | 0% – 25% |
| 6 | Net working capital as % of revenue | 10% | 0% – 40% |
| 7 | Terminal growth rate (TGR) | **2.5%** | Must be < WACC − 2% |
| 8 | WACC | 8% | 5% – 15% |
| 9 | Equity holding period (years) | 5 | 3 – 7 |
| 10 | Base scenario label | Base | Base / Bull / Bear |

**TGR validation rule (hard stop):**
```
IF TGR > (WACC − 2%):
  ALERT: "Terminal growth rate too high. WACC = X%; recommended TGR ≤ (X−2)%."
  FORCE user confirmation before proceeding.
```

**Why TGR = 2.5% default (not 3.0% or 3.5%):**
- Global long-run GDP growth: 2.3–2.7% (IMF long-range)
- Aging demographics in developed markets compress potential growth
- China growth normalising to 4–5% from prior 10%
- Every +0.5pp TGR at WACC=8% inflates Terminal Value ~10–12%
- Conservative default is the responsible default for investors

### Revenue Granularity Rules

**Hardware:**
```
Revenue = Volume × ASP, split by:
  ├─ Geography (North America / Europe / APAC / Other)
  └─ Channel (if any channel > 30% of hardware revenue)
```

**Services / SaaS:**
```
Each service line must have INDEPENDENT drivers:
  User base (or seats) × Price per unit × (1 − Churn rate)
  ← Cannot aggregate across lines
```

**The 30% Rule:**
> If any single sub-line exceeds 30% of its parent line, it must be broken
> into separate rows with independent assumptions.
> Example: Fashion = 45% of Marketplace → split Women's / Men's.

**Cost granularity (minimum):**
```
Revenue
├─ COGS — Product          [separate row always]
├─ COGS — Services         [separate row always; GM% differs 10–40pp]
├─ Gross Profit
├─ R&D                     [never merge into SG&A]
├─ Sales & Marketing       [separate if > 20% revenue]
├─ G&A
├─ EBIT
├─ Interest Income         [separate row]
├─ Interest Expense        [separate row]
├─ FX Gain/(Loss)          [separate if > 1% of Operating Income]
├─ EBT
└─ Tax = EBT × ETR         [formula, not fixed dollar]
```

---

## §2 — Layer 1→4 Pipeline

The model is built in four sequential layers. Each layer has a defined input,
process, and output. A quality gate separates each transition.

```
┌──────────────────────────────────────────────────────┐
│  LAYER 1: Research & Diagnostics                     │
│  Input:   Company information (filings, data, MAVI)  │
│  Process: 8D Framework or Quick-Build 10Q            │
│  Output:  Diagnostic Report + 3-Scenario Table       │
│  Gate:    Gate 1 — Research Completeness             │
└──────────────────────────┬───────────────────────────┘
                           ↓
┌──────────────────────────────────────────────────────┐
│  LAYER 2: Parameter Lock                             │
│  Input:   L1 Diagnostic OR MAVI Variable Lock Sheet  │
│  Process: Confirm / override each assumption;        │
│           bin scenarios (Bull/Base/Bear);            │
│           generate JSON Confirmation Slip            │
│  Output:  Locked Assumptions JSON                    │
│  Gate:    Gate 2 — Partner Confirmation              │
└──────────────────────────┬───────────────────────────┘
                           ↓
┌──────────────────────────────────────────────────────┐
│  LAYER 3: Modeling Specification                     │
│  Input:   Locked Assumptions JSON                    │
│  Process: Define sheet architecture, row structure,  │
│           formula logic, cross-references, color     │
│           coding; produce Modeling Instruction Doc   │
│  Output:  Modeling Instruction Document (MID)        │
│  Gate:    Gate 3 — PM Instructions (user must        │
│           review MID and confirm before Layer 4)     │
└──────────────────────────┬───────────────────────────┘
                           ↓
┌──────────────────────────────────────────────────────┐
│  LAYER 4: Excel Delivery                             │
│  Input:   Approved MID + Locked Assumptions          │
│  Process: Build Excel workbook per spec; color code; │
│           link three statements; run sensitivities   │
│  Output:  Excel workbook + JSON Data Package         │
│  Gate:    Gate 4 — Delivery Quality                  │
└──────────────────────────────────────────────────────┘
```

### Layer 2 JSON Confirmation Slip Template

```json
{
  "lock_timestamp": "YYYY-MM-DDTHH:MM:SSZ",
  "company": "...",
  "track": "HF | BT | Full",
  "source": "MAVI_variable_lock_sheet | user_direct",
  "locked_assumptions": {
    "revenue_base_year": 0,
    "revenue_cagr_5yr": 0.15,
    "ebitda_margin_terminal": 0.25,
    "tax_rate_etr": 0.25,
    "capex_pct_revenue": 0.05,
    "nwc_pct_revenue": 0.10,
    "terminal_growth_rate": 0.025,
    "wacc": 0.08,
    "hold_period_years": 5
  },
  "scenarios": {
    "bear": { "revenue_cagr": 0.05, "ebitda_margin": 0.18, "rationale": "..." },
    "base": { "revenue_cagr": 0.15, "ebitda_margin": 0.25, "rationale": "..." },
    "bull": { "revenue_cagr": 0.25, "ebitda_margin": 0.30, "rationale": "..." }
  },
  "overrides_from_mavi": []
}
```

If MAVI input was provided, populate `source` with `"MAVI_variable_lock_sheet"`
and list any user-overridden fields in `overrides_from_mavi`.

### Layer 3 Modeling Instruction Document (MID)

The MID must be presented to the user for review **before any Excel is built**.
No modeling begins until the user confirms with "Yes, proceed" or specifies
modifications.

MID sections:
1. Model Overview (sheet count, forecast period, historical period, file name)
2. Sheet Architecture (each sheet: purpose, row count, key formula logic, sources)
3. Assumptions Register (complete table of all inputs with defaults and sources)
4. Formula Logic (written in plain language, not code)
5. Cross-Sheet Reference Map (which cells flow where)
6. Color Coding Rules (see §7)
7. Expected Key Outputs (implied price, upside/downside, IRR range estimate)

---

## §3 — Entry Points & Track Routing

### Entry Point: Public vs Private

Before selecting a track, determine the entry point:

| Condition | Entry Point | Reference |
|-----------|------------|-----------|
| 3+ years of public financial data (SEC filings) | **Public Entry** | `tracks/public-entry.md` (1,124 lines) |
| Private company, pre-IPO, or limited public data | **Private Entry** | `tracks/private-entry.md` (1,104 lines) |

- **Public Entry** (`tracks/public-entry.md`): Automatic routing to HF/BT/Full tracks. Includes full data availability mapping, SEC-anchored diagnostics, and market data integration.
- **Private Entry** (`tracks/private-entry.md`): Confidence-tagged assumptions, Adjusted EBITDA framework, data gap handling, management-provided data validation. All outputs carry confidence tags (HIGH/MED/LOW/INFERRED).

### Track Selection

| Track | When to Use | Module | Sheet Count |
|-------|-------------|--------|-------------|
| **HF** | Public market position; thesis-driven; need EV calc + position sizing | `tracks/hf-track.md` | ~9 sheets (Simple) |
| **BT** | PE/LBO transaction; leverage structure; 100-day plan; exit analysis | `tracks/bt-track.md` | 14–18 sheets |
| **Full** | Deep due diligence; M&A; board decision; investor roadshow; maximum rigor | `tracks/full-track.md` | 18+ sheets |

**Decision rules:**

```
IF company is private AND transaction involves debt financing → BT
IF company is public AND primary goal is position sizing / EV → HF
IF audience is board / IC / lenders AND full 3-statement model required → Full
IF user is unsure → ask: "Is this for a trade or a buyout?"
```

**Track files location:** `tracks/` directory alongside this SKILL.md.
Each track file is self-contained and can be read independently.

---

## §4 — Quality Gates

Gates are sequential checkpoints. A gate failure halts progress until resolved.

### Gate 1 — Research Completeness (after Layer 1)

```
☐ All 8D dimensions populated (or substituted by MAVI lock sheet)
☐ Net cash calculated correctly (Cash + ST Inv − ALL interest-bearing debt)
☐ Each key assumption has: historical data + market view + 3-scenario table
☐ Revenue seasonality and CapEx cycle documented
☐ Leading indicators identified (SaaS: ARR/churn; Hardware: units/ASP)
☐ No single sub-line > 30% of parent without being split

PASS: All items checked with quantitative support and source citations.
```

### Gate 2 — Parameter Lock (after Layer 2)

```
☐ All standard Q's answered by user (no auto-fill without confirmation)
☐ MAVI overrides (if any) acknowledged and confirmed
☐ Scenario binning complete: Bull / Base / Bear assumptions defined
☐ JSON Confirmation Slip generated and shared with user
☐ TGR validated: TGR < (WACC − 2%)
☐ Revenue CAGR ≤ 100% (flag if exceeded)

PASS: JSON slip complete, user has reviewed and confirmed.
```

### Gate 3 — PM Instructions (before Layer 4)

```
☐ Modeling Instruction Document (MID) produced in full
☐ Revenue build-up meets minimum granularity (§1 rules)
☐ Cost structure meets minimum granularity (§1 rules)
☐ MID presented to user; user responded "Yes, proceed" or modifications applied
☐ No modeling code generated before this confirmation

PASS: MID confirmed by user. All granularity checks passed.
```

### Gate 4 — Delivery Quality (before handoff)

```
☐ 10 random non-assumption cells verified to start with "="
☐ Sensitivity test: change one assumption → all downstream cells update
☐ No #REF!, #DIV/0!, #VALUE! errors anywhere
☐ Three statements fully linked (IS → CF → BS)
☐ All Bull/Base/Bear scenarios run independently without cross-contamination
☐ IRR sanity check:
    < 5%  → flag (model error or unattractive deal)
    12–30% → normal
    > 30%  → mandatory assumption review (too optimistic?)
☐ Color coding audit passed (see §7)
☐ Currency symbols in cell format, not embedded in formulas

PASS: All 8 items clear. No open flags.
```

---

## §5 — Modeling Principles (Four Iron Rules)

### Rule 1: No Hard Coding

Every number in the model is either:
- A **user input** (in the Assumptions sheet, blue font + yellow background), or
- A **formula** referencing an input or another formula.

```
✗ WRONG: Year 2 Revenue = 120    [hard-coded]
✓ RIGHT: Year 2 Revenue = =Assumptions!B3 * (1 + Assumptions!B4)

✗ WRONG: Tax = 12.5              [hard-coded]
✓ RIGHT: Tax = =EBT * Assumptions!C15
```

### Rule 2: Color Discipline

| Cell Type | Font Color | Background |
|-----------|-----------|-----------|
| Hard-coded input | Blue `#0000FF` | None |
| Key assumption | Blue `#0000FF` | Light Yellow `#FFFF99` |
| Formula | Black `#000000` | None |
| Cross-sheet reference | Green `#008000` | None |
| Key output (IRR/MoIC) | Black bold | Light Gray `#F2F2F2` |
| Bear scenario | Red `#FF0000` | None |

Full spec: `config/color-spec.md`

### Rule 3: Source Annotation

All inputs must carry a source note:
```
Format: Value | Source | Date
Example: 12.0% | FactSet consensus 5Y historical CAGR | 2026-03-15
Example: 8.5%  | CAPM: 4% Rf + 5% ERP × 0.9 Beta | 2026-03-20
```
Implement as an Excel "Source" column adjacent to each Assumptions row,
or as cell comments.

### Rule 4: Error-Free Delivery

Model must pass all Gate 4 checks before delivery. Zero tolerance for broken
references or disconnected statements. Run Excel "Error Checking" scan before
any handoff.

> See also: `checklists/circular-reference.md` for Circular Breaker Switch
> mechanism (required when interest expense or tax creates circularity).

---

## §6 — Comps Integration

### When to Run Comps

| Trigger | Action |
|---------|--------|
| Public company; peer set > 3 listed comps available | Always run comps alongside DCF |
| LBO with strategic exit (not financial buyer) | Run comps to validate exit multiple |
| User asks for "Football Field" valuation | Mandatory comps + DCF + precedent transactions |
| Quick-Build mode | Optional; user may skip |

### How to Run Comps

Full methodology: `comps.md`

Key steps:
1. **Peer selection** — same business model, similar size (±50% revenue), same
   geography. Minimum 5 peers. See `references/peer-selection-guide.md`.
2. **Metrics to spread** — EV/Revenue, EV/EBITDA, P/E, EV/EBIT, EV/FCF.
   Use NTM (next twelve months) multiples as primary; LTM as secondary.
3. **Median anchoring** — use median (not mean) to reduce outlier distortion.
4. **Implied price** — apply median NTM multiple to company's NTM metric.

### Football Field Cross-Validation

After completing DCF and Comps:

```
Football Field — Valuation Range Summary

Method                      Low      Mid      High
─────────────────────────────────────────────────
DCF (Base WACC ± 1%)        $X       $X       $X
EV/EBITDA Comps (P25–P75)   $X       $X       $X
EV/Revenue Comps            $X       $X       $X
Precedent Transactions      $X       $X       $X
─────────────────────────────────────────────────
Current Price               $X
Implied upside / downside   [%]

Convergence test:
  IF all methods overlap in a ±15% band → High confidence in valuation
  IF DCF outlier vs comps → review WACC or terminal assumptions
  IF comps outlier vs DCF → check for structural peer mismatch
```

---

## §7 — Interface: Receiving MAVI Output

### What is a Variable Lock Sheet?

The MAVI skill produces a **Variable Lock Sheet** — a structured JSON document
containing pre-researched, debated, and locked key model assumptions (revenue
growth, margins, WACC, etc.) validated through MAVI's multi-agent process.

### How to Consume a Variable Lock Sheet

**Step 1: Validate the JSON**

Check that all required fields are present:
```json
{
  "schema_version": "1.0",
  "produced_by": "mavi",
  "lock_timestamp": "...",
  "company": "...",
  "variable_locks": {
    "revenue_cagr_base": 0.12,
    "ebitda_margin_terminal": 0.28,
    "wacc": 0.085,
    "terminal_growth_rate": 0.025,
    ...
  },
  "scenario_range": {
    "bear": { ... },
    "base": { ... },
    "bull": { ... }
  },
  "agent_dissents": [ ... ],
  "confidence_flags": { ... }
}
```

**Step 2: Map to JSON Confirmation Slip**

MAVI Variable Lock → Modeling Skill Layer 2:

| MAVI Field | → | L2 JSON Slip Field |
|---|---|---|
| `variable_locks.revenue_cagr_base` | → | `locked_assumptions.revenue_cagr_5yr` |
| `variable_locks.ebitda_margin_terminal` | → | `locked_assumptions.ebitda_margin_terminal` |
| `variable_locks.wacc` | → | `locked_assumptions.wacc` |
| `variable_locks.terminal_growth_rate` | → | `locked_assumptions.terminal_growth_rate` |
| `scenario_range.*` | → | `scenarios.*` |
| `agent_dissents` | → | Show to user; flag dissents for optional override |

**Step 3: User Confirmation**

Even with a MAVI lock sheet, always show the user the mapped parameters and ask:
```
[MAVI VARIABLE LOCK SHEET LOADED]
The following assumptions are pre-locked by MAVI:
  Revenue CAGR (base): 12.0%
  EBITDA Margin (terminal): 28.0%
  WACC: 8.5%
  TGR: 2.5%

Agent dissents recorded:
  - RC (Risk Challenger) flagged WACC as 50bps too low
  - CA (Conservative Anchor) flagged revenue CAGR as 2pp too optimistic

Do you want to:
  [A] Accept all MAVI parameters as-is
  [B] Accept with modifications (specify which)
  [C] Ignore MAVI and run standard 8D
```

**Step 4: If No MAVI Input**

The user fills the JSON Confirmation Slip directly via §2 process.
No difference in Layer 3 or Layer 4 behavior — the slip format is identical.

---

## §8 — Interface: Standard Analysis Signal Output

After model completion, emit a **Standard Analysis Signal** JSON for
backtesting, tracking, and downstream consumption.

```json
{
  "signal_schema": "investment-intelligence-suite/v1.0",
  "signal_type": "financial_model_output",
  "emitted_at": "YYYY-MM-DDTHH:MM:SSZ",
  "company": "...",
  "track": "HF | BT | Full",
  "model_mode": "Quick-Build | Standard | Deep-Dive",
  "mavi_input_used": true,

  "valuation": {
    "implied_price_base": 0.0,
    "upside_downside_pct": 0.0,
    "irr_base": 0.0,
    "moic_base": 0.0,
    "ev_ebitda_implied": 0.0
  },

  "scenario_outputs": {
    "bear": { "irr": 0.0, "moic": 0.0, "implied_price": 0.0 },
    "base": { "irr": 0.0, "moic": 0.0, "implied_price": 0.0 },
    "bull": { "irr": 0.0, "moic": 0.0, "implied_price": 0.0 }
  },

  "key_assumptions_used": {
    "revenue_cagr_5yr": 0.0,
    "ebitda_margin_terminal": 0.0,
    "wacc": 0.0,
    "tgr": 0.0,
    "hold_period_years": 0
  },

  "quality_gates": {
    "gate_1": "PASS | FAIL",
    "gate_2": "PASS | FAIL",
    "gate_3": "PASS | FAIL",
    "gate_4": "PASS | FAIL"
  },

  "backtest_anchor": {
    "catalyst_date": "YYYY-MM-DD",
    "falsification_metrics": [],
    "review_trigger_date": "YYYY-MM-DD"
  }
}
```

This signal JSON is the handoff point to any backtesting or portfolio
monitoring system. For HF track, `backtest_anchor` fields are critical.
For BT track, `backtest_anchor.catalyst_date` maps to the expected exit date.

---

## §9 — Reference File Index

All supporting files referenced by this skill:

```
financial-modeling/
├── SKILL.md                    ← This file (entry point)
│
├── tracks/
│   ├── public-entry.md         ← Public company entry (1,124 lines)
│   ├── private-entry.md        ← Private company entry (1,104 lines)
│   ├── hf-track.md             ← Hedge Fund track (340 lines)
│   ├── bt-track.md             ← Buyout / LBO track (1,709 lines — full depth)
│   └── full-track.md           ← Full track (181 lines)
│
├── dcf.md                      ← DCF skill (copied from FSM v5.0)
├── comps.md                    ← Comps skill (copied from FSM v5.0)
│
├── templates/
│   ├── json-confirmation-slip.json
│   ├── three-scenario-table.md
│   ├── hf-investment-thesis.md
│   ├── hf-expected-value.md
│   ├── hf-post-mortem.md
│   ├── modeling-instruction-doc.md
│   ├── sources-and-uses.md
│   ├── debt-schedule.md
│   ├── 100-day-plan.md
│   ├── revenue-buildup-saas.md
│   ├── revenue-buildup-marketplace.md
│   ├── revenue-buildup-hardware.md
│   ├── dcf-historical-analysis.md
│   ├── comps-table.md
│   ├── backtest-report.md
│   ├── ic-memo.md
│   ├── investment-brief.md
│   └── one-pager.md
│
├── checklists/
│   ├── gate-1-research.md
│   ├── gate-2-partner.md
│   ├── gate-3-pm-instructions.md
│   ├── gate-4-delivery.md
│   ├── circular-reference.md
│   ├── revenue-granularity.md
│   ├── model-audit.md
│   ├── tam-validation.md
│   └── hf-layer-checklists.md
│
├── config/
│   ├── color-spec.md
│   ├── number-formats.md
│   ├── defaults.json
│   ├── quality-gates.md
│   ├── data-sources.md
│   └── mcp-connectors.json
│
├── references/
│   ├── trap-atlas.md
│   ├── pattern-library.md
│   ├── decision-philosophy.md
│   ├── tam-methodology.md
│   ├── customer-dimension-revenue.md
│   ├── peer-selection-guide.md
│   ├── position-sizing.md
│   ├── excel-engineering-guide.md
│   ├── formula-patterns.md
│   ├── recalc-guide.md
│   ├── tgr-benchmarks.md
│   └── validation-ranges.md
│
└── data-contracts/
    └── schemas.md              ← JSON schema definitions for all outputs
```

---

## Version History

**v1.0.0** (2026-04-16) — Initial standalone release
- Extracted from Full-Spectrum Modeling v5.0
- MAVI dependency removed; MAVI Variable Lock Sheet is now optional input
- Added §0 Pre-Start Check with MAVI/no-MAVI routing
- Added §8 Standard Analysis Signal Output for backtesting integration
- All four Iron Rules and four Quality Gates preserved from FSM core
- Track routing (HF/BT/Full) unchanged from FSM v5.0
