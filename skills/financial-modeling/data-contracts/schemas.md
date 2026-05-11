# Data Contracts: Inter-Skill JSON Schemas

## Purpose
When skills call each other, they must pass structured data — not natural language descriptions. These schemas are the single source of truth for input/output interfaces.

## 1. WACC Skill → DCF Skill

```json
{
  "wacc": 0.088,
  "cost_of_equity": 0.105,
  "cost_of_debt_pretax": 0.055,
  "cost_of_debt_after_tax": 0.042,
  "tax_rate": 0.21,
  "debt_weight": 0.30,
  "equity_weight": 0.70,
  "risk_free_rate": 0.043,
  "equity_risk_premium": 0.055,
  "beta_type": "adjusted_blume",
  "beta_value": 1.12,
  "country_risk_premium": 0.00,
  "size_premium": 0.00,
  "currency": "USD",
  "methodology_note": "CAPM with Blume-adjusted beta, US 10Y Treasury as Rf"
}
```

## 2. Terminal Value Skill → DCF Skill

```json
{
  "method": "dual",
  "perpetuity_growth": {
    "terminal_growth_rate": 0.025,
    "terminal_fcf": 398000000,
    "terminal_value": 6320000000,
    "pv_terminal_value": 2598000000,
    "validation": {
      "tgr_vs_wacc_spread": 0.063,
      "tgr_vs_nominal_gdp": "below",
      "gordon_growth_stable": true
    }
  },
  "exit_multiple": {
    "terminal_ebitda": 640000000,
    "scenarios": {
      "conservative": {"multiple": 12.0, "tv": 7680000000, "pv_tv": 3157000000},
      "base": {"multiple": 14.0, "tv": 8960000000, "pv_tv": 3683000000},
      "upside": {"multiple": 16.0, "tv": 10240000000, "pv_tv": 4209000000}
    },
    "multiple_source": "Precedent transactions median: 13.5x, peer trading median: 15.2x"
  }
}
```

## 3. Sensitivity Skill → DCF Skill

```json
{
  "tables": {
    "wacc_vs_tgr": {
      "row_variable": "wacc",
      "col_variable": "terminal_growth_rate",
      "row_values": [0.078, 0.083, 0.088, 0.093, 0.098],
      "col_values": [0.015, 0.020, 0.025, 0.030, 0.035],
      "values": [[2.80, 3.15, 3.52, 3.95, 4.50], [3.02, 3.42, 3.88, 4.42, 5.08], [3.28, 3.72, 4.28, 4.96, 5.82], [3.58, 4.08, 4.75, 5.62, 6.78], [3.95, 4.52, 5.32, 6.48, 8.16]],
      "base_row": 2,
      "base_col": 2
    },
    "growth_vs_margin": {
      "row_variable": "revenue_growth_y1",
      "col_variable": "terminal_ebitda_margin",
      "row_values": [0.06, 0.08, 0.10, 0.12],
      "col_values": [0.28, 0.30, 0.32, 0.34],
      "values": [[28.4, 31.2, 34.1, 37.2], [32.8, 35.9, 39.2, 42.8], [37.4, 40.9, 44.6, 48.7], [42.2, 46.2, 50.6, 55.4]],
      "base_row": 1,
      "base_col": 1
    }
  },
  "tornado": {
    "ranked_assumptions": [
      {"variable": "terminal_growth_rate", "base": 0.025, "low": 0.015, "high": 0.035, "value_low": 2.80, "value_high": 4.50, "impact_range": 1.70},
      {"variable": "wacc", "base": 0.088, "low": 0.078, "high": 0.098, "value_low": 4.20, "value_high": 3.12, "impact_range": 1.08},
      {"variable": "terminal_ebitda_margin", "base": 0.200, "low": 0.180, "high": 0.220, "value_low": 3.65, "value_high": 4.78, "impact_range": 1.13}
    ]
  },
  "tv_ev_ratio": 0.65,
  "tv_ev_warning": false
}
```

