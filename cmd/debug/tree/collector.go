package tree

import (
	"context"
	"fmt"

	lsv1alpha1 "github.com/gardener/landscaper/apis/core/v1alpha1"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

type Collector struct {
	K8sClient client.Client
}

func (c *Collector) CollectInstallationTree(name string, namespace string) (*InstallationTree, error) {
	key := client.ObjectKey{Name: name, Namespace: namespace}
	inst := lsv1alpha1.Installation{}
	err := c.K8sClient.Get(context.TODO(), key, &inst)
	if err != nil {
		return nil, fmt.Errorf("cannot get installation %s: %w", name, err)
	}

	tree := InstallationTree{}
	tree.Installation = inst

	//resolve all sub instalaltions
	for _, subInst := range inst.Status.InstallationReferences {
		subInstTree, err := c.CollectInstallationTree(subInst.Name, namespace)
		if err != nil {
			return nil, fmt.Errorf("cannot get installation %s: %w", subInst.Name, err)
		}
		tree.SubInstallations = append(tree.SubInstallations, *subInstTree)
	}

	//resolve executions
	execution := inst.Status.ExecutionReference
	if execution != nil {
		subExecution, err := c.collectExecutionTree(execution.Name, execution.Namespace)
		if err != nil {
			return nil, fmt.Errorf("cannot get execution %s: %w", execution.Name, err)
		}
		tree.Execution = *subExecution
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

	tree := ExecutionTree{}
	tree.Execution = exec

	//resolve deployItems
	for _, deployItem := range exec.Status.DeployItemReferences {
		deployItemTree, err := c.collectDeployItemTree(deployItem.Reference.Name, deployItem.Reference.Namespace)
		if err != nil {
			return nil, fmt.Errorf("cannot get deployitem %s: %w", deployItem.Reference.Name, err)
		}
		tree.DeployItems = append(tree.DeployItems, *deployItemTree)
	}

	return &tree, nil
}

func (c *Collector) collectDeployItemTree(name string, namespace string) (*DeployItemTree, error) {
	key := client.ObjectKey{Name: name, Namespace: namespace}
	depItem := lsv1alpha1.DeployItem{}
	err := c.K8sClient.Get(context.TODO(), key, &depItem)
	if err != nil {
		return nil, fmt.Errorf("cannot get deployitem %s: %w", name, err)
	}

	tree := DeployItemTree{}
	tree.DeployItem = depItem

	return &tree, nil
}
