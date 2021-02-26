package template

import "html/template"

type Data interface {
	Days() []Day
}

type Day interface {
	Date() string
	Hours() []Hour
}

type Hour interface {
	Time() string
	Temp() string
	UV() UV
	Condition() Condition
}

type UV interface {
	Index() string
	Classification() string
}

type Condition interface {
	Icon() template.URL
	Description() string
}
