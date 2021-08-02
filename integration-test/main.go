package main

import (
	"bytes"
	"context"
	"flag"
	"fmt"
	"io/ioutil"
	"os"
	"os/exec"
	"text/template"
	"time"

	componentclilog "github.com/gardener/component-cli/pkg/logger"
	lsv1alpha1 "github.com/gardener/landscaper/apis/core/v1alpha1"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/labels"
	"k8s.io/apimachinery/pkg/runtime"
	clientgoscheme "k8s.io/client-go/kubernetes/scheme"
	"k8s.io/client-go/tools/clientcmd"
	"sigs.k8s.io/controller-runtime/pkg/client"

	"github.com/gardener/landscapercli/cmd/quickstart"
	"github.com/gardener/landscapercli/integration-test/tests"
	inttestutil "github.com/gardener/landscapercli/integration-test/util"
	"github.com/gardener/landscapercli/pkg/logger"
	"github.com/gardener/landscapercli/pkg/util"
)

var (
	scheme = runtime.NewScheme()
)

func init() {
	_ = clientgoscheme.AddToScheme(scheme)
	_ = lsv1alpha1.AddToScheme(scheme)
}

func runTestSuite(k8sClient client.Client, config *inttestutil.Config, target *lsv1alpha1.Target, helmChartRef string) error {
	fmt.Println("========== RunQuickstartInstallTest() ==========")
	err := tests.RunQuickstartInstallTest(k8sClient, target.DeepCopy(), helmChartRef, config)
	if err != nil {
		return fmt.Errorf("RunQuickstartInstallTest() failed: %w", err)
	}

	fmt.Println("========== RunInstallationsCreateTest() ==========")
	err = tests.RunInstallationsCreateTest(k8sClient, config)
	if err != nil {
		return fmt.Errorf("RunInstallationsCreateTest() failed: %w", err)
	}

	fmt.Println("========== RunComponentCreateTest() ==========")
	err = tests.RunComponentCreateTest(k8sClient, target.DeepCopy(), config)
	if err != nil {
		return fmt.Errorf("RunComponentCreateTest() failed: %w", err)
	}

	// Plug new test cases in here:
	// 1. Create new file in ./tests directory, which exports a single function for running your test.
	//    Your test should perform a cleanup before and after running.
	//    For an example, see ./tests/test_quickstart_install.go.
	// 2. Call your new test from here.

	return nil
}

func main() {
	fmt.Println("========== Starting integration-test ==========")

	err := run()
	if err != nil {
		fmt.Println("Integration-test finished with error:", err)
		os.Exit(1)
	}

	fmt.Println("========== Integration-test finished successfully ==========")
}

