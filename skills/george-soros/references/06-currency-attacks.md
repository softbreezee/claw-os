# Currency Peg Analysis and the Mechanics of Speculative Attacks

> When to read this file: when analyzing an economy with a fixed exchange rate or currency peg, when trying to identify whether a currency regime is sustainable, when studying the mechanics of how speculative attacks on currencies work, when considering a trade that profits from currency devaluation or peg abandonment, or when evaluating sovereign credit in countries with fixed exchange rate regimes.

George Soros's most famous trades — the pound in 1992, the baht in 1997, and related positions in Asian currencies — all involved identifying a currency peg that was fundamentally inconsistent with domestic economic needs and then positioning for the inevitable break. These are not arbitrary speculative bets; they are precise applications of economic analysis combined with reflexivity theory.

> "Speculative attacks on currencies are not the cause of currency crises — they are the mechanism by which pre-existing fundamental imbalances are resolved. Governments blame speculators, but the speculators are simply recognizing what the fundamentals already show: that the current exchange rate cannot be maintained." — George Soros, *The Crisis of Global Capitalism* (1998)

---

## The Mundell Trilemma: The Fundamental Framework

Every currency peg analysis begins with the Mundell-Fleming Trilemma (also called the "Impossible Trinity"). A country cannot simultaneously maintain all three of:

1. **Fixed exchange rate** — pegging the currency to another currency or basket
2. **Free capital flows** — allowing money to move in and out without restriction
3. **Independent monetary policy** — setting interest rates based on domestic economic needs

A country can choose any **two** of these three, but not all three. The choice of which constraint to sacrifice defines the vulnerability of the system:

| Regime | What's Fixed | What's Sacrificed | Example |
|--------|-------------|------------------|---------|
| Currency board | Rate + Capital flows | Monetary independence | Hong Kong vs. USD |
| ERM/peg with controls | Rate + Monetary policy | Capital flows | China partially |
| Inflation targeting, float | Monetary policy + Capital flows | Fixed rate | US, UK post-1992 |
| ERM pre-1992 (the flawed choice) | Rate + Capital flows (partially) | Monetary policy | UK 1990-1992 |

**The critical insight for attack analysis:** When a country tries to maintain all three — particularly when it pegs the exchange rate, allows capital flows, AND tries to set domestic interest rates based on its own needs — it has created a structurally unsustainable system. The attack is not a matter of "if" but "when" and "how."

> "The ERM was fundamentally flawed because it required participating countries to surrender monetary policy without surrendering fiscal policy or building compensating mechanisms. When Germany needed high interest rates (for reunification) and the UK needed low rates (for recession), the system was on borrowed time." — Soros, *Soros on Soros* (1995)

---

## The Soros Framework for Identifying Unsustainable Pegs

A currency peg is potentially attackable when the following conditions are present:

### Condition 1: Fundamental Misalignment
The exchange rate is inconsistent with the economy's competitive position:
- **Current account deficit**: Running a large, persistent current account deficit means the country is consuming more than it produces and paying the difference by borrowing in foreign currency or depleting reserves.
- **Real exchange rate overvaluation**: If domestic inflation has exceeded trading partner inflation while the nominal rate is fixed, the real exchange rate has appreciated — exports become uncompetitive, imports cheap, and the current account deficit widens.
- **Unit labor cost divergence**: If domestic wages have risen faster than productivity relative to trading partners, competitiveness has deteriorated.

**Measurement:**
- Compare domestic CPI with trading partner CPI over the life of the peg
- Calculate the real effective exchange rate (REER)
- Compare to purchasing power parity estimates
- Track current account balance trend

### Condition 2: Incompatible Monetary Policy Requirements
The interest rate required to defend the peg is inconsistent with domestic economic conditions:
- **Recession + high rates**: Defending a peg may require high interest rates that deepen recession (UK 1992, Ireland 2009-2011)
- **Inflation + low rates**: Allowing inflation to erode the peg's value requires rates that are too low to defend it
- **Banking system fragility**: High interest rates to defend a currency peg can simultaneously destroy the banking system through loan defaults (Thailand 1997)

