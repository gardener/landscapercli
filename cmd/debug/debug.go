// SPDX-FileCopyrightText: 2020 SAP SE or an SAP affiliate company and Gardener contributors.
//
// SPDX-License-Identifier: Apache-2.0

package debug

import (
	"context"

	"github.com/spf13/cobra"
)

// NewDebugCommand creates a new debug command.
func NewDebugCommand(ctx context.Context) *cobra.Command {
	cmd := &cobra.Command{
		Use:     "debug",
		Aliases: []string{"deb", "b"},
		Short:   "commands to debug landscaper crs",
	}

	cmd.AddCommand(NewStatusCommand(ctx))
	return cmd
}
