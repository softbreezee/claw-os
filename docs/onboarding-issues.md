# Onboarding 流程交互问题清单

> 起源：用户反馈 "项目初始化、从开始页面进行配置时交互有很多 bug，且模型选择没有 DeepSeek"。本文记录对 [`web/src/app/onboard/page.tsx`](../web/src/app/onboard/page.tsx) + [`internal/setup/handlers.go: handleSaveConfig`](../internal/setup/handlers.go:424-557) 端到端 review 的发现。

按严重程度分级：🔴 数据/逻辑性 bug（必修） · 🟡 体验问题（应修） · 🟢 打磨项

---

## 修复状态（2025-01）

| Bug    | 描述                                  | 状态        | 实现位置 |
| ------ | ------------------------------------ | ----------- | -------- |
| Bug-1  | PROVIDERS 缺失 DeepSeek 等主流厂商     | ✅ 已修复    | [`web/src/app/onboard/page.tsx`](../web/src/app/onboard/page.tsx) `PROVIDERS` 增加 deepseek/openai/anthropic/google/moonshot/qwen，默认改 DeepSeek |
| Bug-2  | OpenRouter 默认模型 ID 是虚构的         | ✅ 已修复    | OpenRouter models 改为 `anthropic/claude-sonnet-4` 等真实 ID |
| Bug-3  | model 字符串被二次拼接前缀              | ✅ 已修复    | [`internal/setup/handlers.go`](../internal/setup/handlers.go) 新增 `resolveModelID()`，已含 `/` 不重复加前缀 |
| Bug-4  | agent.json 与 pawnix.json model 不一致 | ✅ 已修复    | 两处统一用 `canonicalModel`（resolveModelID 结果） |
| Bug-5  | Test 没成功也允许 Next                 | ✅ 已修复    | `canProceed` step 1 加 `testStatus === "success"` 校验 |
| Bug-6  | testStatus 不在表单变更时重置          | ✅ 已修复    | `updateConfig` 检测 `TEST_INVALIDATING_KEYS` 自动 reset |
| Bug-7  | handleLaunch 失败假装成功              | ✅ 已修复    | 解析 `{ok, error}` 契约；失败显示 `launchError` 横条，不放彩带不跳转 |
| Bug-8  | 跳转 hardcode `localhost`              | ✅ 已修复（v0.3）| `window.location.hostname` + protocol；handleLaunch 通过 `waitForGateway(20s)` 轮询 `/api/status`，daemon 报告 `running:true` 后再跳；超时时附带 daemon.log 提示但仍跳转，避免卡住用户 |
| Bug-9  | handleSaveConfig 硬覆盖写入            | ✅ 已修复    | 改为 `config.Load() + merge` 模式，保留 storage/cron/hooks/其它 agent |
| Bug-10 | Telegram 字段无 UI                    | ✅ 已修复（v0.3）| step 2 在 Personality 之前加 Telegram bot 复选框 + token 输入；勾选时 token 必填进入 `canProceed`；后端 `saveConfigRequest.TelegramEnabled / TelegramToken` 已存在，前端只补 UI |
| Bug-19 | storage（DB）配置无 onboard 入口       | ✅ 已修复    | step 2 加 Storage Backend select + PostgreSQL DSN 输入；后端 `saveConfigRequest` 加 `StorageType` / `StorageDSN` |

P1/P2/P3 的体验和打磨项（Bug-11~18）暂未处理，留待下一轮。

---

## 🔴 P0：模型选择与 Provider 配置

### Bug-1: PROVIDERS 缺失 DeepSeek（以及所有主流厂商）

**位置**：[`web/src/app/onboard/page.tsx:28-44`](../web/src/app/onboard/page.tsx:28-44)

```ts
const PROVIDERS = {
  openrouter: { ... },
  ollama:     { ... },
  custom:     { ... },
};
```

只有 3 个 provider，没有 DeepSeek、Anthropic、OpenAI、Moonshot、Qwen、Google 任何一家的直连预设。

**对照现状**：后端 [`modelcatalog.go:50-88`](../internal/modelcatalog/modelcatalog.go:50-88) 内置了 DeepSeek（v4-pro / v4-flash / chat / reasoner）、Claude 全家、GPT-4o / 5 / o-series、Gemini 2.5、Kimi K2/K2.5、Qwen 等 25+ 模型，但 onboard UI 完全没用上这份目录。

