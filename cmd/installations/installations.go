// SPDX-FileCopyrightText: 2020 SAP SE or an SAP affiliate company and Gardener contributors.
//
// SPDX-License-Identifier: Apache-2.0

package installations

import (
	"context"

	lsv1alpha1 "github.com/gardener/landscaper/apis/core/v1alpha1"
	"github.com/spf13/cobra"
	"k8s.io/apimachinery/pkg/runtime"
	clientgoscheme "k8s.io/client-go/kubernetes/scheme"
)

var (
	scheme = runtime.NewScheme()
)

func init() {
	_ = clientgoscheme.AddToScheme(scheme)
	_ = lsv1alpha1.AddToScheme(scheme)
}

// NewInstallationsCommand creates a new installations command.
func NewInstallationsCommand(ctx context.Context) *cobra.Command {
	cmd := &cobra.Command{
		Use:     "installations",
		Aliases: []string{"inst", "installation", "i"},
		Short:   "commands to interact with installations",
	}

	cmd.AddCommand(NewInspectCommand(ctx))
	cmd.AddCommand(NewForceDeleteCommand(ctx))
	cmd.AddCommand(NewReconcileCommand(ctx))
	cmd.AddCommand(NewInterruptCommand(ctx))

	return cmd
}
