package math

import (
	"fmt"
	"math/rand"
	"strconv"
	"strings"
)

// replaceOneNumberWithA 核心逻辑：随机寻找一个数字并替换为 A
func replaceOneNumberWithA(expr string) (string, int) {
	// 1. 将算式按空格拆分成一个个标记（Token），比如 ["(", "12", "+", "5", ")", "*", "3"]
	tokens := strings.Fields(expr)

	// 2. 找出所有包含数字的 Token 的下标
	var numIndices []int
	for i, token := range tokens {
		if hasDigit(token) {
			numIndices = append(numIndices, i)
		}
	}

	// 3. 从数字下标列表中随机选一个
	idx := numIndices[rand.Intn(len(numIndices))]
	targetToken := tokens[idx]

	// 4. 处理 Token。因为数字可能和括号连在一起，比如 "(12" 或 "5)"
	var numStr string
	var prefix, suffix string

	for i, r := range targetToken {
		if r >= '0' && r <= '9' {
			numStr += string(r) // 提取数字部分
		} else if len(numStr) == 0 {
			prefix += string(r) // 数字之前的符号（如左括号）
		} else {
			suffix = targetToken[i:] // 数字之后的符号（如右括号）
			break
		}
	}

	// 5. 将提取出的数字字符串转回 int，作为 A 的真值
	val, _ := strconv.Atoi(numStr)

	// 6. 用 "A" 替换掉原来的数字，保留前缀和后缀
	tokens[idx] = prefix + "A" + suffix

	// 7. 重新组合成字符串并返回
	return strings.Join(tokens, " "), val
}

// 辅助函数：判断字符串中是否包含数字字符
func hasDigit(s string) bool {
	for _, r := range s {
		if r >= '0' && r <= '9' {
			return true
		}
	}
	return false
}

// GenerateExpression 生成符合条件的算式
func GenerateExpression(args ...string) (string, int) {
	numCount, _ := strconv.Atoi(args[0])
	for {
		expr, result, count, bracketCount := generate(numCount, 0, 0)
		// 最终检查：数字个数达标，且括号不超过3个
		if count >= numCount && bracketCount <= 3 {
			expr, replace := replaceOneNumberWithA(expr)
			return expr + " =" + strconv.Itoa(result), replace
		}
	}
}

// generate 递归构造函数
// 返回值：算式字符串, 计算结果, 包含的数字个数, 已使用的括号数
func generate(limit int, depth int, bracketUsed int) (string, int, int, int) {
	// 递归基：如果只需要1个数字或达到深度限制
	if limit <= 1 {
		num := rand.Intn(100) + 1 // 默认生成1-100
		return fmt.Sprintf("%d", num), num, 1, bracketUsed
	}

	// 拆分左右子树的数字个数
	leftLimit := rand.Intn(limit-1) + 1
	rightLimit := limit - leftLimit

	op := []string{"+", "-", "*", "÷"}[rand.Intn(4)]

	leftStr, leftVal, leftCount, bLeft := generate(leftLimit, depth+1, bracketUsed)
	rightStr, rightVal, rightCount, bRight := generate(rightLimit, depth+1, bLeft)

	// 规则校验与修正
	switch op {
	case "+":
		// 规则2: 加法操作数不超过999
		if leftVal > 999 || rightVal > 999 {
			return generate(limit, depth, bracketUsed)
		}
	case "-":
		// 规则1: 全是正整数（结果不能为负或0）
		if leftVal <= rightVal {
			return generate(limit, depth, bracketUsed)
		}
	case "*":
		// 规则3: 乘法只能是一个个位数 * 两位数
		isL1R2 := (leftVal < 10 && rightVal < 100)
		isL2R1 := (leftVal < 100 && rightVal < 10)
		if !isL1R2 && !isL2R1 {
			return generate(limit, depth, bracketUsed)
		}
	case "/":
		// 规则1: 整除且结果为正整数
		if rightVal == 0 || leftVal%rightVal != 0 || leftVal/rightVal == 0 {
			return generate(limit, depth, bracketUsed)
		}
	}

	finalVal := calculate(leftVal, rightVal, op)

	// 规则5: 增加括号逻辑（随机嵌套，总数不超3）
	currentStr := ""
	newBracketCount := bRight
	// 只有当子树是表达式时，才随机加括号
	if (leftCount > 1 || rightCount > 1) && bRight < 3 && rand.Float32() < 0.3 {
		currentStr = fmt.Sprintf("(%s %s %s)", leftStr, op, rightStr)
		newBracketCount++
	} else {
		currentStr = fmt.Sprintf("%s %s %s", leftStr, op, rightStr)
	}

	return currentStr, finalVal, leftCount + rightCount, newBracketCount
}

func calculate(a, b int, op string) int {
	switch op {
	case "+":
		return a + b
	case "-":
		return a - b
	case "*":
		return a * b
	case "/":
		return a / b
	}
	return 0
}
