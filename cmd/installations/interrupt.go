package installations

import (
	"context"
	"os"

	"github.com/gardener/landscaper/apis/core/v1alpha1"
	"github.com/spf13/cobra"
)

func NewInterruptCommand(ctx context.Context) *cobra.Command {
	opts := &annotationOptions{
		annotationKey:   v1alpha1.OperationAnnotation,
		annotationValue: string(v1alpha1.InterruptOperation),
	}

	cmd := &cobra.Command{
		Use:     "interrupt [installation-name] [--namespace namespace] [--kubeconfig kubeconfig.yaml]",
		Args:    cobra.MaximumNArgs(1),
		Example: "landscaper-cli installations interrupt MY_INSTALLATION --namespace MY_NAMESPACE",
		Short: "Interrupts the processing of an installations and its subobjects. All of these objects with an " +
			"unfinished phase (i.e. a phase which is neither 'Succeeded' nor 'Failed' nor 'DeleteFailed') " +
			"are changed to phase 'Failed'. Note that the command affects only the status of Landscaper objects, " +
			"but does not interrupt a running installation process, for example a helm deployment.",
		Run: func(cmd *cobra.Command, args []string) {
			if err := opts.validateArgs(args); err != nil {
				cmd.PrintErr(err.Error())
				os.Exit(1)
			}

			if err := opts.run(ctx); err != nil {
				cmd.PrintErr(err.Error())
				os.Exit(1)
			}

			cmd.Println("The interrupt annotation was added to the installation")
		},
	}

	cmd.SetOut(os.Stdout)

	opts.AddFlags(cmd.Flags())

	return cmd
}
