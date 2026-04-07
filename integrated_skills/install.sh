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
