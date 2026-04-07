#!/bin/bash

# FastClaw 技能集成脚本
# 将 Anthropic Skills 和 Superpowers 集成到 FastClaw

set -e

echo "🚀 FastClaw 技能集成脚本"
echo "=============================="

# 创建集成目录
INTEGRATED_DIR="./integrated_skills"
mkdir -p "$INTEGRATED_DIR"

# 技能源目录
ANTHROPIC_SKILLS="./temp_skills/skills-main/skills"
SUPERPOWERS_SKILLS="./temp_skills/superpowers-main/skills"

echo "📦 正在集成技能..."

# 1. 创建文档处理技能
echo ""
echo "📄 创建文档处理技能..."
mkdir -p "$INTEGRATED_DIR/document"

# DOCX 技能
if [ -d "$ANTHROPIC_SKILLS/docx" ]; then
    mkdir -p "$INTEGRATED_DIR/document/docx"
    cat > "$INTEGRATED_DIR/document/docx/SKILL.md" << 'SKILL_EOF'
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
SKILL_EOF
    echo "  ✅ 已创建 docx 技能"
fi

# PDF 技能
if [ -d "$ANTHROPIC_SKILLS/pdf" ]; then
    mkdir -p "$INTEGRATED_DIR/document/pdf"
    cat > "$INTEGRATED_DIR/document/pdf/SKILL.md" << 'SKILL_EOF'
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
SKILL_EOF
    echo "  ✅ 已创建 pdf 技能"
fi

# 2. 创建开发工作流技能
echo ""
echo "🔄 创建开发工作流技能..."
mkdir -p "$INTEGRATED_DIR/workflow"

# TDD 技能
if [ -d "$SUPERPOWERS_SKILLS/test-driven-development" ]; then
    mkdir -p "$INTEGRATED_DIR/workflow/tdd"
    cat > "$INTEGRATED_DIR/workflow/tdd/SKILL.md" << 'SKILL_EOF'
---
name: tdd
description: 测试驱动开发工作流。当用户提到TDD、测试先行、红绿重构、单元测试、测试驱动时触发此技能。
metadata:
  fastclaw:
    emoji: "✅"
    always: false
    source: "superpowers"
---

# 测试驱动开发(TDD)工作流

基于 Superpowers 的 test-driven-development 技能集成。完整的TDD开发流程。

## 工作流程

### 1. 红阶段 (RED)
- 编写一个失败的测试
- 只测试一个功能点
- 测试要尽可能简单

### 2. 绿阶段 (GREEN)
- 编写最小代码使测试通过
- 不要考虑代码质量
- 只关注功能实现

### 3. 重构阶段 (REFACTOR)
- 改进代码结构
- 消除重复
- 提高可读性

## 核心原则

1. **测试先行** - 先写测试，后写实现
2. **小步快跑** - 每次只实现一个微小功能
3. **持续重构** - 每次通过测试后都要重构
4. **简单设计** - 保持最简单的实现

## 使用方法

```
用户: "用TDD方式实现用户登录功能"
Agent: [触发tdd技能] "好的，让我们开始测试驱动开发:
1. 首先编写失败的登录测试
2. 实现最小登录功能
3. 重构代码..."

用户: "帮我对这个函数写单元测试"
Agent: [触发tdd技能] "我将按照TDD原则编写测试..."
```

## 最佳实践

- 每个测试只验证一个行为
- 测试名称应该描述预期行为
- 使用描述性的断言消息
- 保持测试独立和可重复

## 来源
- **项目**: Superpowers
- **原始技能**: test-driven-development
- **许可证**: MIT
SKILL_EOF
    echo "  ✅ 已创建 TDD 技能"
fi

# 系统化调试技能
if [ -d "$SUPERPOWERS_SKILLS/systematic-debugging" ]; then
    mkdir -p "$INTEGRATED_DIR/workflow/debugging"
    cat > "$INTEGRATED_DIR/workflow/debugging/SKILL.md" << 'SKILL_EOF'
---
name: debugging
description: 系统化调试流程。当用户提到调试、bug排查、问题诊断、错误修复、故障排除时触发此技能。
metadata:
  fastclaw:
    emoji: "🐛"
    always: false
    source: "superpowers"
---

# 系统化调试工作流

基于 Superpowers 的 systematic-debugging 技能集成。科学的问题诊断和修复流程。

## 四阶段调试流程

### 1. 重现问题 (REPRODUCE)
- 确定问题发生的条件
- 创建最小可重现示例
- 记录重现步骤

### 2. 收集信息 (GATHER)
- 收集错误日志和堆栈跟踪
- 记录系统状态和环境信息
- 识别相关变量和状态

### 3. 提出假设 (HYPOTHESIZE)
- 基于信息提出可能原因
- 按可能性排序假设
- 设计验证实验

### 4. 验证修复 (VERIFY)
- 实施最小修复
- 验证问题是否解决
- 确保没有引入新问题

