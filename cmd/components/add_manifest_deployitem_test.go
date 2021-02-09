package components

import (
	"bytes"
	"fmt"
	"testing"
	"text/template"

	"github.com/stretchr/testify/assert"
)

func TestTemplate(t *testing.T) {
	templateString := `the cluster is {{ index .imports "target-cluster" }}`

	temp, err := template.New("").Parse(templateString)
	assert.Nil(t, err, "error when parsing template")

	data := map[string]interface{}{
		"imports": map[string]interface{}{
			"target-cluster": "{{ .rg }}",
		},
	}

	f := &bytes.Buffer{}
	err = temp.Execute(f, data)
	assert.Nil(t, err, "error during template execution")

	fmt.Println(f.String())
}

func TestWriteExecution(t *testing.T) {
	const deployItem1 = `deployItems:
- name: test-deployitem
  type: landscaper.gardener.cloud/kubernetes-manifest
  target:
    name: {{ index .imports "test-target" "metadata" "name" }}
    namespace: {{ index .imports "test-target" "metadata" "namespace" }}
  config:
    apiVersion: manifest.deployer.landscaper.gardener.cloud/v1alpha2
    kind: ProviderConfiguration
    updateStrategy: update
`
	const deployItem2 = `deployItems:
- name: testdeployitem
  type: landscaper.gardener.cloud/kubernetes-manifest
  target:
    name: {{ .imports.testtarget.metadata.name }}
    namespace: {{ .imports.testtarget.metadata.namespace }}
  config:
    apiVersion: manifest.deployer.landscaper.gardener.cloud/v1alpha2
    kind: ProviderConfiguration
    updateStrategy: update
`

	tests := []struct {
		name               string
		options            addManifestDeployItemOptions
		expectedDeployItem string
	}{
		{
			name: "param-name-with-minus",
			options: addManifestDeployItemOptions{
				deployItemName: "test-deployitem",
				updateStrategy: "update",
				clusterParam:   "test-target",
			},
			expectedDeployItem: deployItem1,
		},
		{
			name: "param-name-without-minus",
			options: addManifestDeployItemOptions{
				deployItemName: "testdeployitem",
				updateStrategy: "update",
				clusterParam:   "testtarget",
			},
			expectedDeployItem: deployItem2,
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			w := &bytes.Buffer{}
			err := test.options.writeExecution(w)
			assert.Nil(t, err, "error writing execution: %w", err)
			assert.Equal(t, test.expectedDeployItem, w.String(), "unexpected result for deployitem")
		})
	}
}
