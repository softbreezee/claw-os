# Data Acquisition Protocol

> When to read this file: ALWAYS read first before any analysis. This protocol defines what data to collect and how to collect it before applying the analytical framework. Data must be gathered before analysis begins — do not analyze with incomplete data.

---

## Core Principle: Data Before Analysis

> "The first rule of good thinking is to gather the facts before you form an opinion." — Munger

**Every analysis begins with data gathering. Do not skip this step.** Without data, you are guessing — and Munger despises guessing.

---

## Standard Data Requirements

### Tier 1: Mandatory (Must Have Before Analysis Begins)

These data points are required for any meaningful analysis. If they cannot be obtained, state what is missing and adjust confidence accordingly.

**Financial Metrics (10-Year History Preferred, 5-Year Minimum)**
```
Revenue (annual, quarterly if available)
Net Income
Operating Income
Free Cash Flow
Capital Expenditure
Depreciation & Amortization
Return on Invested Capital (ROIC)
Return on Equity (ROE)
Gross Margin
Operating Margin
Total Debt
Shareholders' Equity
Cash & Equivalents
Outstanding Shares (check for dilution trend)
```

**Valuation Data (Current)**
```
Market Capitalization
Enterprise Value
Current Stock Price
P/E Ratio (trailing and forward)
EV/EBITDA
Price/Free Cash Flow
```

**Business Context**
```
Company description and business segments
Revenue breakdown by segment/geography
Top competitors
Industry classification (GICS or equivalent)
```

### Tier 2: Important (Significantly Enhances Analysis)

**Management & Governance**
```
Insider transactions (last 2 years — buys vs. sells)
Executive compensation structure (latest proxy statement)
Share buyback history (5 years)
Dividend history (10 years)
CEO tenure and background
Board composition (insider vs. independent directors)
```

**Competitive Position**
```
Market share and trend
Customer concentration (top 10 customers as % of revenue)
Supplier concentration
R&D spending as % of revenue
Goodwill & intangible assets
```

### Tier 3: Supplemental (Use When Available)

```
Analyst consensus estimates (revenue, EPS)
Credit rating
Recent material news / press releases (last 90 days)
SEC filings highlights (10-K risk factors, MD&A)
Industry-specific KPIs
Peer comparison data (3-5 closest competitors)
```

---

## Data Collection Methods (Priority Order)

### Method 1: MCP Data Connectors (Preferred — Automated)

If MCP connectors are available and configured (CapIQ, LSEG, FactSet, Daloopa, etc.), use them:

```
Step 1: Query MCP for Tier 1 financial metrics (10-year history)
Step 2: Query MCP for current valuation data
Step 3: Query MCP for Tier 2 data (insider trades, compensation)
Step 4: Cross-validate key metrics across sources if multiple MCPs available
```

### Method 2: Web Data Sources (Free, Always Available)

When MCP connectors are not available, use public data sources:

| Data Need | Free Source | How to Access |
|-----------|------------|---------------|
| Financial statements (US) | SEC EDGAR | `web_fetch` → `sec.gov/cgi-bin/browse-edgar` → 10-K/10-Q filings |
| Financial metrics | StockAnalysis.com | `web_fetch` → `stockanalysis.com/stocks/{ticker}/financials/` |
| Valuation multiples | StockAnalysis.com | `web_fetch` → `stockanalysis.com/stocks/{ticker}/financials/?p=quarterly` |
| Insider transactions | SEC EDGAR Form 4 | `web_fetch` → `sec.gov/cgi-bin/browse-edgar?action=getcompany&type=4` |
| Market cap / price | Yahoo Finance | `web_fetch` → `finance.yahoo.com/quote/{ticker}` |
| Company profile | SEC EDGAR 10-K | Annual report Item 1 (Business Description) |
| Industry peers | StockAnalysis.com | `web_fetch` → `stockanalysis.com/stocks/{ticker}/competitors/` |
| News | Google News | `web_fetch` → `news.google.com/search?q={company_name}` |

### Method 3: User-Provided Data

The user may provide:
- Excel files with financial models
- Annual reports or earnings call transcripts
- Specific data points or estimates
- Industry reports

**Always use user-provided data as the primary source when available** — it may contain proprietary analysis or corrections not available in public sources.

---

## Data Validation Checklist

Before proceeding to analysis, verify:

- [ ] **Completeness**: Do I have at least 5 years of Tier 1 data?
- [ ] **Consistency**: Do revenue/income figures match across sources?
- [ ] **Currency**: Is the data current (latest fiscal year + latest quarter)?
- [ ] **Comparability**: Are accounting standards consistent across periods (watch for GAAP changes, restatements)?
- [ ] **Red flags**: Any data anomalies (sudden jumps in goodwill, revenue recognition changes, auditor switches)?

If data is incomplete, explicitly state:
```
DATA GAPS:
- [List missing data points]
- [Impact on analysis confidence: High/Medium/Low]
- [Mitigation: What proxy or estimate can be used]
```

---

## Output: Structured Data Package

After collection, organize data into this standard format before beginning analysis:

```
## Data Package: [Company Name] ([Ticker])
As of: [Date]

### Financial Summary (Latest FY)
Revenue: $XX.XB | Net Income: $XX.XB | FCF: $XX.XB
ROIC: XX% | ROE: XX% | Gross Margin: XX% | Op Margin: XX%
Total Debt: $XX.XB | D/E: X.XX | Cash: $XX.XB

### 10-Year Trends
[Table: Year | Revenue | Net Income | FCF | ROIC | Shares Outstanding]

### Valuation Snapshot
Market Cap: $XX.XB | P/E: XX | EV/EBITDA: XX | P/FCF: XX
52-Week Range: $XX - $XX | Current: $XX

### Management & Governance
Insider activity (24mo): X buys / Y sells
CEO: [Name], tenure: X years
Buyback activity: [Net shares repurchased over 5 years]

### Competitive Position
Market share: ~XX% | Trend: [Growing/Stable/Declining]
Key competitors: [List]

### Data Quality Assessment
Tier 1 coverage: [Complete/Partial — list gaps]
Tier 2 coverage: [Complete/Partial — list gaps]
Overall data confidence: [High/Medium/Low]
```

---

## For Batch Processing (L1 Quick Scan)

When processing multiple tickers in batch mode, collect only the **L1 Minimal Dataset**:

```
Per ticker (L1 Minimal):
  - Market Cap
  - P/E (trailing)
  - ROIC (latest year)
  - Revenue growth (3-year CAGR)
  - Gross Margin (latest)
  - D/E ratio
  - FCF yield (FCF / Market Cap)
  - Insider buy ratio (24 months)
```

This can be collected in ~30 seconds per ticker via web_fetch and is sufficient for the 4-question quick filter.

---
