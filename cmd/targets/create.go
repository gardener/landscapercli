package targets

import (
	"context"

	"github.com/spf13/cobra"

	"github.com/gardener/landscapercli/cmd/targets/types"
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
