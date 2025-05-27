package test1

type Province struct {
	Code string `json:"code"`
	Name string `json:"name"`
	URL  string `json:"url"`
}

type City struct {
	Code     string `json:"code"`
	Province string `json:"province"`
	City     string `json:"city"`
	URL      string `json:"url"`
}

type Response struct {
	Msg  string `json:"msg"`
	Code int    `json:"code"`
	Data Data   `json:"data"`
}

type Data struct {
	Real        RealWeather        `json:"real"`
	Predict     PredictWeather     `json:"predict"`
	Air         string             `json:"air"`
	TempChart   []TempChart        `json:"tempchart"`
	PassedChart []PassedChartEntry `json:"passedchart"`
}

type RealWeather struct {
	Station       Station   `json:"station"`
	PublishTime   string    `json:"publish_time"`
	Weather       Weather   `json:"weather"`
	Wind          Wind      `json:"wind"`
	Warn          Warn      `json:"warn"`
	SunriseSunset SunPeriod `json:"sunriseSunset"`
}

type PredictWeather struct {
	Station     Station       `json:"station"`
	PublishTime string        `json:"publish_time"`
	Detail      []ForecastDay `json:"detail"`
}

type Station struct {
	Code     string `json:"code"`
	Province string `json:"province"`
	City     string `json:"city"`
	URL      string `json:"url"`
}

type Weather struct {
	Temperature     float64 `json:"temperature"`
	TemperatureDiff float64 `json:"temperatureDiff"`
	AirPressure     float64 `json:"airpressure"`
	Humidity        float64 `json:"humidity"`
	Rain            float64 `json:"rain"`
	RComfort        int     `json:"rcomfort"`
	IComfort        int     `json:"icomfort"`
	Info            string  `json:"info"`
	Img             string  `json:"img"`
	FeelsLike       float64 `json:"feelst"`
}

type Wind struct {
	Direct string  `json:"direct"`
	Degree float64 `json:"degree"`
	Power  string  `json:"power"`
	Speed  float64 `json:"speed"`
}

type Warn struct {
	Alert        string `json:"alert"`
	Pic          string `json:"pic"`
	Province     string `json:"province"`
	City         string `json:"city"`
	URL          string `json:"url"`
	IssueContent string `json:"issuecontent"`
	FMeans       string `json:"fmeans"`
	SignalType   string `json:"signaltype"`
	SignalLevel  string `json:"signallevel"`
	Pic2         string `json:"pic2"`
}

type SunPeriod struct {
	Sunrise string `json:"sunrise"`
	Sunset  string `json:"sunset"`
}

type ForecastDay struct {
	Date          string       `json:"date"`
	Pt            string       `json:"pt"`
	Day           ForecastPart `json:"day"`
	Night         ForecastPart `json:"night"`
	Precipitation float64      `json:"precipitation"`
}

type ForecastPart struct {
	Weather WeatherSimple `json:"weather"`
	Wind    WindSimple    `json:"wind"`
}

type WeatherSimple struct {
	Info        string `json:"info"`
	Img         string `json:"img"`
	Temperature string `json:"temperature"`
}

type WindSimple struct {
	Direct string `json:"direct"`
	Power  string `json:"power"`
}

type TempChart struct {
	Time      string  `json:"time"`
	MaxTemp   float64 `json:"max_temp"`
	MinTemp   float64 `json:"min_temp"`
	DayImg    string  `json:"day_img"`
	DayText   string  `json:"day_text"`
	NightImg  string  `json:"night_img"`
	NightText string  `json:"night_text"`
}

type PassedChartEntry struct {
	Rain1h      float64 `json:"rain1h"`
	Rain24h     float64 `json:"rain24h"`
	Rain12h     float64 `json:"rain12h"`
	Rain6h      float64 `json:"rain6h"`
	Temperature float64 `json:"temperature"`
	TempDiff    string  `json:"tempDiff"`
	Humidity    float64 `json:"humidity"`
	Pressure    float64 `json:"pressure"`
	WindDir     float64 `json:"windDirection"`
	WindSpeed   float64 `json:"windSpeed"`
	Time        string  `json:"time"`
}
