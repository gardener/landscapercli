package tree

import (
	"fmt"
	"strings"

	lsv1alpha1 "github.com/gardener/landscaper/apis/core/v1alpha1"
)

const (
	newLine      = "\n"
	emptySpace   = "    "
	middleItem   = "├── "
	continueItem = "│   "
	lastItem     = "└── "
)

func PrintTree(installationTree InstallationTree) strings.Builder {
	output := strings.Builder{}
	printInstallation(installationTree, "", &output, true)
	return output
}

func printInstallation(installationTree InstallationTree, preFix string, output *strings.Builder, isLast bool) {
	itemFormat := middleItem
	if isLast {
		itemFormat = lastItem
	}

	fmt.Fprintf(output, "%s%s[%s] Installation %s\n", preFix, itemFormat,
		formatStatus(string(installationTree.Installation.Status.Phase)), installationTree.Installation.Name)
	printErrorIfNecessary(installationTree.Installation.Status.LastError, preFix, output)

	if &installationTree.Execution != nil {
		printExecution(installationTree.Execution, addEmptySpaceOrContinueItem(preFix, isLast), output)
	}

	for i, subInst := range installationTree.SubInstallations {
		printInstallation(subInst, addEmptySpaces(preFix), output, i == len(installationTree.SubInstallations)-1)
	}

}

func printExecution(executionTree ExecutionTree, preFix string, output *strings.Builder) {
	fmt.Fprintf(output, "%s%s[%s] Execution %s\n", preFix, lastItem,
		formatStatus(string(executionTree.Execution.Status.Phase)), executionTree.Execution.Name)

	for i, depItem := range executionTree.DeployItems {
		printDeployitem(depItem, addEmptySpaces(preFix), output, i == len(executionTree.DeployItems)-1)
	}

}

func printDeployitem(deployItemTree DeployItemTree, preFix string, output *strings.Builder, isLast bool) {
	itemFormat := middleItem
	if isLast {
		itemFormat = lastItem
	}

	fmt.Fprintf(output, "%s%s[%s] DeployItem %s\n", preFix, itemFormat,
		formatStatus(string(deployItemTree.DeployItem.Status.Phase)), deployItemTree.DeployItem.Name)
	printErrorIfNecessary(deployItemTree.DeployItem.Status.LastError, addEmptySpaceOrContinueItem(preFix, false), output)
}

func formatEmptySpaces(depth int) string {
	emptySpaces := ""
	for i := 0; i < depth; i++ {
		emptySpaces += emptySpace
	}
	return emptySpaces
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

func addEmptySpaces(s string) string {
	return s + emptySpace
}

func addEmptySpaceOrContinueItem(s string, isLast bool) string {
	if isLast {
		return addEmptySpaces(s)
	} else {
		return s + continueItem
	}
}

func printErrorIfNecessary(err *lsv1alpha1.Error, preFix string, output *strings.Builder) {
	if err != nil {
		fmt.Fprintf(output, "%s%s", preFix, err.Message)
	}
}
