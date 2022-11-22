package types

import (
	"context"
	"os"

	lsv1alpha1 "github.com/gardener/landscaper/apis/core/v1alpha1"
	"github.com/gardener/landscaper/apis/core/v1alpha1/targettypes"
	"github.com/spf13/cobra"
	"github.com/spf13/pflag"

	"github.com/gardener/landscapercli/pkg/util"
)

type kubernetesClusterOpts struct {
	targetKubeconfigPath string
}

func NewKubernetesClusterCommand(ctx context.Context, superOpts *TargetCreateOpts) *cobra.Command {
	opts := &kubernetesClusterOpts{}
	cmd := &cobra.Command{
		Use: "kubernetes-cluster --name [name] --namespace [namespace] " +
			"--target-kubeconfig [path to target kubeconfig]",
		Args:    cobra.NoArgs,
		Aliases: []string{"k8s-cluster"},
		Example: "landscaper-cli targets create kubernetes-cluster --name my-target --namespace my-namespace " +
			"--target-kubeconfig  kubeconfig.yaml",
		Short: "create a target of type " + string(targettypes.KubernetesClusterTargetType),
	}

	cmd.SetOut(os.Stdout)

	opts.AddFlags(cmd.Flags())
	err := cmd.MarkFlagRequired("target-kubeconfig")
	if err != nil {
		panic(err)
	}

	return NewTargetCreateSubcommand(ctx, cmd, opts.run, superOpts)
}

func (o *kubernetesClusterOpts) run(ctx context.Context, opts *TargetCreateOpts) ([]byte, lsv1alpha1.TargetType, error) {
	content, err := util.GetKubernetesClusterTargetContent(o.targetKubeconfigPath)
	return content, targettypes.KubernetesClusterTargetType, err
}

func (o *kubernetesClusterOpts) AddFlags(fs *pflag.FlagSet) {
	fs.StringVar(&o.targetKubeconfigPath, "target-kubeconfig", "", "path to the kubeconfig where the created target object will point to")
}