> "The Bank of England had to choose between the pound and the British economy. They could not have both. I was simply recognizing that when forced to choose, they would — as any rational government would — choose the economy." — Soros, interviews

### Condition 3: Declining Reserve Position
The central bank's ability to defend the peg is ultimately limited by its foreign exchange reserves:
- Reserves must be large enough relative to the potential attack (a small central bank with few reserves is more vulnerable)
- Reserves should be measured against: monthly imports, short-term foreign debt, and the potential capital outflow (not just the trade balance)
- **Key threshold**: When reserves fall below 3 months of imports, the system is in acute danger
- **Hidden reserves**: Some central banks use forward contracts to defend currencies without depleting visible spot reserves (as Thailand did in 1997) — these create larger-than-visible vulnerabilities

### Condition 4: Political Willingness Under Pressure
The most important and hardest-to-assess condition:
- Can the government withstand the political pain of defending the peg (high interest rates, recession, austerity)?
- Is there a credible political commitment to the peg, or is it maintained by inertia?
- Would abandoning the peg actually serve the government's political interests?

> "In the end, currency crises are political crises. The economics may make a peg unsustainable, but the timing of the break is determined by the political calculus of the government. When a government finally concludes that maintaining the peg hurts it more than abandoning it, they abandon it — usually very quickly after that conclusion is reached." — Soros, various interviews

### Condition 5: Market Positioning and Sentiment
The technical setup for an attack:
- Are other institutional investors and hedge funds aware of the vulnerability?
- Is there a natural "herd" that will join an attack once it begins? (This is the reflexive element of currency attacks: once the attack starts, it can become self-fulfilling)
- What is the current level of speculative positioning (visible through forward market, options market, prime broker data)?

---

## The Attack Mechanics

When all five conditions are met, the speculative attack follows a predictable structure:

### Phase 1: Accumulation (Before Public Knowledge)
- Investors identifying the opportunity begin building short positions in the currency
- This is done through: forward contracts to sell currency, options (puts on the currency), borrowing in the local currency to convert to foreign, or directly shorting currency futures
- The attack is quiet: positions are built gradually to avoid alerting the central bank

### Phase 2: Initial Attack and Defense
- At some point, the selling pressure becomes large enough to put visible pressure on the peg
- The central bank intervenes: it buys its own currency using foreign exchange reserves
- The central bank may also raise interest rates to make shorting the currency more expensive (increasing the carry cost of the short)
- **This is where the carry cost calculation is critical** (see below)

### Phase 3: The Reflexive Loop Begins
This is where Soros's framework makes the attack particularly powerful:
- **Visible pressure on the peg raises doubts about its sustainability**
- **Doubt causes capital outflows by domestic investors and corporations** (hedging their foreign currency exposure)
- **Capital outflows create more pressure on the peg**
- **More pressure creates more doubt**

The domestic actors (corporations with foreign currency debt, domestic banks, wealthy individuals) are often the most powerful force in this reflexive loop — their actions dwarf those of foreign speculators.

### Phase 4: The Carry Calculation — The Economic Logic of the Attack

The attacker's calculation:

```
Expected profit = (probability of devaluation × size of devaluation) - (carry cost × time)

Where:
- Carry cost = interest rate differential (local rate - foreign rate) × position size × time
- Time = how long until devaluation occurs
```

**Example: UK 1992:**
- Pound was pegged at DM 2.95
- UK interest rates: ~10%; German rates: ~7% → carry cost of ~3% per year
- If the pound devalued 20% when the peg broke → expected profit 20% on one-year carry cost of 3%
- Break-even: devaluation must exceed carry cost within the time horizon

**The asymmetry that makes the attack attractive:**
- The maximum loss for the attacker (if the peg holds indefinitely) is the accumulated carry cost, which is bounded
- The maximum gain if the peg breaks is the size of the devaluation, which can be 20-50% or more
- This asymmetry is the economic foundation of the attack: the attacker has limited downside and large upside

> "The beauty of a currency attack on a fundamentally unsustainable peg is that the asymmetry is favorable. If the peg holds, I lose a little carry. If it breaks, I make 20-30%. The central bank has the opposite asymmetry: it loses reserves defending the peg, and it wins nothing if it succeeds except the right to lose more reserves next time." — Soros, *Soros on Soros* (1995)

