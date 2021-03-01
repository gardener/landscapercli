package tree

import (
	"fmt"

	lsv1alpha1 "github.com/gardener/landscaper/apis/core/v1alpha1"
)

func TransformToPrintableTree(installationTrees []InstallationTree) []TreeElement {
	var printableTrees []TreeElement

	for _, installationTree := range installationTrees {
		printableTrees = append(printableTrees, transformInstallation(installationTree))
	}
	return printableTrees
}

func transformInstallation(installationTree InstallationTree) TreeElement {
	printableNode := TreeElement{}

	printableNode.Headline = fmt.Sprintf("[%s] Installation %s\n",
		formatStatus(string(installationTree.Installation.Status.Phase)), installationTree.Installation.Name)

	if installationTree.Installation.Status.LastError != nil {
		printableNode.Description = fmt.Sprintf("Last error: %s", installationTree.Installation.Status.LastError.Message)
	}

	if len(installationTree.SubInstallations) > 0 {
		for _, inst := range installationTree.SubInstallations {
			printableNode.Childs = append(printableNode.Childs, transformInstallation(inst))
		}
	}

	if installationTree.Execution != nil {
		printableNode.Childs = append(printableNode.Childs, transformExecution(*installationTree.Execution))

	}

	return printableNode
}

func transformExecution(executionTree ExecutionTree) TreeElement {
	printableNode := TreeElement{}

	printableNode.Headline = fmt.Sprintf("[%s] Execution %s\n",
		formatStatus(string(executionTree.Execution.Status.Phase)), executionTree.Execution.Name)

	if len(executionTree.DeployItems) > 0 {
		for _, depItem := range executionTree.DeployItems {
			printableNode.Childs = append(printableNode.Childs, transformDeployItem(depItem))
		}
	}

	return printableNode
}

func transformDeployItem(deployItemTree DeployItemTree) TreeElement {
	printableNode := TreeElement{}

	printableNode.Headline = fmt.Sprintf("[%s] DeployItem %s\n",
		formatStatus(string(deployItemTree.DeployItem.Status.Phase)), deployItemTree.DeployItem.Name)

	if deployItemTree.DeployItem.Status.LastError != nil {
		printableNode.Description = fmt.Sprintf("Last error: %s", deployItemTree.DeployItem.Status.LastError.Message)
	}

	return printableNode
}

func formatStatus(status string) string {
	if status == string(lsv1alpha1.ComponentPhaseSucceeded) {
		return "✅ " + status
	}
	if status == string(lsv1alpha1.ComponentPhaseProgressing) {
		return "🏗️ " + status
	}
	if status == string(lsv1alpha1.ComponentPhaseFailed) {
		return "❌ " + status
	}
	return status
}
