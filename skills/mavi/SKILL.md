---
name: mavi
type: protocol
version: 2.0.0-standalone
description: "MAVI — Multi-Agent Variable Interrogation Protocol（独立版）。7 个分析 Agent 多视角审讯核心变量。每个 Agent 动态引用对应的投资大师 Skill（不内嵌 Reference），产出 Variable Lock Sheet + Macro Overlay + Catalyst Calendar + Standard Analysis Signal。可独立使用（纯讨论），也可对接 financial-modeling（完整流程）。"
---

# MAVI v2.0 — Multi-Agent Variable Interrogation Protocol (Standalone)

## 独立使用 vs 套件使用

**独立使用**：MAVI 是一个完整的多视角分析产品。给它一家公司，它会用 7 个投资大师的视角审讯关键变量，产出 Variable Lock Sheet 和分析报告。不需要建模。

**套件使用**：在 Investment Intelligence Suite 中，MAVI 插入在 Layer 1（8D 诊断）和 Layer 2（参数锁定）之间。MAVI 的 Variable Lock Sheet 直接作为 financial-modeling 的 Layer 2 输入。

---

## 前置条件

### 必需的投资大师 Skill（按 Agent 依赖）

MAVI 不内嵌投资哲学内容——每个 Agent 在执行时**动态读取**对应的投资大师 Skill。

| Agent | 必需 Skill | 读取的 Reference 文件 |
|-------|-----------|---------------------|
| CA 保守锚点 | `howard-marks/` | `references/02-risk.md`, `references/04-contrarian-value.md`, `references/06-defensive-investing.md` |
| GTA 成长引擎 | `cathie-wood/` | `references/01-five-platforms.md`, `references/02-wrights-law.md`, `references/03-s-curve-adoption.md` |
| IC 行业标尺 | （内置，无外部依赖） | — |
| RC 风险挑战者 | `jim-chanos/` | `references/01-forensic-accounting.md`, `references/02-fraud-patterns.md`, `references/07-institutional-failure.md` |
| HC 历史校准器 | `ray-dalio/` | `references/01-economic-machine.md`, `references/02-debt-cycles.md`, `references/05-deleveraging.md` |
| MA 宏观架构师 | `ray-dalio/` + `george-soros/` | `dalio/references/03-all-weather.md`, `soros/references/02-boom-bust-cycle.md` |
| EA 事件裁判 | `john-paulson/` + `paul-singer/` | `paulson/references/01-merger-arb-framework.md`, `singer/references/01-activist-playbook.md` |

**辅助 Skill（可选，增强深度）**：

| Agent | 辅助 Skill | 何时加载 |
|-------|-----------|---------|
| CA | `buffett/references/06-valuation-capital.md` | 估值类变量 |
| GTA | `eric-sanchez-ai/references/01-physical-layer.md` | 科技/AI/半导体公司 |
| RC | `george-soros/references/01-reflexivity.md` | 高杠杆/反身性风险公司 |
| HC | `munger/references/01-mental-models.md` | 需要逆向思维验证时 |
| EA | `jeff-aronson/references/04-distressed-investing.md` | 困境公司 |

### 缺失 Skill 时的降级

如果某个大师 Skill 未安装：
- 该 Agent 仍可运行，但使用 LLM 通用知识（无大师原始引用）
- 置信度自动 -15（缺少知识库锚定）
- 在 Variable Lock Sheet 中标注：`[degraded: {skill_name} not installed]`

---

## 三档模式

| 模式 | Agent | 时间 | 适用场景 |
|------|-------|------|---------|
| **MAVI-Lite** | CA + GTA + IC + EA（4 个） | +30 分钟 | 快速扫描，HF Track |
| **MAVI-Standard** | CA + GTA + IC + RC + HC（5 个） | +2-3 小时 | 标准分析 |
| **MAVI-Full** | 全部 7 个 | +3-5 小时 | 投委会 / LP / M&A |

---

## 执行流程

### Step 0: 变量重要性筛选

```
Variable Priority = Valuation Leverage × Current Uncertainty
```

| 优先级 | 条件 | 处理方式 |
|--------|------|---------|
| **P0** | 杠杆高 AND 不确定高 | 全部 Agent 评估，敏感性必做 |
| **P1** | 杠杆高 OR 不确定高 | 3-4 个 Agent 评估，敏感性推荐 |
| **P2** | 杠杆低 | 跳过 MAVI，单点估计 |

**输出**：5-8 个 P0 变量 + 3-5 个 P1 变量

### Step 1: 7 Agent 独立评估

**核心规则**：Agent 之间**互不知晓对方的评估结果**（防止锚定偏差）。

每个 Agent 在评估前：
1. 读取对应的投资大师 Skill 的 Reference 文件（见映射表）
2. 用该大师的框架和检查清单分析变量
3. 产出标准格式的评估

