package main

import (
	"io"
	"log"
	"os"
	"strconv"
	"time"

	"github.com/joho/godotenv"
	"github.com/ozbe/weather.ozbe.io/openweather"
	"github.com/ozbe/weather.ozbe.io/template"
)

func init() {
	godotenv.Load()
}

func main() {
	long, lat, err := coordinates()
	if err != nil {
		log.Fatal(err)
	}

	loc, err := time.LoadLocation(os.Getenv("LOCATION"))
	if err != nil {
		log.Fatal(err)
	}

	f, err := openweather.GetWeather(long, lat)
	if err != nil {
		log.Fatal(err)
	}

	err = render(os.Stdout, *f, *loc)
	if err != nil {
		log.Fatal(err)
	}
}

func coordinates() (lat float64, long float64, err error) {
	lat, err = strconv.ParseFloat(os.Getenv("LAT"), 64)
	if err != nil {
		return
	}

	long, err = strconv.ParseFloat(os.Getenv("LONG"), 64)
	if err != nil {
		return
	}

	return
}

func render(wr io.Writer, f openweather.Forecast, loc time.Location) error {
	data, err := openweather.TemplateData(f, loc)
	if err != nil {
		return err
	}

	return template.Render(wr, &data)
}
