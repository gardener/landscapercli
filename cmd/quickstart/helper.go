package quickstart

import (
	"context"
	"fmt"
	"github.com/gardener/landscaper/apis/core/v1alpha1"
	extv1 "k8s.io/apiextensions-apiserver/pkg/apis/apiextensions/v1"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"time"
)

var crdNames = []string{
	"componentversionoverwrites.landscaper.gardener.cloud",
	"contexts.landscaper.gardener.cloud",
	"dataobjects.landscaper.gardener.cloud",
	"deployerregistrations.landscaper.gardener.cloud",
	"deployitems.landscaper.gardener.cloud",
	"environments.landscaper.gardener.cloud",
	"executions.landscaper.gardener.cloud",
	"installations.landscaper.gardener.cloud",
	"lshealthchecks.landscaper.gardener.cloud",
	"syncobjects.landscaper.gardener.cloud",
	"targets.landscaper.gardener.cloud",
	"targetsyncs.landscaper.gardener.cloud",
}

func stringInSlice(str string, list []string) bool {
	for _, v := range list {
		if v == str {
			return true
		}
	}
	return false
}

func removeCrdTargetSync(ctx context.Context, k8sClient client.Client, nextName string,
	objectList *v1alpha1.TargetSyncList, nextCrd *extv1.CustomResourceDefinition) error {
	fmt.Println("Removing objects of CRD: " + nextName)

	if err := k8sClient.List(ctx, objectList); err != nil {
		return err
	}

	for i := range objectList.Items {
		nextItem := &objectList.Items[i]
		if err := removeObject(ctx, k8sClient, nextItem); err != nil {
			return err
		}
	}

	return nil
}

func removeCrdTarget(ctx context.Context, k8sClient client.Client, nextName string,
	objectList *v1alpha1.TargetList, nextCrd *extv1.CustomResourceDefinition) error {
	fmt.Println("Removing objects of CRD: " + nextName)

	if err := k8sClient.List(ctx, objectList); err != nil {
		return err
	}

	for i := range objectList.Items {
		nextItem := &objectList.Items[i]
		if err := removeObject(ctx, k8sClient, nextItem); err != nil {
			return err
		}
	}

	return nil
}

func removeCrdSyncObject(ctx context.Context, k8sClient client.Client, nextName string,
	objectList *v1alpha1.SyncObjectList, nextCrd *extv1.CustomResourceDefinition) error {
	fmt.Println("Removing objects of CRD: " + nextName)

	if err := k8sClient.List(ctx, objectList); err != nil {
		return err
	}

	for i := range objectList.Items {
		nextItem := &objectList.Items[i]
		if err := removeObject(ctx, k8sClient, nextItem); err != nil {
			return err
		}
	}

	return nil
}

func removeCrdLsHealthCheck(ctx context.Context, k8sClient client.Client, nextName string,
	objectList *v1alpha1.LsHealthCheckList, nextCrd *extv1.CustomResourceDefinition) error {
	fmt.Println("Removing objects of CRD: " + nextName)

	if err := k8sClient.List(ctx, objectList); err != nil {
		return err
	}

	for i := range objectList.Items {
		nextItem := &objectList.Items[i]
		if err := removeObject(ctx, k8sClient, nextItem); err != nil {
			return err
		}
	}

	return nil
}

func removeCrdInstallation(ctx context.Context, k8sClient client.Client, nextName string,
	objectList *v1alpha1.InstallationList, nextCrd *extv1.CustomResourceDefinition) error {
	fmt.Println("Removing objects of CRD: " + nextName)

	if err := k8sClient.List(ctx, objectList); err != nil {
		return err
	}

	for i := range objectList.Items {
		nextItem := &objectList.Items[i]
		if err := removeObject(ctx, k8sClient, nextItem); err != nil {
			return err
		}
	}

	return nil
}

func removeCrdExecution(ctx context.Context, k8sClient client.Client, nextName string,
	objectList *v1alpha1.ExecutionList, nextCrd *extv1.CustomResourceDefinition) error {
	fmt.Println("Removing objects of CRD: " + nextName)

	if err := k8sClient.List(ctx, objectList); err != nil {
		return err
	}

	for i := range objectList.Items {
		nextItem := &objectList.Items[i]
		if err := removeObject(ctx, k8sClient, nextItem); err != nil {
			return err
		}
	}

	return nil
}