## 4. Integrated Skill Layer 3 → DCF Skill (Engine Mode Input)

```json
{
  "mode": "engine",
  "company": "Sea Limited",
  "ticker": "SE",
  "currency": "USD",
  "forecast_years": 10,
  "historical": {
    "years": ["FY2022", "FY2023", "FY2024", "FY2025E"],
    "revenue": [12450, 13060, 14200, 15800],
    "ebitda": [980, 1520, 2100, 2500],
    "ebitda_margin": [0.079, 0.116, 0.148, 0.158],
    "da": [450, 480, 510, 540],
    "capex": [620, 580, 550, 600],
    "nwc_change": [-120, -80, -90, -100],
    "tax_rate": [0.18, 0.19, 0.20, 0.21]
  },
  "assumptions": {
    "revenue_growth": [0.12, 0.11, 0.10, 0.09, 0.08, 0.07, 0.06, 0.05, 0.04, 0.03],
    "ebitda_margin": [0.165, 0.175, 0.185, 0.190, 0.195, 0.200, 0.200, 0.200, 0.200, 0.200],
    "da_pct_revenue": [0.034, 0.033, 0.032, 0.031, 0.030, 0.030, 0.030, 0.030, 0.030, 0.030],
    "capex_pct_revenue": [0.038, 0.036, 0.034, 0.032, 0.030, 0.030, 0.030, 0.030, 0.030, 0.030],
    "nwc_pct_revenue_change": [0.02, 0.02, 0.02, 0.02, 0.02, 0.02, 0.02, 0.02, 0.02, 0.02],
    "tax_rate": 0.21,
    "terminal_growth_rate": 0.025
  },
  "equity_bridge": {
    "cash": 6800,
    "total_debt": 3200,
    "minority_interest": 150,
    "preferred_stock": 0,
    "diluted_shares": 580
  },
  "scenarios": {
    "bull": {"probability": 0.25, "overrides": {"revenue_growth": [0.14, 0.13, 0.12, 0.11, 0.10, 0.08, 0.07, 0.06, 0.05, 0.04], "terminal_growth_rate": 0.030}},
    "base": {"probability": 0.50},
    "bear": {"probability": 0.25, "overrides": {"revenue_growth": [0.10, 0.09, 0.08, 0.07, 0.06, 0.05, 0.04, 0.03, 0.02, 0.01], "terminal_growth_rate": 0.020}}
  }
}
```

## 5. DCF Skill → Integrated Skill (Engine Mode Output)

```json
{
  "valuation": {
    "enterprise_value": {"bear": 28500, "base": 35200, "bull": 44800},
    "equity_value": {"bear": 31950, "base": 39650, "bull": 48250},
    "per_share": {"bear": 55.09, "base": 68.36, "bull": 83.19}
  },
  "explicit_period_pv": 12800,
  "terminal_value_pv": {"perpetuity": 22400, "exit_multiple_base": 24500},
  "tv_ev_ratio": 0.65,
  "wacc_used": 0.105,
  "ufcf_by_year": [1250, 1380, 1520, 1650, 1770, 1880, 1960, 2030, 2090, 2140],
  "sensitivity_tables": {
    "wacc_vs_tgr": "ref to schema 3",
    "growth_vs_margin": "ref to schema 3"
  },
  "tornado": "ref to schema 3",
  "quality_checks": {
    "tgr_stable": true,
    "tv_ev_below_75": true,
    "all_formulas": true,
    "fcf_growth_pattern": "reasonable",
    "no_negative_fcf": true
  }
}
```

## 6. Model Backtest Input (NEW — Tobi Standard)

```json
{
  "company": "Sea Limited",
  "model_date": "2026-03-20",
  "actual_results": {
    "period": "FY2026",
    "revenue": 17200,
    "ebitda": 3100,
    "ebitda_margin": 0.180,
    "fcf": 1800,
    "capex": 580
  }
}
```

## 7. Confidence Level Schema (Private Company)

