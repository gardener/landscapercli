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
	"github.com/gardener/landscapercli/pkg/logger"
	"github.com/gardener/landscapercli/pkg/util"
	corev1 "k8s.io/api/core/v1"
	k8sErrors "k8s.io/apimachinery/pkg/api/errors"
	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/labels"
	"k8s.io/apimachinery/pkg/runtime"
	clientgoscheme "k8s.io/client-go/kubernetes/scheme"
	"k8s.io/client-go/tools/clientcmd"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

const (
	defaultNamespace = "landscaper"
	targetName = "test-target"
)

var (
	scheme = runtime.NewScheme()
)

func init() {
	_ = clientgoscheme.AddToScheme(scheme)
	_ = landscaper.AddToScheme(scheme)
	// +kubebuilder:scaffold:scheme
}

func CreateTarget(k8sClient client.Client, name, namespace, kubeconfig string) error {
	kubeconfigContent, err := ioutil.ReadFile(kubeconfig)
	if err != nil {
		return fmt.Errorf("Cannot read kubeconfig: %w", err)
	}

	test1 := map[string]interface{}{
		"kubeconfig": string(kubeconfigContent),
	}

	marsh1, err := json.Marshal(test1)
	if err != nil {
		panic(err)
	}

	target := &landscaper.Target{
		ObjectMeta: v1.ObjectMeta{
			Name:      "test-target",
			Namespace: namespace,
		},
		Spec: landscaper.TargetSpec{
			Type:          landscaper.KubernetesClusterTargetType,
			Configuration: marsh1,
		},
	}

	err = k8sClient.Create(context.TODO(), target, &client.CreateOptions{})
	if err != nil {
		return fmt.Errorf("cannot create target: %w", err)
	}

	return nil
}

func StartOCIRegistryPortForward(k8sClient client.Client, namespace, kubeconfigPath string) (*exec.Cmd, error) {
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
		return nil, fmt.Errorf("Cannot list pods: %w", err)
	}

	if len(ociRegistryPods.Items) != 1 {
		return nil, fmt.Errorf("len(ociRegistryPods.Items) != 1")
	}

	portforwardCmd, err := util.ExecCommandNonBlocking("kubectl port-forward " + ociRegistryPods.Items[0].Name + " 5000:5000 --kubeconfig " + kubeconfigPath + " --namespace " + namespace)
	if err != nil {
		return nil, fmt.Errorf("kubectl port-forward failed: %w", err)
	}

	return portforwardCmd, nil
}

func UploadEchoServerHelmChart() error {
	err := util.ExecCommandBlocking("helm pull https://storage.googleapis.com/sap-hub-test/echo-server-1.1.0.tgz")
	if err != nil {
		return fmt.Errorf("helm pull failed: %w", err)
	}
	defer func() {
		err = os.Remove("echo-server-1.1.0.tgz")
		if err != nil {
			fmt.Printf("cannot remove echo-server-1.1.0.tgz: %s", err.Error())
		}
	}()

	err = util.ExecCommandBlocking("helm chart save echo-server-1.1.0.tgz localhost:5000/echo-server-chart:v1.1.0")
	if err != nil {
		return fmt.Errorf("helm chart save failed: %w", err)
	}

	err = util.ExecCommandBlocking("helm chart push localhost:5000/echo-server-chart:v1.1.0")
	if err != nil {
		return fmt.Errorf("helm chart push failed: %w", err)
	}

	return nil
}

func RunQuickstartUninstall(kubeconfigPath, installNamespace string) error {
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
		return fmt.Errorf("Uninstall Command failed: %w", err)
	}

	return nil
}

func RunQuickstartInstall(kubeconfigPath, installNamespace string) error {
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
			fmt.Printf("cannot remove temporary file %s: %s", tmpFile.Name(), err.Error())
		}
	}()

	err = ioutil.WriteFile(tmpFile.Name(), []byte(landscaperValues), 0644)
	if err != nil {
		return fmt.Errorf("Cannot write to file: %w", err)
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
		return fmt.Errorf("Install Command failed: %w", err)
	}

	return nil
}