func removeCrdEnvironment(ctx context.Context, k8sClient client.Client, nextName string,
	objectList *v1alpha1.EnvironmentList, nextCrd *extv1.CustomResourceDefinition) error {
	fmt.Println("Removing objects of CRD: " + nextName)

	if err := k8sClient.List(ctx, objectList); err != nil {
		return err
	}

	for i := range objectList.Items {
		nextItem := &objectList.Items[i]
		if err := removeObject(ctx, k8sClient, nextItem); err != nil {
			return err
		}
	}

	return nil
}

func removeCrdDeployItem(ctx context.Context, k8sClient client.Client, nextName string,
	objectList *v1alpha1.DeployItemList, nextCrd *extv1.CustomResourceDefinition) error {
	fmt.Println("Removing objects of CRD: " + nextName)

	if err := k8sClient.List(ctx, objectList); err != nil {
		return err
	}

	for i := range objectList.Items {
		nextItem := &objectList.Items[i]
		if err := removeObject(ctx, k8sClient, nextItem); err != nil {
			return err
		}
	}

	return nil
}

func removeCrdDeployerRegistration(ctx context.Context, k8sClient client.Client, nextName string,
	objectList *v1alpha1.DeployerRegistrationList, nextCrd *extv1.CustomResourceDefinition) error {
	fmt.Println("Removing objects of CRD: " + nextName)

	if err := k8sClient.List(ctx, objectList); err != nil {
		return err
	}

	for i := range objectList.Items {
		nextItem := &objectList.Items[i]
		if err := removeObject(ctx, k8sClient, nextItem); err != nil {
			return err
		}
	}

	return nil
}

func removeCrdDataObject(ctx context.Context, k8sClient client.Client, nextName string,
	objectList *v1alpha1.DataObjectList, nextCrd *extv1.CustomResourceDefinition) error {
	fmt.Println("Removing objects of CRD: " + nextName)

	if err := k8sClient.List(ctx, objectList); err != nil {
		return err
	}

	for i := range objectList.Items {
		nextItem := &objectList.Items[i]
		if err := removeObject(ctx, k8sClient, nextItem); err != nil {
			return err
		}
	}

	return nil
}

func removeCrdContext(ctx context.Context, k8sClient client.Client, nextName string,
	objectList *v1alpha1.ContextList, nextCrd *extv1.CustomResourceDefinition) error {
	fmt.Println("Removing objects of CRD: " + nextName)

	if err := k8sClient.List(ctx, objectList); err != nil {
		return err
	}

	for i := range objectList.Items {
		nextItem := &objectList.Items[i]
		if err := removeObject(ctx, k8sClient, nextItem); err != nil {
			return err
		}
	}

	return nil
}

func removeCrdComponentversionoverwrites(ctx context.Context, k8sClient client.Client, nextName string,
	objectList *v1alpha1.ComponentVersionOverwritesList, nextCrd *extv1.CustomResourceDefinition) error {
	fmt.Println("Removing objects of CRD: " + nextName)

	if err := k8sClient.List(ctx, objectList); err != nil {
		return err
	}

	for i := range objectList.Items {
		nextItem := &objectList.Items[i]
		if err := removeObject(ctx, k8sClient, nextItem); err != nil {
			return err
		}
	}

	return nil
}

func removeObject(ctx context.Context, k8sClient client.Client, object client.Object) error {
	var err error

	fmt.Printf("Removing object: %s\n", client.ObjectKeyFromObject(object).String())

	for i := 0; i < 10; i++ {
		if err = k8sClient.Get(ctx, client.ObjectKeyFromObject(object), object); err != nil {
			if apierrors.IsNotFound(err) {
				return nil
			}
			fmt.Printf("Removing object: get failed: %s\n", err.Error())
			continue
		}

		object.SetFinalizers(nil)
		if err = k8sClient.Update(ctx, object); err != nil {
			if apierrors.IsNotFound(err) {
				return nil
			}

			fmt.Printf("Removing object: update failed: %s\n", err.Error())
			continue
		}

		if err = k8sClient.Delete(ctx, object); err != nil {
			if apierrors.IsNotFound(err) {
				return nil
			}

			fmt.Printf("Removing object: delete failed: %s\n", err.Error())
			continue
		}

		time.Sleep(time.Millisecond * 10)
	}

	if err != nil {
		return err
	} else {
		return fmt.Errorf("object could not be removed %s", client.ObjectKeyFromObject(object).String())
	}
}