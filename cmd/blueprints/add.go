// SPDX-FileCopyrightText: 2020 SAP SE or an SAP affiliate company and Gardener contributors.
//
// SPDX-License-Identifier: Apache-2.0

package blueprints

import (
	"context"

	"github.com/spf13/cobra"
)

// NewBlueprintsCommand creates a new blueprints command.
func NewAddCommand(ctx context.Context) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "add",
		Short: "command to add parts to a blueprint",
	}

	cmd.AddCommand(NewAddExecutionCommand(ctx))

	return cmd
}
