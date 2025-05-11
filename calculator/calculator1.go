package calculator

import (
	"bufio"
	"fmt"
	"os"
	"strconv"
	"strings"
)

func StartCalculator1() {
	fmt.Println("计算器服务已启动！")
	for {
		fmt.Println("请输入命令：")
		inputs := getInputs()
		if inputs != nil {
			caseCommandForCalculator(inputs)
		}
	}
}

func getInputs() []string {
	scanner := bufio.NewScanner(os.Stdin)

	if !scanner.Scan() {
		fmt.Println("错误：读取输入失败")
		return nil
	}

	input := strings.TrimSpace(scanner.Text())
	if input == "" {
		fmt.Println("错误：输入为空")
		return nil
	}

	parts := strings.Fields(input) // 自动处理多个空格

	// 检查参数数量是否足够
	if len(parts) != 3 {
		fmt.Println("用法：calc <命令> <数字1> <数字2> ...")
		fmt.Println("支持命令：sum, subtract, product, divide")
		os.Exit(1)
	}

	return parts
}

func caseCommandForCalculator(inputs []string) {
	// 解析命令和参数
	cmd := inputs[0]
	args := inputs[1:]

	// 将参数转为浮点数
	var numbers []float64
	for _, arg := range args {
		num, err := strconv.ParseFloat(arg, 64)
		if err != nil {
			fmt.Printf("错误：'%s' 不是有效数字\n", arg)
			os.Exit(1)
		}
		numbers = append(numbers, num)
	}

	// 根据命令执行计算
	switch cmd {
	case "sum":
		result := 0.0
		for _, num := range numbers {
			result += num
		}
		fmt.Printf("结果：%v\n", result)

	case "subtract":
		if len(numbers) < 1 {
			fmt.Println("错误：减法需要至少一个数字")
			os.Exit(1)
		}
		result := numbers[0]
		for _, num := range numbers[1:] {
			result -= num
		}
		fmt.Printf("结果：%v\n", result)

	case "product":
		result := 1.0
		for _, num := range numbers {
			result *= num
		}
		fmt.Printf("结果：%v\n", result)

	case "divide":
		if len(numbers) < 1 {
			fmt.Println("错误：除法需要至少一个数字")
			os.Exit(1)
		}
		result := numbers[0]
		for _, num := range numbers[1:] {
			if num == 0 {
				fmt.Println("错误：除数不能为零")
				os.Exit(1)
			}
			result /= num
		}
		fmt.Printf("结果：%v\n", result)

	default:
		fmt.Printf("错误：不支持的命令 '%s'\n", cmd)
		os.Exit(1)
	}
}