---

## Case Study: British Pound, September 1992

**Background:**
- UK entered ERM in October 1990 at DM 2.95 — widely considered overvalued
- Germany reunified in 1990 and was dealing with inflationary pressure, leading Bundesbank to keep rates high
- UK was in recession; needed lower rates
- UK's interest rate must match or exceed Germany's to maintain the peg; this is contractionary during recession

**Soros's Analysis (applied to the 5-condition framework):**

| Condition | UK 1992 Status |
|-----------|----------------|
| Fundamental misalignment | UK inflation had exceeded German throughout the 1980s; REER appreciated significantly |
| Incompatible monetary policy | UK needed 5-6% rates; defending ERM required 10%+ |
| Reserve position | UK reserves: ~$44B; exposure potential: far larger |
| Political willingness | PM Major had staked political reputation on ERM; Lamont under pressure from recession |
| Market positioning | Growing hedge fund community aware; Soros the largest but not alone |

**The trade:**
- Soros Fund built a $10B short position on the pound through forward contracts
- Added correlated positions: long D-marks, long French francs, short Italian lire, long UK gilts (which would rise if rates fell post-exit)

**The attack:**
- September 16, 1992 ("Black Wednesday"): Bank of England raised rates to 12%, then 15% — then reversed both hikes within hours (signaling political capitulation)
- BoE spent approximately £27B of reserves in a single day
- 7:00 PM: UK suspended ERM membership; pound fell 15-20% against D-mark

**The profit:** Quantum Fund made approximately $1 billion on this trade.

> "I didn't break the Bank of England. The Bank of England broke itself by maintaining an untenable position. I simply recognized that they were in an impossible situation and bet that they would eventually face reality." — Soros, interviews

---

## Case Study: Thai Baht, 1997

**Background:**
- Thailand pegged baht to dollar basket (effectively dollar-pegged) since 1984
- Ran large current account deficits (8% of GDP by 1996)
- Financed by hot money inflows attracted by high interest rate differentials
- Corporate sector borrowed heavily in dollars (unhedged) — a reflexive feedback loop
- Property bubble financed by Thai banks; banks' assets were dollar-funded (unhedged)

**The collapse mechanism:**
- Soros and other hedge funds identified the vulnerability: large CAD + dollar debt + fixed rate + property crash
- Began shorting baht through forward contracts in February 1997
- Bank of Thailand defended with spot reserves AND (crucially) forward contracts
- By June 1997, BoT had used all spot reserves and had $26B in forward commitments it could not honor
- July 2, 1997: Thailand floated the baht; it fell 40% within weeks

**The contagion:** The baht crisis revealed the same structure across the region — Malaysia, Indonesia, Philippines, South Korea all had variants of the same reflexive vulnerability.

---

## Checklist: Is This Currency Peg Attackable?

Score each condition 1 (absent) to 5 (severe):

| Condition | Score | Details |
|-----------|-------|---------|
| Real exchange rate overvaluation | /5 | |
| Current account deficit | /5 | (5 = >6% of GDP) |
| Incompatible monetary policy needs | /5 | |
| Reserve depletion trajectory | /5 | |
| Political willingness deteriorating | /5 | |
| Market positioning building | /5 | |
| **Total** | /30 | |

**Scoring:**
- 20-30: High-conviction attack candidate; build position
- 12-20: Monitor closely; begin small position; wait for confirmation
- Below 12: Not yet; peg is likely sustainable for now

---

## Modern Applications: Non-Classic Pegs

Currency attack analysis applies beyond traditional fixed-rate regimes:

- **Managed floats** (China): PBOC manages daily fixing; fundamental misalignment can still accumulate
- **Euro membership** (Greece, 2010-2015): Cannot devalue, so adjustment happens through deflation and default; attack plays out in sovereign credit spreads rather than currency
- **Currency boards** (Hong Kong): Very hard to attack due to fully-backed structure, but not impossible under extreme stress
- **Implicit pegs** (Saudi Arabia, Gulf states): Dollar-peg dependent on oil revenue; vulnerable when oil price collapses
