package tests

import (
	"bytes"
	"context"
	"fmt"
	"io/ioutil"
	"os"
	"path"
	"strings"

	componentcli "github.com/gardener/component-cli/pkg/commands/componentarchive/remote"
	"github.com/gardener/component-cli/pkg/commands/componentarchive/resources"

	cdv2 "github.com/gardener/component-spec/bindings-go/apis/v2"
	lsv1alpha1 "github.com/gardener/landscaper/apis/core/v1alpha1"
	"github.com/gardener/landscaper/pkg/api"
	"github.com/stretchr/testify/assert"
	yamlv3 "gopkg.in/yaml.v3"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime/serializer"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/yaml"

	"github.com/gardener/landscapercli/cmd/installations"
	"github.com/gardener/landscapercli/cmd/targets/types"
	inttestutil "github.com/gardener/landscapercli/integration-test/util"
	"github.com/gardener/landscapercli/pkg/util"
)

func RunInstallationsCreateTest(k8sClient client.Client, config *inttestutil.Config) error {
	const (
		installationName = "test-installation"
		componentName    = "github.com/dummy-cd"
		componentVersion = "v0.1.0"
		blueprintName    = "dummy-blueprint"
		targetName       = "test-target"
	)

	test := installationsCreateTest{
		k8sClient:        k8sClient,
		installationName: installationName,
		componentName:    componentName,
		componentVersion: componentVersion,
		blueprintName:    blueprintName,
		targetName:       targetName,
		config:           *config,
	}

	err := test.setup()
	if err != nil {
		return fmt.Errorf("setup failed: %w", err)
	}

	err = test.run()
	if err != nil {
		// do not cleanup after erroneous test run to keep failed resources on the cluster
		return fmt.Errorf("test failed: %w", err)
	}

	err = test.teardown()
	if err != nil {
		return fmt.Errorf("teardown failed: %w", err)
	}
	return nil
}

type installationsCreateTest struct {
	k8sClient           client.Client
	installationName    string
	installationDir     string
	installationToApply lsv1alpha1.Installation
	componentName       string
	componentVersion    string
	blueprintName       string
	targetName          string
	config              inttestutil.Config
}

func (t *installationsCreateTest) run() error {
	err := t.createAndUploadComponent()
	if err != nil {
		return fmt.Errorf("creating/uploading component failed: %w", err)
	}

	cmdOutput, err := t.runInstallationsCreateCmd()
	if err != nil {
		return fmt.Errorf("landscaper-cli installations create failed: %w", err)
	}

	err = t.checkInstallation(cmdOutput)
	if err != nil {
		return fmt.Errorf("error checking generated installation: %w", err)
	}

	err = t.writeInstallationToFile(cmdOutput)
	if err != nil {
		return fmt.Errorf("cannot write generated installation to file: %w", err)
	}

	err = t.setImportParameters()
	if err != nil {
		return fmt.Errorf("setting import parameters for installation failed: %w", err)
	}

	err = t.createTarget()
	if err != nil {
		return fmt.Errorf("creating target failed: %w", err)
	}

	err = t.applyToCluster()
	if err != nil {
		return fmt.Errorf("apply to cluster failed: %w", err)
	}

	err = t.waitForInstallationSuccess()
	if err != nil {
		return fmt.Errorf("waiting for installation success failed: %w", err)
	}
	return nil
}

func (t *installationsCreateTest) writeInstallationToFile(cmdOutput *bytes.Buffer) error {
	installationDir, err := ioutil.TempDir(".", "dummy-installation-*")
	if err != nil {
		return fmt.Errorf("cannot create temp directory: %w", err)
	}

	t.installationDir = installationDir

	// The uploaded cd of the component will have 2 repository contexts
	//   - [0]: oci-registry.landscaper.svc.cluster.local:5000 (initially defined)
	//   - [1]: localhost:5000 (appended by component cli during upload)
	// "installations create" cmd will use the last context from that list
	// However, we must set the URL to the first one for the installation to work
	installationStr := strings.Replace(cmdOutput.String(), "localhost:5000", t.config.RegistryBaseURL, -1)

	err = ioutil.WriteFile(path.Join(t.installationDir, "installation-generated.yaml"), []byte(installationStr), os.ModePerm)
	if err != nil {
		return fmt.Errorf("cannot write to file: %w", err)
	}

	return nil
}

func (t *installationsCreateTest) runInstallationsCreateCmd() (*bytes.Buffer, error) {
	ctx := context.TODO()

	fmt.Println("Executing landscaper-cli installations create")
	cmd := installations.NewCreateCommand(ctx)
	outBuf := bytes.Buffer{}
	cmd.SetOut(&outBuf)
	args := []string{
		"localhost:5000",
		t.componentName,
		t.componentVersion,
		"--name",
		t.installationName,
		"--allow-plain-http",
	}
	cmd.SetArgs(args)

	err := cmd.Execute()
	if err != nil {
		return nil, err
	}

	return &outBuf, nil
}

