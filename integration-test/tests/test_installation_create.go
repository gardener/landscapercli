package tests

import (
	"bytes"
	"context"
	"fmt"

	"github.com/gardener/landscapercli/cmd/installations"
)

func RunInstallationCreateTest() error {
	// Check installation create with blueprint in component OCI artifact
	ctx := context.TODO()
	cmd := installations.NewCreateCommand(ctx)
	outBuf := bytes.NewBufferString("")
	cmd.SetOut(outBuf)
	args := []string{
		"oci-registry.landscaper.svc.cluster.local:5000",
		"github.com/gardener/echo-server-cd",
		"v0.1.0",
		"--allow-plain-http",
		"--render-schema-info",
	}
	cmd.SetArgs(args)

	err := cmd.Execute()
	if err != nil {
		return err
	}

	cmdOutput := outBuf.String()
	fmt.Println(cmdOutput)

	// check installation create with blueprint in separate OCI artifact

	return nil
}
