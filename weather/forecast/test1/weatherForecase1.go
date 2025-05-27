package test1

import (
	"bufio"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"strings"
)

const (
	GetProvinces = "provinces"
	GetCities    = "cities"
	GetZones     = "zones"
	GetWeather   = "weather"
)

var commandMap = map[string]func(args []string){
	GetProvinces: func(args []string) {
		handleA() // 无参数函数
	},
	GetCities: func(args []string) {
		handleB(args) // 传入参数
	},
	GetZones: func(args []string) {
		handleC(args) // 传入参数
	},
	GetWeather: func(args []string) {
		handleD(args) // 传入参数
	},
}

var provinceMap map[string]string
var cityMap map[string]string
var zoneMap map[string]string

var provinceCode string
var cityCode string

func StartWeatherForecast1() {
	fmt.Println("天气服务已启动！")
	fmt.Println("请输入命令：")
	for {
		inputs, err := getInputs()
		if err == nil {
			if handler, ok := commandMap[inputs[0]]; ok {
				handler(inputs[1:])
			}
		}
	}
}

func getInputs() (inputs []string, err error) {
	scanner := bufio.NewScanner(os.Stdin)

	if !scanner.Scan() {
		fmt.Println("错误：读取输入失败")
		return nil, fmt.Errorf("错误：读取输入失败")
	}

	input := strings.TrimSpace(scanner.Text())
	if input == "" {
		fmt.Println("错误：输入为空")
		return nil, fmt.Errorf("错误：输入为空")
	}

	parts := strings.Fields(input) // 自动处理多个空格

	// 检查参数数量是否足够
	//if len(parts) != 2 {
	//	return nil, "", fmt.Errorf("错误：参数数量有误")
	//}

	return parts, nil
}

// api1
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

// 气象局api
// 获取省
func getProvinces() error {
	if provinceMap != nil && len(provinceMap) > 0 {
		return nil
	}

	//url := "https://www.weather.com.cn/data/city3jdata/china.html"
	url := "http://www.nmc.cn/rest/province"

	resp, err := http.Get(url)
	if err != nil {
		return fmt.Errorf("请求失败: %w", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return fmt.Errorf("读取响应失败: %w", err)
	}

	var provinces []Province
	err = json.Unmarshal(body, &provinces)
	if err != nil {
		return fmt.Errorf("解析失败: %w", err)
	}

	provinceMap = make(map[string]string)
	for _, province := range provinces {
		provinceMap[province.Code] = province.Name
	}

	//err = json.Unmarshal(body, &provinceMap)
	//if err != nil {
	//	return fmt.Errorf("解析失败: %w", err)
	//}

	//fmt.Printf("状态码: %d\n", resp.StatusCode)
	//fmt.Println("响应内容:")
	//fmt.Println(string(body))

	return nil
}

// 气象局api
// 获取市
func getCities(newCode string) error {
	if provinceCode == newCode {
		return nil
	} else {
		provinceCode = newCode
	}

	baseUrl := "http://www.weather.com.cn/data/city3jdata/provshi/"
	url := baseUrl + provinceCode + ".html"
	url = "http://www.nmc.cn/rest/province/" + provinceCode

	resp, err := http.Get(url)
	if err != nil {
		return fmt.Errorf("请求失败: %w", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return fmt.Errorf("读取响应失败: %w", err)
	}

	var cities []City
	err = json.Unmarshal(body, &cities)
	if err != nil {
		return fmt.Errorf("解析失败: %w", err)
	}

	cityMap = make(map[string]string)
	for _, city := range cities {
		cityMap[city.Code] = city.City
	}

	//err = json.Unmarshal(body, &cityMap)
	//if err != nil {
	//	return fmt.Errorf("解析失败: %w", err)
	//}

	return nil
}

// 气象局api
// 获取区
func getZones(newCode string) error {
	if cityCode == newCode {
		return nil
	} else {
		cityCode = newCode
	}

	baseUrl := "http://www.weather.com.cn/data/city3jdata/station/"
	url := baseUrl + provinceCode + cityCode + ".html"

	resp, err := http.Get(url)
	if err != nil {
		return fmt.Errorf("请求失败: %w", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return fmt.Errorf("读取响应失败: %w", err)
	}

	err = json.Unmarshal(body, &zoneMap)
	if err != nil {
		return fmt.Errorf("解析失败: %w", err)
	}

	return nil
}

// 气象局api
// 获取区天气数据
func getWeatherData(newCode string) error {
	baseUrl := "http://m.weather.com.cn/data/"
	url := baseUrl + provinceCode + cityCode + newCode + ".html"

	resp, err := http.Get(url)
	if err != nil {
		return fmt.Errorf("请求失败: %w", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return fmt.Errorf("读取响应失败: %w", err)
	}

	// 解析数据
	err = json.Unmarshal(body, &zoneMap)
	if err != nil {
		return fmt.Errorf("解析失败: %w", err)
	}

	fmt.Printf("状态码: %d\n", resp.StatusCode)
	fmt.Println("响应内容:")
	fmt.Println(string(body))

	return nil
}

func handleA() {
	err := getProvinces()
	if err != nil {
		println(err)
		return
	}

	var values []string
	for k, v := range provinceMap {
		values = append(values, v+"("+k+")")
	}

	// 输出所有值，空格分隔
	fmt.Println(strings.Join(values, " "))
}

func handleB(inputs []string) {
	if inputs == nil || len(inputs) == 0 {
		fmt.Println("错误：缺少参数，城市代码")
	}

	err := getCities(inputs[0])
	if err != nil {
		println(err)
		return
	}

	var values []string
	for k, v := range cityMap {
		values = append(values, v+"("+k+")")
	}

	// 输出所有值，空格分隔
	fmt.Println(strings.Join(values, " "))
}

func handleC(inputs []string) {
	if inputs == nil || len(inputs) == 0 {
		fmt.Println("错误：缺少参数，城市代码")
	}

	err := getZones(inputs[0])
	if err != nil {
		println(err)
		return
	}

	var values []string
	for k, v := range zoneMap {
		values = append(values, v+"("+k+")")
	}

	// 输出所有值，空格分隔
	fmt.Println(strings.Join(values, " "))
}

func handleD(inputs []string) {
	if inputs == nil || len(inputs) == 0 {
		fmt.Println("错误：缺少参数，城市代码")
	}

	err := getWeatherData(inputs[0])
	if err != nil {
		println(err)
		return
	}

	//var values []string
	//for k, v := range zoneMap {
	//	values = append(values, v+"("+k+")")
	//}
	//
	//// 输出所有值，空格分隔
	//fmt.Println(strings.Join(values, " "))
}
