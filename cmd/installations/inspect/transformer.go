package tree

import (
	"fmt"

	lsv1alpha1 "github.com/gardener/landscaper/apis/core/v1alpha1"
	"sigs.k8s.io/yaml"
)

//Transformer can transform from []*InstallationTree to a printable []TreeElement with different transformation options.
type Transformer struct {
	DetailedMode   bool
	ShowExecutions bool
	OnlyFailed     bool
}

//TransformToPrintableTree transform a []*InstallationTree to a []TreeElement for the Printer.
func (t Transformer) TransformToPrintableTree(installationTrees []*InstallationTree) ([]TreeElement, error) {
	var printableTrees []TreeElement

	if t.OnlyFailed {
		installationTrees = FilterForFailedItemsInTree(installationTrees)
	}

	for _, installationTree := range installationTrees {
		transformedInstalaltion, err := t.transformInstallation(installationTree)
		if err != nil {
			return nil, fmt.Errorf("error in installation %s: %w", installationTree.Installation.Name, err)
		}

		printableTrees = append(printableTrees, *transformedInstalaltion)
	}
	return printableTrees, nil
}

func (t Transformer) transformInstallation(installationTree *InstallationTree) (*TreeElement, error) {
	printableNode := TreeElement{}
	if installationTree == nil {
		return &printableNode, nil
	}

	installationTree.Installation.SetManagedFields(nil)

	printableNode.Headline = fmt.Sprintf("[%s] Installation %s",
		formatStatus(string(installationTree.Installation.Status.Phase)), installationTree.Installation.Name)

	if t.DetailedMode {
		marshaledInstallation, err := yaml.Marshal(installationTree.Installation)
		if err != nil {
			return nil, fmt.Errorf("failed marshaling installation %s: %w", installationTree.Installation.Name, err)
		}
		printableNode.Description = string(marshaledInstallation)

	} else if installationTree.Installation.Status.LastError != nil {
		printableNode.Description = fmt.Sprintf("Last error: %s", installationTree.Installation.Status.LastError.Message)
	}

	if len(installationTree.SubInstallations) > 0 {
		for _, inst := range installationTree.SubInstallations {
			transformedInstalaltion, err := t.transformInstallation(inst)
			if err != nil {
				return nil, fmt.Errorf("error in subinstallation %s: %w", installationTree.Installation.Name, err)
			}
			printableNode.Childs = append(printableNode.Childs, transformedInstalaltion)
		}
	}

	if installationTree.Execution != nil {
		transformedExecution, err := t.transformExecution(installationTree.Execution)
		if err != nil {
			return nil, fmt.Errorf("error in installation %s: %w", installationTree.Installation.Name, err)
		}
		printableNode.Childs = append(printableNode.Childs, transformedExecution)

	}

	return &printableNode, nil
}

func (t Transformer) transformExecution(executionTree *ExecutionTree) (*TreeElement, error) {
	printableNode := TreeElement{}
	if executionTree == nil {
		return &printableNode, nil
	}

	executionTree.Execution.SetManagedFields(nil)

	if t.ShowExecutions || t.DetailedMode {
		printableNode.Headline = fmt.Sprintf("[%s] Execution %s",
			formatStatus(string(executionTree.Execution.Status.Phase)), executionTree.Execution.Name)
	}

	if t.DetailedMode {
		marshaledExecution, err := yaml.Marshal(executionTree.Execution)
		if err != nil {
			return nil, fmt.Errorf("failed marshaling Execution %s: %w", executionTree.Execution.Name, err)
		}
		printableNode.Description = string(marshaledExecution)
	}

	if len(executionTree.DeployItems) > 0 {
		for _, depItem := range executionTree.DeployItems {
			transformedDeployItem, err := t.transformDeployItem(depItem)
			if err != nil {
				return nil, fmt.Errorf("error in Execution %s: %w", executionTree.Execution.Name, err)
			}
			printableNode.Childs = append(printableNode.Childs, transformedDeployItem)
		}
	}

	return &printableNode, nil
}

func (t Transformer) transformDeployItem(deployItemTree *DeployItemLeaf) (*TreeElement, error) {
	printableNode := TreeElement{}
	if deployItemTree == nil {
		return &printableNode, nil
	}

	deployItemTree.DeployItem.SetManagedFields(nil)

	printableNode.Headline = fmt.Sprintf("[%s] DeployItem %s",
		formatStatus(string(deployItemTree.DeployItem.Status.Phase)), deployItemTree.DeployItem.Name)

	if t.DetailedMode {
		marshaledExecution, err := yaml.Marshal(deployItemTree.DeployItem)
		if err != nil {
			return nil, fmt.Errorf("failed marshaling DeployItem %s: %w", deployItemTree.DeployItem.Name, err)
		}
		printableNode.Description = string(marshaledExecution)
	} else if deployItemTree.DeployItem.Status.LastError != nil {
		printableNode.Description = fmt.Sprintf("Last error: %s", deployItemTree.DeployItem.Status.LastError.Message)
	}

	return &printableNode, nil
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
