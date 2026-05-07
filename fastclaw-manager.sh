#!/bin/bash
# FastClaw 开发版管理脚本
# 功能：安装到全局、启动服务、停止服务、检查状态
# 用法：./fastclaw-manager.sh [command] [options]
# 命令：install, start, stop, status, restart, logs, deploy

set -e

# 配置
SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
SOURCE_BIN="${SCRIPT_DIR}/bin/fastclaw"
CONFIG_DIR="${HOME}/.fastclaw"
LOG_DIR="${SCRIPT_DIR}/logs"
PID_FILE="${SCRIPT_DIR}/fastclaw.pid"

# 全局安装配置
INSTALL_DIR="${HOME}/.local/bin"  # 默认安装目录，优先于/usr/local/bin
GLOBAL_BIN_NAME="fastclaw"

# 服务配置
DEFAULT_PORT=18953

# 代理配置（根据需要修改，代理服务未运行时请注释下面两行）
# HTTP_PROXY="http://127.0.0.1:7890"
# HTTPS_PROXY="http://127.0.0.1:7890"
HTTP_PROXY=""
HTTPS_PROXY=""

# 颜色定义
if [ -t 1 ]; then
    RED='\033[0;31m'
    GREEN='\033[0;32m'
    YELLOW='\033[1;33m'
    BLUE='\033[0;34m'
    BOLD='\033[1m'
    NC='\033[0m'
else
    RED=''; GREEN=''; YELLOW=''; BLUE=''; BOLD=''; NC=''
fi

info()    { printf "${GREEN}[INFO]${NC} %s\n" "$*"; }
warn()    { printf "${YELLOW}[WARN]${NC} %s\n" "$*"; }
error()   { printf "${RED}[ERROR]${NC} %s\n" "$*" >&2; }
success() { printf "${GREEN}[✓]${NC} %s\n" "$*"; }
heading() { printf "\n${BLUE}${BOLD}%s${NC}\n" "$*"; }

# 检查源二进制文件
check_source_binary() {
    if [ ! -f "${SOURCE_BIN}" ]; then
        error "FastClaw二进制文件不存在: ${SOURCE_BIN}"
        error "请先运行: make build"
        exit 1
    fi
    
    if [ ! -x "${SOURCE_BIN}" ]; then
        chmod +x "${SOURCE_BIN}"
    fi
    
    # 获取版本信息
    VERSION=$("${SOURCE_BIN}" --version 2>/dev/null | head -1 | sed 's/.*version //' || echo "development")
    info "FastClaw版本: ${VERSION}"
}

# 安装到全局
install_global() {
    heading "安装 FastClaw 到全局"
    
    check_source_binary
    
    # 创建安装目录
    mkdir -p "${INSTALL_DIR}"
    
    # 安装二进制文件
    GLOBAL_BIN="${INSTALL_DIR}/${GLOBAL_BIN_NAME}"
    
    if [ -f "${GLOBAL_BIN}" ]; then
        warn "已存在全局安装，备份原文件..."
        mv "${GLOBAL_BIN}" "${GLOBAL_BIN}.bak.$(date +%Y%m%d%H%M%S)"
    fi
    
    info "安装到: ${GLOBAL_BIN}"
    cp "${SOURCE_BIN}" "${GLOBAL_BIN}"
    chmod +x "${GLOBAL_BIN}"
    
    # 确保PATH包含安装目录
    ensure_path_in_shell
    
    success "安装完成"
    echo ""
    echo "现在可以在任何位置使用: ${BOLD}fastclaw${NC}"
    echo "启动网关: ${BOLD}fastclaw gateway${NC}"
    echo "查看技能: ${BOLD}fastclaw skill list${NC}"
    
    # 测试安装
    if command -v "${GLOBAL_BIN_NAME}" >/dev/null 2>&1; then
        echo ""
        echo "测试命令:"
        "${GLOBAL_BIN_NAME}" --version
    else
        warn "安装目录可能不在PATH中，请运行: source ~/.zshrc 或重新打开终端"
    fi
}

