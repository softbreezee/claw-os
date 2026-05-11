> 本文件是 `full-spectrum-modeling` 的参考资料。
> Core 建模原则见 `../integrated-modeling-core/SKILL.md` §9。

# Excel 工程化执行指南 (Excel Engineering Guide)

AI 稳定产出高质量 Excel 财务模型的核心参考文件。覆盖公式写法、颜色编码、敏感性矩阵、审计追踪和常见陷阱。

---

## 1. 执行环境检测 (Runtime Environment Detection)

在写入任何单元格之前，必须先判断执行环境：

```
┌─────────────────────────────────────────────────────────────────┐
│ IF running inside Excel (Office JS Add-in / Office JS env):    │
│   → Use Office JS API directly                                 │
│     range.formulas = [["=D19*(1+$B$8)"]]                      │
│   → No recalc step needed — Excel calculates natively          │
│   → Use range.format.* for styling (font, fill, borders)      │
│   → Merged cells: 先写值到左上角单元格，再 merge 整个范围       │
│   → Use context.workbook.worksheets for cross-sheet ops        │
│   → Batch operations inside context.sync() blocks             │
│                                                                 │
│ ELSE (standalone .xlsx generation, no live Excel session):      │
│   → Use Python openpyxl                                        │
│   → 必须运行 recalc 步骤（见 recalc-guide.md）                 │
│   → 推荐 XML 双写方案：公式保留 + 缓存值注入                    │
│   → openpyxl.Comment for audit trail                           │
│   → openpyxl.styles for color coding                           │
│   → 文件保存后验证：打开确认公式存在 + 值已显示                   │
└─────────────────────────────────────────────────────────────────┘
```

**环境检测代码（Python）**:
```python
import shutil
import subprocess

def detect_environment():
    """检测当前执行环境，返回推荐的 recalc 方案"""
    # Office JS 环境（通常由 Add-in 框架注入，Python 不会在此环境中运行）
    # 如果在 Python 中执行，一定是 standalone 模式
    
    # 检测 LibreOffice
    lo_path = shutil.which("libreoffice") or shutil.which("soffice")
    if lo_path:
        return "libreoffice", lo_path
    
    # 检测 formulas 库
    try:
        import formulas
        return "xml_dual_write", None  # 推荐方案 C
    except ImportError:
        return "openpyxl_only", None   # 仅保留公式，无缓存值
```

**环境检测代码（Office JS / TypeScript）**:
```typescript
async function detectOfficeJS(): Promise<boolean> {
    try {
        await Excel.run(async (context) => {
            context.workbook.load("name");
            await context.sync();
        });
        return true; // Office JS 环境，无需 recalc
    } catch {
        return false;
    }
}
```

---

## 2. 公式写法铁律 (Formula Writing Rules)

### 核心原则

**投影/计算单元格必须是 Excel 公式，禁止 Python 预计算后写入数值。**

违反此原则的模型无法通过 Gate 4 审查。

### 正确 vs 错误示例

| # | 场景 | ✅ 正确（Excel 公式） | ❌ 错误（Python 预计算值） |
|---|------|----------------------|--------------------------|
| 1 | 收入增长 | `ws["D10"] = "=C10*(1+D5)"` | `ws["D10"] = prev_rev * (1 + growth)` |
| 2 | EBITDA Margin | `ws["D20"] = "=D18/D10"` | `ws["D20"] = ebitda / revenue` |
| 3 | UFCF 计算 | `ws["D30"] = "=D25+D26-D27-D28"` | `ws["D30"] = nopat + da - capex - dnwc` |
| 4 | Tax 计算 | `ws["D16"] = "=D15*Assumptions!$B$8"` | `ws["D16"] = ebt * tax_rate` |
| 5 | 敏感性单元格 | `ws["E42"] = "=NPV(E$40,DCF!$D$30:$H$30)+..."` | `ws["E42"] = npv_result` |
| 6 | 场景切换 IF | `ws["D10"] = '=IF($B$6=1,Bear!D10,IF($B$6=2,Base!D10,Bull!D10))'` | `ws["D10"] = scenario_values[scenario][row]` |

