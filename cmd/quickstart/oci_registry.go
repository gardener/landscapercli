package quickstart

import (
	"context"
	"fmt"
	"os/exec"

	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/api/extensions/v1beta1"
	networking "k8s.io/api/networking/v1"
	k8sErrors "k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/api/resource"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

type ociRegistry struct {
	k8sClient client.Client
	opts      ociRegistryOpts
}

type ociRegistryOpts struct {
	namespace      string
	installIngress bool
	ingressHost    string
	username       string
	password       string

	// set during execution
	ingressAuthData []byte
}

func NewOCIRegistry(opts *ociRegistryOpts, k8sClient client.Client) *ociRegistry {
	obj := &ociRegistry{
		k8sClient: k8sClient,
		opts:      *opts,
	}
	return obj
}

func (r *ociRegistry) install(ctx context.Context) error {
	if r.opts.installIngress {
		cmd := exec.Command("htpasswd", "-n", "-b", r.opts.username, r.opts.password)
		ingressAuthData, err := cmd.CombinedOutput()
		if err != nil {
			return fmt.Errorf("failed to encrypt ingress credentials: %w", err)
		}
		r.opts.ingressAuthData = ingressAuthData
	}

	deployment, pvc, service, authSecret, ingress, _ := r.createK8sObjects()

	fmt.Printf("Creating deployment %s in namespace %s\n", deployment.Name, r.opts.namespace)
	if err := r.k8sClient.Create(ctx, deployment, &client.CreateOptions{}); err != nil {
		if k8sErrors.IsAlreadyExists(err) {
			fmt.Println("Deployment already exists...Skipping")
		} else {
			return err
		}
	}

	fmt.Printf("Creating persitent volume claim %s in namespace %s\n", pvc.Name, r.opts.namespace)
	if err := r.k8sClient.Create(ctx, pvc, &client.CreateOptions{}); err != nil {
		if k8sErrors.IsAlreadyExists(err) {
			fmt.Println("Persitent Volume Claim already exists...Skipping")
		} else {
			return err
		}
	}

	fmt.Printf("Creating service %s in namespace %s\n", service.Name, r.opts.namespace)
	if err := r.k8sClient.Create(ctx, service, &client.CreateOptions{}); err != nil {
		if k8sErrors.IsAlreadyExists(err) {
			fmt.Println("Service already exists...Skipping")
		} else {
			return err
		}
	}

	if r.opts.installIngress {
		fmt.Printf("Creating ingress authentication secret %s in namespace %s\n", authSecret.Name, r.opts.namespace)
		if err := r.k8sClient.Create(ctx, authSecret, &client.CreateOptions{}); err != nil {
			if k8sErrors.IsAlreadyExists(err) {
				fmt.Println("Secret already exists...Skipping")
			} else {
				return err
			}
		}

		fmt.Printf("Creating ingress %s in namespace %s\n", ingress.Name, r.opts.namespace)
		if err := r.k8sClient.Create(ctx, ingress, &client.CreateOptions{}); err != nil {
			if k8sErrors.IsAlreadyExists(err) {
				fmt.Println("Ingress already exists...Skipping")
			} else {
				return err
			}
		}
	}

	return nil
}

func (r *ociRegistry) uninstall(ctx context.Context) error {
	deployment, pvc, service, authSecret, ingress, oldIngress := r.createK8sObjects()

	fmt.Printf("Deleting deployment %s in namespace %s\n", deployment.Name, r.opts.namespace)
	if err := r.k8sClient.Delete(ctx, deployment, &client.DeleteOptions{}); err != nil {
		if k8sErrors.IsNotFound(err) {
			fmt.Println("Deployment not found...Skipping")
		} else {
			return err
		}
	}

	fmt.Printf("Deleting persitent volume claim %s in namespace %s\n", pvc.Name, r.opts.namespace)
	if err := r.k8sClient.Delete(ctx, pvc, &client.DeleteOptions{}); err != nil {
		if k8sErrors.IsNotFound(err) {
			fmt.Println("PersistentVolumeClaim not found...Skipping")
		} else {
			return err
		}
	}

	fmt.Printf("Deleting service %s in namespace %s\n", service.Name, r.opts.namespace)
	if err := r.k8sClient.Delete(ctx, service, &client.DeleteOptions{}); err != nil {
		if k8sErrors.IsNotFound(err) {
			fmt.Println("Service not found...Skipping")
		} else {
			return err
		}
	}

	fmt.Printf("Deleting ingress %s in namespace %s\n", ingress.Name, r.opts.namespace)
	if err := r.k8sClient.Delete(ctx, ingress, &client.DeleteOptions{}); err != nil {
		if k8sErrors.IsNotFound(err) {
			fmt.Println("Ingress not found...Skipping")
		} else {
			fmt.Printf("Deleting old ingress %s in namespace %s\n", ingress.Name, r.opts.namespace)
			err2 := r.k8sClient.Delete(ctx, oldIngress, &client.DeleteOptions{})
			if k8sErrors.IsNotFound(err2) {
				fmt.Println("Old ingress not found...Skipping")
			} else {
				return err
			}
		}
	}

	fmt.Printf("Deleting ingress authentication secret %s in namespace %s\n", authSecret.Name, r.opts.namespace)
	if err := r.k8sClient.Delete(ctx, authSecret, &client.DeleteOptions{}); err != nil {
		if k8sErrors.IsNotFound(err) {
			fmt.Println("Ingress authentication secret not found...Skipping")
		} else {
			return err
		}
	}

	tlsSecretName := ingress.Spec.TLS[0].SecretName
	fmt.Printf("Deleting ingress tls secret %s in namespace %s\n", tlsSecretName, r.opts.namespace)
	tlsSecret := &corev1.Secret{
		ObjectMeta: metav1.ObjectMeta{
			Name:      tlsSecretName,
			Namespace: r.opts.namespace,
		},
	}
	if err := r.k8sClient.Delete(ctx, tlsSecret, &client.DeleteOptions{}); err != nil {
		if k8sErrors.IsNotFound(err) {
			fmt.Println("Ingress tls secret not found...Skipping")
		} else {
			return err
		}
	}

	return nil
}

func (r *ociRegistry) createK8sObjects() (*appsv1.Deployment, *corev1.PersistentVolumeClaim, *corev1.Service, *corev1.Secret, *networking.Ingress, *v1beta1.Ingress) {
	const (
		appName        = "oci-registry"
		pvcName        = appName + "-data"
		pvcSize        = "5Gi"
		authSecretName = appName + "-auth"
		tlsSecretName  = appName + "-tls"
		containerPort  = 5000
	)

	labels := map[string]string{
		"app": appName,
	}

	deployment := &appsv1.Deployment{
		ObjectMeta: metav1.ObjectMeta{
			Name:      appName,
			Namespace: r.opts.namespace,
			Labels:    labels,
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
			Name:      pvcName,
			Namespace: r.opts.namespace,
			Labels:    labels,
		},
		Spec: corev1.PersistentVolumeClaimSpec{
			AccessModes: []corev1.PersistentVolumeAccessMode{
				corev1.ReadWriteOnce,
			},
			Resources: corev1.VolumeResourceRequirements{
				Requests: corev1.ResourceList{
					corev1.ResourceStorage: resource.MustParse(pvcSize),
				},
			},
		},
	}

	service := &corev1.Service{
		ObjectMeta: metav1.ObjectMeta{
			Name:      appName,
			Namespace: r.opts.namespace,
			Labels:    labels,
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

	authSecret := &corev1.Secret{
		ObjectMeta: metav1.ObjectMeta{
			Name:      authSecretName,
			Namespace: r.opts.namespace,
			Labels:    labels,
		},
		StringData: map[string]string{
			"auth": string(r.opts.ingressAuthData),
		},
		Type: corev1.SecretTypeOpaque,
	}

	pathType := networking.PathTypePrefix
	ingress := &networking.Ingress{
		ObjectMeta: metav1.ObjectMeta{
			Name:      appName,
			Namespace: r.opts.namespace,
			Labels:    labels,
			Annotations: map[string]string{
				"cert.gardener.cloud/purpose":                 "managed",
				"dns.gardener.cloud/class":                    "garden",
				"dns.gardener.cloud/dnsnames":                 r.opts.ingressHost,
				"nginx.ingress.kubernetes.io/auth-type":       "basic",
				"nginx.ingress.kubernetes.io/auth-secret":     authSecretName,
				"nginx.ingress.kubernetes.io/auth-realm":      "Authentication Required",
				"nginx.ingress.kubernetes.io/proxy-body-size": "100m",
			},
		},
		Spec: networking.IngressSpec{
			TLS: []networking.IngressTLS{
				{
					Hosts: []string{
						r.opts.ingressHost,
					},
					SecretName: tlsSecretName, // gardener will create a secret with this name
				},
			},
			Rules: []networking.IngressRule{
				{
					Host: r.opts.ingressHost,
					IngressRuleValue: networking.IngressRuleValue{
						HTTP: &networking.HTTPIngressRuleValue{
							Paths: []networking.HTTPIngressPath{
								{
									Path:     "/",
									PathType: &pathType,
									Backend: networking.IngressBackend{
										Service: &networking.IngressServiceBackend{
											Name: appName,
											Port: networking.ServiceBackendPort{
												Number: containerPort,
											},
										},
									},
								},
							},
						},
					},
				},
			},
		},
	}

	oldIngress := &v1beta1.Ingress{
		ObjectMeta: metav1.ObjectMeta{
			Name:      appName,
			Namespace: r.opts.namespace,
		},
	}

	return deployment, pvc, service, authSecret, ingress, oldIngress
}
