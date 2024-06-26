package version

import (
	"fmt"
	"os"

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
			fmt.Printf("\nLandscaper-CLI Version: %s\n", v.GitVersion)

			if v.GitCommit != "" {
				fmt.Printf("  GitCommit: %s\n", v.GitCommit)
			}

			if v.GitTreeState != "" {
				fmt.Printf("  GitTreeState: %s\n", v.GitTreeState)
			}

			if v.GoVersion != "" {
				fmt.Printf("  GoVersion: %s\n", v.GoVersion)
			}

			if v.Compiler != "" {
				fmt.Printf("  Compiler: %s\n", v.Compiler)
			}

			if v.Platform != "" {
				fmt.Printf("  Platform: %s\n", v.Platform)
			}

			release, err := version.GetRelease()
			if err != nil {
				fmt.Println(err.Error())
				os.Exit(1)
			}

			fmt.Printf("\nDefault Landscaper Version: %s", release)
			fmt.Printf("\n\n")
		},
	}
}