### 唯一允许硬编码的单元格类型

| 类型 | 示例 | 原因 |
|------|------|------|
| 原始历史输入 | 2023 Revenue = 1,250.3 | 无法从其他单元格推导 |
| 假设驱动值 | Revenue Growth Rate = 12% | 用户自定义假设，需蓝色字体标记 |
| 当前市场数据 | Risk-free Rate = 4.25% | 外部数据，需 cell comment 标注来源 |
| 标签和文本 | "Revenue", "FY2024E" | 非数值内容 |

### 验证方法

完成模型后执行以下测试：
1. 修改一个关键假设（如 revenue growth 从 12% 改为 15%）
2. 验证所有下游单元格自动更新
3. 确认无 `#REF!`、`#VALUE!`、`#DIV/0!`、`#NAME?` 错误
4. 检查结果变化方向和幅度合理

```python
def validate_formula_integrity(ws, assumption_cell, test_value, original_value):
    """验证公式链完整性"""
    # 记录所有单元格当前值
    before = {cell.coordinate: cell.value for row in ws.iter_rows() for cell in row}
    
    # 修改假设
    ws[assumption_cell] = test_value
    
    # recalc（见 recalc-guide.md）
    # ...
    
    # 验证下游变化
    after = {cell.coordinate: cell.value for row in ws.iter_rows() for cell in row}
    changed = {k for k in before if before[k] != after[k]}
    
    # 恢复
    ws[assumption_cell] = original_value
    
    # 检查无错误
    errors = [c for c in after if isinstance(after[c], str) and after[c].startswith("#")]
    assert len(errors) == 0, f"公式错误: {errors}"
    assert len(changed) > 1, f"下游未更新，可能存在硬编码"
```

---

## 3. 颜色编码的 openpyxl 实现 (Color Coding Implementation)

完整实现 `config/color-spec.md` 定义的所有颜色规范。

### 完整代码块

