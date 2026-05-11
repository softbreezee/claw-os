---
name: integrated-modeling-bt
description: 综合建模Buyout Track模块。PE收购/LBO专属分析框架：管理层可投性评估、PE特有风险、债务容量分析、百日计划、退出蓝图、复杂资本结构建模、Recovery Analysis。当Step 0选择Buyout Track或Full Track时加载。
---

# 综合财务建模系统 - Buyout Track模块

## 模块概述

本模块针对**PE收购 (Private Equity Buyout)** 和 **LBO (Leveraged Buyout)** 交易设计，提供企业级分析框架。遵循**Paul Singer / Elliott Management** 信用分析标准，涵盖管理层评估、债务容量、复杂资本结构、恢复分析和结构化风险评估。

---

## 一、管理层可投性评估 (Management Investability Assessment)

管理层质量是LBO成败的核心驱动因素。本模块采用多维度量化框架替代主观判断。

### 1.1 CEO可投性评估表

| 评估维度 | 权重 | 评分标准 (1-10) | 数据来源 |
|---------|------|----------------|--------|
| **PE经验** | 20% | 1=首次，5=参与过1次LBO，10=3+次LBO成功退出 | CV / References |
| **Board汇报能力** | 15% | 1=沟通差，5=可接受，10=定量驱动、主动透明 | 访谈记录 / 前任Board反馈 |
| **历史执行记录（可量化）** | 30% | 1=错过目标>30%，5=误差10-20%，10=连续达成或超额 | 历史财务数据 |
| **压力下表现** | 20% | 1=危机时失控，5=稳定，10=压力下提升执行 | 历史周期表现 / 危机案例 |
| **激励对齐度** | 15% | 1=利益冲突，5=中立，10=个人净值大幅锁定在equity | Equity结构 / 个人资产配置 |

**核心管理团队风险评估**

| 风险类别 | 评估内容 | 风险等级 |
|---------|--------|--------|
| Key Person离职风险 | 3年内替代概率、内部继任计划 | 高/中/低 |
| 留任激励状态 | Vesting schedule、干股、加速条款 | 是否充分 |
| 替代成本估计 | 重新招聘+培养周期（月）、一次性成本 | 成本占EBITDA % |

**综合得分映射到执行折扣**

```
综合评分 = Σ(各维度评分 × 权重)

如CEO评分 = (8×0.20 + 7×0.15 + 9×0.30 + 8×0.20 + 7×0.15) = 8.05
核心团队风险 = LOW → 合成得分 8.0
```

---

## 二、改进的执行折扣映射 (Execution Discount Mapping)

执行折扣反映管理层将财务计划转化为实际结果的能力。本映射基于PE基金15年实数据校准。

### 2.1 评分与折扣对应表

| 综合评分 | 执行折扣范围 | 行业调整 | 触发条件 |
|---------|----------|--------|--------|
| **8.0 - 10.0** | -2% 到 -5% | 管理密集型上边界；资产密集型下边界 | 优秀，推进 |
| **6.0 - 8.0** | -5% 到 -10% | 保守估计，中点-7% | 可接受，需监控 |
| **< 6.0** | -8% 到 -15% | **触发Layer 2警告** | 需董事会特别讨论 |

### 2.2 行业调整规则

**管理密集型行业** (使用上边界)：
- 软件/SaaS、消费品牌、营销服务、金融科技
- 理由：执行质量决定市场份额、客户保留，管理质量风险高

**资产密集型行业** (使用下边界)：
- 基础设施、房地产、采矿、重工业
- 理由：资产成本固定，运营杠杆稳定，管理风险相对低

**混合型行业** (取中点)：
- 汽车零部件、工业制造、医疗设备、专业服务

### 2.3 关键限制条款

⚠️ **此映射为质量-定量桥接假设。不可过度依赖精确数字。**
- 执行折扣应视为**范围**，不是点估计
- 每年应基于实际业绩重新校准
- 若出现显著偏离，触发管理层变更评审

---

## 三、PE特有风险评估 (PE-Specific Risk Assessment)

### 3.1 风险矩阵

| 风险因子 | 表现形式 | 量化影响 | 缓解手段 |
|---------|--------|--------|--------|
| **Change of Control条款** | 债务/供应商/客户合同触发 | 可能>20%成本增加或收入下降 | 前期同意书、替代方案 |
| **关键人依赖** | CEO/技术创始人不可替代 | 8-15%执行折扣 | 留任协议、期权加速、董事会强化 |
| **信息不对称** | 卖方隐藏问题、尽调不足 | 未知风险乘以 | 详细尽调、锁定陈述和保证保险 |
| **Add-on整合风险** | 跨境、跨产业、文化冲突 | Year 1 EBITDA -10%至-25% | 整合playbook、专项团队、分阶段支付 |
| **杠杆放大风险** | 利率上升、增长放缓时债务失控 | 见3.2 |见3.2 |
| **监管/周期风险** | 行业衰退、合规成本、汇率 | 情景分析（见第六部分） | 多路径分析、成本结构灵活性 |

### 3.2 杠杆放大敏感性分析 (强制性计算)

**公式框架：**

```
入场杠杆 = Entry Debt / LTM EBITDA = X.Xx
股权腐蚀临界点 = Base EBITDA下降 [Y]% 时，Equity → 0
安全边际 = Y% EBITDA 衰退空间
```

**案例计算：**

假设交易参数：
- 购买价格：$500M
- 入场EBITDA (LTM)：$100M
- Senior Debt：$300M
- Subordinated Debt：$100M
- Equity：$100M
- 入场杠杆：4.0x

债务成本假设：
- Senior Debt利率：5.5%，年利息 = $16.5M
- Sub Debt利率：8.5%，年利息 = $8.5M
- 总利息：$25M

**场景分析：**

| EBITDA下降 | 新EBITDA | 新杠杆 | 利息覆盖率 (ICR) | 安全评估 |
|----------|---------|-------|----------------|--------|
| 0% | $100M | 4.0x | 4.0x | 健康 |
| -10% | $90M | 4.4x | 3.6x | 可接受 |
| -20% | $80M | 5.0x | 3.2x | 警告 |
| -30% | $70M | 5.7x | 2.8x | 违约风险 |
| -40% | $60M | 6.7x | 2.4x | **Equity腐蚀** |

**结论：** EBITDA下降超过30%时，ICR低于2.0x，触发违约条款。安全边际 = 30%。

---

## 四、层级债务结构建模 (Complex Capital Structure)

符合Paul Singer / Elliott Management标准，支持任意债务层级。

### 4.1 债务工具清单

每一个债务工具应包含以下属性：

| 工具类型 | 利率类型 | 优先权 | Amortization | 抵押品 | 通常条款 |
|---------|--------|--------|------------|--------|---------|
| **Revolving Credit Facility (RCF)** | Floating (SOFR+200-300bps) | 1 (最优先) | 非摊销 | 第一优先权 | 3-5年期限，财务契约（最小liquidity） |
| **Term Loan A** | SOFR+175-250bps | 2 | 5-7年递减 | 第一优先权 | 有形资产担保，通常无prepay惩罚 |
| **Term Loan B** | SOFR+250-400bps | 3 | 1-3% P&I | 第二优先权 | 无担保或次优，可能有prepay优惠 |
| **Senior Secured Notes** | Fixed 4.5%-6.5% | 2-3 | Bullet或渐进式 | 第一/二优先权 | 7-10年期，call protection（2-3年） |
| **Senior Unsecured Notes** | Fixed 6.0%-8.0% | 4-5 | Bullet | 无担保 | 10年期，通常incurrence covenant |
| **Mezzanine / Sub Notes** | 8.5%-12% PIK选项 | 6 | Bullet | 无担保 | 权益性特征，可转换 |
| **PIK Toggle** | 6% Cash / 9% PIK | 6-7 | Bullet | 无担保 | 发行人可选现金或accrue利息 |
| **Convertible Notes** | 低利率（3-4%） | 若转换前则与Senior同级 | Bullet | 条件性 | 转换价格、soft/hard call条款 |
| **Preferred Stock** | 股息率7-12% | 7 (Equity前) | Non-cumulative 或 Cumulative | 无 | 参与/非参与，转换权 |
| **Management Options / Warrants** | 无成本 | 最次级 | 无 | 无 | 行权价、vesting、稀释计算 |

### 4.2 每层级必需文件元素

对每一债务层级，必须记录：

```
[Debt Layer Name]
├─ Rate Type: Fixed / Floating (指标 + Spread) / PIK / 混合
├─ Outstanding Amount: $XXX M
├─ Maturity Date: YYYY-MM-DD
├─ Annual Interest: $XX M (或% of Principal)
├─ Amortization Schedule:
│  ├─ Year 1-2: 0% 摊销
│  ├─ Year 3-5: 递增至 5% P&A
│  └─ Year 6+: Bullet 到期
├─ Collateral: 1st Lien on [资产清单]
├─ Call Protection:
│  ├─ Period: Year 0-2 不可call
│  ├─ Make-whole provision: 50-100bps
│  └─ Year 3+ par call
├─ Covenant Package:
│  ├─ Type: Maintenance vs Incurrence
│  ├─ Financial Covenants: Max Debt/EBITDA = 4.5x, Min ICR = 2.0x
│  ├─ Operational Covenants: 资产出售限制、dividend限制、并购限制
│  └─ Reporting: 每月/季/年要求
├─ Repayment Waterfall Priority: [Rank 1-8]
└─ Cross-default Triggers: 任何层级>$XX M违约触发
```

