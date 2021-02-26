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
	printInstallation(installationTree, 0, &output)
	return output
}

func printInstallation(installationTree InstallationTree, depth int, output *strings.Builder) {
	itemFormat := middleItem
	if len(installationTree.SubInstallations) == 0 {
		itemFormat = lastItem
	}

	fmt.Fprintf(output, "%s%s[%s] Installation %s\n", formatEmptySpaces(depth), itemFormat,
		formatStatus(string(installationTree.Installation.Status.Phase)), installationTree.Installation.Name)
	printErrorIfNecessary(installationTree.Installation.Status.LastError, depth, output)

	for _, subInst := range installationTree.SubInstallations {
		printInstallation(subInst, depth+1, output)
	}

	if &installationTree.Execution != nil {
		printExecution(installationTree.Execution, depth+1, output)
	}

}

func printExecution(executionTree ExecutionTree, depth int, output *strings.Builder) {
	fmt.Fprintf(output, "%s%s[%s] Execution %s\n", formatEmptySpaces(depth), lastItem,
		formatStatus(string(executionTree.Execution.Status.Phase)), executionTree.Execution.Name)

	for i, depItem := range executionTree.DeployItems {
		printDeployitem(depItem, depth+1, output, i == len(executionTree.DeployItems)-1)
	}

}

func printDeployitem(deployItemTree DeployItemTree, depth int, output *strings.Builder, isLast bool) {
	itemFormat := middleItem
	if isLast {
		itemFormat = lastItem
	}

	fmt.Fprintf(output, "%s%s[%s] DeployItem %s\n", formatEmptySpaces(depth), itemFormat,
		formatStatus(string(deployItemTree.DeployItem.Status.Phase)), deployItemTree.DeployItem.Name)
	printErrorIfNecessary(deployItemTree.DeployItem.Status.LastError, depth, output)
}

func formatEmptySpaces(depth int) string {
	emptySpaces := ""
	for i := 0; i < depth; i++ {
		emptySpaces += emptySpace
	}
	return emptySpaces
}

func formatStatus(status string) string {
	return status
}

func printErrorIfNecessary(err *lsv1alpha1.Error, depth int, output *strings.Builder) {
	if err != nil {
		fmt.Fprintf(output, "%s    |%s", formatEmptySpaces(depth), err.Message)
	}
}
