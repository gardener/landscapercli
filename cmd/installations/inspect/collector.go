package tree

import (
	"context"
	"fmt"

	lsv1alpha1 "github.com/gardener/landscaper/apis/core/v1alpha1"
	"github.com/gardener/landscaper/pkg/landscaper/installations"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

// Collector is responsible for collecting CR (installations, executions, deployItems) from a cluster using the K8sClient.
type Collector struct {
	K8sClient client.Client
}

// CollectInstallationsInCluster collects a single installation (including all referenced executions and deployitems)
// or all installations if name is empty.
func (c *Collector) CollectInstallationsInCluster(name string, namespace string) ([]*InstallationTree, error) {
	ctx := context.TODO()

	if name != "" {
		installation, err := c.collectInstallationTree(ctx, name, namespace)
		if err != nil {
			return nil, fmt.Errorf("error resolving installation %s: %w", name, err)
		}
		return []*InstallationTree{installation}, nil
	}

	instList := lsv1alpha1.InstallationList{}
	installationTreeList := []*InstallationTree{}

	if namespace == "*" {
		if err := c.K8sClient.List(ctx, &instList); err != nil {
			return nil, fmt.Errorf("cannot list installations across all namespaces: %w", err)
		}
	} else {
		if err := c.K8sClient.List(ctx, &instList, client.InNamespace(namespace)); err != nil {
			return nil, fmt.Errorf("cannot list installations for namespace %s: %w", namespace, err)
		}
	}
	for _, inst := range instList.Items {
		if installations.IsRootInstallation(&inst) {
			filledInst, err := c.collectInstallationTree(ctx, inst.Name, inst.Namespace)
			if err != nil {
				return nil, fmt.Errorf("cannot get installation details for %s: %w", inst.Name, err)
			}
			installationTreeList = append(installationTreeList, filledInst)
		}
	}
	return installationTreeList, nil
}

func (c *Collector) collectInstallationTree(ctx context.Context, name string, namespace string) (*InstallationTree, error) {
	key := client.ObjectKey{Name: name, Namespace: namespace}
	inst := lsv1alpha1.Installation{}
	err := c.K8sClient.Get(ctx, key, &inst)
	if err != nil {
		return nil, fmt.Errorf("cannot get installation %s: %w", name, err)
	}

	tree := InstallationTree{
		Installation: &inst,
	}

	//resolve all sub installations
	subInstList := &lsv1alpha1.InstallationList{}
	err = c.K8sClient.List(ctx, subInstList,
		client.InNamespace(inst.Namespace),
		client.MatchingLabels{
			lsv1alpha1.EncompassedByLabel: inst.Name,
		})
	if err != nil {
		return nil, fmt.Errorf("cannot get subinstallations of %s: %w", inst.Name, err)
	}

	for _, subInst := range subInstList.Items {
		subInstTree, err := c.collectInstallationTree(ctx, subInst.Name, namespace)
		if err != nil {
			return nil, fmt.Errorf("cannot get installation %s: %w", subInst.Name, err)
		}
		tree.SubInstallations = append(tree.SubInstallations, subInstTree)
	}

	//resolve executions
	execution := inst.Status.ExecutionReference
	if execution != nil {
		subExecution, err := c.collectExecutionTree(ctx, execution.Name, execution.Namespace)
		if err != nil {
			return nil, fmt.Errorf("cannot get execution %s: %w", execution.Name, err)
		}
		tree.Execution = subExecution
	}

	return &tree, nil
}

func (c *Collector) collectExecutionTree(ctx context.Context, name string, namespace string) (*ExecutionTree, error) {
	key := client.ObjectKey{Name: name, Namespace: namespace}
	exec := lsv1alpha1.Execution{}
	err := c.K8sClient.Get(ctx, key, &exec)
	if err != nil {
		return nil, fmt.Errorf("cannot get execution %s: %w", name, err)
	}

	tree := ExecutionTree{
		Execution: &exec,
	}

	//resolve deployItems
	deployItemList := &lsv1alpha1.DeployItemList{}
	err = c.K8sClient.List(ctx, deployItemList,
		client.MatchingLabels{lsv1alpha1.ExecutionManagedByLabel: exec.Name},
		client.InNamespace(exec.Namespace))
	if err != nil {
		return nil, fmt.Errorf("cannot list deployitems of execution %s: %w", exec.Name, err)
	}

	for _, deployItem := range deployItemList.Items {
		deployItemTree, err := c.collectDeployItemTree(ctx, deployItem.Name, deployItem.Namespace)
		if err != nil {
			return nil, fmt.Errorf("cannot get deployitem %s: %w", deployItem.Name, err)
		}
		tree.DeployItems = append(tree.DeployItems, deployItemTree)
	}

	return &tree, nil
}

func (c *Collector) collectDeployItemTree(ctx context.Context, name string, namespace string) (*DeployItemLeaf, error) {
	key := client.ObjectKey{Name: name, Namespace: namespace}
	depItem := lsv1alpha1.DeployItem{}
	err := c.K8sClient.Get(ctx, key, &depItem)
	if err != nil {
		return nil, fmt.Errorf("cannot get deployitem %s: %w", name, err)
	}

	tree := DeployItemLeaf{
		DeployItem: &depItem,
	}

	return &tree, nil
}
