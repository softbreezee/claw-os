# Geopolitics and Semiconductor Risk

> When to read this file: when assessing geopolitical risk for semiconductor and AI investments, when evaluating export control impact, when analyzing supply chain concentration risk, or when pricing Taiwan risk into valuations.

The semiconductor industry sits at the intersection of technology and geopolitics more than any other sector. Export controls, subsidies, and strategic competition between the US and China have made geopolitical analysis essential for chip investors.

---

## The CHIPS Act and Global Semiconductor Subsidies

### US CHIPS and Science Act ($280B Total, $52B for Chips)

Signed August 2022. The largest industrial policy investment in US history for semiconductors:
- **$39B in manufacturing incentives**: Grants for building fabs in the US
- **$13B for R&D**: National Semiconductor Technology Center, advanced packaging R&D
- **25% investment tax credit**: For semiconductor manufacturing equipment
- **Guardrails**: Recipients cannot expand advanced chip production in China for 10 years

**Key recipients and investments:**
- TSMC Arizona: $6.6B in grants for three fabs ($65B total investment)
- Intel: $8.5B in grants for fabs in Arizona, Ohio, New Mexico, Oregon
- Samsung: $6.4B in grants for Taylor, Texas fab expansion
- Micron: $6.1B for Idaho and New York memory fabs

**Investment implications:**
- Long-term positive for US domestic manufacturing
- Near-term capex inflation (construction costs, equipment demand)
- Doesn't solve the fundamental problem: TSMC's process technology lead is 5+ years

### EU European Chips Act (€43B)

- Targets doubling EU's global chip production share from 10% to 20% by 2030
- Intel Magdeburg fab: €30B project (largest single investment)
- Focus on automotive and industrial chips (not leading-edge logic)

### Asian Subsidies

- **Japan**: ¥2T+ ($15B+) for TSMC Kumamoto fab, Rapidus 2nm fab
- **South Korea**: $7B in subsidies + $200B private sector investment from Samsung/SK Hynix
- **China**: $47B (National IC Fund Phase III) — the most direct state support

---

## Export Controls: The Technology Chokepoint Strategy

### October 2022 Controls (Biden Administration)

The most significant technology trade restriction since the Cold War:

**What was restricted:**
- Advanced chips: GPUs/accelerators above specific compute thresholds (targeting Nvidia A100/H100)
- Semiconductor manufacturing equipment: EUV lithography (ASML), advanced deposition, etch, and inspection tools
- Supercomputer components: Any chip enabling >100 PFLOPS
- **US person restrictions**: US citizens/residents cannot support advanced chip development in China

**Impact on specific companies:**
- **Nvidia**: Lost China revenue (~25% of data center revenue). Created H20 downgrade chip for China market (later also restricted in 2025).
- **ASML**: Banned from selling EUV to China. DUV restrictions added in 2023.
- **Applied Materials / Lam Research / KLA**: Lost access to Chinese fabs for advanced equipment sales.
- **TSMC**: Must refuse orders for advanced node chips destined for Chinese companies.

### The "Chokepoint" Theory

The US identified three chokepoints where it has near-monopoly control:
1. **EDA tools**: Synopsys + Cadence = ~80% of chip design software
2. **EUV lithography**: ASML is the sole manufacturer. Zero alternatives.
3. **Advanced manufacturing equipment**: Applied Materials, Lam Research, KLA dominate deposition, etch, inspection

By restricting these three, the US can control China's ability to manufacture advanced chips — regardless of whether China can design them.

---

## Taiwan Risk: The Most Important Geopolitical Variable

### Why Taiwan Matters

TSMC manufactures:
- ~90% of the world's most advanced chips (<7nm)
- ~55% of all chips globally (by revenue)
- Every leading-edge AI chip (Nvidia, AMD, Apple, Qualcomm)

**If Taiwan's chip production were disrupted for even one quarter, the global economic impact would exceed $1 trillion.**

### Scenario Analysis

**Scenario 1: Status Quo Continues (Base Case, ~70% probability)**
- Tensions continue but no military action
- TSMC diversifies with Arizona, Japan, Germany fabs (but these are 3-5 years from full production)
- Investment implication: TSMC remains best-in-class, slight geopolitical discount

**Scenario 2: Blockade / Quarantine (~15% probability)**
- China imposes naval blockade without invasion
- Chip supply disrupted for months
- Investment implication: Massive short-term disruption, accelerated reshoring, alternative suppliers (Samsung, Intel) benefit

**Scenario 3: Military Conflict (~10% probability)**
- Active military engagement around Taiwan
- TSMC fabs likely destroyed or inoperable (TSMC reportedly has "kill switches")
- Investment implication: Catastrophic for entire semiconductor supply chain. Global recession. Intel and Samsung as sole alternatives for years.

**Scenario 4: Peaceful Resolution (~5% probability)**
- Negotiated arrangement that preserves TSMC operations
- Investment implication: Geopolitical discount removed, TSMC re-rates significantly upward

### How to Price Taiwan Risk

- Apply a 5-15% geopolitical discount to TSMC valuation (current market seems to apply ~10%)
- For companies 100% dependent on TSMC (Nvidia, AMD, Apple): the discount should flow through
- For companies with Samsung/Intel alternatives: lower discount
- **Hedging**: Some investors pair TSMC long with Samsung semi long as a geographic hedge

---

## China's Semiconductor Ambitions

### What China Can Do

- **Mature nodes (28nm+)**: China can manufacture independently. SMIC, Hua Hong have capacity.
- **Memory**: YMTC (NAND) and CXMT (DRAM) are building capacity, though 2-3 generations behind.
- **Packaging**: China is competitive in advanced packaging (OSAT leaders like ASE).
- **Design**: HiSilicon (Huawei), Unisoc, Biren can design competitive chips — but can't manufacture advanced ones domestically.

### What China Cannot Do (Yet)

- **Leading-edge logic (<7nm)**: Cannot manufacture without ASML EUV. SMIC achieved 7nm-equivalent using multi-patterning DUV, but yield is low and cost is high.
- **EUV lithography**: No domestic alternative. The technology requires optics precision at the atomic level. Decades away from replication.
- **Advanced EDA**: Domestic tools (Empyrean) exist but cannot replace Synopsys/Cadence for advanced designs.

---

## Supply Chain Vulnerability Mapping

For any semiconductor investment, map the geopolitical exposure:

| Layer | Chokepoint | Geographic Risk | Alternatives |
|-------|-----------|-----------------|--------------|
| Design tools (EDA) | Synopsys, Cadence | US-controlled | Chinese alternatives years behind |
| IP cores | ARM, Synopsys | UK/US-controlled | RISC-V (open source, growing) |
| Manufacturing (advanced) | TSMC | Taiwan | Samsung (Korea), Intel (US) |
| Manufacturing (mature) | SMIC, UMC, GlobalFoundries | Distributed | Many options |
| Equipment | ASML, Applied, Lam, KLA | Netherlands/US | No alternatives for leading-edge |
| Memory | SK Hynix, Samsung, Micron | Korea/US | YMTC (China, restricted) |
| Packaging | ASE, Amkor, TSMC | Taiwan/US | Distributed, less concentrated |

### Investment Decision Framework

For each holding, assess:
1. **Export control exposure**: Could new restrictions affect this company's revenue or supply?
2. **Taiwan concentration**: What % of supply comes from Taiwan?
3. **China revenue**: What % of revenue comes from China? Could it be restricted?
4. **Alternative sourcing**: If the primary source is disrupted, how quickly can alternatives ramp?
5. **Subsidy beneficiary**: Is this company receiving CHIPS Act or equivalent funding?

---
