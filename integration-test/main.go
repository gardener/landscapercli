package main

import (
	"context"
	"encoding/json"
	"flag"
	"fmt"
	"io/ioutil"
	"os"
	"time"

	landscaper "github.com/gardener/landscaper/pkg/apis/core/v1alpha1"
	"github.com/gardener/landscapercli/cmd/quickstart"
	"github.com/gardener/landscapercli/pkg/logger"
	"github.com/gardener/landscapercli/pkg/util"
	corev1 "k8s.io/api/core/v1"
	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/labels"
	"k8s.io/apimachinery/pkg/runtime"
	clientgoscheme "k8s.io/client-go/kubernetes/scheme"
	"k8s.io/client-go/tools/clientcmd"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

const landscaperValues = `
landscaper:
  registryConfig: # contains optional oci secrets
    allowPlainHttpRegistries: true
    secrets: {}
#     <name>: <docker config json>
  deployers:
  - container
  - helm
#  - mock
`

const (
	defaultNamespace = "landscaper"
)

var (
	scheme   = runtime.NewScheme()
)

func init() {
	_ = clientgoscheme.AddToScheme(scheme)
	_ = landscaper.AddToScheme(scheme)
	// +kubebuilder:scaffold:scheme
}

func main() {
	fmt.Println("Hallo Integration-Test")

	var kubeconfig, namespace string
	flag.StringVar(&kubeconfig, "kubeconfig", "", "kubeconfig of the taget K8s cluster")
	flag.StringVar(&namespace, "namespace", defaultNamespace, "namespace on the target cluster")
	flag.Parse()

	// Run "landscaper-cli quickstart install"
	tmpFile, err := ioutil.TempFile(".", "landscaper-values-")
	defer func() {
		err = os.Remove(tmpFile.Name())
		if err != nil {
			fmt.Printf("cannot remove temporary file %s: %s", tmpFile.Name(), err.Error())
		}
	}()

	err = ioutil.WriteFile(tmpFile.Name(), []byte(landscaperValues), 0644)
	if err != nil {
		fmt.Println("Cannot write to file: %w", err)
		os.Exit(1)
	}

	log, err := logger.NewCliLogger()
	logger.SetLogger(log)
	ctx := context.TODO()

	installArgs := []string{
		"--kubeconfig",
		kubeconfig,
		"--landscaper-values",
		tmpFile.Name(),
		"--install-oci-registry",
		"--namespace",
		namespace,
	}
	installCmd := quickstart.NewInstallCommand(ctx)
	installCmd.SetArgs(installArgs)

	err = installCmd.Execute()
	if err != nil {
		fmt.Println("Install Command failed: %w", err)
		os.Exit(1)
	}

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

	// Wait until Pods are up and running
	for {
		
		podList := corev1.PodList{}
		err = k8sClient.List(ctx, &podList, client.InNamespace(namespace))
		if err != nil {
			fmt.Println("Cannot list pods: %w", err)
			os.Exit(1)
		}
		
		numberOfRunningPods := 0
		for _, pod := range podList.Items {
			for _, condition := range pod.Status.Conditions {
				if condition.Type == corev1.PodReady {
					if condition.Status == corev1.ConditionTrue {
						numberOfRunningPods++
					}
				}
			}
		}
		
		if numberOfRunningPods == len(podList.Items) {
			break
		}

		time.Sleep(5 * time.Second)
	}

	ociRegistryPods := corev1.PodList{}
	err = k8sClient.List(
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
		fmt.Println("Cannot list pods: %w", err)
		os.Exit(1)
	}

	if len(ociRegistryPods.Items) != 1 {
		fmt.Println("len(ociRegistryPods.Items) != 1")
		os.Exit(1)
	}

	// Start Port-forward
	portforwardCmd, err := util.ExecCommandNonBlocking("kubectl port-forward "+ociRegistryPods.Items[0].Name+" 5000:5000 --kubeconfig "+kubeconfig+" --namespace "+namespace)
	if err != nil {
		fmt.Println("kubectl port-forward failed: %w", err)
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

	// Prepare Helm Chart
	err = util.ExecCommandBlocking("helm pull https://storage.googleapis.com/sap-hub-test/echo-server-1.1.0.tgz")
	if err != nil {
		fmt.Println("helm pull failed: %w", err)
		os.Exit(1)
	}
	defer func() {
		err = os.Remove("echo-server-1.1.0.tgz")
		if err != nil {
			fmt.Printf("cannot remove echo-server-1.1.0.tgz: %s", err.Error())
		}
	}()

	err = util.ExecCommandBlocking("helm chart save echo-server-1.1.0.tgz localhost:5000/echo-server-chart:v1.1.0")
	if err != nil {
		fmt.Println("helm chart save failed: %w", err)
		os.Exit(1)
	}

	err = util.ExecCommandBlocking("helm chart push localhost:5000/echo-server-chart:v1.1.0")
	if err != nil {
		fmt.Println("helm chart push failed: %w", err)
		os.Exit(1)
	}

	// Create target
	kubeconfigContent, err := ioutil.ReadFile(kubeconfig)
	if err != nil {
		fmt.Println("Cannot read kubeconfig: %w", err)
		os.Exit(1)
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
			Name: "test-target",
			Namespace: namespace,
		},
		Spec: landscaper.TargetSpec{
			Type: landscaper.KubernetesClusterTargetType,
			Configuration: marsh1,
		},
	}

	err = k8sClient.Create(ctx, target, &client.CreateOptions{})
	if err != nil {
		fmt.Println("cannot create target: %w", err)
		os.Exit(1)
	}

	// Perform Tests
	test := NewTest(k8sClient)
	err = test.Run()
	if err != nil {
		fmt.Println("landscaper test failed: %w", err)
		os.Exit(1)
	}

	// delete target
	err = k8sClient.Delete(ctx, target, &client.DeleteOptions{})
	if err != nil {
		fmt.Println("cannot delete target: %w", err)
		os.Exit(1)
	}

	// Run "landscaper-cli quickstart uninstall"
	uninstallArgs := []string{
		"--kubeconfig",
		kubeconfig,
		"--namespace",
		namespace,
	}
	uninstallCmd := quickstart.NewUninstallCommand(ctx)
	uninstallCmd.SetArgs(uninstallArgs)
	
	err = uninstallCmd.Execute()
	if err != nil {
		fmt.Println("Uninstall Command failed: %w", err)
		os.Exit(1)
	}

	// Check if everything was cleaned up


	fmt.Println("Adieu Integration-Test")
}


// ########################################### Landscaper Test #######################################
type Test struct {
	client client.Client
}

func NewTest(client client.Client) *Test {
	obj := &Test{
		client: client,
	}
	return obj
}

func (t *Test) Run() error {
	ctx := context.TODO()

	inst := buildInstallation("test-1", "landscaper")

	err := t.client.Create(ctx, inst, &client.CreateOptions{})
	if err != nil {
		return err
	}

	time.Sleep(15*time.Second)

	err = t.client.Delete(ctx, inst, &client.DeleteOptions{})
	if err != nil {
		return err
	}

	time.Sleep(10*time.Second)

	return nil
}

func buildInstallation(name, namespace string,) *landscaper.Installation {
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
			Name: name,
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
						Name: "cluster",
						Target: "#test-target",
					},
				},
			},
			ImportDataMappings: map[string]json.RawMessage{
				"appname": json.RawMessage(`"echo-server"`),
				"appnamespace": json.RawMessage(`"landscaper-example"`),
			},
		},
	}

	return obj
}
