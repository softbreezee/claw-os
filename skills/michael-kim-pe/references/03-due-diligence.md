# Due Diligence Framework

> "In PE, the deal memo gets you to signing. Due diligence is what you do to make sure you haven't been fooled."

PE due diligence is not academic research — it is adversarial hypothesis testing. Your job is to find the three things management didn't tell you, the two assumptions that don't hold, and the one risk that could kill the investment.

---

## DD Architecture Overview

A full PE buyout requires 6 parallel workstreams, all reporting to the deal team lead:

| Workstream | Owner | Timeline | Budget |
|-----------|-------|----------|--------|
| Commercial DD | Strategy consulting firm | 6–8 weeks | $500K–1.5M |
| Financial DD (QoE) | Big 4 accounting firm | 6–8 weeks | $400K–1.0M |
| Legal DD | M&A law firm | 6–8 weeks | $300K–800K |
| Management DD | Executive search + reference firm | 4–6 weeks | $50–150K |
| IT/Technology DD | Tech consulting firm | 4–6 weeks | $150–400K |
| Environmental/Regulatory DD | Specialist firm | 4–8 weeks | $100–500K |

**Total DD budget:** $1.5M–4.5M for a mid-size buyout. This is not an expense; it is insurance.

---

## 1. Commercial Due Diligence (CDD)

**Objective:** Independently verify the market size, competitive position, and customer loyalty underpinning the investment thesis.

### 1A: Market Analysis

**TAM / SAM / SOM Assessment:**
```
Total Addressable Market (TAM):    Size of the entire opportunity globally
Serviceable Addressable Market (SAM): Portion the company can realistically reach
Serviceable Obtainable Market (SOM): Realistic near-term market share target
```

**Key questions:**
- Is the market growing, stable, or declining? At what rate?
- What are the structural drivers? (Regulation, demographics, technology, trade flows)
- How has the market evolved in downturns? (Cyclicality test)
- What does the market look like in 5 years? (Will PE hold period coincide with headwinds?)

**Red flags in CDD:**
- TAM inflated by including markets the company cannot serve
- Market share data based on management estimates vs. independent sources
- Market growth projections extrapolated from peak years (2020–2021 distortions)

### 1B: Competitive Position

**Market share analysis:**
- Map competitors by revenue, market share, geographic footprint
- Identify the #1 competitor: what does the target do better or worse?
- Porter's Five Forces: buyer power, supplier power, new entrants, substitutes, rivalry

**Competitive moat assessment (PE lens):**
| Moat Type | Durability | Value in 5-Year Hold |
|-----------|-----------|---------------------|
| Regulatory license (banking, healthcare) | High | Very high |
| Long-term customer contracts (3–5 year) | Medium-High | High |
| Network effects | High | High (if defensible) |
| Brand loyalty | Medium | Medium |
| Cost leadership at scale | Medium | Medium |
| Proprietary technology | Medium | Medium (IP clock runs) |
| Switching costs | Medium-High | High if contracts enforced |
| Low-cost location / labor arbitrage | Low | Low (can be replicated) |

### 1C: Customer Interviews (Most Important DD Activity)

**Minimum standard:** 15–20 independent customer interviews; 5–10 churned or lost customer interviews

**Interview protocol:**
1. **Identity:** Do not reveal you represent a potential buyer — use cover story ("industry research")
2. **Rapport:** Start with their business, not the target company
3. **Net Promoter Score equivalent:** "Would you recommend [Company] to a peer?"
4. **Relationship depth:** "How long have you worked with them? What would make you switch?"
5. **Pricing:** "When did they last raise prices? How did you react?"
6. **Competitive check:** "Who else do you consider? Why do you still use [Company]?"
7. **Warning question:** "What would have to be true for you to reduce your spend with them?"

**Green flags:** Unsolicited positive praise; customer referenceable for marketing; multi-year relationship; budget is locked in

**Red flags:** Hesitation; mentions competitor by name as alternative; describes relationship as "transactional"; mentions dispute or contract negotiation

---

## 2. Financial Due Diligence (Quality of Earnings)

**Objective:** Determine the true, normalized, sustainable EBITDA that is the basis for the acquisition price.

### 2A: EBITDA Normalization Process

**Step 1: Reconcile Management EBITDA to Audited Financials**
- Start with audited P&L EBITDA (or management accounts if unaudited — flag this)
- Reconcile every line item management adjusts
- Document: what is the adjustment, why does management make it, is it legitimate?

**Step 2: Test Each Adjustment**

| Adjustment | Accept / Reject / Reduce |
|-----------|--------------------------|
| Owner salary above market rate | Accept (legitimate normalization) |
| "One-time" restructuring (4th year in a row) | Reject (structural cost) |
| R&D reclassified as capex (GAAP vs. IFRS) | Accept with footnote |
| Revenue pulled forward from next quarter | Reject |
| Vendor rebates booked above the line | Investigate timing |
| Synergies not yet realized | Reject unless contractually committed |

**Step 3: Monthly Revenue / EBITDA Trending**
- Build a monthly revenue and EBITDA chart for 36 months minimum
- Look for: seasonal patterns, one-month spikes, deteriorating trends hidden in annual averages
- Compare each month Y/Y to identify organic growth vs. seasonal normalization

### 2B: Revenue Quality Assessment

| Revenue Quality Factor | High Quality | Low Quality |
|----------------------|-------------|------------|
| Contract type | Multi-year recurring | Spot/project-based |
| Customer concentration | Top 10 = < 30% of rev | Top 1 = > 25% of rev |
| Revenue visibility (forward) | 12+ months backlog | < 3 months visibility |
| Channel | Direct to end customer | Multi-layer distribution |
| Payment terms | Prepaid / milestone | Net 60–90+ |
| Churn / renewal rate | > 90% annual retention | < 80% annual retention |