```python
from openpyxl.styles import Font, PatternFill, Alignment, Border, Side

# ━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━
# 标准颜色定义 (Investment Banking Standard)
# ━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━

# --- 字体 (Fonts) ---
FONT_INPUT = Font(name="Calibri", size=10, color="0000FF")               # 硬编码输入：蓝字
FONT_KEY_ASSUMPTION = Font(name="Calibri", size=10, color="0000FF")      # 关键假设：蓝字（配合黄底）
FONT_FORMULA = Font(name="Calibri", size=10, color="000000")             # 公式：黑字
FONT_CROSSREF = Font(name="Calibri", size=10, color="008000")            # 跨表引用：绿字
FONT_KEY_OUTPUT = Font(name="Calibri", size=10, color="000000", bold=True)  # 关键输出：黑粗
FONT_BEAR = Font(name="Calibri", size=10, color="FF0000")                # 熊市场景：红字
FONT_HEADER = Font(name="Calibri", size=10, color="FFFFFF", bold=True)   # 标题行：白字粗体
FONT_SUBTOTAL = Font(name="Calibri", size=10, color="000000", bold=True) # 分类汇总：黑粗

# --- 填充 (Fills) ---
FILL_NONE = PatternFill(fill_type=None)                                   # 无背景
FILL_KEY_ASSUMPTION = PatternFill(start_color="FFFF99", end_color="FFFF99", fill_type="solid")  # 关键假设：黄底
FILL_KEY_OUTPUT = PatternFill(start_color="F2F2F2", end_color="F2F2F2", fill_type="solid")      # 关键输出：灰底
FILL_HEADER = PatternFill(start_color="003366", end_color="003366", fill_type="solid")          # 标题行：深蓝底
FILL_SUBTOTAL = PatternFill(start_color="FFFFCC", end_color="FFFFCC", fill_type="solid")        # 分类汇总：浅黄底
FILL_SENSITIVITY_CENTER = PatternFill(start_color="BDD7EE", end_color="BDD7EE", fill_type="solid")  # 敏感性中心：浅蓝底

# --- 对齐 (Alignment) ---
ALIGN_CENTER = Alignment(horizontal="center", vertical="center")
ALIGN_RIGHT = Alignment(horizontal="right", vertical="center")
ALIGN_LEFT = Alignment(horizontal="left", vertical="center")

# --- 边框 (Borders) ---
BORDER_THIN = Border(
    left=Side(style="thin"), right=Side(style="thin"),
    top=Side(style="thin"), bottom=Side(style="thin")
)
BORDER_BOTTOM_DOUBLE = Border(bottom=Side(style="double"))  # 汇总行底部双线

# ━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━
# 应用函数 (Application Functions)
# ━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━

def style_input(ws, cell_ref, value):
    """硬编码输入：蓝字，无背景"""
    cell = ws[cell_ref]
    cell.value = value
    cell.font = FONT_INPUT
    cell.fill = FILL_NONE

def style_key_assumption(ws, cell_ref, value):
    """关键假设：蓝字 + 黄底"""
    cell = ws[cell_ref]
    cell.value = value
    cell.font = FONT_KEY_ASSUMPTION
    cell.fill = FILL_KEY_ASSUMPTION

def style_formula(ws, cell_ref, formula):
    """公式：黑字，无背景"""
    cell = ws[cell_ref]
    cell.value = formula  # 必须以 = 开头
    cell.font = FONT_FORMULA
    cell.fill = FILL_NONE

def style_crossref(ws, cell_ref, formula):
    """跨表引用公式：绿字"""
    cell = ws[cell_ref]
    cell.value = formula
    cell.font = FONT_CROSSREF
    cell.fill = FILL_NONE

def style_key_output(ws, cell_ref, formula):
    """关键输出：黑粗 + 灰底"""
    cell = ws[cell_ref]
    cell.value = formula
    cell.font = FONT_KEY_OUTPUT
    cell.fill = FILL_KEY_OUTPUT

def style_bear(ws, cell_ref, formula_or_value):
    """熊市场景：红字"""
    cell = ws[cell_ref]
    cell.value = formula_or_value
    cell.font = FONT_BEAR
    cell.fill = FILL_NONE

def style_header_row(ws, row_num, col_start, col_end):
    """标题行：白字 + 深蓝底"""
    for col in range(col_start, col_end + 1):
        cell = ws.cell(row=row_num, column=col)
        cell.font = FONT_HEADER
        cell.fill = FILL_HEADER
        cell.alignment = ALIGN_CENTER

def style_subtotal(ws, cell_ref, formula):
    """分类汇总：黑粗 + 浅黄底 + 底部双线"""
    cell = ws[cell_ref]
    cell.value = formula
    cell.font = FONT_SUBTOTAL
    cell.fill = FILL_SUBTOTAL
    cell.border = BORDER_BOTTOM_DOUBLE
```

### Office JS 颜色实现

```typescript
// Office JS 等效实现
async function styleInput(range: Excel.Range): Promise<void> {
    range.format.font.color = "#0000FF";
    range.format.fill.clear();
}

async function styleKeyAssumption(range: Excel.Range): Promise<void> {
    range.format.font.color = "#0000FF";
    range.format.fill.color = "#FFFF99";
}

async function styleHeader(range: Excel.Range): Promise<void> {
    range.format.font.color = "#FFFFFF";
    range.format.font.bold = true;
    range.format.fill.color = "#003366";
    range.format.horizontalAlignment = Excel.HorizontalAlignment.center;
}
```

---

## 4. 敏感性矩阵的程序化生成 (Sensitivity Matrix Generation)

### 规则

- 必须用**奇数**行列（5×5 或 7×7），确保中心单元格 = base case
- 中心单元格高亮（`#BDD7EE` 浅蓝底 + bold）
- **每个单元格必须是完整 DCF 重算公式**，不是线性插值
- 行变量和列变量必须覆盖合理区间