# 确保shell配置中包含安装目录
ensure_path_in_shell() {
    # 检查是否已在PATH中
    case ":$PATH:" in
        *":${INSTALL_DIR}:"*) return 0 ;;
    esac
    
    # 检测shell类型
    local shell_rc=""
    local shell_name=$(basename "${SHELL:-sh}")
    
    case "${shell_name}" in
        zsh)    shell_rc="${HOME}/.zshrc" ;;
        bash)   shell_rc="${HOME}/.bashrc" ;;
        fish)   shell_rc="${HOME}/.config/fish/config.fish" ;;
        *)      shell_rc="${HOME}/.profile" ;;
    esac
    
    # 检查是否已配置
    if [ -f "${shell_rc}" ] && grep -q "${INSTALL_DIR}" "${shell_rc}" 2>/dev/null; then
        return 0
    fi
    
    # 添加PATH配置
    info "添加 ${INSTALL_DIR} 到 PATH (${shell_rc})"
    
    if [ "${shell_name}" = "fish" ]; then
        mkdir -p "$(dirname "${shell_rc}")"
        printf '\n# FastClaw (development)\nfish_add_path "%s"\n' "${INSTALL_DIR}" >> "${shell_rc}"
    else
        printf '\n# FastClaw (development)\nexport PATH="%s:$PATH"\n' "${INSTALL_DIR}" >> "${shell_rc}"
    fi
    
    export PATH="${INSTALL_DIR}:${PATH}"
    warn "PATH已更新，请运行: source ${shell_rc} 或重新打开终端使其永久生效"
}

# 启动服务（带日志）
start_service() {
    local port=${1:-${DEFAULT_PORT}}
    # Date-only timestamp so multiple restarts within a day append to the
    # same log file instead of creating a new file per start.
    local timestamp=$(date +%Y%m%d)
    local boundary=$(date +%Y-%m-%d\ %H:%M:%S)
    
    heading "启动 FastClaw 服务"
    
    check_source_binary
    
    # 创建日志目录
    mkdir -p "${LOG_DIR}"
    
    # 设置代理环境变量（HTTP_PROXY/HTTPS_PROXY 为空时不生效）
    [ -n "${HTTP_PROXY}" ] && export http_proxy="${HTTP_PROXY}"
    [ -n "${HTTPS_PROXY}" ] && export https_proxy="${HTTPS_PROXY}"
    [ -n "${HTTP_PROXY}" ] && export HTTP_PROXY="${HTTP_PROXY}"
    [ -n "${HTTPS_PROXY}" ] && export HTTPS_PROXY="${HTTPS_PROXY}"
    
    # 日志文件
    local stdout_log="${LOG_DIR}/fastclaw-${timestamp}.stdout.log"
    local stderr_log="${LOG_DIR}/fastclaw-${timestamp}.stderr.log"
    local combined_log="${LOG_DIR}/fastclaw-${timestamp}.combined.log"
    
    # 检查是否已在运行
    if [ -f "${PID_FILE}" ]; then
        local old_pid=$(cat "${PID_FILE}")
        if ps -p "${old_pid}" >/dev/null 2>&1; then
            error "FastClaw已在运行 (PID: ${old_pid})"
            error "如果要重启，请先运行: $0 stop"
            exit 1
        else
            warn "移除旧的PID文件..."
            rm -f "${PID_FILE}"
        fi
    fi
    
    # 清理旧日志（保留7天）
    cleanup_old_logs
    
    info "启动参数:"
    info "  端口: ${port}"
    info "  代理: ${HTTP_PROXY}"
    info "  日志文件: ${combined_log}"
    echo ""

    # Boundary marker so multiple starts within one day file are visible.
    {
        printf '\n'
        printf '========================================\n'
        printf '[START] %s  PID=will-be-set  port=%s\n' "${boundary}" "${port}"
        printf '========================================\n'
    } | tee -a "${combined_log}" "${stdout_log}" >/dev/null

    # 启动命令
    "${SOURCE_BIN}" gateway --port "${port}" \
        > >(tee -a "${stdout_log}" | tee -a "${combined_log}") \
        2> >(tee -a "${stderr_log}" | tee -a "${combined_log}" >&2) &
    
    local pid=$!
    echo "${pid}" > "${PID_FILE}"
    
    success "FastClaw已启动 (PID: ${pid})"
    echo ""
    echo "访问地址: ${BOLD}http://localhost:${port}${NC}"
    echo "查看日志: ${BOLD}tail -f ${combined_log}${NC}"
    echo "停止服务: ${BOLD}$0 stop${NC}"
    
    # 等待服务启动
    sleep 2
    
    # 检查是否成功启动
    if ! ps -p "${pid}" >/dev/null 2>&1; then
        error "FastClaw启动失败，请检查日志:"
        tail -20 "${stderr_log}"
        rm -f "${PID_FILE}"
        exit 1
    fi
}

