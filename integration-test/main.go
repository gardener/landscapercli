package main

import (
	"context"
	"encoding/json"
	"flag"
	"fmt"
	"io/ioutil"
	"os"
	"os/exec"
	"time"

	landscaper "github.com/gardener/landscaper/pkg/apis/core/v1alpha1"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/labels"
	"k8s.io/apimachinery/pkg/runtime"
	clientgoscheme "k8s.io/client-go/kubernetes/scheme"
	"k8s.io/client-go/tools/clientcmd"
	"sigs.k8s.io/controller-runtime/pkg/client"

	"github.com/gardener/landscapercli/cmd/quickstart"
	"github.com/gardener/landscapercli/integration-test/config"
	"github.com/gardener/landscapercli/integration-test/tests"
	"github.com/gardener/landscapercli/pkg/logger"
	"github.com/gardener/landscapercli/pkg/util"
)

var (
	scheme = runtime.NewScheme()
)

func init() {
	_ = clientgoscheme.AddToScheme(scheme)
	_ = landscaper.AddToScheme(scheme)
}

func runTestSuite(k8sClient client.Client, config *config.Config, target *landscaper.Target, helmChartRef string) error {
	fmt.Println("========== RunQuickstartInstallTest() ==========")
	err := tests.RunQuickstartInstallTest(k8sClient, target, helmChartRef, config)
	if err != nil {
		return fmt.Errorf("RunQuickstartInstallTest() failed: %w", err)
	}

	// Plug new test cases in here:
	// 1. Create new file in ./tests directory, which exports a single function for running your test.
	//    Your test should perform a cleanup before and after running.
	//    For an example, see ./tests/test_quickstart_install.go.
	// 2. Call your new test from here.

	return nil
}

func main() {
	config := parseConfig()

	log, err := logger.NewCliLogger()
	if err != nil {
		fmt.Println("Cannot create logger:", err)
		os.Exit(1)
	}
	logger.SetLogger(log)

	cfg, err := clientcmd.BuildConfigFromFlags("", config.Kubeconfig)
	if err != nil {
		fmt.Println("Cannot parse K8s config:", err)
		os.Exit(1)
	}

	k8sClient, err := client.New(cfg, client.Options{
		Scheme: scheme,
	})
	if err != nil {
		fmt.Println("Cannot build K8s client:", err)
		os.Exit(1)
	}

	fmt.Println("========== Clean Up Before Test Run ==========")

	err = util.DeleteNamespace(k8sClient, config.TestNamespace, config.SleepTime, config.MaxRetries)
	if err != nil {
		fmt.Printf("Cannot delete namespace %s: %s\n", config.TestNamespace, err.Error())
		os.Exit(1)
	}
	err = runQuickstartUninstall(config)
	if err != nil {
		fmt.Println("landscaper-cli quickstart uninstall failed:", err)
		os.Exit(1)
	}

	fmt.Println("Waiting for resources to be deleted on the K8s cluster...")
	time.Sleep(10 * time.Second)

	fmt.Println("========== Running landscaper-cli quickstart install ==========")
	err = runQuickstartInstall(config)
	if err != nil {
		fmt.Println("landscaper-cli quickstart install failed: ", err)
		os.Exit(1)
	}

	fmt.Println("Waiting for Landscaper Pods to get ready")
	timeout, err := util.CheckAndWaitUntilAllPodsAreReady(k8sClient, config.LandscaperNamespace, config.SleepTime, config.MaxRetries)
	if err != nil {
		fmt.Println("error while waiting for Landscaper Pods:", err)
		os.Exit(1)
	}
	if timeout {
		fmt.Println("timeout while waiting for landscaper pods")
		os.Exit(1)
	}

	fmt.Println("========== Starting port-forward to OCI registry ==========")
	portforwardCmd, err := startOCIRegistryPortForward(k8sClient, config.LandscaperNamespace, config.Kubeconfig)
	if err != nil {
		fmt.Println("port-forward to OCI registry failed:", err)
		os.Exit(1)
	}
	defer func() {
		// Disable port-forward
		err = portforwardCmd.Process.Kill()
		if err != nil {
			fmt.Println("Cannot kill port-forward process: %w", err)
		}
	}()

	fmt.Println("========== Upload echo-server Helm Chart to OCI registry ==========")
	helmChartRef, err := uploadEchoServerHelmChart(config.LandscaperNamespace)
	if err != nil {
		fmt.Println("Upload of echo-server Helm Chart failed: %w", err)
		os.Exit(1)
	}

	target, err := buildTarget(config.Kubeconfig)
	if err != nil {
		fmt.Println("Cannot build target:", err)
		os.Exit(1)
	}

	fmt.Println("========== Starting Test Suite ==========")
	err = runTestSuite(k8sClient, config, target, helmChartRef)
	if err != nil {
		fmt.Println("runTestSuite failed:", err)
		os.Exit(1)
	}
	fmt.Println("========== Test Suite Finished Successfully ==========")

	fmt.Println("========== Clean Up After Test Run ==========")
	err = runQuickstartUninstall(config)
	if err != nil {
		fmt.Println("landscaper-cli quickstart uninstall failed:", err)
		os.Exit(1)
	}

	fmt.Println("========== Integration-test finished successfully ==========")
}

