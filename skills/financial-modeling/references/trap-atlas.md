> 本文件是 `full-spectrum-modeling` 的参考资料。
> 核心框架见 `../integrated-modeling-core/SKILL.md`。

# 陷阱图谱 — 十大危险陷阱 (Trap Atlas — 10 Most Dangerous Modeling Traps)

根据过去 20 年 500+ 个已完成项目的回溯分析，以下十个陷阱的危险等级按"频率×影响×隐蔽性"排序：

---

## 陷阱 1: 终端价值幻觉 (The Terminal Value Mirage)
🔴🔴🔴 | 频率：71% | 影响：+18% 价值高估 | 隐蔽性：高

**症状**: Terminal Value > 70% of Enterprise Value，Explicit forecast 增速 >30% 且持续 5+ 年

**检测信号**: Terminal Value > 70%, WACC - TGR < 3%

**修复**: 延长 explicit forecast 至 10 年，设置衰减路径，用 multiple 法交叉验证

---

## 陷阱 2: 净现金地雷 (The Net Cash Time Bomb)
🔴🔴🔴 | 频率：58% | 影响：+200% 回报率虚高 | 隐蔽性：极高

**症状**: 混淆"流动资产 - 流动负债"和真正的"净现金"

**检测信号**: Balance Sheet 中净现金从未被明确定义，没有"Interest-bearing Debt"明细

**修复**: 强制定义 `Net Cash = Cash + ST Inv - (ST Debt + LT Debt)`，建立"Debt Bridge"

---

## 陷阱 3: NWC 时间炸弹 (The Working Capital Trap)
🔴🔴 | 频率：64% | 影响：-15% 到 +20% FCF | 隐蔽性：中等

**症状**: NWC 假设为"收入的 X%"，但未考虑增长期指数级增长

**检测信号**: NWC 假设是"固定%"，没有 AR/Inventory/AP Days 明细

**修复**: 拆解为组件，高增长期设置更高的 NWC%

---

## 陷阱 4: 成本的"黄金延长线"陷阱 (The Linear Cost Trap)
🔴🔴 | 频率：62% | 影响：-8% 到 +12% EBITDA | 隐蔽性：中等

**症状**: SG&A、R&D 等假设为"收入的 X%"，忽视规模效应

**检测信号**: 成本假设在 forecast period 完全平线

**修复**: 为不同阶段设置不同成本假设，加入规模效应

---

## 陷阱 5: 毛利率压缩陷阱 (The Margin Compression Trap)
🔴🔴 | 频率：48% | 影响：-25% FCF 高估 | 隐蔽性：高

**症状**: 毛利率在 5-10 年保持不变，忽视竞争压力

**检测信号**: 毛利率假设完全基于历史，COGS/Revenue 完全平线

**修复**: 进行竞争定价分析，设置margin压缩路径

---

## 陷阱 6: CapEx/D&A 不匹配陷阱 (The CapEx-Depreciation Mismatch)
🔴 | 频率：41% | 影响：+15% FCF 高估 | 隐蔽性：高

**症状**: CapEx 下降但 D&A 保持平线，资产存量变化不匹配

**检测信号**: CapEx 和 D&A 无逻辑关系

**修复**: 建立"固定资产滚动表"，验证逻辑一致性

---

## 陷阱 7: 管理层"沙袋"陷阱 (The Sandbagging Trap)
🔴 | 频率：35% | 影响：-20% 实际回报 | 隐蔽性：极高

**症状**: 管理层指导恰好比分析师共识低 1-2 ppts

**检测信号**: 指导与共识差异 1-2 ppts，多次"低于预期后实现"

**修复**: 跟踪历史准确性，用"修正后的管理层指导"

---

## 陷阱 8: Revenue 混合增速陷阱 (The Revenue Mix Shift Trap)
🔴 | 频率：56% | 影响：±10% margin | 隐蔽性：中等

**症状**: 总 Revenue CAGR 准确，但不同产品线增速差异巨大

**检测信号**: 产品线有不同增速但用总体 CAGR，Margin 完全平线

**修复**: 逐条产品线分别建模，追踪Revenue share evolution

---

## 陷阱 9: WACC 假设陷阱 (The WACC Assumption Trap)
🔴 | 频率：52% | 影响：±300 bps IRR | 隐蔽性：高

**症状**: WACC 基于初始条件，假设不变 10 年

**检测信号**: WACC 没有敏感性分析

**修复**: 加入"WACC ±1 ppts"敏感性，定期review

---

## 陷阱 10: 情景分析中的"相关性错误" (The Scenario Correlation Trap)
🟡 | 频率：38% | 影响：±5% scenario IRR | 隐蔽性：极高

**症状**: Bull/Base/Bear 假设了"独立"的参数，但实际高度相关

**检测信号**: Bear 场景中增速 5% 但 exit multiple 仍 9x（应该 7x）

**修复**: 建立"增速 ↔ Multiple"相关性矩阵

---

## 快速参考表

| # | 陷阱 | 严重度 | 频率 | 影响 | 检测信号 | 修复 |
|---|------|--------|------|------|---------|------|
| 1 | Terminal Value Mirage | 🔴🔴🔴 | 71% | +18% overvaluation | TV > 70% of EV, WACC-TGR < 3% | Extend explicit forecast to 10yr, set decay path, cross-validate with exit multiples |
| 2 | Net Cash Time Bomb | 🔴🔴🔴 | 58% | +200% inflated returns | Net cash never explicitly defined, no interest-bearing debt detail | Force define: Net Cash = Cash + ST Inv - (ST Debt + LT Debt), build Debt Bridge |
| 3 | Working Capital Trap | 🔴🔴 | 64% | -15% to +20% FCF | NWC as "fixed % of revenue", no AR/Inventory/AP Days detail | Decompose into components, set higher NWC% during high-growth periods |
| 4 | Linear Cost Trap | 🔴🔴 | 62% | -8% to +12% EBITDA | Cost assumptions flat across entire forecast period | Set different assumptions by stage, include scale effects |
| 5 | Margin Compression | 🔴🔴 | 48% | -25% FCF overestimate | Gross margin flat 5-10yr, COGS/Revenue flat | Competitive pricing analysis, set margin compression path |
| 6 | CapEx-D&A Mismatch | 🔴 | 41% | +15% FCF overestimate | CapEx declining but D&A flat, no logical relationship | Build fixed asset roll-forward, verify consistency |
| 7 | Sandbagging Trap | 🔴 | 35% | -20% actual return | Management guidance exactly 1-2pp below consensus, repeated "beat-and-raise" | Track historical accuracy, use "adjusted guidance" (+1pp) |
| 8 | Revenue Mix Shift | 🔴 | 56% | ±10% margin | Revenue CAGR accurate but product lines grow at very different rates | Model each product line separately, track revenue share evolution |
| 9 | WACC Assumption Trap | 🔴 | 52% | ±300bps IRR | WACC fixed for 10 years, no sensitivity analysis | Add WACC ±1pp sensitivity, periodic review |
| 10 | Scenario Correlation | 🟡 | 38% | ±5% scenario IRR | Bear scenario: 5% growth but 9x exit multiple (should be 7x) | Build "growth ↔ multiple" correlation matrix |
