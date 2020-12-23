package quickstart

import (
	"context"
	"fmt"

	"github.com/spf13/cobra"
)

func NewQuickstartCommand(ctx context.Context) *cobra.Command {
	cmd := &cobra.Command{
		Use:     "quickstart",
		Aliases: []string{"qs"},
		Short:   "",
		Run: func(cmd *cobra.Command, args []string) {
			fmt.Println("quickstart command")
		},
	}

	cmd.AddCommand(NewSetupCommand(ctx))

	return cmd
}