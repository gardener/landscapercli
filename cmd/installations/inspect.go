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
	"sigs.k8s.io/yaml"

	"github.com/gardener/landscapercli/pkg/logger"
	"github.com/gardener/landscapercli/pkg/util"
)

type statusOptions struct {
	kubeconfig       string
	installationName string
	namespace        string
	omode            string

	allNamespaces  bool
	detailMode     bool
	showExecutions bool
	showOnlyFailed bool

	oyaml bool
	ojson bool
	owide bool
}

const (
	OUTPUT_YAML = "yaml"
	OUTPUT_JSON = "json"
	OUTPUT_WIDE = "wide"
)

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
	k8sClient, namespace, err := util.BuildKubeClientFromConfigOrCurrentClusterContext(o.kubeconfig, scheme)
	if err != nil {
		return fmt.Errorf("cannot build k8s client from config or current cluster context: %w", err)
	}

	if namespace != "" && o.namespace == "" {
		o.namespace = namespace
	}

	if o.allNamespaces {
		if len(o.installationName) != 0 {
			return fmt.Errorf("the --all-namespaces option cannot be used when an installation name is provided")
		}
		o.namespace = "*"
	}
	if o.namespace == "" {
		return fmt.Errorf("namespace was not defined. Use --namespace to specify a namespace")
	}

	switch o.omode {
	case OUTPUT_YAML:
		o.oyaml = true
	case OUTPUT_JSON:
		o.ojson = true
	case OUTPUT_WIDE:
		o.owide = true
	case "":
		// this case occurs if the '-o' flag is not set at all
		// We don't need to do anything here, but it shouldn't go into the default
		// case, since that one is used to detect invalid arguments and throws an error.
	default:
		return fmt.Errorf("invalid option for '--output'/'-o' flag: %q", o.omode)
	}

	// verify mode
	if (o.oyaml || o.ojson || o.owide) && !xor(o.oyaml, o.ojson, o.owide) {
		return fmt.Errorf("no more than one output mode may be set: yaml=%v, json=%v, wide=%v", o.oyaml, o.ojson, o.owide)
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
		ShowNamespaces: o.allNamespaces,
		WideMode:       o.owide,
	}

	transformedTrees, err := transformer.TransformToPrintableTrees(installationTrees)
	if err != nil {
		return fmt.Errorf("error transforming CR to printable tree: %w", err)
	}
	output := inspect.PrintTrees(transformedTrees)

	cmd.Print(output.String())

	return nil
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
	fs.BoolVarP(&o.allNamespaces, "all-namespaces", "A", false, "if present, lists installations across all namespaces. No installation name may be given and any given namespace will be ignored.")
	fs.BoolVarP(&o.detailMode, "show-details", "d", false, "show detailed information about installations, executions and deployitems. Similar to kubectl describe installation installation-name.")
	fs.BoolVarP(&o.showExecutions, "show-executions", "e", false, "show the executions in the tree. By default, the executions are not shown.")
	fs.BoolVarP(&o.showOnlyFailed, "show-failed", "f", false, "show only items that are in phase 'Failed'. It also prints parent elements to the failed items.")
	fs.BoolVarP(&o.oyaml, "oyaml", "y", false, "output in yaml format. Equivalent to '-o yaml'.")
	fs.BoolVarP(&o.ojson, "ojson", "j", false, "output in json format. Equivalent to '-o json'.")
	fs.BoolVarP(&o.owide, "owide", "w", false, "output some additional information. Equivalent to '-o wide'.")
	fs.StringVarP(&o.omode, "output", "o", "", fmt.Sprintf("how the output is formatted. Valid values are %s, %s, and %s.", OUTPUT_YAML, OUTPUT_JSON, OUTPUT_WIDE))
}

// xor returns true if exactly one of the given booleans is true
func xor(bools ...bool) bool {
	isTrue := false
	for _, b := range bools {
		if b {
			if isTrue {
				// more than one boolean is true
				return false
			}
			isTrue = true
		}
	}
	return isTrue
}
