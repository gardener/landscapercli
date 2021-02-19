package targets

import (
	"context"

	"github.com/spf13/cobra"
)

func NewTargetsCommand(ctx context.Context) *cobra.Command {
	cmd := &cobra.Command{
		Use:     "targets",
		Aliases: []string{"t", "target"},
		Short:   "commands for interacting with targets",
	}

	cmd.AddCommand(NewCreateCommand(ctx))

	return cmd
}
