# ILoveMath — Project Guidelines

小学数学练习网站。Go (Gin) 后端 + TypeScript/PixiJS 前端，esbuild 打包，前端产物直接由 Go 服务。

## Build & Run

**开发（两个终端并行）**

```bash
# 终端 1：监听 TypeScript 变化并自动重新打包
make watch          # cd frontend && npm run watch

# 终端 2：启动 Go 服务（port 8080）
make backend        # cd backend && go run .
```

访问 http://localhost:8080

**其他常用命令**

```bash
make typecheck      # 仅做 TS 类型检查，不输出文件
make build          # 生产打包（JS minify + go build -o ilovmath .）
make run            # 生产打包 + 启动二进制
```

> Go 后端必须在 `backend/` 下用 `go run .`，不能用 `go run main.go`（多文件包）。

## Architecture

```
frontend/src/*.ts  →(esbuild)→  backend/static/js/main.js
                                        ↓
                              Go (Gin) 服务 :8080
                                        ↓
                              html/template 渲染
```

- **前端 → 后端通信**：REST JSON，所有请求带 `X-Session-ID` 请求头（从 `localStorage.ilovmath_session_id` 读取）。
- **Session**：`GET /api/list` 初始化（若无 ID 则新建），返回 `session_id`。
- **题目流程**：`POST /api/question/start`（选题型/难度）→ `POST /api/question/next`（循环：提交上题答案 + 获取下题）。

**API 端点**

| Method | Path | 说明 |
|--------|------|------|
| `GET` | `/api/list` | 初始化 session，返回题型列表 |
| `POST` | `/api/question/start` | `{id, difficulty}` 开始练习，返回 `{redirect}` |
| `POST` | `/api/question/next` | `{prev_guid, prev_answer}` 提交答案 + 获取下题 |

## Config-Driven 题目系统

题型配置位于 `backend/config/*.json`，`loader.go` 启动时全部加载。

**配置关键字段**

```jsonc
{
  "id": 1,
  "title": "年龄问题",
  "items": [{
    "difficulty": 2,
    "question": "丁丁今年{a}岁，爸爸比他大{b}岁。...",
    "input": { "a": "random(5,20)", "b": "random(25,35)" },
    "answer": [{ "text": "爸爸的年龄是多少岁", "value": "{a}+{b}" }]
  }]
}
```

- `input` 中的变量按依赖顺序求值，支持 `random(min, max)` 与表达式（见 `backend/math/expr.go`）。
- `{key}` 占位符在 `question` 和 `answer.value` 中替换后再求值。
- 特殊函数 `GenerateExpression(n)` 生成带空格 "A" 的四则运算式（`n` = 运算符数量）。
- 表达式解析基于 Go `go/parser`（AST），支持 `+ - * /`、括号、变量、`random()`，**不支持浮点数**。

**添加新题型**：在 `backend/config/` 新增 `*.json`，无需改代码。

## UI / Styling

- **CSS 框架**：Tailwind CSS v4 + DaisyUI v5，通过 CDN 引入（`index.html` 和 `question.html`）。
- **主题**：`data-theme="fantasy"`（在 `<html>` 标签上），可改为其他 DaisyUI 主题（如 `cupcake`、`garden`）。
- **类名约定**：优先使用 DaisyUI 语义类（`btn btn-primary`、`card`、`badge` 等），自定义样式写在 `frontend/src/style.css`。
- **打印页**（`paper.html`）：不引入 DaisyUI，保留独立内联 CSS，避免干扰打印排版。
- **迁移到 npm**：将来只需删除 CDN 标签、`npm install tailwindcss daisyui`、用 Tailwind CLI 编译 CSS，类名无需改动。

## Conventions

- **Session 存储**：`sync.Map`（内存），服务重启后丢失；`CurrentGUID` 防重放攻击，每题只能提交一次。
- **答案校验**：trim 后精确字符串匹配，题目答案为整数字符串。
- **静态文件缓存**：开发模式下 `Cache-Control: no-cache`（`main.go` 写死，生产需手动改）。
- **PixiJS**：已在 `package.json` 声明但当前代码中尚未使用，为动画功能预留。
- **前端入口**：`frontend/src/main.ts`（首页）和 `frontend/src/question.ts`（题目页），esbuild 分别打包输出 `backend/static/js/main.js` 和 `question.js`。
- 无自动化测试（无 `*_test.go` 或 `*.test.ts`）。
