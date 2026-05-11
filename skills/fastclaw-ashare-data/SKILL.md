---
description: 🟢 Lightweight A-share data fetch — real-time quotes, K-line, hot stocks, fund flows, fundamentals, news. Use this when user asks "多少钱/现价/K线/涨停/跌停/资金流向/财报/板块" — NOT for "分析/建议/判断/预测/能不能买". See also: tradingagents-ashare for full multi-agent analysis.
description: Directly fetch A-share market data (K-line, real-time quotes, fund flows, hot stocks, technical indicators, fundamentals, news) using TradingAgents-AShare's underlying data providers (AkShare + Sina + Eastmoney). Use when the user wants raw market data without full multi-agent analysis — faster and more focused.
---

# A-Share Market Data Skill

**Purpose**: Provide FastClaw with direct access to A-share market data sources, bypassing the heavy multi-agent analysis pipeline. All data goes through the same providers that power TradingAgents-AShare: AkShare (Eastmoney/Sina/Tencent), Sina Finance real-time API, and Yahoo Finance.

**Source code location**: `/Users/liuyang42/workspace/personal/TradingAgents-AShare/`

---

## Architecture

```
┌─────────────────────────────────────────────┐
│                FastClaw                      │
│  (calls data scripts / HTTP API directly)    │
└──────────────┬──────────────────────────────┘
               │
     ┌─────────┴──────────┐
     ▼                    ▼
┌─────────────┐   ┌──────────────┐
│ HTTP API    │   │ Direct Python │
│ localhost:  │   │ via           │
│ 8000        │   │ exec +        │
│             │   │ akshare       │
└──────┬──────┘   └──────┬───────┘
       │                 │
       ▼                 ▼
┌──────────────────────────────────────────┐
│  Provider Chain (with fallback)          │
│  cn_akshare → cn_baostock → yfinance     │
│    ├─ Eastmoney (stock_zh_a_hist)        │
│    ├─ Sina (hq.sinajs.cn realtime)       │
│    └─ Tencent (stock_zh_a_hist_tx)       │
└──────────────────────────────────────────┘
```

---

## Quick Reference: HTTP API (localhost:8000)

These endpoints are available **without authentication**:

### 1. K-line Data
```bash
curl -s "http://localhost:8000/v1/market/kline?symbol=002812.SZ&start_date=2026-04-01&end_date=2026-05-10"
```
Returns: `{ symbol, start_date, end_date, candles: [{date, open, high, low, close, volume}] }`

- **Symbol format**: `XXXXXX.SH` (Shanghai) or `XXXXXX.SZ` (Shenzhen), e.g. `600519.SH`, `002812.SZ`
- **Index symbols**: `000001.SH` (上证), `399001.SZ` (深证), `399006.SZ` (创业板), `000300.SH` (沪深300)
- **Default**: last 120 days if no start_date
- **Supports Chinese names** via stock map: `?symbol=茅台` works

### 2. Hot Stocks
```bash
# Eastmoney hot list (default)
curl -s "http://localhost:8000/v1/market/hot-stocks?source=em&limit=10"

# Xueqiu most-followed
curl -s "http://localhost:8000/v1/market/hot-stocks?source=xq&limit=10"

# Consecutive rising stocks
curl -s "http://localhost:8000/v1/market/hot-stocks?source=ths&limit=10"
```

### 3. Stock Search
```bash
curl -s "http://localhost:8000/v1/market/stock-search?q=恩捷"
```

### 4. Reports (requires auth)
```bash
curl -s "http://localhost:8000/v1/reports?symbol=002812.SZ&limit=5" \
  -H "Authorization: Bearer $TA_API_TOKEN"
```

---

## Direct Python Data Access

For more control, use Python scripts that import TradingAgents' data providers directly. The virtualenv is at:
```
/Users/liuyang42/workspace/personal/TradingAgents-AShare/.venv/bin/python
```

### K-line with Full OHLCV

```bash
/Users/liuyang42/workspace/personal/TradingAgents-AShare/.venv/bin/python3 -c "
import sys
sys.path.insert(0, '/Users/liuyang42/workspace/personal/TradingAgents-AShare')
from tradingagents.dataflows.interface import route_to_vendor
from tradingagents.dataflows.config import get_config

symbol = '002812.SZ'
raw = route_to_vendor('get_stock_data', symbol, '2026-04-01', '2026-05-10')
print(raw[:3000])
"
```

