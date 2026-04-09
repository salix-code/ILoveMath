package math

import (
	"fmt"
	"go/ast"
	"go/parser"
	"go/token"
	"ilovmath/config"
	"math/rand"
	"sort"
	"strconv"
	"strings"
)

// resolveInput evaluates each Input expression to produce a concrete int map.
// Expressions may reference other Input keys using bare names (e.g. "b + 5").
// A retry loop keeps evaluating until all keys are resolved or no further
// progress can be made (circular or undefined dependency).
func resolveInput(input map[string]string) (map[string]int, error) {
	resolved := make(map[string]int, len(input))

	for {
		progress := false
		for k, expr := range input {
			if _, done := resolved[k]; done {
				continue
			}
			// evalExpr looks up unresolved Ident nodes in resolved;
			// if a dependency is not yet present evalNode returns an error
			// and we retry next round.
			if v, err := evalExpr(expr, resolved); err == nil {
				resolved[k] = v
				progress = true
			}
		}
		if !progress {
			break
		}
	}

	if len(resolved) != len(input) {
		unresolved := make([]string, 0, len(input)-len(resolved))
		for k := range input {
			if _, ok := resolved[k]; !ok {
				unresolved = append(unresolved, k)
			}
		}
		sort.Strings(unresolved)
		return nil, fmt.Errorf("could not resolve input keys (circular or undefined dependency): %v", unresolved)
	}

	return resolved, nil
}

// substituteQuestion replaces {key} tokens in s with their resolved integer values.
func substituteQuestion(s string, vars map[string]int) string {
	for k, v := range vars {
		s = strings.ReplaceAll(s, "{"+k+"}", strconv.Itoa(v))
	}
	return s
}

// AnswerItemResult represents a single part of an answer with its resolved value.
type AnswerItemResult struct {
	Text  string `json:"text"`
	Value string `json:"value"`
}

// evalAnswer computes the expected answer string for a question.
//
//   - pure arithmetic expression "a + b" or "{a}+{b}" → evaluated integer string
//   - mixed text template "小明{a}岁，爷爷{b}岁" → token-substituted string
//   - empty string → "" (open-ended question, not validated)
func evalAnswer(items []config.AnswerItem, vars map[string]int) []AnswerItemResult {
	var results []AnswerItemResult
	for _, item := range items {
		val := ""
		if item.Value == "" {
			val = ""
		} else if v, err := evalExpr(item.Value, vars); err == nil {
			val = strconv.Itoa(v)
		} else {
			val = substituteQuestion(item.Value, vars)
		}
		results = append(results, AnswerItemResult{
			Text:  item.Text,
			Value: val,
		})
	}
	return results
}

// evalExpr evaluates an arithmetic expression string using Go's parser.
//
// Supported syntax (a strict subset of valid Go expressions):
//
//	integer literals        42
//	variable names          b  (looked up in vars)
//	binary operators        + - * /   (standard Go precedence)
//	unary minus             -expr
//	parentheses             (expr)
//	random function         random(min, max)   inclusive on both ends
func evalExpr(expr string, vars map[string]int) (int, error) {
	s := strings.TrimSpace(expr)
	tree, err := parser.ParseExpr(s)
	if err != nil {
		return 0, fmt.Errorf("parse %q: %w", s, err)
	}
	return evalNode(tree, vars)
}

// evalNode recursively evaluates a Go AST expression node.
func evalNode(node ast.Expr, vars map[string]int) (int, error) {
	switch n := node.(type) {

	case *ast.Ident:
		if vars == nil {
			return 0, fmt.Errorf("undefined variable %q", n.Name)
		}
		v, ok := vars[n.Name]
		if !ok {
			return 0, fmt.Errorf("undefined variable %q", n.Name)
		}
		return v, nil

	case *ast.BasicLit:
		if n.Kind != token.INT {
			return 0, fmt.Errorf("unsupported literal kind %s", n.Kind)
		}
		return strconv.Atoi(n.Value)

	case *ast.UnaryExpr:
		v, err := evalNode(n.X, vars)
		if err != nil {
			return 0, err
		}
		if n.Op == token.SUB {
			return -v, nil
		}
		return v, nil

	case *ast.BinaryExpr:
		left, err := evalNode(n.X, vars)
		if err != nil {
			return 0, err
		}
		right, err := evalNode(n.Y, vars)
		if err != nil {
			return 0, err
		}
		switch n.Op {
		case token.ADD:
			return left + right, nil
		case token.SUB:
			return left - right, nil
		case token.MUL:
			return left * right, nil
		case token.QUO:
			if right == 0 {
				return 0, fmt.Errorf("division by zero")
			}
			return left / right, nil
		}
		return 0, fmt.Errorf("unsupported operator %s", n.Op)

	case *ast.ParenExpr:
		return evalNode(n.X, vars)

	case *ast.CallExpr:
		fn, ok := n.Fun.(*ast.Ident)
		if !ok || fn.Name != "random" {
			return 0, fmt.Errorf("unsupported function call")
		}
		if len(n.Args) != 2 {
			return 0, fmt.Errorf("random() requires exactly 2 arguments")
		}
		minVal, err := evalNode(n.Args[0], vars)
		if err != nil {
			return 0, err
		}
		maxVal, err := evalNode(n.Args[1], vars)
		if err != nil {
			return 0, err
		}
		if minVal > maxVal {
			minVal, maxVal = maxVal, minVal
		}
		return minVal + rand.Intn(maxVal-minVal+1), nil
	}

	return 0, fmt.Errorf("unsupported AST node %T", node)
}
