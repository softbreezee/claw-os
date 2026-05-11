# Reference 05 — AI Infrastructure: The Hardware Stack in Detail

> "Investors talk about 'AI chips' as if they're a monolith. There is no monolith. There's a specific GPU, a specific HBM generation, a specific packaging process, a specific networking fabric — and each of those is a separate competitive battle."
> — Eric Sanchez

---

## GPU Architecture Deep Dive

### Nvidia Blackwell Architecture (B100/B200/GB200)

**Released:** 2024 (production ramp Q3–Q4 2024)
**Process node:** TSMC N4P (4nm class)
**Die configuration:** Two B100 dies connected via NVLink-C2C on a single package (multi-chip module)
**Key specs:**
- B200: 208 TFLOPS (FP16), 208 TFLOPS (FP8 for training)
- Memory: 192GB HBM3E (on GB200 NVL72 configurations)
- Memory bandwidth: 8 TB/s per B200 (HBM3E)
- TDP: ~700W (B200 standalone), ~1,200W (GB200 NVL module)
- Interconnect: 5th gen NVLink, 1.8 TB/s per GPU

**Critical advance over Hopper (H100):** The Transformer Engine in Blackwell natively supports FP8 precision with dynamic format switching — allowing training at FP8 accuracy (faster) with FP16 fallback where precision matters. Inference throughput improves ~4× on Blackwell vs. H100 for equivalent workloads.

**GB200 NVL72:** The full rack configuration — 36 Grace CPUs + 72 B200 GPUs, connected in a unified memory pool via NVLink-C2C. This is the compute unit that will train GPT-5 class models. It requires ~120kW per rack, making liquid cooling mandatory.

**Supply chain dependency:** Blackwell production requires:
- TSMC N4P wafers (sole source)
- CoWoS-L packaging (TSMC, capacity constrained)
- HBM3E from SK Hynix (allocated 12 months forward)
- Custom NVLink switches (Nvidia design, manufactured at TSMC)

**Investment implication:** Blackwell is the single most important product in the AI infrastructure buildout. Any supply constraint in any of these four supply chain elements creates a revenue shortfall — and tracking CoWoS and HBM3E allocation is more predictive than Nvidia's own guidance.

---

### AMD MI300 Series (MI300X / MI325X / MI350X)

**MI300X (2024):** AMD's first serious data center GPU for AI training/inference
- Architecture: CDNA3 (compute die) + HBM3 (192GB) in chiplet MCM design
- Memory bandwidth: 5.3 TB/s (highest per-GPU at launch)
- TDP: ~750W
- Key advantage: 192GB unified memory pool — enables running larger models in-memory without offloading (critical for inference with long context windows)
- TSMC N5 die for compute, N6 for I/O

**MI325X (late 2024):** Updated HBM to HBM3E (256GB, 6 TB/s bandwidth)

**MI350X (2025):** CDNA4 architecture, TSMC N3, targeting training parity with Blackwell

**AMD's competitive position:**
- Training: Still 2–3 generations behind Nvidia in software ecosystem (ROCm vs. CUDA)
- Inference: Much more competitive. MI300X is preferred by some hyperscalers for large-model inference due to the 192GB memory pool reducing the number of chips needed per inference server
- Key customers: Microsoft, Meta, Oracle (for inference workloads specifically)

**The ROCm vs. CUDA battle:**
- CUDA: 15+ year ecosystem, optimized libraries (cuDNN, NCCL, TensorRT), developer muscle memory
- ROCm: 5 years of serious development, improving rapidly, HIP abstraction layer allows CUDA code to run on ROCm with modification
- Reality: For new inference deployments where customers are willing to port code, AMD can compete. For existing training pipelines, switching cost is 12–18 months of engineering work.

**Investment thesis on AMD:** Not a Nvidia replacement story. An inference-layer value option that captures 15–25% of AI compute spend as inference workloads diverge from training. AMD's semiconductor operations are competent — Lisa Su's execution track record is exceptional.

---

### Google TPU Ironwood (v6)

**Architecture philosophy:** Google builds TPUs because it concluded that Nvidia's GPU is not optimally designed for transformer inference — it's designed for general parallel compute, with transformers as a secondary use case. TPUs are designed from first principles for matrix operations only.

**TPU Ironwood (v6) specs (announced 2024):**
- 4,614 TFLOPS per chip (for AI workloads)
- Designed for inference at scale, not training
- Pods: 9,216 TPUs interconnected via custom Google network fabric
- Power: ~100W per chip (much lower than GPU equivalent)

**TPU competitive advantage:**
- Inference per watt: ~3–5× more efficient than H100 equivalent for specific workloads
- Cost per token: Google's internal TPU cost is structurally lower than purchasing H100 equivalents
- Latency: TPU pods have ultra-low interconnect latency, enabling larger batch sizes

