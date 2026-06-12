#!/bin/bash
# =============================================================================
# claw-browser — 管理常驻的反检测浏览器容器(Podman)
# =============================================================================
# claw-os 跑在物理机,浏览器(camoufox-cli + Camoufox)跑在这个 Podman 容器里。
# claw-os 的 agent 通过 `podman exec claw-browser camoufox-cli ...` 驱动它。
#
# 用法:
#   ./claw-browser.sh build     # 构建镜像(首次,或改了 Dockerfile 后)
#   ./claw-browser.sh start     # 启动常驻容器
#   ./claw-browser.sh stop      # 停止并删除容器
#   ./claw-browser.sh restart   # stop + start
#   ./claw-browser.sh status    # 看容器是否在跑
#   ./claw-browser.sh test      # 冒烟测试:打开一个页面并 snapshot
#   ./claw-browser.sh login <平台>   # 打开 headed 浏览器让你手动登录(见下)
#   ./claw-browser.sh shell     # 进容器交互(调试用)
#   ./claw-browser.sh logs      # 看容器日志
#
# 环境变量:
#   CLAW_BROWSER_PROXY  传给容器的 HTTPS_PROXY(墙内访问境外站点时设,如 X)
#                       例: CLAW_BROWSER_PROXY=http://host.containers.internal:7890
# =============================================================================

set -e

IMAGE="claw-browser:latest"
CONTAINER="claw-browser"
SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
# 登录态 / fingerprint 持久化卷 —— 容器重建后登录不丢
PROFILES_VOLUME="claw-browser-profiles"

# ─── 颜色 ─────────────────────────────────────────────────────────────────────
if [ -t 1 ]; then
    RED='\033[0;31m'; GREEN='\033[0;32m'; YELLOW='\033[1;33m'; BOLD='\033[1m'; NC='\033[0m'
else
    RED=''; GREEN=''; YELLOW=''; BOLD=''; NC=''
fi
info()    { printf "${GREEN}[INFO]${NC} %s\n" "$*"; }
warn()    { printf "${YELLOW}[WARN]${NC} %s\n" "$*"; }
error()   { printf "${RED}[ERROR]${NC} %s\n" "$*" >&2; }
ok()      { printf "${GREEN}[✓]${NC} %s\n" "$*"; }

require_podman() {
    if ! command -v podman >/dev/null 2>&1; then
        error "未找到 podman。安装: brew install podman (macOS) / sudo apt install podman (Linux)"
        error "macOS 还需初始化虚拟机: podman machine init && podman machine start"
        exit 1
    fi
    if ! podman info >/dev/null 2>&1; then
        error "podman 不可用。macOS 上先: podman machine start"
        exit 1
    fi
}

is_running() {
    [ "$(podman inspect -f '{{.State.Running}}' "$CONTAINER" 2>/dev/null)" = "true" ]
}

# ─── 命令 ─────────────────────────────────────────────────────────────────────

cmd_build() {
    require_podman
    info "构建镜像 ${IMAGE}(首次会拉 Camoufox,几分钟,耐心等)..."
    podman build -t "$IMAGE" "$SCRIPT_DIR"
    ok "镜像构建完成: ${IMAGE}"
}

cmd_start() {
    require_podman
    if is_running; then
        warn "容器 ${CONTAINER} 已在运行。"
        return 0
    fi
    podman rm "$CONTAINER" >/dev/null 2>&1 || true

    if ! podman image exists "$IMAGE"; then
        error "镜像 ${IMAGE} 不存在,请先运行: $0 build"
        exit 1
    fi

    local proxy_env=""
    if [ -n "${CLAW_BROWSER_PROXY:-}" ]; then
        proxy_env="-e HTTPS_PROXY=${CLAW_BROWSER_PROXY} -e HTTP_PROXY=${CLAW_BROWSER_PROXY}"
        info "代理已设置: ${CLAW_BROWSER_PROXY}"
    fi

    info "启动常驻容器 ${CONTAINER}..."
    # --shm-size=1g: Firefox 在默认 64MB /dev/shm 下渲染大页面会频繁崩。
    # -v <vol>:/profiles: 持久化登录态/fingerprint,容器重建不丢登录。
    # --restart unless-stopped: podman machine / systemd 重启后自动拉起。
    # shellcheck disable=SC2086
    podman run -d \
        --name "$CONTAINER" \
        --restart unless-stopped \
        --shm-size=1g \
        -v "${PROFILES_VOLUME}:/profiles" \
        $proxy_env \
        "$IMAGE" >/dev/null
    ok "容器已启动。第一次 camoufox-cli open 会冷启动 2-3 分钟,属正常。"
    info "测试: $0 test    |    需要登录的平台先: $0 login <平台>"
}

