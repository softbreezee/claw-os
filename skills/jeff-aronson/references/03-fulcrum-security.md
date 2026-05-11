# Reference 03 · Fulcrum Security Analysis: Finding Value in the Waterfall

> "The fulcrum security is the most intellectually interesting instrument in finance. It's a credit instrument that will become an equity instrument. You have to be right about the value of both to know what it's worth today."
> — Distressed investing principle

---

## What Is a Fulcrum Security?

In a restructuring, a company's enterprise value is distributed to its creditors and equity holders according to absolute priority: secured creditors first, then unsecured, then subordinated, then equity. When the enterprise value is insufficient to pay all claims in full, **some class of claimants receives less than par — or receives equity in the reorganized company instead of cash.**

The **fulcrum security** is the debt tranche that sits at the pivot point: claims senior to the fulcrum are paid in full (either in cash or in new debt at par); claims junior to the fulcrum receive nothing (or a nominal recovery); **the fulcrum tranche receives equity in the reorganized company** as their recovery.

**This is the most important concept in distressed investing**: the fulcrum security is simultaneously:
- The **last** instrument that is "in the money" (still partially covered by enterprise value)
- The **first** instrument that is "out of the money" (cannot be paid at par in a reorganization)
- The instrument that **converts to equity** in a restructuring, giving its holders control of the reorganized company

Whoever owns the fulcrum security effectively owns the reorganized equity — and thus has the most to gain if the underlying business recovers.

---

## How to Find the Fulcrum Security

### Step 1: Estimate Enterprise Value Range

The entire fulcrum analysis depends on getting the enterprise value right. Use multiple methods:

**Method 1: Trading Comparable Multiples**
- Identify 5–10 comparable companies (same industry, similar business model)
- Calculate EV/EBITDA, EV/Revenue multiples for each
- Apply the median (or a conservative discount) to the distressed company's EBITDA
- Example: Comps trade at 7–9x EBITDA; distressed company's EBITDA is $300M; EV = $2,100M–$2,700M

**Method 2: Transaction Comparables**
- Identify recent M&A or LBO transactions in the same industry
- Transaction multiples are usually higher than trading multiples (control premium)
- Apply a discount for distress (distressed sales typically occur at 10–30% below normal M&A multiples)
- Example: Sector transactions at 8–10x; apply 20% distress discount → 6.4–8.0x; EV = $1,920M–$2,400M

**Method 3: Discounted Cash Flow (DCF)**
- Project normalized FCF over 5–10 years (use restructured company without the debt burden)
- Apply a discount rate appropriate for the reorganized entity's risk profile
- Terminal value: EV/EBITDA multiple or Gordon Growth Model

**Method 4: Asset-Based Value**
- For asset-heavy companies (real estate, energy, manufacturing): appraise assets directly
- This gives a liquidation floor — the EV cannot go below this in any scenario (in theory)
- Example: Real estate portfolio appraised at $1.8B; this is the hard floor

**Triangulate across methods.** The range of EV estimates tells you how much uncertainty there is in the fulcrum analysis.

### Step 2: Map the Claims Waterfall

List all debt claims in priority order:

```
Priority    Claim                       Amount ($M)   Notes
────────────────────────────────────────────────────────────────
1st lien    Revolving Credit Facility   $300M        Secured, first lien
1st lien    Term Loan B                 $800M        Secured, first lien
2nd lien    Second Lien Term Loan       $400M        Secured, second lien
Unsecured   Senior Unsecured Notes      $600M        8.5% notes due 2027
Unsecured   Holdco PIK notes            $200M        Subordinated holdco level
Equity      Common Stock                Residual     ~20M shares outstanding

Total Debt: $2,300M
────────────────────────────────────────────────────────────────
```

### Step 3: Overlay Enterprise Value

Now overlay the EV estimates:

| EV Scenario | EV ($M) | Bear | Base | Bull |
|-------------|---------|------|------|------|
| EV estimate | — | $1,600M | $2,200M | $2,800M |
| 1st Lien recovery | $1,100M | 100% (✅) | 100% (✅) | 100% (✅) |
| Remaining after 1st lien | — | $500M | $1,100M | $1,700M |
| 2nd Lien recovery | $400M | 100% (✅) | 100% (✅) | 100% (✅) |
| Remaining after 2nd lien | — | $100M | $700M | $1,300M |
| Senior Unsecured recovery | $600M | **17% (🔴)** | **100% (✅)** | **100% (✅)** |
| Remaining after Sr. Unsec. | — | $0 | $100M | $700M |
| Holdco PIK recovery | $200M | 0% (🔴) | 50% (⚠️) | 100% (✅) |
| Equity recovery | Residual | $0 | $0 | $500M |

### Step 4: Identify the Fulcrum

In the **base case** ($2,200M EV):
- 1st Lien: paid in full ✅ → NOT the fulcrum
- 2nd Lien: paid in full ✅ → NOT the fulcrum
- **Senior Unsecured Notes: paid in full at $700M remaining, with $100M surplus** → borderline, nearly the fulcrum
- Holdco PIK: receives ~50 cents → **this is close to the fulcrum**
- Equity: zero ❌ → out of the money

In the **bear case** ($1,600M EV):
- **Senior Unsecured Notes: receives only 17 cents on the dollar** → **this is the fulcrum in the bear case**

**The fulcrum shifts depending on your EV estimate.** This is the core analytical challenge: if you believe the base case, the Senior Unsecured is fully covered and the Holdco PIK is the fulcrum. If you believe the bear case, the Senior Unsecured is the fulcrum. Getting the EV right — or at least narrowing the range — is essential to identifying the right instrument to buy.

> "In distressed investing, you're not arguing about stock multiples. You're arguing about which creditors get paid in full and which get converted to equity. The EV estimate determines everything."

