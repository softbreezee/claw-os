# Reference 03 — Position in Stack: The 8-Layer AI Value Chain

> "Tell me where a company sits in the stack and I'll tell you what its margins should be. The stack is a profit map."
> — Eric Sanchez

---

## The 8-Layer AI Value Chain

The AI value chain has a definable structure. Every company in the AI ecosystem occupies one or more layers. Value does not accrue uniformly — it concentrates at bottlenecks, at points of IP scarcity, and at layers where substitution is hardest.

```
┌─────────────────────────────────────────────────────────────────┐
│  Layer 8: AI Agents                                             │
│  (Autonomous agents, multi-step task completion, AI workers)    │
├─────────────────────────────────────────────────────────────────┤
│  Layer 7: Applications                                          │
│  (AI-native SaaS, vertical AI, copilot features, AI APIs)      │
├─────────────────────────────────────────────────────────────────┤
│  Layer 6: Cloud Infrastructure                                  │
│  (AWS, Azure, GCP — compute-as-a-service, AI PaaS)             │
├─────────────────────────────────────────────────────────────────┤
│  Layer 5: Systems Integration                                   │
│  (Server OEMs, rack integrators, data center buildout)          │
├─────────────────────────────────────────────────────────────────┤
│  Layer 4: Networking                                            │
│  (InfiniBand, Ethernet, switches, optical interconnects)        │
├─────────────────────────────────────────────────────────────────┤
│  Layer 3: Inference Chips                                       │
│  (Nvidia Hopper/Blackwell inference, AMD MI300, custom ASICs)   │
├─────────────────────────────────────────────────────────────────┤
│  Layer 2: Training Chips                                        │
│  (Nvidia H100/H200/B200, Google TPU v5, custom training ASICs)  │
├─────────────────────────────────────────────────────────────────┤
│  Layer 1: Foundation Models                                     │
│  (OpenAI, Anthropic, Google DeepMind, Meta AI, Mistral)         │
└─────────────────────────────────────────────────────────────────┘
```

*Note: Layers 1 and 2 are co-dependent (foundation model labs are the primary customers of training chips). Memory (HBM, DRAM) is horizontal — it spans Layers 2–5. TSMC and EDA tools are sub-Layer 2 — they are the foundation of the foundation.*

---

## Layer-by-Layer Analysis

### Layer 1: Foundation Models
**Players:** OpenAI, Anthropic, Google DeepMind, Meta AI, Mistral, Cohere, xAI

**Value accrual:** High but contested. Foundation models are the ultimate software layer — they determine what applications are possible. But the competitive dynamics are brutal: open-source models (Llama, Mistral) rapidly commoditize closed-model capabilities.

**IP scarcity:** Moderate and declining. GPT-4 was a significant capability lead. That lead narrowed within 12 months (Claude 3, Gemini Ultra, Llama 3). The moat is data quality + inference efficiency + developer ecosystem, not the architecture itself.

**Long/Short:** Long on specific models with durable distribution advantages (Microsoft/OpenAI ecosystem, Google Search integration). Short on models without distribution moats — capability alone is insufficient.

**Bottleneck dependency:** Layer 1 is entirely dependent on Layer 2 (training chips). When H100 supply was constrained in 2023, model training timelines extended — showing the power dynamic flows from hardware to software, not the reverse.

---

### Layer 2: Training Chips
**Players:** Nvidia (dominant), Google TPU v5 (internal), custom ASICs at hyperscalers, AMD MI300X (emerging)

**Value accrual:** Currently the highest-value layer in the stack. Nvidia's gross margins on data center GPUs (~75–80%) reflect near-monopoly pricing power. The H100 was priced at ~$30K; hyperscalers paid it because there was no alternative.

**IP scarcity:** Very high today. Nvidia's CUDA ecosystem (15+ years of developer tooling, optimized libraries, and workflows) is not a software moat — it is a network effect moat. Retraining the world's ML engineers to use an alternative framework is a 5–7 year process. AMD's ROCm framework has improved but trails by years.

**Bottleneck control:** YES — this is the current primary bottleneck. Every frontier model lab is constrained by H100/H200/B200 allocation, not by money, ideas, or data.

