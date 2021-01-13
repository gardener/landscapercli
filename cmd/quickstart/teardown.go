package quickstart

import (
	"context"
	"fmt"
	"os"

	"github.com/gardener/landscapercli/pkg/logger"
	"github.com/go-logr/logr"
	"github.com/spf13/cobra"
	"github.com/spf13/pflag"
	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/tools/clientcmd"
)

type teardownOptions struct {
	kubeconfigPath string
	namespace      string
}

func NewTeardownCommand(ctx context.Context) *cobra.Command {
	opts := &teardownOptions{}
	cmd := &cobra.Command{
		Use:     "teardown",
		Aliases: []string{"td"},
		Short:   "",
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

func (o *teardownOptions) run(ctx context.Context, log logr.Logger) error {
	cfg, err := clientcmd.BuildConfigFromFlags("", o.kubeconfigPath)
	if err != nil {
		return fmt.Errorf("Cannot parse K8s config: %w", err)
	}

	k8sClient, err := kubernetes.NewForConfig(cfg)
	if err != nil {
		return fmt.Errorf("Cannot build K8s clientset: %w", err)
	}

	_, err = k8sClient.CoreV1().Namespaces().Get(ctx, o.namespace, v1.GetOptions{})
	if err != nil {
		return fmt.Errorf("Cannot get namespace %s: %w", o.namespace, err)
	}

	fmt.Println("Teardown OCI Registry...")
	err = teardownOCIRegistry(ctx, o.namespace, k8sClient)
	if err != nil {
		return fmt.Errorf("Cannot uninstall OCI registry: %w", err)
	}
	fmt.Print("OCI registry teardown succeeded!\n\n")

	fmt.Println("Teardown Landscaper...")
	err = teardownLandscaper(ctx, o.kubeconfigPath, o.namespace)
	if err != nil {
		return fmt.Errorf("Cannot uninstall landscaper: %w", err)
	}
	fmt.Println("Landscaper teardown succeeded!")

	return nil
}

func (o *teardownOptions) Complete(args []string) error {
	return nil
}

func (o *teardownOptions) AddFlags(fs *pflag.FlagSet) {
	fs.StringVar(&o.kubeconfigPath, "kubeconfig", "", "path to the kubeconfig of the target cluster")
	fs.StringVar(&o.namespace, "namespace", defaultNamespace, "namespace where landscaper and OCI registry are installed (default: "+defaultNamespace+")")
}

func teardownOCIRegistry(ctx context.Context, namespace string, k8sClient kubernetes.Interface) error {
	ociRegistry := NewOCIRegistry(namespace, k8sClient)
	return ociRegistry.uninstall(ctx)
}

func teardownLandscaper(ctx context.Context, kubeconfigPath, namespace string) error {
	err := execute(fmt.Sprintf("helm delete --namespace %s landscaper --kubeconfig %s", namespace, kubeconfigPath))
	if err != nil {
		return err
	}

	return nil
}
