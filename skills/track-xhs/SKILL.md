---
name: track-xhs
description: |
  追踪小红书博主更新 + 搜索识别最新内容。触发场景:用户要"盯着某个小红书博主有没有
  发新笔记"、"每天看 XX 话题在小红书上的最新讨论"、"追踪这几个小红书 KOL"、设置小红书
  内容的定时监控、搜索小红书并筛选值得看的内容。
---

# 小红书追踪(平台差异)

> **先读通用骨架**:`{Base directory}/../track-common/SKILL.md` —— 去重、识别
> 标准、cron 模板、登录态原则都在那。本文只写小红书的差异。

平台标识:`platform = 'xhs'`,登录 profile:`/profiles/xhs`。

## URL 形态

- 博主主页:`https://www.xiaohongshu.com/user/profile/<用户ID>`
- 话题搜索:`https://www.xiaohongshu.com/search_result?keyword=<关键词>`
- 单篇笔记:`https://www.xiaohongshu.com/explore/<笔记ID>`

post_id 用笔记 ID(从 explore/<ID> 链接里取)。

## 页面结构(snapshot 里找什么)

小红书主页/搜索是瀑布流卡片,snapshot -i 里每张卡片大致是:
```
- link "笔记标题文字" [ref=eN]      # 点进去是 /explore/<ID>
  - img "封面alt"
  - text "作者名"
  - text "点赞数"
```
- 笔记标题在 `link` 的 accessible name 里
- 点赞/收藏数有时在相邻 text 节点,可作为"热度"参考辅助筛选
- 列表懒加载:`scroll down 1000 && snapshot -i` 能加载更多,但追踪只看最新几条
  即可,别滚太多

## 登录态

- **公开主页/搜索**:多数能不登录直接看,先试。
- **关注流 / 部分完整内容**:需要登录。让用户先
  `deploy/browser/claw-browser.sh login xhs` 登录一次,之后命令统一加
  `--persistent /profiles/xhs`。

## 小红书特有坑

- **风控较严**:小红书对自动化比较敏感。严格遵守通用骨架的"节制"原则,每天
  1-2 次,别短时间反复 open。
- **滑块验证**:撞到滑块/验证页时,snapshot 里会出现验证相关元素 —— 不要尝试
  自动过验证,直接告诉用户"小红书要求验证,需要你手动处理一次",停下。
- **标题截断**:卡片标题在列表里可能被截断,要完整正文得点进 /explore/<ID>
  再 `snapshot -i` / `text`。只在该笔记通过初筛、值得细读时才点进去,省请求。

## 最小闭环验证(先跑这个)

```jsonc
{"command": "podman exec claw-browser camoufox-cli open https://www.xiaohongshu.com/user/profile/<某公开博主ID> && podman exec claw-browser camoufox-cli wait 3000 && podman exec claw-browser camoufox-cli snapshot -i", "timeout": 280}
```
能稳定读到一组笔记标题 link = 闭环成立,按通用骨架往下做去重+筛选+cron。
读不到(登录墙/风控)= 先解决数据源,别急着配 cron。
