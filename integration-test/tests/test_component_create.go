package tests

import (
	"bytes"
	"context"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/gardener/component-cli/pkg/commands/componentarchive/remote"
	"github.com/gardener/component-cli/pkg/commands/componentarchive/resources"
	cdv2 "github.com/gardener/component-spec/bindings-go/apis/v2"
	lsv1alpha1 "github.com/gardener/landscaper/apis/core/v1alpha1"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"sigs.k8s.io/controller-runtime/pkg/client"

	"github.com/gardener/landscapercli/cmd/components"
	inttestutil "github.com/gardener/landscapercli/integration-test/util"
	"github.com/gardener/landscapercli/pkg/util"
)

func RunComponentCreateTest(k8sClient client.Client, target *lsv1alpha1.Target, config *inttestutil.Config) error {
	const (
		componentName    = "github.com/gardener/landscapercli/nginx"
		componentVersion = "v0.1.0"
		testRootDir      = "landscapercli-integrationtest"
		componentDirName = "demo-component"
		installationName = "test-installation"
	)

	test := componentCreateTest{
		k8sClient:        k8sClient,
		target:           target,
		config:           config,
		componentName:    componentName,
		componentVersion: componentVersion,
		testRootDir:      testRootDir,
		componentDir:     filepath.Join(testRootDir, componentDirName),
		installationName: installationName,
	}

	err := test.run()
	if err != nil {
		// do not cleanup after erroneous test run to keep failed resources on the cluster
		return fmt.Errorf("test failed: %w", err)
	}

	return nil
}

type componentCreateTest struct {
	k8sClient        client.Client
	target           *lsv1alpha1.Target
	config           *inttestutil.Config
	componentName    string
	componentVersion string
	testRootDir      string
	componentDir     string
	installationName string
}

func (t *componentCreateTest) run() error {
	ctx := context.TODO()

	err := t.cleanup(ctx)
	if err != nil {
		return err
	}

	err = t.createComponent(ctx)
	if err != nil {
		return err
	}

	err = t.addHelmDeployItemWithExternalChart(ctx)
	if err != nil {
		return err
	}

	err = t.addManifestDeployItem(ctx)
	if err != nil {
		return err
	}

	err = t.addContainerDeployItem(ctx)
	if err != nil {
		return err
	}

	err = t.addResources(ctx)
	if err != nil {
		return err
	}

	err = t.setBaseURL(ctx)
	if err != nil {
		return err
	}

	err = t.pushComponent(ctx)
	if err != nil {
		return err
	}

	err = t.createNamespace(ctx)
	if err != nil {
		return err
	}

	err = t.createTarget(ctx)
	if err != nil {
		return err
	}

	err = t.createInstallation(ctx)
	if err != nil {
		return err
	}

	err = t.waitUntilInstallationSucceeded(ctx)
	if err != nil {
		return err
	}

	err = t.cleanup(ctx)
	if err != nil {
		return err
	}

	return nil
}

func (t *componentCreateTest) createComponent(ctx context.Context) error {
	fmt.Printf("Creating skeleton for component %s\n", t.componentName)
	cmd := components.NewCreateCommand(ctx)
	outBuf := &bytes.Buffer{}
	cmd.SetOut(outBuf)
	args := []string{
		t.componentName,
		t.componentVersion,
		"--component-directory",
		t.componentDir,
	}
	cmd.SetArgs(args)

	err := cmd.Execute()
	if err != nil {
		return fmt.Errorf("landscaper-cli component create failed: %w", err)
	}

	return nil
}

func (t *componentCreateTest) addHelmDeployItemWithExternalChart(ctx context.Context) error {
	fmt.Println("Adding external helm deploy item")
	cmd := components.NewAddHelmLSDeployItemCommand(ctx)
	outBuf := &bytes.Buffer{}
	cmd.SetOut(outBuf)
	args := []string{
		"nginx",
		"--component-directory",
		t.componentDir,
		"--oci-reference",
		"eu.gcr.io/gardener-project/landscaper/tutorials/charts/ingress-nginx:4.0.17",
		"--resource-version",
		"v0.2.0",
		"--cluster-param",
		"target-cluster",
		"--target-ns-param",
		"nginx-namespace",
	}
	cmd.SetArgs(args)

	err := cmd.Execute()
	if err != nil {
		return fmt.Errorf("adding external helm deploy item failed: %w", err)
	}

	return nil
}

