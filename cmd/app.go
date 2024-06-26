// SPDX-FileCopyrightText: 2020 SAP SE or an SAP affiliate company and Gardener contributors.
//
// SPDX-License-Identifier: Apache-2.0

package cmd

import (
	"context"
	"fmt"
	"os"

	"github.com/spf13/cobra"

	"github.com/gardener/landscapercli/cmd/blueprints"
	"github.com/gardener/landscapercli/cmd/completion"
	"github.com/gardener/landscapercli/cmd/installations"
	"github.com/gardener/landscapercli/cmd/quickstart"
	"github.com/gardener/landscapercli/cmd/targets"
	"github.com/gardener/landscapercli/cmd/version"
	"github.com/gardener/landscapercli/pkg/logger"
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
	cmd.AddCommand(blueprints.NewBlueprintsCommand(ctx))
	cmd.AddCommand(quickstart.NewQuickstartCommand(ctx))
	cmd.AddCommand(installations.NewInstallationsCommand(ctx))
	cmd.AddCommand(targets.NewTargetsCommand(ctx))
	cmd.AddCommand(completion.NewCompletionCommand())

	return cmd
}