func parseConfig() *config.Config {
	var kubeconfig, landscaperNamespace, testNamespace string
	var maxRetries int

	flag.StringVar(&kubeconfig, "kubeconfig", "", "path to the kubeconfig of the cluster")
	flag.StringVar(&landscaperNamespace, "landscaper-namespace", "landscaper", "namespace on the cluster to setup Landscaper")
	flag.StringVar(&testNamespace, "test-namespace", "ls-cli-inttest", "namespace where the tests will be runned")
	flag.IntVar(&maxRetries, "max-retries", 6, "max retries (every 5s) for all waiting operations")
	flag.Parse()

	config := config.Config{
		Kubeconfig:          kubeconfig,
		LandscaperNamespace: landscaperNamespace,
		TestNamespace:       testNamespace,
		MaxRetries:          maxRetries,
		SleepTime:           5 * time.Second,
	}

	return &config
}

func buildTarget(kubeconfig string) (*landscaper.Target, error) {
	kubeconfigContent, err := ioutil.ReadFile(kubeconfig)
	if err != nil {
		return nil, fmt.Errorf("cannot read kubeconfig: %w", err)
	}

	config := map[string]interface{}{
		"kubeconfig": string(kubeconfigContent),
	}

	marshalledConfig, err := json.Marshal(config)
	if err != nil {
		return nil, fmt.Errorf("%w", err)
	}

	target := &landscaper.Target{
		ObjectMeta: metav1.ObjectMeta{
			Name: "test-target",
		},
		Spec: landscaper.TargetSpec{
			Type:          landscaper.KubernetesClusterTargetType,
			Configuration: marshalledConfig,
		},
	}

	return target, nil
}

func startOCIRegistryPortForward(k8sClient client.Client, namespace, kubeconfigPath string) (*exec.Cmd, error) {
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
		return nil, fmt.Errorf("expected 1 OCI registry Pod, found %d", len(ociRegistryPods.Items))
	}

	portforwardCmd, err := util.ExecCommandNonBlocking("kubectl port-forward " + ociRegistryPods.Items[0].Name + " 5000:5000 --kubeconfig " + kubeconfigPath + " --namespace " + namespace)
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
			fmt.Printf("Cannot remove echo-server-1.1.0.tgz: %s", err.Error())
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

func runQuickstartUninstall(config *config.Config) error {
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

func runQuickstartInstall(config *config.Config) error {
	const landscaperValues = `
      landscaper:
        registryConfig: # contains optional oci secrets
          allowPlainHttpRegistries: true
          secrets: {}
        deployers:
        - container
        - helm
    `

	tmpFile, err := ioutil.TempFile(".", "landscaper-values-")
	defer func() {
		err = os.Remove(tmpFile.Name())
		if err != nil {
			fmt.Printf("Cannot remove temporary file %s: %s", tmpFile.Name(), err.Error())
		}
	}()

	err = ioutil.WriteFile(tmpFile.Name(), []byte(landscaperValues), 0644)
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