```json
{
  "data_point": "Revenue FY2024",
  "value": 850000000,
  "confidence": 3,
  "confidence_label": "Medium - comparable inference",
  "source": "Inferred from [Comparable A] revenue/employee ratio",
  "sensitivity_range": {"low": 680000000, "high": 1020000000, "range_pct": 0.20}
}
```

## Schema Validation Rules

- All numeric values in basis points or percentages must use decimal notation (0.025 = 2.5%).
- Currency codes must be ISO 4217 3-letter codes (USD, EUR, CNY, etc.).
- Dates must be ISO 8601 format (YYYY-MM-DD).
- Arrays of assumptions must match forecast_years in length.
- Enterprise_value and equity_value must include base, bear, and bull scenarios.
- Terminal value inputs must validate: terminal_growth_rate < (WACC - margin of safety).

## Schema Versioning

Current version: 1.0

All skills must declare which schema version they support in their SKILL.md header using:

```yaml
schema_version: "1.0"
```

Breaking changes trigger version bumps (1.0 → 2.0). Additive changes (new optional fields) do not bump version. Deprecated fields are marked `_deprecated: true` during migration windows.

---

## 8. Comps → Valuation Summary (v4.0 NEW)

Output of the Comps module, consumed by the Valuation Summary sheet and any downstream reporting.