### Real-time Quotes (from Sina Finance)

```bash
/Users/liuyang42/workspace/personal/TradingAgents-AShare/.venv/bin/python3 -c "
import requests, json

# Sina Finance real-time API (no auth, no rate limit, lightning fast)
codes = ['sh600519', 'sz002812', 'sz300750']
url = 'https://hq.sinajs.cn/list=' + ','.join(codes)
resp = requests.get(url, headers={'Referer': 'https://finance.sina.com.cn/'}, timeout=5)
resp.encoding = 'gbk'

for line in resp.text.splitlines():
    if '=\"' not in line: continue
    var, data = line.split('=\"', 1)
    code = var.split('_')[-1]
    fields = data.rstrip('\";').split(',')
    if len(fields) < 10: continue
    name = fields[0]
    open_p, prev_close, price = fields[1], fields[2], fields[3]
    high, low = fields[4], fields[5]
    volume, amount = fields[8], fields[9]
    change = float(price) - float(prev_close)
    pct = round(change / float(prev_close) * 100, 2)
    print(f'{code} {name}: {price} ({pct:+.2f}%) O:{open_p} H:{high} L:{low} V:{volume}')
"
```

The Sina field mapping:
| Index | Field | Index | Field |
|-------|-------|-------|-------|
| 0 | 股票名称 | 5 | 最低价 |
| 1 | 今开盘 | 8 | 成交量(手) |
| 2 | 昨收盘 | 9 | 成交额(万元) |
| 3 | 当前价 | 30 | 日期 |
| 4 | 最高价 | 31 | 时间 |

### Technical Indicators

```bash
/Users/liuyang42/workspace/personal/TradingAgents-AShare/.venv/bin/python3 -c "
import sys
sys.path.insert(0, '/Users/liuyang42/workspace/personal/TradingAgents-AShare')
from tradingagents.dataflows.interface import route_to_vendor

# Supported indicators: close_50_sma, close_200_sma, close_10_ema,
# macd, macds, macdh, rsi, boll, boll_ub, boll_lb, atr, vwma, mfi
indicator = 'rsi'
result = route_to_vendor('get_indicators', '002812.SZ', indicator, '2026-05-10', 14)
print(result)
"
```

### Fundamentals

```bash
/Users/liuyang42/workspace/personal/TradingAgents-AShare/.venv/bin/python3 -c "
import sys
sys.path.insert(0, '/Users/liuyang42/workspace/personal/TradingAgents-AShare')
from tradingagents.dataflows.interface import route_to_vendor

# Company profile + financial abstract
print(route_to_vendor('get_fundamentals', '002812.SZ'))
"
```

### Financial Reports (Balance Sheet / Income / Cashflow)

```bash
/Users/liuyang42/workspace/personal/TradingAgents-AShare/.venv/bin/python3 -c "
import sys
sys.path.insert(0, '/Users/liuyang42/workspace/personal/TradingAgents-AShare')
from tradingagents.dataflows.interface import route_to_vendor

print('=== 利润表 ===')
print(route_to_vendor('get_income_statement', '002812.SZ'))
print()
print('=== 资产负债表 ===')
print(route_to_vendor('get_balance_sheet', '002812.SZ'))
print()
print('=== 现金流量表 ===')
print(route_to_vendor('get_cashflow', '002812.SZ'))
"
```

### News

```bash
/Users/liuyang42/workspace/personal/TradingAgents-AShare/.venv/bin/python3 -c "
import sys
sys.path.insert(0, '/Users/liuyang42/workspace/personal/TradingAgents-AShare')
from tradingagents.dataflows.interface import route_to_vendor

print(route_to_vendor('get_news', '002812.SZ', '2026-04-20', '2026-05-10'))
"
```

### A-Share Market Breadth Data (AkShare direct)

