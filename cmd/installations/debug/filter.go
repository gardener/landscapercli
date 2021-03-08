package tree

import "github.com/gardener/landscaper/apis/core/v1alpha1"

func FilterForFailedItemsInTree(installationTrees []*InstallationTree) []*InstallationTree {
	filteredInstallationTree := []*InstallationTree{}
	for _, installation := range installationTrees {
		if filteredInstalaltion := filterForFailedInstallation(installation); filteredInstalaltion != nil {
			filteredInstallationTree = append(filteredInstallationTree, filteredInstalaltion)
		}
	}
	return filteredInstallationTree
}

func filterForFailedInstallation(installationTree *InstallationTree) *InstallationTree {
	filteredSubInstallations := []*InstallationTree{}
	for _, subInstallation := range installationTree.SubInstallations {
		if filteredSubInstallation := filterForFailedInstallation(subInstallation); filteredSubInstallation != nil {
			filteredSubInstallations = append(filteredSubInstallations, filteredSubInstallation)
		}
	}
	installationTree.SubInstallations = filteredSubInstallations
	installationTree.Execution = filterFailedExecutions(installationTree.Execution)

	if len(filteredSubInstallations) > 0 || installationTree.Execution != nil || installationTree.Installation.Status.Phase == v1alpha1.ComponentPhaseFailed {
		return installationTree
	}
	return nil

}

func filterFailedExecutions(executionTree *ExecutionTree) *ExecutionTree {
	filteredDeployItems := []*DeployItemLeaf{}
	for _, depItem := range executionTree.DeployItems {
		if filteredDeployItem := filterFailedDeployItems(depItem); filteredDeployItem != nil {
			filteredDeployItems = append(filteredDeployItems, filteredDeployItem)
		}
	}

	executionTree.DeployItems = filteredDeployItems

	if len(filteredDeployItems) > 0 || executionTree.Execution.Status.Phase == v1alpha1.ExecutionPhaseFailed {
		return executionTree
	}
	return nil
}

func filterFailedDeployItems(deployItemTree *DeployItemLeaf) *DeployItemLeaf {
	if deployItemTree.DeployItem.Status.Phase == v1alpha1.ExecutionPhaseFailed {
		return deployItemTree
	}
	return nil
}
