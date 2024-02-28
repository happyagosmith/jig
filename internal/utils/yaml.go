package utils

import (
	"bytes"
	"fmt"
	"strings"

	"gopkg.in/yaml.v3"
)

type Yaml struct {
	node *yaml.Node
}

func NewYaml(b []byte) (*Yaml, error) {
	var yn yaml.Node
	if err := yaml.Unmarshal(b, &yn); err != nil {
		return nil, err
	}

	var v map[string]interface{}
	if err := yaml.Unmarshal(b, &v); err != nil {
		return nil, err
	}

	return &Yaml{node: &yn}, nil
}

func (y *Yaml) Delete(key string) error {
	if len := len(y.node.Content); len == 0 {
		return nil
	}

	if len := len(y.node.Content[0].Content); len == 0 {
		return nil
	}

	node := y.node.Content[0]
	nElements := len(node.Content) / 2
	for i := 0; i < nElements; i++ {
		nodeKey := node.Content[i*2].Value
		if nodeKey != key {
			continue
		}

		node.Content = append(node.Content[:i*2], node.Content[i*2+2:]...)
		return nil
	}

	return nil
}

func (y *Yaml) Merge(oy *Yaml) error {
	err := mergeNodes(y.node, oy.node)
	return err
}

func (y Yaml) String() (string, error) {
	out, err := y.Bytes()
	if err != nil {
		return "", err
	}

	return strings.TrimSuffix(string(out), "\n"), nil
}

func (y Yaml) Bytes() ([]byte, error) {
	if len := len(y.node.Content); len == 0 {
		return nil, nil
	}
	if len := len(y.node.Content[0].Content); len == 0 {
		return nil, nil
	}
	var out bytes.Buffer
	encoder := yaml.NewEncoder(&out)
	encoder.SetIndent(1)
	err := encoder.Encode(y.node)
	if err != nil {
		return nil, err
	}
	defer encoder.Close()

	return out.Bytes(), nil
}

func mergeNodes(a, b *yaml.Node) error {
	if a.Kind != b.Kind {
		return fmt.Errorf("it is not possible to merge different types")
	}

	if a.Kind == yaml.DocumentNode {
		err := mergeNodes(a.Content[0], b.Content[0])
		if err != nil {
			return err
		}
	}

	if a.Kind == yaml.MappingNode {
		lmb := lookUpMap(b.Content)

		for i := 0; i < len(a.Content)/2; i++ {
			key := a.Content[i*2].Value
			if n, ok := lmb[key]; ok {
				if n.nodeValue.Kind == yaml.MappingNode {
					_ = mergeNodes(a.Content[i*2+1], n.nodeValue)
					a.Column = a.Column - 2
				} else {
					a.Content[i*2+1] = n.nodeValue
				}
				n.found = true
				lmb[key] = n
			}
		}

		appendContent(a, b, lmb)
	}

	return nil
}

func ToYaml(v any, indent int) string {
	data, err := yaml.Marshal(v)
	if err != nil {
		return ""
	}
	is := "\n" + strings.Repeat(" ", indent)
	ys := string(data)
	ys = is + strings.ReplaceAll(ys, "\n", is)
	ys = strings.TrimSuffix(ys, is)
	return ys
}

type node struct {
	nodeKey   *yaml.Node
	nodeValue *yaml.Node
	found     bool
}

func lookUpMap(nodes []*yaml.Node) map[string]node {
	nb := map[string]node{}
	for i := 0; i < len(nodes)/2; i++ {
		key := nodes[i*2]
		value := nodes[i*2+1]
		nb[key.Value] = node{
			nodeKey:   key,
			nodeValue: value,
			found:     false,
		}
	}

	return nb
}

func appendContent(a *yaml.Node, b *yaml.Node, lmb map[string]node) {
	for i := 0; i < len(b.Content)/2; i++ {
		key := b.Content[i*2].Value
		if n := lmb[key]; !n.found {
			a.Content = append(a.Content, lmb[key].nodeKey)
			a.Content = append(a.Content, lmb[key].nodeValue)
		}
	}
}
