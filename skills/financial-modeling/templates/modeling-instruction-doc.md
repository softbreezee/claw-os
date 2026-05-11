# Modeling Instruction Document Template

> MANDATORY: User must review and approve this document before Layer 4 begins.

```
━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━
MODELING INSTRUCTION — [Company] [Track]
━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━

1. MODEL OVERVIEW
   ├── Sheet Count: __
   ├── Forecast Period: FY____E — FY____E (__years)
   ├── Historical Period: FY____A — FY____A (__years)
   ├── Est. Formula Count: ~___
   └── Output Filename: [company]_[track]_[date].xlsx

2. SHEET ARCHITECTURE
   Sheet 1: [Name]
   ├── Purpose: ...
   ├── Row count / structure: ...
   ├── Key formula logic: ...
   └── Reference sources: ...

   Sheet 2: [Name]
   ├── ... (same structure)

3. ASSUMPTIONS REGISTER
   | # | Assumption | Default | Unit | Source |
   |---|-----------|---------|------|--------|
   | 1 | | | | |
   | 2 | | | | |

4. FORMULA LOGIC (plain language)
   ├── Revenue: [calculation description]
   ├── COGS: [description]
   ├── UFCF: NOPAT + D&A - CapEx - ΔNWC
   ├── Terminal Value: [Gordon Growth / Exit Multiple]
   └── Implied Price: (EV + Net Cash) / Shares

5. CROSS-SHEET REFERENCE MAP
   ├── Assumptions → P&L: [cells]
   ├── P&L → FCF: [cells]
   ├── FCF → DCF: [cells]
   └── ...

6. COLOR CODING (see config/color-spec.md)

7. EXPECTED KEY OUTPUTS
   ├── Implied Share Price: ~$___
   ├── Upside/Downside: ~__%
   └── IRR Range: __% - __%

━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━
Awaiting user confirmation before Layer 4
━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━
```

**User Response Options**:
- Confirm → proceed to Layer 4
- Modify [specific item] → revise and resubmit
- Redo → return to Layer 3
