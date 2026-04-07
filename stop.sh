#!/bin/bash
# FastClaw停止脚本
# 使用方法: ./stop.sh

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
PID_FILE="${SCRIPT_DIR}/fastclaw.pid"

# 检查PID文件是否存在
if [ ! -f "${PID_FILE}" ]; then
    echo "PID文件不存在: ${PID_FILE}"
    echo "尝试查找FastClaw进程..."
    
    # 查找FastClaw进程
    PIDS=$(ps aux | grep -v grep | grep -E "(\./bin/fastclaw|fastclaw gateway)" | awk '{print $2}')
    
    if [ -z "${PIDS}" ]; then
        echo "未找到正在运行的FastClaw进程"
        exit 0
    fi
    
    echo "找到以下FastClaw进程:"
    ps aux | grep -v grep | grep -E "(\./bin/fastclaw|fastclaw gateway)"
    
    echo -n "是否要停止所有FastClaw进程? [y/N] "
    read -r confirm
    if [[ "${confirm}" =~ ^[Yy]$ ]]; then
        echo "正在停止FastClaw进程..."
        kill ${PIDS} 2>/dev/null
        sleep 1
        
        # 检查是否已停止
        REMAINING=$(ps aux | grep -v grep | grep -E "(\./bin/fastclaw|fastclaw gateway)" | awk '{print $2}')
        if [ -n "${REMAINING}" ]; then
            echo "强制停止剩余进程..."
            kill -9 ${REMAINING} 2>/dev/null
        fi
        
        echo "所有FastClaw进程已停止"
    else
        echo "操作取消"
        exit 0
    fi
else
    PID=$(cat "${PID_FILE}")
    
    if ps -p "${PID}" > /dev/null 2>&1; then
        echo "正在停止FastClaw (PID: ${PID})..."
        
        # 优雅停止
        kill "${PID}"
        
        # 等待最多10秒
        timeout=10
        while [ ${timeout} -gt 0 ] && ps -p "${PID}" > /dev/null 2>&1; do
            echo "等待进程停止... (${timeout}s)"
            sleep 1
            timeout=$((timeout - 1))
        done
        
        if ps -p "${PID}" > /dev/null 2>&1; then
            echo "进程未响应，强制停止..."
            kill -9 "${PID}"
            sleep 1
        fi
        
        echo "FastClaw已停止"
    else
        echo "进程 ${PID} 已停止或不存在"
    fi
    
    # 删除PID文件
    rm -f "${PID_FILE}"
    echo "PID文件已移除"
fi

# 清理可能存在的其他临时文件
rm -f "${SCRIPT_DIR}/.fastclaw.running" 2>/dev/null

echo "清理完成"