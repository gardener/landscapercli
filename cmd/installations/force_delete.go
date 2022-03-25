package installations

import (
	"context"
	"fmt"
	"os"

	inspect "github.com/gardener/landscapercli/cmd/installations/inspect"

	lsv1alpha1 "github.com/gardener/landscaper/apis/core/v1alpha1"
	"github.com/go-logr/logr"
	"github.com/spf13/cobra"
	"github.com/spf13/pflag"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	"sigs.k8s.io/controller-runtime/pkg/client"

	"github.com/gardener/landscapercli/pkg/logger"
	"github.com/gardener/landscapercli/pkg/util"
)

type forceDeleteOptions struct {
	kubeconfig       string
	installationName string
	namespace        string
	k8sClient        client.Client
}

func NewForceDeleteCommand(ctx context.Context) *cobra.Command {
	opts := &forceDeleteOptions{}
	cmd := &cobra.Command{
		Use:     "force-delete [installation-name] [--namespace namespace] [--kubeconfig kubeconfig.yaml]",
		Aliases: []string{"fd"},
		Args:    cobra.MaximumNArgs(1),
		Example: "landscaper-cli installations force-delete",
		Short: "Deletes an installations and the depending executions and deployItems in cluster and namespace of the " +
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

	return o.deleteInstallationTrees(ctx, installationTrees)
}

func (o *forceDeleteOptions) deleteInstallationTrees(ctx context.Context, installationTrees []*inspect.InstallationTree) error {
	for _, installationTree := range installationTrees {
		if err := o.deleteInstallation(ctx, installationTree.Installation); err != nil {
			return err
		}

		if err := o.deleteInstallationTrees(ctx, installationTree.SubInstallations); err != nil {
			return err
		}

		if err := o.deleteExecutionTree(ctx, installationTree.Execution); err != nil {
			return err
		}
	}
	return nil
}

func (o *forceDeleteOptions) deleteExecutionTree(ctx context.Context, executionTree *inspect.ExecutionTree) error {
	if executionTree.Execution == nil {
		return nil
	}

	if err := o.deleteExecution(ctx, executionTree.Execution); err != nil {
		return err
	}

	for _, di := range executionTree.DeployItems {
		if err := o.deleteDeployItem(ctx, di.DeployItem); err != nil {
			return err
		}
	}

	return nil
}

func (o *forceDeleteOptions) deleteInstallation(ctx context.Context, inst *lsv1alpha1.Installation) error {
	if err := o.k8sClient.Delete(ctx, inst); err != nil {
		if apierrors.IsNotFound(err) {
			return nil
		}
		return fmt.Errorf("cannot delete installation %s: %w", inst.Name, err)
	}

	inst.SetFinalizers(nil)
	if err := o.k8sClient.Update(ctx, inst); err != nil {
		if apierrors.IsNotFound(err) {
			return nil
		}
		return fmt.Errorf("cannot remove finalizers from installation %s: %w", inst.Name, err)
	}

	return nil
}

func (o *forceDeleteOptions) deleteExecution(ctx context.Context, exec *lsv1alpha1.Execution) error {
	if err := o.k8sClient.Delete(ctx, exec); err != nil {
		if apierrors.IsNotFound(err) {
			return nil
		}
		return fmt.Errorf("cannot delete execution %s: %w", exec.Name, err)
	}

	exec.SetFinalizers(nil)
	if err := o.k8sClient.Update(ctx, exec); err != nil {
		if apierrors.IsNotFound(err) {
			return nil
		}
		return fmt.Errorf("cannot remove finalizers from execution %s: %w", exec.Name, err)
	}

	return nil
}

func (o *forceDeleteOptions) deleteDeployItem(ctx context.Context, di *lsv1alpha1.DeployItem) error {
	if err := o.k8sClient.Delete(ctx, di); err != nil {
		if apierrors.IsNotFound(err) {
			return nil
		}
		return fmt.Errorf("cannot delete deploy item %s: %w", di.Name, err)
	}

	di.SetFinalizers(nil)
	if err := o.k8sClient.Update(ctx, di); err != nil {
		if apierrors.IsNotFound(err) {
			return nil
		}
		return fmt.Errorf("cannot remove finalizers from deploy item %s: %w", di.Name, err)
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