```bash
/Users/liuyang42/workspace/personal/TradingAgents-AShare/.venv/bin/python3 -c "
import akshare as ak

# 板块资金流向排名
df = ak.stock_board_industry_fund_flow_em(symbol='今日')
print('=== 行业板块资金净流入 Top 5 ===')
top = df.sort_values('今日主力净流入-净额', ascending=False).head(5)
for _, r in top.iterrows():
    print(f\"  {r['板块名称']}: {r['今日主力净流入-净额']/1e8:.1f}亿\")

# 涨停板情绪
zt = ak.stock_zt_pool_em(date='20260509')
print(f\"\\n=== 涨停板情绪 ===\")
print(f\"  涨停家数: {len(zt)}\")
if '连板数' in zt.columns:
    lb = zt['连板数'].value_counts().sort_index()
    for k, v in lb.items():
        print(f'  {int(k)}连板: {v}家')

# 个股资金流向
fund = ak.stock_individual_fund_flow(stock='002812', market='sz')
print(f\"\\n=== 恩捷股份近5日资金流向 ===\")
print(fund.tail(5).to_string())
"
```

---

## Data Provider Details

### Provider Chain (auto-fallback)
For A-share stocks, the default chain is:
```
cn_akshare → cn_baostock → yfinance
```

**cn_akshare** uses multiple data sources internally with fallback:
1. **Eastmoney** (`stock_zh_a_hist`): Primary, 前复权 daily OHLCV
2. **Sina** (`stock_zh_a_daily`): Fallback #1
3. **Tencent** (`stock_zh_a_hist_tx`): Fallback #2
4. **Sina real-time** (`hq.sinajs.cn`): For live quotes and ETF data

### Concurrency Control
AkShare calls are rate-limited:
- Max 5 concurrent calls total
- Max 3 from scheduled tasks (reserves 2 for frontend)
- 120s stale timeout (zombie thread recovery)

### Symbol Format
- A-shares: `002812.SZ`, `600519.SH`
- Can also accept: `002812`, `恩捷股份`, `贵州茅台`
- US stocks: `AAPL` (routed to yfinance)
- HK stocks: `00700.HK`

### Indicator Support
| Indicator | Description | Typical Use |
|-----------|-------------|-------------|
| `close_50_sma` | 50-day SMA | 中期趋势 |
| `close_200_sma` | 200-day SMA | 长期趋势 |
| `close_10_ema` | 10-day EMA | 短线动量 |
| `macd` | MACD line | 趋势+动量 |
| `rsi` | RSI (14) | 超买超卖 |
| `boll` / `boll_ub` / `boll_lb` | 布林带(20,2) | 波动区间 |
| `atr` | ATR(14) | 波动率/风控 |
| `vwma` | VWMA(20) | 量价均线 |
| `mfi` | MFI(14) | 资金流量 |

---

## Usage Patterns

### Pattern 1: Quick Price Check
When user asks "恩捷股份现在多少钱？":
1. Call Sina real-time API directly (fastest, no auth)
2. Return price, change%, volume

### Pattern 2: Technical Analysis Prep
When user wants to analyze a stock technically:
1. Fetch K-line via HTTP API (120 days)
2. Fetch 2-3 key indicators (RSI, MACD, MA) via `route_to_vendor`
3. Present summary with key levels

### Pattern 3: Market Sentiment
When user asks about market mood:
1. Call hot stocks endpoint
2. Call `stock_zt_pool_em` for limit-up count
3. Check `stock_board_industry_fund_flow_em` for sector flows

### Pattern 4: Deep Dive
When user wants comprehensive data:
1. K-line history (HTTP API)
2. Technical indicators (route_to_vendor)
3. Fundamentals (route_to_vendor)
4. News (route_to_vendor)
5. Fund flow (akshare direct)

---

## ⚠️ Important Notes

1. **AkShare is NOT thread-safe for all functions**: The concurrency lock (`AkshareLock`) must be used. Our skill scripts automatically respect this via `route_to_vendor`.

2. **Sina real-time API** (`hq.sinajs.cn`): No auth, no rate limit, returns GBK-encoded data. Best for quick price checks. Fields are positional (see table above).

3. **Eastmoney requires akshare**: The `stock_zh_a_hist` function may trigger anti-crawl. The built-in lock prevents hammering.

4. **Trading calendar**: TradingAgents maintains a CN trading calendar. Non-trading days return `N/A` for real-time data.

5. **Data caching**: TradingAgents caches spot data for 8 seconds to avoid hammering Eastmoney under concurrent load.

6. **Timeout**: AkShare calls have a 60s acquire timeout and 120s stale timeout. If you hit a timeout, wait and retry.
