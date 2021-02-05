package tests

import (
	"bytes"
	"context"
	"fmt"
	"io/ioutil"
	"os"
	"path"

	"github.com/gardener/component-cli/pkg/commands/componentarchive/resources"
	lsv1alpha1 "github.com/gardener/landscaper/apis/core/v1alpha1"
	"github.com/gardener/landscapercli/cmd/installations"
	inttestutil "github.com/gardener/landscapercli/integration-test/util"
	"sigs.k8s.io/yaml"
)

func RunInstallationCreateTest() error {
	test := installationCreateTest{}
	err := test.run()
	if err != nil {
		// do not cleanup after erroneous test run to keep failed resources on the cluster
		return fmt.Errorf("test failed: %w", err)
	}

	return nil
}

type installationCreateTest struct{}

func (t *installationCreateTest) run() error {
	fmt.Println("========== Uploading echo-server component descriptor to OCI registry ==========")
	err := t.uploadEchoServerComponentDescriptor()
	if err != nil {
		return fmt.Errorf("upload of echo-server component descriptor failed: %w", err)
	}

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

	err = cmd.Execute()
	if err != nil {
		return err
	}

	cmdOutput := outBuf.String()
	fmt.Println(cmdOutput)

	return nil
}

func (t *installationCreateTest) createDummyBlueprint() *lsv1alpha1.Blueprint {
	bp := &lsv1alpha1.Blueprint{
		Imports: []lsv1alpha1.ImportDefinition{
			{
				FieldValueDefinition: lsv1alpha1.FieldValueDefinition{
					Name: "appname",
				},
			},
			{
				FieldValueDefinition: lsv1alpha1.FieldValueDefinition{
					Name: "appnamespace",
				},
			},
		},
		Exports:          []lsv1alpha1.ExportDefinition{},
		DeployExecutions: []lsv1alpha1.TemplateExecutor{},
	}
	return bp
}

func (t *installationCreateTest) uploadEchoServerComponentDescriptor() error {
	ctx := context.TODO()

	cdDir, err := ioutil.TempDir(".", "echo-server-cd-*")
	defer func() {
		err = os.RemoveAll(cdDir)
		if err != nil {
			fmt.Printf("cannot remove temporary directory %s: %s", cdDir, err.Error())
		}
	}()

	bpDir := path.Join(cdDir, "blueprint")
	err = os.Mkdir(bpDir, os.ModePerm)
	if err != nil {
		return err
	}

	bp := t.createDummyBlueprint()
	marshaledBp, err := yaml.Marshal(bp)
	if err != nil {
		return err
	}
	err = ioutil.WriteFile(path.Join(bpDir, "blueprint.yaml"), marshaledBp, os.ModePerm)
	if err != nil {
		return err
	}

	cd := inttestutil.CreateComponentDescriptor("github.com/gardener/echo-server-cd", "v0.1.0", "oci-registry.landscaper.svc.cluster.local:5000")
	marshaledCd, err := yaml.Marshal(cd)
	if err != nil {
		return err
	}
	err = ioutil.WriteFile(path.Join(cdDir, "component-descriptor.yaml"), marshaledCd, os.ModePerm)
	if err != nil {
		return err
	}

	resourcesYaml := `---
type: blueprint
name: ingress-nginx-blueprint
version: v0.1.0
relation: local
input:
  type: "dir"
  path: "./blueprint"
  compress: true
  mediaType: "application/vnd.gardener.landscaper.blueprint.v1+tar+gzip"
---
`

	resourceFile := path.Join(cdDir, "resources.yaml")
	err = ioutil.WriteFile(resourceFile, []byte(resourcesYaml), os.ModePerm)
	if err != nil {
		return nil
	}

	addResourcesCmd := resources.NewAddCommand(ctx)
	addResourcesArgs := []string{
		cdDir,
		"--resource",
		resourceFile,
	}
	addResourcesCmd.SetArgs(addResourcesArgs)

	err = addResourcesCmd.Execute()
	if err != nil {
		return fmt.Errorf("component-cli add resources failed: %w", err)
	}

	err = inttestutil.UploadComponentArchive(cdDir, "localhost:5000/component-descriptors/github.com/gardener/echo-server-cd:v0.1.0")
	if err != nil {
		return err
	}

	return nil
}