func main() {
	fmt.Println("Hallo Integration-Test")

	var kubeconfig, namespace string
	flag.StringVar(&kubeconfig, "kubeconfig", "", "kubeconfig of the taget K8s cluster")
	flag.StringVar(&namespace, "namespace", defaultNamespace, "namespace on the target cluster")
	flag.Parse()

	log, err := logger.NewCliLogger()
	logger.SetLogger(log)
	ctx := context.TODO()

	cfg, err := clientcmd.BuildConfigFromFlags("", kubeconfig)
	if err != nil {
		fmt.Println("Cannot parse K8s config: %w", err)
		os.Exit(1)
	}

	k8sClient, err := client.New(cfg, client.Options{
		Scheme: scheme,
	})
	if err != nil {
		fmt.Println("Cannot build K8s client: %w", err)
		os.Exit(1)
	}

	RunQuickstartInstall(kubeconfig, namespace)
	if err != nil {
		fmt.Println("quickstart install: %w", err)
		os.Exit(1)
	}
	defer func() {
		err = RunQuickstartUninstall(kubeconfig, namespace)
		if err != nil {
			fmt.Println("quickstart uninstall failed: %w", err)
			os.Exit(1)
		}
	}()

	const (
		sleepTime = 5*time.Second
		maxRetries = 4
	)
	err = util.WaitUntilAllPodsAreReady(k8sClient, namespace, sleepTime, maxRetries)
	if err != nil {
		fmt.Println("waiting for pods failed: %w", err)
		os.Exit(1)
	}

	portforwardCmd, err := StartOCIRegistryPortForward(k8sClient, namespace, kubeconfig)
	if err != nil {
		fmt.Println("portforward to oci registry failed: %w", err)
		os.Exit(1)
	}
	defer func() {
		// Disable port-forward
		err = portforwardCmd.Process.Kill()
		if err != nil {
			fmt.Println("Cannot kill kubectl port-forward process: %w", err)
			os.Exit(1)
		}
	}()

	err = UploadEchoServerHelmChart()
	if err != nil {
		fmt.Println("upload of echo server helm chart failed: %w", err)
		os.Exit(1)
	}

	err = CreateTarget(k8sClient, targetName, namespace, kubeconfig)
	if err != nil {
		fmt.Println("Cannot create target: %w", err)
		os.Exit(1)
	}
	defer func() {
		// delete target
		target := &landscaper.Target{
			ObjectMeta: v1.ObjectMeta{
				Name:      targetName,
				Namespace: namespace,
			},
		}
		err = k8sClient.Delete(ctx, target, &client.DeleteOptions{})
		if err != nil {
			fmt.Println("cannot delete target: %w", err)
			os.Exit(1)
		}
	}()

	// Perform Tests
	test := NewTest(k8sClient, "#"+targetName)
	err = test.Run()
	if err != nil {
		fmt.Println("landscaper test failed: %w", err)
		os.Exit(1)
	}

	// Check if everything was cleaned up. Idea: write a defer() function at the beginning of main()

	fmt.Println("Adieu Integration-Test")
}

// ########################################### Landscaper Test #######################################
type Test struct {
	client     client.Client
	targetName string
}

func NewTest(client client.Client, targetName string) *Test {
	obj := &Test{
		client:     client,
		targetName: targetName,
	}
	return obj
}

func (t *Test) Run() error {
	ctx := context.TODO()

	instName := "test-1"
	instNs := "landscaper"

	inst := buildInstallation(instName, instNs, t.targetName, "landscaper-example")

	err := t.client.Create(ctx, inst, &client.CreateOptions{})
	if err != nil {
		return err
	}

	for {
		err = t.client.Get(ctx, client.ObjectKey{Name: instName, Namespace: instNs}, inst)
		if err != nil {
			return err
		}

		if inst.Status.Phase == landscaper.ComponentPhaseSucceeded {
			fmt.Println("installation.phase == succeeded")
			break
		}

		fmt.Println("installation.phase != succeeded => go to sleep...")
		time.Sleep(5 * time.Second)
	}

	err = t.client.Delete(ctx, inst, &client.DeleteOptions{})
	if err != nil {
		return err
	}

	for {
		err = t.client.Get(ctx, client.ObjectKey{Name: instName, Namespace: instNs}, inst)
		if err != nil {
			if k8sErrors.IsNotFound(err) {
				fmt.Println("installation was deleted successfully")
				break
			}
			return err
		}

		fmt.Println("waiting for installation to be deleted => go to sleep...")
		time.Sleep(5 * time.Second)
	}

	return nil
}

func buildInstallation(name, namespace, targetName, appNamespace string) *landscaper.Installation {
	inlineBlueprint := `
        apiVersion: landscaper.gardener.cloud/v1alpha1
        kind: Blueprint

        imports:
        - name: cluster
          targetType: landscaper.gardener.cloud/kubernetes-cluster

        - name: appname
          schema:
            type: string

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
                  ref: "oci-registry.landscaper.svc.cluster.local:5000/echo-server-chart:v1.1.0"
        
                updateStrategy: patch

                name: {{ .imports.appname }}
                namespace: {{ .imports.appnamespace }}
    `

	fs := map[string]interface{}{
		"blueprint.yaml": inlineBlueprint,
	}

	marsh, err := json.Marshal(fs)
	if err != nil {
		panic(err)
	}

	obj := &landscaper.Installation{
		ObjectMeta: v1.ObjectMeta{
			Name:      name,
			Namespace: namespace,
		},
		Spec: landscaper.InstallationSpec{
			Blueprint: landscaper.BlueprintDefinition{
				Inline: &landscaper.InlineBlueprint{
					Filesystem: json.RawMessage(marsh),
				},
			},
			Imports: landscaper.InstallationImports{
				Targets: []landscaper.TargetImportExport{
					{
						Name:   "cluster",
						Target: targetName,
					},
				},
			},
			ImportDataMappings: map[string]json.RawMessage{
				"appname":      json.RawMessage(`"echo-server"`),
				"appnamespace": json.RawMessage(fmt.Sprintf(`"%s"`, appNamespace)),
			},
		},
	}

	return obj
}