cmd_stop() {
    require_podman
    info "停止并删除容器 ${CONTAINER}(登录态保存在卷里,不丢)..."
    podman stop "$CONTAINER" >/dev/null 2>&1 || true
    podman rm "$CONTAINER" >/dev/null 2>&1 || true
    ok "已停止。"
}

cmd_restart() {
    cmd_stop
    cmd_start
}

cmd_status() {
    require_podman
    if is_running; then
        ok "容器 ${CONTAINER} 运行中。"
        podman ps --filter "name=${CONTAINER}" --format "  {{.Names}}  {{.Status}}  {{.Image}}"
    else
        warn "容器 ${CONTAINER} 未运行。启动: $0 start"
    fi
}

cmd_test() {
    require_podman
    is_running || { error "容器未运行,先 $0 start"; exit 1; }
    info "冒烟测试:打开 example.com 并抓 snapshot(首次冷启动需 2-3 分钟)..."
    podman exec "$CONTAINER" camoufox-cli open https://example.com
    podman exec "$CONTAINER" camoufox-cli snapshot -i | head -20
    ok "浏览器工作正常。"
    info "claw-os 的 agent 现在可以用: podman exec ${CONTAINER} camoufox-cli <命令>"
}

# 手动登录:开一个带持久 profile 的 headed 浏览器,你自己扫码/输密码登录一次,
# 登录态存进卷,之后 agent 用同名 --persistent profile 复用,无需密码。
#
# 注意: headed 模式要看到窗口,需要容器能连到显示。容器内默认无显示,所以这里
# 的实操通常是:在你本机(非容器)装 camoufox-cli 登录一次导出 cookies,或用
# 远程调试。详见 deploy/browser/README.md 的"登录态"章节 —— 这条命令打印指引。
cmd_login() {
    local platform="${2:-}"
    case "$platform" in
        xhs|xiaohongshu) url="https://www.xiaohongshu.com" ; prof="xhs" ;;
        douyin|dy)       url="https://www.douyin.com"       ; prof="douyin" ;;
        x|twitter)       url="https://x.com"                ; prof="x" ;;
        *) error "用法: $0 login <xhs|douyin|x>"; exit 1 ;;
    esac
    cat <<EOF
${BOLD}手动登录 ${platform}${NC}

容器内浏览器无头,无法直接弹窗给你扫码。推荐两种方式登录到持久 profile
(profile 名: ${prof},存在卷 ${PROFILES_VOLUME}):

方式一(推荐,本机登录后导 cookie 进容器):
  1. 你本机浏览器正常登录 ${url}
  2. 用浏览器插件(如 EditThisCookie)导出该站 cookies 为 cookies.json
  3. podman cp cookies.json ${CONTAINER}:/tmp/cookies.json
  4. podman exec ${CONTAINER} camoufox-cli --persistent /profiles/${prof} open ${url}
     podman exec ${CONTAINER} camoufox-cli --persistent /profiles/${prof} cookies import /tmp/cookies.json
     podman exec ${CONTAINER} camoufox-cli --persistent /profiles/${prof} reload

方式二(容器内 headed + VNC,较折腾):
  见 deploy/browser/README.md 的"登录态"章节。

登录后验证:
  podman exec ${CONTAINER} camoufox-cli --persistent /profiles/${prof} open ${url}
  podman exec ${CONTAINER} camoufox-cli --persistent /profiles/${prof} snapshot -i
  能看到登录后才有的内容 = 成功。
EOF
}

cmd_shell() { require_podman; podman exec -it "$CONTAINER" /bin/bash; }
cmd_logs()  { require_podman; podman logs -f "$CONTAINER"; }

usage() { sed -n '2,30p' "$0" | sed 's/^# \{0,1\}//'; }

case "${1:-}" in
    build)   cmd_build ;;
    start)   cmd_start ;;
    stop)    cmd_stop ;;
    restart) cmd_restart ;;
    status)  cmd_status ;;
    test)    cmd_test ;;
    login)   cmd_login "$@" ;;
    shell)   cmd_shell ;;
    logs)    cmd_logs ;;
    *)       usage ;;
esac
