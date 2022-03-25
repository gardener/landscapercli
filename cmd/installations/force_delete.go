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

type forceDeleteOptions struct {
	kubeconfig       string
	installationName string
	namespace        string
	k8sClient client.Client
}


func NewForceDeleteCommand(ctx context.Context) *cobra.Command {
	opts := &forceDeleteOptions{}
	cmd := &cobra.Command{
		Use:     "force-delete [installation-name] [--namespace namespace] [--kubeconfig kubeconfig.yaml]",
		Aliases: []string{"fd"},
		Args:    cobra.MaximumNArgs(1),
		Example: "landscaper-cli installations force-delete",
		Short:   "Deletes an installations and the depending executions and deployItems in cluster and namespace of the " +
			"current kubectl cluster context. Concerning the deployed software no guarantees could be given if it is " +
			"uninstalled or not.",
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

func (o *forceDeleteOptions) run(ctx context.Context, cmd *cobra.Command, log logr.Logger) error {
	k8sClient, namespace, err := util.BuildKubeClientFromConfigOrCurrentClusterContext(o.kubeconfig, scheme)
	if err != nil {
		return fmt.Errorf("cannot build k8s client from config or current cluster context: %w", err)
	}
	o.k8sClient = k8sClient

	if namespace != "" && o.namespace == "" {
		o.namespace = namespace
	}

	if o.namespace == "" {
		return fmt.Errorf("namespace was not defined. Use --namespace to specify a namespace")
	}

	if o.installationName == "" {
		return fmt.Errorf("installationName was not defined.")
	}

	collector := inspect.Collector{
		K8sClient: k8sClient,
	}
	installationTrees, err := collector.CollectInstallationsInCluster(o.installationName, o.namespace)
	if err != nil {
		return fmt.Errorf("cannot collect installations etc.: %w", err)
	}

	o.deleteInstallationsTrees(installationTrees)
	return nil
}

func (o *forceDeleteOptions) deleteInstallationsTrees(installationTrees []*inspect.InstallationTree) error {
	for _, installationTree := range installationTrees {
		// delete(installationTree.Installation)
		o.DeleteExecutionTree(installationTree.Execution)
		o.deleteInstallationsTrees(installationTree.SubInstallations)
	}
	return nil
}

func (o *forceDeleteOptions) DeleteExecutionTree(executionTree *inspect.ExecutionTree) error {
	if executionTree.Execution == nil {
		return nil
	}

	// delete(executionTree.Execution)
	for _, di := range executionTree.DeployItems {
		// delete di.DeployItem
	}

	return nil
}

func (o *forceDeleteOptions) validateArgs(args []string) error {
	if len(args) == 1 {
		o.installationName = args[0]
	}
	return nil
}

func (o *forceDeleteOptions) AddFlags(fs *pflag.FlagSet) {
	fs.StringVar(&o.kubeconfig, "kubeconfig", "", "path to the kubeconfig for the cluster. Required if the cluster is not the same as the current-context of kubectl.")
	fs.StringVarP(&o.namespace, "namespace", "n", "", "namespace of the installation. Required if --kubeconfig is used.")
}
