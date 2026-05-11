# Principles-Based Decision Making

> When to read this file: when applying systematic decision-making, when building repeatable investment processes, when evaluating whether decisions are based on principles or emotions, or when discussing radical transparency, believability weighting, or Dalio's 5-step process.

> "Principles are fundamental truths that serve as the foundations for behavior that gets you what you want out of life." — Principles

---

## Radical Transparency

### The Core Philosophy

> "I believe that the biggest problem that humanity faces is an ego sensitivity to finding out whether one is right or wrong and identifying what one's strengths and weaknesses are." — Principles

Radical transparency at Bridgewater means:
- **All meetings are recorded** and available to anyone in the firm
- **Disagreements are surfaced, not buried** — arguing behind closed doors is prohibited
- **Feedback is direct and immediate** — including to the most senior people
- **Mistakes are studied openly** — as opportunities to learn, not occasions for blame

**Applied to investing:**
- State your thesis explicitly, including assumptions and confidence levels
- Track predictions against outcomes systematically
- When you're wrong, analyze why — openly and honestly
- Don't allow ego to prevent updating your views when evidence changes

> "Pain + Reflection = Progress. The quality of your life will depend on the quality of your decisions. And the quality of your decisions depends on your willingness to confront painful truths." — Principles

---

## Believability-Weighted Decision Making

### The Concept

Not all opinions are equal. Some people have more expertise, better track records, and better reasoning on specific topics. Dalio's system weights opinions by "believability":

> "The most believable opinions are those held by people who have (1) successfully accomplished the thing in question at least three times, and (2) have great explanations of the cause-effect relationships that led to their conclusions." — Principles

**In investment decisions:**
- A person who has correctly called three credit cycles is more believable on credit cycles than one who has called zero
- A person who can explain the causal mechanism (not just pattern-match) is more believable
- Believability is topic-specific — a great equity analyst may have zero believability on macro

### How to Apply Believability Weighting

1. For each relevant question, identify the most believable sources
2. Weight their opinions more heavily than less believable sources
3. If the believability-weighted consensus disagrees with your gut feeling, **seriously consider that you're wrong**
4. Ultimately, you make the decision — but with a clear understanding of where you agree and disagree with the best thinkers

---

## The 5-Step Process

Dalio's universal problem-solving framework:

### Step 1: Set Clear Goals
What are you trying to achieve? Be specific about:
- Return target (e.g., 7% real return)
- Risk tolerance (e.g., max 15% drawdown)
- Time horizon (e.g., 10+ years)
- Constraints (e.g., liquidity needs, regulatory requirements)

### Step 2: Identify Problems (Barriers to Goals)
What is preventing you from achieving your goals?
- Is the portfolio poorly diversified?
- Is there too much concentration in one risk factor?
- Are you paying too much in fees?
- Are you making emotional decisions?

### Step 3: Diagnose Root Causes
Don't just identify symptoms — find the root cause:
- The problem isn't "I lost money in 2008" — the root cause is "my portfolio had 90% equity risk"
- The problem isn't "I sold at the bottom" — the root cause is "I didn't have a systematic rule for when to sell"

### Step 4: Design Solutions (Plans)
Create specific, actionable plans:
- Rebalance to achieve equal risk contribution across environments
- Implement systematic rebalancing rules (not ad hoc emotional decisions)
- Build a checklist for entry and exit decisions

### Step 5: Execute
Do what the plan says:
- Don't second-guess the system in the moment
- Record all decisions and rationale
- Review results periodically and iterate

> "Most people fail at Step 2 — they don't identify problems because problems are painful, and they'd rather not face them." — Principles

---

## Algorithmic Decision-Making

### Why Systems Beat Intuition

> "I've found that in most cases, good principles are more reliable than the people who are making the decisions." — Principles

Dalio's argument for systematic investing:
1. **Humans are biased** (see Munger's 25 tendencies — Dalio agrees with all of them)
2. **Emotions dominate in extremis** — exactly when good decisions matter most
3. **Algorithms are consistent** — they apply the same rules regardless of emotional state
4. **Algorithms can be backtested** — you can see how they would have performed in past environments
5. **Algorithms can be improved** — by systematically studying where they fail

### The Bridgewater Approach

Bridgewater's investment process:
1. **Develop a thesis** about cause-and-effect relationships in the economy/markets
2. **Express the thesis as rules** — "If X happens, do Y"
3. **Backtest the rules** across many historical environments
4. **Run the rules in real-time** — with systematic monitoring
5. **Override only with explicit documentation** — any deviation from the system must be documented and reviewed

> "When I first started out, I made all decisions in my head. Now I make all decisions using my principles expressed as algorithms." — Principles

### Building Investment Decision Rules

For any investment framework, try to express decisions as rules:

**Example — Cycle Positioning:**
```
IF credit_spreads > 75th_percentile_of_20yr_range
AND default_rates_rising
AND central_bank_cutting_rates
THEN: Increase allocation to distressed credit by 10%
      Reduce allocation to equities by 10%
      Set recheck_date = 90 days
```

**Example — Inflation Environment:**
```
IF CPI_YoY > 4% AND rising
AND central_bank_rate < CPI (negative real rates)
THEN: Increase TIPS allocation to 15%
      Increase commodity allocation to 10%
      Reduce long-duration nominal bond allocation to 25%
```

**The value of writing rules:** Even if you don't follow them mechanically, the act of writing them forces clarity of thinking and reveals hidden assumptions.

---

## Learning from Mistakes: The Pain Button

### Dalio's Personal Evolution

Dalio's most formative experience was his catastrophic wrong call in 1982:
- He publicly predicted a depression
- He was completely wrong
- He lost his money, his clients, and had to borrow from his father
- The experience nearly destroyed Bridgewater

**What he learned:**
> "The experience taught me two things: first, to be very humble about what I think I know. Second, to focus on building systems that can handle being wrong." — Principles

This led to:
- Radical transparency (so others can challenge your thinking)
- Believability weighting (so the best thinkers have the most influence)
- Systematic decision-making (so emotions don't override analysis)
- Stress testing (so you know how your portfolio performs when you're wrong)

### The "Pain Button" Framework

When something goes wrong:
1. **Feel the pain** — Don't suppress it. The pain of being wrong is useful information.
2. **Reflect on the pain** — What went wrong? What was the root cause? Was it a bad decision or bad luck?
3. **Create a principle** — Extract a rule from the experience. "In the future, when X happens, I will do Y."
4. **Encode the principle** — Add it to your decision system. Ideally, make it algorithmic.
5. **Test the principle** — Backtest it. See if it would have helped in past situations.

> "The most successful people are not those who avoid mistakes — they're those who learn from them the fastest." — Principles

---
