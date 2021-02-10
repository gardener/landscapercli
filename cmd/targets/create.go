package targets

import (
	"context"
	"fmt"

	"github.com/spf13/cobra"
)

func NewCreateCommand(ctx context.Context) *cobra.Command {
	return &cobra.Command{
		Use:     "create",
		Aliases: []string{"c"},
		Short:   "create a target",
		Run: func(cmd *cobra.Command, args []string) {
			fmt.Println("creating target...")
		},
	}
}
