package types

import (
	"context"
	"fmt"
	"os"

	"github.com/gardener/landscaper/apis/core"
	lsv1alpha1 "github.com/gardener/landscaper/apis/core/v1alpha1"
	"github.com/gardener/landscapercli/pkg/logger"
	"github.com/gardener/landscapercli/pkg/util"
	"github.com/go-logr/logr"
	"github.com/spf13/cobra"
	"github.com/spf13/pflag"
	"k8s.io/apimachinery/pkg/runtime"
	clientgoscheme "k8s.io/client-go/kubernetes/scheme"
	"k8s.io/client-go/tools/clientcmd"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

var (
	scheme = runtime.NewScheme()
)

func init() {
	_ = clientgoscheme.AddToScheme(scheme)
	_ = lsv1alpha1.AddToScheme(scheme)
}

type kubernetesClusterOpts struct {
	kubeconfigPath       string
	targetKubeconfigPath string
}

func NewKubernetesClusterCommand(ctx context.Context) *cobra.Command {
	opts := &kubernetesClusterOpts{}
	cmd := &cobra.Command{
		Use:     "kubernetes-cluster",
		Aliases: []string{"k8s"},
		Short:   "create a target of type " + core.GroupName + "/kubernetes-cluster",
		Run: func(cmd *cobra.Command, args []string) {
			if err := opts.Complete(args); err != nil {
				cmd.PrintErr(err.Error())
				os.Exit(1)
			}

			if err := opts.run(ctx, cmd, logger.Log); err != nil {
				cmd.PrintErr(err.Error())
				os.Exit(1)
			}
		},
	}

	opts.AddFlags(cmd.Flags())

	cmd.SetOut(os.Stdout)

	return cmd
}

func (o *kubernetesClusterOpts) Complete(args []string) error {
	return nil
}

func (o *kubernetesClusterOpts) run(ctx context.Context, cmd *cobra.Command, log logr.Logger) error {
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

	target, err := util.BuildKubernetesClusterTarget("target-name", "target-ns", o.targetKubeconfigPath)
	if err != nil {
		return fmt.Errorf("cannot build target object: %w", err)
	}

	err = k8sClient.Create(ctx, target, &client.CreateOptions{})
	if err != nil {
		return fmt.Errorf("cannot create target: %w", err)
	}

	cmd.Println("Target successfully created")

	return nil
}

func (o *kubernetesClusterOpts) AddFlags(fs *pflag.FlagSet) {
	fs.StringVar(&o.kubeconfigPath, "kubeconfig", "", "path to the kubeconfig of the target cluster")
	fs.StringVar(&o.targetKubeconfigPath, "target-kubeconfig", "", "path to the kubeconfig of the target cluster")
}
