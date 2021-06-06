package template

import (
	"embed"
	"html/template"
	"io"
)

//go:embed template.html
var templates embed.FS

func Render(wr io.Writer, data *Data) error {
	const templateFilename = "template.html"
	template, err := template.ParseFS(templates, templateFilename)
	if err != nil {
		return err
	}

	return template.Execute(wr, data)
}
