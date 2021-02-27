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
	lat, long, err := coordinates()
	if err != nil {
		log.Fatal(err)
	}

	loc, err := location()
	if err != nil {
		log.Fatal(err)
	}

	data, err := templateData(lat, long, *loc)
	if err != nil {
		log.Fatal(err)
	}

	err = render(os.Stdout, data)
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

func location() (*time.Location, error) {
	return time.LoadLocation(os.Getenv("LOCATION"))
}

func templateData(lat float64, long float64, loc time.Location) (template.Data, error) {
	return openweather.TemplateData(lat, long, loc)
}

func render(wr io.Writer, data template.Data) error {
	return template.Render(wr, &data)
}
