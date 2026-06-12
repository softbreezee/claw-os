# claw-browser 服务器部署 SOP

在一台全新服务器上,从零把反检测浏览器容器跑通,接入 claw-os。
照着每一步做,**每步都有验证命令 + 预期输出**,验证不通过不要进下一步。

> 适用:claw-os 跑在物理机/服务器上(非容器),浏览器单独跑 Podman 容器。
> 已在 macOS 验证;Linux 服务器步骤基本一致,差异处单独标注。

---

## 前置检查

```bash
uname -a                 # 确认 OS / 架构
nproc                    # CPU 核数
free -m                  # 物理内存(给浏览器 VM 留 ≥4G)
```

要求:服务器物理内存 ≥ 6G(podman machine VM 要分 4G 给浏览器)。

---

## 步骤 1:安装 Podman

**macOS:**
```bash
brew install podman
podman machine init
podman machine start
```

**Linux(Debian/Ubuntu):**
```bash
sudo apt update && sudo apt install -y podman
# Linux 上 podman 直接跑在宿主机,没有 machine 这层,跳过所有 machine 相关命令
```

**Linux(Fedora/RHEL):**
```bash
sudo dnf install -y podman
```

**验证:**
```bash
podman info
```
预期:输出一大段配置,不报错。
- macOS 若报错 → `podman machine start`
- 报 `Cannot connect` → machine 没起来

---

## 步骤 2:给浏览器 VM 分足够内存(关键,macOS)

> ⚠️ 最容易跳过、然后在"截图重页面崩溃"时才发现。Camoufox 渲染东方财富/
> 小红书这类重页面,2G 默认内存必崩(`Connection closed while reading from
> the driver` = Firefox OOM)。

**macOS:**
```bash
podman machine stop
podman machine set --memory 4096      # 4G;机器内存充裕给 6144
podman machine start
```

**验证:**
```bash
podman machine inspect --format '{{.Resources.Memory}}'
```
预期:`4096`(或你设的值)。

**Linux:** 用宿主机内存,无需此步;确保宿主机 free ≥ 4G。

---

## 步骤 3:解决镜像拉取(DNS 污染 / 墙)

> 国内服务器常见:`docker.io` 被 DNS 污染或墙,build 第一步拉
> `node:22-bookworm-slim` 就失败(`name resolution` / `dial tcp` 错误)。

**先测能不能直连:**
```bash
podman pull node:22-bookworm-slim
```
- 成功 → 跳过本步,去步骤 4
- 失败(name resolution / timeout)→ 往下配镜像源或代理

### 3a. 配国内镜像源(无代理时)

```bash
# macOS 进 VM;Linux 直接在宿主机执行(去掉 machine ssh / exit 两行)
podman machine ssh

sudo tee /etc/containers/registries.conf.d/999-mirror.conf > /dev/null <<'EOF'
[[registry]]
location = "docker.io"

[[registry.mirror]]
location = "docker.m.daocloud.io"
EOF

exit
podman machine stop && podman machine start
```
mirror 时好时坏,daocloud 不通时换:`dockerproxy.com` / `docker.1ms.run`。

### 3b. 配代理(有科学上网时,推荐)

代理治本——镜像、Camoufox 本体、以后访问 X 全走代理。
```bash
podman machine ssh
# 把 7890 换成你的代理端口;host.containers.internal = 宿主机
sudo mkdir -p /etc/systemd/system/podman.service.d
sudo tee /etc/systemd/system/podman.service.d/proxy.conf > /dev/null <<'EOF'
[Service]
Environment="HTTPS_PROXY=http://host.containers.internal:7890"
Environment="HTTP_PROXY=http://host.containers.internal:7890"
EOF
exit
podman machine stop && podman machine start
```

**验证(任选 a/b 后):**
```bash
podman pull node:22-bookworm-slim
```
预期:能拉下来。拉不下来不要进步骤 4。

---

## 步骤 4:构建镜像

```bash
cd <claw-os>/deploy/browser
chmod +x claw-browser.sh          # ⚠️ 新 clone 的仓库脚本没有执行权限
./claw-browser.sh build
```
首次拉 Camoufox(反检测 Firefox)+ 依赖,**几分钟**。

**验证:** 看到 `[✓] 镜像构建完成`。
- `permission denied` → 忘了 `chmod +x`
- `camoufox-cli install` 那步失败 → Camoufox 本体在墙外,需要步骤 3b 的代理

---

## 步骤 5:启动容器

```bash
./claw-browser.sh start
```
- 访问境外站(X)需代理:`CLAW_BROWSER_PROXY=http://host.containers.internal:7890 ./claw-browser.sh start`

**验证:**
```bash
./claw-browser.sh status
```
预期:`[✓] 容器 claw-browser 运行中`。

---

## 步骤 6:冒烟测试

```bash
./claw-browser.sh test
```
**首次冷启动 2-3 分钟**(起 Firefox daemon),正常。

