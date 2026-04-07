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
