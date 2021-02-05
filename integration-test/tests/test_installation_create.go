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
	"sigs.k8s.io/yaml"

	"github.com/gardener/landscapercli/cmd/installations"
	inttestutil "github.com/gardener/landscapercli/integration-test/util"
)

func RunInstallationCreateTest(config *inttestutil.Config) error {
	const (
		componentName         = "github.com/dummy-cd"
		componentVersion      = "v0.1.0"
		blueprintName         = "dummy-blueprint"
		dummyDataImportName   = "dummyDataImport"
		dummyTargetImportName = "dummyTargetImport"
		dummyDataExportName   = "dummyDataExport"
		dummyTargetExportName = "dummyTargetExport"
	)

	test := installationCreateTest{
		registryBaseURL:       config.RegistryBaseURL,
		componentName:         componentName,
		componentVersion:      componentVersion,
		blueprintName:         blueprintName,
		dummyDataImportName:   dummyDataImportName,
		dummyTargetImportName: dummyTargetImportName,
		dummyDataExportName:   dummyDataExportName,
		dummyTargetExportName: dummyTargetExportName,
	}
	err := test.run()
	if err != nil {
		// do not cleanup after erroneous test run to keep failed resources on the cluster
		return fmt.Errorf("test failed: %w", err)
	}

	return nil
}

type installationCreateTest struct {
	registryBaseURL       string
	componentName         string
	componentVersion      string
	blueprintName         string
	dummyDataImportName   string
	dummyTargetImportName string
	dummyDataExportName   string
	dummyTargetExportName string
}

func (t *installationCreateTest) run() error {
	ctx := context.TODO()

	fmt.Println("Creating and uploading dummy component to OCI registry")
	err := t.createAndUploadDummyComponent()
	if err != nil {
		return fmt.Errorf("creating/uploading dummy component failed: %w", err)
	}

	fmt.Println("Executing landscaper-cli installations create")
	cmd := installations.NewCreateCommand(ctx)
	outBuf := bytes.NewBufferString("")
	cmd.SetOut(outBuf)
	args := []string{
		"localhost:5000",
		t.componentName,
		t.componentVersion,
		"--allow-plain-http",
		"--render-schema-info",
	}
	cmd.SetArgs(args)

	err = cmd.Execute()
	if err != nil {
		return fmt.Errorf("landscaper-cli installations create failed: %w", err)
	}

	cmdOutput := outBuf.String()
	fmt.Println(cmdOutput)

	var actualInstallation lsv1alpha1.Installation
	err = yaml.Unmarshal(outBuf.Bytes(), &actualInstallation)
	if err != nil {
		return fmt.Errorf("cannot unmarshal output of landscaper-cli installations create: %w", err)
	}

	return nil
}

func (t *installationCreateTest) createDummyBlueprint() *lsv1alpha1.Blueprint {
	bp := &lsv1alpha1.Blueprint{
		Imports: []lsv1alpha1.ImportDefinition{
			{
				FieldValueDefinition: lsv1alpha1.FieldValueDefinition{
					Name: t.dummyDataImportName,
				},
			},
			{
				FieldValueDefinition: lsv1alpha1.FieldValueDefinition{
					Name:       t.dummyTargetImportName,
					TargetType: string(lsv1alpha1.KubernetesClusterTargetType),
				},
			},
		},
		Exports: []lsv1alpha1.ExportDefinition{
			{
				FieldValueDefinition: lsv1alpha1.FieldValueDefinition{
					Name: t.dummyDataExportName,
				},
			},
			{
				FieldValueDefinition: lsv1alpha1.FieldValueDefinition{
					Name:       t.dummyTargetExportName,
					TargetType: string(lsv1alpha1.KubernetesClusterTargetType),
				},
			},
		},
		DeployExecutions: []lsv1alpha1.TemplateExecutor{},
	}
	return bp
}

func (t *installationCreateTest) createAndUploadDummyComponent() error {
	ctx := context.TODO()

	cdDir, err := ioutil.TempDir(".", "dummy-cd-*")
	if err != nil {
		return fmt.Errorf("cannot create component descriptor directory: %w", err)
	}
	defer func() {
		removeErr := os.RemoveAll(cdDir)
		if removeErr != nil {
			fmt.Printf("cannot remove temporary directory %s: %s", cdDir, removeErr.Error())
		}
	}()

	cd := inttestutil.CreateComponentDescriptor(t.componentName, t.componentVersion, t.registryBaseURL)
	marshaledCd, err := yaml.Marshal(cd)
	if err != nil {
		return fmt.Errorf("cannot marshal component descriptor: %w", err)
	}
	err = ioutil.WriteFile(path.Join(cdDir, "component-descriptor.yaml"), marshaledCd, os.ModePerm)
	if err != nil {
		return fmt.Errorf("cannot write component descriptor file: %w", err)
	}

	bpDir := path.Join(cdDir, "blueprint")
	err = os.Mkdir(bpDir, os.ModePerm)
	if err != nil {
		return fmt.Errorf("cannot create blueprint directory: %w", err)
	}

	bp := t.createDummyBlueprint()
	marshaledBp, err := yaml.Marshal(bp)
	if err != nil {
		return fmt.Errorf("cannot marshal blueprint: %w", err)
	}
	err = ioutil.WriteFile(path.Join(bpDir, "blueprint.yaml"), marshaledBp, os.ModePerm)
	if err != nil {
		return fmt.Errorf("cannot write blueprint.yaml: %w", err)
	}

	resourcesYaml := fmt.Sprintf(`---
type: blueprint
name: %s
version: %s
relation: local
input:
  type: "dir"
  path: "./blueprint"
  compress: true
  mediaType: "application/vnd.gardener.landscaper.blueprint.v1+tar+gzip"
---
`, t.blueprintName, t.componentVersion)

	resourceFile := path.Join(cdDir, "resources.yaml")
	err = ioutil.WriteFile(resourceFile, []byte(resourcesYaml), os.ModePerm)
	if err != nil {
		return fmt.Errorf("cannot write component descriptor resources file: %w", err)
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

	uploadRef := fmt.Sprintf("localhost:5000/component-descriptors/%s:%s", t.componentName, t.componentVersion)
	err = inttestutil.UploadComponentArchive(cdDir, uploadRef)
	if err != nil {
		return fmt.Errorf("cannot upload component archive: %w", err)
	}

	return nil
}
