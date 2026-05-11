> 本文件是 `full-spectrum-modeling` 的参考资料。
> 核心框架见 `../integrated-modeling-core/SKILL.md`。

# 模式识别库 (Pattern Library)

以下 8 个模式捕捉了20年建模中反复出现的"信号"，帮助分析师快速识别风险或机遇。

---

## 模式 1: SaaS 效率悖论
**SIGNAL**: S&M % down, Growth rate also down
**USUALLY MEANS**: Demand saturation, not efficiency
**ACTION**: Compress long-term growth, stop using leverage assumptions after Year 3

**详细说明**: 当 Sales & Marketing 占收入比例下降时，许多分析师误读为"运营效率提升"。但如果同期增速也在下降，真正的原因往往是需求饱和 — 公司在减少获客投入因为可获取的客户池在萎缩。

---

## 模式 2: 应收账款爆炸
**SIGNAL**: AR Days up 3+ quarters, Revenue growth flat/down
**USUALLY MEANS**: Channel stuffing or collections problem
**ACTION**: Interview sales/finance, add "AR Days +10" scenario

**详细说明**: 应收账款天数持续上升而收入不增长，是经典的渠道压货信号。公司可能在季度末向渠道商推货以粉饰收入，但实际终端需求疲软。在建模中必须为AR恶化设置敏感性情景。

---

## 模式 3: 库存积压信号
**SIGNAL**: Inventory Days up, Revenue Growth down
**USUALLY MEANS**: Demand weakness or overproduction
**ACTION**: Check margin for writedown risk, adjust NWC model

**详细说明**: 库存天数上升叠加收入增速下降，意味着产品卖不动。需要检查是否有库存减值风险（特别是时尚/消费电子等有过时风险的行业），并相应调整NWC模型中的库存假设。

---

## 模式 4: CapEx 陡峭化
**SIGNAL**: CapEx spikes from 5% to 18% of Revenue
**USUALLY MEANS**: Major capex program
**ACTION**: Understand nature (one-time vs series), forward-book D&A

**详细说明**: CapEx突然飙升可能是新厂房、新数据中心或重大技术投资。关键是区分一次性投资和系列投资。前者可以在模型中设置为一次性事件后恢复正常，后者则需要重新审视长期CapEx假设和D&A路径。

---

## 模式 5: COGS 成分走高
**SIGNAL**: COGS % rising, competitors cutting prices
**USUALLY MEANS**: Competitive pressure
**ACTION**: Compress margin assumptions, add sensitivity

**详细说明**: 当行业整体在降价而公司的COGS占比却在上升，说明公司正在面临竞争压力且缺乏定价权。在这种情况下，假设margin会维持或改善是危险的。需要压缩margin假设并增加敏感性分析。

---

## 模式 6: 管理层 Guidance 精准下调
**SIGNAL**: Misses by same magnitude every quarter (e.g., always -1 ppt)
**USUALLY MEANS**: Management sandbagging
**ACTION**: Use "guidance + 1 ppt" as base case

**详细说明**: 如果管理层的预测每个季度都精确地低于实际结果同样的幅度（例如总是差1个百分点），这不是巧合而是策略性低估。聪明的分析师会根据历史偏差模式调整管理层指导，在Base Case中使用"guidance + 历史偏差中位数"。

---

## 模式 7: 融资成本上升但不提
**SIGNAL**: Debt spread rising, but management silent on interest cost pressure
**USUALLY MEANS**: Ignoring or downplaying
**ACTION**: Update WACC with new borrowing cost

**详细说明**: 当信用利差在扩大（通过CDS或债券二级市场观察），但管理层在earnings call中完全不提及融资成本压力时，可能是在回避不利信息。分析师应主动更新WACC中的debt cost假设，而不是等管理层承认。

---

## 模式 8: 高管离职信号
**SIGNAL**: 3+ key executives leave in 6 months, financials deteriorate
**USUALLY MEANS**: Strategic crisis
**ACTION**: Add downside scenario, delay investment decision

**详细说明**: 关键高管集体离职是最强烈的非财务预警信号之一。通常意味着内部存在严重的战略分歧、文化问题或未公开的财务困难。在建模中应立即增加downside scenario的权重，并建议推迟投资决策直到新管理层稳定。

---

## 快速参考表

| # | 模式 | 信号 | 通常含义 | 建模行动 |
|---|------|------|---------|----------|
| 1 | SaaS 效率悖论 | S&M % ↓, Growth ↓ | 需求饱和 | 压缩长期增速 |
| 2 | AR 爆炸 | AR Days ↑ 3+Q, Revenue ↔↓ | 渠道压货 | 加 AR Days +10 情景 |
| 3 | 库存积压 | Inventory Days ↑, Revenue ↓ | 需求疲弱 | 检查减值风险, 调NWC |
| 4 | CapEx 陡峭化 | CapEx 从5%→18% | 重大投资 | 区分一次性/系列, 前推D&A |
| 5 | COGS 走高 | COGS % ↑, 竞争降价 | 竞争压力 | 压缩margin, 加敏感性 |
| 6 | Guidance 精准下调 | 每季度差同幅度 | 管理层沙袋 | 用 guidance +1pp |
| 7 | 融资成本静默上升 | Spread ↑, 管理层不提 | 回避坏消息 | 更新WACC中debt cost |
| 8 | 高管离职潮 | 3+ 高管半年离职 | 战略危机 | 加downside, 推迟决策 |
