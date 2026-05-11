# Wright's Law: Cost Curves, Learning Rates & Cumulative Production

> When to read this file: when analyzing cost deflation in any technology platform, projecting future unit economics using learning rates, distinguishing Wright's Law from Moore's Law, evaluating whether a technology is on a predictable cost-decline trajectory, or applying cost-curve analysis to investment thesis construction.

Wright's Law is the single most powerful analytical tool in ARK's framework. Understanding it deeply — its origins, its mechanics, its limitations, and its application to the five platforms — is essential to thinking the way Cathie Wood thinks.

> "The most important thing we do at ARK is cost curve analysis. And the framework we use is Wright's Law, not Moore's Law. The difference matters enormously." — Cathie Wood, ARK Invest Research Presentation, 2021

---

## Origins: T.P. Wright's 1936 Discovery

In 1936, Theodore Paul Wright, an aviation engineer at Curtiss-Wright Corporation, published a paper in the *Journal of Aeronautical Sciences* titled "Factors Affecting the Cost of Airplanes." His finding:

> **For every doubling of cumulative airplane production, labor costs fell by approximately 15%.**

This was not a coincidence or a short-term observation. It was a robust, reproducible relationship between **how much had been made** and **how cheap it had become**. Wright's insight: learning accumulates with production. Workers get faster. Processes get refined. Designs get optimized. Materials get sourced better. All of this happens as a function of **cumulative units produced**, not as a function of **calendar time**.

This single insight, extended to all manufacturing and technology systems, is Wright's Law.

---

## Wright's Law vs. Moore's Law: A Critical Distinction

ARK explicitly uses Wright's Law, not Moore's Law. The distinction is not semantic — it changes how you build models and make predictions.

| Dimension | Moore's Law | Wright's Law |
|-----------|-------------|--------------|
| **Origin** | Gordon Moore's 1965 observation about transistor density | T.P. Wright's 1936 empirical study of aircraft production |
| **Driver** | Calendar time (density doubles every ~2 years) | Cumulative production (cost falls with each doubling of units made) |
| **Mechanism** | Assumed; not mechanistically explained | Learning, process improvement, scale economies, design iteration |
| **Predictive variable** | Years elapsed | Units produced |
| **Universality** | Applies mainly to semiconductors | Applies to virtually any manufactured good with repeated production |
| **Investment implication** | Predicts based on time → inflexible | Predicts based on production volume → actionable for investment |
| **Failure condition** | When physics limits miniaturization | When production growth slows or learning saturates |

> "Moore's Law is about time. Wright's Law is about production. We care about production — because production is what we can model from demand forecasts, not from a calendar." — Brett Winton, ARK Invest Chief Futurist, ARK Research Blog

**Why the distinction matters for investment:**

If cost falls with time (Moore's Law), you cannot accelerate it by buying more. But if cost falls with cumulative production (Wright's Law), then demand growth — driven by falling prices — creates a self-reinforcing cycle:

**Lower cost → More demand → Higher production volume → Even lower cost → More demand**

This is the feedback loop ARK is investing in. It is not just a cost story. It is a demand creation story.

---

## The Mathematics of Wright's Law

**The Core Equation:**

```
C(n) = C(1) × n^(-b)

Where:
  C(n) = cost at the nth cumulative unit produced
  C(1) = cost of the 1st unit
  n    = cumulative units produced
  b    = learning parameter (determines the learning rate)
```

**The Learning Rate (LR):**

The learning rate is expressed as the percentage cost reduction for each doubling of cumulative production.

```
LR = 1 - 2^(-b)

Equivalently:
  b = -log(1 - LR) / log(2)
```

**Worked Example (EV Batteries):**

