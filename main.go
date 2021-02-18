package main

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"
	"time"

	"github.com/joho/godotenv"
)

func main() {
	godotenv.Load()

	f, err := weather(-37.840935, 144.946457)
	if err != nil {
		log.Fatal(err)
	}

	fmt.Printf("%+v\n", f)
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
	Time      EpochTime `json:"dt"`
	FeelsLike float64   `json:"feels_like"`
	Uvi       float64   `json:"uvi"`
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
