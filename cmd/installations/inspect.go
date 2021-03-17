package installations

import (
	"context"
	"encoding/json"
	"fmt"
	"os"

	inspect "github.com/gardener/landscapercli/cmd/installations/inspect"

	"github.com/go-logr/logr"
	"github.com/spf13/cobra"
	"github.com/spf13/pflag"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/yaml"

	"k8s.io/client-go/tools/clientcmd"

	"github.com/gardener/landscapercli/pkg/logger"
	"github.com/gardener/landscapercli/pkg/util"
)

type statusOptions struct {
	kubeconfig       string
	installationName string
	namespace        string

	detailMode     bool
	showExecutions bool
	showOnlyFailed bool

	oyaml bool
	ojson bool
}

func NewInspectCommand(ctx context.Context) *cobra.Command {
	opts := &statusOptions{}
	cmd := &cobra.Command{
		Use:     "inspect [installation-name] [--namespace namespace] [--kubeconfig kubeconfig.yaml]",
		Aliases: []string{"i", "status"},
		Args:    cobra.MaximumNArgs(1),
		Example: "landscaper-cli installations inspect",
		Short:   "Displays status information for all installations and depending executions and deployItems in cluster and namespace of the current kubectl cluster context. To display only one installation, specify the installation-name.",
		Run: func(cmd *cobra.Command, args []string) {
			if err := opts.validateArgs(args); err != nil {
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

func (o *statusOptions) run(ctx context.Context, cmd *cobra.Command, log logr.Logger) error {
	k8sClient, namespace, err := o.buildKubeClientFromConfigOrCurrentClusterContext()
	if err != nil {
		return fmt.Errorf("cannot build k8s client from config or current cluster context: %w", err)
	}

	if namespace != "" && o.namespace == "" {
		o.namespace = namespace
	}

	if o.namespace == "" {
		return fmt.Errorf("namespace was not defined. Use --namespace to specify a namespace")
	}

	collector := inspect.Collector{
		K8sClient: k8sClient,
	}
	installationTrees, err := collector.CollectInstallationsInCluster(o.installationName, o.namespace)
	if err != nil {
		return fmt.Errorf("cannot collect installation: %w", err)
	}

	if o.oyaml {
		marshaledInstallationTrees, err := yaml.Marshal(installationTrees)
		if err != nil {
			return fmt.Errorf("failed marshaling output to yaml: %w", err)
		}
		cmd.Print(string(marshaledInstallationTrees))
		return nil
	}

	if o.ojson {
		marshaledInstallationTrees, err := json.Marshal(installationTrees)
		if err != nil {
			return fmt.Errorf("failed marshaling output to json: %w", err)
		}
		cmd.Print(string(marshaledInstallationTrees))
		return nil
	}

	transformer := inspect.Transformer{
		DetailedMode:   o.detailMode,
		ShowExecutions: o.showExecutions,
		ShowOnlyFailed: o.showOnlyFailed,
	}

	transformedTrees, err := transformer.TransformToPrintableTrees(installationTrees)
	if err != nil {
		return fmt.Errorf("error transforming CR to printable tree: %w", err)
	}
	output := inspect.PrintTrees(transformedTrees)

	cmd.Print(output.String())

	return nil
}

func (o *statusOptions) buildKubeClientFromConfigOrCurrentClusterContext() (client.Client, string, error) {
	var err error
	namespace := ""
	var k8sClient client.Client
	if o.kubeconfig != "" {
		cfg, err := clientcmd.BuildConfigFromFlags("", o.kubeconfig)
		if err != nil {
			return nil, namespace, fmt.Errorf("cannot parse K8s config: %w", err)
		}
		k8sClient, err = client.New(cfg, client.Options{
			Scheme: scheme,
		})
		if err != nil {
			return nil, namespace, fmt.Errorf("cannot build K8s client: %w", err)
		}
	} else {
		k8sClient, namespace, err = util.GetK8sClientFromCurrentConfiguredCluster()
		if err != nil {
			return nil, namespace, fmt.Errorf("cannot build K8s client from current cluster config: %w", err)
		}
	}

	return k8sClient, namespace, nil
}

func (o *statusOptions) validateArgs(args []string) error {
	if len(args) == 1 {
		o.installationName = args[0]
	}
	return nil
}

func (o *statusOptions) AddFlags(fs *pflag.FlagSet) {
	fs.StringVar(&o.kubeconfig, "kubeconfig", "", "path to the kubeconfig for the cluster. Required if the cluster is not the same as the current-context of kubectl.")
	fs.StringVarP(&o.namespace, "namespace", "n", "", "namespace of the installation. Required if --kubeconfig is used.")
	fs.BoolVarP(&o.detailMode, "show-details", "d", false, "show detailed information about installations, executions and deployitems. Similar to kubectl describe installation installation-name.")
	fs.BoolVarP(&o.showExecutions, "show-executions", "e", false, "show the executions in the tree. By default, the executions are not shown.")
	fs.BoolVarP(&o.showOnlyFailed, "show-failed", "f", false, "show only items that are in phase 'Failed'. It also prints parent elements to the failed items.")
	fs.BoolVarP(&o.oyaml, "oyaml", "y", false, "output in yaml format.")
	fs.BoolVarP(&o.ojson, "ojson", "j", false, "output in json format.")
}
