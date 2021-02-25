package debug

import (
	"context"
	"os"

	"github.com/gardener/landscapercli/pkg/logger"
	"github.com/go-logr/logr"
	"github.com/spf13/cobra"
	"github.com/spf13/pflag"
)

type debugOpts struct {
}

func NewStatusCommand(ctx context.Context) *cobra.Command {
	opts := &debugOpts{}
	cmd := &cobra.Command{
		Use:     "create [baseURL] [componentName] [componentVersion]",
		Args:    cobra.ExactArgs(3),
		Aliases: []string{"c"},
		Example: "landscaper-cli installations create my-registry:5000 github.com/my-component v0.1.0",
		Short:   "create an installation template for a component which is stored in an OCI registry",
		Run: func(cmd *cobra.Command, args []string) {
			if err := opts.Complete(args); err != nil {
				cmd.PrintErr(err.Error())
				os.Exit(1)
			}

			if err := opts.run(ctx, cmd, logger.Log); err != nil {
				cmd.PrintErr(err.Error())
				os.Exit(1)
			}
		},
	}

	cmd.SetOut(os.Stdout)

	opts.AddFlags(cmd.Flags())

	return cmd
}

func (o *debugOpts) run(ctx context.Context, cmd *cobra.Command, log logr.Logger) error {
	return nil
}

func (o *debugOpts) Complete(args []string) error {
	return nil
}

func (o *debugOpts) AddFlags(fs *pflag.FlagSet) {
}
