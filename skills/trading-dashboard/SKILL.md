---
name: trading-dashboard
description: |
  生成交易作战仪表盘 —— 把 PG 里的 trading_* 数据(持仓、自选池、绩效、当日分析)
  渲染成一个自包含 HTML 面板。触发场景:用户要"看一下我的持仓/账户盘面"、"生成
  交易仪表盘/驾驶舱"、"今天两个账户怎么样"、"做个能一眼看的面板"、盘前/盘后想要
  一个图文概览而不是聊天里一问一答。
---

# 交易作战仪表盘

把用户 PG 里的 trading_* 数据生成一个**自包含 HTML 仪表盘**(单文件,内联 CSS/JS),
写到 workspace,给用户一个能在浏览器打开的链接。一眼看全两个账户,而不是聊天里
逐条问。

> 前提:PG 后端已配(db_query 可用)。数据表见用户的 REASONIX.md /
> 项目记忆;**字段不要假设,运行时探查**(见下)。

## 关键原则:先探 schema,别猜字段

不同环境表结构可能不同。**第一步永远是探查真实字段**,基于真实结果写后续查询:

```sql
-- 探查所有 trading_ 表的字段
SELECT table_name, column_name, data_type
FROM information_schema.columns
WHERE table_name LIKE 'trading_%'
ORDER BY table_name, ordinal_position;
```

用 db_query 跑这个,看清每张表到底有哪些列,再决定下面的查询怎么写。
**不要照搬本文示例 SQL 的列名** —— 它们是示意,以探查结果为准。

## 数据来源(用户记忆里的约定)

| 表 | 内容 | 查询要点 |
|----|------|---------|
| trading_journal | 交易日志/持仓 | **持仓按 `account` 过滤,不按 agent_id**;当前持仓 `WHERE account=? AND exit_date IS NULL` |
| trading_watchlist | 自选池 | trend-trader / dragon-hunter 写 |
| trading_performance | 绩效快照 | review-coach 写,画曲线用 |
| trading_daily_analysis | 每日分析 | 点某只票展开看历史 |
| trading_daily_plan | 作战计划 | morning-brief 写 |

账户:中山证券(趋势)、平安证券(龙头)。两个账户分 tab 展示。

## 生成流程

1. **探 schema**(上面那条 information_schema 查询)
2. **查每个账户的当前持仓**:
   ```sql
   SELECT * FROM trading_journal
   WHERE account = $1 AND exit_date IS NULL;
   ```
   (列名以探查结果为准:成本、现价、股数、仓位等)
3. **查自选池**:`SELECT * FROM trading_watchlist ...`
4. **查绩效曲线**:`SELECT * FROM trading_performance ORDER BY <日期列>;`
5. **(可选)查今日 daily_plan / market_snapshot** 作为页眉摘要
6. **算盈亏**:有现价就 (现价-成本)/成本;没有现价列就只显示成本+仓位,并在
   页面标注"现价缺失,需接行情源"。**不要编现价**。
7. **生成 HTML**(见下模板),`write_file` 到 `dashboard.html`
8. **给链接**:文件在 agent workspace,用户通过 claw-os 文件服务打开:
   `/api/files?kind=workspace&agentId=<当前agent>&path=dashboard.html`
   —— 把这个相对链接告诉用户,或直接说"已生成 dashboard.html,在 Files 里打开"。

## HTML 模板要求

生成的 HTML 必须:
- **单文件自包含**:CSS 内联 `<style>`,JS 内联 `<script>`,不引外部资源
  (Chart.js 例外:可以用 CDN `<script src="https://cdn.jsdelivr.net/npm/chart.js@4.5.0/dist/chart.umd.js">`)
- **light mode**:白底深字(`:root{color-scheme:light}`),claw-os UI 是亮色
- **双账户 tab**:顶部切「中山证券(趋势)」「平安证券(龙头)」
- **持仓卡片**:每只票一张卡 —— 代码+名称、成本、(现价)、仓位%、股数;
  **盈利绿、亏损红**(盈亏不明时中性灰,不要瞎标颜色)
- **绩效曲线**:Chart.js 折线,x=日期 y=总资产/收益
- **自选池区块**:列表展示 watchlist
- **页眉**:生成时间 + 两账户总资产汇总
- **数据快照声明**:页眉标明"数据快照于 <生成时间>,非实时行情"

## 数据为空 / 表不存在的处理

- 表不存在(information_schema 查不到)→ 告诉用户"trading_* 表还没建/没数据,
  先让交易 agent 跑一轮写入",不要生成空壳面板
- 某账户无持仓 → 该 tab 显示"当前空仓"
- 缺现价列 → 如实标注,只展示成本/仓位维度

