---
name: track-x
description: |
  追踪 X(Twitter)博主更新 + 搜索识别最新推文。触发场景:用户要"盯着某个 X 账号有没有
  发新推"、"看 XX 话题在 X 上的最新讨论"、"追踪这几个 X 上的人"、设置 X 内容的定时监控、
  搜索 X 并筛选值得看的内容。
---

# X(Twitter)追踪(平台差异)

> **先读通用骨架**:`{Base directory}/../track-common/SKILL.md` —— 去重、识别
> 标准、cron 模板、登录态原则都在那。本文只写 X 的差异。

平台标识:`platform = 'x'`,登录 profile:`/profiles/x`。

## URL 形态

- 用户主页:`https://x.com/<handle>`(如 `https://x.com/elonmusk`)
- 话题搜索:`https://x.com/search?q=<关键词>&f=live`(`f=live` = 按最新排序,
  追踪用这个;不加 f=live 默认按热度)
- 单条推文:`https://x.com/<handle>/status/<推文ID>`

post_id 用推文 ID(从 status/<ID> 链接里取)。

## 页面结构(snapshot 里找什么)

X 是时间线 feed,每条推文是一个 article,snapshot -i 大致:
```
- article
  - link "作者名 @handle" [ref=eN]
  - text "推文正文"               # 正文直接可读,这是 X 比抖音强的地方
  - link "时间" [ref=eM]           # 点进去是 status/<ID>
  - 互动按钮: reply / repost / like (带计数)
```
- **推文正文直接在 text 节点**,可读全文 —— X 的"识别"质量比抖音高(有完整文字)
- 互动数(转评赞)可作热度参考
- 转推/引用推:注意区分原创和转发,用户通常更关心原创观点

## 登录态(X 几乎必须登录)

**X 是三个平台里最依赖登录的** —— 2023 后未登录访问被严格限制,公开浏览经常
直接跳登录墙或只给极少内容。所以:
- 默认就走 `--persistent /profiles/x` 登录态,别指望免登录
- 让用户先 `deploy/browser/claw-browser.sh login x` 登录一次
- 没登录态时大概率拿不到 feed,直接告诉用户"X 需要先登录,跑一次 login"

## 墙外访问(代理)

X 在墙内不可直连。claw-browser 容器要能访问 X,需要给容器配代理:
- 启动容器时设 `CLAW_BROWSER_PROXY`(见 claw-browser.sh):
  ```
  CLAW_BROWSER_PROXY=http://host.containers.internal:7890 deploy/browser/claw-browser.sh restart
  ```
- 容器内 camoufox-cli 会自动把代理注入 --proxy(Dockerfile 里的 shim 做的)
- 若 open X 报 `NS_ERROR_NET_INTERRUPT` / 超时,基本是代理没配对 —— 告诉用户
  检查 CLAW_BROWSER_PROXY

## X 特有坑

- **登录态最关键**:X 的所有问题里 90% 是"没登录"或"代理不通"。先确认这两个。
- **正文可读是优势**:充分利用 —— X 追踪能做到真正的内容理解和筛选,不像抖音
  只能看标题。
- **时间线算法**:用户主页按时间倒序,但 search 默认按热度 —— 追踪"最新"务必
  用 `&f=live`。
- **频率**:X 对登录账号的自动化也有限制,同样节制,每天 1-2 次。

## 最小闭环验证(先跑这个)

前提:已 login x + 代理配好。
```jsonc
{"command": "podman exec claw-browser camoufox-cli --persistent /profiles/x open https://x.com/<某handle> && podman exec claw-browser camoufox-cli wait 3000 && podman exec claw-browser camoufox-cli snapshot -i", "timeout": 280}
```
能读到一组 article + 推文正文 = 闭环成立,按通用骨架做去重+筛选+cron。
跳登录墙 = login 没生效或过期,重新 login;连不上 = 查代理。
