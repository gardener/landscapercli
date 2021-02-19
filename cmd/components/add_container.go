// SPDX-FileCopyrightText: 2020 SAP SE or an SAP affiliate company and Gardener contributors.
//
// SPDX-License-Identifier: Apache-2.0

package components

import (
	"context"

	"github.com/spf13/cobra"
)

func NewAddContainerCommand(ctx context.Context) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "container",
		Short: "command to add parts to a component concerning a container deployment",
	}

	cmd.AddCommand(NewAddContainerDeployItemCommand(ctx))

	return cmd
}
