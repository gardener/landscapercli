// SPDX-FileCopyrightText: 2020 SAP SE or an SAP affiliate company and Gardener contributors.
//
// SPDX-License-Identifier: Apache-2.0

package components

import (
	"context"

	"github.com/spf13/cobra"
)

// NewBlueprintsCommand creates a new blueprints command.
func NewAddHelmCommand(ctx context.Context) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "helm-ls",
		Short: "command to add parts to a component concerning a helm landscaper deployment",
	}

	cmd.AddCommand(NewAddHelmLSDeployItemCommand(ctx))

	return cmd
}
