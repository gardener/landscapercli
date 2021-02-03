package util

import (
	"testing"

	"github.com/stretchr/testify/assert"
	yamlv3 "gopkg.in/yaml.v3"
)

func TestFindNodeByPath(t *testing.T) {
	tests := []struct {
		name              string
		yaml              string
		keyPath           string
		expectedKeyNode   *yamlv3.Node
		expectedValueNode *yamlv3.Node
	}{
		{
			name: "simple positive test",
			yaml: `
            key1: val1
            key2:
              key3: val3
            `,
			keyPath: "key2.key3",
			expectedKeyNode: &yamlv3.Node{
				Kind:  yamlv3.ScalarNode,
				Tag:   "!!str",
				Value: "key3",
			},
			expectedValueNode: &yamlv3.Node{
				Kind:  yamlv3.ScalarNode,
				Tag:   "!!str",
				Value: "val3",
			},
		},
		{
			name: "invalid path 1",
			yaml: `
            key1: val1
            key2:
              key3: val3
            `,
			keyPath: "key2.key4",
		},
		{
			name: "invalid path 2",
			yaml: `
            key1: val1
            key2:
              key3: val3
            `,
			keyPath: "key4",
		},
		{
			name: "empty path",
			yaml: `
            key1: val1
            key2:
              key3: val3
            `,
		},
		{
			name:    "empty node",
			keyPath: "key1",
		},
		{
			name: "key defined, but value not",
			yaml: `
            key1: val1
            key2:
              key3:
            `,
			keyPath: "key2.key3",
			expectedKeyNode: &yamlv3.Node{
				Kind:  yamlv3.ScalarNode,
				Tag:   "!!str",
				Value: "key3",
			},
			expectedValueNode: &yamlv3.Node{
				Kind: yamlv3.ScalarNode,
				Tag:  "!!null",
			},
		},
	}

	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			rootNode := &yamlv3.Node{}
			err := yamlv3.Unmarshal([]byte(tt.yaml), rootNode)
			assert.NoError(t, err)

			keyNode, valueNode := FindNodeByPath(rootNode, tt.keyPath)

			assertNodeEquality(t, tt.expectedKeyNode, keyNode)
			assertNodeEquality(t, tt.expectedValueNode, valueNode)
		})
	}
}

func assertNodeEquality(t *testing.T, expectedNode, actualNode *yamlv3.Node) {
	if expectedNode != nil {
		assert.NotNil(t, actualNode)
		assert.Equal(t, expectedNode.Kind, actualNode.Kind)
		assert.Equal(t, expectedNode.Tag, actualNode.Tag)
		assert.Equal(t, expectedNode.Value, actualNode.Value)
		assert.Equal(t, expectedNode.Content, actualNode.Content)
	} else {
		assert.Nil(t, actualNode)
	}
}
