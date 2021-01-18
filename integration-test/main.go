package main

import (
	"context"
	"flag"
	"fmt"
	"io/ioutil"
	"net/http"
	"net/url"
	"os"
	"os/signal"
	"strings"
	"sync"
	"syscall"
	"time"

	"github.com/gardener/landscapercli/cmd/quickstart"
	"github.com/gardener/landscapercli/pkg/logger"
	corev1 "k8s.io/api/core/v1"
	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/cli-runtime/pkg/genericclioptions"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/clientcmd"
	"k8s.io/client-go/tools/portforward"
	"k8s.io/client-go/transport/spdy"
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

	// Start Port-forward, source https://gianarb.it/blog/programmatically-kube-port-forward-in-go
	sigs := make(chan os.Signal, 1)
	signal.Notify(sigs, syscall.SIGINT, syscall.SIGTERM)

	stopPortForwarding, wgPortForwarding := startPortForwarding(cfg, &ociRegistryPods.Items[0])

	// Perform Tests
	time.Sleep(1 * time.Minute)

	// Run "landscaper-cli quickstart uninstall"

	stopPortForwarding()
	wgPortForwarding.Wait()

	// Check if everything was cleaned up

}

func startPortForwarding(cfg *rest.Config, pod *corev1.Pod) (func(), *sync.WaitGroup) {
	var wgPortForwarding sync.WaitGroup
	wgPortForwarding.Add(1)

	stopCh := make(chan struct{}, 1)
	readyCh := make(chan struct{})

	stream := genericclioptions.IOStreams{
		In:     os.Stdin,
		Out:    os.Stdout,
		ErrOut: os.Stderr,
	}

	sigs := make(chan os.Signal, 1)
	signal.Notify(sigs, syscall.SIGINT, syscall.SIGTERM)
	go func() {
		<-sigs
		fmt.Println("Stopping port forwarding...")
		close(stopCh)
		wgPortForwarding.Done()
	}()

	go func() {
		// PortForward the pod specified from its port 9090 to the local port
		// 8080
		err := forwardAPod(PortForwardAPodRequest{
			RestConfig: cfg,
			Pod: v1.Pod{
				ObjectMeta: metav1.ObjectMeta{
					Name:      pod.Name,
					Namespace: pod.Namespace,
				},
			},
			LocalPort: 5000,
			PodPort:   5000,
			Streams:   stream,
			StopCh:    stopCh,
			ReadyCh:   readyCh,
		})
		if err != nil {
			panic(err)
		}
	}()

	<-readyCh
	println("Port forwarding is ready to get traffic.")

	stopPortForwarding := func() {
		fmt.Println("Stopping port forwarding...")
		close(stopCh)
		wgPortForwarding.Done()
	}

	return stopPortForwarding, &wgPortForwarding

}

func forwardAPod(req PortForwardAPodRequest) error {
	path := fmt.Sprintf("/api/v1/namespaces/%s/pods/%s/portforward",
		req.Pod.Namespace, req.Pod.Name)
	hostIP := strings.TrimLeft(req.RestConfig.Host, "htps:/")

	transport, upgrader, err := spdy.RoundTripperFor(req.RestConfig)
	if err != nil {
		return err
	}

	dialer := spdy.NewDialer(upgrader, &http.Client{Transport: transport}, http.MethodPost, &url.URL{Scheme: "https", Path: path, Host: hostIP})
	fw, err := portforward.New(dialer, []string{fmt.Sprintf("%d:%d", req.LocalPort, req.PodPort)}, req.StopCh, req.ReadyCh, req.Streams.Out, req.Streams.ErrOut)
	if err != nil {
		return err
	}
	return fw.ForwardPorts()
}

type PortForwardAPodRequest struct {
	// RestConfig is the kubernetes config
	RestConfig *rest.Config
	// Pod is the selected pod for this port forwarding
	Pod v1.Pod
	// LocalPort is the local port that will be selected to expose the PodPort
	LocalPort int
	// PodPort is the target port for the pod
	PodPort int
	// Steams configures where to write or read input from
	Streams genericclioptions.IOStreams
	// StopCh is the channel used to manage the port forward lifecycle
	StopCh <-chan struct{}
	// ReadyCh communicates when the tunnel is ready to receive traffic
	ReadyCh chan struct{}
}
