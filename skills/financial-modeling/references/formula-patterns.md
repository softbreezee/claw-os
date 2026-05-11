> 本文件是 `full-spectrum-modeling` 的参考资料。
> Core 建模原则见 `../integrated-modeling-core/SKILL.md` §9。
> 颜色编码实现见 `excel-engineering-guide.md` §3。

# 金融公式标准写法库 (Formula Patterns Library)

常用金融建模公式的标准 openpyxl 写法。所有公式必须作为 Excel 公式写入，禁止 Python 预计算后写值。

---

## 1. 收入预测 (Revenue Projections)

```python
# ━━━ 收入增长（Year-over-Year Growth） ━━━
# 上一年收入 × (1 + 增长率)
ws[f"D{rev_row}"] = f"=C{rev_row}*(1+D{growth_row})"

# ━━━ 分部收入合计（Revenue by Segment） ━━━
ws[f"D{total_rev_row}"] = f"=SUM(D{seg_start}:D{seg_end})"

# ━━━ SaaS 收入拆分 ━━━
# ARR = 上年末 ARR + 新增 ARR + 扩展 - 流失
ws[f"D{arr_row}"] = f"=C{arr_end_row}+D{new_arr_row}+D{expansion_row}-D{churn_row}"
# 订阅收入 = ARR（年初 + 年末）/ 2（假设线性增长）
ws[f"D{sub_rev_row}"] = f"=(C{arr_end_row}+D{arr_row})/2"

# ━━━ 平台/Marketplace 收入拆分 ━━━
# GMV = 活跃买家 × 购买频次 × 平均订单值
ws[f"D{gmv_row}"] = f"=D{buyers_row}*D{freq_row}*D{aov_row}"
# 平台收入 = GMV × Take Rate
ws[f"D{platform_rev_row}"] = f"=D{gmv_row}*D{take_rate_row}"

# ━━━ 硬件收入拆分 ━━━
# 出货量 × ASP
ws[f"D{hw_rev_row}"] = f"=D{units_row}*D{asp_row}"
```

---

## 2. 利润表 (Income Statement)

```python
# ━━━ 毛利润 (Gross Profit) ━━━
ws[f"D{gp_row}"] = f"=D{rev_row}-D{cogs_row}"

# ━━━ 毛利率 (Gross Margin) ━━━
ws[f"D{gm_row}"] = f"=D{gp_row}/D{rev_row}"

# ━━━ COGS 拆分（必须分产品和服务） ━━━
ws[f"D{cogs_product_row}"] = f"=D{rev_product_row}*D{cogs_pct_product_row}"
ws[f"D{cogs_service_row}"] = f"=D{rev_service_row}*D{cogs_pct_service_row}"
ws[f"D{cogs_total_row}"] = f"=D{cogs_product_row}+D{cogs_service_row}"

# ━━━ EBITDA ━━━
# EBITDA = Gross Profit - R&D - SG&A（R&D 和 SG&A 必须分行）
ws[f"D{ebitda_row}"] = f"=D{gp_row}-D{rd_row}-D{sga_row}"

# ━━━ EBITDA Margin ━━━
ws[f"D{ebitda_margin_row}"] = f"=D{ebitda_row}/D{rev_row}"

# ━━━ EBIT ━━━
ws[f"D{ebit_row}"] = f"=D{ebitda_row}-D{da_row}"

# ━━━ EBT (Earnings Before Tax) ━━━
# Interest Income 和 Interest Expense 必须分行（禁止净数呈现）
ws[f"D{ebt_row}"] = f"=D{ebit_row}+D{int_income_row}-D{int_expense_row}"

# ━━━ Tax（必须用 ETR 假设，禁止固定金额） ━━━
ws[f"D{tax_row}"] = f"=D{ebt_row}*Assumptions!$B${etr_row}"
# 注意：ETR 用绝对引用 $B$，使所有预测年使用同一税率假设

# ━━━ Net Income ━━━
ws[f"D{ni_row}"] = f"=D{ebt_row}-D{tax_row}"

# ━━━ EPS (Earnings Per Share) ━━━
ws[f"D{eps_row}"] = f"=D{ni_row}/D{diluted_shares_row}"
```

---

## 3. 自由现金流 (Free Cash Flow)

