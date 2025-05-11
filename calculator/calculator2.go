package calculator

import (
	"bufio"
	"fmt"
	"os"
	"strconv"
	"strings"
	"unicode"
)

func StartCalculator2() {
	fmt.Println("计算器服务已启动！")
	for {
		fmt.Println("请输入命令：")
		input := getInput()
		expression, err := evaluateExpression(input)
		if err != nil {
			fmt.Printf("计算有误，请检查输入表达式：%s\n", input)
			continue
		}
		fmt.Printf("计算结果：%d\n", expression)
	}
}

func getInput() string {
	scanner := bufio.NewScanner(os.Stdin)

	if !scanner.Scan() {
		fmt.Println("错误：读取输入失败")
		return ""
	}

	input := strings.TrimSpace(scanner.Text())
	if input == "" {
		fmt.Println("错误：输入为空")
		return ""
	}

	return input
}

// 主入口：计算表达式
func evaluateExpression(expr string) (int, error) {
	tokens := tokenize(expr)
	postfix, err := infixToPostfix(tokens)
	if err != nil {
		return 0, err
	}
	return evalPostfix(postfix)
}

// Tokenize 词法分析：将字符串拆成 token，支持负数前缀和 -(
func tokenize(s string) []string {
	var tokens []string
	i := 0
	for i < len(s) {
		ch := s[i]
		if unicode.IsSpace(rune(ch)) {
			i++
			continue
		}
		if ch == '-' {
			// 负号逻辑
			if i == 0 || isOperatorOrParen(tokens[len(tokens)-1]) {
				// 一元负号
				if i+1 < len(s) && unicode.IsDigit(rune(s[i+1])) {
					j := i + 1
					for j < len(s) && unicode.IsDigit(rune(s[j])) {
						j++
					}
					tokens = append(tokens, s[i:j])
					i = j
					continue
				} else if i+1 < len(s) && s[i+1] == '(' {
					// - ( 开头，视为 -1 * (
					tokens = append(tokens, "-1", "*")
					i++ // 跳过 '-'，下轮处理 (
					continue
				}
			}
			// 二元减号
			tokens = append(tokens, string(ch))
			i++
		} else if unicode.IsDigit(rune(ch)) {
			j := i
			for j < len(s) && unicode.IsDigit(rune(s[j])) {
				j++
			}
			tokens = append(tokens, s[i:j])
			i = j
		} else if strings.ContainsRune("()+*/", rune(ch)) {
			tokens = append(tokens, string(ch))
			i++
		} else {
			panic(fmt.Sprintf("invalid character: %c", ch))
		}
	}
	return tokens
}

func isOperatorOrParen(token string) bool {
	return token == "" || token == "+" || token == "-" || token == "*" || token == "/" || token == "("
}

// 中缀转后缀表达式（Shunting Yard 算法）
func infixToPostfix(tokens []string) ([]string, error) {
	var output []string
	var stack []string
	precedence := map[string]int{
		"+": 1, "-": 1,
		"*": 2, "/": 2,
	}
	for _, token := range tokens {
		switch {
		case isNumber(token):
			output = append(output, token)
		case token == "(":
			stack = append(stack, token)
		case token == ")":
			for len(stack) > 0 && stack[len(stack)-1] != "(" {
				output = append(output, stack[len(stack)-1])
				stack = stack[:len(stack)-1]
			}
			if len(stack) == 0 || stack[len(stack)-1] != "(" {
				return nil, fmt.Errorf("mismatched parentheses")
			}
			stack = stack[:len(stack)-1] // 弹出 '('
		default: // 运算符
			for len(stack) > 0 &&
				stack[len(stack)-1] != "(" &&
				precedence[stack[len(stack)-1]] >= precedence[token] {
				output = append(output, stack[len(stack)-1])
				stack = stack[:len(stack)-1]
			}
			stack = append(stack, token)
		}
	}
	// 弹出剩余操作符
	for len(stack) > 0 {
		if stack[len(stack)-1] == "(" {
			return nil, fmt.Errorf("mismatched parentheses")
		}
		output = append(output, stack[len(stack)-1])
		stack = stack[:len(stack)-1]
	}
	return output, nil
}

// 计算后缀表达式
func evalPostfix(postfix []string) (int, error) {
	var stack []int
	for _, token := range postfix {
		if isNumber(token) {
			val, _ := strconv.Atoi(token)
			stack = append(stack, val)
		} else {
			if len(stack) < 2 {
				return 0, fmt.Errorf("invalid expression")
			}
			b := stack[len(stack)-1]
			a := stack[len(stack)-2]
			stack = stack[:len(stack)-2]
			switch token {
			case "+":
				stack = append(stack, a+b)
			case "-":
				stack = append(stack, a-b)
			case "*":
				stack = append(stack, a*b)
			case "/":
				if b == 0 {
					return 0, fmt.Errorf("division by zero")
				}
				stack = append(stack, a/b)
			default:
				return 0, fmt.Errorf("unknown operator: %s", token)
			}
		}
	}
	if len(stack) != 1 {
		return 0, fmt.Errorf("invalid expression")
	}
	return stack[0], nil
}

func isNumber(s string) bool {
	_, err := strconv.Atoi(s)
	return err == nil
}
