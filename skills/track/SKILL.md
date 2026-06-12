---
name: track
description: |
  追踪社媒博主更新 + 搜索识别最新内容(小红书 / 抖音 / X)。触发场景:用户要"盯着某个
  博主有没有发新内容"、"每天看 XX 话题的最新讨论"、"追踪这几个 KOL"、设置社媒内容的
  定时监控、搜索某平台并筛选值得看的内容。本 skill 是通用骨架;具体平台差异按需读
  references/<平台>.md(xhs / douyin / x)。
---

# 社媒追踪(通用骨架)

三个平台(小红书 / 抖音 / X)追踪流程本质相同,差异只在 URL、页面结构、登录方式。
通用部分在这里;**确定平台后,读对应差异**:
`{Base directory}/references/xhs.md` / `douyin.md` / `x.md`。

> **前置**:必须先会 browser skill。读 `{Base directory}/../browser/SKILL.md`。
> 所有取数靠 `podman exec claw-browser camoufox-cli ...`。

## 核心价值

不是把原始数据倒给用户,而是**帮用户过滤判断、只推有信息量的**。这是 agent
比纯爬虫强的地方。

## 通用流程(五步)

1. **打开**目标页(博主主页 or 搜索结果页)。首次 open 必带 `timeout: 280`:
   ```jsonc
   {"command": "podman exec claw-browser camoufox-cli open <URL> && podman exec claw-browser camoufox-cli wait 3000 && podman exec claw-browser camoufox-cli snapshot -i", "timeout": 280}
   ```
2. **提取**内容列表 —— 从 snapshot 的 aria 树里找帖子/视频条目(平台差异见 references)。
3. **去重** —— 用 db 记住已推过的,只处理新的(见下)。
4. **筛选 + 总结** —— 按识别标准过滤垃圾,对值得看的 LLM 总结 2-3 句。
5. **通知** —— `notify` 推给用户;没有新内容就**不打扰**。

## 去重(只推"新"内容)

第一次设置时建一张通用表(三平台共用):
```
db_create_table: social_seen_posts
  字段: platform(TEXT), post_id(TEXT), author(TEXT), title(TEXT), seen_at(TIMESTAMP)
  用途: 记录已推送过的社媒内容,跨平台去重
```

每次检查:snapshot 拿列表 → 每条算 post_id(优先用链接里的 ID;没有就
platform+author+标题 做指纹) → db_query 查该 (platform, author) 已记录的 →
**只处理没见过的** → 总结后写回表 → notify。

## 识别标准(用户的"值得看" vs 垃圾)

默认规则(用户可在触发时覆盖):
- ✅ **推**:有具体数据/数字、行业内部视角、方法论、对比分析、真实经验
- ❌ **过滤**:纯带货/广告、纯情绪输出、标题党无实质、与用户关注点无关

总结格式(notify 的 body):
```
【平台·博主名】<标题>
核心:<2-3 句提炼,有数据/观点的优先写出>
链接:<URL>
```
多条合并成一条 notify,别刷屏。

## 定时追踪(cron 模板)

闭环验证通过后再配 cron。message 只写做什么,不写投递细节:
```
create_cron_job:
  name: "track-<平台>-<博主名>"
  type: cron
  schedule: "0 20 * * *"        # 每天 20:00,频率别太高
  message: "用 track skill 检查 <博主URL> 今天有没有新内容。
            用 social_seen_posts 表去重(platform=<平台>),只总结没推过的。
            按识别标准过滤带货和纯情绪贴。有新内容 notify 总结给我,没有就不打扰。"
```
博主多时一个 cron 里循环查多个 URL(省冷启动),少时一博主一 cron。

## 登录态原则(三平台通用)

- 公开主页/搜索通常**不用登录**,先试无登录。
- 需要登录的:**绝不自动输密码**。用 browser skill 的 `--persistent
  /profiles/<平台>` profile,用户手动登录一次(见 claw-browser.sh login),复用。

## 通用注意

- **首次 open 必带 `timeout: 280`** —— 忘了 daemon 被杀,后续全失败。
- **风控是猫鼠游戏** —— 某天拿不到数据是正常的,先告诉用户"今天被挡了",别狂重试。
- **节制** —— 每天 1-2 次足够,高频访问更容易触发风控。
- **SPA 用 snapshot 不用 text body** —— 三平台都是 JS 渲染。

## 平台差异(确定平台后读对应文件)

- 小红书 → `references/xhs.md`
- 抖音 → `references/douyin.md`
- X(Twitter) → `references/x.md`
