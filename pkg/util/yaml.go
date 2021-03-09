package util

import (
	"bytes"
	"strings"

	yamlv3 "gopkg.in/yaml.v3"
)

func MarshalYaml(node *yamlv3.Node) ([]byte, error) {
	buf := bytes.Buffer{}
	enc := yamlv3.NewEncoder(&buf)
	enc.SetIndent(2)
	err := enc.Encode(node)
	if err != nil {
		return nil, err
	}
	return buf.Bytes(), nil
}

func FindNodeByPath(node *yamlv3.Node, path string) (*yamlv3.Node, *yamlv3.Node) {
	if node == nil || path == "" {
		return nil, nil
	}

	var keyNode, valueNode *yamlv3.Node
	if node.Kind == yamlv3.DocumentNode {
		valueNode = node.Content[0]
	} else {
		valueNode = node
	}
	splittedPath := strings.Split(path, ".")

	for _, p := range splittedPath {
		keyNode, valueNode = findNode(valueNode.Content, p)
		if keyNode == nil && valueNode == nil {
			break
		}
	}

	return keyNode, valueNode
}

func findNode(nodes []*yamlv3.Node, name string) (*yamlv3.Node, *yamlv3.Node) {
	if nodes == nil {
		return nil, nil
	}

	var keyNode, valueNode *yamlv3.Node
	for i, node := range nodes {
		if node.Value == name {
			keyNode = node
			if i < len(nodes)-1 {
				valueNode = nodes[i+1]
			}
		} else if node.Kind == yamlv3.SequenceNode || node.Kind == yamlv3.MappingNode {
			keyNode, valueNode = findNode(node.Content, name)
		}

		if keyNode != nil && valueNode != nil {
			break
		}
	}

	return keyNode, valueNode
}
