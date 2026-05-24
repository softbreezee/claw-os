---
name: debugging
description: 系统化调试流程。当用户提到调试、bug排查、问题诊断、错误修复、故障排除时触发此技能。
tags: [dev, tool, troubleshooting]
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
