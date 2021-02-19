// SPDX-FileCopyrightText: 2020 SAP SE or an SAP affiliate company and Gardener contributors.
//
// SPDX-License-Identifier: Apache-2.0

package components

import (
	"context"

	"github.com/spf13/cobra"
)

// NewBlueprintsCommand creates a new blueprints command.
func NewAddCommand(ctx context.Context) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "add",
		Short: "command to add parts to a component",
	}

	cmd.AddCommand(NewAddHelmCommand(ctx))
	cmd.AddCommand(NewAddManifestCommand(ctx))
	cmd.AddCommand(NewAddContainerCommand(ctx))

	return cmd
}
