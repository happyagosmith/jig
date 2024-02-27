package utils

import (
	"fmt"

	"github.com/vmware-labs/yaml-jsonpath/pkg/yamlpath"
	"gopkg.in/yaml.v3"
)

func GetYamlValue(yamlData []byte, path string) (string, error) {
	var node yaml.Node
	err := yaml.Unmarshal(yamlData, &node)
	if err != nil {
		return "", err
	}

	v, err := yamlpath.NewPath(path)
	if err != nil {
		return "", err
	}

	result, err := v.Find(&node)
	if err != nil {
		return "", err
	}

	if len(result) == 0 {
		return "", fmt.Errorf("path not found: %s", path)
	}

	return result[0].Value, nil
}
