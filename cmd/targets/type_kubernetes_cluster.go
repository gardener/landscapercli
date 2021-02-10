package targets

import (
	"context"
	"fmt"

	"github.com/gardener/landscaper/apis/core"
	"github.com/spf13/cobra"
)

type kubernetesClusterOpts struct {
	kubeconfig bool
}

func NewKubernetesClusterCommand(ctx context.Context) *cobra.Command {
	return &cobra.Command{
		Use:     "kubernetes-cluster",
		Aliases: []string{"k8s"},
		Short:   "create a target of type " + core.GroupName + "/kubernetes-cluster",
		Run: func(cmd *cobra.Command, args []string) {
			fmt.Println("creating target...")
		},
	}
}