```python
# ━━━ NOPAT (Net Operating Profit After Tax) ━━━
ws[f"D{nopat_row}"] = f"=D{ebit_row}*(1-Assumptions!$B${etr_row})"

# ━━━ Unlevered Free Cash Flow (UFCF) ━━━
# UFCF = NOPAT + D&A - CapEx - ΔNWC
ws[f"D{ufcf_row}"] = f"=D{nopat_row}+D{da_row}-D{capex_row}-D{dnwc_row}"

# ━━━ ΔNet Working Capital (NWC 变动) ━━━
# ΔNWC = Current Period NWC - Prior Period NWC
# NWC = Current Assets (ex-Cash) - Current Liabilities (ex-Debt)
ws[f"D{nwc_row}"] = f"=D{ca_ex_cash_row}-D{cl_ex_debt_row}"
ws[f"D{dnwc_row}"] = f"=D{nwc_row}-C{nwc_row}"

# ━━━ CapEx ━━━
# 方法 A: 占收入比
ws[f"D{capex_row}"] = f"=D{rev_row}*Assumptions!$B${capex_pct_row}"
# 方法 B: 维护性 + 增长性
ws[f"D{capex_row}"] = f"=D{maint_capex_row}+D{growth_capex_row}"

# ━━━ Levered Free Cash Flow (LFCF) ━━━
ws[f"D{lfcf_row}"] = f"=D{ufcf_row}-D{int_expense_row}*(1-Assumptions!$B${etr_row})+D{net_borrowing_row}"
```

---

## 4. WACC (加权平均资本成本)

```python
# ━━━ Cost of Equity (CAPM) ━━━
# Ke = Rf + β × ERP (+ Size Premium if applicable)
ws[f"B{coe_row}"] = f"=B{rf_row}+B{beta_row}*B{erp_row}"
# 含规模溢价版本
ws[f"B{coe_row}"] = f"=B{rf_row}+B{beta_row}*B{erp_row}+B{size_premium_row}"

# ━━━ After-tax Cost of Debt ━━━
# Kd_at = Kd_pretax × (1 - Tax Rate)
ws[f"B{cod_row}"] = f"=B{pretax_cod_row}*(1-B{tax_row})"

# ━━━ WACC ━━━
# WACC = Ke × We + Kd_at × Wd
ws[f"B{wacc_row}"] = f"=B{coe_row}*B{eq_weight_row}+B{cod_row}*B{debt_weight_row}"

# ━━━ 验证（WACC 合理性检查） ━━━
# 通常 6% - 14%，超出此范围需要解释
ws[f"B{wacc_check_row}"] = f'=IF(OR(B{wacc_row}<0.06,B{wacc_row}>0.14),"⚠ REVIEW","✓ OK")'
```

---

## 5. 终值 (Terminal Value)

```python
# ━━━ Gordon Growth Model (永续增长法) ━━━
# TV = FCF_terminal × (1 + TGR) / (WACC - TGR)
ws[f"B{tv_gordon_row}"] = f"={last_col}{ufcf_row}*(1+B{tgr_row})/(B{wacc_row}-B{tgr_row})"

# ━━━ Exit Multiple Method (退出倍数法) ━━━
# TV = Terminal EBITDA × Exit Multiple
ws[f"B{tv_exit_row}"] = f"={last_col}{ebitda_row}*B{exit_mult_row}"

# ━━━ 终值选择（加权平均） ━━━
# 可选：两种方法的加权平均
ws[f"B{tv_blended_row}"] = f"=B{tv_gordon_row}*B{gordon_weight_row}+B{tv_exit_row}*B{exit_weight_row}"

# ━━━ TGR 安全检查（Core 陷阱 #1） ━━━
# TGR 不得超过 WACC - 2%
ws[f"B{tgr_check_row}"] = f'=IF(B{tgr_row}>B{wacc_row}-0.02,"🔴 TGR too high!","✓ OK")'
```

---

## 6. 折现 (Discounting)