### Python openpyxl 模板（5×5 WACC vs TGR）

```python
from openpyxl.utils import get_column_letter

def build_sensitivity_matrix(ws, start_row, start_col,
                              wacc_base, tgr_base,
                              wacc_step=0.005, tgr_step=0.005,
                              dcf_sheet_name="DCF",
                              terminal_fcf_cell="H30",
                              sum_pv_fcf_cell="B35",
                              net_cash_cell="B38",
                              shares_cell="B40",
                              size=5):
    """
    生成 size x size 敏感性矩阵。
    每个单元格是完整的隐含股价重算公式。
    """
    half = size // 2
    
    # 写标题
    ws.cell(row=start_row, column=start_col).value = "WACC \\ TGR"
    ws.cell(row=start_row, column=start_col).font = FONT_HEADER
    ws.cell(row=start_row, column=start_col).fill = FILL_HEADER
    
    # 列头：TGR 值
    for j in range(size):
        tgr_val = tgr_base + (j - half) * tgr_step
        col = start_col + 1 + j
        cell = ws.cell(row=start_row, column=col)
        cell.value = tgr_val
        cell.number_format = "0.0%"
        cell.font = FONT_HEADER
        cell.fill = FILL_HEADER
        cell.alignment = ALIGN_CENTER
    
    # 行头：WACC 值 + 矩阵单元格
    for i in range(size):
        wacc_val = wacc_base + (i - half) * wacc_step
        row = start_row + 1 + i
        
        # WACC 标签
        label_cell = ws.cell(row=row, column=start_col)
        label_cell.value = wacc_val
        label_cell.number_format = "0.0%"
        label_cell.font = Font(name="Calibri", size=10, color="FFFFFF", bold=True)
        label_cell.fill = FILL_HEADER
        
        for j in range(size):
            tgr_val = tgr_base + (j - half) * tgr_step
            col = start_col + 1 + j
            cell = ws.cell(row=row, column=col)
            
            # 获取 TGR 和 WACC 单元格引用（从矩阵表头读取）
            tgr_ref = f"{get_column_letter(col)}${start_row}"
            wacc_ref = f"${get_column_letter(start_col)}${row}"
            
            # 完整 DCF 重算公式（非线性插值）：
            # = (TV / (1+WACC)^n + PV_of_FCFs + Net_Cash) / Shares
            # TV = Terminal_FCF * (1 + TGR) / (WACC - TGR)
            formula = (
                f"=({dcf_sheet_name}!{terminal_fcf_cell}*(1+{tgr_ref})"
                f"/({wacc_ref}-{tgr_ref})"
                f"/(1+{wacc_ref})^5"    # 假设5年预测期
                f"+{dcf_sheet_name}!{sum_pv_fcf_cell}"
                f"+{dcf_sheet_name}!{net_cash_cell})"
                f"/{dcf_sheet_name}!{shares_cell}"
            )
            
            cell.value = formula
            cell.number_format = "$#,##0.00"
            cell.alignment = ALIGN_CENTER
            cell.border = BORDER_THIN
            
            # 中心单元格高亮
            if i == half and j == half:
                cell.fill = FILL_SENSITIVITY_CENTER
                cell.font = Font(name="Calibri", size=10, bold=True)
            else:
                cell.font = FONT_FORMULA

# 使用示例
# build_sensitivity_matrix(ws_val, start_row=40, start_col=2,
#                          wacc_base=0.09, tgr_base=0.025)
```

### Office JS 模板