```json
{
  "schema_version": "2.0",
  "schema_id": "comps_valuation_summary",
  "generated_at": "2026-04-12T10:30:00Z",
  "target_company": "Sea Limited",
  "ticker": "SE",
  "analysis_date": "2026-04-12",

  "peer_group": {
    "total_peers": 8,
    "tier_breakdown": { "tier_1": 3, "tier_2": 3, "tier_3": 2 },
    "peers": [
      { "name": "Shopify", "ticker": "SHOP", "tier": 1, "country": "CA", "market_cap_usd_m": 98000 },
      { "name": "MercadoLibre", "ticker": "MELI", "tier": 1, "country": "AR", "market_cap_usd_m": 92000 },
      { "name": "Global-E Online", "ticker": "GLBE", "tier": 1, "country": "IL", "market_cap_usd_m": 6800 },
      { "name": "Grab Holdings", "ticker": "GRAB", "tier": 2, "country": "SG", "market_cap_usd_m": 12000 },
      { "name": "Coupang", "ticker": "CPNG", "tier": 2, "country": "KR", "market_cap_usd_m": 28000 },
      { "name": "PDD Holdings", "ticker": "PDD", "tier": 2, "country": "CN", "market_cap_usd_m": 142000 },
      { "name": "Amazon", "ticker": "AMZN", "tier": 3, "country": "US", "market_cap_usd_m": 2100000 },
      { "name": "Alibaba", "ticker": "BABA", "tier": 3, "country": "CN", "market_cap_usd_m": 220000 }
    ],
    "outliers_excluded": [
      { "ticker": "AMZN", "metric": "EV/EBITDA", "value": 42.5, "reason": ">2σ from peer mean; scale differential", "treatment": "excluded_from_statistics" }
    ]
  },

  "multiples": {
    "as_of_date": "2026-04-12",
    "metrics": {
      "ev_ltm_revenue": {
        "all_peers": [13.8, 4.8, 9.4, 4.4, 5.2, 3.1, 3.4, 2.1],
        "tier_1_only": [13.8, 4.8, 9.4]
      },
      "ev_ntm_revenue": {
        "all_peers": [11.2, 4.1, 8.2, 3.8, 4.4, 2.7, 3.0, 1.8],
        "tier_1_only": [11.2, 4.1, 8.2]
      },
      "ev_ltm_ebitda": {
        "all_peers": [28.5, 12.4, null, 31.0, 15.8, 9.2, null, 8.8],
        "_note": "null = negative EBITDA, excluded from stats"
      },
      "ev_ntm_ebitda": {
        "all_peers": [22.0, 10.8, 52.0, 18.5, 11.2, 7.8, null, 7.1],
        "tier_1_only": [22.0, 10.8, 52.0]
      }
    }
  },

  "statistics": {
    "ev_ltm_revenue": {
      "all_peers": { "mean": 6.6, "median": 4.8, "p25": 3.6, "p75": 9.4, "std_dev": 3.9 },
      "tier_1_only": { "mean": 9.3, "median": 9.4, "p25": 7.1, "p75": 11.6 }
    },
    "ev_ntm_revenue": {
      "all_peers": { "mean": 5.4, "median": 4.1, "p25": 3.0, "p75": 8.2, "std_dev": 3.2 },
      "tier_1_only": { "mean": 7.8, "median": 8.2, "p25": 6.2, "p75": 9.7 }
    },
    "ev_ntm_ebitda": {
      "all_peers_excl_outliers": { "mean": 12.9, "median": 11.0, "p25": 8.5, "p75": 20.3 },
      "tier_1_only": { "mean": 28.3, "median": 22.0, "p25": 16.4, "p75": 37.0 }
    }
  },

  "implied_valuation": {
    "target_financials": {
      "ltm_revenue_usd_m": 14200,
      "ntm_revenue_usd_m": 15800,
      "ltm_ebitda_usd_m": 2100,
      "ntm_ebitda_usd_m": 2500,
      "net_debt_usd_m": -5000,
      "diluted_shares_m": 570
    },
    "by_metric": {
      "ev_ntm_revenue": {
        "multiple_low": 3.0, "multiple_mid": 4.1, "multiple_high": 8.2,
        "implied_ev_low": 47400, "implied_ev_mid": 64780, "implied_ev_high": 129560,
        "implied_equity_low": 52400, "implied_equity_mid": 69780, "implied_equity_high": 134560,
        "implied_price_low": 91.93, "implied_price_mid": 122.42, "implied_price_high": 236.07
      },
      "ev_ntm_ebitda": {
        "multiple_low": 8.5, "multiple_mid": 11.0, "multiple_high": 20.3,
        "implied_ev_low": 21250, "implied_ev_mid": 27500, "implied_ev_high": 50750,
        "implied_equity_low": 26250, "implied_equity_mid": 32500, "implied_equity_high": 55750,
        "implied_price_low": 46.05, "implied_price_mid": 57.02, "implied_price_high": 97.81
      }
    },
    "weighted_average": {
      "weights": { "ev_ntm_revenue": 0.40, "ev_ntm_ebitda": 0.60 },
      "implied_price_low": 72.69,
      "implied_price_mid": 83.19,
      "implied_price_high": 152.80,
      "current_price": 68.50,
      "upside_mid": 0.215
    },
    "dcf_cross_validation": {
      "dcf_implied_price": 74.50,
      "comps_implied_price_mid": 83.19,
      "deviation_pct": 0.117,
      "status": "WITHIN_15PCT",
      "assessment": "DCF and Comps broadly consistent. Minor premium in Comps may reflect market growth optimism."
    }
  }
}
```

---

## 9. Data Layer → Core (v4.0 NEW)

Standard output from the Data Layer module, consumed by Core for model initialization.

