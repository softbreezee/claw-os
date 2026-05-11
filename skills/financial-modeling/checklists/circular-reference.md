> 本文件是 `full-spectrum-modeling` 的参考资料。
> 核心框架见 `../integrated-modeling-core/SKILL.md`。

# 循环引用管理 (Circular Reference Management)

> **💀 曼巴注记 — 来源**: Project Stratus在每个子公司sheet的第1行设置了"Circ. Switch"开关。这不是偷懒，是处理真实世界中不可避免的循环依赖的专业方法。

## 什么是建模中的循环引用

```
常见循环场景:
  1. 利息计算循环:
     Interest Expense → Net Income → Cash Flow → Debt Balance → Interest Expense

  2. 股利循环:
     Dividend Payment → Cash Balance → Available for Distribution → Dividend Payment

  3. Tax循环 (某些司法管辖区):
     Tax Expense → Net Income → Taxable Income → Tax Expense
```

## 处理方法: Circular Breaker Switch

```
在每个sheet的Row 1, Column A, 放置一个开关:

Cell A1: "Circ. Switch"
Cell B1: 1  (1 = ON, 0 = OFF)

使用方式:
  原始公式 (有循环):
    Interest = Avg_Debt_Balance × Interest_Rate
    其中 Avg_Debt_Balance 依赖 Cash Flow，而 Cash Flow 依赖 Interest

  改写后 (带开关):
    Interest = IF($B$1=1, Avg_Debt_Balance × Interest_Rate, 0)

  运行步骤:
    1. 设置 B1 = 0 (关闭循环)
    2. 输入所有其他假设
    3. 设置 B1 = 1 (开启循环)
    4. Excel自动迭代求解 (需开启: File > Options > Formulas > Enable iterative calculation)
    5. 验证: 迭代后值稳定 (变化<0.01%)
```

## Excel设置要求

```
File > Options > Formulas:
  ☑ Enable iterative calculation
  Maximum Iterations: 100
  Maximum Change: 0.001

注意: 如果迭代不收敛(值持续振荡)，说明模型结构有问题，需要检查循环逻辑。
```

## 何时允许、何时禁止循环

```
✓ 允许 (真实业务逻辑要求):
  • 利息-现金流循环 (LBO/杠杆交易)
  • 股利-可分配利润循环 (分红政策建模)
  • Tax shield循环 (利息抵税→税→现金流→债务→利息)

✗ 禁止 (建模错误导致):
  • A引用B、B引用A但无业务逻辑 (公式错误)
  • 不必要的跨表循环引用 (重构公式可消除)

判断标准: 如果关闭循环开关(B1=0)后，模型逻辑仍然完整只是精度略降，
那就是合理循环。如果关闭后模型崩溃，说明是结构性错误。
```

## 验证检查表

```
☐ 如模型包含循环引用，已设置Circ. Switch
☐ 迭代计算已开启且收敛 (Maximum Change < 0.001)
☐ 关闭循环开关后模型仍可运行(值近似正确)
☐ 循环引用有明确业务逻辑说明(利息/股利/tax)
☐ 无意外循环引用(非设计的circular dependency)
```

## 判断标准总结

**Test**: 如果关闭循环开关(B1=0)后，模型逻辑仍然完整只是精度略降，那就是合理循环。如果关闭后模型崩溃，说明是结构性错误，需要重构公式而非使用循环。