### 4.3 Sources & Uses表（详细行项目）

```
SOURCES OF FUNDS                      $M        %
─────────────────────────────────────────────────
Senior Secured RCF (drawn)           50        10%
Term Loan A (7-year)                200        40%
Term Loan B (institutional)         150        30%
Senior Secured Notes (8-year)        80        16%
Mezzanine Financing                  20         4%
─────────────────────────────────────────────────
Total Debt Financing                500        100%

Equity (Sponsor)                    150
Equity (Management, options pool)    50
─────────────────────────────────────────────────
TOTAL SOURCES                       700

USES OF FUNDS                         $M        %
─────────────────────────────────────────────────
Purchase Price (Enterprise Value)   600        86%
│  Base Business (going concern)   500
│  Add-on acquisition (synergies)  100
Refinance Existing Debt              20         3%
Transaction Fees (banking, legal)    45         6%
Financing Fees & Arrangement Fees    20         3%
Working Capital / Adjustment          15         2%
─────────────────────────────────────────────────
TOTAL USES                          700
```

---

## 五、债务容量实质分析 (Debt Capacity Analysis)

**必须在用户回答S4前向其展示本分析。**

### 5.1 自由现金流转换分析

```
FCFF Conversion Model
─────────────────────────────────────

EBITDA (LTM / Year 1 Proj)           $100M
Less: Depreciation & Amortization    ($12M)      ← Historic or normalized
Less: Maintenance CapEx               ($8M)      ← Sustain operations
─────────────────────────────────────
Operating Profit (EBIT)               $80M

Less: Taxes (28% normalized)          ($22M)
─────────────────────────────────────
NOPAT                                 $58M

Add: D&A (non-cash)                   $12M
Less: Change in NWC                   ($2M)       ← Conservative
─────────────────────────────────────
Free Cash Flow to Firm               $68M

FCF / EBITDA Conversion Rate         68%

Industry Benchmark (Mfg)              65% - 75%
Assessment: ✓ 在范围内，可信
```

### 5.2 Interest Coverage Ratio (ICR) 测试表

```
Leverage & Debt Service Coverage
────────────────────────────────────────────────
Entry Leverage (Debt / EBITDA)       4.0x

Year 1 Debt Service Schedule:
  Senior RCF Interest        $2.8M
  Term Loan A Interest       $11.0M
  Term Loan B Interest       $6.4M
  Senior Notes Interest      $5.2M
  Mezzanine Interest         $1.6M
  ────────────────────────
  Total Annual Interest      $27.0M

Year 1 FCF (with execution discount)  $62M (9% discount applied)

ICR = FCF / Total Interest           2.3x

Covenant Floor (typical)              2.0x

Headroom                             +15%

Risk Assessment:
  ✓ 通过初始测试
  ⚠️ 有限余地；对增长/成本假设敏感
```

### 5.3 五年偿债容量表

```
5-Year Debt Repayment Capacity Analysis
────────────────────────────────────────────────────────
        Year 1    Year 2    Year 3    Year 4    Year 5
────────────────────────────────────────────────────────
EBITDA  $100M    $107M    $115M    $123M    $132M
  Growth %    7%      7%       8%       7%
  (conservative vs. industry 9%)

FCF (pre-debt)  $62M   $68M   $73M   $80M   $87M
  FCF/EBITDA   62%    64%    63%    65%    66%

Total Debt (BOP)  $400M   $380M   $355M   $325M   $290M
  Reduction/yr    $20M    $25M    $30M    $35M

Interest Expense  $27M    $25M    $23M    $20M    $18M
  Rate (average)  6.75%   6.6%    6.5%    6.2%    6.2%

ICR Ratio         2.30x   2.72x   3.17x   4.00x   4.83x

Debt / EBITDA     4.00x   3.55x   3.09x   2.65x   2.20x

Covenant Test:
  Max Debt/EBITDA Covenant (4.5x):    ✓✓✓✓✓ Clear
  Min ICR Covenant (2.0x):             ✓✓✓✓✓ Clear
  Headroom Year 1:                    +15% (vulnerable window)
  Headroom Year 5:                    +140% (strong)
```

### 5.4 推荐债务天花板

| 指标 | 建议上限 | 理由 |
|------|--------|------|
| **最大杠杆倍数** | 4.5x Debt / EBITDA | ICR>2.0x + 15% headroom @ Year 1 |
| **最大利息支出** | < 27% FCF | 保留投资/应急资本 |
| **Maturity Wall** | 无单年债务>总债务35% | 避免再融资集中风险 |
| **Floating Rate Exposure** | < 60% 总债务 | 利率敏感性限制 |
| **最小Liquidity** | $15M + 12个月费用 | 运营应急 |

---

## 六、价值创造杠杆 (Value Creation Levers)

### 6.1 短期杠杆 (0-12个月)

| 杠杆 | 目标增幅 | 量化方法 | 风险 |
|------|--------|--------|------|
| **收入保留** | +2-4% | 减少客户流失，提升定价 | 客户反弹、竞争压力 |
| **成本优化** | +3-6% EBITDA margin | SG&A削减、采购节约、制造效率 | 管理层执行风险、员工流动 |
| **营运资本释放** | $5-15M 一次性现金 | 存货周期优化、应收账款加速 | 运营风险、供应链中断 |
| **工作资本管理** | 改善Cash Conversion Cycle | Days Sales Outstanding -5d, DIO -10d | 供应商关系恶化 |

### 6.2 中期杠杆 (1-3年)

| 杠杆 | 目标增幅 | 关键驱动 | 投资需求 |
|------|--------|--------|---------|
| **有机增长** | +5-8% 年度收入 | 新市场进入、产品创新、销售力量 | 营销和R&D: $2-5M/年 |
| **Add-on并购** | +10-15% 年度收入 | 标的收购整合、交叉销售 | 收购成本，见Sources&Uses |
| **运营杠杆** | +200-300bps margin | 规模化、自动化、供应链整合 | CapEx $5-10M/年 |
| **定价权** | +2-3% ASP | 产品升级、品牌强化、细分市场 | 相对低CapEx |

### 6.3 长期杠杆 (3-7年，退出前)

| 杠杆 | 目标增幅 | 机制 | 退出倍数影响 |
|------|--------|------|------------|
| **行业整合** | 市场份额 +500-800bps | 进一步add-on，巩固头部位置 | EBITDA倍数 +0.5-1.0x |
| **国际扩张** | 收入多元化到海外 | 出口、FDI、当地合作伙伴 | 风险溢价降低 |
| **数字化转型** | 重复性收入 +15-20% | 平台/SaaS转型、订阅模式 | EBITDA倍数 +1.0-2.0x |
| **企业品牌** | B2B → B2C品牌力 | 市场推广、高管可见性 | 溢价倍数 +0.3-0.5x |

---

## 七、百日计划框架 (100-Day Plan)

LBO成功的90%取决于前100天的执行。此框架细分关键里程碑。

### 7.1 第1-30天：稳定化 (Stabilization)

**目标：** 保持业务连续性，避免信息真空导致人才流失和客户离心。

| 活动 | 责任方 | 截止日期 | 输出 | 财务影响 |
|------|--------|---------|------|---------|
| **1. 董事会第一次会议** | CEO + Sponsor | Day 5 | 董事会章程、年度目标、governance模板 | 治理框架 |
| **2. 关键管理层留任协议签署** | HR + Sponsor | Day 10 | 3-5年留任协议，equity acceleration条款 | 保留风险↓ |
| **3. 客户沟通计划** | 销售副总 + CEO | Day 7 | 客户拜访脚本、续约确认 | 收入保护 |
| **4. 供应商&债权人沟通** | CFO | Day 10 | 付款确认、条款无变化声明 | 运营连续性 |
| **5. IT系统与数据访问** | CTO + Sponsor | Day 3 | 系统审计、数据备份、访问权限清单 | 信息可获得性 |
| **6. 银行/债权人通知** | CFO + 律师 | Day 1 | 融资步骤证书、covenant baseline baseline | 债务人信任 |
| **7. 内部通信计划** | HR + Communications | Day 2 | 员工全员说明会、FAQ文档、Hotline | 留用率 |
| **8. 前期融资步骤** | CFO + 财务部 | Day 20 | 初期payroll、operating expenses的提前支付安排 | 流动性保障 |

**第30天结束时的里程碑：**
- ✓ 关键管理层(3-5名)签署留任协议，无意外离职
- ✓ 客户续约率 ≥ 95% (或至少确认意向)
- ✓ 供应商付款延期未发生
- ✓ 银行covenant baseline 确认无争议
- ✓ 员工流动率 < 2% (正常范围)

### 7.2 第31-60天：诊断 (Diagnosis)

**目标：** 快速识别EBITDA改进机会，为Year 1计划奠定数据基础。

