# ILoveMath

小学数学练习网站。后端使用 Go（Gin 框架）提供 REST API 并渲染 HTML 模板，前端动画使用 TypeScript + PixiJS 编写，经 esbuild 编译后由 Go 服务。

---

## 目录结构

```
ILoveMath/
├── backend/
│   ├── main.go              # Go 服务入口
│   ├── go.mod / go.sum
│   ├── handlers/
│   │   └── index.go         # 页面路由 + REST API
│   ├── templates/
│   │   └── index.html       # Go html/template 模板
│   └── static/
│       ├── css/style.css    # 样式（Go 直接服务）
│       └── js/main.js       # TypeScript 编译输出（Go 直接服务）
├── frontend/
│   ├── src/
│   │   └── main.ts          # TypeScript 源码入口
│   ├── package.json
│   └── tsconfig.json
└── Makefile
```

---

## 环境要求

| 工具 | 版本要求 |
|------|----------|
| Go   | ≥ 1.22   |
| Node.js | ≥ 18  |
| npm  | ≥ 9     |

---

## 首次初始化

```bash
# 1. 安装 Go 依赖
cd backend
go mod tidy

# 2. 安装前端依赖
cd ../frontend
npm install
```

---

## TypeScript 编译

所有命令在 `frontend/` 目录下执行。

编译工具为 **esbuild**，将 `src/main.ts`（含 PixiJS 等依赖）打包成单文件，直接输出到 Go 的静态目录。

```bash
cd frontend

# 开发：编译一次
npm run build

# 开发：监听模式（保存 .ts 文件后自动重新编译）
npm run watch

# 生产：编译并压缩 (minify)
npm run build:prod

# 仅做类型检查，不输出文件
npm run typecheck
```

**编译输出路径：** `backend/static/js/main.js`

HTML 模板通过以下标签引用编译产物：

```html
<script src="/static/js/main.js"></script>
```

---

## 启动 Go 后端

```bash
cd backend

# 开发运行（不编译二进制）
go run .

# 或编译后运行
go build -o ilovmath .
./ilovmath
```

服务启动后访问：**http://localhost:8080**

---

## 完整开发流程

打开两个终端：

**终端 1 — 监听 TypeScript 变化**
```bash
cd frontend
npm run watch
```

**终端 2 — 启动 Go 服务**
```bash
cd backend
go run .
```

修改 `.ts` 文件保存后，esbuild 自动重新编译，刷新浏览器即可看到更新。

---

## REST API

| 方法 | 路径 | 说明 |
|------|------|------|
| GET | `/api/list` | 获取所有题目类型，首次调用自动创建 session |

请求时在 Header 中携带 session ID（由首次调用返回）：

```
X-Session-ID: <session_id>
```

---

## 添加新的 TypeScript 模块

1. 在 `frontend/src/` 下新建 `.ts` 文件，例如 `animations/fractal.ts`
2. 在 `frontend/package.json` 的 `scripts` 中添加编译命令：
   ```json
   "build:fractal": "esbuild src/animations/fractal.ts --bundle --outfile=../backend/static/js/fractal.js"
   ```
3. 在对应的 Go 模板中引用：
   ```html
   <script src="/static/js/fractal.js"></script>
   ```
