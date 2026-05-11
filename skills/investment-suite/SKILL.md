---
name: investment-intelligence-suite
type: suite
version: 1.0.0
description: "Investment Intelligence Suite — 投资智能套件。一个编排层 + N 个独立 Skill 的模块化投资分析系统。包含 12 位投资大师 Skill、MAVI 多视角审讯协议、财务建模引擎、数据获取层、文档生成器、回测引擎。每个 Skill 可独立使用，也可通过编排层组合为完整工作流。"
author: Investment Intelligence Suite
---

# Investment Intelligence Suite

**一个编排层 + N 个独立 Skill 的模块化投资分析系统。**

每个 Skill 有独立价值，可以单独安装使用。本文件是编排器——只做路由和胶水，不含领域知识。

---

## 意图识别与路由

| 用户意图 | 关键词示例 | 加载的 Skill | 流程路径 |
|---------|-----------|-------------|---------|
| **多视角讨论** | "帮我分析/讨论/怎么看/多视角" | mavi + 对应大师 Skill + investment-data-layer | 路径 A |
| **纯建模** | "建模/模型/Excel/DCF/LBO" | financial-modeling + investment-data-layer | 路径 B |
| **完整流程** | "完整分析/MAVI建模/全流程" | 全部 | 路径 C |
| **单个大师视角** | "用巴菲特视角/用芒格框架/用Chanos方法" | 对应大师 Skill + investment-data-layer | 路径 D |
| **回测验证** | "回测/历史验证/对比信号" | investment-backtester + 大师 Skill | 路径 E |
| **写报告** | "写 IC Memo/报告/one-pager" | investment-docs | 路径 F |
| **组织诊断** | "组织分析/管理层评估/竞争格局" | org-x-ray + investment-data-layer | 路径 G |
| **获取数据** | "拉财务数据/SEC数据/市场数据" | investment-data-layer | 路径 H |

---

## 路径 A：多视角讨论（不建模）

```
用户想要多个投资大师的视角讨论一家公司，但不需要建模。

1. investment-data-layer → 获取基本面数据
2. mavi → Step 0: 识别 P0 变量
3. mavi → Step 1: 7 Agent 独立评估
   每个 Agent 动态加载对应的投资大师 Skill（见 Agent-大师映射表）
4. mavi → Step 2: CMO 聚合
5. mavi → Step 3: Variable Lock Sheet + Macro Overlay + Catalyst Calendar
6. mavi → Step 4: Standard Analysis Signal

产出物：
  ✅ Variable Lock Sheet（可下载）
  ✅ 每个 Agent 的完整分析记录
  ✅ Macro Overlay + Catalyst Calendar
  ✅ Standard Analysis Signal JSON
  ❌ 不产出 Excel 模型
  ❌ 不产出 Modeling Spec
```

---

## 路径 B：纯建模（不讨论）

```
用户已经知道自己的假设，只需要高质量的建模引擎。

1. investment-data-layer → 获取基本面数据
2. financial-modeling → 8D 诊断（Layer 1）
3. financial-modeling → 用户确认假设（Layer 2，无 MAVI）
4. financial-modeling → 选择轨道（HF / BT / Full）
5. financial-modeling → Modeling Spec（Layer 3）→ Excel（Layer 4）
6. financial-modeling → DCF + Comps 交叉验证

产出物：
  ✅ Excel 模型
  ✅ Comps Table + Football Field
  ✅ Standard Analysis Signal JSON
  ❌ 不产出 Variable Lock Sheet
  ❌ 不产出 Agent 审讯记录
```

---

## 路径 C：完整流程（讨论 + 建模）

```
最完整的流程。MAVI 审讯假设 → 建模 → 估值 → 回测 → 报告。

1. investment-data-layer → 获取数据
2. financial-modeling → 8D 诊断（Layer 1）
3. mavi → MAVI 审讯（Layer 1.5）
   ← 动态加载对应投资大师 Skill
4. mavi → Variable Lock Sheet → financial-modeling Layer 2 参数锁定
5. financial-modeling → Layer 3 Modeling Spec → Layer 4 Excel
6. financial-modeling → DCF + Comps 交叉验证
7. investment-backtester → Signal 回测验证（可选）
8. investment-docs → IC Memo / One-Pager（可选）

产出物：全部
```

---

## 路径 D：单个大师视角

