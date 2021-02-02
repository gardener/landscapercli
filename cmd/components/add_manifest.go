// SPDX-FileCopyrightText: 2020 SAP SE or an SAP affiliate company and Gardener contributors.
//
// SPDX-License-Identifier: Apache-2.0

package components

import (
	"context"

	"github.com/spf13/cobra"
)

// NewBlueprintsCommand creates a new blueprints command.
func NewAddManifestCommand(ctx context.Context) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "manifest",
		Short: "command to add parts to a component concerning a manifest deployment",
	}

	cmd.AddCommand(NewAddManifestDeployItemCommand(ctx))

	return cmd
}