**修复**：
1. 把 `PROVIDERS` 从前端硬编码迁出，改为后端 `GET /api/onboard/providers` 拉取（或在前端按 modelcatalog.go 的内容补齐）
2. 至少补全：
   - `deepseek`：apiBase=`https://api.deepseek.com/v1`，models=`["deepseek-chat","deepseek-reasoner"]`
   - `anthropic`：apiBase=`https://api.anthropic.com`，apiType=`anthropic-messages`，models=`["claude-sonnet-4","claude-opus-4","claude-3-5-sonnet"]`
   - `openai`：apiBase=`https://api.openai.com/v1`，models=`["gpt-5","gpt-4o","gpt-4o-mini","o3","o3-mini","o4-mini"]`
   - `moonshot`：apiBase=`https://api.moonshot.cn/v1`，models=`["kimi-k2","kimi-k2.5"]`
   - `qwen`（DashScope OpenAI 兼容入口）：apiBase=`https://dashscope.aliyuncs.com/compatible-mode/v1`，models=`["qwen-max","qwen2.5-72b"]`
   - `google`（Gemini OpenAI 兼容入口）：apiBase=`https://generativelanguage.googleapis.com/v1beta/openai`，models=`["gemini-2.5-pro","gemini-2.5-flash"]`

### Bug-2: OpenRouter 默认模型 ID 是虚构的

**位置**：[`web/src/app/onboard/page.tsx:32-36`](../web/src/app/onboard/page.tsx:32-36)

```ts
models: [
  "openai/gpt-5.4",
  "anthropic/claude-sonnet-4.6",
  "google/gemini-3.1-flash-lite-preview",
],
```

这三个模型 ID 在 OpenRouter 上都不存在（`gpt-5.4` 没发布、`claude-sonnet-4.6` 没发布、`gemini-3.1-flash-lite-preview` 没发布）。用户用默认值点 Test Connection 必然 400/404。

**修复**：换成已发布的真实 ID，例如：
```ts
[
  "anthropic/claude-sonnet-4",
  "openai/gpt-5",
  "google/gemini-2.5-pro",
  "deepseek/deepseek-chat",
]
```

### Bug-3: 模型字符串被二次拼接前缀，测试通过≠启动可用

**位置**：[`internal/setup/handlers.go:455`](../internal/setup/handlers.go:455)

```go
Model: providerKey + "/" + req.Model,
```

如果用户选 `provider=openrouter`、UI 默认 `model=openai/gpt-5.4`，最终保存到 `pawnix.json` 的是：
```
openrouter/openai/gpt-5.4   ← 三段 slash
```

而 [`testProvider`](../internal/setup/handlers.go:155-250) 测试的 model 字符串是 `openai/gpt-5.4`（两段），**测试通过的串和实际写入的串完全不一样**。结果：onboard 全绿通过，启动后 provider registry 解析失败。

**修复**：
- 当 `req.Model` 已经包含 `/` 时不再加前缀，直接使用
- 或者在 provider 选择时明确"OpenRouter 等 aggregator 的模型用完整 ID，直连 provider 的模型不带前缀"，前端做规范化
- 写入前再过一遍和 testProvider 同款的 normalize 逻辑

### Bug-4: agent.json 与 pawnix.json 的 model 字段不一致

**位置**：[`handlers.go:455` vs `handlers.go:542`](../internal/setup/handlers.go:455)

```go
// pawnix.json
Defaults.Model = providerKey + "/" + req.Model   // 带前缀

// agent.json
agentCfg := config.AgentFileConfig{Model: req.Model}  // 不带前缀
```

同一个 agent 的 model 在两处定义不一致，行为依赖谁加载得早 / normalize 是否覆盖到位。

**修复**：两处统一用 normalize 后的同一份字符串。

---

## 🔴 P0：交互逻辑 bug

### Bug-5: Test Connection 没成功也允许 Next

**位置**：[`canProceed` step 1](../web/src/app/onboard/page.tsx:213-226)

```ts
case 1:
  return config.apiBase.length > 0 && config.model.length > 0;
```

完全不校验 `testStatus === "success"`。用户填错 API Key、Test Connection 报红，仍然可以 Next → Launch → 启动失败。

