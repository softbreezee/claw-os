# Model Audit Checklist

> 独立审计检查清单。可被 Gate 4 内部调用，也可在审计模块（`integrated-modeling-audit/SKILL.md`）中独立使用。
> 颜色规范定义见 `../config/color-spec.md`。

---

## Formula-Level（所有审计范围通用）

### 公式错误
- [ ] 无 `#REF!` 错误 — 所有跨表/跨单元格引用完整
- [ ] 无 `#VALUE!` 错误 — 数据类型匹配（数字 vs 文本）
- [ ] 无 `#N/A` 错误 — VLOOKUP/INDEX/MATCH 找到匹配项
- [ ] 无 `#DIV/0!` 错误 — 分母非零或已用 IFERROR 保护
- [ ] 无 `#NAME?` 错误 — 函数名拼写正确，无中文感叹号

### 公式质量
- [ ] 无公式内嵌硬编码 — 如 `=A1*1.05` 中的 `1.05` 应为单元格引用
- [ ] 相邻单元格公式模式一致 — 同一行/列的投影公式结构相同
- [ ] SUM/AVERAGE 范围完整 — 未遗漏首行或末行
- [ ] 无被覆盖的公式单元格 — 应为公式的位置未被硬编码值替换
- [ ] 无断裂的跨表链接 — 引用的 Sheet 和单元格均存在

### 循环引用
- [ ] 无意循环引用已消除 — 追踪循环路径并修复
- [ ] 有意循环引用有控制 — Circ. Switch 存在且功能正常
- [ ] 迭代设置正确 — Maximum Iterations ≥ 100, Maximum Change ≤ 0.001

### 数据一致性
- [ ] 单位/量级一致 — 全模型统一（百万/千/个），无混用
- [ ] 百分比格式正确 — 以小数存储（0.15 而非 15），格式为 `0.0%`
- [ ] 无隐藏行/列/表包含覆盖数据 — 隐藏内容已审查

---

## Sheet-Level（sheet 和 model 范围）

### 结构
- [ ] 输入/公式分离 — 假设集中在 Assumptions 表
- [ ] Tab 逻辑顺序 — Assumptions → IS → BS → CF → Valuation → Sensitivity
- [ ] 日期表头一致 — 所有表列头对齐，历史/预测分隔清晰
- [ ] 单位标注 — 表头或 Assumptions 中注明单位（$M, $K, etc.）

### 格式
- [ ] 颜色规范一致 — 蓝字 = 输入，黑字 = 公式，绿字 = 跨表引用
- [ ] 关键假设标记 — 蓝字 + 黄底（`#FFFF99`）
- [ ] 关键输出标记 — 黑粗 + 灰底（`#F2F2F2`）
- [ ] 标题行格式 — 白字 + 深蓝底（`#003366`）
- [ ] 数字格式统一 — 千分位、小数位按 `config/number-formats.md` 执行

---

## Model-Level（仅完整模型审计）

### Balance Sheet
- [ ] BS 平衡 — `Total Assets = Total Liabilities + Equity`（**每个期间**）
- [ ] BS 差额量化 — 如不平衡，列出每个期间的差额和断裂点
- [ ] RE 滚动正确 — `Current RE = Prior RE + Net Income - Dividends`
- [ ] PP&E 滚动正确 — `Ending PP&E = Beginning + CapEx - D&A`

### Cash Flow Statement
- [ ] 现金勾稽 — `CF Ending Cash = BS Cash`（每个期间）
- [ ] CF 合计 — `CFO + CFI + CFF = ΔCash`
- [ ] D&A 匹配 — CF 加回 D&A = IS D&A 费用
- [ ] CapEx 匹配 — CF CapEx 与 PP&E 滚动一致
- [ ] WC 变动符号 — ΔAR 增加 → CF 为负；ΔAP 增加 → CF 为正

### Income Statement
- [ ] 收入构成一致 — Revenue Total = Sum of Segments
- [ ] Tax 用 ETR 假设 — `Tax = EBT × ETR_assumption`，非固定金额
- [ ] 股份数来源 — Diluted Shares 与稀释时间表一致

### 三表联动
- [ ] IS → BS 联动 — Net Income 流入 Retained Earnings
- [ ] BS → CF 联动 — WC 变动从 BS 推导
- [ ] CF → BS 联动 — Ending Cash 回写 BS
- [ ] 联动验证 — 修改 Revenue → IS → BS → CF 全链条自动更新

### 逻辑合理性
- [ ] 增长率合理 — Revenue YoY < 100%（除非有并购）
- [ ] Margin 合理 — EBITDA Margin 在行业标准范围内
- [ ] 无 hockey stick — 预测期增速无不合理跳升
- [ ] 边界案例 — 0% 增长、负 EBITDA 时模型无错误

---

## IM Enhancement（Full-Spectrum Modeling 独有）

### 估值检查
- [ ] TGR < WACC - 2% — Terminal Growth Rate 在安全距离内
- [ ] TV/EV < 70% — Terminal Value 占比合理（>70% 黄灯，>80% 红灯）
- [ ] 净现金口径正确 — `Net Cash = Cash + ST Inv - ALL Interest-bearing Debt`
- [ ] Comps vs DCF 偏差 < 30% — 两种估值方法交叉验证偏差合理
- [ ] IRR 在 12-20% 范围 — 超出范围需有文字解释

### 颗粒度规范
- [ ] 30% 拆分规则 — 单条产品线 > 30% 总收入需拆分
- [ ] COGS 产品/服务分行 — 禁止混合 COGS
- [ ] R&D 与 SG&A 分行 — 禁止合并
- [ ] Interest Income/Expense 分行 — 禁止净数呈现

### 质量标准
- [ ] 场景独立性 — Bear/Base/Bull 互不干扰
- [ ] 数据来源注解 — 关键假设 Cell Comment 标注 source + date
- [ ] 颜色规范完整 — 全套颜色编码通过抽样验证
- [ ] Gate 状态 — Gate 1-4 全部通过（如适用）

---

## 审计结果评级

| 评级 | 条件 |
|------|------|
| **Clean ✅** | 0 Critical, ≤ 2 Warning |
| **Minor Issues ⚠️** | 0 Critical, > 2 Warning |
| **Major Issues 🔴** | ≥ 1 Critical |

---

**版本**: 4.0.0 | **最后更新**: 2026-04-12
