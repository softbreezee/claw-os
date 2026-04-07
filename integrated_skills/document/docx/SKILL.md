---
name: docx
description: 处理Microsoft Word文档(.docx文件)。当用户提到Word文档、docx、文档编辑、格式排版、文档转换时触发此技能。
metadata:
  fastclaw:
    emoji: "📄"
    always: false
    source: "anthropic"
---

# DOCX 文档处理技能

基于 Anthropic Skills 的 docx 技能集成。用于创建、读取、编辑和操作 Word 文档。

## 主要功能

1. **文档创建** - 创建新的Word文档
2. **内容提取** - 从.docx文件中提取文本和结构
3. **格式编辑** - 修改文档格式和样式
4. **文档转换** - 与其他格式互转
5. **批量处理** - 自动化文档操作

## 使用方法

```
用户: "帮我把这个文本转换成Word文档"
Agent: [触发docx技能] "我将使用docx技能创建Word文档..."

用户: "提取这个Word文档的所有标题"
Agent: [触发docx技能] "正在分析文档结构..."
```

## 技术依赖

- `pandoc`: 文档格式转换
- `python-docx`: Python文档处理库
- `LibreOffice`: 文档转换工具

## 来源
- **项目**: Anthropic Skills
- **原始技能**: docx
- **许可证**: Proprietary (详见原项目)
