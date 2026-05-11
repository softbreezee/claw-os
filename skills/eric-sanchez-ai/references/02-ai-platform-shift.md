# Reference 02 — The AI Platform Shift: First Principles, Not Analogy

> "Everyone keeps asking me: 'Is this like the internet? Is this like 5G?' My answer is: stop using analogies. Think from first principles. This is the industrialization of intelligence. Nothing like that has ever happened before."
> — Eric Sanchez

---

## The TAM Framing That Changes Everything

Most technology TAM analyses start with "the software market" or "the cloud market." Eric starts differently.

**The real TAM for AI is not IT spending. It is human labor.**

Current global IT spend: ~$3T/year
Current global labor cost: ~$50T/year (rough estimate of wages paid to knowledge workers)

If AI captures 10% of global knowledge worker productivity over 10 years, the incremental economic value is ~$5T — larger than the entire current IT industry. If it captures 30%, the numbers become almost incomprehensible.

**Why this framing matters for investing:**
- A $3T IT market with AI taking share is a zero-sum redistribution story — some vendors win, others lose
- A $50T labor market being augmented is an entirely new demand creation story — the market expands as AI penetrates
- The infrastructure required to augment $50T of labor is orders of magnitude larger than current IT infrastructure

**The investment implication:** We are not in a market share game. We are in a market creation event. Companies that provide the foundational infrastructure for that creation — compute, memory, networking — are in a structurally different position than typical IT vendors competing for existing budget.

---

## Why AI Is NOT the Dot-Com Bubble

The dot-com comparison is intellectually lazy. Here is why:

**Dot-com (1995–2000): A distribution shift**
- The internet was a new distribution channel for existing goods and services
- The TAM was "how much of retail, media, advertising moves online?"
- Infrastructure (fiber, routers) was massively overbuilt relative to near-term demand
- Companies built for distribution without revenue models — pets.com, Webvan
- **The bust happened because demand for the distribution layer was real but took 15 years to materialize, and equity was priced for 3 years**

**AI (2022–present): An intelligence shift**
- AI is not a new distribution channel — it is a new production function
- The TAM is "how much of human cognitive work can be automated or augmented?"
- Infrastructure (GPUs, HBM, data centers) is being absorbed as fast as it is built — there is no overbuilding as of 2024
- Companies are building on demonstrated revenue (GitHub Copilot, Google Search integration, enterprise SaaS)
- **The comparison fails because the demand is structural and the use cases generate measurable productivity**

**The 5G comparison also fails:**
- 5G was a latency/bandwidth improvement on an existing service (mobile connectivity)
- Users did not fundamentally change behavior — they continued calling, texting, browsing
- The "killer app" for 5G still hasn't emerged
- AI creates new categories of output that did not exist — code generation, protein folding, drug discovery, image synthesis

**Correct framing:** The closest historical analog is probably electricity (1890–1930) — a general-purpose technology that restructured the entire production system over 30 years, with enormous infrastructure investment front-loaded before the full economic value was visible.

---

## Scaling Laws: The Physics of Intelligence

The most important empirical discovery in AI since the transformer architecture (2017) is the **scaling law** — the relationship between compute, data, and model capability.

### The Chinchilla Scaling Law (2022, DeepMind)

Optimal training requires scaling model size and data size proportionally:
- **Pre-Chinchilla:** The field scaled model parameters aggressively (GPT-3: 175B params, 300B tokens)
- **Chinchilla:** A 70B parameter model trained on 1.4 trillion tokens outperforms GPT-3
- **Implication:** The field was underspending on training data and overspending on raw model size

For investors, Chinchilla meant: **more training compute per unit of model capability**, which increased HBM demand and extended the duration of the training compute bottleneck.

### Post-Chinchilla Progression

| Model | Est. Training Compute (FLOP) | Key Advance |
|-------|------------------------------|-------------|
| GPT-3 (2020) | ~3.1 × 10²³ | Demonstrated scale |
| GPT-4 (2023) | ~2 × 10²⁵ (est.) | Multimodal, reasoning |
| GPT-5 (2025, est.) | ~10²⁶–10²⁷ | Unknown |
| Gemini Ultra (2024) | ~3 × 10²⁵ (est.) | Native multimodal |

**Key observation:** Each generation requires roughly 10–50× more compute than the previous. This is not slowing. The GPU clusters required to train frontier models have grown from 1,000 A100s (GPT-3) to 16,000+ H100s (GPT-4) to potentially 100,000+ H100 equivalents for next-generation models.

**Investment implication:** Every frontier model generation expansion directly translates into hardware demand. This is not demand that can be served with last-generation chips — each training run requires the highest-bandwidth memory and highest-compute GPUs available.

---

## Compute Economics: Training vs. Inference

The AI infrastructure market is bifurcated. Most investors treat it as homogeneous. It is not.

### Training Economics

**What it is:** The process of adjusting model weights on massive datasets using gradient descent. Done once (or periodically for updates). Computationally intensive, memory-bandwidth intensive.

**Hardware requirements:**
- Maximum compute density (H100 SXM, future Blackwell)
- Maximum HBM bandwidth (HBM3E, 3.2+ TB/s)
- High-speed interconnect (NVLink for inter-GPU, InfiniBand for inter-node)
- Reliability is secondary — training can checkpoint and resume

