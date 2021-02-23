package util

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"os"
	"os/exec"
	"strings"
	"time"

	lsv1alpha1 "github.com/gardener/landscaper/apis/core/v1alpha1"
	yamlv3 "gopkg.in/yaml.v3"
	corev1 "k8s.io/api/core/v1"
	k8sErrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

// GetValueFromNestedMap extracts the value in a given value path (e.g. landscaper.registryConfig.allowPlainHttpRegistries) or returns an error
// if the path does not exists.
func GetValueFromNestedMap(data map[string]interface{}, valuePath string) (interface{}, error) {
	var val interface{}
	var ok bool

	keys := strings.Split(valuePath, ".")
	for index, key := range keys {
		if index == len(keys)-1 {
			val, ok = data[key]
			if !ok {
				return nil, fmt.Errorf("Cannot get value for path %s", valuePath)
			}
		} else {
			tmp := data[key]
			data, ok = tmp.(map[string]interface{})
			if !ok {
				return nil, fmt.Errorf("Cannot get value for path %s", valuePath)
			}
		}
	}

	return val, nil
}

// ExecCommandBlocking executes a command and wait for its completion.  Returns a Cmd that can be used to stop the command
func ExecCommandBlocking(command string) error {
	fmt.Printf("Executing: %s\n", command)

	arr := strings.Split(command, " ")

	if arr[0] == "helm" {
		helmPath := os.Getenv("HELM_EXECUTABLE")
		if helmPath != "" {
			arr[0] = helmPath
			fmt.Printf("Using helm binary: %s\n", arr[0])
		}
	}

	cmd := exec.Command(arr[0], arr[1:]...)
	cmd.Env = []string{"HELM_EXPERIMENTAL_OCI=1", "HOME=" + os.Getenv("HOME"), "PATH=" + os.Getenv("PATH")}
	out, err := cmd.CombinedOutput()
	outStr := string(out)

	if err != nil {
		return fmt.Errorf("failed with error: %s:\n%s\n", err, outStr)
	}
	fmt.Println("Executed sucessfully!")

	return nil
}

type CmdResult struct {
	Error  error
	Stdout string
	StdErr string
}

// ExecCommandNonBlocking executes a command without without blocking. Returns a Cmd that can be used to stop the command.
// When the command has stopped or failed, the result is written into the channel resultCh.
func ExecCommandNonBlocking(command string, resultCh chan<- CmdResult) (*exec.Cmd, error) {
	fmt.Printf("Executing: %s\n", command)

	arr := strings.Split(command, " ")

	if arr[0] == "helm" {
		helmPath := os.Getenv("HELM_EXECUTABLE")
		if helmPath != "" {
			arr[0] = helmPath
			fmt.Printf("Using helm binary: %s\n", arr[0])
		}
	}

	outbuf := bytes.Buffer{}
	errbuf := bytes.Buffer{}

	cmd := exec.Command(arr[0], arr[1:]...)
	cmd.Env = []string{"HELM_EXPERIMENTAL_OCI=1", "HOME=" + os.Getenv("HOME"), "PATH=" + os.Getenv("PATH")}
	cmd.Stderr = &outbuf
	cmd.Stdout = &errbuf

	err := cmd.Start()
	if err != nil {
		fmt.Printf("Failed with error: %s:\n", err)
		return nil, err
	}
	fmt.Println("Started sucessfully!")

	go func() {
		exitErr := cmd.Wait()
		res := CmdResult{
			Error:  exitErr,
			Stdout: outbuf.String(),
			StdErr: outbuf.String(),
		}
		resultCh <- res
		close(resultCh)
	}()

	return cmd, nil
}

// CheckConditionPeriodically checks the success of a function peridically. Returns timeout(bool) to indicate the success of the function
// and propagates possible errors of the function.
func CheckConditionPeriodically(conditionFunc func() (bool, error), sleepTime time.Duration, maxRetries int) (bool, error) {
	retries := 0
	for {
		fmt.Printf("Checking condition... retries: %d\n", retries)

		ok, err := conditionFunc()
		if err != nil {
			return false, err
		}

		if ok {
			break
		}

		if retries >= maxRetries {
			return true, nil
		}
		retries++

		time.Sleep(sleepTime)
	}
	return false, nil
}

