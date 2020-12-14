// SPDX-FileCopyrightText: 2020 SAP SE or an SAP affiliate company and Gardener contributors.
//
// SPDX-License-Identifier: Apache-2.0

package cmd

import (
	"context"
	"fmt"
	"os"

	"github.com/gardener/landscaper/pkg/logger"
	"github.com/gardener/landscapercli/cmd/blueprints"
	"github.com/gardener/landscapercli/cmd/componentdescriptor"
	"github.com/gardener/landscapercli/cmd/version"

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
	cmd.AddCommand(blueprints.NewBlueprintsCommand(ctx))
	cmd.AddCommand(componentdescriptor.NewComponentsCommand(ctx))

	return cmd
}