func (t *installationsCreateTest) setImportParameters() error {
	ctx := context.TODO()

	fmt.Println("Executing landscaper-cli installations set-import-parameters")
	cmdImportParams := installations.NewSetImportParametersCommand(ctx)
	outBufImportParams := &bytes.Buffer{}
	cmdImportParams.SetOut(outBufImportParams)
	argsImportParams := []string{
		path.Join(t.installationDir, "installation-generated.yaml"),
		"appnamespace=" + t.config.TestNamespace,
		"-o=" + path.Join(t.installationDir, "installation-set-import-params.yaml"),
	}
	cmdImportParams.SetArgs(argsImportParams)

	err := cmdImportParams.Execute()
	if err != nil {
		return fmt.Errorf("landscaper-cli installations set-import-parameters failed: %w", err)
	}
	return nil
}

func (t *installationsCreateTest) createTarget() error {
	ctx := context.TODO()

	fmt.Println("Executing landscaper-cli targets create kubernetes-cluster")
	cmdTargetCreate := types.NewKubernetesClusterCommand(ctx)
	outBufTargetCreate := &bytes.Buffer{}
	cmdTargetCreate.SetOut(outBufTargetCreate)
	argsTargetCreateParams := []string{
		"--name=" + t.targetName,
		"--namespace=" + t.config.TestNamespace,
		"--target-kubeconfig=" + t.config.Kubeconfig,
	}
	cmdTargetCreate.SetArgs(argsTargetCreateParams)

	err := cmdTargetCreate.Execute()
	if err != nil {
		return fmt.Errorf("landscaper-cli targets create kubernetes-cluster failed: %w", err)
	}

	targetFromCommand := lsv1alpha1.Target{}
	if err = yaml.Unmarshal(outBufTargetCreate.Bytes(), &targetFromCommand); err != nil {
		return fmt.Errorf("cannot decode temp target: %w", err)
	}
	err = t.k8sClient.Create(ctx, &targetFromCommand, &client.CreateOptions{})
	if err != nil {
		return fmt.Errorf("cannot create target: %w", err)
	}
	return nil
}

func (t *installationsCreateTest) applyToCluster() error {
	ctx := context.TODO()

	fmt.Printf("Preparing installation")
	installationFileData, err := ioutil.ReadFile(path.Join(t.installationDir, "installation-set-import-params.yaml"))
	if err != nil {
		return fmt.Errorf("cannot read temp installation file: %w", err)
	}
	t.installationToApply = lsv1alpha1.Installation{}
	if _, _, err := serializer.NewCodecFactory(api.LandscaperScheme).UniversalDecoder().Decode(installationFileData, nil, &t.installationToApply); err != nil {
		return fmt.Errorf("cannot decode temp installation: %w", err)
	}

	fmt.Printf("Creating installation %s in namespace %s\n", t.installationToApply.Name, t.config.TestNamespace)
	t.installationToApply.ObjectMeta.Namespace = t.config.TestNamespace
	t.installationToApply.Spec.Imports.Targets[0].Target = "#" + t.targetName
	err = t.k8sClient.Create(ctx, &t.installationToApply, &client.CreateOptions{})
	if err != nil {
		return fmt.Errorf("cannot create installation: %w", err)
	}
	return nil
}

func (t *installationsCreateTest) waitForInstallationSuccess() error {
	//check if installation has status succeeded
	fmt.Printf("Wait for installation %s in namespace %s to succeed\n", t.installationToApply.Name, t.config.TestNamespace)
	timeout, err := util.CheckAndWaitUntilLandscaperInstallationSucceeded(t.k8sClient, client.ObjectKey{Name: t.installationToApply.Name, Namespace: t.installationToApply.Namespace}, t.config.SleepTime, t.config.MaxRetries)
	if err != nil {
		return fmt.Errorf("error while waiting for installation to succeed: %w", err)
	}
	if timeout {
		return fmt.Errorf("timeout at waiting for installation")
	}
	return nil
}

