# ILoveMath 题型配置文件格式

`backend/config/*.json` 中每个文件描述一类题型，由 `loader.go` 在启动时加载。
添加新题型只需新增 JSON 文件，**无需修改任何 Go 代码**。

---

## 顶层结构

```jsonc
{
  "ID": 3,           // 唯一整数，题型 ID，不可重复
  "Title": "时钟重合问题", // 显示名称
  "Items": [ /* ProblemItem[] */ ] // 用来描述题目基本类型
}
```

---

## ProblemItem 结构

```jsonc
{
  "Difficulty": 1,           // 难度等级：1=低 / 2=中 / 3=高
  "Question": "现在是{A}点整，", // 题干，{变量名} 会被替换为实际数值
  "Input": {                  // 变量声明，按依赖顺序求值
    "A": "random(1, 8)", // random 随机整数
    "B": "random(A + 1, 11)"
  },
  "Answer": [                 // 子问题列表，可多个，在生成题目随机选择一个
    {
      "text": "到{B}点整，时针分针共重合几次？", // 子题题面，支持 {变量名} 和特殊函数
      "value": "B - A"                          // 答案表达式，求值后为整数字符串
    }
  ],
  "Tips": "说明文字（不影响题目逻辑）",
  "Solution": {}  // 预留字段，当前未使用，保持空对象
}
```

---

## Input 变量系统

- **字面量**：`"A": "5"` → A = 5
- **随机整数**：`"A": "random(5, 20)"` → A ∈ [5, 20]（闭区间，整数）
- **表达式**：可引用已声明的变量，`"B": "A + 3"`
- **复合**：`"M": "random(1, 5 * A)"` — `random` 参数本身也可含表达式
- **求值顺序**：按 JSON 中键的出现顺序依次求值，因此**被依赖的变量必须先声明**

支持运算符：`+  -  *  /`（整数除法）及括号，**不支持浮点数**。

---

## {占位符} 替换规则

`Question`、`Answer[].text`、`Answer[].value` 中的 `{变量名}` 均会被替换为 `Input` 中对应变量的值，替换后再对 `value` 表达式求值。

---

## 特殊函数（用于 Answer.text）

| 语法 | 说明 |
|------|------|
| `{GenerateExpression(n)}` | 生成含 n 个运算符的四则运算式，带一个空格"A"，同时自动生成答案 |

当 `Answer[].text` 为特殊函数调用时，`Answer[].value` **留空或省略**，答案由函数自动生成。

---

## 约束与校验规则

1. `ID` 在所有 JSON 文件中唯一。
2. `Difficulty` 仅允许 1、2、3。
3. `Answer[].value` 求值结果必须为**整数**（不能含小数）。
4. `Input` 变量名区分大小写；建议大写字母（与题干 `{A}`、`{B}` 对应）。
5. `random(min, max)` 要求 min ≤ max，且结果集非空；若表达式依赖其他变量，确保依赖已在前面声明。
6. `Answer` 数组至少包含一个元素。
7. JSON 为标准格式，不支持注释；说明文字写入 `Tips` 字段。

---

## 典型示例

### 简单单答案题

```json
{
  "Difficulty": 1,
  "Question": "丁丁今年{a}岁，爸爸比他大{b}岁。",
  "Answer": [
    { "text": "爸爸今年多少岁？", "value": "a + b" }
  ],
  "Input": {
    "a": "random(5, 15)",
    "b": "random(25, 35)"
  },
  "Tips": "",
  "Solution": {}
}
```

### 多答案题

```json
{
  "Difficulty": 2,
  "Question": "爸爸妈妈年龄和是{a}岁，爸爸比妈妈大{c}岁。",
  "Answer": [
    { "text": "爸爸今年多少岁？", "value": "(a + c) / 2" },
    { "text": "妈妈今年多少岁？", "value": "(a - c) / 2" }
  ],
  "Input": {
    "x": "random(30, 50)",
    "c": "random(1, 8)",
    "y": "x - c",
    "a": "x + y"
  },
  "Tips": "",
  "Solution": {}
}
```

### 使用特殊函数

```json
{
  "Difficulty": 1,
  "Question": "如果下面算式成立，",
  "Answer": [
    { "text": "{GenerateExpression(4)}" }
  ],
  "Input": {},
  "Tips": "",
  "Solution": {}
}
```
