package statistics

import (
	"bufio"
	"fmt"
	"os"
	"strings"
)

func StartStatistics1() {
	fmt.Println("文件统计服务已启动！")
	for {
		fmt.Println("请输入文件路径：")

		input := getInput()
		if input != "" {
			lineCount, aCount, err := statisticsFile(input)
			if err == nil {
				fmt.Printf("文件共%d行\n", lineCount)
				fmt.Printf("文件共%d个a字母\n", aCount)
			}
		}
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

	parts := strings.Fields(input) // 自动处理多个空格

	// 检查参数数量是否足够
	if len(parts) != 1 {
		fmt.Println("用法：输入错误")
		return ""
	}

	return parts[0]
}

func statisticsFile(path string) (lineCount int, aCount int, err error) {
	file, err := os.Open(path)
	if err != nil {
		return 0, 0, err
	}
	defer file.Close()

	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		line := scanner.Text()
		lineCount++
		aCount += strings.Count(strings.ToLower(line), "a")
	}

	if err := scanner.Err(); err != nil {
		return 0, 0, err
	}

	return lineCount, aCount, nil
}