**Investment implication:** TPUs are why Google Cloud is competitive on AI inference pricing despite weaker enterprise sales. Google does not sell TPU access to third parties at scale (limited v4/v5 access on GCP). This limits the investable opportunity — but it signals that Nvidia's inference dominance is not guaranteed at hyperscaler scale.

---

## HBM: The Memory Bottleneck

High Bandwidth Memory is not a commodity product. It is the most supply-constrained component in the AI stack after the GPU itself.

### HBM Technology Generations

| Generation | Bandwidth | Capacity | Key Buyer | Availability |
|------------|-----------|----------|-----------|--------------|
| HBM2e | 3.2 TB/s per stack | 16GB | A100 (legacy) | Surplus |
| HBM3 | 4.8 TB/s per stack | 24GB | H100, MI300 | Available |
| HBM3E | 4.8–5.6+ TB/s per stack | 24–36GB | H200, B200, MI325X | Constrained |
| HBM4 | ~6.4+ TB/s per stack | 32–48GB | Rubin/next-gen | 2025+ |

**Why HBM matters:** Transformer models are memory-bandwidth bound, not compute bound, for most production inference workloads. More HBM bandwidth = higher throughput for the same number of chips.

### HBM Competitive Landscape

**SK Hynix (~50% market share, technology leader):**
- First to mass produce HBM3E (6-month lead over Samsung)
- Sole supplier of HBM3E to Nvidia for H200 and B200 launch
- Manufacturing advantage: superior Through-Silicon Via (TSV) process
- Valuation implication: SK Hynix commands premium multiple vs. Samsung on HBM exposure

**Samsung (~30% market share, technology follower):**
- Late to HBM3E qualification for Nvidia (yield/reliability issues in 2023–2024)
- Has advanced DRAM manufacturing at 1c nm (cutting-edge DRAM) but TSV integration lagging
- Incentivized to regain HBM share — pricing more aggressively
- Risk: If Samsung fixes its HBM3E yield and captures more Nvidia allocation, SK Hynix faces pricing pressure

**Micron (~20% market share, fast follower):**
- US-based, CHIPS Act beneficiary ($6.1B in proposed grants)
- HBM3E qualification achieved in 2024, ramping at Nvidia
- Strategic value as a geopolitically secure source (US-headquartered, non-Korea supply chain)
- Long thesis: Micron's HBM share grows from 20% to 25–30% over 3 years, mix shift improves margins

**Supply constraint dynamics:**
- HBM requires dedicated DRAM capacity (cannot easily convert standard DRAM lines to HBM)
- TSV bonding is a specialized process — not all DRAM fabs have the equipment
- New HBM capacity additions take 18–24 months from announcement to production
- 2024–2025: HBM capacity is fully allocated. GPU shipment volumes are gated by HBM availability.

---

## Networking: The Invisible Bottleneck

A 10,000-GPU cluster is only as fast as the network connecting it. As cluster sizes grow, networking becomes an increasingly binding constraint on GPU utilization efficiency.

### InfiniBand vs. AI Ethernet

**InfiniBand (Nvidia/Mellanox):**
- Designed for high-performance computing, adopted by AI clusters
- Ultra-low latency: ~1 microsecond end-to-end
- High bandwidth: NDR InfiniBand = 400Gb/s per port; XDR = 800Gb/s (2025)
- Proprietary: Only Nvidia/Mellanox manufactures InfiniBand equipment
- Used by: OpenAI, Anthropic, most Tier 2 cloud providers
- Cost: ~$10,000–15,000 per port at full build-out

**AI Ethernet (Ultra Ethernet Consortium):**
- Industry coalition (AMD, Intel, Broadcom, Microsoft, Meta, Google) to build AI-optimized Ethernet
- Goal: Match InfiniBand latency at lower cost using commodity Ethernet infrastructure
- Technical challenge: Ethernet's TCP/IP stack has too much latency for tight GPU synchronization; UEC adds RDMA and congestion management
- Status: First generation specifications released 2024; deployments expected 2025–2026
- Strategic driver: Hyperscalers want to reduce dependence on Nvidia networking monopoly

**Investment implication of the networking battle:**
- If InfiniBand maintains dominance: Nvidia captures networking revenue in addition to compute
- If AI Ethernet wins: Broadcom (the dominant Ethernet chip designer), Arista (switches), and Marvell (PHY chips) benefit; Nvidia loses networking revenue stream

**Broadcom's positioning:** Broadcom's custom ASIC business (Google TPU networking, Meta AI fabric) and its Ethernet chip dominance make it the most balanced networking winner regardless of InfiniBand vs. Ethernet outcome.