# 停止服务
stop_service() {
    heading "停止 FastClaw 服务"
    
    if [ ! -f "${PID_FILE}" ]; then
        error "PID文件不存在: ${PID_FILE}"
        error "尝试查找FastClaw进程..."
        
        local pids=$(ps aux | grep -v grep | grep -E "(\./bin/fastclaw|fastclaw gateway)" | awk '{print $2}')
        
        if [ -z "${pids}" ]; then
            warn "未找到正在运行的FastClaw进程"
            return 0
        fi
        
        echo "找到以下FastClaw进程:"
        ps aux | grep -v grep | grep -E "(\./bin/fastclaw|fastclaw gateway)"
        
        echo -n "是否要停止所有FastClaw进程? [y/N] "
        read -r confirm
        if [[ "${confirm}" =~ ^[Yy]$ ]]; then
            info "正在停止FastClaw进程..."
            kill ${pids} 2>/dev/null
            sleep 1
            
            # 强制停止剩余进程
            local remaining=$(ps aux | grep -v grep | grep -E "(\./bin/fastclaw|fastclaw gateway)" | awk '{print $2}')
            if [ -n "${remaining}" ]; then
                warn "强制停止剩余进程..."
                kill -9 ${remaining} 2>/dev/null
            fi
            
            success "所有FastClaw进程已停止"
        else
            info "操作取消"
            exit 0
        fi
    else
        local pid=$(cat "${PID_FILE}")
        
        if ps -p "${pid}" >/dev/null 2>&1; then
            info "正在停止FastClaw (PID: ${pid})..."
            
            # 优雅停止
            kill "${pid}"
            
            # 等待最多10秒
            local timeout=10
            while [ ${timeout} -gt 0 ] && ps -p "${pid}" >/dev/null 2>&1; do
                echo "等待进程停止... (${timeout}s)"
                sleep 1
                timeout=$((timeout - 1))
            done
            
            if ps -p "${pid}" >/dev/null 2>&1; then
                warn "进程未响应，强制停止..."
                kill -9 "${pid}"
                sleep 1
            fi
            
            success "FastClaw已停止"
        else
            warn "进程 ${pid} 已停止或不存在"
        fi
        
        # 删除PID文件
        rm -f "${PID_FILE}"
        info "PID文件已移除"
    fi
}

