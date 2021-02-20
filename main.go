package main

import (
	"encoding/base64"
	"encoding/json"
	"fmt"
	"html/template"
	"io"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"time"

	"github.com/joho/godotenv"
)

var loc time.Location

func init() {
	l, err := time.LoadLocation("Australia/Melbourne")
	if err != nil {
		log.Fatal(err)
	}
	loc = *l
}

func main() {
	godotenv.Load()

	f, err := weather(-37.840935, 144.946457)
	if err != nil {
		log.Fatal(err)
	}

	err = render(os.Stdout, *f)
	if err != nil {
		log.Fatal(err)
	}
}

type Forecast struct {
	Hourly []HourlyForecast `json:"hourly"`
}

type EpochTime time.Time

func (t *EpochTime) UnmarshalJSON(b []byte) error {
	var i int64
	err := json.Unmarshal(b, &i)
	if err != nil {
		return err
	}

	*t = EpochTime(time.Unix(i, 0))
	return nil
}

type HourlyForecast struct {
	Hour      EpochTime `json:"dt"`
	FeelsLike float64   `json:"feels_like"`
	UVI       float64   `json:"uvi"`
	Weather   []Weather `json:"weather"`
}

func (f HourlyForecast) LocalTime() time.Time {
	return time.Time(f.Hour).In(&loc)
}

func (f HourlyForecast) Date() string {
	return f.LocalTime().Format("Mon Jan 2")
}

func (f HourlyForecast) Time() string {
	return f.LocalTime().Format("03:00 PM")
}

func (f HourlyForecast) Temp() string {
	return fmt.Sprintf("%.1f°", f.FeelsLike)
}

func (f HourlyForecast) UV() string {
	return fmt.Sprintf("%.f", f.UVI)
}

func (f HourlyForecast) UVLevel() string {
	if f.UVI < 5 {
		return "Low"
	} else if f.UVI < 8 {
		return "Moderate"
	} else {
		return "Extreme"
	}
}

func (f HourlyForecast) Condition() string {
	return f.Weather[0].Description
}

func (f HourlyForecast) Icon() string {
	return f.Weather[0].Icon
}

type Weather struct {
	Description string `json:"description"`
	Icon        string `json:"icon"`
}

func weather(lat float64, long float64) (*Forecast, error) {
	apiKey := os.Getenv("OPENWEATHER_API_KEY")
	url := fmt.Sprintf("https://api.openweathermap.org/data/2.5/onecall?lat=%f&lon=%f&exclude=current,minutely,daily,alerts&units=metric&appid=%s", lat, long, apiKey)

	res, err := http.Get(url)
	if err != nil {
		return nil, err
	}

	var result Forecast
	err = json.NewDecoder(res.Body).Decode(&result)
	if err != nil {
		return nil, err
	}

	return &result, nil
}

type TemplateData struct {
	Days           []string
	ForecastsByDay map[string][]HourlyForecast
	IconSrc        map[string]template.URL
}

func render(wr io.Writer, f Forecast) error {
	data, err := templateData(f)
	if err != nil {
		return err
	}

	return renderTemplate(wr, data)
}

func templateData(f Forecast) (*TemplateData, error) {
	days := make([]string, 0, 1)
	forecastsByDay := make(map[string][]HourlyForecast)

	for _, h := range f.Hourly {
		date := h.Date()
		hours, exists := forecastsByDay[date]
		if !exists {
			forecastsByDay[date] = []HourlyForecast{h}
			days = append(days, date)
		} else {
			forecastsByDay[date] = append(hours, h)
		}
	}

	icons, err := mapIconImgSrc(f)
	if err != nil {
		return nil, err
	}

	return &TemplateData{
		Days:           days,
		ForecastsByDay: forecastsByDay,
		IconSrc:        *icons,
	}, nil
}

type iconSrc struct {
	icon string
	src  template.URL
	err  error
}

func mapIconImgSrc(f Forecast) (*map[string]template.URL, error) {
	icons := make(map[string]bool)
	result := make(map[string]template.URL)

	for _, h := range f.Hourly {
		icons[h.Icon()] = true
	}

	numIcons := len(icons)
	ch := make(chan iconSrc, numIcons)
	for k := range icons {
		go func(icon string) {
			src, err := iconImgSrc(icon)
			ch <- iconSrc{icon, src, err}
		}(k)
	}

	for i := numIcons; i > 0; i-- {
		i := <-ch
		if i.err != nil {
			return nil, i.err
		}
		result[i.icon] = i.src
	}

	return &result, nil
}

func iconImgSrc(icon string) (template.URL, error) {
	res, err := http.Get(fmt.Sprintf("https://openweathermap.org/img/w/%s.png", icon))
	if err != nil {
		return "", err
	}

	body, err := ioutil.ReadAll(res.Body)
	if err != nil {
		return "", err
	}

	encodedImg := base64.StdEncoding.EncodeToString(body)
	src := fmt.Sprintf("data:image/png;base64,%s", encodedImg)
	return template.URL(src), nil
}

func renderTemplate(wr io.Writer, data *TemplateData) error {
	const templateFilename = "template.html"
	template, err := template.ParseFiles(templateFilename)
	if err != nil {
		return err
	}

	return template.Execute(wr, data)
}