## 调试工具

- **日志分析** - 结构化日志查询
- **断点调试** - 交互式代码检查
- **性能分析** - 瓶颈识别
- **内存分析** - 泄漏检测

## 使用方法

```
用户: "这个程序有bug，帮我调试"
Agent: [触发debugging技能] "我将使用系统化调试流程:
1. 首先重现问题
2. 收集错误信息
3. 分析可能原因..."

用户: "为什么这个函数返回错误结果?"
Agent: [触发debugging技能] "让我们逐步调试这个函数..."
```

## 最佳实践

- 一次只改变一个变量
- 保持详细的调试记录
- 使用版本控制二分查找
- 编写回归测试防止复发

## 来源
- **项目**: Superpowers
- **原始技能**: systematic-debugging
- **许可证**: MIT
SKILL_EOF
    echo "  ✅ 已创建 调试 技能"
fi

# 3. 创建安装和使用说明
echo ""
echo "📚 创建文档..."
cat > "$INTEGRATED_DIR/README.md" << 'README_EOF'
# FastClaw 集成技能包

本目录包含从 Anthropic Skills 和 Superpowers 项目集成的核心技能，可直接用于 FastClaw AI Agent。

## 🎯 快速开始

```bash
# 1. 复制技能到 FastClaw 目录
cp -r integrated_skills/* ~/.fastclaw/skills/

# 2. 启动 FastClaw
fastclaw

# 3. 开始使用技能
#    在聊天中尝试:
#    - "处理这个PDF文档" (触发pdf技能)
#    - "用TDD方式开发登录功能" (触发tdd技能)
#    - "帮我调试这个程序" (触发debugging技能)
```

## 📦 包含的技能

### 📄 文档处理技能 (document/)
- `docx/` - Microsoft Word 文档处理
- `pdf/` - PDF 文档提取和操作

### 🔄 开发工作流技能 (workflow/)
- `tdd/` - 测试驱动开发(TDD)完整工作流
- `debugging/` - 系统化调试和问题诊断

## 🚀 技能自动触发机制

FastClaw 会根据技能描述中的关键词自动触发技能加载:

- `docx` 技能: "Word文档、docx、文档编辑、格式排版"
- `pdf` 技能: "PDF、文档提取、表单处理、页面分割"  
- `tdd` 技能: "TDD、测试先行、红绿重构、单元测试"
- `debugging` 技能: "调试、bug排查、问题诊断、错误修复"

当用户聊天中包含这些关键词时，对应技能会自动加载。

## 🔧 自定义和扩展

### 修改技能触发
编辑技能目录中的 `SKILL.md` 文件，修改 `description` 字段:

```yaml
description: "处理Word文档。当用户提到[你的关键词]时触发此技能。"
```

### 添加新技能
1. 在对应类别目录下创建新技能文件夹
2. 按照 `SKILL.md` 格式编写技能
3. FastClaw 会自动发现和加载

### 技能目录结构
```
集成技能包/
├── document/          # 文档处理技能
│   ├── docx/
│   │   └── SKILL.md
│   └── pdf/
│       └── SKILL.md
├── workflow/          # 开发工作流技能
│   ├── tdd/
│   │   └── SKILL.md
│   └── debugging/
│       └── SKILL.md
└── README.md          # 说明文档
```

## 📋 使用示例

### 示例1: 文档处理
```
用户: "帮我把这篇Markdown转换成Word文档"
Agent: [自动触发docx技能]
"我将使用docx技能创建Word文档。
首先，我需要..."
```

### 示例2: TDD开发
```
用户: "用测试驱动开发实现用户注册功能"
Agent: [自动触发tdd技能]
"好的，让我们按照TDD流程:
1. 红阶段: 编写失败的注册测试
2. 绿阶段: 实现最小注册功能
3. 重构阶段: 优化代码结构..."
```

### 示例3: 系统调试
```
用户: "这个API返回500错误，帮我调试"
Agent: [自动触发debugging技能]
"开始系统化调试:
1. 重现问题: 什么请求导致500?
2. 收集信息: 查看错误日志
3. 提出假设: 可能是数据库连接问题
4. 验证修复: 测试修复方案..."
```

## ⚙️ 技术依赖

部分技能可能需要额外安装:

```bash
# 文档处理依赖
brew install pandoc poppler  # macOS
apt install pandoc poppler-utils  # Ubuntu

# Python库
pip install python-docx PyPDF2

# 开发工具
npm install -g jest mocha  # 测试框架
```

## 📝 许可证说明

- **Anthropic Skills**: 部分技能为专有许可证，仅限个人/教育使用
- **Superpowers**: MIT 许可证，可自由使用和修改

请遵守原始项目的许可证条款。

## 🔗 参考资源

