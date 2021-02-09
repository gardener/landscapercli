package tests

import (
	"bytes"
	"context"
	"fmt"
	"io/ioutil"
	"os"
	"path"

	"github.com/gardener/component-cli/pkg/commands/componentarchive/resources"
	cdv2 "github.com/gardener/component-spec/bindings-go/apis/v2"
	lsv1alpha1 "github.com/gardener/landscaper/apis/core/v1alpha1"
	"github.com/stretchr/testify/assert"
	yamlv3 "gopkg.in/yaml.v3"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"sigs.k8s.io/yaml"

	"github.com/gardener/landscapercli/cmd/installations"
	inttestutil "github.com/gardener/landscapercli/integration-test/util"
	"github.com/gardener/landscapercli/pkg/util"
)

func RunInstallationCreateTest(config *inttestutil.Config) error {
	const (
		installationName = "test-installation"
		componentName    = "github.com/dummy-cd"
		componentVersion = "v0.1.0"
		blueprintName    = "dummy-blueprint"
	)

	test := installationCreateTest{
		registryBaseURL:  config.RegistryBaseURL,
		installationName: installationName,
		componentName:    componentName,
		componentVersion: componentVersion,
		blueprintName:    blueprintName,
	}

	err := test.run()
	if err != nil {
		// do not cleanup after erroneous test run to keep failed resources on the cluster
		return fmt.Errorf("test failed: %w", err)
	}

	return nil
}

type installationCreateTest struct {
	registryBaseURL  string
	installationName string
	componentName    string
	componentVersion string
	blueprintName    string
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
	outBuf := &bytes.Buffer{}
	cmd.SetOut(outBuf)
	args := []string{
		"localhost:5000",
		t.componentName,
		t.componentVersion,
		"--name",
		t.installationName,
		"--allow-plain-http",
		"--render-schema-info",
	}
	cmd.SetArgs(args)

	err = cmd.Execute()
	if err != nil {
		return fmt.Errorf("landscaper-cli installations create failed: %w", err)
	}

	actualInstallation := lsv1alpha1.Installation{}
	err = yaml.Unmarshal(outBuf.Bytes(), &actualInstallation)
	if err != nil {
		return fmt.Errorf("cannot unmarshal output of landscaper-cli installations create: %w", err)
	}

	expectedInstallation := lsv1alpha1.Installation{
		TypeMeta: metav1.TypeMeta{
			Kind:       "Installation",
			APIVersion: lsv1alpha1.SchemeGroupVersion.String(),
		},
		ObjectMeta: metav1.ObjectMeta{
			Name: t.installationName,
		},
		Spec: lsv1alpha1.InstallationSpec{
			ComponentDescriptor: &lsv1alpha1.ComponentDescriptorDefinition{
				Reference: &lsv1alpha1.ComponentDescriptorReference{
					Version:       t.componentVersion,
					ComponentName: t.componentName,
					RepositoryContext: &cdv2.RepositoryContext{
						Type:    cdv2.OCIRegistryType,
						BaseURL: t.registryBaseURL,
					},
				},
			},
			Blueprint: lsv1alpha1.BlueprintDefinition{
				Reference: &lsv1alpha1.RemoteBlueprintReference{
					ResourceName: t.blueprintName,
				},
			},
			Imports: lsv1alpha1.InstallationImports{
				Data: []lsv1alpha1.DataImport{
					{
						Name: "dummyDataImport",
					},
				},
				Targets: []lsv1alpha1.TargetImportExport{
					{
						Name: "dummyTargetImport",
					},
				},
			},
			Exports: lsv1alpha1.InstallationExports{
				Data: []lsv1alpha1.DataExport{
					{
						Name: "dummyDataExport",
					},
				},
				Targets: []lsv1alpha1.TargetImportExport{
					{
						Name: "dummyTargetExport",
					},
				},
			},
		},
	}

	fmt.Println("Checking generated installation")

	ok := assert.Equal(inttestutil.DummyTestingT{}, expectedInstallation, actualInstallation)
	if !ok {
		return fmt.Errorf("")
	}

	rootNode := &yamlv3.Node{}
	err = yamlv3.Unmarshal(outBuf.Bytes(), rootNode)

	_, dataImportsNode := util.FindNodeByPath(rootNode, "spec.imports.data")
	expectedSchema := `# JSON schema
# {
#   "type": "string"
# }`
	ok = assert.Equal(inttestutil.DummyTestingT{}, expectedSchema, dataImportsNode.Content[0].Content[0].HeadComment)
	if !ok {
		return fmt.Errorf("schema comments for spec.imports.data are invalid")
	}

	_, targetImportsNode := util.FindNodeByPath(rootNode, "spec.imports.targets")
	expectedSchema = "# Target type: landscaper.gardener.cloud/kubernetes-cluster"
	ok = assert.Equal(inttestutil.DummyTestingT{}, expectedSchema, targetImportsNode.Content[0].Content[0].HeadComment)
	if !ok {
		return fmt.Errorf("schema comments for spec.imports.targets are invalid")
	}

	_, dataExportsNode := util.FindNodeByPath(rootNode, "spec.exports.data")
	dataExportsNode, _ = util.FindNodeByPath(dataExportsNode.Content[0], "name")
	expectedSchema = `# JSON schema
# {
#   "type": "string"
# }`
	ok = assert.Equal(inttestutil.DummyTestingT{}, expectedSchema, dataExportsNode.HeadComment)
	if !ok {
		return fmt.Errorf("schema comments for spec.exports.data are invalid")
	}

	_, targetExportsNode := util.FindNodeByPath(rootNode, "spec.exports.targets")
	expectedSchema = "# Target type: landscaper.gardener.cloud/kubernetes-cluster"
	ok = assert.Equal(inttestutil.DummyTestingT{}, expectedSchema, targetExportsNode.Content[0].Content[0].HeadComment)
	if !ok {
		return fmt.Errorf("schema comments for spec.exports.targets are invalid")
	}

	return nil
}

func (t *installationCreateTest) createDummyBlueprint() *lsv1alpha1.Blueprint {
	bp := &lsv1alpha1.Blueprint{
		Imports: []lsv1alpha1.ImportDefinition{
			{
				FieldValueDefinition: lsv1alpha1.FieldValueDefinition{
					Name:   "dummyDataImport",
					Schema: lsv1alpha1.JSONSchemaDefinition("{ \"type\": \"string\" }"),
				},
			},
			{
				FieldValueDefinition: lsv1alpha1.FieldValueDefinition{
					Name:       "dummyTargetImport",
					TargetType: string(lsv1alpha1.KubernetesClusterTargetType),
				},
			},
		},
		Exports: []lsv1alpha1.ExportDefinition{
			{
				FieldValueDefinition: lsv1alpha1.FieldValueDefinition{
					Name:   "dummyDataExport",
					Schema: lsv1alpha1.JSONSchemaDefinition("{ \"type\": \"string\" }"),
				},
			},
			{
				FieldValueDefinition: lsv1alpha1.FieldValueDefinition{
					Name:       "dummyTargetExport",
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
