#!/bin/bash
# Pawnix 开发版管理脚本
#
# 设计原则（重构后）：
#   - 进程管理（start/stop/restart/status）全部委托给 `bin/pawnix daemon`，
#     manager.sh 不再维护自己的 PID 文件、不再 nohup tee。
#   - 单一事实来源：PID = ~/.pawnix/pawnix.pid，日志 = ~/.pawnix/logs/gateway.log。
#     这样无论用 `./pawnix-manager.sh` 还是 `bin/pawnix daemon` 启动，
#     另一种命令都能正确感知/管理。
#   - manager.sh 只在 daemon 命令之外提供：开发便利（install/build/deploy）、
#     代理 env 注入（HTTP_PROXY/HTTPS_PROXY）、彩色日志输出和富文本 status。
#
# 用法：./pawnix-manager.sh [command] [options]
# 命令：install / start / stop / restart / status / logs / deploy / build / clean

set -e

# ─── 配置 ─────────────────────────────────────────────────────────────────────

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
SOURCE_BIN="${SCRIPT_DIR}/bin/pawnix"
CONFIG_DIR="${HOME}/.pawnix"
DAEMON_LOG="${CONFIG_DIR}/logs/gateway.log"   # 与 bin/pawnix daemon 一致
DAEMON_PID="${CONFIG_DIR}/pawnix.pid"       # 与 internal/daemon.Paths() 一致

# 全局安装配置
INSTALL_DIR="${HOME}/.local/bin"
GLOBAL_BIN_NAME="pawnix"

# 服务配置
DEFAULT_PORT=18953

# 代理配置（HTTP_PROXY/HTTPS_PROXY 为空时不导出，留给系统代理或不走代理）
# HTTP_PROXY="http://localhost:7892"
# HTTPS_PROXY="http://localhost:7892"

# ─── 输出辅助 ────────────────────────────────────────────────────────────────

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

# ─── 公共检查 ────────────────────────────────────────────────────────────────

check_source_binary() {
    if [ ! -f "${SOURCE_BIN}" ]; then
        error "Pawnix 二进制文件不存在: ${SOURCE_BIN}"
        error "请先运行: $0 build"
        exit 1
    fi
}

# 注入代理 env，让后续 fork 出来的 daemon 子进程继承。
# 注意：用 if 块而不是 `[ -n "X" ] && ...` 链式写法，否则在 set -e 下
# 当变量为空时整个 && 链返回非零，会让脚本提前退出（被坑过一次）。
inject_proxy_env() {
    if [ -n "${HTTP_PROXY}" ]; then
        export http_proxy="${HTTP_PROXY}"
        export HTTP_PROXY="${HTTP_PROXY}"
    fi
    if [ -n "${HTTPS_PROXY}" ]; then
        export https_proxy="${HTTPS_PROXY}"
        export HTTPS_PROXY="${HTTPS_PROXY}"
    fi
}

# ─── install / build / deploy ────────────────────────────────────────────────

install_global() {
    heading "安装 Pawnix 到全局"
    check_source_binary
    mkdir -p "${INSTALL_DIR}"
    install -m 0755 "${SOURCE_BIN}" "${INSTALL_DIR}/${GLOBAL_BIN_NAME}"
    success "已安装: ${INSTALL_DIR}/${GLOBAL_BIN_NAME}"
    case ":$PATH:" in
        *":${INSTALL_DIR}:"*) ;;
        *) warn "${INSTALL_DIR} 不在 PATH 中。建议添加到 ~/.zshrc 或 ~/.bashrc："
           echo "    export PATH=\"${INSTALL_DIR}:\$PATH\"" ;;
    esac
}

make_build() {
    heading "构建 Pawnix"
    if [ ! -f "${SCRIPT_DIR}/Makefile" ] && [ ! -f "${SCRIPT_DIR}/go.mod" ]; then
        error "未找到构建文件，请确认在 Pawnix 项目目录中"
        exit 1
    fi
    info "开始构建（make build）..."
    if [ -f "${SCRIPT_DIR}/Makefile" ]; then
        ( cd "${SCRIPT_DIR}" && make build )
    else
        ( cd "${SCRIPT_DIR}" && go build -o "${SOURCE_BIN}" ./cmd/pawnix )
    fi
    if [ -f "${SOURCE_BIN}" ]; then
        success "构建成功: ${SOURCE_BIN}"
    else
        error "构建失败"
        exit 1
    fi
}

deploy_all() {
    heading "一键部署 Pawnix（build → install → restart）"
    make_build
    install_global
    echo ""
    restart_service "${DEFAULT_PORT}"
    success "部署完成"
}

# ─── 服务管理（thin wrapper around `bin/pawnix daemon`）───────────────────

start_service() {
    local port=${1:-${DEFAULT_PORT}}
    heading "启动 Pawnix 服务"

    check_source_binary
    inject_proxy_env

    info "启动参数:"
    info "  端口: ${port}"
    if [ -n "${HTTP_PROXY}" ]; then
        info "  代理: ${HTTP_PROXY}"
    fi
    info "  PID 文件: ${DAEMON_PID}"
    info "  日志文件: ${DAEMON_LOG}"
    echo ""

    # 委托给 daemon start。--port 在 cmd_daemon 接受作为子进程参数。
    "${SOURCE_BIN}" daemon start --port "${port}"

    echo ""
    echo "访问地址: ${BOLD}http://localhost:${port}${NC}"
    echo "查看日志: ${BOLD}$0 logs${NC}     (或 tail -f ${DAEMON_LOG})"
    echo "停止服务: ${BOLD}$0 stop${NC}"
}

