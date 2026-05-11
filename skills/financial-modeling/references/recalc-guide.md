> 本文件是 `full-spectrum-modeling` 的参考资料。
> Core 建模原则见 `../integrated-modeling-core/SKILL.md` §9。
> Excel 工程化执行指南见 `excel-engineering-guide.md`。

# Excel 重算指南 (Recalculation Guide)

standalone `.xlsx` 生成（非 Office JS 环境）时，openpyxl 写入的公式不会自动计算——打开文件时单元格显示为空或旧值。本指南提供三种重算方案及选择逻辑。

---

## 方案选择逻辑

```
┌───────────────────────────────────────────────────────┐
│ IF Office JS 环境 (Add-in / Excel 在线):              │
│   → 不需要 recalc — Excel 引擎原生计算                 │
│   → 直接跳过本指南                                     │
│                                                        │
│ ELIF LibreOffice 已安装:                               │
│   → 方案 B (LibreOffice headless)                     │
│   → 保留公式 + 准确计算 + 无 Python 依赖              │
│                                                        │
│ ELIF 需要保留公式且有 formulas 库:                      │
│   → 方案 C (XML 双写) ← 推荐                          │
│   → 公式保留 + 缓存值注入 + 纯 Python                  │
│                                                        │
│ ELIF 不关心公式保留 (临时分析):                         │
│   → 方案 A (formulas 库直算)                           │
│   → 最快最简单，但公式被值覆盖                          │
│                                                        │
│ ELSE (无 formulas 库、无 LibreOffice):                 │
│   → 仅保留公式，不注入缓存值                           │
│   → 提示用户: "请在 Excel 中 Ctrl+Shift+F9 强制重算"   │
└───────────────────────────────────────────────────────┘
```

---

## 方案 A: formulas 库（Python 原生）

### 特点

| 优点 | 缺点 |
|------|------|
| 无外部依赖（纯 Python） | 公式被计算值覆盖 |
| 安装简单：`pip install formulas` | 打开 xlsx 后看不到公式 |
| 支持大多数 Excel 函数 | 不支持 XLOOKUP/FILTER 等新函数 |

### 代码模板

```python
import formulas

def recalc_method_a(input_path: str, output_path: str):
    """
    方案 A: 使用 formulas 库直接计算。
    警告: 公式会被计算结果覆盖。
    """
    xl_model = formulas.ExcelModel().loads(input_path).finish()
    solution = xl_model.calculate()
    xl_model.write(dirpath=output_path)
    
    print(f"✅ 方案 A 完成: {output_path}")
    print("⚠️ 注意: 公式已被计算值覆盖")
```

### 适用场景

- 一次性分析，不需要交付给他人
- 快速验证模型正确性
- 后续会用方案 C 生成最终版本

---

## 方案 B: LibreOffice headless

### 特点

| 优点 | 缺点 |
|------|------|
| 保留所有公式 | 需要安装 LibreOffice |
| 计算引擎准确 | 某些 Excel 专有函数不支持 |
| 批量处理友好 | 服务器环境需额外配置 |

### 环境检测

```python
import shutil
import subprocess

def check_libreoffice() -> str | None:
    """检测 LibreOffice 是否可用，返回路径或 None"""
    for cmd in ["libreoffice", "soffice"]:
        path = shutil.which(cmd)
        if path:
            return path
    return None
```

### 代码模板

```python
import subprocess
import shutil
from pathlib import Path

def recalc_method_b(input_path: str, output_dir: str = "."):
    """
    方案 B: 使用 LibreOffice headless 重算。
    保留公式，计算缓存值。
    """
    lo_path = check_libreoffice()
    if not lo_path:
        raise RuntimeError("LibreOffice 未安装。请安装后重试或使用方案 C。")
    
    cmd = [
        lo_path,
        "--headless",
        "--calc",
        "--convert-to", "xlsx",
        "--outdir", output_dir,
        input_path
    ]
    
    result = subprocess.run(cmd, capture_output=True, text=True, timeout=120)
    
    if result.returncode != 0:
        raise RuntimeError(f"LibreOffice 重算失败: {result.stderr}")
    
    print(f"✅ 方案 B 完成: {output_dir}/{Path(input_path).name}")
    print("✓ 公式已保留，缓存值已更新")
```

