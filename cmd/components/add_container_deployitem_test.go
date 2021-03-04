package components

import (
	"bytes"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestWriteContainerExecution(t *testing.T) {
	tests := []struct {
		name           string
		options        addContainerDeployItemOptions
		expectedResult string
	}{
		{
			name: "container execution without component data",
			options: addContainerDeployItemOptions{
				componentPath:     "",
				deployItemName:    "test",
				image:             "alpine",
				command:           &[]string{"sh", "-c"},
				args:              &[]string{"env", "ls"},
				importParams:      &[]string{"replicas:integer", "enabled:boolean"},
				importDefinitions: nil,
				exportParams:      nil,
				exportDefinitions: nil,
				clusterParam:      "target-cluster",
			},
			expectedResult: `deployItems:
- name: test
  type: landscaper.gardener.cloud/container
  config:
    apiVersion: container.deployer.landscaper.gardener.cloud/v1alpha1
    kind: ProviderConfiguration
    image: alpine
    args:
    - env
    - ls
    command:
    - sh
    - -c
    importValues: 
      {{ toJson .imports | indent 2 }}
    `,
		},
		{
			name: "container execution with component data",
			options: addContainerDeployItemOptions{
				componentPath:     "",
				deployItemName:    "test",
				image:             "alpine",
				command:           &[]string{"sh", "-c"},
				args:              &[]string{},
				importParams:      &[]string{"replicas:integer", "enabled:boolean"},
				importDefinitions: nil,
				exportParams:      nil,
				exportDefinitions: nil,
				clusterParam:      "target-cluster",
				addComponentData:  true,
			},
			expectedResult: `deployItems:
- name: test
  type: landscaper.gardener.cloud/container
  config:
    apiVersion: container.deployer.landscaper.gardener.cloud/v1alpha1
    kind: ProviderConfiguration
    image: alpine
    args: []
    command:
    - sh
    - -c
    importValues: 
      {{ toJson .imports | indent 2 }}
    componentDescriptor: 
      {{ toJson .componentDescriptorDef | indent 2 }}
    blueprint: 
      {{ toJson .blueprint | indent 2 }}
    `,
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			err := test.options.parseImportDefinitions()
			assert.Nil(t, err, "failed to parse import definitions: %w", err)
			f := bytes.Buffer{}
			err = test.options.writeExecution(&f)
			assert.Nil(t, err, "failed to write execution template: %w", err)
			actualResult := f.String()
			assert.Equal(t, test.expectedResult, actualResult, "unexpected result")
		})
	}
}
