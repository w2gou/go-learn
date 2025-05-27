package main

import (
	"bufio"
	"fmt"
	"go-learn/weather/forecast/test1"
	"log"
	"net/http"
	"os"
	"strconv"
)

func main() {
	//startServer()
	//calculator.StartCalculator1()
	//calculator.StartCalculator2()
	//statistics.StartStatistics1()
	test1.StartWeatherForecast1()
}

func startServer() {
	//http.HandleFunc("/", handler)      // each request calls handler
	http.HandleFunc("/count", counter) // each request calls handler
	log.Fatal(http.ListenAndServe("localhost:8010", nil))
}

func handler(w http.ResponseWriter, r *http.Request) {
	fmt.Fprintf(w, "URL.Path = %q\n", r.URL.Path)
}

func counter(w http.ResponseWriter, r *http.Request) {
	// 1. 检查请求方法是否为POST
	if r.Method != http.MethodPost {
		http.Error(w, "只支持POST请求", http.StatusMethodNotAllowed)
		return
	}

	scanner := bufio.NewScanner(os.Stdin)

	if !scanner.Scan() {
		fmt.Println("错误：读取输入失败")
		return
	}

	// 检查参数数量是否足够
	if len(os.Args) < 3 {
		fmt.Println("用法：calc <命令> <数字1> <数字2> ...")
		fmt.Println("支持命令：sum, subtract, product, divide")
		os.Exit(1)
	}

	// 解析命令和参数
	cmd := os.Args[1]
	args := os.Args[2:]

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