```json
{
  "schema_version": "2.0",
  "schema_id": "data_layer_to_core",
  "generated_at": "2026-04-12T09:15:00Z",
  "entity": {
    "name": "Sea Limited",
    "ticker": "SE",
    "exchange": "NYSE",
    "gics_sector": "Consumer Discretionary",
    "gics_industry": "Internet & Direct Marketing Retail",
    "fiscal_year_end": "December",
    "currency": "USD",
    "reporting_unit": "millions"
  },
  "historical_financials": {
    "years": ["FY2022", "FY2023", "FY2024"],
    "revenue": [12450, 13060, 14200],
    "revenue_growth": [null, 0.049, 0.087],
    "ebitda": [980, 1520, 2100],
    "ebitda_margin": [0.079, 0.116, 0.148],
    "ebit": [530, 1040, 1590],
    "da": [450, 480, 510],
    "net_income": [320, 720, 1180],
    "capex": [620, 580, 550],
    "capex_pct_revenue": [0.050, 0.044, 0.039],
    "nwc_change": [-120, -80, -90],
    "nwc_pct_revenue": [null, null, 0.063],
    "cash_and_equivalents": [6800, 7200, 7600],
    "short_term_investments": [1200, 900, 800],
    "total_interest_bearing_debt": [3200, 2900, 2600],
    "net_debt": [-4800, -5200, -5800],
    "diluted_shares": [580, 575, 570],
    "tax_rate_etr": [0.18, 0.19, 0.20],
    "source": "10-K FY2024 (SEC EDGAR, filed 2025-02-15)",
    "confidence": "HIGH"
  },
  "market_data": {
    "as_of_date": "2026-04-12",
    "stock_price": 68.50,
    "shares_outstanding_diluted": 570,
    "market_cap": 39055,
    "enterprise_value": 34055,
    "ev_calc_note": "EV = Mkt Cap ($39,055M) + Net Debt (-$5,800M) + Minority Interest ($800M)",
    "beta_raw": 1.32,
    "beta_blume_adjusted": 1.21,
    "beta_source": "Bloomberg 5yr weekly, Blume adjustment (2/3 raw + 1/3 × 1.0)",
    "week_52_high": 82.30,
    "week_52_low": 41.20,
    "consensus_estimates": {
      "ntm_revenue": 15800,
      "ntm_ebitda": 2500,
      "ntm_ebitda_margin": 0.158,
      "ntm_eps_diluted": 2.45,
      "revenue_growth_ntm": 0.113,
      "analyst_count": 24,
      "source": "FactSet consensus as of 2026-04-11",
      "confidence": "HIGH"
    }
  },
  "debt_breakdown": {
    "total_interest_bearing_debt": 2600,
    "items": [
      {
        "instrument": "4.75% Senior Notes",
        "amount": 2000,
        "currency": "USD",
        "maturity_date": "2028-03-15",
        "rate_type": "fixed",
        "coupon_rate": 0.0475,
        "source": "10-K FY2024, Note 8"
      },
      {
        "instrument": "Term Loan Facility",
        "amount": 600,
        "currency": "USD",
        "maturity_date": "2027-06-30",
        "rate_type": "floating",
        "spread_bps": 200,
        "reference_rate": "SOFR",
        "source": "10-K FY2024, Note 8"
      }
    ],
    "maturity_profile_usd_m": {
      "2026": 0,
      "2027": 600,
      "2028": 2000,
      "2029_and_beyond": 0
    },
    "fixed_rate_pct": 0.769,
    "floating_rate_pct": 0.231,
    "weighted_avg_cost_of_debt_pretax": 0.051,
    "confidence": "HIGH"
  },
  "peer_data": {
    "status": "PENDING",
    "note": "Peer data populated after Comps module completes Peer Selection"
  },
  "wacc_inputs": {
    "risk_free_rate": 0.043,
    "source_rf": "US 10Y Treasury yield as of 2026-04-12",
    "equity_risk_premium": 0.055,
    "source_erp": "Damodaran implied ERP, January 2026",
    "beta_levered": 1.21,
    "cost_of_equity_capm": 0.110,
    "pre_tax_cost_of_debt": 0.051,
    "tax_rate_for_wacc": 0.21,
    "after_tax_cost_of_debt": 0.040,
    "target_equity_weight": 0.92,
    "target_debt_weight": 0.08,
    "wacc_calculated": 0.104,
    "confidence": "MEDIUM",
    "notes": "Debt weight based on current market structure; may need adjustment for LBO or recapitalization scenarios"
  },
  "data_quality_log": [
    {
      "field": "nwc_pct_revenue",
      "confidence": "MEDIUM",
      "issue": "NWC derived from BS delta; direct CF disclosure uses different definition",
      "recommendation": "Cross-check against working capital note in 10-K"
    },
    {
      "field": "beta_blume_adjusted",
      "confidence": "MEDIUM",
      "issue": "Beta calculated from 5yr weekly returns; may not reflect current risk profile post-business pivot",
      "recommendation": "Review if business model significantly changed in last 12 months"
    }
  ]
}
```

