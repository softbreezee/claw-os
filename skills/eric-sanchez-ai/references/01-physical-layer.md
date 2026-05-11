# Reference 01 — The Physical Layer: Why Physics Determines Winners

> "Software is an abstraction over physics. When the abstraction breaks down, the physics always wins."
> — Eric Sanchez

---

## The Collapse of Dennard Scaling

From 1974 to ~2004, the semiconductor industry operated under a gift: Dennard scaling. As transistors shrank, power density stayed constant — meaning every node shrink delivered smaller, faster, *and* cooler chips. Engineers could pretend physics was cooperative.

Then, around 2004–2005, Dennard scaling broke.

**Why it broke:** As transistors approached ~90nm and below, leakage current exploded. Electrons tunnel through gate oxides even when the transistor is supposed to be "off." The smaller the transistor, the worse the leakage. Power density stopped being constant and started rising — fast.

**Consequence: The Power Wall**
- Clock speeds plateaued around 3–4 GHz circa 2005 and have barely moved since
- The chip industry pivoted from single-core frequency scaling to multi-core parallelism
- This pivot was the original forcing function for GPU compute: GPUs are massively parallel architectures born from the death of Dennard scaling

**Investment implication:** Any company whose competitive moat depends on "we'll just run it faster next year" is not a real moat. Clock speed is dead. The companies that understood this early (Nvidia, 2006–2010) and pivoted to parallel compute architectures built 15-year advantages.

---

## The Power Wall and Architecture Diversity

The death of Dennard scaling created the era of **heterogeneous computing** — the recognition that no single processor architecture is optimal for all workloads. This is not a trend. It is physics.

**Power budget constraints in modern AI systems:**
- A single H100 GPU: ~700W TDP
- An H100 NVL server (8x GPU): ~10.2kW
- A full rack of H100s: ~40–50kW
- A hyperscaler AI data center pod: 5–20MW

At these power densities, **cooling becomes architecture**. Liquid cooling transitions from optional to mandatory above ~40kW/rack. Air-cooled data centers built in the 2010s are being retrofit or abandoned for AI workloads.

**Architecture diversity created by the power wall:**
- CPUs: General purpose, optimized for latency and branch prediction
- GPUs: Massively parallel, optimized for throughput on regular matrix operations
- TPUs/ASICs: Purpose-built for specific operations (matrix multiply for transformers), most efficient per operation
- FPGAs: Reconfigurable, useful for inference at specific quantization levels
- Memory-compute integration (PIM/CIM): Moves compute to where data lives, sidesteps memory bandwidth wall

Each of these architectures has a physics-constrained niche. The mistake investors make is assuming one architecture will dominate. The correct view is that the AI era accelerates architectural specialization.

---

## Process Nodes: What the Numbers Actually Mean

Node naming became marketing around 2015. "3nm" does not mean transistor gates are 3 nanometers wide.

### How to read node names correctly:

| Node Label | Actual Gate Length Equiv. | Key Vendor | Volume Ramp Timeline |
|------------|--------------------------|------------|----------------------|
| N3B (3nm) | ~18nm FinFET equivalent | TSMC | Ramped 2023 |
| N3E (3nm enhanced) | Improved N3B, better yield | TSMC | 2024 production |
| N2 (2nm) | GAA nanosheet transistors | TSMC | 2025 target |
| 18A | Intel's GAA + backside power | Intel | 2025 target |
| SF3 (3nm) | GAA | Samsung | Limited yield |

**Gate-All-Around (GAA):** The critical next architecture after FinFET. Surrounds the channel on all four sides, better electrostatic control, enables scaling below 3nm. TSMC moves to GAA at N2. Intel claims to be there now with 18A. Samsung's GAA yield has been poor.

**Yield is everything:** A node that yields 30% is economically useless. TSMC's consistent execution advantage is not just technology — it is yield engineering accumulated over 30 years. Samsung's GAA yield challenges in 2023–2024 cost it major customers. Intel's 18A yield is the central question for its foundry strategy.

**For investors:** When a CEO announces "we're moving to 2nm," the question is not "can they tape out a chip?" — it's "what will the yield be at volume?" A 30% yield improvement can swing gross margins by 15+ points.

---

## The Packaging Revolution

As classical scaling slowed, the industry found a new dimension: **advanced packaging**. Instead of putting more transistors on one die, connect multiple dies together as if they were one.

### Key packaging technologies:

**CoWoS (Chip-on-Wafer-on-Substrate)**
- TSMC's proprietary interposer technology
- Allows dies to communicate at near-monolithic bandwidth without going off-chip
- Critical for Nvidia GPUs: the H100 uses CoWoS-S; H200 and Blackwell use CoWoS-L (larger interposer)
- **Supply constraint 2023–2024:** CoWoS capacity was the binding constraint on H100 supply, not chip production. TSMC's CoWoS capacity became more valuable than the chip dies themselves.