### 安装 LibreOffice

```bash
# Ubuntu / Debian
sudo apt-get install libreoffice-calc

# macOS (Homebrew)
brew install --cask libreoffice

# Docker (Alpine)
apk add libreoffice-calc
```

---

## 方案 C: XML 双写（推荐）

### 原理

xlsx 文件本质是 ZIP 包，内部 XML 格式支持同时保存公式和缓存值：

```xml
<c r="D10" s="1">
    <f>C10*(1+D5)</f>      <!-- 公式 -->
    <v>1234.56</v>          <!-- 缓存值 -->
</c>
```

openpyxl 只写入 `<f>` 标签（公式），不写 `<v>` 标签。XML 双写方案的核心是：
1. 用 openpyxl 生成带公式的 xlsx
2. 用 formulas 库计算值
3. 将计算值注入到 XML 的 `<v>` 标签中
4. 最终 xlsx 同时包含公式和显示值

### 完整代码模板

```python
import zipfile
import shutil
import os
import re
from pathlib import Path
from lxml import etree

def recalc_method_c(formula_xlsx: str, output_xlsx: str):
    """
    方案 C (推荐): XML 双写。
    保留公式 + 注入缓存值。
    
    Args:
        formula_xlsx: openpyxl 生成的带公式的 xlsx 文件
        output_xlsx: 最终输出路径
    """
    import formulas
    
    # Step 1: 用 formulas 库计算值（会覆盖公式）
    values_xlsx = formula_xlsx.replace(".xlsx", "_values.xlsx")
    xl_model = formulas.ExcelModel().loads(formula_xlsx).finish()
    xl_model.calculate()
    xl_model.write(dirpath=os.path.dirname(values_xlsx) or ".")
    
    # Step 2: 解压两个 xlsx
    formula_dir = formula_xlsx + "_extracted"
    values_dir = values_xlsx + "_extracted"
    
    with zipfile.ZipFile(formula_xlsx, "r") as z:
        z.extractall(formula_dir)
    with zipfile.ZipFile(values_xlsx, "r") as z:
        z.extractall(values_dir)
    
    # Step 3: 从 values 版本提取 <v> 标签，注入 formula 版本
    ns = {"s": "http://schemas.openxmlformats.org/spreadsheetml/2006/main"}
    
    sheet_dir_f = os.path.join(formula_dir, "xl", "worksheets")
    sheet_dir_v = os.path.join(values_dir, "xl", "worksheets")
    
    for sheet_file in os.listdir(sheet_dir_f):
        if not sheet_file.endswith(".xml"):
            continue
        
        tree_f = etree.parse(os.path.join(sheet_dir_f, sheet_file))
        values_path = os.path.join(sheet_dir_v, sheet_file)
        
        if not os.path.exists(values_path):
            continue
        
        tree_v = etree.parse(values_path)
        
        # 建立 values 版本的 cell → value 映射
        value_map = {}
        for cell_v in tree_v.findall(".//s:c", ns):
            ref = cell_v.get("r")
            v_elem = cell_v.find("s:v", ns)
            if v_elem is not None and ref:
                value_map[ref] = v_elem.text
        
        # 为 formula 版本中有公式的 cell 注入 <v>
        for cell_f in tree_f.findall(".//s:c", ns):
            ref = cell_f.get("r")
            f_elem = cell_f.find("s:f", ns)
            if f_elem is not None and ref in value_map:
                # 移除已有的 <v>（如果有）
                for old_v in cell_f.findall("s:v", ns):
                    cell_f.remove(old_v)
                # 注入新的 <v>
                new_v = etree.SubElement(cell_f, f"{{{ns['s']}}}v")
                new_v.text = value_map[ref]
        
        tree_f.write(os.path.join(sheet_dir_f, sheet_file),
                     xml_declaration=True, encoding="UTF-8", standalone=True)
    
    # Step 4: 重新打包为 xlsx
    with zipfile.ZipFile(output_xlsx, "w", zipfile.ZIP_DEFLATED) as zout:
        for root, dirs, files in os.walk(formula_dir):
            for file in files:
                file_path = os.path.join(root, file)
                arcname = os.path.relpath(file_path, formula_dir)
                zout.write(file_path, arcname)
    
    # 清理临时文件
    shutil.rmtree(formula_dir, ignore_errors=True)
    shutil.rmtree(values_dir, ignore_errors=True)
    if os.path.exists(values_xlsx):
        os.remove(values_xlsx)
    
    print(f"✅ 方案 C 完成: {output_xlsx}")
    print("✓ 公式已保留 + 缓存值已注入")

# 验证
def verify_dual_write(xlsx_path: str):
    """验证 xlsx 同时包含公式和缓存值"""
    from openpyxl import load_workbook
    
    # 检查公式存在
    wb_formula = load_workbook(xlsx_path)
    # 检查值存在
    wb_value = load_workbook(xlsx_path, data_only=True)
    
    for ws_name in wb_formula.sheetnames:
        ws_f = wb_formula[ws_name]
        ws_v = wb_value[ws_name]
        formula_count = 0
        value_count = 0
        
        for row in ws_f.iter_rows():
            for cell in row:
                if isinstance(cell.value, str) and cell.value.startswith("="):
                    formula_count += 1
                    # 对应的 data_only 版本应有值
                    val_cell = ws_v[cell.coordinate]
                    if val_cell.value is not None:
                        value_count += 1
        
        if formula_count > 0:
            coverage = value_count / formula_count * 100
            print(f"  {ws_name}: {formula_count} formulas, {value_count} cached values ({coverage:.0f}% coverage)")
    
    wb_formula.close()
    wb_value.close()
```

