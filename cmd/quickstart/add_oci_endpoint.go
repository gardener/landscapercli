// SPDX-FileCopyrightText: 2020 SAP SE or an SAP affiliate company and Gardener contributors.
//
// SPDX-License-Identifier: Apache-2.0

package quickstart

import (
	"context"
	"fmt"
	"os"
	"os/exec"
	"strings"

	"k8s.io/apimachinery/pkg/util/intstr"

	"github.com/go-logr/logr"
	"github.com/spf13/cobra"
	"github.com/spf13/pflag"
	v1 "k8s.io/api/core/v1"
	"k8s.io/api/extensions/v1beta1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/tools/clientcmd"
	"sigs.k8s.io/controller-runtime/pkg/client"

	"github.com/gardener/landscapercli/pkg/logger"
)

const (
	authSecretName = "oci-secret"
	ingressName    = "oci-ingress"
	tlsSecretName  = "oci-tls-secret"
)

type addOciEndpointOptions struct {
	kubeconfigPath string
	namespace      string
	user           string
	password       string
}

// NewCreateCommand creates a new blueprint command to create a blueprint
func NewAddOciEndpointCommand(ctx context.Context) *cobra.Command {
	opts := &addOciEndpointOptions{}
	cmd := &cobra.Command{
		Use:  "add-oci-endpoint",
		Args: cobra.ExactArgs(0),
		Example: "landscaper-cli quickstart add-oci-endpoint \\\n" +
			"    --kubeconfig ./kubconfig.yaml \\\n" +
			"    --namespace landscaper \\\n" +
			"    --user testuser \\\n" +
			"    --password sic7a5snk",
		Short: "command to add an external https endpoint for the oci registry (experimental)",
		Long: "This command add an external https endpoint to an oci registry installed with the Landscaper CLI quickstart " +
			"command. This command is only supported for garden shoot clusters with activated nginx and a cert manager. " +
			"Furthermore htpasswd must be installed on your local machine.",
		Run: func(cmd *cobra.Command, args []string) {
			if err := opts.Complete(args); err != nil {
				fmt.Println(err.Error())
				os.Exit(1)
			}

			if err := opts.run(ctx, logger.Log); err != nil {
				fmt.Println(err.Error())
				os.Exit(1)
			}
		},
	}

	opts.AddFlags(cmd.Flags())

	return cmd
}

func (o *addOciEndpointOptions) AddFlags(fs *pflag.FlagSet) {
	fs.StringVar(&o.kubeconfigPath, "kubeconfig", "", "path to the kubeconfig of the target cluster")
	fs.StringVar(&o.namespace, "namespace", defaultNamespace, "namespace where the landscaper and the OCI registry are installed")
	fs.StringVar(&o.user, "user", "", "user for authentication at the oci endpoint")
	fs.StringVar(&o.password, "password", "", "password for authentication at the oci endpoint")
}

func (o *addOciEndpointOptions) Complete(args []string) error {
	return o.validate()
}

func (o *addOciEndpointOptions) validate() error {
	if o.kubeconfigPath == "" {
		return fmt.Errorf("no kubeconfig provided")
	}

	if o.namespace == "" {
		return fmt.Errorf("no namespace provided")
	}

	if o.user == "" {
		return fmt.Errorf("no user provided")
	}

	if o.password == "" {
		return fmt.Errorf("no password provided")
	}

	return nil
}

func (o *addOciEndpointOptions) run(ctx context.Context, log logr.Logger) error {
	cfg, err := clientcmd.BuildConfigFromFlags("", o.kubeconfigPath)
	if err != nil {
		return fmt.Errorf("cannot parse K8s config: %w", err)
	}

	k8sClient, err := client.New(cfg, client.Options{})
	if err != nil {
		return fmt.Errorf("cannot build K8s client: %w", err)
	}

	ingressHost := strings.Replace(cfg.Host, "https://api", "o.ingress", 1)
	if len(ingressHost) > 64 {
		return fmt.Errorf("No certificate could be created because your domain exceeds 64 characters: " + ingressHost)
	}

	err = o.createAuthSecret(ctx, k8sClient, authSecretName)
	if err != nil {
		return err
	}

	err = o.createIngress(ctx, k8sClient, authSecretName, ingressName, ingressHost, tlsSecretName)
	if err != nil {
		return err
	}

	fmt.Println("Installation succeeded - it might need some minutes until the certificate is created")
	fmt.Println("The OCI endpoint is: https://" + ingressHost)

	return nil
}

func (o *addOciEndpointOptions) createAuthSecret(ctx context.Context, k8sClient client.Client, authSecretName string) error {
	cmd := exec.Command("htpasswd", "-n", "-b", o.user, o.password)
	passwordBytes, err := cmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("failed to encrypt password: %w", err)
	}

	secret := v1.Secret{
		ObjectMeta: metav1.ObjectMeta{
			Name:      authSecretName,
			Namespace: o.namespace,
		},
		StringData: map[string]string{
			"auth": string(passwordBytes),
		},
		Type: v1.SecretTypeOpaque,
	}

	err = k8sClient.Create(ctx, &secret)
	if err != nil {
		return fmt.Errorf("failed to create authentication secret: %w", err)
	}

	fmt.Printf("created secret %s in namespace %s with the authentication data for the oci endpoint",
		authSecretName, o.namespace)

	return nil
}

func (o *addOciEndpointOptions) createIngress(ctx context.Context, k8sClient client.Client,
	authSecretName, ingressName, ingressHost, tlsSecretName string) error {

	ingress := v1beta1.Ingress{
		ObjectMeta: metav1.ObjectMeta{
			Name:      ingressName,
			Namespace: o.namespace,
			Annotations: map[string]string{
				"cert.gardener.cloud/purpose":             "managed",
				"dns.gardener.cloud/class":                "garden",
				"dns.gardener.cloud/dnsnames":             ingressHost,
				"nginx.ingress.kubernetes.io/auth-type":   "basic",
				"nginx.ingress.kubernetes.io/auth-secret": authSecretName,
				"nginx.ingress.kubernetes.io/auth-realm":  "Authentication Required",
			},
		},
		Spec: v1beta1.IngressSpec{
			TLS: []v1beta1.IngressTLS{
				{
					Hosts: []string{
						ingressHost,
					},
					SecretName: tlsSecretName, // gardener will create a secret with this name
				},
			},
			Rules: []v1beta1.IngressRule{
				{
					Host: ingressHost,
					IngressRuleValue: v1beta1.IngressRuleValue{
						HTTP: &v1beta1.HTTPIngressRuleValue{
							Paths: []v1beta1.HTTPIngressPath{
								{
									Path: "/",
									Backend: v1beta1.IngressBackend{
										ServiceName: "oci-registry",
										ServicePort: intstr.IntOrString{
											IntVal: 5000,
											Type:   intstr.Int,
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

	err := k8sClient.Create(ctx, &ingress)
	if err != nil {
		return fmt.Errorf("failed to create ingress: %w", err)
	}

	fmt.Printf("created ingress %s in namespace %s", ingressName, o.namespace)

	return nil
}