```
用户想用特定投资大师的框架分析公司。

1. investment-data-layer → 获取数据
2. 加载对应大师 Skill（如 howard-marks/）
3. 按该 Skill 的 A/B/C 分发逻辑分析
4. 产出 Standard Analysis Signal

产出物：
  ✅ 该大师框架下的完整分析
  ✅ Standard Analysis Signal JSON
```

---

## 路径 E：回测验证

```
对历史投资决策做事后验证。

1. investment-backtester → 接收 Signal JSON（来自 MAVI 或大师 Skill）
2. 选择 portfolio strategy（concentrated / diversified / hedged / cyclical）
3. 运行时间隔离回测
4. 产出回测报告

产出物：
  ✅ Backtest Report（可由 investment-docs 格式化）
```

---

## 路径 F / G：写报告 / 获取数据

这两个 Skill 通常作为其他路径的辅助被调用，也可独立使用。

---

## MAVI Agent ↔ 投资大师 Skill 映射

这是编排层最核心的内容——声明式地连接 MAVI Agent 和独立大师 Skill。

| Agent | 角色 | 主要 Skill | 读取的 Reference 文件 | 辅助 Skill |
|-------|------|-----------|---------------------|-----------|
| **CA** 保守锚点 | 找可信底部 | `howard-marks/` | 02-risk, 04-contrarian-value, 06-defensive-investing | `buffett/` → 06-valuation-capital |
| **GTA** 成长引擎 | 找天花板 + 先例 | `cathie-wood/` | 01-five-platforms, 02-wrights-law, 03-s-curve-adoption | `eric-sanchez-ai/` → 01-physical-layer, 02-ai-platform-shift |
| **IC** 行业标尺 | 同业分位对标 | `financial-modeling/` | comps.md (内置 Comps 模块) | — |
| **RC** 风险挑战者 | 崩溃阈值 | `jim-chanos/` | 01-forensic-accounting, 02-fraud-patterns, 04-governance-red-flags, 07-institutional-failure | `george-soros/` → 01-reflexivity |
| **HC** 历史校准器 | 均值回归 | `ray-dalio/` | 01-economic-machine, 02-debt-cycles, 05-deleveraging | `munger/` → 01-mental-models |
| **MA** 宏观架构师 | 周期 + 利率 + 地缘 | `ray-dalio/` + `george-soros/` | dalio/03-all-weather, dalio/06-changing-world-order, soros/02-boom-bust-cycle | — |
| **EA** 事件裁判 | 催化剂 + 概率 | `john-paulson/` + `paul-singer/` | paulson/01-merger-arb, paulson/05-antitrust, singer/01-activist-playbook, singer/06-governance-analysis | `jeff-aronson/` → 04-distressed-investing |

### 自定义 Agent 阵容

用户可以替换映射：
- 不想用 Cathie Wood 做 GTA？把主要 Skill 换成 `eric-sanchez-ai/`（纯硅基视角）
- 想加强 MA？额外引用 `george-soros/03-currency-attacks.md`
- 想新增 Agent？安装一个新的大师 Skill，在映射表中添加一行

---

## Skill 间接口契约

### 1. Data Layer → 任何 Skill

```json
{
  "schema": "investment-data-layer-v1",
  "ticker": "AAPL",
  "financials": { "IS": {...}, "BS": {...}, "CF": {...} },
  "market": { "price": 185.5, "market_cap": 2850000, "ev": 2830000 },
  "metadata": { "source": "SEC EDGAR XBRL", "confidence": "HIGH", "as_of": "2026-04-16" }
}
```

### 2. MAVI → Financial Modeling（Variable Lock Sheet）

```json
{
  "schema": "mavi-variable-lock-sheet-v2",
  "company": "AAPL",
  "track": "HF",
  "mavi_mode": "MAVI-Full",
  "variables": [
    {
      "name": "FY+1 Revenue Growth",
      "priority": "P0",
      "base": "8%", "bear": "4%", "bull": "12%",
      "di": "HIGH",
      "cmo_confidence": 72,
      "agents": {
        "CA": {"estimate": "5%", "confidence": 80},
        "GTA": {"estimate": "12%", "confidence": 55},
        "IC": {"estimate": "8%", "confidence": 75},
        "RC": {"collapse_threshold": "2%", "confidence": 70},
        "HC": {"estimate": "7%", "confidence": 78},
        "MA": {"macro_adjustment": "-1%", "confidence": 65},
        "EA": {"catalyst": "Q2 earnings Aug 15", "probability": "70%"}
      }
    }
  ],
  "macro_overlay": { "cycle_position": "Late Cycle", "rate_sensitivity": "HIGH" },
  "catalyst_calendar": [ {"date": "2026-08-15", "event": "Q2 Earnings", "type": "A"} ],
  "signal": { "signal": "buy", "confidence": 0.65, "target_allocation_pct": 2.0 }
}
```

