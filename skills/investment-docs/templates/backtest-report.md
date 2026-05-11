> 本文件是 `full-spectrum-modeling` 的文档模板。
> Layer 5 模块见 `../integrated-modeling-docs/SKILL.md`。

# MODEL BACKTEST REPORT

**Company**: [公司名] | **Model Date**: [建模日期] | **Review Date**: [复盘日期]
**Investment Date**: [投资日期] | **Hold Period**: [X] months | **Track**: [HF/BT/Full]

---

## 1. OVERVIEW

| Item | Detail |
|------|--------|
| Original Model IRR (Base) | [X]% |
| Updated IRR (Based on Actuals) | [X]% |
| IRR Deviation | [±X] ppts |
| Original MoIC (Base) | [X.Xx] |
| Current Status | [Holding / Exited / Partially Exited] |
| Realized Return (if exited) | [X]% |

---

## 2. FORECAST vs ACTUAL

| Metric | Year 1 Forecast | Year 1 Actual | Variance | Variance % | Status |
|--------|----------------|--------------|----------|-----------|--------|
| Revenue ($M) | [X] | [X] | [±X] | [±X]% | [On-track / Over / Under] |
| EBITDA ($M) | [X] | [X] | [±X] | [±X]% | [On-track / Over / Under] |
| EBITDA Margin | [X]% | [X]% | [±X]pp | — | [On-track / Over / Under] |
| CapEx ($M) | [X] | [X] | [±X] | [±X]% | [On-track / Over / Under] |
| FCF ($M) | [X] | [X] | [±X] | [±X]% | [On-track / Over / Under] |
| NWC ($M) | [X] | [X] | [±X] | [±X]% | [On-track / Over / Under] |

**Status 定义**:
- On-track: 偏差 <5%
- Over/Under: 偏差 5-15%
- **Significantly Over/Under**: 偏差 >15%

---

## 3. ASSUMPTION DEVIATION RANKING

| Rank | Assumption | Forecast | Actual-Implied | Deviation | Impact on IRR | Cumulative IRR Impact |
|------|-----------|---------|---------------|-----------|--------------|----------------------|
| 1 | [假设 A, e.g. Revenue CAGR] | [X]% | [X]% | [±X]pp | [±X]pp IRR | [±X]pp |
| 2 | [假设 B, e.g. COGS % Revenue] | [X]% | [X]% | [±X]pp | [±X]pp IRR | [±X]pp |
| 3 | [假设 C, e.g. CapEx % Revenue] | [X]% | [X]% | [±X]pp | [±X]pp IRR | [±X]pp |
| 4 | [假设 D] | [X]% | [X]% | [±X]pp | [±X]pp IRR | [±X]pp |
| 5 | [假设 E] | [X]% | [X]% | [±X]pp | [±X]pp IRR | [±X]pp |

**最大偏差来源**: [假设名] — [解释为什么预测偏离]

---

## 4. CALIBRATION RECOMMENDATIONS

基于本次回测，对未来同类模型的假设修正建议：

| # | 假设 | 原始默认 | 建议修正为 | 依据 |
|---|------|---------|----------|------|
| 1 | [假设 A] | [X]% | [Y]% | [实际数据表明...] |
| 2 | [假设 B] | [X]% | [Y]% | [行业趋势显示...] |
| 3 | [假设 C] | [X]% | [Y]% | [管理层执行力表明...] |

**流程改进建议**:
- [建议 1: e.g. "CapEx 假设应考虑管理层资本投入周期，而非简单用历史平均"]
- [建议 2: e.g. "NWC 在高增长期应设更高假设，不能用稳态 % Revenue"]

---

## 5. LESSONS LEARNED

### 预判正确的方面
- [哪些假设/判断被验证]

### 预判错误的方面
- [哪些假设/判断偏离最大，原因分析]

### 方法论改进
- [建模方法层面有什么可以改进]

### 信息获取改进
- [哪些信息如果提前获取，可以改善预测准确度]

---

*Prepared using Full-Spectrum Modeling v4.0 — Backtest Module*