| 诊断工作流 | 分析方法 | 输出 | 财务影响 |
|----------|--------|------|--------|
| **A. EBITDA改进机会扫描** | • 历史P&L分析，对标行业基准 • 成本结构明细表(by department) • Contribution margin by product/customer | EBITDA改进机会清单，优先级排序，预期收益范围 | +$2-5M Year 1 (3-5%) |
| **B. 成本结构优化** | • SG&A成本瀑布图 • 支出按category分组 • 对标同行比率 | 可削减成本清单：招聘冻结、外包评审、采购谈判空间 | -$1-2M SG&A / 年 |
| **C. 定价权评估** | • 产品/客户level margin分析 • 竞争对标 • 价格弹性测试（样本） | 提价空间评估：哪些产品/客户可接受2-3% ASP提升 | +$0.5-1.5M revenue / 年 |
| **D. 资本效率审查** | • CapEx历史支出 • PP&E周期更新计划 • 维护 vs 增长CapEx分离 | 优化CapEx计划，削减不必要的资本开支 | 释放$2-3M现金 / 年 |
| **E. 营运资本优化** | • Days Inventory Outstanding、DPO、DSO计算 • 对标行业标准 • 库存老化分析 | NWC释放计划：库存削减、应收加速、应付延期 | 一次性释放$3-8M 现金 |
| **F. 收入漏斗分析** | • Pipeline review by stage • Win rate & avg deal size 历史 • 产品销售mix | 收入增长机会识别：product upsell, market expansion, customer expansion | +$1-3M revenue 年度 |

**第60天结束时的里程碑：**
- ✓ EBITDA改进清单确定，总潜力 $4-8M (4-8% margin expansion)
- ✓ 成本削减初步实施：招聘冻结、采购谈判开启
- ✓ 定价评估完成，合格的ASP提升targets确定
- ✓ 营运资本释放计划确认，预计Day 100前释放 $3-5M
- ✓ Year 1 财务目标初稿完成（基于诊断数据）

### 7.3 第61-100天：启动 (Launch)

**目标：** 启动具体计划，建立18个月财务与运营里程碑追踪。

#### 7.3.1 具体计划与里程碑

```
LAUNCH INITIATIVES TRACKER

Initiative 1: SG&A成本削减 (Target: -$1.2M / 年)
  ├─ Day 65: 招聘冻结确认，HR重新规划
  ├─ Day 75: 外包评审提交（哪些函数可外包）
  ├─ Day 90: 前3个月削减 -$0.3M (quarterly run-rate = -$1.2M)
  └─ 18个月里程碑: 累计 -$1.8M

Initiative 2: 定价改革 (Target: +$1.0M / 年收入)
  ├─ Day 70: 产品群定价模型完成
  ├─ Day 85: 高层客户谈判启动（top 20 customers）
  ├─ Day 95: 标准化定价表生效（中长尾客户）
  └─ 18个月里程碑: Q1 +$150K, Q2 +$250K, Q3+ +$600K

Initiative 3: 营运资本释放 (Target: $5M 一次性 cash)
  ├─ Day 68: 库存优化计划 (target: -15 days of inventory)
  ├─ Day 82: 应收账款激励启动 (早付折扣2%)
  ├─ Day 100: 释放现金 $4-5M (通过库存削减 + AR加速)
  └─ 18个月里程碑: 维持改进状态，prevent reversion

Initiative 4: Add-on并购（若可行）(Target: +$8-10M 年收入)
  ├─ Day 75: 购买list 确定，招标启动
  ├─ Day 100: LoI 或初步协议 (若找到合格对象)
  └─ 18个月里程碑: 完成整合，达成协同 targets

Initiative 5: 运营改进项目 (Target: +$0.8M EBITDA / 年)
  ├─ Day 72: 精益或持续改进培训启动
  ├─ Day 90: 优先改进项目 (流程简化、自动化) 上线
  └─ 18个月里程碑: 累计改进 +$1.2M EBITDA
```

#### 7.3.2 Year 1 财务计划与映射规则

```
BASE CASE (Year 1 EBITDA)
─────────────────────────────────────
Entry EBITDA (LTM)              $100.0M
Less: Execution Discount (-8%)   ($8.0M)
─────────────────────────────────────
Base EBITDA (Year 1, normalized) $92.0M

VALUE CREATION INITIATIVES (发布schedule)
──────────────────────────────────────────
Timeline              Initiative              Impact      Cumulative
Day 61-90            SG&A + pricing kickoff  +$0.4M      $92.4M
Q1 (Jan-Mar)         Initiatives ramping     +$0.6M      $93.0M
Q2 (Apr-Jun)         CapEx benefit + pricing +$0.8M      $93.8M
Q3 (Jul-Sep)         NWC release + OpEx      +$0.6M      $94.4M
Q4 (Oct-Dec)         Full-run rate initiatives +$0.5M     $94.9M
──────────────────────────────────────────────────────────
Year 1 Projected EBITDA                      $94.9M

Variance to Year 1 Plan: -0.1M (-0.1%) ← Conservative

Key Assumption:
✓ Execution discount applies to BASE, not incremental
✓ Initiatives released per milestone calendar
✓ NOT full run-rate in Year 1 (ramp-up profile)
✓ 18-month targets (by mid-Year 2) incorporate higher realization rates
```

#### 7.3.3 100日计划资本配置

```
Available Year 1 Cash:
  FCF (from operations, with discount)    $62M
  Less: Debt amortization                 ($20M)
  Less: CapEx (normalized)                ($10M)
  ────────────────────────────────────
  Available for Initiatives & Contingency $32M

Allocation:
  • Initiative investments (OpEx)          $5M  (training, systems, temporary costs)
  • Add-on acquisition earnout / notes    $8M  (if applicable)
  • Debt paydown (accelerated)            $12M (covenant coverage)
  • Contingency / working capital         $7M
  ────────────────────────────────────
  Total                                   $32M ✓ Balanced
```

---

## 八、退出蓝图 (Exit Blueprint)

成功的LBO从Day 1就规划退出。本节框架化退出路径选择与时间表。

### 8.1 退出路径可行性矩阵

| 退出路径 | 企业条件 | 可行性评分(1-10) | 时间框架 | 潜在倍数 | 风险因子 |
|---------|--------|----------------|--------|--------|--------|
| **IPO** | 利润率>15%, 增长>10%, 市值$100M+, 公众认知 | 8 (若条件满足) | 4-5年 | 10-14x EBITDA | 市场周期、审计成本、SOX合规 |
| **战略出售** | 与行业玩家协同价值、市场地位顶部3 | 8-9 (最灵活) | 3-7年可行 | 8-12x EBITDA | 竞争威胁、监管批准 |
| **二级PE出售** | EBITDA $50-150M, 稳健增长, 并购创意 | 7-8 (流动性最强) | 3-5年 | 7-10x EBITDA | PE融资环境、对标公司涌入 |
| **Spin-off/Carve-out** | 目标自立运营、独特商业模式、分部独立性强 | 5-6 (复杂) | 5-7年 | 可能溢价10%+ | 运营复杂性、成本结构分离难 |
| **Dividend Recapitalization** | 成熟业务, 稳定FCF > $20M, 低增长 | 5 (临时) | 2-3年 | 现金回报，EBITDA倍数不变 | 债务增加风险、再融资成本 |

### 8.2 退出路径对持有期战略的影响

#### IPO路径 (Holding Period 4-5 年)

```
关键运营建设 (Year 1-3):
  • 合规性强化: SOX framework, 审计师聘用, internal controls
  • 财务透明度: 3年完整季报 + 年报历史, consistent accounting
  • 管理层品牌: CEO/CFO road show就绪、独立董事会
  • 业务多元化: 不过度依赖单一客户(>15%), 地理多元化
  • 增长故事: 持续>10% CAGR revenue, margin expansion

IPO前两年 (Year 3-4):
  • Underwriter聘用、路演准备
  • 目标定价: 12-14x forward EBITDA
  • 融资成本预提: IPO成本 3-5% 交易值

退出收益 (Year 5):
  • 假设IPO定价 13x EBITDA (Year 4 projected $110M) = $1,430M
  • 减去债务 (Year 4 末 $250M) = $1,180M equity value
  • MoIC (Multiple on Invested Capital):
    Initial Equity Investment   $150M
    Year 5 Equity Value        $1,180M
    ─────────────────────────────────
    MoIC = 7.9x

    Assumptions on dividends reinvested, no management dilution
```

#### 战略出售路径 (Holding Period 3-5 年)

```
关键战略建设:
  • 行业地位强化: top 3 市场份额, 品牌认知
  • 协同价值识别: 明确为战略买家的价值(成本协同、交叉销售)
  • 客户集中度管理: 防止大客户流失触发MAC
  • 国际扩张: 多地域运营降低风险

买方洽谈准备:
  • Data room完整性、legal 尽调就绪
  • Management presentation: "我们是你理想的Bolt-on"
  • 非竞争协议预规划(key person post-deal role)

退出收益估计 (Year 4-5):
  假设 EBITDA Year 4 = $110M
  战略买家支付多数: 11-13x EBITDA (溢价>PE 8-10x)
  Assume 12.0x = $1,320M

  Equity payoff (after debt reduction to $200M) = $1,120M
  MoIC = 7.5x (略低于IPO但风险更低、时间更短可能)
```

#### 二级PE出售路径 (Holding Period 3-5 年)

```
运营目标:
  • EBITDA 稳定且增长: $80-150M range，CAGR >7%
  • 利润率提升 200-300bps (via initiatives)
  • 杠杆数倍下降至 2.5-3.0x (debt reduction)

二级买方吸引力因子:
  • Management team经过验证
  • 并购平台已建立(add-on plays identified)
  • Margin expansion空间剩余(给新PE主人)

估值 & 收益:
  Year 4 EBITDA $110M × 8.0x (二级LBO标准倍数)
  = $880M Enterprise Value
  Equity value (假设债务 $220M) = $660M

  原Equity投入 $150M
  中期分红(若有) $50M
  MoIC = (660 + 50) / 150 = 4.7x

  Note: 二级相对IPO/战略风险高(取决于new sponsor)，但流动性强
```

