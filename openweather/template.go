package openweather

import (
	"encoding/base64"
	"fmt"
	"html/template"
	"log"
	"os"
	"time"

	t "github.com/ozbe/weather.ozbe.io/template"
)

var loc time.Location

func init() {
	// TODO - consider moving this to an explicit argument
	l, err := time.LoadLocation(os.Getenv("LOCATION"))
	if err != nil {
		log.Fatal(err)
	}
	loc = *l
}

type templateData struct {
	days []templateDay
}

type templateDay struct {
	date  string
	hours []templateHour
}

func (d templateDay) Date() string {
	return d.date
}

func (d templateDay) Hours() []t.Hour {
	hours := make([]t.Hour, len(d.hours))

	for i, hour := range d.hours {
		hours[i] = hour
	}

	return hours
}

type templateHour struct {
	time      string
	temp      string
	uv        templateUV
	condition templateCondition
}

func (h templateHour) Time() string {
	return h.time
}

func (h templateHour) Temp() string {
	return h.temp
}

func (h templateHour) UV() t.UV {
	return h.uv
}

func (h templateHour) Condition() t.Condition {
	return h.condition
}

type templateUV struct {
	index          string
	classification string
}

func (uv templateUV) Index() string {
	return uv.index
}

func (uv templateUV) Classification() string {
	return uv.classification
}

type templateCondition struct {
	icon        template.URL
	description string
}

func (c templateCondition) Icon() template.URL {
	return c.icon
}

func (c templateCondition) Description() string {
	return c.description
}

func (d templateData) Days() []t.Day {
	days := make([]t.Day, len(d.days))

	for i, day := range d.days {
		days[i] = day
	}

	return days
}

func (f HourlyForecast) localTime() time.Time {
	return time.Time(f.Hour).In(&loc)
}

func (f HourlyForecast) date() string {
	return f.localTime().Format("Mon Jan 2")
}

func (f HourlyForecast) time() string {
	return f.localTime().Format("03:00 PM")
}

func (f HourlyForecast) temp() string {
	return fmt.Sprintf("%.1f°", f.FeelsLike)
}

func (f HourlyForecast) uv() string {
	return fmt.Sprintf("%.f", f.UVI)
}

func (f HourlyForecast) uvLevel() string {
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

func TemplateData(f Forecast) (t.Data, error) {
	dates := make([]string, 0, 2)
	forecastsByDate := make(map[string][]HourlyForecast)

	for _, h := range f.Hourly {
		date := h.date()
		hours, exists := forecastsByDate[date]
		if !exists {
			forecastsByDate[date] = []HourlyForecast{h}
			dates = append(dates, date)
		} else {
			forecastsByDate[date] = append(hours, h)
		}
	}

	icons, err := mapIconImgSrc(f)
	if err != nil {
		return nil, err
	}

	tds := make([]templateDay, len(dates))

	fmt.Printf("%+v\n", dates)

	for di, date := range dates {
		hours := forecastsByDate[date]
		ths := make([]templateHour, len(hours))

		for hi, hour := range hours {
			ths[hi] = templateHour{
				time: hour.time(),
				temp: hour.temp(),
				uv: templateUV{
					index:          hour.uv(),
					classification: hour.uvLevel(),
				},
				condition: templateCondition{
					icon:        template.URL((*icons)[hour.Icon()]),
					description: hour.Condition(),
				},
			}
		}

		tds[di] = templateDay{date, ths}
	}

	fmt.Printf("%+v\n", tds)

	return &templateData{
		days: tds,
	}, nil
}

type iconSrc struct {
	icon string
	src  string
	err  error
}

func mapIconImgSrc(f Forecast) (*map[string]string, error) {
	icons := make(map[string]bool)
	result := make(map[string]string)

	for _, h := range f.Hourly {
		icons[h.Icon()] = true
	}

	numIcons := len(icons)
	ch := make(chan iconSrc, numIcons)
	for i := range icons {
		go func(icon string) {
			img, err := iconImgSrc(icon)
			encodedImg := base64.StdEncoding.EncodeToString(img)
			src := fmt.Sprintf("data:image/png;base64,%s", encodedImg)
			ch <- iconSrc{icon, src, err}
		}(i)
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
