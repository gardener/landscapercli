package installations

import (
	"context"
	"os"

	"github.com/gardener/landscaper/apis/core/v1alpha1"
	"github.com/spf13/cobra"
)

func NewReconcileCommand(ctx context.Context) *cobra.Command {
	opts := &annotationOptions{
		annotationKey:         v1alpha1.OperationAnnotation,
		annotationValue:       string(v1alpha1.ReconcileOperation),
		rootInstallationsOnly: true,
	}

	cmd := &cobra.Command{
		Use:     "reconcile [installation-name] [--namespace namespace] [--kubeconfig kubeconfig.yaml]",
		Args:    cobra.MaximumNArgs(1),
		Example: "landscaper-cli installations reconcile MY_INSTALLATION --namespace MY_NAMESPACE",
		Short: "Starts a new reconciliation of the specified root installation. If the command is invoked while a " +
			"reconciliation is already running, the new reconciliation is postponed until the current one has " +
			"finished. The command is only supported for root installations.",
		Run: func(cmd *cobra.Command, args []string) {
			if err := opts.validateArgs(args); err != nil {
				cmd.PrintErr(err.Error())
				os.Exit(1)
			}

			if err := opts.run(ctx); err != nil {
				cmd.PrintErr(err.Error())
				os.Exit(1)
			}

			cmd.Println("The reconcile annotation was added to the installation")
		},
	}

	cmd.SetOut(os.Stdout)

	opts.AddFlags(cmd.Flags())

	return cmd
}
