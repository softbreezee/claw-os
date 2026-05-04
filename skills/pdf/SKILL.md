---
name: pdf
description: 处理PDF文档。当用户提到PDF、文档提取、表单处理、页面分割、PDF转换时触发此技能。
metadata:
  fastclaw:
    emoji: "📑"
    always: false
    source: "anthropic"
---

# PDF 文档处理技能

基于 Anthropic Skills 的 pdf 技能集成。用于提取、编辑和操作 PDF 文档。

## 主要功能

1. **文本提取** - 从PDF提取文本内容
2. **表单处理** - 处理PDF表单字段
3. **页面操作** - 分割、合并、旋转页面
4. **文档转换** - PDF与其他格式互转
5. **元数据提取** - 获取文档信息

## 使用方法

```
用户: "提取这个PDF文件的第3-5页"
Agent: [触发pdf技能] "正在使用PDF技能提取指定页面..."

用户: "把这个PDF转换成Word文档"
Agent: [触发pdf技能] "正在转换PDF到Word格式..."
```

## 技术依赖

- `pdftotext`: PDF文本提取
- `pdftk`: PDF工具包
- `python-pdf2`: Python PDF处理库

## 来源
- **项目**: Anthropic Skills
- **原始技能**: pdf
- **许可证**: Proprietary (详见原项目)
