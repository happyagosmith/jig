package releaseNote

import (
	"fmt"
	"io"
	"path/filepath"
	"text/template"

	"github.com/Masterminds/sprig/v3"
	"gopkg.in/yaml.v2"
)

type ExtractedIssue = map[interface{}]interface{}

func Generate(tplPath string, values []byte, output io.Writer) error {
	if output == nil {
		return fmt.Errorf("please provide an output file")
	}
	var model map[interface{}]interface{}
	err := yaml.Unmarshal(values, &model)
	if err != nil {
		return err
	}

	jigFuncMap := template.FuncMap{
		"issuesFlatList": func(issuesMap ExtractedIssue) []ExtractedIssue {
			uniqueIssuesIdx := make(map[string]int)
			var uniqueIssuesSlice []ExtractedIssue
			for key, issues := range issuesMap {
				service := key.(string)
				for _, issue := range issues.([]interface{}) {
					i := issue.(map[interface{}]interface{})
					issueKey := i["issueKey"].(string)
					if _, ok := uniqueIssuesIdx[issueKey]; !ok {
						uIssue := issue.(map[interface{}]interface{})
						uIssue["impactedService"] = []string{service}

						uniqueIssuesIdx[issueKey] = len(uniqueIssuesSlice)
						uniqueIssuesSlice = append(uniqueIssuesSlice, uIssue)

						continue
					}

					idx := uniqueIssuesIdx[issueKey]
					existingIssue := uniqueIssuesSlice[idx]
					existingIssue["impactedService"] = append(existingIssue["impactedService"].([]string), service)
					uniqueIssuesSlice[idx] = existingIssue
				}
			}

			return uniqueIssuesSlice
		},
	}

	tpl, err := template.New(filepath.Base(tplPath)).Funcs(sprig.FuncMap()).Funcs(jigFuncMap).ParseFiles(tplPath)
	if err != nil {
		return err
	}

	return tpl.Execute(output, model)
}
