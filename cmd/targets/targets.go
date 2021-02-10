package targets

import (
	"context"

	"github.com/spf13/cobra"
)

func NewTargetsCommand(ctx context.Context) *cobra.Command {
	cmd := &cobra.Command{
		Use:     "targets",
		Aliases: []string{"t"},
		Short:   "commands to interact with targets",
	}

	cmd.AddCommand(NewCreateCommand(ctx))

	return cmd
}