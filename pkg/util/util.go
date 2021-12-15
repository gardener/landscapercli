package util

import (
	"context"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"time"

	cdv2 "github.com/gardener/component-spec/bindings-go/apis/v2"
	lsv1alpha1 "github.com/gardener/landscaper/apis/core/v1alpha1"
	"github.com/gardener/landscaper/apis/mediatype"
	corev1 "k8s.io/api/core/v1"
	k8sErrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/utils/pointer"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

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
			return false, fmt.Errorf("cannot list pods: %w", err)
		}

		numberOfRunningPods := 0
		for _, pod := range podList.Items {
			if pod.Status.Phase == corev1.PodSucceeded {
				numberOfRunningPods++
			} else {
				for _, condition := range pod.Status.Conditions {
					if condition.Type == corev1.PodReady {
						if condition.Status == corev1.ConditionTrue {
							numberOfRunningPods++
							break
						}
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
		return fmt.Errorf("Deleting namespace gracefully timed out")
	}
	return nil
}

// gracefully try to delete all Landscaper installations in a namespace, then delete the namespace itself
func gracefullyDeleteNamespace(k8sClient client.Client, namespace string, sleepTime time.Duration, maxRetries int) (bool, error) {
	ctx := context.TODO()

	installationList := lsv1alpha1.InstallationList{}
	if err := k8sClient.List(ctx, &installationList, &client.ListOptions{Namespace: namespace}); err != nil {
		return false, err
	}
	for _, installation := range installationList.Items {
		fmt.Println("Deleting installation:", installation.Name)
		if err := k8sClient.Delete(ctx, &installation, &client.DeleteOptions{}); err != nil {
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

	podList := corev1.PodList{}
	if err := k8sClient.List(ctx, &podList, &client.ListOptions{Namespace: namespace}); err != nil {
		return false, err
	}

	if len(podList.Items) > 0 {
		for i := range podList.Items {
			next := &podList.Items[i]

			if len(next.GetFinalizers()) > 0 {
				return false, fmt.Errorf("there still exists pods with finalizer in namespace %s though all installations were deleted: %w", namespace, err)
			}
		}
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

func BuildKubernetesClusterTarget(name, namespace, kubeconfigPath string) (*lsv1alpha1.Target, error) {
	kubeconfigContent, err := ioutil.ReadFile(kubeconfigPath)
	if err != nil {
		return nil, fmt.Errorf("cannot read kubeconfig: %w", err)
	}

	config := lsv1alpha1.KubernetesClusterTargetConfig{
		Kubeconfig: lsv1alpha1.ValueRef{
			StrVal: pointer.StringPtr(string(kubeconfigContent)),
		},
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
				RawMessage: marshalledConfig,
			},
		},
	}

	return target, nil
}

func GetBlueprintResource(cd *cdv2.ComponentDescriptor, blueprintResourceName string) (*cdv2.Resource, error) {
	blueprintResources := map[string]cdv2.Resource{}
	for _, resource := range cd.ComponentSpec.Resources {
		if resource.IdentityObjectMeta.Type == mediatype.BlueprintType || resource.IdentityObjectMeta.Type == mediatype.OldBlueprintType {
			blueprintResources[resource.Name] = resource
		}
	}

	var blueprintRes cdv2.Resource
	numberOfBlueprints := len(blueprintResources)
	if numberOfBlueprints == 0 {
		return nil, fmt.Errorf("no blueprint resources defined in the component descriptor")
	} else if numberOfBlueprints == 1 && blueprintResourceName == "" {
		// access the only blueprint in the map. the flag blueprint-resource-name is ignored in this case.
		for _, entry := range blueprintResources {
			blueprintRes = entry
		}
	} else {
		if blueprintResourceName == "" {
			return nil, fmt.Errorf("the blueprint resource name must be defined since multiple blueprint resources exist in the component descriptor")
		}
		ok := false
		blueprintRes, ok = blueprintResources[blueprintResourceName]
		if !ok {
			return nil, fmt.Errorf("blueprint %s is not defined as a resource in the component descriptor", blueprintResourceName)
		}
	}

	return &blueprintRes, nil
}
