# Snapshot 与 Refs

给 AI agent 用的紧凑元素引用系统,高效定位页面元素。

> 命令前缀同样是 `podman exec claw-browser camoufox-cli ...`,本文为简洁省略。

**Related**: [commands.md](commands.md) — 完整命令参考。

## Refs 原理

传统方式:
```
完整 DOM/HTML → AI 解析 → CSS selector → 动作(~3000-5000 tokens)
```

camoufox-cli 方式:
```
紧凑 aria 快照 → 分配 @refs → 直接交互(~200-400 tokens)
```

Refs 是顺序编号(`@e1`、`@e2`…),每次 snapshot 按 DOM 遍历顺序分配。页面内容
变了,同一元素可能拿到不同的 ref。

## Snapshot 命令

```bash
camoufox-cli snapshot              # 完整页面结构
camoufox-cli snapshot -i           # 仅交互元素(推荐)
camoufox-cli snapshot -s "#main"   # 限定到容器
camoufox-cli snapshot -i -s "form.login"
```

### 输出格式

```
- heading "Example Domain" [ref=e1]
- navigation
  - link "Home" [ref=e2]
  - link "Products" [ref=e3]
- button "Sign In" [ref=e5]
- main
  - textbox "Email" [ref=e7]
  - textbox "Password" [ref=e8]
  - button "Log In" [ref=e9]
```

`-i`(仅交互)会过滤掉非交互元素,只留 link/button/textbox 等。

## 用 Refs 交互

```bash
camoufox-cli click @e5            # 点 "Sign In"
camoufox-cli fill @e7 "user@x.com"
camoufox-cli fill @e8 "password"
camoufox-cli click @e9            # 提交
```

## Ref 生命周期(重要)

**Refs 在页面变化时失效!** 它们绑定特定页面状态。页面变化后旧 ref 可能指向
错误元素或不存在。

```bash
camoufox-cli snapshot -i          # - button "Next" [ref=e1]
camoufox-cli click @e1            # 触发页面跳转
camoufox-cli snapshot -i          # 必须重新 snapshot!ref 含义已变
```

### 什么会让 ref 失效

- **导航**:点链接、表单提交、重定向
- **动态内容**:下拉展开、模态弹出、AJAX 更新
- **滚动**:若触发懒加载
- **JS**:任何改变元素顺序的 DOM 变更

### 用了失效 ref 会怎样

- 指向**不同元素**(点错目标)
- ref **不存在**(报错)
- **静默作用在错误元素上**(最坑)

拿不准就重新 snapshot。

## 最佳实践

1. **交互前先 snapshot** — 没 snapshot 就用 @ref 必然失败
2. **导航后重新 snapshot** — 点了跳转链接后旧 ref 全废
3. **动态变化后重新 snapshot** — 下拉/模态出现后
4. **复杂页面 scope snapshot** — `snapshot -i -s "#login-form"` 减噪
5. **慢页面先 wait 再 snapshot** — `wait 3000 && snapshot -i`

## 重复元素(同 role + name)

多个元素同 role 同 name(如两个 "Submit" 按钮),各拿独立 ref,snapshot 用
`[nth=N]` 区分:

```
- button "Submit" [ref=e3]
- button "Submit" [ref=e7] [nth=1]
```

两个 ref 都能独立用,选对应你要的那个。
