package components

import (
	"bytes"
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestWriteContainerExecution(t *testing.T) {
	o := addContainerDeployItemOptions{
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
	}

	err := o.parseImportDefinitions()
	assert.Nil(t, err, "failed to parse import definitions: %w", err)

	f := bytes.Buffer{}

	err = o.writeExecution(&f)
	assert.Nil(t, err, "failed to write execution template: %w", err)

	fmt.Println(f.String())
}