### 8.3 退出时机IRR敏感性表

```
IRR Sensitivity: Exit Multiple vs Hold Period
───────────────────────────────────────────────────────────
Holding Period      8x EBITDA      10x EBITDA     12x EBITDA
───────────────────────────────────────────────────────────
Year 3 Exit         32% IRR         38% IRR        45% IRR
Year 4 Exit         25% IRR         31% IRR        37% IRR
Year 5 Exit         21% IRR         27% IRR        33% IRR
Year 7 Exit         18% IRR         24% IRR        30% IRR

假设基础:
  • Entry: 6.5x EBITDA，$100M base EBITDA
  • 年EBITDA增长 +7% (conservative)
  • Debt paydown: $15-25M/年
  • No interim distributions (equity retained)

IRR Targets by PE Fund Type:
  Mid-market (our focus): 20-30% target range
  Lower mid-market: 25-35% target range
  Large-cap: 15-25% target range

Assessment:
  ✓ Year 5 @ 10-12x multiple = 27-33% IRR → Attractive
  ⚠️ Year 3 @ <10x = Below target → Risk of dividend recap needed
```

---

## 九、Recovery Analysis (Paul Singer / Elliott标准)

若投资方进入压力/重组，各债权/股权层的回收价值。

### 9.1 清算分析 (Liquidation Analysis)

**假设：** 公司破产清算，资产按市场价格出售。

| 资产类别 | 账面价值 | 恢复率 | 清算价值 | 说明 |
|---------|--------|--------|---------|------|
| 应收账款 | $30M | 80% | $24M | 正常客户，少数坏账 |
| 库存 | $25M | 50% | $12.5M | 快速清库折扣，过时品 |
| PP&E (Machinery) | $40M | 30% | $12M | 二手机器折价 |
| Real Estate (若有) | $20M | 60% | $12M | 房产市场折价 |
| 无形资产 (Goodwill, IP) | $35M | 0% | $0M | 独立清算无价值 |
| Other (Cash, misc) | $5M | 100% | $5M | 现金即现金 |
| ─────────── | ────── | ────── | ─────── | |
| **总资产清算价值** | | | **$65.5M** | |

**清算成本抵扣：**

```
Liquidation Proceeds       $65.5M
Less: 清算费用 (律师, 破产受托人, 管理成本, 10-15% range)
  Assume: -$10.0M (15% 费用率，高成本)
──────────────────────────
Net Proceeds               $55.5M
```

### 9.2 持续经营重组分析 (Going-Concern Reorganization)

**假设：** 公司进入Chapter 11，以持续经营价值进行重组。

```
GOING-CONCERN VALUATION IN DISTRESS SCENARIO
─────────────────────────────────────────────────
Assume EBITDA在压力下降至 $70M (30% decline from $100M)
持续经营利率折扣幅度 (distress multiple): 6.0x vs 8.5x (normal)

Distressed Going-Concern Value:
  EBITDA (stressed)              $70M
  × Distressed Multiple          6.0x
  = Enterprise Value             $420M

Less: Senior Secured Debt @ fair value
  RCF + Term Loans + Sen Notes   $330M (假设no haircut on senior)
─────────────────────────────────────
Equity (or unsecured recovery)   $90M

Total Distressed EV (senior claims priority) = $330M senior + $90M junior
```

### 9.3 Recovery Waterfall (按绝对优先原则)

```
DISTRESSED ENTERPRISE VALUE = $420M (going-concern)
─────────────────────────────────────────────────────────

Layer 1: Secured Lenders (1st Lien on collateral)
  └─ RCF Outstanding                    $30M
  └─ Term Loan A Outstanding            $180M
  └─ Senior Notes (secured)             $70M
  └─ Subtotal Senior Secured Claims: $280M

  Secured Asset Value (under water)    $420M (total EV)
  Recovery: $280M / $280M = 100% ✓ Full recovery
  ─────────────────────────────────────

Layer 2: Junior Secured Lenders (2nd Lien)
  └─ Term Loan B Outstanding           $100M

  Available for Layer 2:                $140M remaining
  Recovery: $100M / $100M = 100% ✓ Likely full recovery
  ─────────────────────────────────────

Layer 3: Senior Unsecured Notes
  └─ Senior Unsecured Notes Outst.    $50M

  Available for Layer 3:                $40M remaining
  Recovery: $40M / $50M = 80% (80 cents on dollar)
  ─────────────────────────────────────

Layer 4: Mezzanine / Sub Notes
  └─ Mezzanine Debt Outstanding       $20M

  Available for Layer 4:                $0M remaining
  Recovery: $0M / $20M = 0% (0 cents on dollar)
  ─────────────────────────────────────

Layer 5: Preferred Stock
  └─ Preferred Equity Outstanding     $15M (cumulative, par)

  Recovery: $0M (equity layer, no recovery)
  ─────────────────────────────────────

Layer 6: Common Equity (Management + Sponsor)
  └─ Common Equity                    $150M

  Recovery: $0M (wiped out)
  ─────────────────────────────────────

RECOVERY SUMMARY TABLE:
┌─────────────────────────────────────────────────────┐
│ Instrument         │ Outstanding │ Recovery Rate │  │
├────────────────────┼─────────────┼───────────────┤  │
│ RCF                │ $30M        │ 100%          │  │
│ Term Loan A        │ $180M       │ 100%          │  │
│ Senior Notes (Sec) │ $70M        │ 100%          │  │
│ Term Loan B        │ $100M       │ 100%          │  │
│ Senior Unsecured   │ $50M        │ 80%           │  │
│ Mezzanine          │ $20M        │ 0%            │  │
│ Preferred          │ $15M        │ 0%            │  │
│ Common Equity      │ $150M       │ 0%            │  │
├────────────────────┼─────────────┼───────────────┤  │
│ TOTAL CLAIMS       │ $615M       │               │  │
└─────────────────────────────────────────────────────┘

Key Insight:
  ⚠️ Sponsor equity在持续经营重组情景下面临0% recovery
  ✓ Senior layers(至Term Loan B)相对安全
  ⚠️ Senior Unsecured Notes受损20%
  ✓ 结论: 债务结构中，不鼓励过度的Sub/Mezzanine发行
```

### 9.4 Recovery分析对融资结构的implications

```
Recovery Analysis → Debt Structure Optimization

根据上述恢复分析，建议结构：

✓ RECOMMENDED:
  • Senior Secured (RCF + TLA + Senior Notes): $280M (安全)
  • Term Loan B (2nd lien, but strong recovery): $100M (可接受)
  • 总杠杆: $380M (3.8x) 与$100M EBITDA相比

⚠️ CAUTION:
  • 限制Unsecured Notes至< $30M (recovery风险高)
  • 避免过多Mezzanine (在distress中无recovery)
  • 确保Senior层级"bite"足以覆盖所有senior claims

核心原则:
  在Elliott/Singer标准下，recovery analysis should drive capital structure,
  not just leverage ratios & covenant tests.
  一个看起来"绝对值"的杠杆(如4.5x)，如果在distress中导致
  equity+sub层完全wipeout + senior有haircut，就是过度.
```

---

## 十、Covenant 头部轨迹分析 (Covenant Headroom Trajectory)

不仅仅是年度covenant测试，而是12个季度的滚动头部轨迹。

### 10.1 债务/EBITDA头部图表

```
Debt / EBITDA Headroom Trajectory (12 quarters forward)
Max Covenant = 4.5x, Min Safe Headroom = 15%

  Leverage
    5.0x ┌─────────────────────────────────────────────
    4.8x │
    4.5x │ ════════ COVENANT MAX (400 bps)
    4.3x │     ╱──────────────────
    4.0x ├───╱      ↓ Q5 "Pinch Point"
    3.8x │ ╱ Entry leverage 4.0x
    3.5x │
    3.0x │
    2.5x └────────────────────────────────────────────
         Q1  Q2  Q3  Q4  Q5  Q6  Q7  Q8  Q9 Q10 Q11 Q12

  Scenario: Base Case
  ├─ Debt paydown: $5M per quarter
  ├─ EBITDA growth: 1.75% per quarter (7% annual)
  └─ Result: Steady de-leveraging to 2.5x by Q12

  ⚠️ RISK: Q4-Q6 headroom仅200bps
    → If EBITDA misses 3%, breach
    → Refinancing calendar must account for Q5 pressure
```

### 10.2 Interest Coverage Ratio头部图表

```
ICR Headroom Trajectory (12 quarters forward)
Min Covenant = 2.0x, Strong Headroom = >3.0x

  ICR
    4.5x │
    4.0x │                        ╱─── Strong
    3.5x │                    ╱──
    3.0x │ ═════════════════╱ Safe zone
    2.5x │              ╱
    2.0x ├────────────╱────────────── COVENANT MIN
    1.5x │        ╱ ← DANGER if dips below
         └─────────────────────────────────────────
         Q1  Q2  Q3  Q4  Q5  Q6  Q7  Q8  Q9 Q10 Q11 Q12

  Driver: Interest expense declining (debt paydown) + EBITDA growing
  Result: Improvement trajectory, no covenant risk

  ⚠️ If EBITDA shocks negative in Q3-4, could trigger covenant test
     → Need $15M+ liquidity buffer
```

### 10.3 关键Pinch Points识别

