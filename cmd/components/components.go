// SPDX-FileCopyrightText: 2020 SAP SE or an SAP affiliate company and Gardener contributors.
//
// SPDX-License-Identifier: Apache-2.0

package components

import (
	"context"

	"github.com/spf13/cobra"
)

// NewComponentsCommand creates a new components command.
func NewComponentsCommand(ctx context.Context) *cobra.Command {
	cmd := &cobra.Command{
		Use:     "components",
		Aliases: []string{"comp", "component", "co"},
		Short:   "command to interact with components based on blueprints",
	}

	cmd.AddCommand(NewCreateCommand(ctx))
	cmd.AddCommand(NewAddCommand(ctx))

	return cmd
}
