package releaseNote

import (
	"os"
	"text/template"

	"gopkg.in/yaml.v2"
)

func Generate(tplPath string, values []byte) error {
	var model map[string]any
	err := yaml.Unmarshal(values, &model)
	if err != nil {
		return err
	}

	tpl, err := template.ParseFiles(tplPath)
	if err != nil {
		return err
	}

	return tpl.Execute(os.Stdout, model)
}
