package main

import (
	"context"
	"flag"
	"fmt"
	"io/ioutil"
	"os"
	"time"

	"github.com/gardener/landscapercli/cmd/quickstart"
	"github.com/gardener/landscapercli/pkg/logger"
	"github.com/gardener/landscapercli/pkg/util"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/cli-runtime/pkg/genericclioptions"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/tools/clientcmd"
	"k8s.io/kubectl/pkg/cmd/portforward"
	cmdutil "k8s.io/kubectl/pkg/cmd/util"
)

const landscaperValues = `
landscaper:
  registryConfig: # contains optional oci secrets
    allowPlainHttpRegistries: false
    secrets: {}
#     <name>: <docker config json>
  deployers:
  - container
  - helm
#  - mock
`

const (
	defaultNamespace = "integration-test"
)

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
	args := []string{
		"--kubeconfig",
		kubeconfig,
		"--landscaper-values",
		tmpFile.Name(),
		"--install-oci-registry",
		"--namespace",
		namespace,
	}
	cmd := quickstart.NewInstallCommand(ctx)
	cmd.SetArgs(args)

	err = cmd.Execute()
	if err != nil {
		fmt.Println("Install Command failed: %w", err)
		os.Exit(1)
	}

	cfg, err := clientcmd.BuildConfigFromFlags("", kubeconfig)
	if err != nil {
		fmt.Println("Cannot parse K8s config: %w", err)
		os.Exit(1)
	}
	
	k8sClient, err := kubernetes.NewForConfig(cfg)
	if err != nil {
		fmt.Println("Cannot build K8s clientset: %w", err)
		os.Exit(1)
	}
	
	// Wait until Pods are up and running
	for {
		time.Sleep(10 * time.Second)
		podList, err := k8sClient.CoreV1().Pods(namespace).List(ctx, metav1.ListOptions{})
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
	}

	ociRegistryPods, err := k8sClient.CoreV1().Pods(namespace).List(ctx, metav1.ListOptions{
		LabelSelector: "app=oci-registry",
	})
	if err != nil {
		fmt.Println("Cannot list pods: %w", err)
		os.Exit(1)
	}

	if len(ociRegistryPods.Items) != 1 {
		fmt.Println("len(ociRegistryPods.Items) != 1")
		os.Exit(1)
	}


	// Start Port-forward
	restClientGetter := util.NewRESTClientGetter(namespace, kubeconfig)

	factory := cmdutil.NewFactory(restClientGetter)
	iostreams := genericclioptions.IOStreams{
		In: os.Stdin,
		Out: os.Stdout,
		ErrOut: os.Stderr,
	}

	portforwardArgs := []string{
		ociRegistryPods.Items[0].Name,
		"5000:5000",
	}
	portforwardCmd := portforward.NewCmdPortForward(factory, iostreams)
	portforwardCmd.SetArgs(portforwardArgs)

	err = portforwardCmd.Execute()
	if err != nil {
		fmt.Println("Portforward cmd failed: %w", err)
		os.Exit(1)
	}

	time.Sleep(10 * time.Minute)

	// Perform Tests


	// Run "landscaper-cli quickstart uninstall"


	// Check if everything was cleaned up

}