---

### Agent 1: Conservative Anchor (CA) — 保守锚点

**动态依赖**：`howard-marks/references/02-risk.md` + `04-contrarian-value.md` + `06-defensive-investing.md`
**辅助**：`buffett/references/06-valuation-capital.md`

**核心问题**：`"这个变量的可信底部在哪？"`

**必须使用的框架**（来自 howard-marks Skill）：
- Marks 风险六维度评估（不是波动率！是永久资本损失的概率）
- 不对称性检查（下行是否被资产/合同/监管保护？）
- "钟摆"周期定位（当前情绪在恐惧端还是贪婪端？）

**必须使用的框架**（来自 buffett Skill）：
- 安全边际计算（范围估值，非点估计）
- 内在价值 vs 市场价格差距

**输出格式**：
```
CA评估 | [变量名]
估值: [X%/X倍]
置信度: [0-100]
方向偏向: ↓ 偏保守
底线红线: 如果低于[Y]，整个命题需要重新审视
安全边际: [当前假设距红线的百分比距离]
推理: [一句话，引用具体框架]
```

---

### Agent 2: Growth Thesis Agent (GTA) — 成长引擎

**动态依赖**：`cathie-wood/references/01-five-platforms.md` + `02-wrights-law.md` + `03-s-curve-adoption.md`
**辅助**：`eric-sanchez-ai/references/01-physical-layer.md`（科技公司时）

**核心问题**：`"强执行下天花板在哪？这个数字在行业内是否被其他公司实现过？"`

**必须使用的框架**（来自 cathie-wood Skill）：
- Wright's Law 成本曲线分析：C(q) = C₁ × q^(-b)
- S 曲线采用阶段判断（pre-chasm / early majority / mainstream）
- TAM 扩张 vs 静态 TAM

**先例证据强制要求**：GTA **必须**引用一家已实现过该增长/利润水平的公司。无先例 → 置信度上限 50。

**输出格式**：
```
GTA评估 | [变量名]
估值: [X%/X倍]
置信度: [0-100]
方向偏向: ↑ 偏乐观
先例证据: [公司名 + 具体数字 + 来源]
Wright's Law适用性: [适用/不适用]
推理: [一句话，必须有可验证依据]
```

---

### Agent 3: Industry Calibrator (IC) — 行业标尺

**动态依赖**：无外部依赖（内置 Comps 方法论）

**核心问题**：`"这家公司在行业分布中站在哪个分位？模型假设是否要求它维持或提升这个排名？"`

**内置框架**：
- Peer Selection 五维框架（业务模型 / 地理 / 规模 / 增长阶段 / 资本结构）
- P25/P50/P75 分位定位
- 行业生命周期阶段（Introduction → Growth → Maturity → Decline）
- Porter 五力作为变量约束器
- "隐含排名" 警报：如果模型假设公司从 P50 升至 P80，必须标注并要求解释

**输出格式**：
```
IC评估 | [变量名]
可比组均值: [X%]
可比组P25-P75区间: [A%] — [B%]
目标公司当前位置: P[XX]
模型假设隐含位置: P[YY]
行业生命周期阶段: [Growth/Maturity/...]
置信度: [0-100]
推理: 隐含位置[合理/偏高/偏低]
```

---

### Agent 4: Risk Challenger (RC) — 风险挑战者

**动态依赖**：`jim-chanos/references/01-forensic-accounting.md` + `02-fraud-patterns.md` + `07-institutional-failure.md`
**辅助**：`george-soros/references/01-reflexivity.md`

**核心问题**：`"哪一个假设如果错了，会让这个模型失去意义？"`

**必须使用的框架**（来自 jim-chanos Skill）：
- 法证会计 8 大红旗（DSO 趋势 / 费用资本化 / 应计质量 / 关联交易 / 审计师变更 / 内部人卖出 / SEC 信函 / SBC 操纵）
- 四大做空桶分类
- 制度失败审计（谁在看门？他们的激励是什么？）

**必须使用的框架**（来自 george-soros Skill）：
- 反身性下行螺旋识别（股价下跌 → 信用评级下调 → 融资成本上升 → 增长放缓 → 进一步下跌）

**输出格式**：
```
RC评估 | [变量名]
压力测试底部: [X%]（此水平下投资逻辑仍成立）
崩溃阈值: [Y%]（低于此点模型结论翻转）
法证红旗: [0-8 个检测到的红旗]
反身性风险: [高/中/低]
制度守卫评估: 审计师[可信/存疑]，分析师[独立/有冲突]
置信度: [0-100]
推理: [触发路径一句话]
```

---

### Agent 5: Historical Calibrator (HC) — 历史校准器

