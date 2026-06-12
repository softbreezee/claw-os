# claw-browser — 浏览器容器启动流程(Podman)

claw-os 跑在物理机上,浏览器(camoufox-cli + 反检测 Firefox)跑在一个常驻的
Podman 容器 `claw-browser` 里。claw-os 的 agent 通过
`podman exec claw-browser camoufox-cli ...` 驱动它来爬数据 / 追踪博主。

本文是从零到能用的完整流程。

---

## 0. 为什么用容器(而不是物理机直装)

camoufox-cli 要装一个完整 Firefox + 一堆系统依赖,还会常驻浏览器 daemon、执行
不可信的网页内容。把它关进容器:物理机保持干净、出问题 `podman rm` 一键清、
有隔离更安全。claw-os 主进程**不进容器**,还是物理机直跑。

---

## 1. 装 Podman(一次性)

**macOS**:
```bash
brew install podman
podman machine init        # 初始化虚拟机(Podman 在 mac 上跑在轻量 VM 里)
podman machine start
```

**Linux**:
```bash
sudo apt install podman    # Debian/Ubuntu
# 或 sudo dnf install podman (Fedora)
```

验证:`podman info` 不报错即可。

---

## 2. 构建镜像(一次性,或改了 Dockerfile 后)

```bash
cd deploy/browser
./claw-browser.sh build
```

首次会拉 Camoufox(反检测 Firefox)+ 一堆依赖,**几分钟**,耐心等。看到
`[✓] 镜像构建完成` 即可。

---

## 3. 启动常驻容器

```bash
./claw-browser.sh start
```

- 容器名 `claw-browser`,`--restart unless-stopped`,podman/机器重启后自动拉起
- 登录态存在 podman volume `claw-browser-profiles`(挂到容器 `/profiles`),
  容器重建不丢登录
- `--shm-size=1g`:Firefox 渲染大页面需要,默认 64MB 会崩

**墙外站点(X / Twitter)**:启动前设代理:
```bash
CLAW_BROWSER_PROXY=http://host.containers.internal:7890 ./claw-browser.sh start
```
(`host.containers.internal` 是 podman 容器访问宿主机的特殊主机名;7890 换成你
本机代理端口)

---

## 4. 冒烟测试

```bash
./claw-browser.sh test
```

它会 `open example.com` + `snapshot`。**首次 open 冷启动 2-3 分钟**(起 Firefox
daemon),正常。看到 snapshot 输出一堆元素 = 浏览器工作正常。

---

## 5. 登录需要登录态的平台(按需)

公开主页/搜索通常不用登录。但小红书关注流、抖音完整浏览、**X(几乎必须)**
需要登录。

```bash
./claw-browser.sh login xhs      # 或 douyin / x
```

它会打印登录指引。推荐方式:**本机浏览器登录后导出 cookie 进容器**(因为容器内
浏览器无头、无法弹窗扫码):
1. 你本机浏览器正常登录该平台
2. 用浏览器插件(如 EditThisCookie)导出该站 cookies.json
3. `podman cp cookies.json claw-browser:/tmp/cookies.json`
4. `podman exec claw-browser camoufox-cli --persistent /profiles/xhs open <平台URL>`
   `podman exec claw-browser camoufox-cli --persistent /profiles/xhs cookies import /tmp/cookies.json`
   `podman exec claw-browser camoufox-cli --persistent /profiles/xhs reload`

登录态存进 volume,以后 agent 加 `--persistent /profiles/<平台>` 复用,无需密码。

> **安全**:agent 永远不碰你的密码,只复用你手动登录好的 session。登录这一步
> 永远是你手动做的。

---

## 6. claw-os 这边:什么都不用改

agent 通过现有的 `exec` 工具调 `podman exec claw-browser camoufox-cli ...`,
不需要改 claw-os 代码。相关 skill 已经写好:
- `skills/browser/` —— 浏览器操作通用能力(camoufox-cli 用法)
- `skills/track-common/` —— 追踪/监控通用骨架(去重、筛选、cron)
- `skills/track-xhs/` `skills/track-douyin/` `skills/track-x/` —— 各平台差异

agent 遇到"追踪某博主/搜话题"时会自动加载对应 skill。

---

## 7. 日常管理

```bash
./claw-browser.sh status     # 看容器在不在跑
./claw-browser.sh restart    # 重启(改了代理 / 卡住了)
./claw-browser.sh stop       # 停(登录态保留在 volume,不丢)
./claw-browser.sh logs       # 看容器日志
./claw-browser.sh shell      # 进容器调试
```

---

## 常见问题

| 现象 | 原因 / 解决 |
|---|---|
| `command not found: podman` | 没装,见第 1 步 |
| `podman info` 报错(mac) | `podman machine start` |
| 首次 open 超时被杀 | agent 第一次 open 要传 `timeout: 280`(skill 已写明) |
| `NS_ERROR_NET_INTERRUPT`(X) | 代理没配,见第 3 步 `CLAW_BROWSER_PROXY` |
| 渲染崩 / 白屏 | shm 不够,确认 start 带了 `--shm-size=1g`(脚本默认有) |
| 小红书/抖音要求滑块验证 | 风控触发,降低频率;别自动过验证,手动处理一次 |
| 登录态过期 | 重新走第 5 步 login |

---

## 一句话流程

```
装 podman → ./claw-browser.sh build → start → test
                                              ↓
                          (需登录的平台) login <平台>
                                              ↓
              claw-os agent 自动用 skill 调 podman exec 爬数据
```
