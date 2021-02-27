package openweather

import (
	"encoding/base64"
	"fmt"
	"html/template"
	"time"

	t "github.com/ozbe/weather.ozbe.io/template"
)

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

func (d templateDay) Times() []string {
	times := make([]string, len(d.hours))

	for i, hour := range d.hours {
		times[i] = hour.Time()
	}

	return times
}

func (d templateDay) Temps() []string {
	temps := make([]string, len(d.hours))

	for i, hour := range d.hours {
		temps[i] = hour.Temp()
	}

	return temps
}

func (d templateDay) UVs() []string {
	uvs := make([]string, len(d.hours))

	for i, hour := range d.hours {
		uvs[i] = hour.UV().Index()
	}

	return uvs
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

func TemplateData(lat float64, long float64, loc time.Location) (t.Data, error) {
	f, err := getWeather(lat, long)
	if err != nil {
		return nil, err
	}

	dates := make([]string, 0, 2)
	forecastsByDate := make(map[string][]hourlyForecast)

	for _, h := range f.Hourly {
		localTime := time.Time(h.Hour).In(&loc)
		date := localTime.Format("Mon Jan 2")
		hours, exists := forecastsByDate[date]
		if !exists {
			forecastsByDate[date] = []hourlyForecast{h}
			dates = append(dates, date)
		} else {
			forecastsByDate[date] = append(hours, h)
		}
	}

	icons, err := mapIconImgSrc(*f)
	if err != nil {
		return nil, err
	}

	tds := make([]templateDay, len(dates))

	for di, date := range dates {
		hours := forecastsByDate[date]
		ths := make([]templateHour, len(hours))

		for hi, hour := range hours {
			ths[hi] = templateHour{
				time: localTime(loc, hour),
				temp: temp(hour),
				uv: templateUV{
					index:          uvIndex(hour),
					classification: uvLevel(hour),
				},
				condition: templateCondition{
					icon:        icon(icons, hour),
					description: conditionDescription(hour),
				},
			}
		}

		tds[di] = templateDay{date, ths}
	}

	return &templateData{
		days: tds,
	}, nil
}

func localTime(loc time.Location, h hourlyForecast) string {
	localTime := time.Time(h.Hour).In(&loc)
	return localTime.Format("3 PM")
}

func temp(h hourlyForecast) string {
	return fmt.Sprintf("%.f", h.FeelsLike)
}

func uvIndex(h hourlyForecast) string {
	return fmt.Sprintf("%.f", h.UVI)
}

func uvLevel(h hourlyForecast) string {
	if h.UVI < 5 {
		return "Low"
	} else if h.UVI < 8 {
		return "Moderate"
	} else {
		return "Extreme"
	}
}

func icon(icons *map[string]string, h hourlyForecast) template.URL {
	return template.URL((*icons)[h.Weather[0].Icon])
}

func conditionDescription(h hourlyForecast) string {
	return h.Weather[0].Description
}

type iconSrc struct {
	icon string
	src  string
	err  error
}

func mapIconImgSrc(f forecast) (*map[string]string, error) {
	icons := make(map[string]bool)
	result := make(map[string]string)

	for _, h := range f.Hourly {
		icons[h.Weather[0].Icon] = true
	}

	numIcons := len(icons)
	ch := make(chan iconSrc, numIcons)
	for i := range icons {
		go func(icon string) {
			img, err := iconImage(icon)
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
