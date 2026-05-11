> 本文件是 `full-spectrum-modeling` 的参考资料。
> 核心框架见 `../integrated-modeling-core/SKILL.md`。

# 客户维度收入建模 (Customer-Dimension Revenue Modeling)

> **💀 曼巴注记 — 来源**: Project Stratus (MBK Partners, 韩国云基础设施) 不按"业务线"拆分收入，而是按**客户维度**建模。当一个公司的收入来自多个供应商/平台时，这是唯一正确的方法。

**适用场景**: 公司作为中间商/代理/分销商，收入来源与上游供应商绑定。

**架构**:
```
传统方法 (业务线维度):
  Revenue = Cloud Resale + Managed Services + Consulting
  ← 问题: 无法回答"如果AWS提价，对我们影响多大？"

客户维度方法 (Stratus模型):
  Revenue = Σ(by 供应商/平台) [
    Beginning Accounts × (1 + New Account Growth %)
    × ARPU
    × (1 + ARPU Growth %)
  ]

  具体展开:
  ├─ AWS Resale Revenue
  │   ├─ # of Accounts: 400 → 460 → 530
  │   ├─ ARPU: $142K → $155K → $170K
  │   ├─ Reserved Instance %: 35% → 40% → 45%
  │   └─ On-Demand / Spot Mix: 65% → 60% → 55%
  │
  ├─ Azure Resale Revenue
  │   ├─ # of Accounts: 150 → 180 → 215
  │   ├─ ARPU: $95K → $108K → $120K
  │   └─ Enterprise Agreement %: 60% → 65% → 70%
  │
  ├─ GCP Resale Revenue
  │   ├─ # of Accounts: 50 → 65 → 85
  │   ├─ ARPU: $60K → $72K → $85K
  │   └─ Growth driven by AI/ML workloads
  │
  └─ Other Platforms (China Cloud, etc.)
      └─ Catch-all, minimal assumptions
```

**为什么客户维度更好**:
1. **供应商风险可见**: 如果AWS占60%收入，模型直接显示集中度风险
2. **ARPU差异化**: AWS ARPU ($142K) ≠ Azure ARPU ($95K) — 混合计算会掩盖定价结构
3. **增速差异化**: GCP可能增长40%/年而AWS只增长15% — 总体CAGR无法捕捉这种结构变化
4. **组合效应可模拟**: 可以模拟"如果Azure份额从25%升至35%，整体margin如何变化？"

**触发条件**: 当以下条件满足时，强制使用客户维度建模:
```
IF company is reseller/distributor/agent
  AND top supplier > 30% of revenue
  THEN require customer-dimension revenue build-up
  WITH separate Account/ARPU/Growth for each supplier/platform
```

**验证追加项**:
```
☐ 每个供应商/平台都有独立的Account数、ARPU、增速假设
☐ 各平台ARPU差异有合理解释(企业级vs SMB、产品组合)
☐ 供应商集中度已计算(Top 1, Top 3占比)
☐ 集中度>50%时，已添加"供应商流失"敏感性情景
```
