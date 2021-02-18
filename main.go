package main

import (
	"encoding/json"
	"fmt"
	"html/template"
	"io"
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

func (t EpochTime) String() string {
	return time.Time(t).In(&loc).Format("Mon Jan 2 03:00PM")
}

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

func (f HourlyForecast) Date() string {
	return time.Time(f.Hour).In(&loc).Format("Mon Jan 2")
}

func (f HourlyForecast) Time() string {
	return time.Time(f.Hour).In(&loc).Format("03:00 PM")
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

func render(wr io.Writer, f Forecast) error {
	group := make(map[string][]HourlyForecast)
	groupOrder := make([]string, 0, 1)

	for _, h := range f.Hourly {
		date := h.Date()
		hours, exists := group[date]
		if !exists {
			group[date] = []HourlyForecast{h}
			groupOrder = append(groupOrder, date)
		} else {
			group[date] = append(hours, h)
		}
	}

	const templateFilename = "template.html"
	template, err := template.ParseFiles(templateFilename)
	if err != nil {
		return err
	}

	return template.Execute(wr, struct {
		Days           []string
		ForecastsByDay map[string][]HourlyForecast
	}{
		Days:           groupOrder,
		ForecastsByDay: group,
	})
}
