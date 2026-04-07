#!/bin/bash
# FastClaw启动脚本 - 带日志功能
# 使用方法: ./start-with-logs.sh [port]
# 默认端口: 18953

set -e

# 配置
SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
FASTCLAW_BIN="${SCRIPT_DIR}/bin/fastclaw"
CONFIG_DIR="${HOME}/.fastclaw"
LOG_DIR="${SCRIPT_DIR}/logs"
TIMESTAMP=$(date +%Y%m%d-%H%M%S)

# 参数处理
PORT=${1:-18953}

# 创建日志目录
mkdir -p "${LOG_DIR}"

# 设置代理
export http_proxy=http://127.0.0.1:7890
export https_proxy=http://127.0.0.1:7890
export HTTP_PROXY=http://127.0.0.1:7890
export HTTPS_PROXY=http://127.0.0.1:7890

# 日志文件
STDOUT_LOG="${LOG_DIR}/fastclaw-${TIMESTAMP}.stdout.log"
STDERR_LOG="${LOG_DIR}/fastclaw-${TIMESTAMP}.stderr.log"
COMBINED_LOG="${LOG_DIR}/fastclaw-${TIMESTAMP}.combined.log"

# PID文件
PID_FILE="${SCRIPT_DIR}/fastclaw.pid"

# 检查二进制文件
if [ ! -f "${FASTCLAW_BIN}" ]; then
    echo "错误: FastClaw二进制文件不存在: ${FASTCLAW_BIN}"
    echo "请先运行: make build"
    exit 1
fi

# 检查权限
if [ ! -x "${FASTCLAW_BIN}" ]; then
    chmod +x "${FASTCLAW_BIN}"
fi

# 检查是否已在运行
if [ -f "${PID_FILE}" ]; then
    OLD_PID=$(cat "${PID_FILE}")
    if ps -p "${OLD_PID}" > /dev/null 2>&1; then
        echo "FastClaw已在运行 (PID: ${OLD_PID})"
        echo "如果要重启，请先运行: ./stop.sh"
        exit 0
    else
        echo "移除旧的PID文件..."
        rm -f "${PID_FILE}"
    fi
fi

# 启动函数
start_fastclaw() {
    echo "启动 FastClaw..."
    echo "端口: ${PORT}"
    echo "标准输出日志: ${STDOUT_LOG}"
    echo "标准错误日志: ${STDERR_LOG}"
    echo "组合日志: ${COMBINED_LOG}"
    echo ""
    
    # 启动命令
    "${FASTCLAW_BIN}" gateway --port "${PORT}" \
        > >(tee -a "${STDOUT_LOG}" | tee -a "${COMBINED_LOG}") \
        2> >(tee -a "${STDERR_LOG}" | tee -a "${COMBINED_LOG}" >&2) &
    
    PID=$!
    echo "${PID}" > "${PID_FILE}"
    
    echo "FastClaw已启动 (PID: ${PID})"
    echo "访问: http://localhost:${PORT}"
    echo ""
    echo "查看实时日志: tail -f ${COMBINED_LOG}"
    echo "停止服务: ./stop.sh"
    
    # 等待服务启动
    sleep 2
    
    # 检查是否成功启动
    if ! ps -p "${PID}" > /dev/null 2>&1; then
        echo "错误: FastClaw启动失败，请检查日志:"
        tail -20 "${STDERR_LOG}"
        rm -f "${PID_FILE}"
        exit 1
    fi
    
    return 0
}

# 清理旧日志函数
cleanup_old_logs() {
    local keep_days=7
    echo "清理${keep_days}天前的日志..."
    
    # 查找并删除旧日志
    find "${LOG_DIR}" -name "fastclaw-*.stdout.log" -mtime +${keep_days} -delete 2>/dev/null || true
    find "${LOG_DIR}" -name "fastclaw-*.stderr.log" -mtime +${keep_days} -delete 2>/dev/null || true
    find "${LOG_DIR}" -name "fastclaw-*.combined.log" -mtime +${keep_days} -delete 2>/dev/null || true
    
    echo "日志清理完成"
}

# 显示配置信息
show_config_info() {
    echo "=== FastClaw配置信息 ==="
    echo "二进制文件: ${FASTCLAW_BIN}"
    echo "配置目录: ${CONFIG_DIR}"
    echo "日志目录: ${LOG_DIR}"
    
    if [ -f "${CONFIG_DIR}/fastclaw.json" ]; then
        echo ""
        echo "当前存储配置:"
        grep -A3 '"storage"' "${CONFIG_DIR}/fastclaw.json" || echo "  使用默认文件存储"
        
        echo ""
        echo "已配置的代理:"
        grep -o '"id":"[^"]*"' "${CONFIG_DIR}/fastclaw.json" | sed 's/"id":"//;s/"//' | uniq
    else
        echo "未找到配置文件，将启动设置向导"
    fi
    echo ""
}

# 主程序
main() {
    echo "=========================================="
    echo "     FastClaw 启动脚本 v1.0"
    echo "=========================================="
    
    # 显示配置信息
    show_config_info
    
    # 清理旧日志
    cleanup_old_logs
    
    # 启动服务
    start_fastclaw
    
    echo "=========================================="
}

# 运行主程序
main "$@"