**修复**：
- 把 Next 按钮 disabled 条件加上 `testStatus === "success"`，或弹 confirm "未测试连接，确认继续？"
- 至少在 testStatus=error 时给 Next 按钮一个警告标

### Bug-6: testStatus 不会在表单变更时重置

**位置**：[`updateConfig`](../web/src/app/onboard/page.tsx:149-154)

修改 `apiBase` / `apiKey` / `model` 后，`testStatus` 仍然是 "success"，会显示绿色 Connected 徽章，**误导用户以为新配置已通过测试**。

**修复**：在 `updateConfig` 里检测 apiBase / apiKey / model / apiType / authType 的变化，自动 `setTestStatus("idle")`。

### Bug-7: handleLaunch 失败时假装成功

**位置**：[`handleLaunch`](../web/src/app/onboard/page.tsx:196-211)

```ts
} catch {
  setLaunched(true);     // ← 失败也标记 launched
  setShowConfetti(true); // ← 还放彩带
  ...
}
```

`saveConfig` 失败时静默吞错，UI 显示 "You're All Set!" + 彩带，3 秒后跳转到 chat 页面（实际 config 没存）。这是非常严重的"骗用户"型 bug。

**修复**：
```ts
} catch (err) {
  setLaunchError(String(err));
  setLaunchStatus("error");
  return;  // 不放彩带、不跳转
}
```

### Bug-8: handleLaunch 跳转用 hardcode `localhost`，且不等 daemon 起来

**位置**：[`handleLaunch:202-205`](../web/src/app/onboard/page.tsx:202-205)

```ts
setTimeout(() => {
  const port = config.port || window.location.port;
  window.location.href = `http://localhost:${port}/chat/`;
}, 3000);
```

三个隐患：
1. **hostname 写死 `localhost`**：远程 / 局域网访问（如 `http://192.168.1.10:18953/onboard`）跳转后变成 `http://localhost:18953/chat/` → 直接失效
2. **port 用了用户输入的新值**：但 daemon 还没重启，新 port 上根本没服务，浏览器 ECONNREFUSED
3. **3 秒固定等待，不轮询服务可用性**：[`api.ts`](../web/src/lib/api.ts:176-186) 已经实现了 `waitForGateway`，这里没用

**修复**：
```ts
await saveConfig(...);
await restartDaemon();
const ok = await waitForGateway(15000);
if (!ok) { setLaunchError("daemon failed to start"); return; }
window.location.href = `${window.location.origin}/chat/`;  // 复用当前 origin
```

### Bug-9: 后端 handleSaveConfig 是"硬重置"覆盖写入

**位置**：[`handlers.go:444-466`](../internal/setup/handlers.go:444-466)

```go
cfg := &config.Config{ Providers: ..., Agents: ..., Channels: ..., Bindings: [] }
// 然后整个写回 pawnix.json
```

直接构造全新 cfg 写回磁盘，会**清空已有的**：其他 providers、其他 agents、cron jobs、storage（PG/SQLite 配置！）、hooks、bindings 等。如果用户已经配过 PG，再走一遍 onboard 直接把 DB 配置抹掉。

虽然正常用户路径下 `page.tsx` 会用 `status.configured` 拦截，但任何手动访问 `/onboard/` 的人会触发数据丢失。

**修复**：先 `config.Load()` 已有 cfg，只覆盖 onboard 涉及的字段，其它字段 merge 保留。

### Bug-10: Telegram 渲染缺失 → 死代码 bug

**位置**：[`OnboardConfig`](../web/src/app/onboard/page.tsx:56-69) + [`handleSaveConfig`](../internal/setup/handlers.go:476-485)

`OnboardConfig` 定义了 `telegramEnabled` / `telegramToken`，后端 `handleSaveConfig` 也有完整处理逻辑，**但 step 2 的 UI 里完全没渲染任何 Telegram 输入控件**。功能两端都准备好了，唯独中间没 UI——典型的"前端遗失"。

**修复**：在 step 2 加一个折叠区 "（可选）连接 Telegram bot"，含 enable toggle + token input + Test Bot 按钮。

---

## 🟡 P1：体验问题

### Bug-11: Step 2 标题叫 "Gateway"，内容却是 Agent + Personality

**位置**：[`STEP_LABELS`](../web/src/app/onboard/page.tsx:71-76) vs [`step === 2 内容`](../web/src/app/onboard/page.tsx:526-584)

