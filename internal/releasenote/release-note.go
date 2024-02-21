package releaseNote

import (
	"fmt"
	"os"
	"text/template"

	"gopkg.in/yaml.v2"
)

func Generate(tplPath string, values []byte, output *os.File) error {
	if output == nil {
		return fmt.Errorf("please provide an output file")
	}
	var model map[string]any
	err := yaml.Unmarshal(values, &model)
	if err != nil {
		return err
	}

	tpl, err := template.ParseFiles(tplPath)
	if err != nil {
		return err
	}

	return tpl.Execute(output, model)
}