- LR = 18% (ARK's estimate for lithium-ion batteries, conservatively)
- b = -log(1 - 0.18) / log(2) = 0.284
- If current cost is $100/kWh at cumulative production X, then when production doubles to 2X, the new cost is:
  - C(2X) = $100 × 2^(-0.284) = $100 × 0.82 = $82/kWh

Every doubling of total batteries ever produced → cost falls to 82% of current level.

**Key insight:** Early doublings are cheap (absolute production is small). Later doublings require enormous production volume. This is why cost curves steepen initially, then slow — and why investors who miss the early phase miss most of the return.

---

## Empirical Learning Rates by Platform

ARK has documented learning rates across the five platforms. These are not projections — they are historically observed rates validated over multiple decades.

### Solar Photovoltaics: ~20% Learning Rate

- Cost per watt of solar panels: ~$76 in 1977 → ~$0.20 in 2023
- Every time cumulative production of solar panels doubled, cost fell by ~20%
- Total cost decline: >99.7% over ~45 years of Wright's Law operation
- This is why solar is now the cheapest source of electricity ever produced in human history

> "Solar is the greatest vindication of Wright's Law. People said for 30 years that solar would never be competitive. Those people didn't understand Wright's Law." — Cathie Wood, Energy Innovation Summit, 2022

### Lithium-Ion Batteries: ~18-28% Learning Rate

- Cost per kWh: ~$1,000 in 2010 → ~$100 in 2023 (ARK's tracking)
- Learning rate has been volatile: 18% conservatively, 28% in high-adoption scenarios
- **The critical threshold:** ARK has identified $100/kWh as the point where EVs reach upfront cost parity with ICE vehicles on an unsubsidized basis. This threshold was crossed in 2023 for many segments.
- Projection: If learning rate holds at 18%, batteries reach $50/kWh by ~2027-2030 (depending on production growth rate)

**What $50/kWh means:**
- EV manufacturing cost 20-30% below equivalent ICE vehicle
- Grid storage becomes competitive with natural gas peaker plants
- Electrification of aviation becomes economically plausible

### DNA Sequencing: ~40% Learning Rate (Fastest Ever Recorded)

- Cost per genome: ~$3,000,000,000 (Human Genome Project, completed 2003) → ~$200 (2023)
- This is a decline of ~99.9999% in 20 years
- The learning rate of ~40% per doubling is the highest ARK has ever documented — exceeding solar, batteries, and semiconductors
- **What this means:** At current trajectory, $0 effective cost per sequencing read is a mathematical inevitability

> "DNA sequencing is the most extraordinary example of Wright's Law we have ever tracked. The cost curve has exceeded every analyst projection, every year, for 20 consecutive years. We expect this to continue." — ARK Genomic Revolution Research, 2022

### EV Drivetrains: ~15% Learning Rate

- Separate from battery costs (which dominate), the mechanical and electronic drivetrain also follows Wright's Law
- Learning rate approximately 15% per doubling of cumulative EV production
- Combined with battery cost decline, total EV powertrain cost is declining at 22-25% per doubling

### AI Compute (Training Cost): ~60% Annual Decline

- This one is unusual because AI compute improvement is driven by both Wright's Law (chip production scale) AND algorithmic efficiency improvements
- Training a large language model equivalent in 2023 costs ~1,000x less than in 2020
- The combined driver is production volume (chips) + algorithm innovation (reducing compute required per task)

---

## Applying Wright's Law: The ARK Cost Projection Template

**Step 1: Establish the current cost baseline**
```
Current cost per unit: $XX per [kWh / genome / FLOP / km of robotic operation / ...]
As of: [date]
Source: [ARK tracking, NREL data, Illumina pricing, etc.]
```

**Step 2: Establish cumulative production baseline**
```
Cumulative production to date: XX [GWh / genomes / exaFLOPs / units]
Source: [industry data]
```

**Step 3: Project cumulative production at investment horizon**
```
Annual production today: XX units/year
Expected CAGR of annual production: XX% (based on demand model)
Cumulative production at 5-year horizon: XX units
Number of doublings implied: log(future/current) / log(2) = X.X doublings
```

**Step 4: Apply the learning rate**
```
Learning rate: XX%
Cost per doubling: multiply by (1 - LR)
Projected cost at 5-year horizon: $XX per unit
```

**Step 5: Check the "magic number"**
```
What is the cost threshold at which mass-market demand unlocks?
  - EVs: ~$100/kWh for battery cost parity
  - Genomics: ~$100/genome for routine clinical use
  - Solar: Already passed (~$0.02/kWh LCOE)
  - Robotics: ~$10,000/unit for mass-market industrial robots
Is the projected cost below or above this magic number at the 5-year horizon?
```

**Step 6: Investment conclusion**
```
If projected cost crosses the "magic number" within the investment horizon:
  → TAM expansion event is coming
  → This is an ARK-style buy opportunity (if S-curve hasn't already reflected it)

If projected cost remains above the "magic number":
  → Opportunity is real but further out than 5 years
  → Remain in Watch category

If cost curve is plateauing or learning rate is decelerating:
  → Investigate: saturation of learning? Physical limits? Supply chain constraint?
  → Potentially reduce conviction
```

---

## Historical Case Studies: Wright's Law in Action

### Case Study 1: Solar — The Textbook Vindication

In 2008, Al Gore and others predicted solar would become competitive with fossil fuels by 2030. Analysts said 2050. ARK said: apply Wright's Law. If the learning rate holds at ~20% and production grows at ~30% annually, parity arrives around 2015 in the sunbelt, 2020 globally.

Outcome: Solar reached cost parity in 2015-2017 in most markets. Wright's Law was right. The analysts using static models were wrong.

Investment implication: SolarCity, First Solar, and the broader solar ecosystem were enormous opportunities in 2010-2012 for investors who understood the cost curve.

### Case Study 2: EV Batteries — The Investment Thesis in Real Time

ARK began tracking battery costs in 2014. Their Wright's Law model projected:
- 2020: ~$150/kWh ✓ (actual: ~$130/kWh — ahead of schedule)
- 2023: ~$100/kWh ✓ (actual: ~$100-110/kWh — on track)
- 2026 projection: ~$60-70/kWh (on current Wright's Law trajectory)

Tesla, as the manufacturer driving more cumulative battery production than any other company, was the primary beneficiary of this cost curve. ARK's Tesla thesis is fundamentally a Wright's Law thesis.

> "When we bought Tesla in 2016 for ARK's portfolio, the battery cost curve told us this company was going to destroy the conventional auto industry. The cost curve doesn't lie." — Cathie Wood, Bloomberg Technology, 2021

### Case Study 3: DNA Sequencing — The Fastest Cost Curve Ever

The Human Genome Project (1990-2003) cost $3 billion. By 2007, a genome could be sequenced for $10 million. By 2014, $1,000. By 2022, $200. By 2023, approaching $100.

This ~40% learning rate far exceeded any prior technology's documented rate. The implication: precision medicine becomes economically viable faster than anyone predicted.

ARK's investment thesis in Illumina (the dominant sequencing platform provider) and later in multi-omics companies (Veracyte, Pacific Biosciences, Oxford Nanopore) was grounded in this cost curve analysis.

---

## Wright's Law Limitations: When the Model Fails

**Limitation 1: Production growth must continue**

Wright's Law requires cumulative production to grow. If production stalls — due to supply chain disruption, regulatory barriers, or demand collapse — the cost curve stalls. The learning does not automatically continue.

Investment implication: Track production growth rates as religiously as cost rates. If production growth slows, the Wright's Law thesis weakens.

**Limitation 2: Physical limits can truncate the curve**

Every technology has physical cost floors — the cost of raw materials, energy, and irreducible labor. When the cost curve approaches these limits, the learning rate decelerates. Moore's Law has encountered this with transistor physics; genomics will eventually encounter the cost of sample preparation and data storage.

**Limitation 3: Architecture discontinuities reset the curve**

A new technology architecture (e.g., solid-state batteries replacing lithium-ion) does not inherit the learning from the prior architecture. The Wright's Law curve restarts. This is both a risk (thesis invalidation for incumbent technology companies) and an opportunity (the new architecture's curve begins at a higher cost, creating a new investment window).

**Limitation 4: Supply chain constraints can create temporary dislocations**

Lithium supply, rare earth elements for EV motors, and polysilicon for solar all experienced supply-driven cost increases in 2021-2022 despite continued production volume growth. Wright's Law describes learning-driven cost reduction; external commodity inputs are separate variables.

> "Wright's Law doesn't mean cost always falls. It means that if you produce enough, it tends to fall. Supply chain problems, geopolitical disruptions, and raw material shortages can cause temporary reversals. We model these separately." — ARK Invest Research Note, 2022

---

## Common Analytical Errors

**Error 1: Confusing learning rate with growth rate**
The learning rate (% cost decline per doubling) is not the same as the annual cost decline rate. Confusing these leads to wildly inaccurate projections. Always convert to "per doubling" terms.

**Error 2: Extrapolating linearly in time**
Wright's Law operates in production-volume space, not time space. In periods of slow production growth, cost decline slows in calendar time. In periods of rapid production growth (ARK's expected scenario), cost can decline faster than annual projections suggest.

**Error 3: Ignoring the feedback loop**
Falling costs drive demand, which drives production, which drives further cost declines. A linear model misses this. ARK's models include demand feedback — lower cost → more demand → higher cumulative production → faster cost decline.

**Error 4: Applying one learning rate to the entire value chain**
Different components of a system may have different learning rates. EV batteries (18-28% LR) have a different trajectory from EV electronics (faster) from EV structural components (slower). The integrated system's cost decline is a weighted average.

---
