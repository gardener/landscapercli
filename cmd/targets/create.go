package targets

import (
	"context"

	"github.com/spf13/cobra"

	"github.com/gardener/landscapercli/cmd/targets/types"
)

func NewCreateCommand(ctx context.Context) *cobra.Command {
	opts := &types.TargetCreateOpts{}
	cmd := &cobra.Command{
		Use:     "create",
		Aliases: []string{"c"},
		Short:   "command for creating different types of targets",
	}
	opts.AddFlags(cmd.PersistentFlags())
	err := cmd.MarkPersistentFlagRequired("name")
	if err != nil {
		panic(err)
	}

	cmd.AddCommand(types.NewKubernetesClusterCommand(ctx, opts))

	return cmd
}