```
CRITICAL REFINANCING RISK WINDOW

Maturity Calendar:
  ├─ RCF: committed, no maturity pressure
  ├─ Term Loan A: Year 5 bullet (Q20) → Low risk (4 years away)
  ├─ Senior Notes: Year 8 bullet (Q32) → Manageable
  └─ Term Loan B: Year 3 bullet (Q12) → PINCH POINT

Q5-Q8 Pinch Point Drivers:
  • Term Loan B refinancing discussions should start Q2-3
  • Rising rate environment → refinancing costs increase
  • If EBITDA < guidance, debt capacity shrinks
  • Leverage reaches peak headroom constraint (4.3x)

Mitigation:
  ✓ Begin Term Loan B refinancing 12 months pre-maturity
  ✓ Build incremental headroom in Q1-Q3 (accelerated paydown)
  ✓ Maintain $20M+ minimum liquidity (RCF availability)
  ✓ Ensure EBITDA forecast cushion (model downside scenarios)
```

---

## 十一、法律/结构风险扫描 (Legal & Structural Risk Scan)

PE在Elliott/Singer风格的尽调中必须识别结构化legal风险。

### 11.1 Change of Control触发条款扫描

```
CHANGE OF CONTROL TRIGGER SCAN
─────────────────────────────────────────
Document: [Debt/Supplier/Customer Contract]
Risk: CoC 触发 → 额外成本/收入丧失/债务加速

Risk Inventory:
┌─────────────────────────────────────────────────────────┐
│ Contract Type     │ Trigger      │ Impact          │ 缓解 │
├────────────────────┼──────────────┼─────────────────┼─────┤
│ Debt documents     │ Ownership    │ Debt加速、利率  │ 债权人 │
│                    │ >50%变更     │ step-up         │ 同意书 │
├────────────────────┼──────────────┼─────────────────┼─────┤
│ 主要客户合同       │ 控制权变更   │ 合同可终止或    │ 客户 │
│                    │ >20-30%      │ 费率调整        │ 承诺书 │
├────────────────────┼──────────────┼─────────────────┼─────┤
│ 供应商合同         │ 控制权变更   │ 信用条款收紧    │ 供应商 │
│ (主要供应商)       │              │ 或费用增加      │ 承诺书 │
├────────────────────┼──────────────┼─────────────────┼─────┤
│ IP许可协议         │ 控制权变更   │ 许可权被回收    │ 许可方 │
│ (如重要)           │              │ 或费用增加      │ 同意书 │
├────────────────────┼──────────────┼─────────────────┼─────┤
│ 员工与关键人       │ 变更控制权   │ 留任奖金支付    │ 留任协议 │
│ retention pools    │              │ / Key person离职│ & carve-out │
└─────────────────────────────────────────────────────────┘

行动项:
  □ 审查前50大客户合同 CoC条款
  □ 审查前10大供应商合同 CoC条款
  □ 债务文件：争取CoC waiver / 明确步骤成本
  □ 软件许可：如critical，获取许可方同意书
  □ 员工：提前lock in 关键人留任安排

Quantification:
  潜在成本 = [客户流失revenue $XX M] + [债务step-up $X M] + [供应商费用增加 $X M]
  需在Sources & Uses中计提contingency
```

### 11.2 MAC条款评估 (Material Adverse Change)

```
MAC CLAUSE ASSESSMENT
────────────────────────────────
MAC定义：法律/经济事件导致业务价值显著(通常>20%)降低

在融资文件中的风险:
  • 某些贷款人在发生定义的MAC时可declare default
  • 历史上在经济衰退/疫情时实际触发过

评估框架:
  ├─ Debt文件中MAC定义的宽度: Narrow (可取) vs Broad (高风险)
  ├─ 排除条款: 既有市场/经济环境是否排除
  ├─ Materiality Threshold: 数值型 vs 定性型
  └─ 例外条款: 行业通用影响是否excluded

建议措施:
  ✓ Negotiate narrow MAC definition in financing docs
  ✓ Include broad carve-outs (industry, general economic, government action)
  ✓ Obtain MAC insurance if available & material
  ✓ Document baseline business state (Day 1 representations)
```

### 11.3 跨境/异地结构风险 (VIE、离岸持股)

```
CROSS-BORDER STRUCTURE RISK
──────────────────────────────
For Chinese Companies & Offshore Structures:

VIE (Variable Interest Entity) Risks:
  ├─ 中国监管变化可能导致VIE contracts 无效
  ├─ PE所有权可能被中国政府challenge
  ├─ 现金汇出受限可能性
  └─ 贷款方可能要求VIE unwinding或保障

典型VIE结构:
  Offshore PE Fund (Delaware)
    └─ Offshore HoldCo (BVI/Cayman)
        └─ PRC Operating Company (WFOEs)
            └─ VIE Contracts with PRC Opcos

评估:
  □ VIE合同合法性: 中国律师意见
  □ 汇款权: 中国税务和外汇政策变化风险
  □ 监管趋势: 中国政府对VIE态度(Tech严管)
  □ 保险: 是否可获VIE insurance

融资影响:
  • Lenders可能要求额外担保或escrow
  • 融资成本可能增加 50-100bps
  • Maturity可能被shortening至3-4年(vs 5-7年)

建议:
  ✓ 获取PRC & offshore双重法律意见
  ✓ 向lenders充分披露VIE风险
  ✓ Consider partial unwinding或WFOE restructure (if feasible)
```

### 11.4 交叉违约/交叉加速条款 (Cross-Default)

```
CROSS-DEFAULT CLAUSE RISK
──────────────────────────────
交叉违约: 一个债务instrument违约 → 自动触发其他debt默认

典型trigger:
  "任何债务 > $XX M 的违约将导致本facility下所有债务加速"

风险:
  • 一个小债务(如supplier financing)违约 → 银行贷款被加速
  • Covenant breach（如leverage test失败）→ 全面default

缓解:
  □ Negotiate materiality thresholds: 仅 > $10-20M 债务触发
  □ Cure period: 通常给予30-60天补救时间
  □ Cross-default carve-outs:
      - Supplier/trade credit excluded
      - 被动违约(non-payment by customer) excluded
      - 自动生效延缓/豁免条款

尽调Action:
  • 映射所有outstanding debt
  • 识别任何existing default / technical breach
  • 评估相关covenant headroom
  • 确保financing close前所有breach cured
```

---

## 十二、IRR 合理性检查 (IRR Reasonableness Checkpoint)

LBO成败最终以IRR衡量。此检查表防止结构化过度承诺。

### 12.1 IRR范围合理性表

| IRR范围 | 评估 | 风险 | 行动 |
|---------|------|------|------|
| **< 12%** | 回报不足 | Capital opportunity cost过高 | 重新议价、增加杠杆、削成本 |
| **12% - 20%** | 正常LBO范围下限 | 可接受但margin窄；对执行敏感 | 强化management、明确value levers |
| **20% - 30%** | 良好/高回报 | 对假设敏感性强 | 深入敏感性测试，验证主要驱动 |
| **> 30%** | 异常高 | **MUST CHECK:** 基础假设可靠性 | 见12.2 |

### 12.2 异常高IRR (> 30%)的检查清单

```
当计算IRR > 30% 时，触发以下检查:

□ NET CASH DEFINITION CHECK:
   • 确认"Net Cash"定义(Debt净额减去可用liquidity)
   • 有无隐藏的operating cash requirements
   • 有无overstated exit proceeds (过度乐观multiple)

□ DEBT CAPACITY CHECK:
   • Entry leverage vs 业界可比: 我们4.0x，可比2.5-3.5x?
   • 利率假设: 我们5.5% vs market 6.0-6.5%?
   • 有无artificially低估费用?

□ EXIT MULTIPLE CHECK:
   • 出口倍数假设: 我们12x，历史可比8-10x?
   • 增长假设: CAGR 我们8%，行业4-6%?
   • 有无包含unlikely add-on scenarios

□ EXECUTION DISCOUNT CHECK:
   • 是否使用了过度optimistic的discount rate
   • Management track record真的值得-2%折扣吗

□ 时间敏感性:
   • 若Year 3 (而非Year 5)出售，IRR仍>25%?
   • 若出口倍数 -10%, IRR跌至多少

□ 场景分析:
   • Base case IRR = 28% ✓
   • Bear case (EBITDA -20%, exit 1x lower) = ?
     └─ 若Bear IRR < 8%, 风险过大

┌─────────────────────────────────────────────┐
│ RED FLAG CHECKLIST:                         │
│ ☐ Entry leverage > 4.5x (above market)     │
│ ☐ Exit multiple > 13x (above comps)        │
│ ☐ Management discount < -3% (over-bullish) │
│ ☐ EBITDA growth CAGR > 9% (vs 6% industry) │
│ ☐ No add-on M&A in base case (under-lever) │
│ ☐ Bear case IRR < 10% (insufficient cushion)│
│ ☐ Zero interim cash returns assumed        │
└─────────────────────────────────────────────┘

如发现RED FLAG，不应proceed unless resolved
```

### 12.3 行业基准 & PE基金目标

```
KKR / Blackstone / Apollo 目标IRR 范围 (2023-2025):

Lower Mid-Market LBO (EBITDA $10-50M):  22% - 32% 目标
  理由: 较高风险，较少流动性选项，需要higher returns

Mid-Market LBO (EBITDA $50-200M):       18% - 28% 目标
  理由: 平衡风险/流动性，标准LBO结构

Upper Mid-Market (EBITDA $200M+):       15% - 25% 目标
  理由: 规模优势，更多退出选项，接受lower IRR

我们交易特征:
  • EBITDA $100M → Mid-market
  • 目标IRR: 20-28% (best-in-class team) 或 18-25% (average team)
  • 若computed IRR > 30%，high-case scenario须有强逻辑支撑
  • 若computed IRR < 16%，需重估交易价格或结构

Sensitivity for sensibility:
  ✓ 25% baseline IRR + 200bps management execution benefit
    = 适度aggressive but defensible

  ⚠️ 32% baseline IRR + 300bps execution benefit
    = Requires extraordinary assumptions & strong risk mitigation
```

