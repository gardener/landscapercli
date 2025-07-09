package main

import (
	"bytes"
	"context"
	"flag"
	"fmt"
	"os"
	"os/exec"
	"path"
	"text/template"
	"time"

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

	// Plug new test cases in here:
	// 1. Create new file in ./tests directory, which exports a single function for running your test.
	//    Your test should perform a cleanup before and after running.
	//    For an example, see ./tests/test_quickstart_install.go.
	// 2. Call your new test from here.
	//
	// fmt.Println("========== Run......Test() ==========")
	// err = tests.Run......Test(k8sClient, target.DeepCopy(), config)
	// if err != nil {
	//     return fmt.Errorf("Run......Test() failed: %w", err)
	// }

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

	cfg, err := clientcmd.BuildConfigFromFlags("", config.Kubeconfig)
	if err != nil {
		return fmt.Errorf("cannot parse K8s config: %w", err)
	}

	fmt.Println("========== Cleaning up before test ==========")
	if err := runQuickstartUninstall(config); err != nil {
		return fmt.Errorf("landscaper-cli quickstart uninstall failed: %w", err)
	}

	fmt.Println("Waiting for resources to be deleted on the K8s cluster...")
	time.Sleep(10 * time.Second)

	fmt.Println("========== Running landscaper-cli quickstart install ==========")
	if err := runQuickstartInstall(config); err != nil {
		return fmt.Errorf("landscaper-cli quickstart install failed: %w", err)
	}

	k8sClient, err := client.New(cfg, client.Options{
		Scheme: scheme,
	})
	if err != nil {
		return fmt.Errorf("cannot build K8s client: %w", err)
	}

	fmt.Println("Waiting for pods to get ready")
	timeout, err := util.CheckAndWaitUntilAllPodsAreReady(k8sClient, config.LandscaperNamespace, config.SleepTime, config.MaxRetries)
	if err != nil {
		return fmt.Errorf("error while waiting for pods: %w", err)
	}
	if timeout {
		return fmt.Errorf("timeout while waiting for pods")
	}

	fmt.Println("========== Fetching ingress to OCI registry ==========")
	ingressUrl, err := util.CheckIngressReady(k8sClient, config.LandscaperNamespace, config.SleepTime, config.MaxRetries)
	if err != nil {
		return fmt.Errorf("error fetching ingress url: %w", err)
	}
	config.ExternalRegistryBaseURL = ingressUrl
	fmt.Println("External registry base URL: " + config.ExternalRegistryBaseURL)

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

	fmt.Println("========== Uploading test helm chart to OCI registry ==========")
	helmChartRef, err := uploadTestHelmChart(config.ExternalRegistryBaseURL)
	if err != nil {
		return fmt.Errorf("upload of test helm chart failed: %w", err)
	}

	target, _, err := util.BuildKubernetesClusterTarget("test-target", "", config.Kubeconfig, "")
	if err != nil {
		return fmt.Errorf("cannot build target: %w", err)
	}

	fmt.Println("========== Starting test suite ==========")
	if err := runTestSuite(k8sClient, config, target, helmChartRef); err != nil {
		return fmt.Errorf("runTestSuite() failed: %w", err)
	}
	fmt.Println("========== Test suite finished successfully ==========")

	fmt.Println("========== Cleaning up after test ==========")
	if err := runQuickstartUninstall(config); err != nil {
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

func uploadTestHelmChart(externalRegistryBaseURL string) (string, error) {
	chartDir := path.Join(".", "testdata", "01", "chart")

	tempDir, err := os.MkdirTemp(".", "landscaper-chart-tmp2-*")
	if err != nil {
		return "", err
	}
	defer func() {
		if err := os.RemoveAll(tempDir); err != nil {
			fmt.Printf("cannot remove temporary directory %s: %s", tempDir, err.Error())
		}
	}()

	err = util.ExecCommandBlocking(fmt.Sprintf("helm package %s -d %s", chartDir, tempDir))
	if err != nil {
		return "", fmt.Errorf("helm package failed: %w", err)
	}

	if err := util.ExecCommandBlocking(fmt.Sprintf("helm push %s/test-chart-v0.1.0.tgz oci://localhost:5000", tempDir)); err != nil {
		return "", fmt.Errorf("helm push failed: %w", err)
	}

	helmChartRef := externalRegistryBaseURL + "/test-chart:v0.1.0"
	return helmChartRef, nil
}

func runQuickstartUninstall(config *inttestutil.Config) error {
	uninstallArgs := []string{
		"--kubeconfig",
		config.Kubeconfig,
		"--namespace",
		config.LandscaperNamespace,
		"--delete-namespace",
		"--delete-crd",
	}
	uninstallCmd := quickstart.NewUninstallCommand(context.TODO())
	uninstallCmd.SetArgs(uninstallArgs)

	if err := uninstallCmd.Execute(); err != nil {
		return fmt.Errorf("uninstall command failed: %w", err)
	}

	return nil
}

func runQuickstartInstall(config *inttestutil.Config) error {
	landscaperValues, err := buildLandscaperValues(config.LandscaperNamespace)
	if err != nil {
		return fmt.Errorf("cannot template landscaper values: %w", err)
	}

	tmpFile, err := os.CreateTemp(".", "landscaper-values-")
	if err != nil {
		return fmt.Errorf("cannot create temporary file: %w", err)
	}
	defer func() {
		if err := os.Remove(tmpFile.Name()); err != nil {
			fmt.Printf("Cannot remove temporary file %s: %s", tmpFile.Name(), err.Error())
		}
	}()

	if err := os.WriteFile(tmpFile.Name(), []byte(landscaperValues), os.ModePerm); err != nil {
		return fmt.Errorf("cannot write to file: %w", err)
	}

	installArgs := []string{
		"--kubeconfig",
		config.Kubeconfig,
		"--install-oci-registry",
		"--install-registry-ingress",
		"--registry-username",
		inttestutil.Testuser,
		"--registry-password",
		inttestutil.Testpw,
		"--namespace",
		config.LandscaperNamespace,
	}
	installCmd := quickstart.NewInstallCommand(context.TODO())
	installCmd.SetArgs(installArgs)

	if err := installCmd.Execute(); err != nil {
		return fmt.Errorf("install command failed: %w", err)
	}

	return nil
}

func buildLandscaperValues(namespace string) ([]byte, error) {
	const valuesTemplate = `
landscaper:
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
	if err := t.Execute(b, data); err != nil {
		return nil, fmt.Errorf("could not template helm values: %w", err)
	}

	return b.Bytes(), nil
}