func (t *componentCreateTest) addManifestDeployItem(ctx context.Context) error {
	fmt.Println("Adding manifest deploy item")

	secretPath1, err := t.createSecret(ctx, "secret1", "password-1")
	if err != nil {
		return err
	}

	secretPath2, err := t.createSecret(ctx, "secret2", "password-2")
	if err != nil {
		return err
	}

	cmd := components.NewAddManifestDeployItemCommand(ctx)
	outBuf := &bytes.Buffer{}
	cmd.SetOut(outBuf)
	args := []string{
		"secret",
		"--component-directory",
		t.componentDir,
		"--manifest-file",
		secretPath1,
		"--manifest-file",
		secretPath2,
		"--import-param",
		"password-1:string",
		"--import-param",
		"password-2:string",
		"--cluster-param",
		"target-cluster",
	}
	cmd.SetArgs(args)

	err = cmd.Execute()
	if err != nil {
		return fmt.Errorf("adding manifest deploy item failed: %w", err)
	}

	return nil
}

func (t *componentCreateTest) addContainerDeployItem(ctx context.Context) error {
	fmt.Println("Adding container deploy item")

	const (
		image   = "eu.gcr.io/sap-gcp-cp-k8s-stable-hub/examples/landscaper/integrationtests/images/containerexample:0.1.0"
		command = "./script.sh"
	)

	cmd := components.NewAddContainerDeployItemCommand(ctx)
	outBuf := &bytes.Buffer{}
	cmd.SetOut(outBuf)
	args := []string{
		"containertestitem",
		"--resource-version", "0.1.0",
		"--component-directory", t.componentDir,
		"--cluster-param", "target-cluster",
		"--import-param", "word:string",
		"--import-param", "sleepTimeBefore:integer",
		"--import-param", "sleepTimeAfter:integer",
		"--export-param", "sentence:string",
		"--image", image,
		"--command", command,
	}
	cmd.SetArgs(args)

	err := cmd.Execute()
	if err != nil {
		return fmt.Errorf("adding container deploy item failed: %w", err)
	}

	return nil
}

func (t *componentCreateTest) createSecret(ctx context.Context, secretName, passwordParamName string) (string, error) {
	secretPath := filepath.Join(t.testRootDir, secretName+".yaml")

	fmt.Printf("Writing secret to file %s\n", secretPath)

	f, err := os.Create(secretPath)
	if err != nil {
		return "", fmt.Errorf("creating secret failed: %w", err)
	}

	defer f.Close()

	secretFormat := `
apiVersion: v1
kind: Secret
metadata:
  name: %s
  namespace: %s
type: kubernetes.io/basic-auth
stringData:
  username: admin
  password: %s
`

	_, err = f.WriteString(fmt.Sprintf(secretFormat, secretName, t.config.TestNamespace, passwordParamName))
	if err != nil {
		return "", fmt.Errorf("writing secret failed: %w", err)
	}

	return secretPath, nil
}

func (t *componentCreateTest) addResources(ctx context.Context) error {
	fmt.Println("Adding resources to component descriptor")
	cmd := resources.NewAddCommand(ctx)
	outBuf := &bytes.Buffer{}
	cmd.SetOut(outBuf)
	args := []string{
		t.componentDir,
		"-r",
		t.componentDir + "/resources.yaml",
	}
	cmd.SetArgs(args)

	err := cmd.Execute()
	if err != nil {
		return fmt.Errorf("adding resources to component descriptor failed: %w", err)
	}

	return nil

}

func (t *componentCreateTest) setBaseURL(ctx context.Context) error {
	fmt.Println("Set base url in component descriptor")

	componentDescriptorPath := filepath.Join(t.componentDir, "component-descriptor.yaml")

	data, err := os.ReadFile(componentDescriptorPath)
	if err != nil {
		return fmt.Errorf("could not read component descriptor: %w", err)
	}

	oldCdString := string(data)
	baseURLElement := fmt.Sprintf(`baseUrl: "%s"`, t.config.RegistryBaseURL)
	newCdString := strings.Replace(oldCdString, `baseUrl: ""`, baseURLElement, 1)

	err = os.WriteFile(componentDescriptorPath, []byte(newCdString), 0755)
	if err != nil {
		return fmt.Errorf("could not write component descriptor: %w", err)
	}

	return nil

}

func (t *componentCreateTest) pushComponent(ctx context.Context) error {
	fmt.Printf("Pushing component %s\n", t.componentName)
	cmd := remote.NewPushCommand(ctx)
	outBuf := &bytes.Buffer{}
	cmd.SetOut(outBuf)
	args := []string{
		"localhost:5000",
		t.componentName,
		t.componentVersion,
		t.componentDir,
	}
	cmd.SetArgs(args)

	err := cmd.Execute()
	if err != nil {
		return fmt.Errorf("pushing component failed: %w", err)
	}

	return nil
}

