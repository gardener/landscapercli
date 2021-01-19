// SPDX-FileCopyrightText: 2020 SAP SE or an SAP affiliate company and Gardener contributors.
//
// SPDX-License-Identifier: Apache-2.0

package componentarchive

import (
	"context"

	"github.com/spf13/cobra"

	"github.com/gardener/component-cli/pkg/commands/componentarchive/componentreferences"
	"github.com/gardener/component-cli/pkg/commands/componentarchive/resources"
	"github.com/gardener/component-cli/pkg/commands/componentarchive/sources"
)

// NewComponentArchiveCommand creates a new component archive command.
func NewComponentArchiveCommand(ctx context.Context) *cobra.Command {
	cmd := &cobra.Command{
		Use:     "component-archive",
		Aliases: []string{"componentarchive", "ca", "archive"},
	}
	cmd.AddCommand(NewExportCommand(ctx))
	cmd.AddCommand(resources.NewResourcesCommand(ctx))
	cmd.AddCommand(componentreferences.NewCompRefCommand(ctx))
	cmd.AddCommand(sources.NewSourcesCommand(ctx))
	return cmd
}