**动态依赖**：`ray-dalio/references/01-economic-machine.md` + `02-debt-cycles.md` + `05-deleveraging.md`
**辅助**：`munger/references/01-mental-models.md`

**核心问题**：`"这个假设与这家公司自己的历史轨道一致吗？偏离历史基础利率的幅度是否有充分解释？"`

**必须使用的框架**（来自 ray-dalio Skill）：
- 经济机器三大力量定位（生产率趋势 + 短期债务周期 + 长期债务周期）
- 均值回归框架（z-score 偏差检测）

**必须使用的框架**（来自 munger Skill）：
- 基础利率思维："这个行业有多少公司曾连续 3 年以上实现过这个增速？"
- 逆向思维："什么会让这个假设失败？"

**"曲棍球棒" 检测**：当历史数据平坦但模型假设加速时，HC 必须标注并要求 GTA 解释结构性变化。

**"这次不同" 的 6 项检验**：
1. 技术不连续性？ 2. 监管环境不可逆变化？ 3. 竞争结构根本改变？
4. 客户基础结构性不同？ 5. 成本结构永久转移？ 6. 黑天鹅使历史不可比？

**输出格式**：
```
HC评估 | [变量名]
历史均值（近5年）: [X%]
模型假设偏离均值: [+/- ZZ%]（[N]个标准差）
曲棍球棒检测: [通过/⚠️ 检测到拐点]
"这次不同"检验: [通过X/6项]
管理层guidance达成率: [XX%]
置信度: [0-100]
推理: [偏离是否有结构性解释]
```

---

### Agent 6: Macro Architect (MA) — 宏观架构师

**动态依赖**：`ray-dalio/references/03-all-weather.md` + `george-soros/references/02-boom-bust-cycle.md`

**核心问题**：`"当前的宏观环境对这个变量假设是顺风还是逆风？"`

**必须使用的框架**：
- Dalio "我们在周期的哪里"（短期债务周期 5-8 年 + 长期债务周期 50-75 年）
- Soros 反身性在宏观层面（央行政策 → 资产价格 → 实体经济反馈）
- 流动性追踪（M2、央行资产负债表、逆回购）
- 利率影响矩阵（每个 P0 变量对 ±100bp 的敏感性）

**MA 仅评估宏观有实质影响的 P0 变量**。

**输出格式**：
```
MA评估 | [变量名]
宏观环境: [顺风/逆风/中性]
短期债务周期位置: [Early Expansion/Mid-Cycle/Late Cycle/Contraction]
利率敏感性: +100bp → [变量变化%]
流动性条件: [宽松/收紧/中性]
地缘政治风险: [具体风险] P=[X%]
置信度: [0-100]
宏观调整建议: Base case 应 [上调/下调/不变] [X%]
```

---

### Agent 7: Event Arbiter (EA) — 事件裁判

**动态依赖**：`john-paulson/references/01-merger-arb-framework.md` + `paul-singer/references/01-activist-playbook.md`
**辅助**：`jeff-aronson/references/04-distressed-investing.md`（困境公司时）

**核心问题**：`"有没有具体的、可定时的事件会改变这个变量的预期？"`

**必须使用的框架**：
- Paulson 7 维度交易分析（反垄断 / 融资确定性 / 股东投票 / 关闭条件 / 违约金 / 战略逻辑 / 搅局者）
- Singer 5 种激进战役类型 + 治理红旗评估
- 催化剂分类：Type A 已排期 / Type B 预期中 / Type C 意外
- 含催化剂时间的年化期望值

**EA 仅在存在具体催化事件时评估**。无催化剂时输出 "No catalyst identified, skip"。

**输出格式**：
```
EA评估 | [变量名]
已识别催化剂: [事件描述]
催化剂类型: [A/B/C]
催化剂日期: [YYYY-MM-DD 或时间窗口]
催化剂概率: [X%]
如触发对变量影响: [变量→Y%]
市场已定价程度: [完全/部分/未]
激进主义潜力: [高/中/低/不适用]
年化期望值: [X%]
置信度: [0-100]
```

---

## Step 2: CMO 聚合

CMO（首席建模官）不取平均值，而是基于信息优势加权：

### 分歧指数 (DI)

```
DI = (max(estimates) - min(estimates)) / median(estimates) × 100%
```

| DI | 级别 | 敏感性要求 |
|----|------|---------|
| > 50% | 🔴 HIGH | 必须做完整敏感性矩阵 |
| 20-50% | 🟡 MED | 推荐敏感性分析 |
| < 20% | 🟢 LOW | 单点估计可接受 |

### CMO 加权原则

