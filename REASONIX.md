# Reasonix project memory

Notes the user pinned via the `#` prompt prefix. The whole file is
loaded into the immutable system prefix every session — keep it terse.

- Memory

## 交易系统架构

### Agent 清单
| Agent | 账户 | 职责 | Cron |
|-------|------|------|------|
| trend-trader | 趋势 | 中长线趋势扫描+持仓诊断 | 16:00 交易日 |
| dragon-hunter | 龙头 | 涨停板复盘+龙头辨识+接力预判 | 15:10 交易日 |
| review-coach | 双账户 | 逐笔归因+错误模式+绩效统计 | 15:30 交易日 |
| morning-brief | 双账户 | 盘前简报+作战计划+竞价速览 | 08:30/09:25 交易日 |
| deep-research | 双账户 | 深度基本面研究(按需触发) | 无 |

### TA（TradingAgents-AShare）
- 独立的多 Agent A 股分析系统，位于 `/Users/leon/Projects/TradingAgents-AShare/`
- 数据库：SQLite `/Users/leon/Projects/TradingAgents-AShare/tradingagents.db`
- 7 个分析 Agent：市场技术面、量价、新闻、社交舆情、基本面、宏观板块、主力资金
- 通过辩论裁判收口，产出最终交易决策（BUY/SELL + 信心度 + 目标价/止损价）
- 触发方式：通过 API 或直接调用

### 账户与持仓（2026-05-17 基准）

**账户一：中山证券** | 总资产 81,224.61 | 标记：趋势
| 股票 | 代码 | 成本 | 仓位 | 股数 |
|------|------|------|------|------|
| 恩捷股份 | 002812 | 83.610 | 84.7% | 900 |
| 新疆天业 | 600075 | 8.250 | 8.5% | ~800 |

**账户二：平安证券** | 总资产 193,575.38
| 股票 | 代码 | 成本 | 仓位 | 股数 |
|------|------|------|------|------|
| 中国长城 | 000066 | 23.619 | 34.9% | ~2900 |
| 中钨高新 | 000657 | 58.502 | 32.1% | ~1100 |
| 汇绿生态 | 001267 | 64.006 | 22.7% | ~700 |
| 九丰能源 | 605090 | 45.119 | 9.0% | ~400 |

### 数据库 (fastclaw_db, psql -h /tmp)
| 表 | 用途 | 写入者 |
|----|------|--------|
| trading_journal | 交易日志 | review-coach, 用户手动 |
| trading_watchlist | 自选池 | trend-trader, dragon-hunter |
| trading_daily_analysis | 每日分析 | 所有 Agent |
| trading_performance | 绩效快照 | review-coach |
| trading_daily_plan | 作战计划 | morning-brief |
| trading_market_snapshot | 市场快照 | morning-brief |

### 查询约定（重要）
- **trading_journal 查持仓按 `account` 过滤，不按 `agent_id`**。agent_id 仅作审计字段（记录谁写入的）。
- 查当前持仓：`WHERE account = '平安证券' AND exit_date IS NULL`
- 各 Agent 按需过滤 account：trend-trader → 中山证券，dragon-hunter → 平安证券，review-coach/morning-brief/deep-research → 两个都查

### 数据源
- pawnix-ashare-data (AkShare + Sina API)
- HTTP API: http://localhost:8000
- TA 系统: 独立运行

### Cron 调度
- 08:30 morning-brief 盘前简报
- 09:25 morning-brief 集合竞价
- 15:10 dragon-hunter 涨停板初扫
- 15:30 review-coach 当日复盘
- 16:00 trend-trader 趋势扫描

### 编排方式
- 无独立编排 Agent。pawnpawn 本身就是总管
- Cron 触发 pawnpawn → pawnpawn spawn_subagent 调用对应 Agent → 收回复 → notify 用户
- 用户直接对话各 Agent：在 Pawnix 中选择对应 Agent 发起对话
pawnpawn 的 memory 里不就这些内容吗，哪里来的 46 条记忆呢？
