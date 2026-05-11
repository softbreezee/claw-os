# Quality Gates

Four mandatory gates that every model must pass before delivery. Each gate is a hard checkpoint — no skipping, no partial passes.

---

## Gate 1 — Layer 1 Research Completeness

All foundational research must be complete before proceeding to partner confirmation.

### Checklist

- [ ] **8D Framework complete** — all eight dimensions researched and documented:
  1. **Market Size** — TAM/SAM/SOM with sources and methodology
  2. **Growth Rate** — historical and projected, with driver decomposition
  3. **Customer Economics** — unit economics, LTV/CAC, retention/churn
  4. **Competitive Position** — market share, moat assessment, Porter's forces
  5. **Regulatory/ESG** — compliance requirements, ESG risks, policy tailwinds/headwinds
  6. **Technology/Product** — product differentiation, tech stack, R&D pipeline
  7. **Management Quality** — track record, incentive alignment, bench depth
  8. **Financial Health** — balance sheet strength, cash flow quality, credit profile

- [ ] **Net cash calculation verified**:
  ```
  Net Cash = Cash + Short-term Investments - All Interest-bearing Debt
  ```
  Must reconcile to latest filing. Include operating leases if material.

- [ ] **Each target question has three components**:
  1. Historical data anchor (minimum 3 years where available)
  2. Industry peer view (at least 3 comparable companies)
  3. Three-scenario table (Bull / Base / Bear with explicit assumptions)

- [ ] **Revenue/Investment cycle diagnosed**:
  - Seasonality pattern identified and quantified
  - CapEx cycle mapped (replacement vs. growth)
  - Working capital pattern documented (DSO, DIO, DPO trends)

- [ ] **Leading indicators identified** — at least 3 forward-looking metrics that predict revenue/margin changes before they appear in financials

### Pass Condition
All items checked. Any gap requires remediation before Gate 2.

---

## Gate 2 — Layer 2 Partner Confirmation

Senior partner must review and lock all key assumptions before model construction begins.

### Checklist

- [ ] **All 4 standard questions answered**:
  1. **Business Model** — How does the company make money? Revenue model classification (subscription, transactional, licensing, hybrid)
  2. **Growth Stage** — Where is the company in its lifecycle? (Early-stage, Growth, Mature, Turnaround)
  3. **Capital Intensity** — Asset-light vs. asset-heavy? CapEx/Revenue ratio benchmark
  4. **Target Audience** — Who is the model for? (Internal IC, external client, board presentation)

- [ ] **All targeted questions confirmed** — each question from Layer 1 reviewed with:
  - Data locked (no further changes without partner re-approval)
  - Scenario binning defined (explicit Bull/Base/Bear ranges for each key variable)

- [ ] **JSON confirmation slip generated** with the following fields:
  ```json
  {
    "lock_timestamp": "2024-XX-XXTXX:XX:XXZ",
    "partner_name": "Name of approving partner",
    "locked_assumptions": ["list of all locked variables"],
    "scenarios_binned": {
      "variable_name": {
        "bull": "value",
        "base": "value",
        "bear": "value"
      }
    }
  }
  ```

### Pass Condition
JSON confirmation slip generated and stored. All assumptions locked.

---

## Gate 3 — Layer 3 PM Instructions

The Modeling Instruction Document must be reviewed and approved by the user before any Excel construction.

### Checklist

- [ ] **MANDATORY: User must review & approve Modeling Instruction Document before Layer 4** — this is a hard gate, no exceptions

- [ ] **Sheet architecture complete** (per Modeling Instruction Document template):
  - All sheets listed with names, purposes, and cross-references
  - Data flow direction documented (left-to-right within sheets, top-to-bottom across sheets)
  - Color coding spec applied (see `config/color-spec.md`)

- [ ] **Revenue build-up meets minimum granularity**:
  - Hardware companies: geographic + channel split
  - Services companies: independent driver decomposition
  - **No single revenue line > 30% of total revenue unsplit** (see `validation_rules.revenue_line_split_threshold` in `config/defaults.json`)

- [ ] **Cost build-up meets minimum standard**:
  - COGS split into product costs and services costs
  - R&D separated from SG&A
  - Interest Income and Interest Expense on separate lines
  - D&A broken out (not buried in COGS or OpEx)

- [ ] **Full modeling instruction output to user** — complete document delivered, not a summary

- [ ] **User explicit confirmation received** — written acknowledgment that they have reviewed and approved the instruction document

### Pass Condition
User has explicitly confirmed approval of the Modeling Instruction Document. All architecture and granularity requirements met.

---

## Gate 4 — Layer 4 Delivery Quality

Final quality assurance before the model is delivered to the end user.

### Checklist

- [ ] **Sample check**: Select 10 non-assumption cells at random — every one must start with `=` (i.e., must be a formula, not a hard-coded value)

- [ ] **Assumption sensitivity test**: Change one key assumption (e.g., revenue growth rate) and verify:
  - All downstream cells auto-update without manual intervention
  - No `#REF!`, `#VALUE!`, `#DIV/0!`, or `#NAME?` errors appear anywhere
  - Results change in the expected direction and magnitude

- [ ] **Currency symbol format check**: All monetary values follow `config/number-formats.md` — no embedded currency symbols in formulas

- [ ] **Sensitivity matrix verification**:
  - WACC x Growth Rate two-way data table present
  - Data table is formula-driven (not hard-coded values)
  - Matrix updates automatically when base case assumptions change

- [ ] **Three financial statements fully linked**:
  - Income Statement → Balance Sheet (retained earnings, tax payable, etc.)
  - Balance Sheet → Cash Flow Statement (change in working capital, D&A, CapEx)
  - Cash Flow Statement → Balance Sheet (ending cash balance)
  - Verify: changing revenue flows through all three statements correctly

- [ ] **All scenarios (Bull/Base/Bear) run independently**:
  - Scenario switch toggles cleanly between all three cases
  - Each scenario produces complete, error-free outputs
  - No cross-contamination between scenarios

- [ ] **IRR reasonability check** (per `config/defaults.json` thresholds):
  - Normal range: 12-20% — no further action needed
  - Above 30%: **forced review** — document why returns are this high, identify which assumption drives it
  - Below 5%: **forced review** — confirm this is not an error, document the thesis for why the deal still makes sense

- [ ] **Circular reference handling** (if applicable):
  - Circular Switch (Circ. Switch) cell present and documented
  - Iterative calculation enabled in Excel settings
  - Model converges (does not oscillate or diverge)
  - Model works correctly with Circ. Switch set to OFF (no errors)

### Pass Condition
All checks pass with no anomaly flags. Model is cleared for delivery.