**Long conviction:** Nvidia is the most obvious long in AI. The question is not "is this a good business?" but "what is the right price?" At 30–35x forward earnings, it requires continued hyperscaler capex growth AND maintenance of the training chip bottleneck position.

---

### Layer 3: Inference Chips
**Players:** Nvidia (Hopper/Blackwell), AMD MI300X, Google TPU (inference), Amazon Trainium/Inferentia, Microsoft Maia, Broadcom custom ASICs, Qualcomm AI 100

**Value accrual:** Currently lower than training chips, but growing. As training cluster buildout matures, inference becomes the dominant workload. The inference market is 10–100× larger by unit volume than training.

**IP scarcity:** Moderate and competitive. Inference is less memory-bandwidth constrained than training (smaller KV caches, more flexible quantization). AMD MI300X is competitive for inference. Custom ASICs (Google TPU, Amazon Trainium) are designed for inference economics and are 2–3× more efficient per dollar than Nvidia for specific workloads.

**Bottleneck migration:** The key insight is that today's training bottleneck (Layer 2) will migrate to Layer 3 as inference deployment scales. The inference TAM expansion begins in 2025–2026.

**Best risk/reward:** Broadcom — designs custom inference ASICs for hyperscalers (Google TPU, Meta MTIA, unknown third customer). Pure design company (fabless), high margins, benefits from hyperscaler desire to reduce Nvidia dependence.

---

### Layer 4: Networking
**Players:** Nvidia (InfiniBand + NVLink), Broadcom (Ethernet custom chips), Arista Networks, Juniper/HPE, Marvell, Intel (Ethernet)

**Value accrual:** High and underappreciated by many generalist investors. A 10,000 GPU training cluster is useless if the GPUs cannot communicate fast enough. Network bandwidth is often the binding constraint on GPU utilization efficiency.

**IP scarcity:** High for InfiniBand (Nvidia/Mellanox monopoly), moderate for AI Ethernet (Broadcom dominant but competitive). The InfiniBand vs. AI Ethernet debate is crucial:
- InfiniBand: Purpose-built for AI clusters, ultra-low latency, higher cost, Nvidia monopoly
- AI Ethernet (Ultra Ethernet Consortium): Industry effort to build AI-optimized Ethernet, backed by AMD/Intel/Broadcom/hyperscalers as a strategic alternative to InfiniBand dependence

**Key dynamic:** Hyperscalers are strategically motivated to develop InfiniBand alternatives. Meta built its own fabric. Google uses its own TPU network. If Ethernet-based AI fabric matures, Nvidia loses a networking revenue stream.

**Long:** Arista Networks — positioned on both sides (wins regardless of InfiniBand vs. Ethernet outcome), dominant switching position in hyperscale data centers, expanding into AI networking.

---

### Layer 5: Systems Integration
**Players:** Super Micro Computer (SMCI), Dell Technologies, HPE, Foxconn, Wistron, Inspur (China)

**Value accrual:** Low to moderate. System integrators are primarily assembly and logistics plays. Margins are structurally compressed because components (GPU, HBM, networking) are the value, not the box.

**IP scarcity:** Low. System integration is easily replicated. SMCI's advantage was being first to AI-optimized servers with liquid cooling — but Dell and HPE are closing the gap.

**Short candidate:** SMCI is the clearest short in the systems integration layer. Benefits from AI server demand (real) but: accounting concerns, gross margins structurally low (~14–16%), competitive position is not durable, management quality concerns. Multiple re-ratings likely.

**Watch:** Dell has a more defensible position due to direct salesforce relationships with enterprise customers — AI server demand is one path for Dell's services business to expand.

---

### Layer 6: Cloud Infrastructure
**Players:** AWS, Azure, Google Cloud, Oracle Cloud Infrastructure (OCI)

**Value accrual:** High and durable. Hyperscalers are simultaneously the biggest customers of the AI stack (Layers 1–5) and the distribution layer for AI to enterprise customers. They sit at a strategic choke point.

**IP scarcity:** High. The combination of global data center footprint, enterprise relationships, and compliance infrastructure (SOC2, FedRAMP, HIPAA) is not replicable by startups.

