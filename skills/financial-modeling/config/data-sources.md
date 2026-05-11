# Data Sources Priority Protocol

## Overview

All financial data entering the modeling system must follow this priority hierarchy. Lower-priority sources are used only when higher-priority sources are unavailable. Every data point must be audited with source and date.

---

## Priority 1 — Authoritative Primary Sources (最高优先级)

| Data Type | Source | Format | Freshness Requirement |
|-----------|--------|--------|-----------------------|
| Historical financials (IS/BS/CF) | Company SEC filings (10-K, 10-Q), HKEX/CSRC disclosures | Audited PDF/XBRL | Within 6 months of filing |
| Shares outstanding | Cap table (via filing), transfer agent records | Official filing | Latest available |
| Debt schedule | Loan agreements, bond indentures, credit facility docs | Legal docs | As of transaction date |
| Tax rate (ETR) | Company historical filings (3-year avg) | Filed financials | Last 3 fiscal years |
| Capex (actual) | Cash flow statement | Filed financials | LTM basis |

**Audit requirement**: Must cite document name, page number, and filing date for each P1 data point.

---

## Priority 2 — Real-Time Market Data Terminals (实时市场数据)

| Data Type | Source Options | Notes |
|-----------|---------------|-------|
| Market capitalization | Bloomberg, FactSet, Capital IQ, Reuters | Use closing price as of model date |
| Enterprise value | Terminal-calculated (Mkt Cap + Net Debt) | Verify Net Debt against P1 filings |
| Consensus estimates (NTM) | FactSet consensus, Bloomberg BEST, Capital IQ | Min 2 analysts required |
| Peer trading multiples | Same terminals | Snapshot date must match model date |
| Bond prices / credit spreads | Bloomberg, Refinitiv | For WACC debt cost calculation |
| Equity beta | Bloomberg (5yr weekly), FactSet | Use Blume-adjusted beta |

**Audit requirement**: Note terminal name, data pull date, and number of contributing analysts for consensus data.

---

## Priority 3 — Trusted Research & Industry Data (行业研究数据)

| Data Type | Source Options | Notes |
|-----------|---------------|-------|
| Industry growth rates (TAM/SAM) | Gartner, IDC, Forrester, Euromonitor | Cross-validate ≥2 sources |
| Macro assumptions (GDP, inflation) | IMF WEO, World Bank, local central banks | Use latest published edition |
| Regulatory data | Government publications, CSRC/SEC databases | Official only; no secondary sources |
| Comparable transaction multiples | Mergerstat, Bloomberg MA, Capital IQ MA | Note deal date, deal type, conditions |
| Industry benchmark ratios | McKinsey/Bain sector reports, Damodaran.com | Note publication year |

**Audit requirement**: Cite report title, publisher, year, and specific page/table. Never use single-source TAM without independent cross-check.

---

## Priority 4 — Inference & Estimation (推算与估计，最低优先级)

| Scenario | Method | Confidence |
|----------|--------|-----------|
| Missing NTM consensus (<2 analysts) | Use management guidance ± analyst delta from comparable period | Low — flag explicitly |
| Private company revenue | Revenue/employee ratio from public comps × known headcount | Low — ±30% range |
| Historical data unavailable | Industry median from P3 sources | Very Low — flagged with `[INFERRED]` |
| Missing segment breakdown | Proportional allocation from disclosed totals | Low — document methodology |

**Audit requirement**: All P4 data MUST be labeled `[INFERRED]` in the model with methodology note. Never present P4 data without confidence range.

---

## Fallback Decision Tree

```
For each data point:

  Is P1 (audited filing) available?
  ├─ YES → Use P1, cite document + page
  └─ NO → Is P2 (market terminal) available?
           ├─ YES → Use P2, note terminal + date
           └─ NO → Is P3 (research report) available?
                   ├─ YES → Use P3, cite report; cross-validate ≥2 sources
                   └─ NO → Use P4 (inference), label [INFERRED],
                           document methodology, set confidence = Low
```

---

## MCP Connector Integration

When MCP data connectors are enabled (`config/mcp-connectors.json`), Priority 2 data retrieval is automated. The system will:

1. Query enabled connectors in order of connector priority
2. Validate returned data against P1 filing baseline (variance alert if >5%)
3. Cache results with timestamp for reproducibility
4. Fall back to manual entry if all connectors fail

---

## Data Staleness Rules

| Data Type | Maximum Age | Alert Threshold |
|-----------|-------------|-----------------|
| Market prices (EV/multiples) | Model date | >1 trading day → warning |
| NTM consensus | 30 days | >60 days → flag as stale |
| Audited financials | Last fiscal year + 6 months | >18 months → red flag |
| Industry/macro reports | 12 months | >24 months → must revalidate |
| Comparable transactions | 36 months | >60 months → disclose limitations |
