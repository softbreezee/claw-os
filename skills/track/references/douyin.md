# 抖音追踪(平台差异)

> 通用流程见上级 `SKILL.md`。本文只写抖音差异。

平台标识:`platform = 'douyin'`,登录 profile:`/profiles/douyin`。

## URL 形态

- 博主主页:`https://www.douyin.com/user/<sec_uid>`
- 话题搜索:`https://www.douyin.com/search/<关键词>`
- 单条视频:`https://www.douyin.com/video/<视频ID>`

post_id 用视频 ID(从 video/<ID> 链接里取)。

## 页面结构(snapshot 里找什么)

抖音主页是视频网格,snapshot -i 里每个视频条目大致:
```
- link "视频标题/描述" [ref=eN]    # 点进去是 /video/<ID>
  - img "封面"
  - text "播放量/点赞数"
```
- 视频文案(标题)在 link 的 name 里;抖音文案常带大量话题标签 #xxx#,筛选时
  关注实质内容,别被标签带偏
- 抖音内容是**视频**,文字信息少 —— "识别"主要靠标题文案 + 可见的互动数据。
  视频内容本身无法读(没有字幕提取),只能基于文案+热度判断,告诉用户"这是
  视频,我基于标题和热度判断,具体内容需你自己看"

## 登录态

- **公开主页**:部分能不登录看到作品列表,先试。
- **完整浏览 / 避免频繁弹登录框**:登录更稳。让用户先
  `deploy/browser/claw-browser.sh login douyin`,之后加 `--persistent
  /profiles/douyin`。

## 抖音特有坑

- **签名风控重**:抖音是反爬最狠的之一。Camoufox 能过浏览器指纹,但抖音还有
  行为检测 —— 严格节制频率,每天 1-2 次。
- **登录弹窗**:抖音频繁弹登录引导遮挡内容。snapshot 里若出现登录弹窗元素,
  先尝试关闭(找 close/✕ 的 ref 点掉)再 snapshot;关不掉就用 persistent 登录态。
- **视频不可读**:抖音追踪只能基于文案+互动数判断,无法理解视频内容本身。
  用户要"视频里讲了啥",老实告诉他做不到,只能给标题。
- **scroll 加载**:作品列表懒加载,但追踪只看最新几条,别滚太多触发风控。

## 最小闭环验证(先跑这个)

```jsonc
{"command": "podman exec claw-browser camoufox-cli open https://www.douyin.com/user/<某公开博主sec_uid> && podman exec claw-browser camoufox-cli wait 3000 && podman exec claw-browser camoufox-cli snapshot -i", "timeout": 280}
```
能读到一组视频条目 link = 闭环成立,按通用骨架做去重+筛选+cron。
读不到(登录墙/风控/全是登录弹窗)= 先用 persistent 登录态,还不行就先告诉
用户抖音这条暂时走不通,别硬刚。
