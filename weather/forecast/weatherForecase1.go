package forecast

import (
	"bufio"
	"fmt"
	"io"
	"net/http"
	"os"
	"strings"
)

func StartWeatherForecast1() {
	fmt.Println("计算器服务已启动！")
	for {
		fmt.Println("请输入命令：")
		sheng, place, err := getInputs()
		if err != nil {
			getData(sheng, place)
		}
	}
}

func getInputs() (sheng string, place string, err error) {
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

func getData(sheng string, place string) {
	id := "88888888"
	key := "88888888"
	basicUrl := "https://cn.apihz.cn/api/tianqi/tqyb.php"
	url := basicUrl + "?id=" + id + "&key=" + key + "&sheng=" + sheng + "&place=" + place

	resp, err := http.Get(url)
	if err != nil {
		fmt.Println("请求失败:", err)
		return
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		fmt.Println("读取响应失败:", err)
		return
	}

	fmt.Printf("状态码: %d\n", resp.StatusCode)
	fmt.Println("响应内容:")
	fmt.Println(string(body))
}