- [Anthropic Skills 项目](https://github.com/anthropics/skills)
- [Superpowers 项目](https://github.com/obra/superpowers)
- [FastClaw 文档](https://github.com/fastclaw-ai/fastclaw)
- [技能开发指南](./documentation/skill-guide.md)

## 🆘 问题排查

### 技能未触发
1. 检查技能描述中的关键词是否匹配
2. 确认技能目录结构正确
3. 查看 FastClaw 日志获取详细信息

### 功能不工作
1. 检查是否安装了必要的依赖
2. 验证文件权限和路径
3. 测试原始项目中的示例是否正常

### 性能问题
1. 技能内容过多可拆分成子技能
2. 使用按需加载而不是always模式
3. 优化技能描述减少误触发

## 🤝 贡献指南

欢迎贡献改进和新的集成技能!

1. Fork 本仓库
2. 创建新的技能集成
3. 提交 Pull Request
4. 确保包含完整的文档

## 📞 支持

- GitHub Issues: 报告问题
- FastClaw 社区: 讨论和帮助
- 原始项目: 特定技能问题

---

**提示**: 首次使用时，建议先在测试环境中验证所有技能功能正常。
README_EOF

# 创建安装脚本
cat > "$INTEGRATED_DIR/install.sh" << 'INSTALL_EOF'
#!/bin/bash

echo "🚀 FastClaw 集成技能安装脚本"
echo "=============================="

# 检测安装目录
FASTCLAW_DIR="$HOME/.fastclaw"
if [ ! -d "$FASTCLAW_DIR" ]; then
    echo "⚠️  未找到 FastClaw 目录，将创建: $FASTCLAW_DIR"
    mkdir -p "$FASTCLAW_DIR"
fi

SKILLS_DIR="$FASTCLAW_DIR/skills"
mkdir -p "$SKILLS_DIR"

echo ""
echo "📦 正在安装集成技能..."

# 安装文档技能
if [ -d "document" ]; then
    echo "1. 安装文档处理技能..."
    cp -r document/* "$SKILLS_DIR/" 2>/dev/null || true
    echo "   ✅ 文档技能已安装"
fi

# 安装工作流技能
if [ -d "workflow" ]; then
    echo "2. 安装开发工作流技能..."
    cp -r workflow/* "$SKILLS_DIR/" 2>/dev/null || true
    echo "   ✅ 工作流技能已安装"
fi

# 统计技能数量
SKILL_COUNT=$(find "$SKILLS_DIR" -name "SKILL.md" 2>/dev/null | wc -l | tr -d ' ')

echo ""
echo "✅ 安装完成!"
echo "=============================="
echo "📊 安装统计:"
echo "  - FastClaw 目录: $FASTCLAW_DIR"
echo "  - 技能目录: $SKILLS_DIR"
echo "  - 已安装技能: $SKILL_COUNT 个"
echo ""
echo "🚀 下一步:"
echo "  1. 启动 FastClaw: fastclaw"
echo "  2. 在Web界面或聊天中测试技能"
echo "  3. 查看 README.md 获取详细使用指南"
echo "=============================="
INSTALL_EOF

chmod +x "$INTEGRATED_DIR/install.sh"

# 创建快速测试脚本
cat > "$INTEGRATED_DIR/test_skills.sh" << 'TEST_EOF'
#!/bin/bash

echo "🧪 FastClaw 技能测试脚本"
echo "=============================="

echo "测试技能触发关键词..."
echo ""

echo "1. 文档处理技能测试:"
echo "   - '处理Word文档' (应触发docx)"
echo "   - '提取PDF内容' (应触发pdf)"
echo ""

echo "2. 开发工作流技能测试:"
echo "   - '用TDD方式开发' (应触发tdd)"
echo "   - '帮我调试程序' (应触发debugging)"
echo ""

echo "3. 混合场景测试:"
echo "   - '先用TDD开发，再生成Word文档' (应触发tdd+docx)"
echo "   - '调试PDF处理问题' (应触发debugging+pdf)"
echo ""

echo "✅ 测试用例准备完成"
echo "在FastClaw中尝试上述短语，检查技能是否正确触发。"
echo "=============================="
TEST_EOF

chmod +x "$INTEGRATED_DIR/test_skills.sh"

# 统计创建的技能
SKILL_COUNT=$(find "$INTEGRATED_DIR" -name "SKILL.md" | wc -l | tr -d ' ')

echo ""
echo "✅ 集成完成!"
echo "=============================="
echo "📊 集成统计:"
echo "  - 总技能数: $SKILL_COUNT"
echo "  - 文档技能: 2个 (docx, pdf)"
echo "  - 工作流技能: 2个 (tdd, debugging)"
echo "  - 输出目录: $INTEGRATED_DIR"
echo ""
echo "🚀 使用方法:"
echo "  1. 进入集成目录: cd $INTEGRATED_DIR"
echo "  2. 运行安装脚本: ./install.sh"
echo "  3. 启动 FastClaw: fastclaw"
echo "  4. 测试技能触发"
echo ""
echo "📚 详细文档: $INTEGRATED_DIR/README.md"
echo "=============================="
