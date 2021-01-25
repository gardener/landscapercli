package util

import (
	"context"
	"fmt"
	"os"
	"os/exec"
	"strings"
	"time"

	landscaper "github.com/gardener/landscaper/pkg/apis/core/v1alpha1"
	corev1 "k8s.io/api/core/v1"
	k8sErrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
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

	cmd := exec.Command(arr[0], arr[1:]...)
	cmd.Env = []string{"HELM_EXPERIMENTAL_OCI=1", "HOME=" + os.Getenv("HOME"), "PATH=" + os.Getenv("PATH")}
	out, err := cmd.CombinedOutput()

	if err != nil {
		fmt.Printf("Failed with error: %s:\n%s\n", err, string(out))
		return err
	}
	fmt.Println("Executed sucessfully!")

	return nil
}

// ExecCommandNonBlocking executes a command without without blocking. Returns a Cmd that can be used to stop the command.
func ExecCommandNonBlocking(command string) (*exec.Cmd, error) {
	fmt.Printf("Executing: %s\n", command)

	arr := strings.Split(command, " ")

	cmd := exec.Command(arr[0], arr[1:]...)
	cmd.Env = []string{"HELM_EXPERIMENTAL_OCI=1", "HOME=" + os.Getenv("HOME"), "PATH=" + os.Getenv("PATH")}
	err := cmd.Start()

	if err != nil {
		fmt.Printf("Failed with error: %s:\n", err)
		return nil, err
	}
	fmt.Println("Executed sucessfully!")

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
		inst := &landscaper.Installation{}
		err := k8sClient.Get(context.TODO(), key, inst)
		if err != nil {
			return false, fmt.Errorf("cannot get installation: %w", err)
		}

		return inst.Status.Phase == landscaper.ComponentPhaseSucceeded, nil
	}

	return CheckConditionPeriodically(conditionFunc, sleepTime, maxRetries)
}

// CheckAndWaitUntilObjectNotExistAnymore periodically checks and wait until the object does not exist anymore. Returns an error on failure
// or if timeout (sleepTime and maxRetries is reached). Returns a boolean indicating if the object remains on return.
func CheckAndWaitUntilObjectNotExistAnymore(k8sClient client.Client, objKey types.NamespacedName, obj runtime.Object, sleepTime time.Duration, maxRetries int) (bool, error) {
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

		installationList := &landscaper.InstallationList{}
		err := k8sClient.List(ctx, installationList, &client.ListOptions{Namespace: namespace})
		if err != nil {
			return false, err
		}

		return len(installationList.Items) == 0, nil
	}

	return CheckConditionPeriodically(conditionFunc, sleepTime, maxRetries)
}

// DeleteNamespace deletes a namespace (if it exists). First, a graceful delete will be tried. On timeout, the finalizers will be deleted
// and the namespace will be deleted forcefully.
func DeleteNamespace(k8sClient client.Client, namespace string, sleepTime time.Duration, maxRetries int) error {
	namespaceToCheck := &corev1.Namespace{
		ObjectMeta: metav1.ObjectMeta{
			Name: namespace,
		},
	}
	err := k8sClient.Get(context.TODO(), client.ObjectKey{Name: namespace}, namespaceToCheck)
	if err != nil {
		if k8sErrors.IsNotFound(err) {
			return nil
		}
		return fmt.Errorf("Error getting namespace %s: %w", namespace, err)
	}

	timeout, err := gracefulyDeleteNamespace(k8sClient, namespace, sleepTime, maxRetries)
	if err != nil {
		fmt.Printf("Deleting namespace gracefully failed with error %s, using foce delete...", err.Error())
		forceDeleteNamespace(k8sClient, namespace, sleepTime, maxRetries)
		return err
	}
	if timeout {
		fmt.Printf("Deleting namespace gracefully timed out, using force delete...")
		forceDeleteNamespace(k8sClient, namespace, sleepTime, maxRetries)
	}
	return nil
}

// gracefully try to delete all Landscaper installations in a namespace, then delete the namespace itself
func gracefulyDeleteNamespace(k8sClient client.Client, namespace string, sleepTime time.Duration, maxRetries int) (bool, error) {
	ctx := context.TODO()

	installationList := landscaper.InstallationList{}
	err := k8sClient.List(ctx, &installationList, &client.ListOptions{Namespace: namespace})
	if err != nil {
		return false, err
	}
	for _, installation := range installationList.Items {
		fmt.Println("Deleting installation:", installation.Name)
		err := k8sClient.Delete(ctx, &installation, &client.DeleteOptions{})
		if err != nil {
			return false, fmt.Errorf("Cannot delete installation: %w", err)
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
		return false, err
	}

	return timeout, nil
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
		//unterscheide zwischen timeout und general error
		return false, err
	}

	return timeout, nil

}

// removes the finalizers from all Landscaper CRs in a namespace
func removeFinalizersFromLandscaperCRs(k8sClient client.Client, namespace string) error {
	ctx := context.TODO()

	installationList := landscaper.InstallationList{}
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

	executionList := &landscaper.ExecutionList{}
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

	deployItemList := &landscaper.DeployItemList{}
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
	return k8sClient.Update(ctx, object.(runtime.Object))
}
