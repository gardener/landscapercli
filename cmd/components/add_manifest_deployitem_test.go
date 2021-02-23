package components

import (
	"bytes"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestParseImportDefinition(t *testing.T) {
	tests := []struct {
		name           string
		paramDef       string
		expectedName   string
		expectedSchema string
		expectError    bool
	}{
		{
			name:           "parse-string-parameter-definition",
			paramDef:       "namespace:string",
			expectedName:   "namespace",
			expectedSchema: `{ "type": "string" }`,
		},
		{
			name:           "parse-integer-parameter-definition",
			paramDef:       "replicas:integer",
			expectedName:   "replicas",
			expectedSchema: `{ "type": "integer" }`,
		},
		{
			name:           "parse-boolean-parameter-definition",
			paramDef:       "ready:boolean",
			expectedName:   "ready",
			expectedSchema: `{ "type": "boolean" }`,
		},
		{
			name:        "parse-parameter-definition-with-unsupported-schema",
			paramDef:    "z:complex-number",
			expectError: true,
		},
		{
			name:        "parse-parameter-definition-without-schema",
			paramDef:    "replicas",
			expectError: true,
		},
		{
			name:        "parse-parameter-definition-with-empty-name",
			paramDef:    ":replicas",
			expectError: true,
		},
		{
			name:        "parse-parameter-definition-with-empty-schema",
			paramDef:    "replicas:",
			expectError: true,
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			o := addManifestDeployItemOptions{}
			importDefinition, err := o.parseImportDefinition(test.paramDef)
			if test.expectError {
				assert.NotNil(t, err, "expected an error when parsing parameter definition")
			} else {
				actualSchema := string(importDefinition.Schema.RawMessage)
				assert.Nil(t, err, "error parsing parameter definition")
				assert.Equal(t, test.expectedName, importDefinition.Name, "unexpected parameter name")
				assert.Equal(t, test.expectedSchema, actualSchema, "unexpected parameter schema")
			}
		})
	}
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
    name: {{ index .imports "testtarget" "metadata" "name" }}
    namespace: {{ index .imports "testtarget" "metadata" "namespace" }}
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