**验证:** 输出里有 `Example Domain` + `link "Learn more" [ref=e1]` + `[✓] 浏览器工作正常`。

---

## 步骤 7:重页面验证(确认内存够)

冒烟用的 example.com 极轻,过了不代表重页面也行。必须测一个真实重页面:
```bash
podman exec claw-browser camoufox-cli open "https://quote.eastmoney.com/sz002812.html"
podman exec claw-browser camoufox-cli url          # 浏览器没崩才拿得到 URL
podman exec claw-browser camoufox-cli wait 5000
podman exec claw-browser camoufox-cli screenshot /workspace/test.png
podman exec claw-browser ls -la /workspace/test.png
podman cp claw-browser:/workspace/test.png ./test.png
```

**验证:** `test.png` 生成且能打开看到行情页。
- `Connection closed while reading from the driver` = Firefox OOM →
  回步骤 2 加内存(4G→6G),`./claw-browser.sh restart` 重试
- 页面"没加载完"= 正常,异步内容慢;agent 靠 snapshot 读数据,不靠截图,
  不影响。要完整截图就 `wait` 更久。

诊断命令(崩溃时用):
```bash
podman exec claw-browser sh -c "df -h /dev/shm && free -m"   # shm 应 1G,看 Mem free
podman logs --tail 30 claw-browser                           # 看 crash 痕迹
```

---

## 步骤 8:接入 claw-os

浏览器容器和 claw-os 是两个独立进程,claw-os 通过 `podman exec` 调容器。

```bash
cd <claw-os>
./pawnix-manager.sh deploy        # 重新加载,让 agent 看到 browser / track-* skill
```
Web UI 打开后 **Cmd+Shift+R 强制刷新**(清前端缓存)。

相关 skill 已随仓库:
- `skills/browser/` — camoufox-cli 通用操作
- `skills/track-common/` — 追踪通用骨架(去重/筛选/cron)
- `skills/track-xhs/` `track-douyin/` `track-x/` — 各平台

**验证:** 回聊天界面发:
```
帮我用浏览器打开 https://quote.eastmoney.com/sz002812.html，
等加载完，告诉我恩捷股份现在的价格和涨跌幅
```
预期:agent 自动加载 browser skill、调 `podman exec claw-browser camoufox-cli`、
snapshot 读出价格。

> ⚠️ 容器冷启动时 agent 第一条 open 要 `timeout: 280`(skill 已写)。弱模型
> 可能忘 → 你先手动 `camoufox-cli open <url>` 预热,或提醒它设 timeout。

---

## 步骤 9:登录态(按需,小红书/抖音/X)

公开内容不用登录。需要登录的(X 几乎必须),用持久 profile,**绝不让 agent 输密码**:
```bash
./claw-browser.sh login xhs       # 打印登录指引(本机登录→导 cookie→import)
```
登录态存 podman volume `claw-browser-profiles`,容器重建不丢。之后 agent 命令
加 `--persistent /profiles/xhs` 复用。

---

## 日常运维

```bash
./claw-browser.sh status / restart / stop / logs / shell
```
- 改了 Dockerfile → 重新 `build`
- 容器卡住 / 改了代理 → `restart`
- VM 重启后容器自动拉起(`--restart unless-stopped`),但第一次 open 又会冷启动

---

## 排错速查

| 现象 | 原因 | 解决 |
|---|---|---|
| `permission denied: ./claw-browser.sh` | 脚本无执行权限 | `chmod +x claw-browser.sh` |
| build 第一步 `name resolution` 失败 | docker.io 被污染/墙 | 步骤 3 配镜像源或代理 |
| `Connection closed while reading from the driver` | Firefox OOM | 步骤 2 加 VM 内存到 4-6G |
| 渲染白屏/崩 | shm 不够 | 确认 start 带 `--shm-size=1g`(脚本默认有) |
| `camoufox-cli install` 失败 | Camoufox 本体在墙外 | 步骤 3b 配代理 |
| 首次 open 超时 | 冷启动 2-3 分钟被默认 30s 杀 | 第一条 open 设 `timeout: 280` |
| X 报 `NS_ERROR_NET_INTERRUPT` | 没配代理 | `CLAW_BROWSER_PROXY=... ./claw-browser.sh restart` |
| 小红书/抖音要滑块验证 | 风控触发 | 降频;别自动过验证,手动处理一次 |
| agent 不用浏览器 | claw-os 没加载新 skill | `./pawnix-manager.sh deploy` + 浏览器硬刷新 |

---

## 一页纸流程

```
装 podman → (mac) machine set --memory 4096 → 测 podman pull
   → 不通则配镜像源/代理(步骤 3)
   → chmod +x + ./claw-browser.sh build
   → start → test(冒烟)→ 重页面验证(步骤 7,确认不 OOM)
   → claw-os: ./pawnix-manager.sh deploy + 硬刷新
   → 聊天里让 agent 打开东方财富验证
   → (按需) login <平台> 配登录态
```