```typescript
async function buildSensitivityMatrix(
    context: Excel.RequestContext,
    sheetName: string,
    startRow: number,
    startCol: number,
    waccBase: number,
    tgrBase: number,
    waccStep: number = 0.005,
    tgrStep: number = 0.005,
    size: number = 5
): Promise<void> {
    const sheet = context.workbook.worksheets.getItem(sheetName);
    const half = Math.floor(size / 2);
    
    for (let i = 0; i < size; i++) {
        for (let j = 0; j < size; j++) {
            const row = startRow + 1 + i;
            const col = startCol + 1 + j;
            const cell = sheet.getCell(row - 1, col - 1); // 0-indexed
            
            // TGR 和 WACC 引用矩阵表头
            const tgrRef = `${getColLetter(col)}$${startRow}`;
            const waccRef = `$${getColLetter(startCol)}$${row}`;
            
            // 完整 DCF 重算公式
            const formula = `=('DCF'!H30*(1+${tgrRef})/(${waccRef}-${tgrRef})/(1+${waccRef})^5+'DCF'!B35+'DCF'!B38)/'DCF'!B40`;
            cell.formulas = [[formula]];
            cell.format.numberFormat = "$#,##0.00";
            
            // 中心高亮
            if (i === half && j === half) {
                cell.format.fill.color = "#BDD7EE";
                cell.format.font.bold = true;
            }
        }
    }
    await context.sync();
}
```

---

## 5. Cell Comments 审计追踪 (Audit Trail via Comments)

### 规则

- **每个硬编码输入单元格必须有 cell comment**
- 格式标准：`"Source: [来源], [日期], [引用], [URL if applicable]"`
- **写值时立即添加 comment**，不要推迟到最后
- Comment 是 Gate 4 审核的必查项（`checklists/gate-4-delivery.md`）

### openpyxl Comment 代码

```python
from openpyxl.comments import Comment

def add_source_comment(ws, cell_ref, source, date, reference="", url=""):
    """为硬编码输入添加来源注释"""
    parts = [f"Source: {source}", date]
    if reference:
        parts.append(reference)
    if url:
        parts.append(url)
    
    comment_text = ", ".join(parts)
    ws[cell_ref].comment = Comment(comment_text, "Model Builder")

# 使用示例
def write_input_with_source(ws, cell_ref, value, source, date, 
                             reference="", url="", is_key_assumption=False):
    """写入硬编码值并同时添加来源注释和颜色"""
    cell = ws[cell_ref]
    cell.value = value
    
    # 颜色编码
    if is_key_assumption:
        style_key_assumption(ws, cell_ref, value)
    else:
        style_input(ws, cell_ref, value)
    
    # 审计追踪
    add_source_comment(ws, cell_ref, source, date, reference, url)

# 实操示例
write_input_with_source(
    ws, "B5", 1250.3,
    source="Company 10-K Filing",
    date="2024-02-15",
    reference="FY2023 Revenue, p.45",
    url="https://www.sec.gov/..."
)

write_input_with_source(
    ws, "B8", 0.12,
    source="Management Guidance",
    date="2024-01-20",
    reference="Q4 2023 Earnings Call",
    is_key_assumption=True
)
```

### Office JS Comment 等效

```typescript
async function addSourceComment(
    context: Excel.RequestContext,
    sheetName: string,
    cellRef: string,
    source: string,
    date: string
): Promise<void> {
    const sheet = context.workbook.worksheets.getItem(sheetName);
    const range = sheet.getRange(cellRef);
    range.addConditionalFormat; // Note: Office JS uses Notes API
    // context.workbook.comments.add(range, `Source: ${source}, ${date}`);
    // 注意：Office JS 的 comment API 在不同版本中差异较大
    // 建议使用 range.setNote() 或 context.workbook.comments
}
```

---

## 6. 模型布局规划原则 (Model Layout Planning)

### 黄金流程：先定位，后填充

```
Step 1: 定义 Row Map（所有 section 的行号位置）
  ↓
Step 2: 写所有 Headers 和 Labels（文本内容）
  ↓
Step 3: 写所有 Section Dividers 和空行
  ↓
Step 4: 写公式（使用锁定的行号变量）
  ↓
Step 5: 应用颜色编码和数字格式
  ↓
Step 6: 测试公式 + 立即验证
```

### Row Map 模板

