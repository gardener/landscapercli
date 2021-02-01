package version

import (
	"fmt"

	"github.com/gardener/landscapercli/pkg/version"

	"github.com/spf13/cobra"
)

func NewVersionCommand() *cobra.Command {
	return &cobra.Command{
		Use:     "version",
		Aliases: []string{"v"},
		Short:   "displays the version",
		Run: func(cmd *cobra.Command, args []string) {
			v := version.Get()
			fmt.Printf("%#v", v)
			fmt.Printf("\nCompatible Landscaper Version: %s", version.LandscaperVersion)
			fmt.Printf("\nCompatible and included Component-Cli Version: %s", version.ComponentCliVersion)
			fmt.Printf("\n")
		},
	}
}
