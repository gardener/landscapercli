// SPDX-FileCopyrightText: 2020 SAP SE or an SAP affiliate company and Gardener contributors.
//
// SPDX-License-Identifier: Apache-2.0

package installations

import (
	"context"

	"github.com/spf13/cobra"
)

// NewInstallationsCommand creates a new installations command.
func NewInstallationsCommand(ctx context.Context) *cobra.Command {
	cmd := &cobra.Command{
		Use:     "installations",
		Aliases: []string{"inst", "installation", "i"},
		Short:   "commands to interact with installations",
	}

	cmd.AddCommand(NewCreateCommand(ctx))
	cmd.AddCommand(NewSetImportParametersCommand(ctx))

	return cmd
}
