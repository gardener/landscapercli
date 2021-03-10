package installations

import (
	"context"
	"encoding/json"
	"fmt"
	"os"

	tree "github.com/gardener/landscapercli/cmd/installations/inspect"

	"github.com/gardener/landscapercli/pkg/logger"
	"github.com/go-logr/logr"
	"github.com/spf13/cobra"
	"github.com/spf13/pflag"
	"k8s.io/apimachinery/pkg/runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/yaml"

	lsv1alpha1 "github.com/gardener/landscaper/apis/core/v1alpha1"
	clientgoscheme "k8s.io/client-go/kubernetes/scheme"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/clientcmd"
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

var (
	scheme = runtime.NewScheme()
)

func init() {
	_ = clientgoscheme.AddToScheme(scheme)
	_ = lsv1alpha1.AddToScheme(scheme)
}

func NewInspectCommand(ctx context.Context) *cobra.Command {
	opts := &statusOptions{}
	cmd := &cobra.Command{
		Use:     "inspect [installationName] [--namespace namespace] [--kubeconfig kubeconfig.yaml]",
		Aliases: []string{"i", "status"},
		Args:    cobra.MaximumNArgs(1),
		Example: "landscaper-cli installation inspect my-installation --namespace my-namespace",
		Short:   "displays status installations for the installation and depending executions and deployitems",
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

	coll := tree.Collector{
		K8sClient: k8sClient,
	}
	installationTrees, err := coll.CollectInstallationsInCluster(o.installationName, o.namespace)
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

	transformer := tree.Transformer{
		DetailedMode:   o.detailMode,
		ShowExecutions: o.showExecutions,
		OnlyFailed:     o.showOnlyFailed,
	}

	transformedTree, err := transformer.TransformToPrintableTree(installationTrees)
	if err != nil {
		return fmt.Errorf("error transforming CR to printable tree: %w", err)
	}
	output := tree.PrintTrees(transformedTree)

	cmd.Print(output.String())

	return nil
}

func (o *statusOptions) buildKubeClientFromConfigOrCurrentClusterContext() (client.Client, string, error) {
	var cfg *rest.Config
	var err error
	namespace := ""
	if o.kubeconfig != "" {
		cfg, err = clientcmd.BuildConfigFromFlags("", o.kubeconfig)
		if err != nil {
			return nil, namespace, fmt.Errorf("cannot parse K8s config: %w", err)
		}
	} else {
		rules := clientcmd.NewDefaultClientConfigLoadingRules()
		rules.DefaultClientConfig = &clientcmd.DefaultClientConfig

		overrides := &clientcmd.ConfigOverrides{ClusterDefaults: clientcmd.ClusterDefaults}
		clientConfig := clientcmd.NewNonInteractiveDeferredLoadingClientConfig(rules, overrides)
		cfg, err = clientConfig.ClientConfig()

		if err != nil {
			return nil, namespace, fmt.Errorf("cannot build k8s config %w", err)
		}
		namespace, _, err = clientConfig.Namespace()
		if err != nil {
			return nil, namespace, fmt.Errorf("error extracting namespace from current k8s context. %w", err)
		}

	}

	k8sClient, err := client.New(cfg, client.Options{
		Scheme: scheme,
	})
	if err != nil {
		return nil, namespace, fmt.Errorf("cannot build K8s client: %w", err)
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
	fs.StringVar(&o.kubeconfig, "kubeconfig", "", "path to the kubeconfig of the cluster")
	fs.StringVarP(&o.kubeconfig, "namespace", "n", "", "namespace of the installation. Required if --kubeconfig is used.")
	fs.BoolVarP(&o.detailMode, "show-details", "d", false, "show detailed information about installations, executions and deployitems")
	fs.BoolVarP(&o.showExecutions, "show-executions", "e", false, "show the executions in the tree")
	fs.BoolVarP(&o.showOnlyFailed, "show-failed", "f", false, "show only failed items")
	fs.BoolVarP(&o.oyaml, "oyaml", "y", false, "output in yaml format")
	fs.BoolVarP(&o.ojson, "ojson", "j", false, "output in json format")
}