---

## 10. Layer 4 → Layer 5 预留 (v4.0 NEW — Wave 2 Placeholder)

Reserved data contract for the Layer 5 (Document/Report Generation) module, planned for Wave 2.

```json
{
  "schema_version": "2.0",
  "schema_id": "layer4_to_layer5",
  "status": "RESERVED_FOR_WAVE2",
  "description": "Data contract from completed financial model (Layer 4) to the document/report generation layer (Layer 5). Layer 5 will auto-generate investment memos, LP reports, and board presentations from structured model output.",
  "generated_at": "2026-04-12T14:00:00Z",

  "model_summary": {
    "company": "Sea Limited",
    "model_date": "2026-04-12",
    "track": "HF",
    "granularity_tier": "Standard",
    "all_gates_passed": true,
    "gate_timestamps": {
      "gate_1_research": "2026-04-11T10:00:00Z",
      "gate_2_partner": "2026-04-11T15:30:00Z",
      "gate_3_pm_instructions": "2026-04-12T09:00:00Z",
      "gate_4_delivery": "2026-04-12T13:45:00Z"
    }
  },

  "key_outputs": {
    "valuation": {
      "methodology": ["DCF", "Comps"],
      "implied_price_range": { "low": 55.0, "mid": 75.0, "high": 105.0 },
      "current_price": 68.50,
      "upside_mid": 0.095,
      "recommendation": "BUY",
      "conviction": "MEDIUM_HIGH"
    },
    "scenarios": {
      "bear": { "irr": 0.082, "moic": 1.55, "key_risk": "GMV growth decelerates to <10%" },
      "base": { "irr": 0.148, "moic": 2.15, "key_driver": "Take-rate expansion + gaming recovery" },
      "bull": { "irr": 0.225, "moic": 3.10, "key_upside": "AI commerce penetration accelerates" }
    }
  },

  "narrative_inputs": {
    "thesis_statement": "Sea Limited is a misunderstood re-acceleration story...",
    "key_risks": ["regulatory risk in SEA markets", "Shopee GMV competition from TikTok Shop"],
    "key_catalysts": ["gaming turnaround", "Shopee take-rate increase", "SeaMoney profitability"],
    "variant_view": "Consensus underestimates take-rate expansion potential by ~100bps"
  },

  "document_requests": {
    "investment_memo": {
      "requested": true,
      "format": "markdown",
      "target_pages": 8,
      "sections": ["executive_summary", "business_overview", "thesis", "valuation", "risks", "catalysts"]
    },
    "one_pager": {
      "requested": true,
      "format": "html",
      "style": "IB_tearsheet"
    },
    "sensitivity_appendix": {
      "requested": true,
      "format": "excel_export"
    }
  },

  "_wave2_note": "Layer 5 module will be implemented in Wave 2. This schema defines the interface contract so Layer 4 outputs can be forward-compatible."
}
```

---

## Schema Versioning Update (v2.0)

Current version: **2.0** (Wave 1 upgrade adds Schemas 8, 9, 10)

All v4.0 skills declare:

```yaml
schema_version: "2.0"
```

| Schema | Version | Status |
|--------|---------|--------|
| 1 — WACC → DCF | 1.0 | Stable |
| 2 — Terminal Value → DCF | 1.0 | Stable |
| 3 — Sensitivity → DCF | 1.0 | Stable |
| 4 — Integrated L3 → DCF | 1.0 | Stable |
| 5 — DCF → Integrated | 1.0 | Stable |
| 6 — Model Backtest Input | 1.0 | Stable |
| 7 — Confidence Level (Private) | 1.0 | Stable |
| 8 — Comps → Valuation Summary | 2.0 | **NEW v4.0** |
| 9 — Data Layer → Core | 2.0 | **NEW v4.0** |
| 10 — Layer 4 → Layer 5 | 2.0 | **RESERVED Wave 2** |
