package targets

import (
	"context"

	"github.com/gardener/landscapercli/cmd/targets/types"
	"github.com/spf13/cobra"
)

type createOpts struct {
	name      string
	namespace string
}

func NewCreateCommand(ctx context.Context) *cobra.Command {
	cmd := &cobra.Command{
		Use:     "create",
		Aliases: []string{"c"},
		Short:   "create a target",
	}

	cmd.AddCommand(types.NewKubernetesClusterCommand(ctx))

	return cmd
}