标签 `Gateway` 让人以为配的是 host/port/网关相关，但实际包含 Agent Name、Port、Personality（SOUL.md），杂糅。

**修复**：要么拆成 "Agent Identity" + "Server" 两步，要么统一改名为 "Agent & Server"。

### Bug-12: Port 输入框无法清空

**位置**：[`page.tsx:551-552`](../web/src/app/onboard/page.tsx:551-552)

```ts
onChange={(e) => updateConfig({ port: parseInt(e.target.value) || 18953 })}
```

用户清空后立刻被改回 18953，无法编辑（必须先输新值再删旧值）。

**修复**：用受控字符串 + 提交时再 parse，或允许空但 Next 时校验。

### Bug-13: Port 冲突无检测

填了一个被占用的端口，后端 `handleSaveConfig` 不校验 → daemon restart 失败 → 用户卡死。

**修复**：`saveConfig` 前调一个 `POST /api/check-port?port=xxx`，端口被占就提示换端口。

### Bug-14: Welcome 步骤缺少 NavigationButtons，导航割裂

Step 0 用独立的 `Get Started` 按钮跳转，其他 step 用 `NavigationButtons`。后续步骤的 Back 按钮可以一路退到 step 0，但 step 0 的"Get Started"和 NavigationButtons 长得不一样，进退体验割裂。

**修复**：要么 step 0 也用 NavigationButtons，要么允许圆点点击直接跳到 step 0。

### Bug-15: 步骤进度圆点的可点击逻辑不直观

**位置**：[`page.tsx:242-251`](../web/src/app/onboard/page.tsx:242-251)

```ts
onClick={() => i < step && setStep(i)}
disabled={i > step}
```

`i === step` 时按钮 enabled 但 onClick 不响应，用户点了"当前步圆点"会以为按钮坏了。

**修复**：`i === step` 时也设 disabled，视觉上保持高亮。

### Bug-16: Provider 切换后 testError 残留

**位置**：[`handleProviderChange`](../web/src/app/onboard/page.tsx:156-170)

```ts
setTestStatus("idle");  // 没清 testError
```

Test 失败 → 切换 provider → testStatus 变 idle → 红字消失（因为 status=idle 不渲染）。这条勉强不显形，但 testError 应该和 testStatus 一起清，避免后续 race。

---

## 🟢 P2：打磨

### Bug-17: 全英文 UI，与中文目标用户脱节

整个 onboard 没有 i18n，按 README 中文定位看应该至少补一份中文文案或自动按浏览器语言切换。

### Bug-18: useRouter 导入但只放在依赖数组里

[`page.tsx:119,211`](../web/src/app/onboard/page.tsx:119)：`router` 只用在 `useCallback` 依赖里，从未实际调用 `router.push/replace`，跳转用的是 `window.location.href`。Next 路由和原生 location 混用会有一帧白屏。

**修复**：跳转改为 `router.push("/chat")` 即可。

### Bug-19: storage（DB）配置在 onboard 完全无入口

引入 PG 后用户期望能在 onboard 选存储后端，但 `OnboardConfig` 和 step 1/2/3 完全无 storage 字段，必须 onboard 完后再去 Settings 里改 → 启动时无法享受 DB 提供的 session/memory 持久化。

**修复**：在 step 2 或新加一步 "Storage" 加可选字段 "PostgreSQL DSN"，留空则用本地文件。

---

## 改进优先级建议

```
Phase 1 (修血): Bug-1, 2, 3, 4, 7, 9
  └ 涉及"骗用户" / "数据丢失" / "测了等于没测"
Phase 2 (修流): Bug-5, 6, 8, 10
  └ 涉及"功能能跑通" / "跨网络可用"
Phase 3 (修体验): Bug-11~16
Phase 4 (打磨): Bug-17~19
```

## 一句话总结

> 当前 onboard 是 "demo 跑通了就上线"的状态：模型预设是占位假数据、测试通过的字符串≠保存的字符串、saveConfig 失败时假装成功放彩带、跳转用 hardcode localhost、覆盖写会清空 DB 配置——**第一次接触产品的用户极容易在 onboard 里全绿通过然后打开 chat 看到错误**。优先修 Phase 1 的 6 条数据/逻辑 bug，其它可以按节奏来。
