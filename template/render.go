package template

import (
	"html/template"
	"io"
)

func Render(wr io.Writer, data *Data) error {
	const templateFilename = "template.html"
	template, err := template.ParseFiles(templateFilename)
	if err != nil {
		return err
	}

	return template.Execute(wr, data)
}
