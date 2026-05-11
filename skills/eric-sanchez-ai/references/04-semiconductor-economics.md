# Reference 04 — Semiconductor Economics: The Physics of Money

> "Moore's Law is dead. But Wright's Law lives. And Wright's Law is what makes AI economically inevitable."
> — Eric Sanchez

---

## Moore's Law Is Dead — What Replaced It

**Moore's Law (1965):** The number of transistors on a chip doubles approximately every two years.

This was both a technological observation and an industry roadmap. TSMC, Intel, Samsung, and the entire semiconductor supply chain organized their investment cycles around it for 50 years.

**Why Moore's Law is dead:**
- At sub-3nm nodes, transistors are ~10–15 silicon atoms wide
- Quantum tunneling effects (electrons appearing on the wrong side of a gate) become significant
- Power leakage per transistor does not decrease as transistors shrink (Dennard scaling failure)
- The cost per transistor stopped declining consistently around 2016–2018 at leading-edge nodes

**Key data:** Intel's 10nm node (marketed as equivalent to TSMC 7nm) cost more per transistor than its 14nm node. For the first time in the industry's history, shrinking was more expensive, not less.

**What replaced Moore's Law:**
1. **Advanced packaging** — connecting multiple dies as if they were one (CoWoS, SoIC, HBM stacking)
2. **Architecture specialization** — domain-specific chips (GPUs, TPUs, ASICs) optimized for specific workloads
3. **Wright's Law** — cost reduction through manufacturing experience accumulation, independent of geometry shrinks

---

## Wright's Law: The Surviving Cost Reduction Engine

**Wright's Law (1936):** For every doubling of cumulative production volume, cost falls by a constant percentage (the "learning rate").

This was first observed in aircraft manufacturing in the 1930s. It applies universally to manufacturing with learning curves — solar panels, batteries, and semiconductors.

### Wright's Law Applied to Semiconductors

**Typical learning rate for semiconductor manufacturing: 15–30% cost reduction per doubling of cumulative production**

For AI chips specifically:
- H100 production began ramp in late 2022 / early 2023
- By end of 2024, Nvidia has shipped ~500,000+ H100-class units
- Each doubling of cumulative volume corresponds to ~20–25% manufacturing cost reduction
- Die yield improves, packaging yield improves, test process efficiency improves

**Practical implications:**

*For Nvidia:* The H100 launched at ~$30K list. Manufacturing cost at ramp was ~$3,500–4,500. As volume doubles, that manufacturing cost falls — but Nvidia's pricing power lets it capture most of the margin expansion. Gross margins expanded from ~65% to 78%+ as volumes scaled.

*For inferene cost per token:* The training cost curve feeds directly into inference cost curves. As more chips are manufactured (cumulative volume doubles), chip cost falls → inference cost per token falls. This is why inference economics improve so rapidly even without architectural breakthroughs.

### Wright's Law vs. Moore's Law for Investors

| Dimension | Moore's Law | Wright's Law |
|-----------|-------------|--------------|
| Driver | Physics (transistor miniaturization) | Experience (manufacturing learning) |
| Status | Dead at leading-edge nodes | Alive and applicable |
| Benefit recipient | Entire industry (chips got cheaper for everyone) | Production leaders (TSMC, Nvidia benefit most) |
| Predictability | Now unpredictable (node timelines extend) | Highly predictable (apply to cumulative volume) |
| Investment implication | Node timing surprises create opportunity | Volume leader advantages compound |

**The critical insight:** Moore's Law benefited chip buyers (each generation cheaper). Wright's Law benefits production leaders — first movers who accumulate volume faster build cost advantages that compound.

---

## Fab Economics: The Capital Intensity of Physics

### What It Costs to Build a Leading-Edge Fab

| Fab Type | Capital Cost | Timeline | Output Capacity |
|----------|--------------|----------|-----------------|
| Leading-edge (≤3nm) | $15–25B | 4–6 years | ~100,000 wafer starts/month |
| Advanced (5–7nm) | $8–15B | 3–4 years | ~80,000 WSM |
| Mature (28nm+) | $2–5B | 2–3 years | ~100,000+ WSM |
| EUV refit (existing fab) | $2–5B additional | 18–24 months | Conversion, not new capacity |

