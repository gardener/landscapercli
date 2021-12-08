package quickstart

import (
	"context"
	"fmt"
	"os"
	"strings"
	"time"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/util/wait"

	"github.com/go-logr/logr"
	"github.com/spf13/cobra"
	"github.com/spf13/pflag"
	corev1 "k8s.io/api/core/v1"
	k8sErrors "k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/client-go/tools/clientcmd"
	"sigs.k8s.io/controller-runtime/pkg/client"

	"github.com/gardener/landscaper/apis/core/v1alpha1"

	"github.com/gardener/landscapercli/pkg/logger"
	"github.com/gardener/landscapercli/pkg/util"
)

type uninstallOptions struct {
	kubeconfigPath  string
	namespace       string
	deleteNamespace bool
}

func NewUninstallCommand(ctx context.Context) *cobra.Command {
	opts := &uninstallOptions{}
	cmd := &cobra.Command{
		Use:     "uninstall --kubeconfig [kubconfig.yaml] --delete-namespace",
		Aliases: []string{"u"},
		Short:   "command to uninstall Landscaper and OCI registry (from the install command) in a target cluster",
		Example: "landscaper-cli quickstart uninstall --kubeconfig ./kubconfig.yaml --namespace landscaper --delete-namespace",
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
	if err := o.uninstallLandscaper(ctx, k8sClient); err != nil {
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
	fs.BoolVar(&o.deleteNamespace, "delete-namespace", false, "deletes the namespace (otherwise secrets, service accounts etc. of the landscaper installation in the namespace are not removed)")
}

func (o *uninstallOptions) uninstallOCIRegistry(ctx context.Context, k8sClient client.Client) error {
	ociRegistryOpts := &ociRegistryOpts{
		namespace: o.namespace,
	}
	ociRegistry := NewOCIRegistry(ociRegistryOpts, k8sClient)
	return ociRegistry.uninstall(ctx)
}

func (o *uninstallOptions) uninstallLandscaper(ctx context.Context, k8sClient client.Client) error {
	fmt.Println("Removing deployer registrations")
	deployerRegistrations := v1alpha1.DeployerRegistrationList{}
	if err := k8sClient.List(ctx, &deployerRegistrations); err != nil {
		return err
	}

	for _, registration := range deployerRegistrations.Items {
		if err := k8sClient.Delete(ctx, &registration); err != nil {
			return err
		}
	}

	if err := waitForRegistrationsRemoved(ctx, k8sClient); err != nil {
		return err
	}

	err := util.ExecCommandBlocking(fmt.Sprintf("helm delete --namespace %s landscaper --kubeconfig %s", o.namespace, o.kubeconfigPath))
	if err != nil {
		if strings.Contains(err.Error(), "not found") {
			// Ignore error if the release that should be deleted was not found ;)
			fmt.Println("Release not found...Skipping")
		} else {
			return err
		}
	}

	if o.deleteNamespace {
		fmt.Println("Removing namespace")
		namespace := corev1.Namespace{
			ObjectMeta: metav1.ObjectMeta{Name: o.namespace},
		}

		if err := k8sClient.Delete(ctx, &namespace); err != nil {
			return err
		}
	}

	return nil
}

func waitForRegistrationsRemoved(ctx context.Context, k8sClient client.Client) error {
	timeoutCtx, cancel := context.WithTimeout(ctx, 4*time.Minute)
	defer cancel()

	err := wait.PollImmediateUntil(10*time.Second, func() (done bool, err error) {
		fmt.Println("waiting for deployer registrations removed")

		deployerRegistrations := v1alpha1.DeployerRegistrationList{}
		if err := k8sClient.List(ctx, &deployerRegistrations); err != nil {
			return false, err
		}

		if len(deployerRegistrations.Items) == 0 {
			return true, nil
		}

		return false, nil
	}, timeoutCtx.Done())

	if err != nil {
		return fmt.Errorf("error while waiting for deployer registrations being removed: %w", err)
	}

	return nil
}