---

## 十三、多法人合并引擎 (Multi-Entity Consolidation Engine)

> **💀 曼巴注记 — 来源**: Project Jewel (MBK Partners, 日本家庭保险LBO) — 65张sheets，5个运营子公司各自独立建模后合并。没有这套引擎，你在做的不是LBO，是猜数字。

### 13.1 架构原则

多法人LBO中，每个子公司必须是**独立的自足模型**，而非母公司的一行拆分。

```
集团控股公司 (HoldCo)
├─ 子公司A (独立P&L + BS + CF) ← 完整独立模型
├─ 子公司B (独立P&L + BS + CF) ← 完整独立模型
├─ 子公司C (独立P&L + BS + CF) ← 完整独立模型
├─ 合并调整 (Consol Adj.) ← 内部交易消除
└─ 合并报表 (Consolidated) ← 集团层面输出
```

### 13.2 子公司独立模型最低要求

每个子公司模型sheet必须包含：

| 模块 | 最低行数 | 内容 |
|------|---------|------|
| Revenue Build-up | 50+ rows | 按Core第5节颗粒度规范，独立驱动假设 |
| Operating Expenses | 30+ rows | COGS/SG&A/R&D独立假设，不继承母公司% |
| EBITDA → Net Income | 15+ rows | 完整P&L推导 |
| CapEx & D&A | 10+ rows | 独立资产滚动表 |
| Working Capital | 10+ rows | AR/Inventory/AP Days独立假设 |

**铁律**: 子公司之间的假设必须独立。不允许"子公司B增速 = 子公司A × 0.8"这种偷懒公式。

> **💀 判断注记 — 为什么不能偷懒？**
> Project Jewel案例中，JWS（保修服务）和ACTG（会计服务）的收入驱动完全不同：JWS靠会员数×ARPU，ACTG靠交易量×单价。如果用母公司总增速拆分，Year 3的预测偏差达到23%。原因是两个子公司处于不同的生命周期阶段。

### 13.3 内部交易消除 (Intercompany Elimination)

```
合并调整表 (Consol Adj. Sheet):

1. 内部收入消除 (Revenue Elimination)
   ├─ 子公司A向B提供服务 → A的Revenue中扣除
   ├─ 同时B的COGS中扣除对应成本
   └─ 净影响: 合并层Revenue和COGS同时减少

2. 内部应收应付消除 (AR/AP Elimination)
   ├─ A的应收 = B的应付 → 合并层双向抵消
   └─ 验证: 内部AR余额 ≡ 内部AP余额 (不平则有错误)

3. 内部投资消除 (Investment Elimination)
   ├─ 母公司对子公司的股权投资 → 与子公司净资产抵消
   └─ 差额 = 商誉 (Goodwill)

4. 少数股东权益 (Minority Interest)
   ├─ 如子公司非100%控股
   └─ 合并Net Income × (1 - 持股%) = 少数股东损益
```

**消除公式模板**:
```
Consol Revenue = Σ(各子公司Revenue) - Intercompany Revenue
Consol COGS = Σ(各子公司COGS) - Intercompany COGS
Consol Net Income = Σ(各子公司NI) - Minority Interest

验证: Consol Revenue ≠ Σ(子公司Revenue) → 消除必须生效
```

### 13.4 共享服务成本分摊 (Shared Service Allocation)

```
SHARED SERVICE COST ALLOCATION FRAMEWORK
─────────────────────────────────────────
来源: 母公司/控股层面的成本中心

分摊方法选择 (按优先级):
  1. 直接归属法 (Direct Attribution): 成本可直接追溯到子公司 → 100%归属
  2. 使用量法 (Usage-Based): 按实际使用量分摊 (如IT系统按用户数、HR按员工数)
  3. 收入比例法 (Revenue Pro-rata): 按各子公司收入占比分摊 → 仅在无法直接归属时使用

分摊表模板:
┌──────────────┬──────────┬──────────┬──────────┬──────────┐
│ 成本类别      │ 总额     │ 子公司A  │ 子公司B  │ 子公司C  │
├──────────────┼──────────┼──────────┼──────────┼──────────┤
│ IT系统        │ $5M      │ $2M(40%) │ $2M(40%) │ $1M(20%) │
│ HR/行政       │ $3M      │ $1.2M    │ $1.2M    │ $0.6M    │
│ 法务/合规     │ $2M      │ $0.8M    │ $0.8M    │ $0.4M    │
│ 财务/审计     │ $1.5M    │ $0.6M    │ $0.6M    │ $0.3M    │
│ 管理层薪酬    │ $4M      │ 直接归属  │ 直接归属  │ 直接归属  │
└──────────────┴──────────┴──────────┴──────────┴──────────┘

分摊基础:
  IT系统: 按用户数 (A=200, B=200, C=100)
  HR/行政: 按员工数 (A=500, B=500, C=250)
  法务/合规: 按收入比例
  财务/审计: 按收入比例
```

### 13.5 合并验证检查表

```
CONSOLIDATION VALIDATION CHECKLIST
───────────────────────────────────
☐ 各子公司P&L独立balance (Revenue - Costs = NI ✓)
☐ 内部交易消除后，合并Revenue < Σ(子公司Revenue)
☐ 内部AR ≡ 内部AP (差额 = 0)
☐ 少数股东权益正确计算 (如适用)
☐ 商誉 = 收购价 - 子公司净资产公允价值 (如适用)
☐ 合并CF中，内部股利/分红已消除
☐ 共享服务成本分摊总额 = 母公司层面总成本 (不多不少)
☐ 合并BS两边平衡 (Assets = Liabilities + Equity)

FAILURE MODE: 如任何检查失败，停止建模，回溯找到错误源。
常见错误源: 时间错位(子公司fiscal year不同)、币种未统一、消除遗漏。
```

---

## 十四、Covenant季度自动化测试引擎 (Quarterly Covenant Automation Engine)

> **💀 曼巴注记 — 来源**: Project Jewel的CF_Credit sheet实现了真正的季度covenant自动监控。大多数模型只做年度测试，但真实世界中银行每季度测一次——Year 1 Q3-Q4往往是最危险的"pinch point"。

### 14.1 测试频率与触发机制

```
COVENANT TESTING CADENCE
─────────────────────────
测试频率: 每季度 (Q1/Q2/Q3/Q4)
报告频率: 季度结束后30天内
触发机制:
  ├─ 自动测试: 每季度末自动计算所有covenant指标
  ├─ Breach预警: 如headroom < 15%，触发黄灯
  ├─ Breach确认: 如实际breach，触发红灯 + cash sweep机制
  └─ Cure期: 通常30-60天补救窗口
```

### 14.2 季度Covenant监控表

```
QUARTERLY COVENANT DASHBOARD (20 quarters forward)
──────────────────────────────────────────────────────────────
指标          │ Covenant │ Q1Y1 │ Q2Y1 │ Q3Y1 │ Q4Y1 │ Q1Y2 │ ...
──────────────┼──────────┼──────┼──────┼──────┼──────┼──────┼────
Debt/EBITDA   │ ≤ 4.5x   │ 4.0x │ 3.9x │ 3.8x │ 3.7x │ 3.5x │
  Headroom    │          │ 11%  │ 13%  │ 16%  │ 18%  │ 22%  │
  Status      │          │ 🟡   │ 🟡   │ 🟢   │ 🟢   │ 🟢   │
──────────────┼──────────┼──────┼──────┼──────┼──────┼──────┼────
ICR           │ ≥ 2.0x   │ 2.3x │ 2.4x │ 2.5x │ 2.7x │ 2.9x │
  Headroom    │          │ 15%  │ 20%  │ 25%  │ 35%  │ 45%  │
  Status      │          │ 🟡   │ 🟢   │ 🟢   │ 🟢   │ 🟢   │
──────────────┼──────────┼──────┼──────┼──────┼──────┼──────┼────
DSCR          │ ≥ 1.2x   │ 1.4x │ 1.5x │ 1.5x │ 1.6x │ 1.7x │
  Headroom    │          │ 17%  │ 25%  │ 25%  │ 33%  │ 42%  │
  Status      │          │ 🟢   │ 🟢   │ 🟢   │ 🟢   │ 🟢   │
──────────────┼──────────┼──────┼──────┼──────┼──────┼──────┼────
Min Liquidity │ ≥ $15M   │ $18M │ $20M │ $22M │ $25M │ $28M │
  Status      │          │ 🟢   │ 🟢   │ 🟢   │ 🟢   │ 🟢   │
──────────────────────────────────────────────────────────────

Status Legend: 🟢 Headroom > 15% │ 🟡 Headroom 5-15% │ 🔴 Breach

公式:
  Headroom = (Actual - Covenant) / Covenant × 100%
  Status = IF(Headroom > 15%, "🟢", IF(Headroom > 5%, "🟡", "🔴"))
```

### 14.3 Excess Cash Sweep机制