## 实时性

这是**按需生成的快照**,不是自动刷新的实时面板。用户要"盘中半实时":
建议配 cron 定时重新生成(如盘中每 30 分钟),用 create_cron_job:
```
name: "dashboard-refresh"
type: cron
schedule: "*/30 9-15 * * 1-5"   # 交易时段每 30 分钟
message: "用 trading-dashboard skill 重新生成交易仪表盘 dashboard.html"
```
但默认不要自动配 cron,等用户明确要。

## 可交互:卡片按钮回调 agent(C)

仪表盘不只是看的,卡片上可以带按钮,点了**直接触发 agent 干下一步** ——
比如持仓卡上一个「深入分析」按钮,点了就让 agent 去查这只票的 daily_analysis
+ 最新行情给出诊断。这把"看面板"变成"在面板上操作"。

原理:页面和 claw-os 同源(都在 :18953),按钮 onclick 用 fetch POST
`/api/chat/submit` 触发一个新 agent 任务。

**生成 HTML 时,把 agentId / sessionId 内联成 JS 常量**(页面自己猜不到,
必须你生成时写进去 —— agentId = 当前 agent 名,sessionId = 当前会话 key):

```html
<script>
  const AGENT_ID = "<当前agent名>";      // 生成时填真实值
  const SESSION_ID = "<当前sessionId>";  // 生成时填真实值
  function ask(message) {
    fetch("/api/chat/submit", {
      method: "POST",
      headers: { "Content-Type": "application/json" },
      body: JSON.stringify({ agentId: AGENT_ID, sessionId: SESSION_ID, message })
    }).then(function(r){ return r.json(); })
      .then(function(){ alert("已发给 agent,回聊天窗口看结果"); })
      .catch(function(e){ alert("发送失败: " + e); });
  }
</script>
```

按钮示例(持仓卡片上):
```html
<button onclick="ask('深入分析 恩捷股份(002812):查它的 daily_analysis 历史 + 最新行情,给我诊断')">深入分析</button>
```

**用按钮回调时的注意:**
- 任务在聊天窗口里执行(不是面板里),点完按钮提示用户回聊天看 —— 面板只是
  触发器,不显示结果(面板拿不到 agent 的流式输出)。
- message 要写完整自包含的指令(agent 收到的是一个全新任务,不带面板上下文)。
- 别滥用:只在"明确的下一步动作"上加按钮(深入分析某只票、刷新面板、把某票
  加入 watchlist)。不是每个数字都要能点。
- 这是可选增强 —— 用户没要交互按钮时,生成纯展示面板即可,别强塞。

## HTML 骨架参考(按探查到的真实数据填充)

```html
<!doctype html><html lang="zh"><head><meta charset="utf-8">
<meta name="viewport" content="width=device-width,initial-scale=1">
<title>交易作战仪表盘</title>
<style>
  :root{color-scheme:light}
  body{margin:0;font:14px/1.5 -apple-system,sans-serif;background:#f7f8fa;color:#1a1a1a}
  .wrap{max-width:1100px;margin:0 auto;padding:20px}
  header{display:flex;justify-content:space-between;align-items:baseline;margin-bottom:16px}
  .muted{color:#888;font-size:12px}
  .tabs{display:flex;gap:8px;margin-bottom:16px}
  .tab{padding:6px 14px;border:1px solid #ddd;border-radius:8px;cursor:pointer;background:#fff}
  .tab.active{background:#1a1a1a;color:#fff;border-color:#1a1a1a}
  .cards{display:grid;grid-template-columns:repeat(auto-fill,minmax(240px,1fr));gap:12px}
  .card{background:#fff;border:1px solid #eee;border-radius:12px;padding:14px}
  .pos .gain{color:#16a34a} .pos .loss{color:#dc2626}
  .big{font-size:20px;font-weight:600}
  canvas{background:#fff;border:1px solid #eee;border-radius:12px;padding:12px;margin-top:16px}
  .acct{display:none} .acct.active{display:block}
</style></head>
<body><div class="wrap">
  <header><div class="big">交易作战仪表盘</div>
    <div class="muted">数据快照于 <!--生成时间--> · 非实时行情</div></header>
  <!-- tab 按钮 + 各账户持仓卡片 + watchlist + 绩效 canvas -->
  <script src="https://cdn.jsdelivr.net/npm/chart.js@4.5.0/dist/chart.umd.js"></script>
  <script>
    // tab 切换 + new Chart(...) 画绩效曲线;数据由 agent 内联进来
  </script>
</div></body></html>
```

把探查到的真实数据内联进这个骨架,别留占位符给浏览器去拉(它拉不到 db)。
