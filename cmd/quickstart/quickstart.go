package quickstart

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

func NewQuickstartCommand(ctx context.Context) *cobra.Command {
	cmd := &cobra.Command{
		Use:     "quickstart",
		Aliases: []string{"qs"},
		Short:   "useful commands for getting quickly up and running with Landscaper",
	}

	cmd.AddCommand(NewInstallCommand(ctx))
	cmd.AddCommand(NewUninstallCommand(ctx))

	return cmd
}