func run() error {
	config := parseConfig()

	log, err := logger.NewCliLogger()
	if err != nil {
		return fmt.Errorf("cannot create logger: %w", err)
	}
	logger.SetLogger(log)
	componentclilog.SetLogger(log)

	cfg, err := clientcmd.BuildConfigFromFlags("", config.Kubeconfig)
	if err != nil {
		return fmt.Errorf("cannot parse K8s config: %w", err)
	}

	k8sClient, err := client.New(cfg, client.Options{
		Scheme: scheme,
	})
	if err != nil {
		return fmt.Errorf("cannot build K8s client: %w", err)
	}

	fmt.Println("========== Cleaning up before test ==========")

	err = util.DeleteNamespace(k8sClient, config.TestNamespace, config.SleepTime, config.MaxRetries)
	if err != nil {
		return fmt.Errorf("cannot delete namespace %s: %w", config.TestNamespace, err)
	}
	err = runQuickstartUninstall(config)
	if err != nil {
		return fmt.Errorf("landscaper-cli quickstart uninstall failed: %w", err)
	}

	fmt.Println("Waiting for resources to be deleted on the K8s cluster...")
	time.Sleep(10 * time.Second)

	fmt.Println("========== Running landscaper-cli quickstart install ==========")
	err = runQuickstartInstall(config)
	if err != nil {
		return fmt.Errorf("landscaper-cli quickstart install failed: %w", err)
	}

	fmt.Println("Waiting for pods to get ready")
	timeout, err := util.CheckAndWaitUntilAllPodsAreReady(k8sClient, config.LandscaperNamespace, config.SleepTime, config.MaxRetries)
	if err != nil {
		return fmt.Errorf("error while waiting for pods: %w", err)
	}
	if timeout {
		return fmt.Errorf("timeout while waiting for pods")
	}

	fmt.Println("========== Starting port-forward to OCI registry ==========")
	resultChan := make(chan util.CmdResult)
	portforwardCmd, err := startOCIRegistryPortForward(k8sClient, config.LandscaperNamespace, config.Kubeconfig, resultChan)
	if err != nil {
		return fmt.Errorf("cannot start port-forward to OCI: %w", err)
	}

	go func() {
		result := <-resultChan
		if result.Error != nil {
			fmt.Printf("port-forward to OCI registry failed: %s:\n%s\n", result.Error.Error(), result.StdErr)
		}
	}()

	defer func() {
		// Disable port-forward
		killPortforwardErr := portforwardCmd.Process.Kill()
		if killPortforwardErr != nil {
			fmt.Println("cannot kill port-forward process:", killPortforwardErr)
		}
	}()

	// Port forwarding starts non-blocking (asynchronous), so we cant be sure it is completed.
	// Hopefully completed after 5s.
	fmt.Println("Waiting 5s for port forward to start...")
	time.Sleep(5 * time.Second)

	fmt.Println("========== Uploading echo-server helm chart to OCI registry ==========")
	helmChartRef, err := uploadEchoServerHelmChart(config.LandscaperNamespace)
	if err != nil {
		return fmt.Errorf("upload of echo-server helm chart failed: %w", err)
	}

	target, err := util.BuildKubernetesClusterTarget("test-target", "", config.Kubeconfig)
	if err != nil {
		return fmt.Errorf("cannot build target: %w", err)
	}

	fmt.Println("========== Starting test suite ==========")
	err = runTestSuite(k8sClient, config, target, helmChartRef)
	if err != nil {
		return fmt.Errorf("runTestSuite() failed: %w", err)
	}
	fmt.Println("========== Test suite finished successfully ==========")

	fmt.Println("========== Cleaning up after test ==========")
	err = runQuickstartUninstall(config)
	if err != nil {
		return fmt.Errorf("landscaper-cli quickstart uninstall failed: %w", err)
	}

	return nil
}

func parseConfig() *inttestutil.Config {
	var kubeconfig, landscaperNamespace, testNamespace string
	var maxRetries int

	flag.StringVar(&kubeconfig, "kubeconfig", "", "path to the kubeconfig of the cluster")
	flag.StringVar(&landscaperNamespace, "landscaper-namespace", "landscaper", "namespace on the cluster to setup Landscaper")
	flag.StringVar(&testNamespace, "test-namespace", "ls-cli-inttest", "namespace where the tests will be runned")
	flag.IntVar(&maxRetries, "max-retries", 10, "max retries (every 5s) for all waiting operations")
	flag.Parse()

	registryBaseURL := fmt.Sprintf("oci-registry.%s.svc.cluster.local:5000", landscaperNamespace)

	config := inttestutil.Config{
		Kubeconfig:          kubeconfig,
		LandscaperNamespace: landscaperNamespace,
		TestNamespace:       testNamespace,
		MaxRetries:          maxRetries,
		SleepTime:           10 * time.Second,
		RegistryBaseURL:     registryBaseURL,
	}

	return &config
}

func startOCIRegistryPortForward(k8sClient client.Client, namespace, kubeconfigPath string, ch chan<- util.CmdResult) (*exec.Cmd, error) {
	ctx := context.TODO()
	ociRegistryPods := corev1.PodList{}
	err := k8sClient.List(
		ctx,
		&ociRegistryPods,
		client.InNamespace(namespace),
		client.MatchingLabelsSelector{
			Selector: labels.SelectorFromSet(
				labels.Set(map[string]string{
					"app": "oci-registry",
				}),
			),
		},
	)

	if err != nil {
		return nil, fmt.Errorf("cannot list pods: %w", err)
	}

	if len(ociRegistryPods.Items) != 1 {
		return nil, fmt.Errorf("expected 1 OCI registry pod, found %d", len(ociRegistryPods.Items))
	}

	portforwardCmd, err := util.ExecCommandNonBlocking("kubectl port-forward "+ociRegistryPods.Items[0].Name+" 5000:5000 --kubeconfig "+kubeconfigPath+" --namespace "+namespace, ch)
	if err != nil {
		return nil, fmt.Errorf("kubectl port-forward failed: %w", err)
	}

	return portforwardCmd, nil
}

