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
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/types"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

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

func WaitUntilAllPodsAreReady(k8sClient client.Client, namespace string, sleepTime time.Duration, maxRetries int) error {
	ctx := context.TODO()
	retries := 0
	for {
		podList := corev1.PodList{}
		err := k8sClient.List(ctx, &podList, client.InNamespace(namespace))
		if err != nil {
			return fmt.Errorf("Cannot list pods: %w", err)
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

		if retries >= maxRetries {
			return fmt.Errorf("Pods not ready after sleepTime=%dns and maxRetries=%d", sleepTime, maxRetries)
		}
		retries++
		
		time.Sleep(sleepTime)
	}

	return nil
}

func WaitUntilLandscaperInstallationSucceeded(k8sClient client.Client, key types.NamespacedName, sleepTime time.Duration, maxRetries int) error {
	ctx := context.TODO()
	retries := 0
	inst := &landscaper.Installation{}
	
	for {
		err := k8sClient.Get(ctx, key, inst)
		if err != nil {
			return fmt.Errorf("Cannot get installation: %w", err)
		}

		if inst.Status.Phase == landscaper.ComponentPhaseSucceeded {
			break
		}

		if retries >= maxRetries {
			return fmt.Errorf("Installation not succeeded after sleepTime=%dns and maxRetries=%d", sleepTime, maxRetries)
		}
		retries++
		
		time.Sleep(sleepTime)
	}

	return nil
}

func WaitUntilObjectIsDeleted(k8sClient client.Client, objKey types.NamespacedName, obj runtime.Object, sleepTime time.Duration, maxRetries int) error {
	ctx := context.TODO()
	retries := 0

	for {
		err := k8sClient.Get(ctx, objKey, obj)
		if err != nil {
			if k8sErrors.IsNotFound(err) {
				break
			}
			return err
		}

		if retries >= maxRetries {
			return fmt.Errorf("Object %s still exists after sleepTime=%dns and maxRetries=%d", objKey, sleepTime, maxRetries)
		}
		retries++
		
		time.Sleep(sleepTime)
	}

	return nil
}
