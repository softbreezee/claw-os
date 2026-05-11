# Gate 4: Layer 4 Delivery Quality

**Trigger**: After model build, before delivery.

## Checklist

- [ ] **Formula Audit**: 10 random non-assumption cells start with "="
- [ ] **Assumption Sensitivity Test**: Change one assumption → all downstream cells update, no #REF! or #DIV/0!
- [ ] **Currency Symbol Check**: Symbols in format codes, not embedded in formulas
- [ ] **Sensitivity Matrix Verification**: WACC × Growth Rate table — values change when parameters change, not hard-coded
- [ ] **Three-Statement Linkage**: IS → BS → CF fully connected
  - Net Income flows to CF (Operating Activities)
  - Net Income flows to BS (Retained Earnings)
  - No manual inputs or breaks
- [ ] **Scenario Independence**: Bull/Base/Bear each have own assumptions; modifying one doesn't affect others; all three show different IRR/MoIC
- [ ] **IRR Reasonability**:
  - 12-20%: Normal
  - > 30%: Forced review (assumptions may be too optimistic)
  - < 5%: Review (possible model error or unattractive investment)
- [ ] **Circular Reference Management** (if applicable):
  - [ ] Circ. Switch in Row 1, Col A
  - [ ] Iterative calculation enabled (Max iterations: 100, Max change: 0.001)
  - [ ] Converges without oscillation
  - [ ] Model works with switch OFF (approximate values)
  - [ ] Clear business logic documented for each circular dependency
  - [ ] No unintended circular references

**Pass**: All 11 checks pass, no anomaly flags.