func (t *componentCreateTest) createNamespace(ctx context.Context) error {
	fmt.Printf("Creating namespace %s\n", t.config.TestNamespace)

	namespace := &corev1.Namespace{
		ObjectMeta: metav1.ObjectMeta{
			Name: t.config.TestNamespace,
		},
	}

	err := t.k8sClient.Create(ctx, namespace)
	if err != nil {
		return fmt.Errorf("cannot create namespace %s: %w", t.config.TestNamespace, err)
	}

	return nil
}

func (t *componentCreateTest) createTarget(ctx context.Context) error {
	fmt.Printf("Creating target %s\n", t.target.Name)

	target := lsv1alpha1.Target{
		ObjectMeta: metav1.ObjectMeta{
			Name:      t.target.Name,
			Namespace: t.config.TestNamespace,
		},
		Spec: t.target.Spec,
	}

	err := t.k8sClient.Create(ctx, &target)
	if err != nil {
		return fmt.Errorf("cannot create target: %w", err)
	}

	return nil
}

func (t *componentCreateTest) createInstallation(ctx context.Context) error {
	fmt.Printf("Creating installation %s\n", t.installationName)

	repoCtx, _ := cdv2.NewUnstructured(cdv2.NewOCIRegistryRepository(t.config.RegistryBaseURL, ""))
	installation := lsv1alpha1.Installation{
		ObjectMeta: metav1.ObjectMeta{
			Name:      t.installationName,
			Namespace: t.config.TestNamespace,
			Annotations: map[string]string{
				lsv1alpha1.OperationAnnotation: string(lsv1alpha1.ReconcileOperation),
			},
		},
		Spec: lsv1alpha1.InstallationSpec{
			ComponentDescriptor: &lsv1alpha1.ComponentDescriptorDefinition{
				Reference: &lsv1alpha1.ComponentDescriptorReference{
					RepositoryContext: &repoCtx,
					ComponentName:     t.componentName,
					Version:           t.componentVersion,
				},
			},
			Blueprint: lsv1alpha1.BlueprintDefinition{
				Reference: &lsv1alpha1.RemoteBlueprintReference{
					ResourceName: "blueprint",
				},
			},
			Imports: lsv1alpha1.InstallationImports{
				Targets: []lsv1alpha1.TargetImport{
					{
						Name:   "target-cluster",
						Target: "#" + t.target.Name,
					},
				},
			},
			ImportDataMappings: map[string]lsv1alpha1.AnyJSON{
				"nginx-namespace": {RawMessage: []byte(`"` + t.config.TestNamespace + `"`)},
				"password-1":      {RawMessage: []byte(`"pw1"`)},
				"password-2":      {RawMessage: []byte(`"pw2"`)},
				"word":            {RawMessage: []byte(`"test"`)},
				"sleepTimeBefore": {RawMessage: []byte(`0`)},
				"sleepTimeAfter":  {RawMessage: []byte(`0`)},
			},
		},
	}

	err := t.k8sClient.Create(ctx, &installation)
	if err != nil {
		return fmt.Errorf("creating installation failed: %w", err)
	}

	return nil
}

func (t *componentCreateTest) waitUntilInstallationSucceeded(ctx context.Context) error {
	fmt.Printf("Waiting until installation %s has succeeded\n", t.installationName)

	// Returns (true, nil)  if the installation has succeeded.
	// Returns (false, nil) if the installation has not yet succeeded.
	// Returns (false, err) if the check has failed whether or not the installation succeeded.
	conditionFunc := func() (ok bool, err error) {
		key := client.ObjectKey{
			Namespace: t.config.TestNamespace,
			Name:      t.installationName,
		}
		installation := lsv1alpha1.Installation{}
		err = t.k8sClient.Get(ctx, key, &installation)
		if err != nil {
			return false, fmt.Errorf("failed fetching installation: %w", err)
		}

		return installation.Status.InstallationPhase == lsv1alpha1.InstallationPhases.Succeeded, nil
	}

	timeout, err := util.CheckConditionPeriodically(conditionFunc, t.config.SleepTime, t.config.MaxRetries)
	if err != nil {
		return fmt.Errorf("cannot check installation phase: %w", err)
	}
	if timeout {
		return fmt.Errorf("installation did not succeed within the given time")
	}

	fmt.Println("Installation has succeeded")
	return nil
}

func (t *componentCreateTest) cleanup(ctx context.Context) error {
	fmt.Println("Cleanup")

	fmt.Println("- Removing test directory")
	err := os.RemoveAll(t.testRootDir)
	if err != nil {
		return fmt.Errorf("Cleanup failed when deleting component directory: %w", err)
	}

	fmt.Printf("- Removing test namespace %s\n", t.config.TestNamespace)
	err = util.DeleteNamespace(t.k8sClient, t.config.TestNamespace, t.config.SleepTime, t.config.MaxRetries)
	if err != nil {
		return fmt.Errorf("Cleanup failed when deleting namespace: %w", err)
	}

	return nil
}