stop_service() {
    heading "停止 Pawnix 服务"
    check_source_binary
    "${SOURCE_BIN}" daemon stop
}

restart_service() {
    local port=${1:-${DEFAULT_PORT}}
    heading "重启 Pawnix 服务"
    check_source_binary
    inject_proxy_env
    "${SOURCE_BIN}" daemon restart --port "${port}"
    echo ""
    echo "访问地址: ${BOLD}http://localhost:${port}${NC}"
}

check_status() {
    heading "Pawnix 服务状态"
    check_source_binary

    echo "二进制文件: ${SOURCE_BIN}"
    echo "PID 文件:   ${DAEMON_PID}"
    echo "日志文件:   ${DAEMON_LOG}"
    echo ""

    # 委托给 daemon status，这是权威来源
    "${SOURCE_BIN}" daemon status || true
    echo ""

    # 叠加 manager 自己的端口/进程详情
    local port_info
    port_info=$(lsof -nP -iTCP:${DEFAULT_PORT} -sTCP:LISTEN 2>/dev/null | tail -n +2 | head -1)
    if [ -n "${port_info}" ]; then
        echo "端口 ${DEFAULT_PORT} 监听:"
        echo "  ${port_info}"
    fi

    # 配置摘要
    if [ -f "${CONFIG_DIR}/pawnix.json" ]; then
        echo ""
        echo "配置摘要:"
        local store_type
        store_type=$(grep -A1 '"storage"' "${CONFIG_DIR}/pawnix.json" 2>/dev/null \
                     | grep '"type"' | head -1 | cut -d'"' -f4)
        echo "  存储类型: ${store_type:-file}"
    fi

    # 最近日志 (最后 10 行)
    if [ -f "${DAEMON_LOG}" ]; then
        echo ""
        echo "最近日志 (${DAEMON_LOG} 末 10 行):"
        tail -10 "${DAEMON_LOG}" | sed 's/^/  /'
    fi
}

show_logs() {
    heading "Pawnix 日志"
    if [ ! -f "${DAEMON_LOG}" ]; then
        error "日志文件不存在: ${DAEMON_LOG}"
        warn "可能服务尚未启动，先运行: $0 start"
        return 1
    fi
    info "tail -f ${DAEMON_LOG}（Ctrl+C 退出）"
    echo ""
    tail -f "${DAEMON_LOG}"
}

# ─── 清理 ─────────────────────────────────────────────────────────────────────

cleanup_old_logs() {
    heading "清理旧产物"

    # daemon 命令的 gateway.log 是单文件，不需要按日清理。
    # 这里清理的是历史遗留：旧版 manager.sh 在 <repo>/logs/ 下按日生成的文件，
    # 以及 <repo>/pawnix.pid 这个废弃的 PID 文件。
    local stale_pid="${SCRIPT_DIR}/pawnix.pid"
    local stale_logdir="${SCRIPT_DIR}/logs"

    if [ -f "${stale_pid}" ]; then
        info "移除旧 PID 文件: ${stale_pid}"
        rm -f "${stale_pid}"
    fi

    if [ -d "${stale_logdir}" ]; then
        local count
        count=$(ls -1 "${stale_logdir}"/pawnix-*.log 2>/dev/null | wc -l | tr -d ' ')
        if [ "${count}" != "0" ]; then
            info "移除旧版按日志文件: ${stale_logdir}/pawnix-*.log (${count} 个)"
            find "${stale_logdir}" -maxdepth 1 -name "pawnix-*.log" -delete 2>/dev/null || true
        fi
        # 如果目录已空就一并删掉
        if [ -z "$(ls -A "${stale_logdir}" 2>/dev/null)" ]; then
            rmdir "${stale_logdir}" 2>/dev/null || true
        fi
    fi

    success "清理完成"
}

# ─── 帮助 + 入口 ─────────────────────────────────────────────────────────────

show_help() {
    cat << EOF
${BOLD}Pawnix 开发版管理脚本${NC}

用法: $0 [命令] [选项]

命令:
  install     安装到全局 (~/.local/bin/pawnix)
  build       重新构建项目 (等价于 make build)
  deploy      一键部署 (build → install → restart)

  start       启动服务 [端口，默认: ${DEFAULT_PORT}]    （委托给 pawnix daemon start）
  stop        停止服务                              （委托给 pawnix daemon stop）
  restart     重启服务 [端口，默认: ${DEFAULT_PORT}]    （委托给 pawnix daemon restart）
  status      查看服务状态 + 端口/配置/日志摘要      （委托给 pawnix daemon status）
  logs        tail -f 实时日志

  clean       清理历史遗留 (<repo>/pawnix.pid 和 <repo>/logs/pawnix-*.log)

选项:
  -h, --help  显示此帮助信息

工作流程:
  1. 改代码
  2. ${BOLD}$0 deploy${NC}    # 构建 + 安装 + 重启
  3. ${BOLD}$0 logs${NC}      # 查看日志

PID 文件: ${DAEMON_PID}
日志文件: ${DAEMON_LOG}
EOF
}

main() {
    local command="${1:-}"
    local arg="${2:-}"

    case "${command}" in
        ""|-h|--help)   show_help ;;
        install)        install_global ;;
        build)          make_build ;;
        deploy)         deploy_all ;;
        start)          start_service "${arg}" ;;
        stop)           stop_service ;;
        restart)        restart_service "${arg}" ;;
        status)         check_status ;;
        logs)           show_logs ;;
        clean)          cleanup_old_logs ;;
        *)              error "未知命令: ${command}"; show_help; exit 1 ;;
    esac
}

main "$@"
