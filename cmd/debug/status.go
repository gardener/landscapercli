package debug

import (
	"context"
	"fmt"
	"os"

	"github.com/gardener/landscapercli/cmd/debug/tree"

	"github.com/gardener/landscapercli/pkg/logger"
	"github.com/go-logr/logr"
	"github.com/spf13/cobra"
	"github.com/spf13/pflag"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/client-go/tools/clientcmd"
	"sigs.k8s.io/controller-runtime/pkg/client"

	lsv1alpha1 "github.com/gardener/landscaper/apis/core/v1alpha1"
	clientgoscheme "k8s.io/client-go/kubernetes/scheme"
)

type statusOptions struct {
	kubeconfig       string
	installationName string
	namespace        string
}

var (
	scheme = runtime.NewScheme()
)

func init() {
	_ = clientgoscheme.AddToScheme(scheme)
	_ = lsv1alpha1.AddToScheme(scheme)
}

func NewStatusCommand(ctx context.Context) *cobra.Command {
	opts := &statusOptions{}
	cmd := &cobra.Command{
		Use:     "status [installationName] [namespace] --kubeconfig [kubeconfig.yaml]",
		Aliases: []string{"s"},
		Args:    cobra.ExactArgs(2),
		Example: "landscaper-cli debug status my-installation my-namespace --kubeconfig kubeconfig.yaml",
		Short:   "create an installation template for a component which is stored in an OCI registry",
		Run: func(cmd *cobra.Command, args []string) {
			if err := opts.validateArgs(args); err != nil {
				cmd.PrintErr(err.Error())
				os.Exit(1)
			}

			if err := opts.run(ctx, cmd, logger.Log); err != nil {
				cmd.PrintErr(err.Error())
				os.Exit(1)
			}
		},
	}

	cmd.SetOut(os.Stdout)

	opts.AddFlags(cmd.Flags())
	cmd.MarkFlagRequired("kubeconfig")

	return cmd
}

func (o *statusOptions) run(ctx context.Context, cmd *cobra.Command, log logr.Logger) error {
	cfg, err := clientcmd.BuildConfigFromFlags("", o.kubeconfig)
	if err != nil {
		return fmt.Errorf("cannot parse K8s config: %w", err)
	}

	k8sClient, err := client.New(cfg, client.Options{
		Scheme: scheme,
	})
	if err != nil {
		return fmt.Errorf("cannot build K8s client: %w", err)
	}

	coll := tree.Collector{
		K8sClient: k8sClient,
	}
	installationTree, err := coll.CollectInstallationTree(o.installationName, o.namespace)
	if err != nil {
		return fmt.Errorf("cannot collect installation: %w", err)
	}

	output := tree.PrintTree(*installationTree)
	fmt.Print(output.String())

	return nil
}

func (o *statusOptions) validateArgs(args []string) error {
	o.installationName = args[0]
	o.namespace = args[1]

	return nil
}

func (o *statusOptions) AddFlags(fs *pflag.FlagSet) {
	fs.StringVar(&o.kubeconfig, "kubeconfig", "", "path to the kubeconfig of the cluster")
}