| 场景 | 权重倾向 |
|------|---------|
| 数据充分的成熟公司 | IC + HC 权重更高 |
| 早期/成长期公司 | GTA 的先例证据质量决定权重 |
| 高杠杆/压力场景 | RC 的崩溃阈值必须纳入 Bear |
| 宏观不确定性高 | MA 权重提升 |
| 有明确催化剂 | EA 权重提升 |

### Base/Bear/Bull 确定

```
Base = CMO 加权中位数
Bear = max(RC 崩溃阈值, CA 底线红线) → 如 MA 识别逆风再下调
Bull = GTA 先例证据支撑的天花板 → 如 EA 识别正面催化剂可上调
```

---

## Step 3: Variable Lock Sheet 输出

```
━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━
VARIABLE LOCK SHEET v2.0
公司: [Company]  |  模式: [Lite/Standard/Full]
日期: [YYYY-MM-DD]  |  版本: [Draft/Final]
━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━

## Agent 评估矩阵
| 变量 | CA | GTA | IC | RC | HC | MA | EA | DI | 级别 |
|------|-----|------|-----|-----|-----|-----|-----|------|------|

## CMO 锁定结论
| 变量 | Base | Bear | Bull | CMO置信度 | 锁定依据 |
|------|------|------|------|---------|---------|

## 宏观叠加层 (Macro Overlay)
| 维度 | 当前状态 | 对 Base Case 影响 |

## 催化剂日历 (Catalyst Calendar)
| 日期 | 事件 | 类型 | 影响变量 | 概率 | 市场定价 |

## 敏感性优先级清单
P0: [HIGH 分歧变量，必做敏感性]
P1: [MED 分歧变量，推荐]
P2: [LOW 分歧，单点即可]

## CMO 特别注记
[高分歧成因 / 隐性相关性 / 宏观风险 / 催化剂窗口 / 曲棍球棒警告]
━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━
```

---

## Step 4: Standard Analysis Signal 输出

```json
{
  "skill": "mavi",
  "version": "2.0",
  "ticker": "[TICKER]",
  "analysis_date": "YYYY-MM-DD",
  "signal": "strong_buy | buy | hold | sell | strong_sell",
  "confidence": 0.00-1.00,
  "target_allocation_pct": 0.0-15.0,
  "rationale": {
    "primary": "[核心理由]",
    "bull_case": "[Bull + 概率]",
    "bear_case": "[Bear + 概率]",
    "catalyst": "[下一个催化剂 + 日期]"
  },
  "mavi_metadata": {
    "agent_count": 7,
    "mode": "Lite | Standard | Full",
    "max_di": 0.00-1.00,
    "macro_overlay": "tailwind | headwind | neutral",
    "catalyst_count": 2
  }
}
```

### Signal 映射

| CMO 判断 | Signal | 条件 |
|----------|--------|------|
| EV > 12%, Asymmetry > 3x, 催化剂 < 6 月 | `strong_buy` | 高信念 + 高不对称 + 近期催化 |
| EV 5-12%, Asymmetry > 1.5x | `buy` | 正 EV + 合理不对称 |
| EV 0-5%, 或缺乏催化剂 | `hold` | 边际正 EV 或无催化 |
| EV < 0%, 或 RC 崩溃阈值接近 | `sell` | 负 EV 或高风险 |
| EV << 0%, 法证红旗, 反身性下行 | `strong_sell` | 严重问题 |

---

## 质量 Gate

| Gate | 检查项 | 标准 |
|------|-------|------|
| G1 | P0 变量完整评估 | 无空白格（MA/EA 可 n/a） |
| G2 | 置信度已填写 | 全部非空 |
| G3 | DI 已计算 | P0/P1/P2 分类完成 |
| G4 | CMO Lock 三值 | Base ≠ Bear ≠ Bull |
| G5 | GTA 先例证据 | 空则置信度 ≤ 50 |
| G6 | 敏感性区间 | P0 变量步进已确定 |
| G7 | HC 曲棍球棒检测 | 偏离解释已填写 |
| G8 | Signal JSON | 非空 |
| G9 | Macro Overlay（Full） | 周期位置 + 利率敏感性已填 |
| G10 | Catalyst Calendar（Full） | 至少标注"无催化剂"或列出 |

---

## 独立使用产出物

当 MAVI 独立使用（不接建模）时，产出：

1. **Variable Lock Sheet**（可下载 MD/PDF）
2. **每个 Agent 的完整分析记录**（带引用来源）
3. **Macro Overlay 摘要**
4. **Catalyst Calendar**
5. **Standard Analysis Signal JSON**

可选调用 `investment-docs/` 生成格式化报告。

---

**版本**: 2.0.0-standalone
**依赖**: 投资大师 Skill（动态引用，缺失时降级运行）
**接口**: Variable Lock Sheet JSON → financial-modeling / Standard Analysis Signal → investment-backtester
**最后更新**: 2026-04-16