```python
# ━━━━━ Income Statement Row Map ━━━━━
ROWS = {
    "header":       1,
    "period_labels": 2,
    "blank_1":      3,
    # Revenue Section
    "rev_header":   4,
    "rev_seg_1":    5,
    "rev_seg_2":    6,
    "rev_seg_3":    7,
    "rev_total":    8,
    "blank_2":      9,
    # COGS Section
    "cogs_header":  10,
    "cogs_product": 11,
    "cogs_service": 12,
    "cogs_total":   13,
    "blank_3":      14,
    # Gross Profit
    "gross_profit": 15,
    "gp_margin":    16,
    "blank_4":      17,
    # OpEx Section
    "opex_header":  18,
    "rd":           19,
    "sga":          20,
    "opex_total":   21,
    "blank_5":      22,
    # EBITDA
    "ebitda":       23,
    "ebitda_margin": 24,
    "da":           25,
    "ebit":         26,
    # Below the line
    "int_income":   27,
    "int_expense":  28,
    "ebt":          29,
    "tax":          30,
    "net_income":   31,
    "ni_margin":    32,
}

# 使用时通过 ROWS dict 引用行号，永不硬编码行号
# ws[f"D{ROWS['rev_total']}"] = f"=SUM(D{ROWS['rev_seg_1']}:D{ROWS['rev_seg_3']})"
```

### 关键原则

1. **Row Map 定义后不可随意修改** — 所有公式依赖这些行号
2. **先写 labels 再写公式** — 确保 row map 没有遗漏
3. **section 之间留空行** — 提高可读性，空行行号也要定义
4. **公式中使用 f-string + ROWS dict** — 可维护性远优于硬编码 `"=D15-D13"`
5. **测试后立即验证** — 每写完一个 section 就运行 recalc 检查

---

## 7. 常见陷阱与修复 (Common Pitfalls & Fixes)

### 陷阱 1: 合并单元格 (Merged Cells)

**问题**: merge 后写值导致 `ValueError` 或值丢失。

**修复**: **先写值到左上角单元格，再 merge 整个范围。**

```python
# openpyxl — 正确顺序
ws["B1"] = "Investment Committee Memorandum"  # 先写值
ws["B1"].font = FONT_HEADER
ws["B1"].alignment = Alignment(horizontal="center")
ws.merge_cells("B1:H1")                       # 再 merge
```

```typescript
// Office JS — 正确顺序
const range = sheet.getRange("B1:H1");
range.getCell(0, 0).values = [["Investment Committee Memorandum"]];  // 先写值
range.merge(true);                                                    // 再 merge
range.format.font.bold = true;
```

### 陷阱 2: formulas 库替换公式为值

**问题**: Python `formulas` 库计算后，公式被值覆盖，打开 xlsx 看不到公式。

**修复**: 使用 XML 双写方案（详见 `recalc-guide.md` 方案 C）。

```python
# 核心逻辑：
# 1. openpyxl 写公式 → 保存为 model_formulas.xlsx
# 2. formulas 库计算 → 保存为 model_values.xlsx
# 3. 解压两个 xlsx → 从 values 版本提取 <v> 标签 → 注入 formulas 版本
# 4. 最终 xlsx 同时包含公式和缓存值
```

### 陷阱 3: recalc 失败的 Fallback

```
方案 A (formulas 库) 失败
  → 检查: 是否有不支持的函数 (INDEX/MATCH/XLOOKUP)
  → Fallback: 方案 B (LibreOffice)

方案 B (LibreOffice) 失败
  → 检查: LibreOffice 是否安装 (which libreoffice)
  → Fallback: 方案 C (XML 双写)

方案 C (XML 双写) 失败
  → 检查: XML 解析错误，通常是命名空间问题
  → Fallback: 仅保留公式，不注入缓存值
  → 警告用户: "请在 Excel 中打开后按 Ctrl+Shift+F9 强制重算"
```

### 陷阱 4: 循环引用的 Excel 迭代设置

**问题**: LBO 模型中 Interest → Debt → Cash → Interest 形成有意循环，Excel 默认不允许。

