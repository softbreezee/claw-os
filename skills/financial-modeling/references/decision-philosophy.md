> 本文件是 `full-spectrum-modeling` 的参考资料。
> 核心框架见 `../integrated-modeling-core/SKILL.md`。

# 决策哲学层 — 模型之道 (Decision Philosophy — The Way of Modeling)

建模的最终目的不是造出一个"完美"的数字，而是在**不确定中做出更聪明的决策**。

## 精度 vs 准确度 (Precision vs Accuracy)

一个模型可以精确到小数点后两位，但完全错误的方向。高精确度给人虚假的自信。

**正确做法**:
- 以"范围"而非"单点"呈现：IRR 在 15-22% 之间
- 定期（每季度）更新，不要 just extrapolate
- 强调"主要假设"而非"精确数字"

## 管理层乐观主义 vs 市场悲观主义 (Management Optimism vs Market Pessimism)

**校准方式**:
```
Base Case = Max(Management - 2 ppts, Market Consensus - 1 ppt)

理由：
  - 管理层 -2 ppts：执行风险折扣
  - 市场 -1 ppts：上行惊喜空间
```

## 证据的等级制度 (Hierarchy of Evidence)

**第一等级**: 审计过的财务数据（95%+ 可信）
**第二等级**: 管理层指导（60-80% 可信）
**第三等级**: 分析师共识（50-70% 可信）
**第四等级**: 自建模型（40-60% 可信）

**正确用法**: 先用第一级建立基线，逐级补充，发现不一致时调查原因。

## 何时信任模型，何时超越模型 (When to Trust, When to Override)

**应该信任**:
- 通过所有 Quality Gates
- 关键假设与历史和行业数据对齐
- Bear case 下投资仍有吸引力

**应该超越**:
1. 战略转折点（管理层宣布重大转向）
2. 市场黑天鹅（疫情、AI浪潮等）
3. 管理层可信度崩塌
4. 模型结果与直觉严重背离

## "报纸测试" — The Newspaper Test

**问自己**: "如果这份分析明天上《金融时报》，会不会很傻？"

这强制你从第三方眼光审视自己的工作。

## 精通不确定性 (Mastering Uncertainty)

最高层次的建模能力不是消除不确定性，而是在不确定中做出最好决策。

**资深分析师的标志**:
- 知道自己不知道什么
- 能在多个场景间权衡
- 能根据新信息动态调整
- 能用"概率加权"思维
