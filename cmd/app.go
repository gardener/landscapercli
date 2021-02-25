// SPDX-FileCopyrightText: 2020 SAP SE or an SAP affiliate company and Gardener contributors.
//
// SPDX-License-Identifier: Apache-2.0

//go:generate go run -mod=vendor ../hack/generate-docs ../docs/reference

package cmd

import (
	"context"
	"fmt"
	"os"

	"github.com/gardener/landscapercli/cmd/blueprints"
	"github.com/gardener/landscapercli/cmd/components"
	"github.com/gardener/landscapercli/cmd/debug"
	"github.com/gardener/landscapercli/cmd/installations"
	"github.com/gardener/landscapercli/cmd/quickstart"
	"github.com/gardener/landscapercli/cmd/targets"
	"github.com/gardener/landscapercli/cmd/version"
	"github.com/gardener/landscapercli/pkg/logger"

	componentcli "github.com/gardener/component-cli/cmd/component-cli/app"
	"github.com/spf13/cobra"
)

func NewLandscaperCliCommand(ctx context.Context) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "landscaper-cli",
		Short: "landscaper cli",
		PersistentPreRun: func(cmd *cobra.Command, args []string) {
			log, err := logger.NewCliLogger()
			if err != nil {
				fmt.Println("unable to setup logger")
				fmt.Println(err.Error())
				os.Exit(1)
			}
			logger.SetLogger(log)
		},
	}

	logger.InitFlags(cmd.PersistentFlags())

	cmd.AddCommand(version.NewVersionCommand())
	cmd.AddCommand(components.NewComponentsCommand(ctx))
	cmd.AddCommand(blueprints.NewBlueprintsCommand(ctx))
	cmd.AddCommand(quickstart.NewQuickstartCommand(ctx))
	cmd.AddCommand(installations.NewInstallationsCommand(ctx))
	cmd.AddCommand(targets.NewTargetsCommand(ctx))
	cmd.AddCommand(debug.NewDebugCommand(ctx))

	// Integrate commands of the component cli
	componentsCliCommand := componentcli.NewComponentsCliCommand(ctx)
	componentsCliCommand.Short = "commands of the components cli"
	cmd.AddCommand(componentsCliCommand)

	return cmd
}