**SoIC (System-on-Integrated-Chips)**
- TSMC's 3D stacking technology (face-to-face die bonding)
- Enables stacking logic dies on top of each other, dramatically reducing interconnect distance
- Used in Apple's M-series chips (stacking DRAM controllers)
- Next-generation AI accelerators will use SoIC for compute-memory integration

**HBM (High Bandwidth Memory)**
- Not just faster DRAM — a fundamentally different architecture
- Stacks DRAM dies vertically using through-silicon vias (TSVs)
- Achieves memory bandwidth of 3.2–4.8 TB/s (vs ~100 GB/s for DDR5)
- **Why AI needs it:** Transformer attention mechanisms are memory-bandwidth limited, not compute limited. HBM is not optional for training large models.
- Current HBM supply: SK Hynix (~50% share, technology leader), Samsung (~30%), Micron (~20%)
- HBM3E (the current generation): SK Hynix is 6–12 months ahead of Samsung on yield and production volume

**Backside Power Delivery**
- Traditional: power arrives at the chip through the same layers as signals — creating interference and resistance
- Backside: power routed through the back of the wafer, freeing the front for logic/signal
- Intel's 18A includes this. TSMC's N2P will include it.
- **Effect:** ~10–15% power reduction or equivalent performance boost at the same power budget

---

## TSMC's Monopoly Position

TSMC is the most important company in the global technology supply chain that most investors underweight.

### Why the monopoly exists:

**1. Accumulated yield learning:** TSMC has run more leading-edge wafers than any other foundry by an order of magnitude. Yield engineering is empirical — it requires data. TSMC's data advantage is 30 years deep.

**2. Customer concentration creates virtuous cycle:** Apple, Nvidia, AMD, Qualcomm, and now Google and Amazon all use TSMC for leading-edge. Their combined volume funds TSMC's R&D. No competitor can match this.

**3. Equipment co-development:** TSMC co-develops process technology with ASML, Applied Materials, Lam Research. The equipment is not commercially available to competitors in the same configuration — TSMC works with vendors on a next-node basis.

**4. ASML EUV as a choke point:** ASML is the sole manufacturer of EUV (Extreme Ultraviolet) lithography machines, which are required for leading-edge nodes (sub-7nm). Each machine costs ~$200M and takes 2 years to build. TSMC has ~60% of installed EUV capacity globally. Samsung and Intel are catching up, but the gap is structural.

### Competitive landscape:

| Foundry | Leading-Edge Capability | Yield Quality | Volume |
|---------|------------------------|---------------|--------|
| TSMC | N3/N2/A16 | Excellent | Dominant |
| Samsung | SF3/SF2 | Poor-Moderate | Limited |
| Intel Foundry | 18A (GAA) | Unproven | Minimal |
| SMIC (China) | ~7nm equivalent | Below leading edge | China-only |

**Investment implication:** Companies that are sole-sourced to TSMC for leading-edge have a fundamental supply chain dependency that is also a competitive moat. No one else can make their chip. This is simultaneously a risk (Taiwan) and a feature (no competitor can easily replicate).

---

## "Everything is Malleable Except Physics"

Eric's central investment heuristic deserves elaboration.

**Software is malleable:** An application can be rewritten in 6 months. A model can be retrained in weeks. A cloud service can be deprecated and relaunched. Software moats exist but they erode faster than engineers admit.

**Silicon is not malleable:**
- A new semiconductor fab takes 3–5 years and $15–20B to build
- A new process node takes 5–7 years of R&D before volume production
- Packaging capacity (CoWoS, SoIC) expands at 18–24 month intervals
- HBM capacity is supply-constrained 12–18 months forward

This asymmetry defines the investment landscape:
- **Short**: Companies whose moat is software-based in markets where silicon players are moving up the stack
- **Long**: Companies that control the physical infrastructure for which there is no fast substitute

**Case study — Nvidia 2022–2024:**
The thesis was not "Nvidia has great software (CUDA)." The thesis was: Nvidia controls the CoWoS allocation (packaging), the HBM allocation (memory), and the only trained team of GPU microarchitects in existence. These cannot be replicated in 18 months. A software competitor (like PyTorch optimized for alternative hardware) can emerge. A physical competitor cannot. The stock 5x'd.

---

## Practical Checklist: Physical Layer Due Diligence

Before underwriting any semiconductor or AI infrastructure investment:

- [ ] What process node does the flagship product use? Who manufactures it?
- [ ] What is the packaging technology? Is CoWoS or SoIC required?
- [ ] What is the power envelope per chip / per rack? Does it require liquid cooling?
- [ ] What is the HBM generation and who supplies it? Is supply committed?
- [ ] What is the yield assumption in the financial model? What does management guide?
- [ ] When is the next node transition? What is the realistic timeline?
- [ ] Can the product roadmap be replicated at an alternative foundry? At what cost and time?
- [ ] Is ASML EUV access a gating factor for any competitor?
- [ ] What happens to the economics if TSMC raises prices 15%? (They do this every 1–2 years.)
- [ ] Is backside power delivery part of the next-generation roadmap? What's the timing?

If you cannot answer all of these, you don't know the physical layer well enough to make a conviction call.