**Arista Networks:** The dominant switch provider to hyperscalers. Every AI cluster (whether InfiniBand or Ethernet at the higher levels) needs Arista switches for the spine layer. Revenue growth from AI build-out is structural and largely independent of the compute architecture debate.

---

## Data Center Architecture: Power and Cooling

AI workloads are rewriting data center design requirements in ways that create multi-billion dollar infrastructure obsolescence risk.

### Power Density Evolution

| Era | Rack Power Density | Cooling Method |
|-----|-------------------|----------------|
| Traditional IT (2010s) | 5–10 kW/rack | Air cooling |
| GPU servers (2020–2022) | 10–25 kW/rack | High-density air cooling |
| H100 clusters (2023) | 30–50 kW/rack | Hybrid air + direct liquid |
| Blackwell GB200 NVL72 (2024) | 100–120 kW/rack | Full direct liquid cooling (DLC) |
| Next-gen (2026+) | 150–200 kW/rack | Liquid immersion or advanced DLC |

**The cooling inflection point:** Air cooling becomes economically and physically impractical above ~50kW/rack. The transition to direct liquid cooling (DLC) — where liquid coolant flows directly over chip packages — is now a capex prerequisite for AI data centers.

**Winners from the power/cooling transition:**
- Vertiv Holdings: Data center cooling infrastructure (precision cooling, thermal management)
- Eaton, Schneider Electric: Power distribution units, UPS systems
- Custom DLC vendors (CoolIT, Asetek): Direct liquid cooling loops

**Data center power dynamics:**
- Each 1 MW of AI compute requires ~1.5–2 MW of total power (including cooling, networking, UPS)
- A 100MW AI data center campus: ~$800M–1.2B capital cost
- Power availability (grid connection, substation capacity) is now a gating factor for AI data center expansion — not just real estate or permits

---

## Custom Silicon: The Hyperscaler AI ASIC Race

All four major hyperscalers have internal silicon programs explicitly designed to reduce Nvidia dependence and optimize cost-per-inference.

### Google TPU Program
- **TPU v4:** Used for internal training and GCP offering
- **TPU v5e:** Optimized for inference, available on GCP
- **TPU v6 (Ironwood):** Next-generation, inference-optimized
- **Strategic goal:** Reduce the cost of serving Gemini models by 5–10× vs. GPU equivalents

### Amazon Trainium / Inferentia
- **Trainium 2:** Training chip, clusters of 100,000+ units deployed internally
- **Inferentia 2:** Inference chip, deployed for AWS Bedrock inference services
- **Designed by:** Annapurna Labs (Amazon subsidiary), manufactured at TSMC
- **Position:** Amazon uses its own chips for Amazon models; still uses Nvidia for external model training workloads

### Microsoft Maia / Cobalt
- **Maia 100:** AI accelerator for Azure AI training workloads
- **Cobalt:** Arm-based CPU for general Azure compute
- **Status:** Limited deployment as of 2024; Microsoft still massively buys Nvidia

### Meta MTIA (Meta Training and Inference Accelerator)
- Custom inference chip for Meta's recommendation systems and Llama inference
- Reduces inference cost for Meta's 3B+ daily active users at scale
- Not offered externally (internal only)

**Investment implication of custom silicon:**
- Each hyperscaler chip reduces future Nvidia total addressable market by the volume it displaces
- Near-term (2024–2026): Too early to meaningfully dent Nvidia. Custom chips have limited software ecosystems.
- Medium-term (2026–2028): Custom inference chips take 15–25% of hyperscaler inference workloads
- Beneficiary: Broadcom — all of these custom chips use Broadcom networking, and Broadcom designs some of the custom ASICs itself (Google TPU silicon partnership)
- Risk to monitor: How many of Nvidia's hyperscaler units are "truly contested" by custom silicon?

---

## Infrastructure Investment Priority Matrix

| Component | Current Value Accrual | Duration | Key Player(s) | Investable? |
|-----------|-----------------------|----------|----------------|-------------|
| Training GPUs | Extreme | 2–4 years dominant | Nvidia | Yes (priced in) |
| HBM Memory | Very High | 3–5 years structural | SK Hynix, Micron | Yes |
| TSMC Foundry | Very High | Decade+ | TSMC (ADR: TSM) | Yes |
| CoWoS Packaging | High | 2–4 years constrained | TSMC | Via TSMC exposure |
| Inference chips | Growing | 2025–2028 inflection | AMD, Broadcom | Yes |
| AI networking | High | Multi-year build | Arista, Broadcom, Marvell | Yes |
| Data center power/cooling | Growing | Decade transition | Vertiv, Eaton | Yes |
| Custom silicon | Eventual threat | 2026+ impact | Broadcom (design wins) | Via Broadcom |
| Server OEMs | Low | Commodity risk | Avoid SMCI | Short candidate |