### 3. 任何 Skill → Backtester（Standard Analysis Signal）

```json
{
  "skill": "mavi | howard-marks | financial-modeling | ...",
  "ticker": "AAPL",
  "analysis_date": "2026-04-16",
  "signal": "strong_buy | buy | hold | sell | strong_sell",
  "confidence": 0.00-1.00,
  "target_allocation_pct": 0.0-15.0,
  "rationale": { "primary": "...", "bull_case": "...", "bear_case": "..." }
}
```

### 4. 任何 Skill → Investment Docs

investment-docs 接受任意组合的输入——Variable Lock Sheet、模型数据、原始分析笔记——并产出格式化文档。

---

## Suite 包含的 Skill 清单

### 基础设施 Skill（4 个）

| Skill | 用途 | 可独立使用 |
|-------|------|-----------|
| `mavi/` | 多视角假设审讯协议（7 Agent） | ✅ |
| `financial-modeling/` | 财务建模引擎（8D + L1-L4 + DCF + Comps） | ✅ |
| `investment-data-layer/` | 数据获取（MCP + SEC EDGAR + 市场数据） | ✅ |
| `investment-docs/` | 文档生成（IC Memo / Brief / One-Pager） | ✅ |

### 投资大师 Skill（12 个）

| Skill | 投资大师 | 核心差异化 | MAVI Agent |
|-------|---------|-----------|-----------|
| `buffett/` | Warren Buffett | 内在价值 + 护城河 + 管理层 | CA 辅助 |
| `munger/` | Charlie Munger | 心智模型 + 检查清单 + 逆向 | HC 辅助 |
| `howard-marks/` | Howard Marks | 风险≠波动 + 周期 + 第二层思维 | CA 主要 |
| `ray-dalio/` | Ray Dalio | 经济机器 + 债务周期 + 全天候 | HC 主要, MA 主要 |
| `george-soros/` | George Soros | 反身性 + 繁荣-萧条 + 货币攻击 | RC 辅助, MA 主要 |
| `cathie-wood/` | Cathie Wood | 颠覆式创新 + Wright's Law + S 曲线 | GTA 主要 |
| `jim-chanos/` | Jim Chanos | 法证做空 + 会计欺诈 + 制度失败 | RC 主要 |
| `john-paulson/` | John Paulson | 并购套利 + 事件驱动 + 催化剂 | EA 主要 |
| `paul-singer/` | Paul Singer | 激进主义 + 诉讼 + 治理分析 | EA 主要 |
| `jeff-aronson/` | Jeff Aronson | 跨资本结构 + 困境债 + 支点证券 | EA 辅助 |
| `michael-kim-pe/` | Michael Kim | PE/LBO + 亚洲PE + 行业手册 | — |
| `eric-sanchez-ai/` | Eric Sanchez | AI/半导体 + 硅层 + 地缘政治 | GTA 辅助 |

### 工具 Skill（2 个）

| Skill | 用途 | 可独立使用 |
|-------|------|-----------|
| `investment-backtester/` | 通用回测引擎（4 策略模板） | ✅ |
| `org-x-ray/` | 组织诊断（定性分析 + 竞争格局） | ✅ |

---

## 安装

### 完整安装（整个 Suite）

```bash
cp -r investment-intelligence-suite/ ~/.claude/skills/
```

### 按需安装（只装需要的）

```bash
# 只要 MAVI + 几个大师
cp -r investment-intelligence-suite/mavi/ ~/.claude/skills/
cp -r investment-intelligence-suite/howard-marks/ ~/.claude/skills/
cp -r investment-intelligence-suite/jim-chanos/ ~/.claude/skills/
cp -r investment-intelligence-suite/ray-dalio/ ~/.claude/skills/

# 只要建模
cp -r investment-intelligence-suite/financial-modeling/ ~/.claude/skills/
cp -r investment-intelligence-suite/investment-data-layer/ ~/.claude/skills/
```

---

## 版本信息

**版本**: 1.0.0
**包含**: 18 个独立 Skill + 1 个编排器
**总规模**: ~180 files, ~50,000 lines
**最后更新**: 2026-04-16
