package openweather

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"os"
	"time"
)

type forecast struct {
	Hourly []hourlyForecast `json:"hourly"`
}

type epochTime time.Time

func (t *epochTime) UnmarshalJSON(b []byte) error {
	var i int64
	err := json.Unmarshal(b, &i)
	if err != nil {
		return err
	}

	*t = epochTime(time.Unix(i, 0))
	return nil
}

type hourlyForecast struct {
	Hour      epochTime `json:"dt"`
	FeelsLike float64   `json:"feels_like"`
	UVI       float64   `json:"uvi"`
	Weather   []weather `json:"weather"`
}

type weather struct {
	Description string `json:"description"`
	Icon        string `json:"icon"`
}

func getWeather(lat float64, long float64) (*forecast, error) {
	apiKey := os.Getenv("OPENWEATHER_API_KEY")
	url := fmt.Sprintf("https://api.openweathermap.org/data/2.5/onecall?lat=%f&lon=%f&exclude=current,minutely,daily,alerts&units=metric&appid=%s", lat, long, apiKey)

	res, err := http.Get(url)
	if err != nil {
		return nil, err
	}

	var result forecast
	err = json.NewDecoder(res.Body).Decode(&result)
	if err != nil {
		return nil, err
	}

	return &result, nil
}

func iconImage(icon string) ([]byte, error) {
	res, err := http.Get(fmt.Sprintf("https://openweathermap.org/img/w/%s.png", icon))
	if err != nil {
		return nil, err
	}

	body, err := ioutil.ReadAll(res.Body)
	if err != nil {
		return nil, err
	}

	return body, nil
}