```python
# ━━━ 年中折现因子 (Mid-year Convention) ━━━
# DF = 1 / (1 + WACC) ^ (year - 0.5)
ws[f"D{df_row}"] = f"=1/(1+$B${wacc_row})^(D{year_num_row}-0.5)"

# ━━━ 年末折现因子 (Year-end Convention) ━━━
ws[f"D{df_row}"] = f"=1/(1+$B${wacc_row})^D{year_num_row}"

# ━━━ 各年 FCF 现值 (PV of FCF) ━━━
ws[f"D{pv_fcf_row}"] = f"=D{ufcf_row}*D{df_row}"

# ━━━ 终值现值 (PV of Terminal Value) ━━━
# 终值折到最后一个预测年末
ws[f"B{pv_tv_row}"] = f"=B{tv_row}/(1+$B${wacc_row})^{last_year}"
# 年中惯例版本
ws[f"B{pv_tv_row}"] = f"=B{tv_row}/(1+$B${wacc_row})^({last_year}-0.5)"

# ━━━ FCF 现值合计 (Sum of PV of FCFs) ━━━
ws[f"B{sum_pv_row}"] = f"=SUM(D{pv_fcf_row}:{last_col}{pv_fcf_row})"
```

---

## 7. 估值桥 (Valuation Bridge)

```python
# ━━━ Enterprise Value ━━━
# EV = PV of FCFs + PV of Terminal Value
ws[f"B{ev_row}"] = f"=B{sum_pv_row}+B{pv_tv_row}"

# ━━━ Equity Value ━━━
# Equity Value = EV + Net Cash (or - Net Debt)
# Net Cash = Cash + ST Investments - ALL Interest-bearing Debt
ws[f"B{net_cash_row}"] = f"=B{cash_row}+B{st_inv_row}-B{total_debt_row}"
ws[f"B{equity_row}"] = f"=B{ev_row}+B{net_cash_row}"

# ━━━ Implied Share Price ━━━
ws[f"B{price_row}"] = f"=B{equity_row}/B{shares_row}"

# ━━━ TV/EV 占比检查（Core 陷阱 #1） ━━━
ws[f"B{tv_ev_pct_row}"] = f"=B{pv_tv_row}/B{ev_row}"
ws[f"B{tv_ev_check_row}"] = f'=IF(B{tv_ev_pct_row}>0.8,"🔴 TV>80%",IF(B{tv_ev_pct_row}>0.7,"⚠️ TV>70%","✓ OK"))'

# ━━━ Upside/Downside ━━━
ws[f"B{upside_row}"] = f"=B{price_row}/B{current_price_row}-1"
```

---

## 8. 场景切换 (Scenario Switch)

```python
# ━━━ 场景选择器 (Scenario Selector) ━━━
# Assumptions!$B$6 = 1(Bear), 2(Base), 3(Bull)
# 每个预测单元格引用对应场景 sheet

# 方法 A: 嵌套 IF（适合3个场景）
ws[f"D{row}"] = f'=IF($B$6=1,Bear!D{row},IF($B$6=2,Base!D{row},Bull!D{row}))'

# 方法 B: CHOOSE（更简洁）
ws[f"D{row}"] = f'=CHOOSE($B$6,Bear!D{row},Base!D{row},Bull!D{row})'

# 方法 C: INDEX + INDIRECT（适合多场景）
# 场景名在 A1:A5，使用 INDIRECT 动态引用
ws[f"D{row}"] = f"=INDIRECT(INDEX($A$1:$A$5,$B$6)&\"!D{row}\")"

# ━━━ 场景独立性验证 ━━━
# 切换场景后，非活跃场景的值不应改变
# Gate 4 要求: Bear/Base/Bull 互不干扰
```

---

## 9. 敏感性单元格 (Sensitivity Formulas)

