# camoufox-cli 完整命令参考(claw-browser 容器版)

所有命令都通过 `podman exec claw-browser` 前缀执行。浏览器跑在容器里,
claw-os 在物理机上通过 exec 工具驱动它。

> 为简洁,下面命令省略前缀,只写 `camoufox-cli ...`。**实际调用时务必
> 写成** `podman exec claw-browser camoufox-cli ...`。

**Related**: [snapshot-refs.md](snapshot-refs.md) — ref 系统详解。

## 命令链式调用

命令可以用 `&&` 在一次 exec 调用里串起来。浏览器通过后台 daemon 在命令间
保持状态,所以串联是安全且更高效的(省 exec 往返)。

```bash
# 串联 open + snapshot
podman exec claw-browser camoufox-cli open https://example.com && \
podman exec claw-browser camoufox-cli snapshot -i
```

**何时串联**:当你不需要读中间命令的输出就能继续时(如 open + screenshot)。
当你需要先解析输出再决定下一步时分开调用(如 snapshot 拿到 refs 后再交互)。

## 导航

```bash
camoufox-cli open <url>       # 打开 URL(必要时启动 daemon)
camoufox-cli back             # 后退
camoufox-cli forward          # 前进
camoufox-cli reload           # 刷新
camoufox-cli url              # 打印当前 URL
camoufox-cli title            # 打印页面标题
camoufox-cli close            # 关闭浏览器 + 停 daemon
camoufox-cli close --all      # 关闭所有 session
```

## 快照(Snapshot)

```bash
camoufox-cli snapshot                 # 完整 aria 树
camoufox-cli snapshot -i              # 仅交互元素(推荐)
camoufox-cli snapshot -s "#selector"  # 限定到 CSS selector 范围
```

## 交互(用 snapshot 拿到的 @refs)

```bash
camoufox-cli click @e1            # 点击
camoufox-cli fill @e1 "text"      # 清空后输入
camoufox-cli type @e1 "text"      # 不清空,追加输入
camoufox-cli select @e1 "option"  # 选下拉项
camoufox-cli check @e1            # 勾选/取消勾选
camoufox-cli hover @e1            # 悬停
camoufox-cli press Enter          # 按键
camoufox-cli press "Control+a"    # 组合键
```

## 数据提取

```bash
camoufox-cli text @e1                # 取某元素文本
camoufox-cli text body               # 取整页文本(CSS selector;SPA 上不可靠)
camoufox-cli eval "document.title"   # 执行 JS
```

## 截图 / 导出

```bash
camoufox-cli screenshot              # 截图返回 JSON {"base64": "..."}
camoufox-cli screenshot page.png     # 截图存文件
camoufox-cli screenshot --full p.png # 整页截图
camoufox-cli pdf output.pdf          # 存为 PDF
```

> 截图存文件时注意:文件在容器内。要取出来用
> `podman cp claw-browser:/workspace/page.png ./page.png`,或截图存到
> 一个挂载的 volume(见 deploy/browser 配置)。

## 滚动 / 等待

```bash
camoufox-cli scroll down             # 下滚 500px
camoufox-cli scroll up               # 上滚 500px
camoufox-cli scroll down 1000        # 下滚 1000px
camoufox-cli wait @e1                # 等元素出现
camoufox-cli wait 2000               # 等毫秒数
camoufox-cli wait --url "*/dashboard" # 等 URL 匹配
```

## 标签页

```bash
camoufox-cli tabs                    # 列出标签
camoufox-cli switch 2                # 按索引切换
camoufox-cli close-tab               # 关闭当前标签
```

## Cookie / 状态

```bash
camoufox-cli cookies                 # dump cookies(JSON)
camoufox-cli cookies import file.json # 导入 cookies
camoufox-cli cookies export file.json # 导出 cookies
```

## Session(并发隔离)

并发跑多个任务时,用命名 session 避免互相踩:

```bash
camoufox-cli --session s1 open https://site-a.com
camoufox-cli --session s2 open https://site-b.com
camoufox-cli sessions                # 列出活跃 session
camoufox-cli --session s1 snapshot -i
camoufox-cli --session s1 close      # 关闭指定 session
camoufox-cli close --all             # 关闭所有
```

## 全局标志

```
--session <name>       命名 session(默认 "default")
--headed               显示浏览器窗口(默认无头;容器内无头通常无显示,
                       调试时用 claw-browser.sh shell 进容器)
--timeout <seconds>    daemon 空闲超时(默认 1800)
--json                 输出 JSON 而非人类可读
--persistent [path]    持久身份 — 跨启动复用同一 fingerprint + cookies
--proxy <url>          代理(http:// 或 https://;认证: http://user:pass@host:port)
--no-geoip             禁用自动 GeoIP 伪装(--proxy 时自动开启)
--locale <tag>         强制 locale(如 "en-US" 或 "zh-CN")
```

## 持久身份(账号绑定 / 登录态)

默认每次启动用全新随机 fingerprint。加 `--persistent [path]` 跨启动复用同一
fingerprint + cookies。

**何时用**:同一设备需要在多次访问间呈现同一指纹(账号绑定任务、需要登录态的
小红书关注流、或单靠 cookie import/export 不够因为站点还查设备稳定性时)。
**何时跳过**:一次性抓取、快速调试。

```bash
# 用户手动登录一次(headed 模式需要在容器里有显示,通常先用扫码/cookie 方式)
camoufox-cli --persistent /profiles/xhs open https://www.xiaohongshu.com
# 之后 agent 复用这个 profile,无需密码
camoufox-cli --persistent /profiles/xhs open <url>

# 重置身份:删目录
rm -rf /profiles/xhs
```

> claw-browser 容器内 `/profiles` 已挂成 podman volume(claw-browser.sh
> 默认 `-v claw-browser-profiles:/profiles`),登录态在容器重建后仍保留。

## 常见问题

### "Ref @eN not found"
ref 失效了,重新 snapshot:`camoufox-cli snapshot -i`

### 元素不在 snapshot 里
```bash
camoufox-cli scroll down 1000 && camoufox-cli snapshot -i   # 滚动露出
camoufox-cli wait 2000 && camoufox-cli snapshot -i          # 等动态内容
```

### snapshot 元素太多
```bash
camoufox-cli snapshot -i -s "#main-content"   # 限定容器
```

### 页面没加载完
```bash
camoufox-cli wait --url "*/dashboard" && camoufox-cli snapshot -i
camoufox-cli wait 3000 && camoufox-cli snapshot -i
```

### "command not found: podman" 或容器连不上
claw-browser 容器没起。让用户在物理机跑:
`deploy/browser/claw-browser.sh start`(没构建过先 `build`)。

## 文档

- [camoufox-cli 官方文档](https://github.com/Bin-Huang/camoufox-cli)
