package tree

import (
	"context"
	"fmt"

	lsv1alpha1 "github.com/gardener/landscaper/apis/core/v1alpha1"
	installations "github.com/gardener/landscaper/pkg/landscaper/installations"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

// Collector is responsible for collecting CR (installations, executions, deployItems) from a cluster using the K8sClient.
type Collector struct {
	K8sClient client.Client
}

// CollectInstallationsInCluster collects a single installation (including all referenced executions and deployitems)
// or all installations if name is empty.
func (c *Collector) CollectInstallationsInCluster(name string, namespace string) ([]*InstallationTree, error) {
	if name != "" {
		installation, err := c.collectInstallationTree(name, namespace)
		if err != nil {
			return nil, fmt.Errorf("error resolving installation %s: %w", name, err)
		}
		return []*InstallationTree{installation}, nil
	}

	instList := lsv1alpha1.InstallationList{}
	installationTreeList := []*InstallationTree{}

	if namespace == "*" {
		if err := c.K8sClient.List(context.TODO(), &instList); err != nil {
			return nil, fmt.Errorf("cannot list installations across all namespaces: %w", err)
		}
	} else {
		if err := c.K8sClient.List(context.TODO(), &instList, client.InNamespace(namespace)); err != nil {
			return nil, fmt.Errorf("cannot list installations for namespace %s: %w", namespace, err)
		}
	}
	for _, inst := range instList.Items {
		if installations.IsRootInstallation(&inst) {
			filledInst, err := c.collectInstallationTree(inst.Name, inst.Namespace)
			if err != nil {
				return nil, fmt.Errorf("cannot get installation details for %s: %w", inst.Name, err)
			}
			installationTreeList = append(installationTreeList, filledInst)
		}
	}
	return installationTreeList, nil
}

func (c *Collector) collectInstallationTree(name string, namespace string) (*InstallationTree, error) {
	key := client.ObjectKey{Name: name, Namespace: namespace}
	inst := lsv1alpha1.Installation{}
	err := c.K8sClient.Get(context.TODO(), key, &inst)
	if err != nil {
		return nil, fmt.Errorf("cannot get installation %s: %w", name, err)
	}

	tree := InstallationTree{
		Installation: &inst,
	}

	//resolve all sub installations
	for _, subInst := range inst.Status.InstallationReferences {
		subInstTree, err := c.collectInstallationTree(subInst.Reference.Name, namespace)
		if err != nil {
			return nil, fmt.Errorf("cannot get installation %s: %w", subInst.Reference.Name, err)
		}
		tree.SubInstallations = append(tree.SubInstallations, subInstTree)
	}

	//resolve executions
	execution := inst.Status.ExecutionReference
	if execution != nil {
		subExecution, err := c.collectExecutionTree(execution.Name, execution.Namespace)
		if err != nil {
			return nil, fmt.Errorf("cannot get execution %s: %w", execution.Name, err)
		}
		tree.Execution = subExecution
	}

	return &tree, nil
}

func (c *Collector) collectExecutionTree(name string, namespace string) (*ExecutionTree, error) {
	key := client.ObjectKey{Name: name, Namespace: namespace}
	exec := lsv1alpha1.Execution{}
	err := c.K8sClient.Get(context.TODO(), key, &exec)
	if err != nil {
		return nil, fmt.Errorf("cannot get execution %s: %w", name, err)
	}

	tree := ExecutionTree{
		Execution: &exec,
	}

	//resolve deployItems
	for _, deployItem := range exec.Status.DeployItemReferences {
		deployItemTree, err := c.collectDeployItemTree(deployItem.Reference.Name, deployItem.Reference.Namespace)
		if err != nil {
			return nil, fmt.Errorf("cannot get deployitem %s: %w", deployItem.Reference.Name, err)
		}
		tree.DeployItems = append(tree.DeployItems, deployItemTree)
	}

	return &tree, nil
}

func (c *Collector) collectDeployItemTree(name string, namespace string) (*DeployItemLeaf, error) {
	key := client.ObjectKey{Name: name, Namespace: namespace}
	depItem := lsv1alpha1.DeployItem{}
	err := c.K8sClient.Get(context.TODO(), key, &depItem)
	if err != nil {
		return nil, fmt.Errorf("cannot get deployitem %s: %w", name, err)
	}

	tree := DeployItemLeaf{
		DeployItem: &depItem,
	}

	return &tree, nil
}