// CheckAndWaitUntilAllPodsAreReady checks and wait until all pods have the ready condition set to true. Returns an error on failure
// or if timeout (sleepTime and maxRetries is reached). Returns a boolean indicating if all pods have the ready condition true on return.
func CheckAndWaitUntilAllPodsAreReady(k8sClient client.Client, namespace string, sleepTime time.Duration, maxRetries int) (bool, error) {
	conditionFunc := func() (bool, error) {
		podList := corev1.PodList{}
		err := k8sClient.List(context.TODO(), &podList, client.InNamespace(namespace))
		if err != nil {
			return false, fmt.Errorf("Cannot list pods: %w", err)
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

		return numberOfRunningPods == len(podList.Items), nil
	}

	return CheckConditionPeriodically(conditionFunc, sleepTime, maxRetries)
}

// CheckAndWaitUntilLandscaperInstallationSucceeded checks and wait until the installation is on status succeeded. Returns an error on failure
// or if timeout (sleepTime and maxRetries is reached). Returns a boolean indicating if the installation succeeded.
func CheckAndWaitUntilLandscaperInstallationSucceeded(k8sClient client.Client, key types.NamespacedName, sleepTime time.Duration, maxRetries int) (bool, error) {
	conditionFunc := func() (bool, error) {
		inst := &lsv1alpha1.Installation{}
		err := k8sClient.Get(context.TODO(), key, inst)
		if err != nil {
			return false, fmt.Errorf("cannot get installation: %w", err)
		}

		return inst.Status.Phase == lsv1alpha1.ComponentPhaseSucceeded, nil
	}

	return CheckConditionPeriodically(conditionFunc, sleepTime, maxRetries)
}

// CheckAndWaitUntilObjectNotExistAnymore periodically checks and wait until the object does not exist anymore. Returns an error on failure
// or if timeout (sleepTime and maxRetries is reached). Returns a boolean indicating if the object remains on return.
func CheckAndWaitUntilObjectNotExistAnymore(k8sClient client.Client, objKey types.NamespacedName, obj client.Object, sleepTime time.Duration, maxRetries int) (bool, error) {
	conditionFunc := func() (bool, error) {
		err := k8sClient.Get(context.TODO(), objKey, obj)
		if err != nil {
			if k8sErrors.IsNotFound(err) {
				return true, nil
			}
			return false, err
		}
		return false, nil
	}

	return CheckConditionPeriodically(conditionFunc, sleepTime, maxRetries)
}

// CheckAndWaitUntilNoInstallationsInNamespaceExists periodically checks and wait until no installation in the namespace remains. Returns an error on failure
// or if timeout (sleepTime and maxRetries is reached). Returns a boolean indicating if no installations remains on return.
func CheckAndWaitUntilNoInstallationsInNamespaceExists(k8sClient client.Client, namespace string, sleepTime time.Duration, maxRetries int) (bool, error) {
	conditionFunc := func() (bool, error) {
		ctx := context.TODO()

		installationList := &lsv1alpha1.InstallationList{}
		err := k8sClient.List(ctx, installationList, &client.ListOptions{Namespace: namespace})
		if err != nil {
			return false, err
		}

		return len(installationList.Items) == 0, nil
	}

	return CheckConditionPeriodically(conditionFunc, sleepTime, maxRetries)
}

// DeleteNamespace deletes a namespace (if it exists). First, a graceful delete will be tried with a timeout.
// On timeout, the namespace will be deleted forcefully by removing the finalizers on the Landscaper CRs.
func DeleteNamespace(k8sClient client.Client, namespace string, sleepTime time.Duration, maxRetries int) error {
	ns := &corev1.Namespace{
		ObjectMeta: metav1.ObjectMeta{
			Name: namespace,
		},
	}
	err := k8sClient.Get(context.TODO(), client.ObjectKey{Name: namespace}, ns)
	if err != nil {
		if k8sErrors.IsNotFound(err) {
			return nil
		}
		return fmt.Errorf("error getting namespace %s: %w", namespace, err)
	}

	timeout, err := gracefullyDeleteNamespace(k8sClient, namespace, sleepTime, maxRetries)
	if err != nil {
		return fmt.Errorf("deleting namespace gracefully failed: %w", err)
	}
	if timeout {
		fmt.Println("Deleting namespace gracefully timed out, using force delete...")
		timeout, err = forceDeleteNamespace(k8sClient, namespace, sleepTime, maxRetries)
		if err != nil {
			return fmt.Errorf("deleting namespace forcefully failed: %w", err)
		}
		if timeout {
			return fmt.Errorf("deleting namespace forcefully timed out")
		}
	}
	return nil
}

// gracefully try to delete all Landscaper installations in a namespace, then delete the namespace itself
func gracefullyDeleteNamespace(k8sClient client.Client, namespace string, sleepTime time.Duration, maxRetries int) (bool, error) {
	ctx := context.TODO()

	installationList := lsv1alpha1.InstallationList{}
	err := k8sClient.List(ctx, &installationList, &client.ListOptions{Namespace: namespace})
	if err != nil {
		return false, err
	}
	for _, installation := range installationList.Items {
		fmt.Println("Deleting installation:", installation.Name)
		err = k8sClient.Delete(ctx, &installation, &client.DeleteOptions{})
		if err != nil {
			return false, fmt.Errorf("cannot delete installation: %w", err)
		}
	}

	timeout, err := CheckAndWaitUntilNoInstallationsInNamespaceExists(k8sClient, namespace, sleepTime, maxRetries)
	if err != nil {
		return false, err
	}
	if timeout {
		return timeout, nil
	}

	ns := &corev1.Namespace{
		ObjectMeta: metav1.ObjectMeta{
			Name: namespace,
		},
	}
	fmt.Println("Deleting namespace:", namespace)
	err = k8sClient.Delete(ctx, ns, &client.DeleteOptions{})
	if err != nil {
		return false, fmt.Errorf("cannot delete namespace: %w", err)
	}

	timeout, err = CheckAndWaitUntilObjectNotExistAnymore(k8sClient, client.ObjectKey{Name: namespace}, ns, sleepTime, maxRetries)
	if err != nil {
		return false, fmt.Errorf("error while waiting for namespace to be deleted: %w", err)
	}
	if timeout {
		return true, nil
	}

	return false, nil
}

func forceDeleteNamespace(k8sClient client.Client, namespace string, sleepTime time.Duration, maxRetries int) (bool, error) {
	ctx := context.TODO()

	ns := &corev1.Namespace{
		ObjectMeta: metav1.ObjectMeta{
			Name: namespace,
		},
	}
	err := k8sClient.Delete(ctx, ns, &client.DeleteOptions{})
	if err != nil {
		return false, fmt.Errorf("cannot delete namespace: %w", err)
	}

	err = removeFinalizersFromLandscaperCRs(k8sClient, namespace)
	if err != nil {
		return false, fmt.Errorf("cannot remove finalizer: %w", err)
	}

	timeout, err := CheckAndWaitUntilObjectNotExistAnymore(k8sClient, client.ObjectKey{Name: namespace}, ns, sleepTime, maxRetries)
	if err != nil {
		return false, fmt.Errorf("error while waiting for namespace to be deleted: %w", err)
	}
	if timeout {
		return true, nil
	}

	return false, nil
}

// removes the finalizers from all Landscaper CRs in a namespace
func removeFinalizersFromLandscaperCRs(k8sClient client.Client, namespace string) error {
	ctx := context.TODO()

	installationList := lsv1alpha1.InstallationList{}
	err := k8sClient.List(ctx, &installationList, &client.ListOptions{Namespace: namespace})
	if err != nil {
		return fmt.Errorf("cannot list installations: %w", err)
	}
	for _, installation := range installationList.Items {
		err = removeFinalizers(ctx, k8sClient, &installation)
		if err != nil {
			return fmt.Errorf("cannot remove finalizers for installation: %w", err)
		}
	}

	executionList := &lsv1alpha1.ExecutionList{}
	err = k8sClient.List(ctx, executionList, &client.ListOptions{Namespace: namespace})
	if err != nil {
		return fmt.Errorf("cannot list executions: %w", err)
	}
	for _, execution := range executionList.Items {
		err = removeFinalizers(ctx, k8sClient, &execution)
		if err != nil {
			return fmt.Errorf("cannot remove finalizers for execution: %w", err)
		}
	}

	deployItemList := &lsv1alpha1.DeployItemList{}
	err = k8sClient.List(ctx, deployItemList, &client.ListOptions{Namespace: namespace})
	if err != nil {
		return fmt.Errorf("cannot list deployitems: %w", err)
	}
	for _, deployItem := range deployItemList.Items {
		err = removeFinalizers(ctx, k8sClient, &deployItem)
		if err != nil {
			return fmt.Errorf("cannot remove finalizers for deployitem: %w", err)
		}
	}

	return nil
}

func removeFinalizers(ctx context.Context, k8sClient client.Client, object metav1.Object) error {
	if len(object.GetFinalizers()) == 0 {
		return nil
	}
	object.SetFinalizers([]string{})
	return k8sClient.Update(ctx, object.(client.Object))
}

func MarshalYaml(node *yamlv3.Node) ([]byte, error) {
	buf := bytes.Buffer{}
	enc := yamlv3.NewEncoder(&buf)
	enc.SetIndent(2)
	err := enc.Encode(node)
	if err != nil {
		return nil, err
	}
	return buf.Bytes(), nil
}

func FindNodeByPath(node *yamlv3.Node, path string) (*yamlv3.Node, *yamlv3.Node) {
	if node == nil || path == "" {
		return nil, nil
	}

	var keyNode, valueNode *yamlv3.Node
	if node.Kind == yamlv3.DocumentNode {
		valueNode = node.Content[0]
	} else {
		valueNode = node
	}
	splittedPath := strings.Split(path, ".")

	for _, p := range splittedPath {
		keyNode, valueNode = findNode(valueNode.Content, p)
		if keyNode == nil && valueNode == nil {
			break
		}
	}

	return keyNode, valueNode
}

func findNode(nodes []*yamlv3.Node, name string) (*yamlv3.Node, *yamlv3.Node) {
	if nodes == nil {
		return nil, nil
	}

	var keyNode, valueNode *yamlv3.Node
	for i, node := range nodes {
		if node.Value == name {
			keyNode = node
			if i < len(nodes)-1 {
				valueNode = nodes[i+1]
			}
		} else if node.Kind == yamlv3.SequenceNode || node.Kind == yamlv3.MappingNode {
			keyNode, valueNode = findNode(node.Content, name)
		}

		if keyNode != nil && valueNode != nil {
			break
		}
	}

	return keyNode, valueNode
}

func BuildKubernetesClusterTarget(name, namespace, kubeconfig string) (*lsv1alpha1.Target, error) {
	kubeconfigContent, err := ioutil.ReadFile(kubeconfig)
	if err != nil {
		return nil, fmt.Errorf("cannot read kubeconfig: %w", err)
	}

	config := lsv1alpha1.KubernetesClusterTargetConfig{
		Kubeconfig: string(kubeconfigContent),
	}

	marshalledConfig, err := json.Marshal(config)
	if err != nil {
		return nil, fmt.Errorf("%w", err)
	}

	target := &lsv1alpha1.Target{
		TypeMeta: metav1.TypeMeta{
			APIVersion: lsv1alpha1.SchemeGroupVersion.String(),
			Kind:       "Target",
		},
		ObjectMeta: metav1.ObjectMeta{
			Name:      name,
			Namespace: namespace,
		},
		Spec: lsv1alpha1.TargetSpec{
			Type: lsv1alpha1.KubernetesClusterTargetType,
			Configuration: lsv1alpha1.AnyJSON{
				marshalledConfig,
			},
		},
	}

	return target, nil
}