**The $20B leading-edge fab in context:**
- TSMC's Arizona Fab 21 (N4/N3): $40B committed capital across two phases
- Samsung's Taylor, TX fab: $17B investment
- Intel's Ohio mega-fab: $20B Phase 1, up to $100B committed

**Economic payback:**
- At ~100,000 wafer starts/month, a fab with ~$500/wafer revenue generates $600M/month = $7.2B/year
- At 30–35% operating margin, $2.2–2.5B EBIT/year
- On $20B capital: 8–9 year raw payback (before CHIPS Act subsidies), 5–7 years with subsidies
- This explains why the semiconductor industry is brutally capital-intensive and why only 3 companies (TSMC, Samsung, Intel) can even attempt leading-edge fabs

**The CHIPS Act subsidy effect:**
- US CHIPS Act: $52B total, with ~$39B for fab construction grants
- TSMC Arizona: Received ~$6.6B in grants + $5B in loans
- Intel: Received ~$8.5B in grants + $11B in loans
- Samsung: Received ~$6.4B in grants
- Subsidy economics: Reduce payback period by 1–2 years, but do not change the fundamental capital intensity

---

## TSMC's Pricing Power: A Case Study in Monopoly Economics

TSMC has raised wafer prices at leading-edge nodes consistently:
- N5 (5nm): ~$17,000/wafer at launch; ~$16,000 current (modest decline)
- N3 (3nm): ~$20,000–22,000/wafer at launch (premium for first advanced FinFET+ node)
- N2 (2nm, projected): ~$25,000–30,000/wafer

**Why customers pay:** Because there is no alternative. For anyone building the world's most advanced chip, TSMC is the only supplier. Nvidia cannot go to Samsung for B200 production — Samsung's yield on leading-edge nodes has been insufficient for Nvidia's requirements.

**The pricing power test:** In 2022–2023, TSMC raised prices ~6–20% across various nodes. No customer left. No customer could leave. This is the definition of pricing power.

