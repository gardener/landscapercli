package quickstart

import (
	"context"
	"fmt"

	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	k8sErrors "k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/api/resource"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
)

type ociRegistry struct {
	deployment *appsv1.Deployment
	pvc        *corev1.PersistentVolumeClaim
	service    *corev1.Service
	k8sClient  kubernetes.Interface
	namespace  string
}

func NewOCIRegistry(namespace string, k8sClient kubernetes.Interface) *ociRegistry {
	deployment, pvc, service := createK8sObjects()

	obj := &ociRegistry{
		deployment: deployment,
		pvc:        pvc,
		service:    service,
		k8sClient:  k8sClient,
		namespace:  namespace,
	}

	return obj
}

func (r *ociRegistry) install(ctx context.Context) error {
	fmt.Printf("Creating Deployment %s in namespace %s\n", r.deployment.Name, r.namespace)
	_, err := r.k8sClient.AppsV1().Deployments(r.namespace).Create(ctx, r.deployment, metav1.CreateOptions{})
	if err != nil {
		if k8sErrors.IsAlreadyExists(err) {
			fmt.Println("Deployment already exists...Skipping")
		} else {
			return err
		}
	}

	fmt.Printf("Creating Persitent Volume Claim %s in namespace %s\n", r.pvc.Name, r.namespace)
	_, err = r.k8sClient.CoreV1().PersistentVolumeClaims(r.namespace).Create(ctx, r.pvc, metav1.CreateOptions{})
	if err != nil {
		if k8sErrors.IsAlreadyExists(err) {
			fmt.Println("Persitent Volume Claim already exists...Skipping")
		} else {
			return err
		}
	}

	fmt.Printf("Creating Service %s in namespace %s\n", r.service.Name, r.namespace)
	_, err = r.k8sClient.CoreV1().Services(r.namespace).Create(ctx, r.service, metav1.CreateOptions{})
	if err != nil {
		if k8sErrors.IsAlreadyExists(err) {
			fmt.Println("Service already exists...Skipping")
		} else {
			return err
		}
	}

	return nil
}

func (r *ociRegistry) uninstall(ctx context.Context) error {
	fmt.Printf("Deleting Deployment %s in namespace %s\n", r.deployment.Name, r.namespace)
	err := r.k8sClient.AppsV1().Deployments(r.namespace).Delete(ctx, r.deployment.Name, metav1.DeleteOptions{})
	if err != nil {
		if k8sErrors.IsNotFound(err) {
			fmt.Println("Deployment not found...Skipping")
		} else {
			return err
		}
	}

	fmt.Printf("Deleting Persitent Volume Claim %s in namespace %s\n", r.pvc.Name, r.namespace)
	err = r.k8sClient.CoreV1().PersistentVolumeClaims(r.namespace).Delete(ctx, r.pvc.Name, metav1.DeleteOptions{})
	if err != nil {
		if k8sErrors.IsNotFound(err) {
			fmt.Println("PersistentVolumeClaim not found...Skipping")
		} else {
			return err
		}
	}

	fmt.Printf("Deleting Service %s in namespace %s\n", r.service.Name, r.namespace)
	err = r.k8sClient.CoreV1().Services(r.namespace).Delete(ctx, r.service.Name, metav1.DeleteOptions{})
	if err != nil {
		if k8sErrors.IsNotFound(err) {
			fmt.Println("Service not found...Skipping")
		} else {
			return err
		}
	}

	return nil
}

func createK8sObjects() (*appsv1.Deployment, *corev1.PersistentVolumeClaim, *corev1.Service) {
	const (
		appName       = "oci-registry"
		pvcName       = "oci-registry-data"
		containerPort = 5000
	)

	var labels = map[string]string{
		"app": appName,
	}

	deployment := &appsv1.Deployment{
		ObjectMeta: metav1.ObjectMeta{
			Name:   appName,
			Labels: labels,
		},
		Spec: appsv1.DeploymentSpec{
			Selector: &metav1.LabelSelector{
				MatchLabels: labels,
			},
			Template: corev1.PodTemplateSpec{
				ObjectMeta: metav1.ObjectMeta{
					Labels: labels,
				},
				Spec: corev1.PodSpec{
					Containers: []corev1.Container{
						{
							Name:            "registry",
							Image:           "registry:2",
							ImagePullPolicy: corev1.PullIfNotPresent,
							Ports: []corev1.ContainerPort{
								{
									ContainerPort: containerPort,
								},
							},
							VolumeMounts: []corev1.VolumeMount{
								{
									Name:      "registry-data",
									MountPath: "/var/lib/registry",
								},
							},
						},
					},
					Volumes: []corev1.Volume{
						{
							Name: "registry-data",
							VolumeSource: corev1.VolumeSource{
								PersistentVolumeClaim: &corev1.PersistentVolumeClaimVolumeSource{
									ClaimName: pvcName,
								},
							},
						},
					},
				},
			},
		},
	}

	pvc := &corev1.PersistentVolumeClaim{
		ObjectMeta: metav1.ObjectMeta{
			Name:   pvcName,
			Labels: labels,
		},
		Spec: corev1.PersistentVolumeClaimSpec{
			AccessModes: []corev1.PersistentVolumeAccessMode{
				corev1.ReadWriteOnce,
			},
			Resources: corev1.ResourceRequirements{
				Requests: corev1.ResourceList{
					corev1.ResourceStorage: resource.MustParse("5Gi"),
				},
			},
		},
	}

	var service = &corev1.Service{
		ObjectMeta: metav1.ObjectMeta{
			Name:   appName,
			Labels: labels,
		},
		Spec: corev1.ServiceSpec{
			Type:     corev1.ServiceTypeClusterIP,
			Selector: labels,
			Ports: []corev1.ServicePort{
				{
					Port:     containerPort,
					Protocol: corev1.ProtocolTCP,
				},
			},
		},
	}

	return deployment, pvc, service
}