### 2C: CapEx and FCF Analysis

**CapEx classification:**
- **Maintenance CapEx:** Required to maintain current revenue base (should be ≤ 3–5% of revenue for non-capital-intensive businesses)
- **Growth CapEx:** Discretionary investment for future expansion (exclude from normalized EBITDA-to-FCF conversion)
- **One-time CapEx:** System implementations, facility build-outs (understand timeline)

**FCF Conversion Check:**
```
Normalized EBITDA:           $100M
Less: Taxes (effective rate): ($20M)
Less: Maintenance CapEx:      ($8M)
Less: Working Capital change: ($5M)
Normalized FCF:               $67M

FCF Conversion = $67M / $100M = 67% → Passes screen (≥ 60%)
```

### 2D: Balance Sheet and Contingent Liabilities

- **Debt schedule:** All instruments, maturities, covenants, change of control provisions
- **Pension obligations:** Underfunded pensions are PE liabilities (especially Japan — very common)
- **Off-balance-sheet items:** Operating leases, factored receivables, vendor financing
- **Tax exposures:** Transfer pricing, deferred tax assets/liabilities, jurisdictional risks
- **Earn-outs payable:** Prior acquisition earn-outs still running

---

## 3. Legal Due Diligence

**Objective:** Identify legal risks that could reduce value, create liability, or block regulatory approval.

### Priority Legal Items

**Material Contracts:**
- Customer contracts: termination provisions, change of control triggers, exclusivity
- Supplier contracts: pricing, capacity guarantees, change of control
- Lease agreements: remaining term, renewal rights, landlord consent required?
- Debt agreements: existing covenants, change of control provisions, prepayment penalties

**Intellectual Property:**
- Patents: owned vs. licensed; expiry dates; pending litigation
- Trademarks: registered in operating jurisdictions; any challenges?
- Trade secrets and know-how: documented? employee agreements in place?
- Software: open-source license compliance (GPL contamination risk)
- Customer data: GDPR/PIPL compliance; data breach history

**Litigation:**
- Active cases: probability of adverse outcome; maximum exposure
- Regulatory investigations: government inquiries, sanctions exposure
- Employment claims: class action risk; labor law compliance (especially Korea, Japan)
- Environmental: site contamination; regulatory orders

**Regulatory Approvals Required:**
- Antitrust / competition authority (KFTC in Korea, JFTC in Japan, SAMR in China)
- Foreign investment review (CFIUS equivalent)
- Industry-specific licensing (financial services, healthcare, telecommunications)

---

## 4. Management Due Diligence

**Objective:** Independently verify the management team's capability, integrity, and PE-compatibility before committing capital.

### Background Verification

**For CEO and CFO (minimum):**
- [ ] Employment history verification (every position listed)
- [ ] Criminal and civil litigation record (all jurisdictions)
- [ ] Professional license verification (CPA, securities, etc.)
- [ ] Education verification (common fraud area, especially international candidates)
- [ ] Credit check (personal financial stress = red flag)
- [ ] Social media and press search (reputational issues)

**Reference Calls (15 minimum for CEO; 8 for CFO):**
- Format: structured, 30-minute calls; ask for references from bosses, peers, AND subordinates
- Key questions:
  1. "What are their top three strengths?"
  2. "In what situations did they struggle?"
  3. "Would you hire them again? Without hesitation?"
  4. "How do they handle adversity / missing a target?"
  5. "Are they builders or maintainers?"

### Management Compensation and Incentive Review

- Current total compensation: base, bonus, equity, benefits
- How is bonus calculated? What percentage was paid in last 3 years?
- Any retention contracts or golden parachutes triggered by the sale?
- Identify non-compete scope: duration, geography, sector

---

## 5. IT and Operations Due Diligence

**Objective:** Assess technology stack health, operational infrastructure, and digital risk.

### IT Health Check

**Systems inventory:**
- ERP system (age, customization level, vendor support status)
- CRM and customer data systems
- Financial reporting systems (can they close books in < 15 days?)
- Manufacturing / operational systems

**Red flags:**
- Core system end-of-life (SAP version no longer supported)
- Heavy reliance on Excel for operational management (fragile, unscalable)
- No single version of financial truth (multiple systems, manual reconciliation)
- Cybersecurity: no formal incident response plan; prior breaches not disclosed

**Technology investment required:**
- Estimate cost and timeline for critical system upgrades
- Deduct this from the purchase price or model as Year 1–2 capex drag

### Operations Assessment

- **Manufacturing/production capacity:** Utilization rate; bottlenecks; maintenance backlog
- **Supply chain:** Concentration in single supplier; geopolitical exposure (China supplier risk)
- **Quality systems:** ISO certifications; defect rates; customer complaint history
- **Facility condition:** Age, maintenance status, owned vs. leased

---

## DD Summary Red Flag Checklist

**Automatic deal-killers:**
- [ ] Material misrepresentation by management during DD
- [ ] Undisclosed regulatory investigation with criminal exposure
- [ ] Customer concentration > 40% in single customer
- [ ] EBITDA cannot be verified within 15% of management claim
- [ ] Ongoing litigation with expected liability > 20% of purchase price

**High-concern items (require significant price reduction or reps/warranties insurance):**
- [ ] Pension underfunding > 10% of EBITDA
- [ ] Change of control provision triggers key customer exits
- [ ] Management team unwilling to roll equity
- [ ] Single product / service with no revenue diversification
- [ ] IT systems require > $20M in near-term remediation