```
CASH SWEEP MECHANICS
────────────────────
触发条件: Leverage > [Sweep Trigger Level] (通常 3.5x-4.0x)
Sweep比例: 按leverage bracket递增

Leverage Range    │ Sweep % of Excess CF │ 说明
< 3.0x            │ 0%                   │ 无强制偿还
3.0x - 3.5x       │ 25%                  │ 轻度sweep
3.5x - 4.0x       │ 50%                  │ 中度sweep
> 4.0x             │ 75%                  │ 重度sweep (保留最低流动性)

Excess CF定义:
  Excess CF = Operating CF - Maintenance CapEx - Required Amortization - Tax

Sweep分配优先级:
  1. RCF余额偿还 (最优先)
  2. Term Loan A摊销加速
  3. Term Loan B (如有excess)

注意: Sweep后必须重新测试Covenant，确认sweep不会导致流动性低于最低要求。
```

### 14.4 Covenant Breach应急预案

```
BREACH RESPONSE PROTOCOL
─────────────────────────
Level 1 — 预警 (Headroom < 15%):
  ├─ 通知: 向CFO和Sponsor发出预警
  ├─ 行动: 重新审视当季EBITDA预测，识别提升空间
  └─ 频率: 切换为月度监控

Level 2 — 技术性Breach (Covenant刚好触及):
  ├─ 通知: 正式通知银行团
  ├─ 行动: 启动30天Cure Period
  │   ├─ 选项A: 注入额外股权 (Equity Cure)
  │   ├─ 选项B: 出售非核心资产产生现金
  │   └─ 选项C: 请求银行Waiver (通常需支付费用)
  └─ 文档: 准备详细的remediation plan

Level 3 — 实质性Breach (连续2季度):
  ├─ 通知: 触发Cross-default条款审查
  ├─ 行动: 启动全面重组讨论
  │   ├─ 聘用重组顾问
  │   ├─ 准备修订后的Business Plan
  │   └─ 评估Standstill Agreement可能性
  └─ 后果: 可能导致debt acceleration
```

---

## 十五、多退出路径建模引擎 (Multi-Exit Pathway Modeling Engine)

> **💀 曼巴注记 — 来源**: Project Stratus建模了4种截然不同的退出路径(Put/IPO/Block Trade/Drag-Along)，每种都有独立的IRR/MoE计算。只建一种退出路径的模型是"一条腿走路"。

### 15.1 四标准退出路径

| 路径 | 触发条件 | IRR计算特殊性 | 典型适用场景 |
|------|---------|-------------|------------|
| **A: 协议回购 (Put Option)** | 原股东/创始人回购PE持股 | 回购价 = 约定公式(通常EBITDA×约定倍数) | 少数股权投资、战略合作 |
| **B: IPO退出** | 公司达到上市标准 | IPO定价 - 承销费 - lock-up折扣 | 成熟期、增长故事明确 |
| **C: 大宗交易 (Block Trade)** | IPO后PE抛售持股 | 市价 - 大宗折扣(5-10%) - 交易费 | IPO后锁定期结束 |
| **D: Drag-Along (强制出售+Upside Sharing)** | 控股方行使强制出售权 | 含IRR hurdle + 超额收益分成 | 控股/并购退出 |

### 15.2 每条路径的独立计算模块

**路径A: Put Option Exit**
```
Put Price = EBITDA(exit year) × Agreed Multiple
         OR = Fixed Formula (如: Book Value × 1.2)

PE Exit Proceeds:
  = Put Price × PE Stake %
  - Transaction Fees (1-2%)

Cash Flow Timeline:
  T0: -Investment Amount
  T1-Tn: +Interim Dividends (if any)
  Tn: +Put Exit Proceeds

IRR = XIRR(Cash Flows, Dates)
MoE = Total Proceeds / Investment
```

**路径B: IPO Exit**
```
IPO Valuation:
  Forward EBITDA(exit year+1) × Public Market Multiple
  OR: Revenue × Revenue Multiple (for growth companies)

IPO Adjustments:
  - Underwriting Discount: 3-7% of proceeds
  - IPO Expenses: $2-5M
  - Over-allotment (Greenshoe): +15% if exercised
  - Lock-up Period: 通常180天 → 延迟变现

PE Exit Equity:
  = (IPO Market Cap - Net Debt) × PE Stake %
  - IPO Costs attributable to PE

特殊考量:
  ☐ Dilution from IPO primary shares
  ☐ ESOP/Option pool dilution at IPO
  ☐ Lock-up period discount to NPV (3-5%)
```

**路径C: Post-IPO Block Trade**
```
Block Trade Proceeds:
  = Market Price at Block Date
  × PE Shares
  × (1 - Block Discount %) [通常5-10%]
  - Broker Fee [通常1-2%]

时间线:
  Lock-up结束后6-12个月执行
  需考虑市场窗口(避开earnings blackout)

IPO→Block Trade总收益:
  = IPO期间部分套现 + Block Trade净收益
```

**路径D: Drag-Along with Upside Sharing**
```
Drag-Along Mechanics:
  控股方(Sponsor)行使强制出售权 → 所有股东必须按同条件出售

Upside Sharing (来自Stratus模型):

  Case I — MOU-based Sharing:
    IF actual IRR > Hurdle IRR (如25%)
    THEN upside_amount = (Actual_Proceeds - Hurdle_Proceeds) × Sharing %
    Sharing % 通常 = 10-20% 给管理层/原股东

  Case II — SHA-based Tiered Sharing:
    Tier 1: IRR 25-40% → Sharing 15%
    Tier 2: IRR > 40% → Sharing 20%

  Adjusted Returns:
    Sponsor IRR = XIRR(Sponsor CFs after sharing)
    Management IRR = XIRR(Mgmt CFs including sharing bonus)
```

### 15.3 退出路径并行比较表

```
EXIT PATHWAY COMPARISON MATRIX
──────────────────────────────────────────────────────────
                │ Put Option │ IPO      │ Block Trade │ Drag-Along │
────────────────┼────────────┼──────────┼─────────────┼────────────┤
Exit Year       │ Year 3-4   │ Year 4-5 │ Year 5-6    │ Year 3-5   │
Exit Multiple   │ 7-8x agreed│ 12-15x   │ 10-12x mkt  │ 9-11x nego │
Gross Proceeds  │ $XXM       │ $XXM     │ $XXM        │ $XXM       │
Fees/Discounts  │ 1-2%       │ 5-8%     │ 6-12%       │ 2-3%       │
Net Proceeds    │ $XXM       │ $XXM     │ $XXM        │ $XXM       │
IRR             │ XX%        │ XX%      │ XX%         │ XX%        │
MoE             │ X.Xx       │ X.Xx     │ X.Xx        │ X.Xx       │
Risk Level      │ Low-Med    │ Med-High │ High        │ Medium     │
Controllability │ High       │ Low      │ Low         │ High       │
──────────────────────────────────────────────────────────

选择原则:
  • 控制权>50%: 倾向Drag-Along或Strategic Sale
  • 少数股权: 倾向Put Option (保底) + IPO (上行)
  • 市场环境好: IPO → Block Trade序列
  • 市场环境差: Put Option作为保底，等待窗口
```

### 15.4 退出路径QA检查

```
EXIT PATHWAY VALIDATION
───────────────────────
☐ 至少建模3条退出路径 (2条主路径 + 1条保底)
☐ 每条路径都有独立的IRR/MoE计算
☐ Put Option(如有)的约定公式已验证(不依赖市场)
☐ IPO路径包含dilution、lock-up、费用调整
☐ Drag-Along的upside sharing机制已正确建模
☐ 所有路径的现金流时间线一致(T0 = 投资日)
☐ 并行比较表已生成，供IC决策使用
```

---

## 十六、行业专属建模附录 — 保险/准备金 (Industry Module: Insurance & Reserves)

> **💀 曼巴注记 — 来源**: Project Jewel的保险子公司模型包含准备金充足性测试、理赔率建模、再保险分出跟踪——这些是金融服务LBO的"隐形地雷"。忽略它们等于在地雷阵里闭眼走路。

### 16.1 适用场景

当标的公司涉及以下业务时，强制加载本模块：
- 保险 (财产/人寿/健康/再保险)
- 保修/延保服务
- 金融担保
- 任何含"准备金"或"浮存金"的业务

### 16.2 保费收入确认模型

```
保费收入 ≠ 收到的保费

PREMIUM REVENUE RECOGNITION:
─────────────────────────────
收到保费 (Gross Written Premium, GWP):
  = 保单数 × 年化保费

已赚保费 (Earned Premium):
  = GWP × 赚取比例 (按保单生效天数/保单总天数)

未赚保费 (Unearned Premium Reserve, UPR):
  = GWP - Earned Premium
  = 负债项 (Balance Sheet Liability)

公式:
  Revenue = Earned Premium = GWP × (Coverage Days Elapsed / Total Coverage Days)

  Year 1 GWP = $100M (新保单$30M + 续保$70M)
  Year 1 Earned = $85M (部分新保单尚未完全赚取)
  Year 1 UPR = $15M (BS负债增加)
```

### 16.3 理赔准备金模型

```
CLAIMS RESERVE MODEL
────────────────────
Loss Ratio = Incurred Claims / Earned Premium

三层准备金:
  1. 已报告未决赔款 (Outstanding Claims Reserve, OCR)
     = 已收到理赔申请但尚未支付的金额

  2. 已发生未报告赔款 (IBNR - Incurred But Not Reported)
     = 估计已发生但尚未收到报告的赔款
     = Earned Premium × IBNR Factor (通常5-15%)

  3. 理赔调整费用 (LAE - Loss Adjustment Expense)
     = 处理理赔的运营成本
     = Incurred Claims × LAE Ratio (通常8-12%)

总准备金 = OCR + IBNR + LAE

P&L影响:
  准备金增加 → P&L费用增加 → EBITDA减少
  准备金释放 → P&L收入增加 → EBITDA增加

⚠️ 陷阱: 管理层可能通过操纵IBNR因子来"管理"利润。
检测: 追踪IBNR/Earned Premium比率的季度趋势，异常波动需调查。
```

