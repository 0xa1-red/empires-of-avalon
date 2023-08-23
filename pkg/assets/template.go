package assets

import (
	"html/template"
	"io"
)

func LoadTemplate(path string) (*template.Template, error) {
	t := template.New("")

	f, err := Assets.Open(path)
	if err != nil {
		return nil, err
	}

	d, err := io.ReadAll(f)
	if err != nil {
		return nil, err
	}

	t, err = t.Parse(string(d))
	if err != nil {
		return nil, err
	}

	return t, err
}
