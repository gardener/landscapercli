package quickstart

import (
	"context"
	"fmt"
	"time"

	"k8s.io/apimachinery/pkg/util/wait"

	extv1 "k8s.io/apiextensions-apiserver/pkg/apis/apiextensions/v1"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

func removeObjects(ctx context.Context, k8sClient client.Client, crd *extv1.CustomResourceDefinition) error {
	for k := range crd.Spec.Versions {
		gvk := schema.GroupVersionKind{
			Group:   crd.Spec.Group,
			Version: crd.Spec.Versions[k].Name,
			Kind:    crd.Spec.Names.Kind,
		}

		fmt.Printf("Removing objects of type %s\n", gvk)

		objectList := &unstructured.UnstructuredList{}
		objectList.SetGroupVersionKind(gvk)
		if err := k8sClient.List(ctx, objectList); err != nil {
			return err
		}

		for i := range objectList.Items {
			item := &objectList.Items[i]
			if err := removeObject(ctx, k8sClient, item); err != nil {
				return err
			}
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

func removeObjectsPatiently(ctx context.Context, k8sClient client.Client, crd *extv1.CustomResourceDefinition) error {
	for k := range crd.Spec.Versions {
		gvk := schema.GroupVersionKind{
			Group:   crd.Spec.Group,
			Version: crd.Spec.Versions[k].Name,
			Kind:    crd.Spec.Names.Kind,
		}

		fmt.Printf("Removing objects of type %s\n", gvk)

		objectList := &unstructured.UnstructuredList{}
		objectList.SetGroupVersionKind(gvk)
		if err := k8sClient.List(ctx, objectList); err != nil {
			return err
		}

		for i := range objectList.Items {
			item := &objectList.Items[i]
			if err := k8sClient.Delete(ctx, item); err != nil {
				return err
			}
		}

		err := wait.PollUntilContextTimeout(ctx, 10*time.Second, 4*time.Minute, true, func(ctx context.Context) (done bool, err error) {
			fmt.Println("waiting for deployer registrations removed")

			remainingObjectList := &unstructured.UnstructuredList{}
			remainingObjectList.SetGroupVersionKind(gvk)
			if err := k8sClient.List(ctx, remainingObjectList); err != nil {
				return false, err
			}

			return len(remainingObjectList.Items) == 0, nil
		})
		if err != nil {
			return fmt.Errorf("error while waiting for deployer registrations being removed: %w", err)
		}
	}

	return nil
}