# 检查服务状态
check_status() {
    heading "FastClaw 服务状态"
    
    check_source_binary
    
    echo "二进制文件: ${SOURCE_BIN}"
    echo "配置文件目录: ${CONFIG_DIR}"
    echo "日志目录: ${LOG_DIR}"
    echo ""
    
    if [ -f "${PID_FILE}" ]; then
        local pid=$(cat "${PID_FILE}")
        if ps -p "${pid}" >/dev/null 2>&1; then
            success "服务运行中 (PID: ${pid})"
            
            # 获取端口信息
            local port_info=$(lsof -i :${DEFAULT_PORT} -P 2>/dev/null | grep LISTEN || echo "未知")
            echo "端口 ${DEFAULT_PORT}: ${port_info}"
            
            # 获取进程信息
            echo "进程信息:"
            ps -p "${pid}" -o pid,ppid,user,%cpu,%mem,start,time,command
            
            # 显示最近日志
            local latest_log=$(ls -t "${LOG_DIR}/fastclaw-"*.combined.log 2>/dev/null | head -1)
            if [ -f "${latest_log}" ]; then
                echo ""
                echo "最近日志 (最后10行):"
                tail -10 "${latest_log}"
            fi
        else
            error "服务已停止 (PID文件存在但进程不存在)"
            warn "建议清理PID文件: rm -f ${PID_FILE}"
        fi
    else
        local pids=$(ps aux | grep -v grep | grep -E "(\./bin/fastclaw|fastclaw gateway)" | awk '{print $2}')
        if [ -n "${pids}" ]; then
            warn "发现FastClaw进程但没有PID文件:"
            ps aux | grep -v grep | grep -E "(\./bin/fastclaw|fastclaw gateway)"
        else
            info "服务未运行"
        fi
    fi
    
    # 显示配置信息
    if [ -f "${CONFIG_DIR}/fastclaw.json" ]; then
        echo ""
        echo "配置信息:"
        echo "  存储类型: $(grep -A1 '"storage"' "${CONFIG_DIR}/fastclaw.json" | grep '"type"' | cut -d'"' -f4 || echo '文件存储')"
        echo "  代理数量: $(grep -o '"id":"[^"]*"' "${CONFIG_DIR}/fastclaw.json" | wc -l | tr -d ' ')"
    fi
}

# 查看日志
show_logs() {
    heading "FastClaw 日志"
    
    if [ ! -d "${LOG_DIR}" ]; then
        error "日志目录不存在: ${LOG_DIR}"
        return 1
    fi
    
    local log_count=$(ls "${LOG_DIR}/"*.log 2>/dev/null | wc -l | tr -d ' ')
    
    if [ "${log_count}" -eq "0" ]; then
        warn "没有找到日志文件"
        return 0
    fi
    
    echo "日志文件 (${log_count}个):"
    ls -laht "${LOG_DIR}/"*.log 2>/dev/null | head -20
    
    local latest_combined=$(ls -t "${LOG_DIR}/fastclaw-"*.combined.log 2>/dev/null | head -1)
    if [ -f "${latest_combined}" ]; then
        echo ""
        echo -n "查看最新组合日志? [Y/n] "
        read -r confirm
        if [[ ! "${confirm}" =~ ^[Nn]$ ]]; then
            echo "=== ${latest_combined} ==="
            tail -50 "${latest_combined}"
        fi
    fi
}

# 清理旧日志
cleanup_old_logs() {
    local keep_days=7
    info "清理${keep_days}天前的日志..."
    
    find "${LOG_DIR}" -name "fastclaw-*.stdout.log" -mtime +${keep_days} -delete 2>/dev/null || true
    find "${LOG_DIR}" -name "fastclaw-*.stderr.log" -mtime +${keep_days} -delete 2>/dev/null || true
    find "${LOG_DIR}" -name "fastclaw-*.combined.log" -mtime +${keep_days} -delete 2>/dev/null || true
    
    info "日志清理完成"
}

# 一键部署（构建+安装+启动）
deploy_all() {
    heading "一键部署 FastClaw"
    
    # 检查是否需要构建
    if [ ! -f "${SOURCE_BIN}" ]; then
        info "二进制文件不存在，开始构建..."
        make_build
    else
        local build_time=$(stat -f "%m" "${SOURCE_BIN}" 2>/dev/null || stat -c "%Y" "${SOURCE_BIN}")
        local current_time=$(date +%s)
        local age_hours=$(( (current_time - build_time) / 3600 ))
        
        if [ ${age_hours} -gt 24 ]; then
            warn "二进制文件已超过24小时，建议重新构建"
            echo -n "是否重新构建? [y/N] "
            read -r confirm
            if [[ "${confirm}" =~ ^[Yy]$ ]]; then
                make_build
            fi
        fi
    fi
    
    # 安装到全局
    install_global
    
    # 启动服务
    echo ""
    echo -n "是否启动服务? [Y/n] "
    read -r confirm
    if [[ ! "${confirm}" =~ ^[Nn]$ ]]; then
        start_service "${DEFAULT_PORT}"
    fi
    
    success "部署完成"
}

# 构建项目
make_build() {
    heading "构建 FastClaw"
    
    if [ ! -f "${SCRIPT_DIR}/Makefile" ] && [ ! -f "${SCRIPT_DIR}/go.mod" ]; then
        error "未找到构建文件，请确认在FastClaw项目目录中"
        exit 1
    fi
    
    info "开始构建..."
    
    if [ -f "${SCRIPT_DIR}/Makefile" ]; then
        make build
    else
        go build -o "${SCRIPT_DIR}/bin/fastclaw" ./cmd/fastclaw
    fi
    
    if [ $? -eq 0 ] && [ -f "${SOURCE_BIN}" ]; then
        success "构建成功: ${SOURCE_BIN}"
    else
        error "构建失败"
        exit 1
    fi
}

# 显示帮助
show_help() {
    cat << EOF
${BOLD}FastClaw 开发版管理脚本${NC}

用法: $0 [命令] [选项]

命令:
  install     安装到全局 (~/.local/bin)
  start       启动服务 [端口，默认: ${DEFAULT_PORT}]
  stop        停止服务
  restart     重启服务
  status      查看服务状态
  logs        查看日志
  deploy      一键部署 (构建+安装+启动)
  build       重新构建项目
  clean       清理旧日志

选项:
  -h, --help  显示此帮助信息

示例:
  $0 install           # 安装到全局
  $0 start             # 启动服务 (端口${DEFAULT_PORT})
  $0 start 18954       # 启动服务 (指定端口)
  $0 stop              # 停止服务
  $0 status            # 查看状态
  $0 deploy            # 一键部署

工作流程:
  1. 完善代码
  2. make build        # 或使用 $0 build
  3. $0 deploy         # 一键部署
  4. fastclaw gateway  # 使用全局命令启动

代理配置: ${HTTP_PROXY}
日志目录: ${LOG_DIR}
EOF
}

# 主程序
main() {
    local command="${1}"
    local arg="${2}"
    
    case "${command}" in
        ""|-h|--help)
            show_help
            ;;
        install)
            install_global
            ;;
        start)
            start_service "${arg}"
            ;;
        stop)
            stop_service
            ;;
        restart)
            stop_service
            sleep 2
            start_service "${arg:-${DEFAULT_PORT}}"
            ;;
        status)
            check_status
            ;;
        logs)
            show_logs
            ;;
        deploy)
            deploy_all
            ;;
        build)
            make_build
            ;;
        clean)
            cleanup_old_logs
            ;;
        *)
            error "未知命令: ${command}"
            show_help
            exit 1
            ;;
    esac
}

# 运行主程序
main "$@"