**Cost trajectory:**
- Training GPT-3 cost ~$4M in 2020
- Training GPT-4 cost ~$60–100M (estimated) in 2022–2023
- Training frontier models in 2025 is projected at $200M–$1B+
- These costs are rising faster than compute efficiency gains — frontier model training is getting more expensive, not cheaper

**Who benefits from training economics:** Nvidia (dominant training chip), SK Hynix (HBM3E leader), TSMC (sole source for leading-edge), Arista/Nvidia (networking)

### Inference Economics

**What it is:** The process of running a trained model to generate outputs. Done continuously, billions of times per day. Different optimization target — latency and cost per token, not peak throughput.

**Hardware requirements:**
- Lower compute density than training (inference is sequential, not embarrassingly parallel)
- Lower HBM requirement per chip (can use DDR5 or LPDDR5 for some tasks)
- Extremely sensitive to batch size and quantization
- High volume, so unit cost matters enormously

**Cost trajectory (inference per token):**
- GPT-4 launch (2023): ~$0.06/1K tokens
- Late 2024: ~$0.002–0.01/1K tokens (10–30× reduction in 18 months)
- 2026 projection: potentially $0.0001/1K tokens
- This is Wright's Law in action on inference

**Who benefits from inference economics:** AMD (competitive position stronger in inference than training), custom silicon players (Google TPU, Amazon Trainium, Microsoft Maia — inference efficiency is their primary design goal), Broadcom (ASICs for hyperscaler inference workloads)

**The bottleneck migration:** Today, the AI infrastructure bottleneck is training compute (Nvidia GPU allocation). As training clusters are built out, the next bottleneck will shift to inference at scale — a different competitive landscape.

---

## Why "First Principles Not Analogy"

Eric's framework for avoiding the analogy trap:

**Step 1: What is the fundamental economic action?**
- Internet: Moving information from point A to B faster and cheaper
- 5G: Moving wireless data faster with lower latency
- AI: Generating novel cognitive outputs (text, code, images, decisions) from learned patterns

**Step 2: What does this change in the production function?**
- Internet: Did not replace workers, redistributed information access
- 5G: Made existing mobile workers more connected, did not replace workers
- AI: Can produce cognitive outputs previously requiring human cognition — this is replacement, not augmentation only

**Step 3: What infrastructure is physically required?**
- Internet: Fiber, DNS, routing protocols (one-time infrastructure build)
- 5G: Tower densification, spectrum allocation, small cells (network build)
- AI: Continuous compute demand that scales with usage — not a one-time build but a permanent capex stream

**Step 4: What is the competitive structure?**
- Internet infrastructure: Became commodity (Cisco was not the big winner long-term)
- 5G equipment: Became oligopoly (Ericsson, Nokia, Huawei)
- AI infrastructure: Structural monopoly at leading edge (TSMC), architectural monopoly (Nvidia CUDA), memory oligopoly (SK Hynix/Samsung/Micron for HBM)

The structural answer at Step 4 is what drives the investment thesis. AI infrastructure has much more durable monopoly positions than internet or 5G infrastructure.

---

## The Inference Transition: What Comes Next

Current market state (as of 2024): Training dominates capex conversations.

**Near-term (2025–2026):** Inference begins to dominate deployment economics. As models become "good enough" for production use cases (customer service, code generation, document processing), the relevant cost is not "how much to train the model" but "how much to serve a billion inference requests per day."

**Implications of the inference transition:**
1. Custom silicon becomes cost-competitive with Nvidia for inference workloads
2. AMD gains relevance — its inference performance/$ is closer to parity than its training performance/$
3. Memory mix shifts — HBM remains important, but DDR5 and LPDDR5 become viable for some inference hardware
4. Networking becomes more important — inference at scale requires low-latency routing at massive concurrency
5. Edge inference emerges — Qualcomm, Apple Silicon, Arm-based chips become relevant at the device layer

**Short opportunity in the inference transition:** Companies priced as if training capex dominates forever, whose competitive position is weaker in inference. This is not a Nvidia short — it is a "what adjacent companies get re-rated lower?" question.

---

## Sanity Check: What Would Break the AI Platform Shift Thesis?

Intellectual honesty requires maintaining the bear case:

1. **Scaling laws hit a wall:** If model capability stops improving with more compute (diminishing returns kick in before the labor TAM is penetrated), the infrastructure buildout stalls
2. **Energy economics become prohibitive:** At 100kW/rack and rising, power availability and cost could constrain deployment more than silicon supply
3. **Regulatory intervention:** Forced training data restrictions, liability for model outputs, or usage caps could compress the commercial TAM
4. **Security failures at scale:** A catastrophic AI security failure (e.g., widely deployed model manipulated for mass harm) could trigger a moratorium
5. **Enterprise adoption stalls:** If knowledge worker productivity gains are not measurable, enterprise buyers cut AI capex budgets

Eric's view: Items 1 and 2 are the real risks. The regulatory and security risks are real but unlikely to stop the platform shift — they will shape it. Item 5 is a timing risk, not a structural risk.
