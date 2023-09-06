package installations

import (
	"context"
	"fmt"
	"os"
	"time"

	"k8s.io/apimachinery/pkg/util/wait"

	inspect "github.com/gardener/landscapercli/cmd/installations/inspect"

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
			cmd.Println("All objects deleted")
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
		if err := o.deleteObject(ctx, installationTree.Installation, "installation"); err != nil {
			return err
		}

		if err := o.deleteInstallationTrees(ctx, installationTree.SubInstallations); err != nil {
			return err
		}

		if err := o.deleteExecutionTree(ctx, installationTree.Execution); err != nil {
			return err
		}

		if err := o.removeFinalizer(ctx, installationTree.Installation, "installation"); err != nil {
			return err
		}
	}
	return nil
}

func (o *forceDeleteOptions) deleteExecutionTree(ctx context.Context, executionTree *inspect.ExecutionTree) error {
	if executionTree == nil || executionTree.Execution == nil {
		return nil
	}

	if err := o.deleteObject(ctx, executionTree.Execution, "execution"); err != nil {
		return err
	}

	for _, di := range executionTree.DeployItems {
		if err := o.deleteObject(ctx, di.DeployItem, "deployItem"); err != nil {
			return err
		}
		if err := o.removeFinalizer(ctx, di.DeployItem, "deployItem"); err != nil {
			return err
		}
	}

	if err := o.removeFinalizer(ctx, executionTree.Execution, "execution"); err != nil {
		return err
	}

	return nil
}

func (o *forceDeleteOptions) deleteObject(ctx context.Context, object client.Object, objectType string) error {
	if err := o.k8sClient.Delete(ctx, object); err != nil {
		if apierrors.IsNotFound(err) {
			fmt.Printf("- already gone: %s %s\n", objectType, object.GetName())
			return nil
		}
		return fmt.Errorf("cannot delete %s %s: %w", objectType, object.GetName(), err)
	}

	return nil
}

func (o *forceDeleteOptions) removeFinalizer(ctx context.Context, object client.Object, objectType string) error {
	var lastErr error = nil

	if err := wait.PollUntilContextTimeout(ctx, time.Second, 10*time.Second, true, func(ctx context.Context) (done bool, err error) {
		if err := o.k8sClient.Get(ctx, client.ObjectKeyFromObject(object), object); err != nil {
			if apierrors.IsNotFound(err) {
				return true, nil
			}
			lastErr = fmt.Errorf("cannot fetch %s %s: %w", objectType, object.GetName(), err)
			preliminaryMessage := fmt.Sprintf("- cannot fetch %s %s - will retry", objectType, object.GetName())
			fmt.Println(preliminaryMessage)
			return false, nil
		}

		object.SetFinalizers(nil)
		if err := o.k8sClient.Update(ctx, object); err != nil {
			if apierrors.IsNotFound(err) {
				return true, nil
			}
			lastErr = fmt.Errorf("cannot remove finalizers from %s %s: %w", objectType, object.GetName(), err)
			preliminaryMessage := fmt.Sprintf("- cannot remove finalizers from %s %s - will retry", objectType, object.GetName())
			fmt.Println(preliminaryMessage)
			return false, nil
		}

		return true, nil

	}); err != nil {
		return lastErr
	}

	fmt.Printf("- deleted %s %s\n", objectType, object.GetName())
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