**The hyperscaler AI dynamic:**
- AWS: Strongest in enterprise adoption, weakest in foundation model capability
- Azure: OpenAI partnership is a structural advantage; enterprise distribution dominates
- GCP: Best foundation model capability (Google DeepMind); weakest enterprise sales motion
- OCI: Niche position — lowest-cost GPU cloud, attracting training workloads that prioritize cost

**Investment consideration:** Hyperscalers' AI capex spending is what drives Layers 2–5 demand. Monitoring their guidance is the most important demand signal for the whole stack.

---

### Layer 7: Applications
**Players:** Salesforce, ServiceNow, Workday, GitHub/Microsoft, Adobe, Palantir, Snowflake, dozens of AI-native startups

**Value accrual:** Highly variable. Applications with proprietary data moats or deep workflow integration can sustain high multiples. Applications that are thin wrappers on foundation models (API resellers) will be commoditized.

**IP scarcity:** Bimodal distribution:
- High: Companies with proprietary training data (clinical records, legal databases, financial transaction data), deep enterprise integrations (ServiceNow, SAP), or network effects (GitHub Copilot)
- Low: Generic AI features added to existing software without data differentiation

**Key question for every application investment:** "Can this feature be replicated by the foundation model provider in 12 months?" If yes, the application layer value is borrowed, not owned.

---

### Layer 8: AI Agents
**Players:** Emerging (OpenAI Operator, Anthropic Computer Use, Salesforce AgentForce, Microsoft Copilot Studio, startups)

**Value accrual:** Unknown — this layer is pre-revenue for most players as of 2024–2025.

**IP scarcity:** To be determined. Agent frameworks are currently open-source and commodity. Durable agent moats will likely come from: task-specific training data, reliability track records in high-stakes domains, and enterprise trust/compliance positioning.

**Investment strategy:** Too early to underwrite with high conviction. Monitor for: first enterprise deployments with measurable ROI, reliability benchmarks, liability/insurance frameworks emerging.

---

## Bottleneck Migration Analysis

The key to positioning the portfolio correctly is anticipating where the bottleneck migrates over time.

### Current Bottleneck (2024): Training Compute
- H100/H200 allocation is constrained
- Hyperscalers ordering 12–18 months forward
- TSMC CoWoS packaging is the sub-bottleneck
- Winner: Nvidia, TSMC, SK Hynix (HBM3E)

### Next Bottleneck (2025–2026): Inference at Scale
- Training cluster buildout catches up to near-term demand
- Inference request volume grows exponentially as applications deploy
- Inference optimization (quantization, speculative decoding, KV cache efficiency) becomes critical
- Winner: AMD (inference parity better), custom ASICs (Google TPU, Trainium), Broadcom
- Loser: Nvidia loses pricing power on inference-optimized configurations; pure-play training chip demand cools relative to expectations

### Future Bottleneck (2027+): Data and Energy
- Compute cost per token falls dramatically (Wright's Law)
- High-quality, proprietary training data becomes the binding constraint
- Energy infrastructure (power generation, transmission, cooling) becomes the physical limit
- Winner: Data owners (unique datasets, proprietary workflows), energy infrastructure companies
- Loser: Pure compute providers without data moats

---

## Long/Short Framework by Stack Layer

| Layer | Long Candidates | Short Candidates | Key Risk to Monitor |
|-------|----------------|------------------|---------------------|
| 1 — Foundation Models | Microsoft (OpenAI dist.), Alphabet (search integration) | Pure-play model API companies without distribution | Open-source model parity |
| 2 — Training Chips | Nvidia | — (too expensive to short outright) | Custom ASIC displacement |
| 3 — Inference Chips | Broadcom, AMD | Nvidia on inference-specific thesis | Inference efficiency improvements reducing demand |
| 4 — Networking | Arista, Marvell | Juniper (inferior AI positioning) | Ethernet displacing InfiniBand |
| 5 — Systems Integration | — | SMCI (margin compression + governance) | Dell closing gap |
| 6 — Cloud Infrastructure | Microsoft Azure, AWS | GCP (go-to-market weakness) | Sovereign cloud fragmentation |
| 7 — Applications | ServiceNow, Palantir | Generic AI-feature SaaS | Foundation model providers moving up stack |
| 8 — AI Agents | Too early | Too early | Regulatory liability frameworks |