**Long-term pricing trajectory:** TSMC raises prices by 3–8% per year in normal conditions. At leading-edge nodes, the price increase is faster because:
1. Customer alternatives are non-existent
2. TSMC's R&D costs are rising (each node costs more to develop)
3. Demand from AI chips is structurally price-inelastic (a 10% wafer price increase is immaterial when the chip's value to the customer is 10–100× the wafer cost)

---

## Wafer Start Economics: Reading the Supply/Demand Signal

**Wafer starts** are the primary production volume metric for semiconductor fabs. Understanding the economics of wafer starts is essential for timing semiconductor investments.

### Key metrics:

**WSM (Wafer Starts per Month):** The raw volume metric. TSMC's global capacity: ~3M 12-inch equivalent WSM total, ~200K+ WSM at leading-edge (N3 + N5).

**Yield:** The percentage of functional chips per wafer. A 12-inch wafer might produce 200 potential chip dies. At 80% yield, that's 160 saleable chips. Yield varies by process node maturity and chip design complexity.

**Die size:** Larger chips (larger die area) mean fewer chips per wafer. The H100 die is ~814mm² — meaning relatively few per wafer (~100 gross dies on a 300mm wafer at typical distribution). A smartphone application processor at ~100mm² produces ~800+ gross dies per wafer.

**Die yield economics:**
```
Revenue per wafer = Gross dies × Yield × ASP
At $30K H100 price, ~100 gross dies, ~80% yield:
Revenue per wafer ≈ 80 × $30,000 = $2.4M per wafer
Wafer cost at N3: ~$20,000
Die cost: ~$250
Package + test: ~$500
Total cost: ~$750
Margin: ~97.5% on ASP (before Nvidia's own capex, R&D, sales)
```

This explains Nvidia's extraordinary gross margins — the chip's value ($30K) is vastly disconnected from its manufacturing cost (~$750). The value is in the architecture, software ecosystem, and supply constraints — not the silicon.

---

## Capex Cycles and Their Predictability

Semiconductor capex cycles are among the most predictable in all of investing — if you understand the physics of fab construction timelines.

### The cycle structure:

**Upswing triggers:**
- New application demand (AI, 5G, automotive) creates acute supply shortage
- Lead times extend to 12–52 weeks (normal: 8–12 weeks)
- ASPs rise, gross margins expand
- Fab operators order equipment and announce capacity expansions

**The 18–36 month lag:**
- Fab construction takes 18–24 months from groundbreaking to first wafer out
- New equipment (ASML EUV machines) has 18–24 month lead times itself
- Capacity additions thus lag demand signals by 18–36 months — causing overshoot

**Downswing trigger:**
- New capacity comes online simultaneously (everyone expanded at the same time)
- Inventory builds (customers who double-ordered to secure supply now have excess)
- ASPs fall, utilization drops, margins compress
- Stock prices lead the cycle by 6–9 months in each direction

**The AI exception (2023–2025):** The leading-edge fab capacity for AI chips (N3/N4/N5 via TSMC) is not being overbuilt. TSMC's CoWoS packaging capacity is even tighter. The cycle dynamics differ because:
1. Leading-edge capacity is structurally limited (not everyone can build these fabs)
2. Demand is structural (hyperscaler capex is a multi-year commitment, not order-based)
3. TSMC controls the allocation — it is rationing capacity, not competing for customers

**Mature node cycle (28nm, 40nm):** This IS following normal cyclical patterns. COVID-era demand (automotive, IoT, consumer electronics) caused aggressive capacity expansion; overcapacity has emerged in 2023–2024. Companies exposed to mature nodes are experiencing ASP pressure.

---

## ASP Trends and Yield Curves

**ASP (Average Selling Price) dynamics by market segment:**

| Segment | Current ASP Trend | Primary Driver |
|---------|------------------|----------------|
| AI Data Center GPUs (H100/B200) | Rising → Stable | Supply constraint + demand growth |
| AI networking chips | Rising | Design win concentration |
| PC/mobile application processors | Declining | Oversupply, consumer weakness |
| Automotive semiconductors | Declining | Inventory correction |
| Industrial MCUs | Declining | Significant overcapacity |
| Memory (DRAM, NAND) | Recovering | HBM premium, NAND floor |

**Yield curve evolution:** New process nodes always launch with low yield, improving over 12–24 months:
- N3 launch yield (TSMC, 2022): ~50–60%
- N3 mature yield (2024): ~80–85%
- Yield improvement = cost reduction + capacity expansion (same tools produce more good chips)

This creates a characteristic earnings pattern: early quarters after a node launch have gross margin compression (low yield, high cost), followed by rapid gross margin expansion as yield matures. This is a recurring investment opportunity.

---

## Practical Financial Model Checklist

When building a semiconductor company model:

- [ ] **Revenue:** Units × ASP. What's the ASP trajectory? Node transition = ASP jump + initial margin compression
- [ ] **Gross margin:** Wafer cost + packaging + test = COGS. What's the yield assumption?
- [ ] **R&D:** Semiconductor R&D is 15–25% of revenue for fabless; 8–15% for IDMs
- [ ] **Capex (IDMs/foundries):** 30–50% of revenue for leading-edge fabs. Model EBITDA-capex = "economic earnings"
- [ ] **Inventory:** Watch days-inventory-outstanding. >180 days is a warning sign. <60 days = supply-constrained
- [ ] **Customer concentration:** >30% one customer = structural risk. >50% = red flag
- [ ] **Wright's Law projection:** Apply 20–25% learning rate to cumulative volume trajectory. Where is ASP in 3 years?
- [ ] **Capex cycle position:** Are we in capacity expansion (rising capex, rising backlog) or digestion (flat capex, falling backlog)?
- [ ] **Node transition timing:** When is the next major product transition? Model the margin dip + recovery
- [ ] **TSMC wafer cost allocation:** Estimate COGS sensitivity to TSMC price increases (3–8%/year)
