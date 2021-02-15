package quickstart

import (
	"context"

	"github.com/spf13/cobra"
)

func NewQuickstartCommand(ctx context.Context) *cobra.Command {
	cmd := &cobra.Command{
		Use:     "quickstart",
		Aliases: []string{"qs"},
		Short:   "useful commands for getting quickly up and running with the landscaper",
	}

	cmd.AddCommand(NewInstallCommand(ctx))
	cmd.AddCommand(NewAddOciEndpointCommand(ctx))
	cmd.AddCommand(NewUninstallCommand(ctx))

	return cmd
}
