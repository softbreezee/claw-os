# Gate 3: Layer 3 PM Instructions Review

**Trigger**: Before entering Layer 4.

## Checklist

- [ ] **Sheet Architecture Defined**
  - [ ] Sheet count specified
  - [ ] Each sheet has: purpose, row/column structure, key formulas, reference sources

- [ ] **Revenue Build-up Meets Minimum Granularity**
  - [ ] Hardware: Geography + Channel split
  - [ ] Services: Independent drivers per line (users, pricing, churn)
  - [ ] No single line > 30% of parent without further split
  - [ ] All row sums = report-level revenue

- [ ] **Cost Build-up Meets Minimum Standard**
  - [ ] COGS split: Product vs Services (separate rows)
  - [ ] R&D separate from SG&A (never merged)
  - [ ] Interest Income / Expense separate (never netted)
  - [ ] Tax uses ETR assumption (not fixed dollar amount)

- [ ] **Modeling Instruction Document Output to User** ← MANDATORY
  - See template: `templates/modeling-instruction-doc.md`

- [ ] **User Explicit Confirmation Received**
  - "Confirm" → proceed to Layer 4
  - "Modify" → revise and resubmit
  - "Redo" → return to Layer 3

**Pass**: All granularity rules verified + instruction document output + user confirmed.
**STRICT**: No modeling code may be generated before user confirmation.