func (t *installationsCreateTest) checkInstallation(outBuf *bytes.Buffer) error {
	fmt.Println("Checking generated installation")

	// The uploaded cd of the component will have 2 repository contexts
	//   - [0]: oci-registry.landscaper.svc.cluster.local:5000 (initially defined)
	//   - [1]: localhost:5000 (appended by component cli during upload)
	// "installations create" cmd will use the last context from that list
	repoCtx, _ := cdv2.NewUnstructured(cdv2.NewOCIRegistryRepository("localhost:5000", ""))
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
					Version:           t.componentVersion,
					ComponentName:     t.componentName,
					RepositoryContext: &repoCtx,
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
						Name: "appnamespace",
					},
				},
				Targets: []lsv1alpha1.TargetImport{
					{
						Name: "cluster",
					},
				},
			},
		},
	}

	actualInstallation := lsv1alpha1.Installation{}
	err := yaml.Unmarshal(outBuf.Bytes(), &actualInstallation)
	if err != nil {
		return fmt.Errorf("cannot unmarshal cmd output into installation: %w", err)
	}

	// exclude raw fields from check, because the order of fields can vary
	expectedInstallation.Spec.ComponentDescriptor.Reference.RepositoryContext.Raw = nil
	actualInstallation.Spec.ComponentDescriptor.Reference.RepositoryContext.Raw = nil

	ok := assert.Equal(inttestutil.DummyTestingT{}, expectedInstallation, actualInstallation)
	if !ok {
		return fmt.Errorf("expected installation does not match with actual installation")
	}

	rootNode := &yamlv3.Node{}
	err = yamlv3.Unmarshal(outBuf.Bytes(), rootNode)
	if err != nil {
		return err
	}

	_, dataImportsNode := util.FindNodeByPath(rootNode, "spec.imports.data")
	expectedSchemaComment := `# JSON schema
# {
#   "type": "string"
# }`
	ok = assert.Equal(inttestutil.DummyTestingT{}, expectedSchemaComment, dataImportsNode.Content[0].Content[0].HeadComment)
	if !ok {
		return fmt.Errorf("schema comments for spec.imports.data are invalid")
	}

	_, targetImportsNode := util.FindNodeByPath(rootNode, "spec.imports.targets")
	expectedTargetComment := "# Target type: landscaper.gardener.cloud/kubernetes-cluster"
	ok = assert.Equal(inttestutil.DummyTestingT{}, expectedTargetComment, targetImportsNode.Content[0].Content[0].HeadComment)
	if !ok {
		return fmt.Errorf("schema comments for spec.imports.targets are invalid")
	}

	return nil
}

func (t *installationsCreateTest) createBlueprint() string {
	return `
apiVersion: landscaper.gardener.cloud/v1alpha1
kind: Blueprint

imports:
- name: cluster
  targetType: landscaper.gardener.cloud/kubernetes-cluster
- name: appnamespace
  schema:
    type: string

deployExecutions:
- name: default
  type: GoTemplate
  template: |
    deployItems:
    - name: deploy
      type: landscaper.gardener.cloud/helm
      target:
        name: {{ .imports.cluster.metadata.name }}
        namespace: {{ .imports.cluster.metadata.namespace }}
      config:
        apiVersion: helm.deployer.landscaper.gardener.cloud/v1alpha1
        kind: ProviderConfiguration

        chart:
          {{ $resource := getResource .cd "name" "echo-server-chart" }}
          ref: {{ $resource.access.imageReference }}

        updateStrategy: patch

        name: test-name
        namespace: {{ .imports.appnamespace }} `
}

func (t *installationsCreateTest) createAndUploadComponent() error {
	fmt.Println("Creating and uploading blueprint component")

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

	cd := inttestutil.CreateComponentDescriptor(t.componentName, t.componentVersion, t.config.RegistryBaseURL)

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

	marshaledBp := t.createBlueprint()

	err = ioutil.WriteFile(path.Join(bpDir, "blueprint.yaml"), []byte(marshaledBp), os.ModePerm)
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
type: helm
name: echo-server-chart
version: v0.1.0
relation: local
access:
  type: ociRegistry
  imageReference: %s/echo-server-chart:v1.1.0
---
`, t.blueprintName, t.componentVersion, t.config.RegistryBaseURL)

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

	cmdPush := componentcli.NewPushCommand(ctx)
	outBufPush := &bytes.Buffer{}
	cmdPush.SetOut(outBufPush)
	argsPush := []string{
		"localhost:5000",
		cd.Name,
		cd.Version,
		cdDir,
	}
	cmdPush.SetArgs(argsPush)

	err = cmdPush.Execute()
	if err != nil {
		return fmt.Errorf("component-cli component-archive remote push failed: %w", err)
	}

	return nil
}

func (t *installationsCreateTest) setup() error {
	err := util.DeleteNamespace(t.k8sClient, t.config.TestNamespace, t.config.SleepTime, t.config.MaxRetries)
	if err != nil {
		return fmt.Errorf("cannot delete namespace %s: %w", t.config.TestNamespace, err)
	}

	fmt.Printf("Creating namespace %s\n", t.config.TestNamespace)
	namespace := &corev1.Namespace{
		ObjectMeta: metav1.ObjectMeta{
			Name: t.config.TestNamespace,
		},
	}
	ctx := context.TODO()

	//create namespace
	err = t.k8sClient.Create(ctx, namespace, &client.CreateOptions{})
	if err != nil {
		return fmt.Errorf("cannot create namespace %s: %w", t.config.TestNamespace, err)
	}
	return nil
}

func (t *installationsCreateTest) teardown() error {
	removeErr := os.RemoveAll(t.installationDir)
	if removeErr != nil {
		fmt.Printf("cannot remove temporary directory %s: %s", t.installationDir, removeErr.Error())
	}

	err := util.DeleteNamespace(t.k8sClient, t.config.TestNamespace, t.config.SleepTime, t.config.MaxRetries)
	if err != nil {
		return fmt.Errorf("cannot delete namespace %s: %w", t.config.TestNamespace, err)
	}
	return nil
}