---

## 方案调度器 (Auto-Selector)

```python
def auto_recalc(input_xlsx: str, output_xlsx: str):
    """自动选择最佳 recalc 方案并执行"""
    
    # 尝试方案 B (LibreOffice)
    lo_path = check_libreoffice()
    if lo_path:
        print("→ 检测到 LibreOffice，使用方案 B")
        recalc_method_b(input_xlsx, os.path.dirname(output_xlsx))
        return
    
    # 尝试方案 C (XML 双写)
    try:
        import formulas
        print("→ 检测到 formulas 库，使用方案 C (XML 双写)")
        recalc_method_c(input_xlsx, output_xlsx)
        return
    except ImportError:
        pass
    
    # Fallback: 无 recalc
    print("⚠️ 无可用 recalc 方案")
    print("→ 公式已保留，但无缓存值")
    print("→ 请在 Excel 中打开后按 Ctrl+Shift+F9 强制重算")
    shutil.copy2(input_xlsx, output_xlsx)
```

---

## 注意事项

| 项目 | 说明 |
|------|------|
| 函数覆盖 | formulas 库不支持 XLOOKUP、FILTER、UNIQUE 等新函数；使用这些函数时只能用方案 B 或跳过 recalc |
| 循环引用 | 方案 A 和 C 不支持迭代计算；有循环引用的 LBO 模型建议用方案 B 或手动在 Excel 中 recalc |
| 性能 | 大型模型（>10MB）方案 C 的 XML 解析可能较慢；考虑方案 B |
| 编码 | XML 双写时确保 UTF-8 编码，避免中文字符乱码 |
| 验证 | 无论使用哪种方案，最终都应运行 `verify_dual_write()` 确认结果 |

---

**版本**: 4.0.0 | **最后更新**: 2026-04-12 | **相关文件**: `excel-engineering-guide.md`, `formula-patterns.md`
