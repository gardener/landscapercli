package targets

import (
	"context"

	"github.com/gardener/landscapercli/cmd/targets/types"
	"github.com/spf13/cobra"
)

func NewCreateCommand(ctx context.Context) *cobra.Command {
	cmd := &cobra.Command{
		Use:     "create",
		Aliases: []string{"c"},
		Short:   "command for creating different types of targets",
	}

	cmd.AddCommand(types.NewKubernetesClusterCommand(ctx))

	return cmd
}
