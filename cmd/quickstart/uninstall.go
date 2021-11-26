package quickstart

import (
	"context"
	"fmt"
	"os"
	"strings"

	"github.com/go-logr/logr"
	"github.com/spf13/cobra"
	"github.com/spf13/pflag"
	corev1 "k8s.io/api/core/v1"
	k8sErrors "k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/client-go/tools/clientcmd"
	"sigs.k8s.io/controller-runtime/pkg/client"

	"github.com/gardener/landscapercli/pkg/logger"
	"github.com/gardener/landscapercli/pkg/util"
)

type uninstallOptions struct {
	kubeconfigPath string
	namespace      string
}

func NewUninstallCommand(ctx context.Context) *cobra.Command {
	opts := &uninstallOptions{}
	cmd := &cobra.Command{
		Use:     "uninstall --kubeconfig [kubconfig.yaml]",
		Aliases: []string{"u"},
		Short:   "command to uninstall Landscaper and OCI registry (from the install command) in a target cluster",
		Example: "landscaper-cli quickstart uninstall --kubeconfig ./kubconfig.yaml --namespace landscaper",
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

func (o *uninstallOptions) run(ctx context.Context, log logr.Logger) error {
	cfg, err := clientcmd.BuildConfigFromFlags("", o.kubeconfigPath)
	if err != nil {
		return fmt.Errorf("cannot parse K8s config: %w", err)
	}

	k8sClient, err := client.New(cfg, client.Options{
		Scheme: scheme,
	})
	if err != nil {
		return fmt.Errorf("cannot build K8s client: %w", err)
	}

	key := client.ObjectKey{
		Name: o.namespace,
	}
	ns := corev1.Namespace{}
	if err := k8sClient.Get(ctx, key, &ns); err != nil {
		if k8sErrors.IsNotFound(err) {
			fmt.Printf("Cannot find namespace %s\n", o.namespace)
			return nil
		}
		return fmt.Errorf("cannot get namespace: %w", err)
	}

	fmt.Println("Uninstall OCI Registry")
	if err := o.uninstallOCIRegistry(ctx, k8sClient); err != nil {
		return fmt.Errorf("cannot uninstall OCI registry: %w", err)
	}
	fmt.Print("OCI registry uninstall succeeded!\n\n")

	fmt.Println("Uninstall Landscaper")
	if err := o.uninstallLandscaper(ctx); err != nil {
		return fmt.Errorf("cannot uninstall landscaper: %w", err)
	}
	fmt.Println("Landscaper uninstall succeeded!")

	return nil
}

func (o *uninstallOptions) Complete(args []string) error {
	return nil
}

func (o *uninstallOptions) AddFlags(fs *pflag.FlagSet) {
	fs.StringVar(&o.kubeconfigPath, "kubeconfig", "", "path to the kubeconfig of the target cluster")
	fs.StringVar(&o.namespace, "namespace", defaultNamespace, "namespace where Landscaper and the OCI registry are installed")
}

func (o *uninstallOptions) uninstallOCIRegistry(ctx context.Context, k8sClient client.Client) error {
	ociRegistryOpts := &ociRegistryOpts{
		namespace: o.namespace,
	}
	ociRegistry := NewOCIRegistry(ociRegistryOpts, k8sClient)
	return ociRegistry.uninstall(ctx)
}

func (o *uninstallOptions) uninstallLandscaper(ctx context.Context) error {
	err := util.ExecCommandBlocking(fmt.Sprintf("helm delete --namespace %s landscaper --kubeconfig %s", o.namespace, o.kubeconfigPath))
	if err != nil {
		if strings.Contains(err.Error(), "not found") {
			// Ignore error if the release that should be deleted was not found ;)
			fmt.Println("Release not found...Skipping")
			return nil
		}
		return err
	}

	return nil
}
