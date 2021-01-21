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
	"github.com/gardener/landscapercli/cmd/quickstart"
	"github.com/gardener/landscapercli/integration-test/tests"
	"github.com/gardener/landscapercli/pkg/logger"
	"github.com/gardener/landscapercli/pkg/util"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/labels"
	"k8s.io/apimachinery/pkg/runtime"
	clientgoscheme "k8s.io/client-go/kubernetes/scheme"
	"k8s.io/client-go/tools/clientcmd"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

const (
	defaultLandscaperNamespace = "landscaper"
)

var (
	scheme = runtime.NewScheme()
)

func init() {
	_ = clientgoscheme.AddToScheme(scheme)
	_ = landscaper.AddToScheme(scheme)
}

func main() {
	fmt.Println("========== Starting integration-test ==========")
	err := run()
	if err != nil {
		fmt.Println("Error while running integration-test:", err)
		os.Exit(1)
	}
	fmt.Println("========== Integration-test finished successfully ==========")
}

func run() error {
	var kubeconfig, landscaperNamespace string
	flag.StringVar(&kubeconfig, "kubeconfig", "", "path to the kubeconfig of the target cluster")
	flag.StringVar(&landscaperNamespace, "landscaper-namespace", defaultLandscaperNamespace, "namespace on the target cluster to setup Landscaper")
	flag.Parse()

	log, err := logger.NewCliLogger()
	logger.SetLogger(log)

	cfg, err := clientcmd.BuildConfigFromFlags("", kubeconfig)
	if err != nil {
		return fmt.Errorf("cannot parse K8s config: %w", err)
	}

	k8sClient, err := client.New(cfg, client.Options{
		Scheme: scheme,
	})
	if err != nil {
		return fmt.Errorf("cannot build K8s client: %w", err)
	}

	defer func() {
		// Check if everything was cleaned up
		fmt.Println("Running cleanup check routine")
	}()

	fmt.Println("Running landscaper-cli quickstart install")
	runQuickstartInstall(kubeconfig, landscaperNamespace)
	if err != nil {
		return fmt.Errorf("landscaper-cli quickstart install failed: %w", err)
	}
	defer func() {
		err = runQuickstartUninstall(kubeconfig, landscaperNamespace)
		if err != nil {
			fmt.Println("landscaper-cli quickstart uninstall failed: %w", err)
		}
	}()

	const (
		sleepTime  = 5 * time.Second
		maxRetries = 6
	)

	fmt.Println("Waiting for Landscaper Pods to get ready")
	err = util.WaitUntilAllPodsAreReady(k8sClient, landscaperNamespace, sleepTime, maxRetries)
	if err != nil {
		return fmt.Errorf("error while waiting for Landscaper Pods: %w", err)
	}

	fmt.Println("Starting port-forward to OCI registry")
	portforwardCmd, err := startOCIRegistryPortForward(k8sClient, landscaperNamespace, kubeconfig)
	if err != nil {
		return fmt.Errorf("port-forward to OCI registry failed: %w", err)
	}
	defer func() {
		// Disable port-forward
		err = portforwardCmd.Process.Kill()
		if err != nil {
			fmt.Println("Cannot kill port-forward process: %w", err)
		}
	}()

	fmt.Println("Upload echo-server Helm Chart to OCI registry")
	helmChartRef, err := uploadEchoServerHelmChart(landscaperNamespace)
	if err != nil {
		return fmt.Errorf("upload of echo-server Helm Chart failed: %w", err)
	}

	fmt.Println("Build target")
	target, err := buildTarget(kubeconfig)
	if err != nil {
		return fmt.Errorf("cannot build target: %w", err)
	}

	fmt.Println("Running test suite")
	// ################################ <Run Tests> ################################

	err = tests.RunQuickstartInstallTest(k8sClient, target, helmChartRef)
	if err != nil {
		return fmt.Errorf("RunQuickstartInstallTest() failed: %w", err)
	}

	// ############################### </Run Tests> ################################
	fmt.Println("Test suite finished successfully")

	return nil
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

func runQuickstartUninstall(kubeconfigPath, installNamespace string) error {
	uninstallArgs := []string{
		"--kubeconfig",
		kubeconfigPath,
		"--namespace",
		installNamespace,
	}
	uninstallCmd := quickstart.NewUninstallCommand(context.TODO())
	uninstallCmd.SetArgs(uninstallArgs)

	err := uninstallCmd.Execute()
	if err != nil {
		return fmt.Errorf("uninstall command failed: %w", err)
	}

	return nil
}

func runQuickstartInstall(kubeconfigPath, installNamespace string) error {
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
		kubeconfigPath,
		"--landscaper-values",
		tmpFile.Name(),
		"--install-oci-registry",
		"--namespace",
		installNamespace,
	}
	installCmd := quickstart.NewInstallCommand(context.TODO())
	installCmd.SetArgs(installArgs)

	err = installCmd.Execute()
	if err != nil {
		return fmt.Errorf("install command failed: %w", err)
	}

	return nil
}