func uploadEchoServerHelmChart(landscaperNamespace string) (string, error) {
	err := util.ExecCommandBlocking("helm pull https://storage.googleapis.com/sap-hub-test/echo-server-1.1.0.tgz")
	if err != nil {
		return "", fmt.Errorf("helm pull failed: %w", err)
	}
	defer func() {
		err = os.Remove("echo-server-1.1.0.tgz")
		if err != nil {
			fmt.Printf("Cannot remove file echo-server-1.1.0.tgz: %s\n", err.Error())
		}
	}()

	err = util.ExecCommandBlocking("helm chart save echo-server-1.1.0.tgz localhost:5000/echo-server-chart:v1.1.0")
	if err != nil {
		return "", fmt.Errorf("helm chart save failed: %w", err)
	}

	err = util.ExecCommandBlocking("helm chart push localhost:5000/echo-server-chart:v1.1.0")
	if err != nil {
		return "", fmt.Errorf("helm chart push failed: %w", err)
	}

	helmChartRef := fmt.Sprintf("oci-registry.%s.svc.cluster.local:5000/echo-server-chart:v1.1.0", landscaperNamespace)
	return helmChartRef, nil
}

func runQuickstartUninstall(config *inttestutil.Config) error {
	uninstallArgs := []string{
		"--kubeconfig",
		config.Kubeconfig,
		"--namespace",
		config.LandscaperNamespace,
	}
	uninstallCmd := quickstart.NewUninstallCommand(context.TODO())
	uninstallCmd.SetArgs(uninstallArgs)

	err := uninstallCmd.Execute()
	if err != nil {
		return fmt.Errorf("uninstall command failed: %w", err)
	}

	return nil
}

func runQuickstartInstall(config *inttestutil.Config) error {
	landscaperValues, err := buildLandscaperValues(config.LandscaperNamespace)
	if err != nil {
		return fmt.Errorf("cannot template landscaper values: %w", err)
	}

	tmpFile, err := ioutil.TempFile(".", "landscaper-values-")
	if err != nil {
		return fmt.Errorf("cannot create temporary file: %w", err)
	}
	defer func() {
		err := os.Remove(tmpFile.Name())
		if err != nil {
			fmt.Printf("Cannot remove temporary file %s: %s", tmpFile.Name(), err.Error())
		}
	}()

	err = ioutil.WriteFile(tmpFile.Name(), []byte(landscaperValues), os.ModePerm)
	if err != nil {
		return fmt.Errorf("cannot write to file: %w", err)
	}

	installArgs := []string{
		"--kubeconfig",
		config.Kubeconfig,
		"--landscaper-values",
		tmpFile.Name(),
		"--install-oci-registry",
		"--namespace",
		config.LandscaperNamespace,
	}
	installCmd := quickstart.NewInstallCommand(context.TODO())
	installCmd.SetArgs(installArgs)

	err = installCmd.Execute()
	if err != nil {
		return fmt.Errorf("install command failed: %w", err)
	}

	return nil
}

func buildLandscaperValues(namespace string) ([]byte, error) {
	const valuesTemplate = `
landscaper:
  registryConfig: # contains optional oci secrets
    allowPlainHttpRegistries: true
    secrets: {}
  deployers:
  - container
  - helm
  - manifest
  deployerManagement:
    namespace: {{ .Namespace }}
    agent:
      namespace: {{ .Namespace }}
`

	t, err := template.New("valuesTemplate").Parse(valuesTemplate)
	if err != nil {
		return nil, err
	}

	data := struct {
		Namespace string
	}{
		Namespace: namespace,
	}

	b := &bytes.Buffer{}
	err = t.Execute(b, data)
	if err != nil {
		return nil, fmt.Errorf("could not template helm values: %w", err)
	}

	return b.Bytes(), nil
}
