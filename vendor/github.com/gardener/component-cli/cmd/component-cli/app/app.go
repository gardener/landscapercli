// SPDX-FileCopyrightText: 2020 SAP SE or an SAP affiliate company and Gardener contributors.
//
// SPDX-License-Identifier: Apache-2.0

//go:generate go run -mod=vendor ../../../hack/generate-docs ../../../docs/reference

package app

import (
	"context"
	"fmt"
	"os"

	"github.com/gardener/component-cli/pkg/commands/componentarchive"
	"github.com/gardener/component-cli/pkg/commands/ctf"
	"github.com/gardener/component-cli/pkg/commands/imagevector"
	"github.com/gardener/component-cli/pkg/logger"
	"github.com/gardener/component-cli/pkg/version"

	"github.com/spf13/cobra"
)

func NewComponentsCliCommand(ctx context.Context) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "components-cli",
		Short: "components cli",
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

	cmd.AddCommand(NewVersionCommand())
	cmd.AddCommand(ctf.NewCTFCommand(ctx))
	cmd.AddCommand(componentarchive.NewComponentArchiveCommand(ctx))
	cmd.AddCommand(imagevector.NewImageVectorCommand(ctx))

	return cmd
}

func NewVersionCommand() *cobra.Command {
	return &cobra.Command{
		Use:     "version",
		Aliases: []string{"v"},
		Short:   "displays the version",
		Run: func(cmd *cobra.Command, args []string) {
			v := version.Get()
			fmt.Printf("%#v", v)
		},
	}
}
