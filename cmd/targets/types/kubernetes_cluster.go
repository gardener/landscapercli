package types

import (
	"context"
	"errors"
	"fmt"
	"os"

	lsv1alpha1 "github.com/gardener/landscaper/apis/core/v1alpha1"
	"github.com/go-logr/logr"
	"github.com/spf13/cobra"
	"github.com/spf13/pflag"
	"sigs.k8s.io/yaml"

	"github.com/gardener/landscapercli/pkg/logger"
	"github.com/gardener/landscapercli/pkg/util"
)

type kubernetesClusterOpts struct {
	name                 string
	namespace            string
	targetKubeconfigPath string

	//outputPath is the path to write the installation.yaml to
	outputPath string
}

func NewKubernetesClusterCommand(ctx context.Context) *cobra.Command {
	opts := &kubernetesClusterOpts{}
	cmd := &cobra.Command{
		Use: "kubernetes-cluster --name [name] --namespace [namespace] " +
			"--target-kubeconfig [path to target kubeconfig]",
		Args:    cobra.NoArgs,
		Aliases: []string{"k8s-cluster"},
		Example: "landscaper-cli targets create kubernetes-cluster --name my-target --namespace my-namespace " +
			"--target-kubeconfig  kubeconfig.yaml",
		Short: "create a target of type " + string(lsv1alpha1.KubernetesClusterTargetType),
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

func (o *kubernetesClusterOpts) Complete(args []string) error {
	if o.name == "" {
		return errors.New("--name must be defined")
	}

	if o.targetKubeconfigPath == "" {
		return errors.New("--target-kubeconfig must be defined")
	}

	return nil
}

func (o *kubernetesClusterOpts) run(ctx context.Context, cmd *cobra.Command, log logr.Logger) error {
	target, err := util.BuildKubernetesClusterTarget(o.name, o.namespace, o.targetKubeconfigPath)
	if err != nil {
		return fmt.Errorf("cannot build target object: %w", err)
	}

	marshaledYaml, err := yaml.Marshal(target)
	if err != nil {
		return fmt.Errorf("cannot marshal target yaml: %w", err)
	}

	if o.outputPath == "" {
		cmd.Println(string(marshaledYaml))
	} else {
		f, err := os.Create(o.outputPath)
		if err != nil {
			return fmt.Errorf("error creating file %s: %w", o.outputPath, err)
		}
		_, err = f.Write(marshaledYaml)
		if err != nil {
			return fmt.Errorf("error writing file %s: %w", o.outputPath, err)
		}
		cmd.Printf("Wrote target to %s", o.outputPath)
	}
	return nil
}

func (o *kubernetesClusterOpts) AddFlags(fs *pflag.FlagSet) {
	fs.StringVar(&o.name, "name", "", "name of the target")
	fs.StringVar(&o.namespace, "namespace", "", "namespace of the target (optional)")
	fs.StringVar(&o.targetKubeconfigPath, "target-kubeconfig", "", "path to the kubeconfig where the created target object will point to")
	fs.StringVarP(&o.outputPath, "output-file", "o", "", "file path for the resulting target yaml")
}