---

## Pricing the Fulcrum Security

### Dollar Price vs. Implied EV

If the Senior Unsecured Notes are trading at **65 cents** in the market:
- Market is implying that recovery on the Sr. Unsecured is 65 cents
- Implied EV: 1st lien ($1,100M) + 2nd lien ($400M) + 65% × $600M = $1,100M + $400M + $390M = **$1,890M**

**The key question**: Does the market's implied EV of $1,890M make sense?

If your analysis says the business is worth $2,200M in the base case, then:
- The bonds at 65 cents are **cheap** — you expect to recover close to par
- Expected return: buy at 65, recover at ~100, plus carry = 25–40%+ total return

If your analysis says the business is worth $1,600M in the bear case, then:
- The bonds at 65 cents are **fairly priced** (recovery would be ~17 cents... wait, that doesn't add up)
- *Actually*, the market at 65 cents is pricing in the **blended probability** of different scenarios

**Scenario-weighted recovery calculation**:
- Bear case (30% probability): 17 cents recovery → contribution: 5 cents
- Base case (50% probability): ~100 cents recovery → contribution: 50 cents
- Bull case (20% probability): 100 cents recovery → contribution: 20 cents
- **Expected recovery: 75 cents**

If bonds are trading at 65 cents and expected recovery is 75 cents, the bonds offer a 15% upside to expected value — but the distribution is bimodal (either great outcome or terrible outcome). This is the risk/return profile of the fulcrum.

### The Fulcrum Return Profile

The fulcrum security has a fundamentally different risk/return profile than regular credit or equity:

**Outcomes**:
1. **Company recovers operationally** (avoids restructuring): Bonds return to par + carry; return = 20–40% based on purchase price and timeline
2. **Restructuring occurs, fulcrum converts to equity**: Return depends on quality of reorganized equity; can be 0–500%+ depending on execution and business recovery
3. **Liquidation** (rare, usually value-destructive): Recovery well below par; return = loss

The bimodal nature of the fulcrum means that:
- Small errors in EV estimation lead to large errors in return
- Position sizing must account for the wide outcome distribution
- The analyst must be genuinely confident in the EV estimate — or size the position conservatively to account for estimation error

---

## The Absolute Priority Rule (APR)

In U.S. Chapter 11 bankruptcy, the **Absolute Priority Rule** (APR) requires that junior claimants receive nothing until senior claimants are paid in full. This is the legal foundation of the waterfall analysis.

**In practice**, the APR is frequently negotiated around:
- **New value exception**: Equity holders can contribute new capital and retain equity even if creditors aren't paid in full — but the new capital must be "new value," and the plan must be fair
- **Settlement/consent**: Junior classes can receive value if senior classes consent (which they will if it facilitates a quick resolution)
- **Gifting**: Senior creditors can "gift" value to junior creditors as part of a consensual plan

**Why this matters**: When building your waterfall model, understand that the theoretical APR outcome is the floor. In practice, negotiations between classes lead to deviations. Equity often gets 5–10 cents even when technically out of the money, because management needs to be incentivized and fights take time and money. Model the APR waterfall first, then adjust for likely negotiated deviations.

---

## Historical Fulcrum Examples

### Washington Mutual (2008)
- **Capital Structure**: FDIC seized the bank; the holding company (WMI) went into Chapter 11 separately
- **Fulcrum**: WMI senior unsecured notes and subordinated debt — they had a claim on value trapped in the holding company
- **Complexity**: Multiple years of litigation over billions in tax refunds and assets
- **Outcome**: Extended litigation; eventually settled with significant but below-par recoveries for most unsecured claimants
- **Lesson**: Legal complexity can massively extend the timeline and reduce effective returns even when the EV is sufficient

### Energy Future Holdings / TXU (2014)
- **Capital Structure**: The largest LBO in history ($45B); suffered from collapsing natural gas prices
- **Fulcrum**: First lien EFIH (Energy Future Intermediate Holding) debt at the holdco level — they had a claim on the Oncor regulated utility subsidiary
- **Complexity**: Tax considerations prevented a clean sale of Oncor for years; multiple plan attempts failed
- **Outcome**: After years of litigation, Oncor was sold to Sempra; EFIH first lien recovered ~85–90 cents; EFIH second lien ~30–40 cents
- **Lesson**: Regulatory complexity (Oncor was regulated) can trap value for years; regulatory change-of-control approvals add duration risk to distressed investments

### CIT Group (2009)
- **Capital Structure**: $75B in assets, primarily commercial loans; funded with medium-term notes and commercial paper
- **Fulcrum**: Senior unsecured notes (no secured debt in the traditional sense — CIT's assets were financial assets)
- **Speed**: Pre-packaged bankruptcy — filed and emerged in 38 days
- **Outcome**: Noteholders received equity in reorganized CIT; reorganized CIT was eventually acquired by First Citizens BancShares
- **Lesson**: Pre-packaged bankruptcies can move extremely fast when creditors are concentrated and aligned; speed limits time value of money cost and operational damage

---

## Key Formulas and Quick Reference

**Implied EV from bond price**:
Implied EV = (Senior claims above fulcrum) + (Fulcrum bond face value × dollar price ÷ 100)

**Fulcrum recovery in restructuring**:
Fulcrum recovery = (EV − Senior claims above fulcrum) ÷ Fulcrum bond face value

**Return from fulcrum (no restructuring / trading back to par)**:
Return = (Par + Accrued interest − Purchase price) ÷ Purchase price

**Return from fulcrum (restructuring, converts to equity)**:
Return = (Reorganized equity value per bond × shares received) ÷ Purchase price

**Key question to resolve before buying**:
At this dollar price, what EV am I implying? Is that EV conservative, fair, or optimistic?