**修复**:
1. 在 Assumptions 表添加 `Circ. Switch` 单元格（值 = 1 或 0）
2. 所有循环入口公式乘以 `Circ. Switch`
3. 提示用户开启迭代：`File > Options > Formulas > Enable Iterative Calculation`
4. 设置 Maximum Iterations = 100, Maximum Change = 0.001

```python
# openpyxl — 设置迭代计算（需要直接修改 workbook XML）
wb.calculation = openpyxl.workbook.properties.CalcProperties(
    iterate=True, iterateCount=100, iterateDelta=0.001
)
```

### 陷阱 5: 中文感叹号 "！" vs 英文 "!"

**问题**: 跨表引用中使用中文感叹号 `Assumptions！B5` 导致 `#NAME?` 错误。

**修复**: 始终使用英文半角 `!`。

```python
# ❌ 错误
formula = f"=Assumptions！$B${row}"

# ✅ 正确
formula = f"=Assumptions!$B${row}"
```

### 陷阱 6: 数字存为文本 (Leading Apostrophe)

**问题**: openpyxl 写入字符串 `"123.45"` 而非数字 `123.45`，导致 SUM 不计入。

**修复**: 确保 Python 写入的是 `float` 或 `int`，不是 `str`。

```python
# ❌ 错误
ws["B5"] = "1250.3"        # 存为文本
ws["B5"] = str(revenue)    # 也是文本

# ✅ 正确
ws["B5"] = 1250.3          # 数值
ws["B5"] = float(revenue)  # 确保是数值类型
```

### 陷阱 7: 千分位逗号在不同 locale 下的差异

**问题**: 某些欧洲 locale 使用 `.` 作为千分位，`,` 作为小数点，导致公式解析错误。

**修复**: 
- 公式中**永远不要**使用千分位符号
- 使用 `number_format` 属性控制显示格式
- 公式参数分隔符在 openpyxl 中始终用英文逗号 `,`

```python
# ❌ 错误：公式中放千分位
ws["B5"] = "=1,250.3*1.1"

# ✅ 正确：数值无千分位，格式单独设
ws["B5"] = 1250.3
ws["B5"].number_format = "#,##0.0"  # 显示时加千分位
```

### 陷阱 8: 数字格式不匹配

**问题**: 百分比存为 `15` 而非 `0.15`，或反之。

**修复**:
```python
# 百分比值应以小数存储
ws["B8"] = 0.15                    # 不是 15
ws["B8"].number_format = "0.0%"    # 显示为 15.0%

# 公式中引用百分比无需额外转换
ws["D10"] = f"=C10*(1+D{growth_row})"  # growth_row 存的是 0.15
```

---

## 附录: 快速参考表

| 操作 | openpyxl | Office JS |
|------|----------|-----------|
| 写公式 | `ws["A1"] = "=B1+C1"` | `range.formulas = [["=B1+C1"]]` |
| 写数值 | `ws["A1"] = 123.45` | `range.values = [[123.45]]` |
| 字体颜色 | `Font(color="0000FF")` | `range.format.font.color = "#0000FF"` |
| 背景色 | `PatternFill(start_color="FFFF99", fill_type="solid")` | `range.format.fill.color = "#FFFF99"` |
| 合并单元格 | `ws.merge_cells("A1:D1")` | `range.merge(true)` |
| 数字格式 | `cell.number_format = "#,##0.0"` | `range.numberFormat = "#,##0.0"` |
| Comment | `Comment("text", "author")` | `range.addConditionalFormat` / Notes API |
| 列宽 | `ws.column_dimensions["A"].width = 15` | `range.format.columnWidth = 100` (pt) |
| 冻结窗格 | `ws.freeze_panes = "C3"` | `sheet.freezePanes = sheet.getRange("C3")` |

---

**版本**: 4.0.0 | **最后更新**: 2026-04-12 | **相关文件**: `formula-patterns.md`, `recalc-guide.md`, `../config/color-spec.md`