```python
# ━━━ 2D 敏感性矩阵（WACC vs TGR） ━━━
# 每个单元格必须是完整 DCF 重算，不是线性插值
#
# 矩阵结构：
#          TGR_1   TGR_2   TGR_3   TGR_4   TGR_5
# WACC_1   [价格]  [价格]  [价格]  [价格]  [价格]
# WACC_2   [价格]  [价格]  [价格]  [价格]  [价格]
# WACC_3   [价格]  [价格]  [BASE]  [价格]  [价格]  ← 中心
# WACC_4   [价格]  [价格]  [价格]  [价格]  [价格]
# WACC_5   [价格]  [价格]  [价格]  [价格]  [价格]

# 每个单元格公式模板:
# = (Terminal_FCF * (1+TGR) / (WACC-TGR) / (1+WACC)^n + Sum_PV_FCFs + Net_Cash) / Shares
for i in range(5):
    for j in range(5):
        row = matrix_start_row + 1 + i
        col_letter = get_column_letter(matrix_start_col + 1 + j)
        tgr_ref = f"{col_letter}${matrix_start_row}"       # 列头 TGR
        wacc_ref = f"${get_column_letter(matrix_start_col)}${row}"  # 行头 WACC
        
        formula = (
            f"=(DCF!{last_col}{ufcf_row}*(1+{tgr_ref})"
            f"/({wacc_ref}-{tgr_ref})"
            f"/(1+{wacc_ref})^{projection_years}"
            f"+DCF!$B${sum_pv_row}"
            f"+DCF!$B${net_cash_row})"
            f"/DCF!$B${shares_row}"
        )
        ws.cell(row=row, column=matrix_start_col + 1 + j).value = formula
```

---

## 10. Balance Sheet 联动 (BS Linkage)

```python
# ━━━ Retained Earnings 滚动 ━━━
ws[f"D{re_row}"] = f"=C{re_row}+D{ni_row}-D{div_row}"

# ━━━ BS 平衡检查 ━━━
ws[f"D{bs_check_row}"] = f"=D{total_assets_row}-D{total_le_row}"
# 应为 0；非零表示模型断裂

# ━━━ PP&E 滚动 ━━━
ws[f"D{ppe_row}"] = f"=C{ppe_row}+D{capex_row}-D{da_row}"

# ━━━ Cash 勾稽（从 CF 推导） ━━━
ws[f"D{cash_row}"] = f"=C{cash_row}+D{cfo_row}+D{cfi_row}+D{cff_row}"
```

---

## 11. LBO 专用 (LBO-Specific)

```python
# ━━━ 债务偿还 (Debt Paydown) ━━━
# 强制还款 = MIN(可用现金, 本期应还)
ws[f"D{paydown_row}"] = f"=MIN(D{available_cash_row},D{mandatory_repay_row})"

# ━━━ 债务余额滚动 (Debt Balance Roll) ━━━
ws[f"D{debt_bal_row}"] = f"=C{debt_bal_row}-D{paydown_row}"

# ━━━ 利息计算（用平均余额） ━━━
ws[f"D{interest_row}"] = f"=(C{debt_bal_row}+D{debt_bal_row})/2*Assumptions!$B${rate_row}"

# ━━━ IRR ━━━
# 假设第3列为Entry，最后列为Exit
ws[f"B{irr_row}"] = f"=IRR(C{cf_row}:{last_col}{cf_row})"

# ━━━ MoIC ━━━
ws[f"B{moic_row}"] = f"={last_col}{exit_equity_row}/C{entry_equity_row}"
```

---

## 12. Comps 专用 (Comps-Specific)

```python
# ━━━ EV/EBITDA ━━━
ws[f"D{ev_ebitda_row}"] = f"=D{ev_row}/D{ebitda_row}"

# ━━━ EV/Revenue ━━━
ws[f"D{ev_rev_row}"] = f"=D{ev_row}/D{rev_row}"

# ━━━ P/E ━━━
ws[f"D{pe_row}"] = f"=D{price_row}/D{eps_row}"

# ━━━ Comps 统计 ━━━
ws[f"B{median_row}"] = f"=MEDIAN(D{comp_start}:D{comp_end})"
ws[f"B{mean_row}"] = f"=AVERAGE(D{comp_start}:D{comp_end})"
ws[f"B{p25_row}"] = f"=PERCENTILE(D{comp_start}:D{comp_end},0.25)"
ws[f"B{p75_row}"] = f"=PERCENTILE(D{comp_start}:D{comp_end},0.75)"

# ━━━ Implied Value from Comps ━━━
ws[f"B{implied_ev_row}"] = f"=B{median_row}*B{target_ebitda_row}"
ws[f"B{implied_equity_row}"] = f"=B{implied_ev_row}+B{net_cash_row}"
ws[f"B{implied_price_row}"] = f"=B{implied_equity_row}/B{shares_row}"
```

---

**版本**: 4.0.0 | **最后更新**: 2026-04-12 | **相关文件**: `excel-engineering-guide.md`, `recalc-guide.md`