### 16.4 再保险分出 (Reinsurance Ceded)

```
REINSURANCE MODEL
─────────────────
再保险 = 保险公司将部分风险转移给再保险公司

Ceded Premium = GWP × Cession Rate (通常20-40%)
Retained Premium = GWP - Ceded Premium

Ceded Claims = Gross Claims × Cession Rate (同比例)
Net Claims = Gross Claims - Ceded Claims

P&L展示:
  Gross Written Premium        $100M
  - Ceded Premium              ($25M)  ← 转移给再保险公司
  ─────────────────────────────────
  Net Written Premium           $75M
  × Earning Rate                85%
  ─────────────────────────────────
  Net Earned Premium            $63.75M  ← 真正的Revenue

  Net Claims Incurred           ($38M)
  Net Loss Ratio                60%      ← 目标: 55-65%
```

### 16.5 保险浮存金投资收益 (Insurance Float)

```
FLOAT INVESTMENT MODEL
──────────────────────
浮存金 = 已收保费但尚未支付理赔的资金
       = UPR + Claims Reserves - Reinsurance Receivable

浮存金规模 = 通常为Annual Premium的30-50%

投资收益:
  Float × Investment Yield (2-4% for conservative portfolio)
  = $50M × 3% = $1.5M/年 → 直接贡献Pre-tax Income

投资组合限制:
  ├─ 政府债券: 40-60% (安全)
  ├─ 企业债券: 20-30% (AA级以上)
  ├─ 权益类: 0-10% (监管限制)
  └─ 现金: 10-20% (流动性)

⚠️ LBO影响: 高杠杆可能导致rating agency降级 → 投资组合限制收紧 → 投资收益下降
```

### 16.6 保险行业建模QA检查

```
INSURANCE MODULE VALIDATION
───────────────────────────
☐ Revenue = Earned Premium (非GWP)
☐ UPR正确计算且在BS负债端
☐ IBNR准备金有独立计算(非手工调整)
☐ Loss Ratio趋势合理(55-70%，行业对标)
☐ 再保险分出比例与合同一致
☐ 浮存金投资收益独立建模(非假设固定数)
☐ 准备金变动反映在P&L中
☐ 监管资本要求(Solvency)已考虑
```

---

## 十七、Buyout Track模块的整体集成

### 17.1 模块加载触发条件

```
Step 0 Track Selection:
  ├─ "Track A: Growth / Platform Investment" → 加载growth-specific模块
  ├─ "Track B: PE Buyout / LBO" → 加载本 integrated-modeling-bt 模块
  └─ "Track C: Full Integrated" → 加载 integrated-modeling-bt + 综合模块

本模块关键依赖:
  • Base financial model from integrated-modeling-base
  • Management assessment framework from integrated-modeling-mgmt (if exists)
  • Capital structure library from integrated-modeling-cap (if exists)

输出产物:
  • LBO-specific business case (Sources & Uses, covenant analysis)
  • Management investability scorecard
  • 100-Day Plan with financial roadmap
  • Exit strategy memo with IRR sensitivities
  • Recovery analysis waterfall
  • Financing structure term sheet summary
```

### 17.2 QA / 质量检查Gate

在model最终化前，确保以下gate通过:

```
QUALITY ASSURANCE CHECKLIST FOR BT MODULE
──────────────────────────────────────────

□ Management Assessment:
  ✓ CEO 评分明确，且 <6 的触发Layer 2警告
  ✓ 执行折扣范围 已根据industry调整
  ✓ Key person 离职风险 已量化

□ Debt Capacity:
  ✓ FCF conversion 与industry benchmark对标
  ✓ ICR test 明确显示Year 1 headroom
  ✓ 推荐债务天花板 有定量支撑

□ Capital Structure:
  ✓ 所有debt layers 均有rate type, amortization, collateral, covenants 定义
  ✓ Repayment waterfall 已建立 & cross-default triggers 明确
  ✓ Sources & Uses balance + 费用合理

□ 100-Day Plan:
  ✓ Day 1-30, 31-60, 61-100 分段具体
  ✓ 所有initiatives 有量化targets & 18月里程碑
  ✓ Year 1 EBITDA = Base × (1 - Exec Discount) + gradual value creation
  ✓ 未假设full run-rate in Year 1

□ Exit Blueprint:
  ✓ 3-4条退出路径 已评估
  ✓ 每条路径的时间/倍数/风险 明确
  ✓ 持有期敏感性分析 覆盖Year 3-7

□ Recovery Analysis:
  ✓ Liquidation vs going-concern scenarios 已建立
  ✓ Recovery waterfall 显示各layer的回收率
  ✓ 结果inform了capital structure decisions

□ Covenant Headroom:
  ✓ 12季度 Debt/EBITDA 和 ICR 轨迹图
  ✓ Pinch points & refinancing风险 已识别
  ✓ 流动性buffer 足够应对应力

□ Legal / Structural Risks:
  ✓ CoC triggers 已审计 & 缓解计划制定
  ✓ MAC条款 已评估，必要时获保险
  ✓ 若涉及VIE/跨境，已获double legal opinion
  ✓ Cross-default triggers 已映射

□ IRR Check:
  ✓ Computed IRR 在行业合理范围内 (18-28% for mid-market)
  ✓ 若 > 30%，red flag已逐项检查 & resolve
  ✓ Bear case scenario IRR > 10% (cushion)
  ✓ 敏感性测试 覆盖±20% entry assumptions

若任何gate FAIL，Return to model refinement前不应proceed.
```

### 17.3 Buyout Track标准输出包 (Deliverables)

本模块应生成以下integrated outputs:

```
1. INVESTMENT SUMMARY
   • Investment thesis (2-3 pages)
   • Key value drivers & risks
   • Management team assessment
   • Recommended capital structure with rationale

2. FINANCIAL MODEL (5-statement integrated)
   • Base case 5-year projections with exec discount applied
   • Alternative scenarios (bull/bear/stress)
   • Debt schedule & covenant compliance tracking
   • Exit value bridge (entry to exit multiple)

3. SOURCES & USES
   • Detailed breakdown by debt tranche
   • Transaction fees & implementation costs
   • Equity investment by source (sponsor, mgmt, employees)

4. 100-DAY PLAN DECK
   • Stabilization (Day 1-30) action items
   • Diagnosis (Day 31-60) findings summary
   • Launch (Day 61-100) detailed initiatives
   • Financial roadmap: milestones + cumulative value

5. EXIT STRATEGY MEMO
   • Feasibility of IPO / Strategic / Secondary / Dividend
   • Holding period IRR sensitivity table
   • Management build-out needs per exit path

6. DEBT STRUCTURE MEMO
   • Capital structure diagram (全层级)
   • Covenant test schedule (annual + quarterly)
   • Covenant headroom trajectory chart
   • Refinancing risk calendar

7. RISK ASSESSMENT REPORT
   • PE-specific risks: key person, change of control, leverage
   • Recovery analysis: liquidation + going-concern waterfall
   • Legal/structural risk scan: MAC, VIE, cross-default
   • Sensitivity analysis & scenario stress tests

8. IRR WATERFALL & SENSITIVITIES
   • Entry value bridge
   • EBITDA growth contribution to returns
   • Multiple expansion / contraction impact
   • Exit timing impact on IRR
   • Sensitivity matrix: entry multiple vs exit multiple vs hold period

全部输出应以模型驱动，每项假设可trace back to基础数据或行业基准.
```

---

## 结语

本**Buyout Track (BT)** 模块遵循**Paul Singer / Elliott Management** 企业级尽调标准，涵盖：

1. **管理层可投性** - 多维度量化评估，执行折扣与行业调整
2. **PE特有风险** - 杠杆放大敏感性、关键人风险、CoC触发
3. **债务容量与结构** - FCF转换、ICR测试、复杂债务层级建模、Recovery分析
4. **百日计划** - 分阶段稳定、诊断、启动，与财务目标绑定
5. **退出蓝图** - IPO / Strategic / Secondary路径，时间与倍数灵敏度
6. **Legal & Covenant** - CoC扫描、MAC条款、VIE风险、12季度headroom轨迹
7. **IRR合理性** - Red flag检查，基准对标，敏感性验证
8. **多法人合并** - 子公司独立建模、内部交易消除、共享服务分摊
9. **Covenant自动化** - 季度监控、Cash Sweep机制、Breach应急预案
10. **多退出路径** - Put/IPO/Block Trade/Drag-Along并行建模
11. **行业专属模块** - 保险/准备金/浮存金/再保险专业化建模

**关键原则：** 定量化、透明化、场景化、stress-tested。
任何LBO model须经本模块Quality Assurance gate通过，方可进入交易决策阶段。

---

**版本:** v2.0 (2026-03-20) — 曼巴标准增强版：注入Project Jewel/Stratus毕生所学技术
**适用范围:** PE / LBO交易分析、中型以上企业收购、多法人集团LBO、金融服务行业专业化
**合规标准:** Elliott Management / KKR / Blackstone / MBK Partners尽调水准
**更新周期:** 年度或根据市场环境变化
