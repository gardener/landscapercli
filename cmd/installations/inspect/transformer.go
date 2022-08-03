package tree

import (
	"encoding/json"
	"fmt"
	"strings"

	lsv1alpha1 "github.com/gardener/landscaper/apis/core/v1alpha1"
	containerv1alpha1 "github.com/gardener/landscaper/apis/deployer/container/v1alpha1"
	helmv1alpha1 "github.com/gardener/landscaper/apis/deployer/helm/v1alpha1"
	"sigs.k8s.io/yaml"
)

//Transformer can transform from []*InstallationTree to printable []PrintableTreeNodes with different transformation options.
type Transformer struct {
	DetailedMode   bool
	ShowExecutions bool
	ShowOnlyFailed bool
	ShowNamespaces bool
	WideMode       bool
}

//TransformToPrintableTrees transform a []*InstallationTree to []PrintableTreeNodes for the Printer.
func (t Transformer) TransformToPrintableTrees(installationTrees []*InstallationTree) ([]PrintableTreeNode, error) {
	var printableTrees []PrintableTreeNode

	for _, installationTree := range installationTrees {
		if t.ShowOnlyFailed {
			installationTree = installationTree.filterForFailedInstallation()
		}

		transformedInstallaltion, err := t.transformInstallation(installationTree)
		if err != nil {
			return nil, fmt.Errorf("error in installation %s: %w", installationTree.Installation.Name, err)
		}

		printableTrees = append(printableTrees, *transformedInstallaltion)
	}
	return printableTrees, nil
}

func (t Transformer) transformInstallation(installationTree *InstallationTree) (*PrintableTreeNode, error) {
	printableNode := PrintableTreeNode{}
	if installationTree == nil {
		return &printableNode, nil
	}

	installationTree.Installation.SetManagedFields(nil)

	namespaceInfo := ""
	if t.ShowNamespaces {
		namespaceInfo = fmt.Sprintf("%s/", installationTree.Installation.Namespace)
	}

	printableNode.Headline = fmt.Sprintf("[%s] Installation %s%s",
		formatStatus(string(installationTree.Installation.Status.InstallationPhase)), namespaceInfo, installationTree.Installation.Name)

	if t.WideMode {
		wide := strings.Builder{}
		cd := "inline"
		if installationTree.Installation.Spec.ComponentDescriptor.Reference != nil {
			cd = fmt.Sprintf("%s:%s", installationTree.Installation.Spec.ComponentDescriptor.Reference.ComponentName, installationTree.Installation.Spec.ComponentDescriptor.Reference.Version)
		}
		wide.WriteString(fmt.Sprintf("Component Descriptor: %s\n", cd))
		bp := "inline"
		if installationTree.Installation.Spec.Blueprint.Reference != nil {
			bp = installationTree.Installation.Spec.Blueprint.Reference.ResourceName
		}
		wide.WriteString(fmt.Sprintf("Blueprint: %s", bp))
		printableNode.WideData = wide.String()
	}

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

func (t Transformer) transformExecution(executionTree *ExecutionTree) (*PrintableTreeNode, error) {
	printableNode := PrintableTreeNode{}
	if executionTree == nil {
		return &printableNode, nil
	}

	executionTree.Execution.SetManagedFields(nil)

	if t.ShowExecutions || t.DetailedMode {
		printableNode.Headline = fmt.Sprintf("[%s] Execution %s",
			formatStatus(string(executionTree.Execution.Status.ExecutionPhase)), executionTree.Execution.Name)
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

func (t Transformer) transformDeployItem(deployItem *DeployItemLeaf) (*PrintableTreeNode, error) {
	printableNode := PrintableTreeNode{}
	if deployItem == nil {
		return &printableNode, nil
	}

	deployItem.DeployItem.SetManagedFields(nil)

	printableNode.Headline = fmt.Sprintf("[%s] DeployItem %s",
		formatStatus(string(deployItem.DeployItem.Status.Phase)), deployItem.DeployItem.Name)

	if t.WideMode {
		wide := strings.Builder{}
		diType := deployItem.DeployItem.Spec.Type
		wide.WriteString(fmt.Sprintf("Type: %s", diType))
		switch diType {
		case "landscaper.gardener.cloud/helm":
			// print helm chart location
			wide.WriteString("\n")
			config := &helmv1alpha1.ProviderConfiguration{}
			err := json.Unmarshal(deployItem.DeployItem.Spec.Configuration.Raw, config)
			if err != nil {
				wide.WriteString("unable to parse helm config")
			} else {
				chartLocation := ""
				if len(config.Chart.Ref) != 0 {
					chartLocation = config.Chart.Ref
				} else if config.Chart.Archive != nil {
					if len(config.Chart.Archive.Raw) != 0 {
						chartLocation = "inline (archive)"
					} else if config.Chart.Archive.Remote != nil {
						chartLocation = fmt.Sprintf("%s (archive)", config.Chart.Archive.Remote.URL)
					}
				} else if config.Chart.FromResource != nil {
					cd := "inline component descriptor"
					if config.Chart.FromResource.Reference != nil {
						cd = fmt.Sprintf("%s:%s", config.Chart.FromResource.Reference.ComponentName, config.Chart.FromResource.Reference.Version)
					}
					chartLocation = fmt.Sprintf("resource %q from %s", config.Chart.FromResource.ResourceName, cd)
				}
				if len(chartLocation) == 0 {
					chartLocation = "unknown"
				}
				wide.WriteString("Chart: ")
				wide.WriteString(chartLocation)
			}
		case "landscaper.gardener.cloud/container":
			// print container image and command
			wide.WriteString("\n")
			config := &containerv1alpha1.ProviderConfiguration{}
			err := json.Unmarshal(deployItem.DeployItem.Spec.Configuration.Raw, config)
			if err != nil {
				wide.WriteString("unable to parse container config")
			} else {
				wide.WriteString("Image: ")
				wide.WriteString(config.Image)
				if len(config.Command) != 0 {
					wide.WriteString("\nCommand:")
					for _, e := range config.Command {
						wide.WriteString(fmt.Sprintf(" %q", e))
					}
				}
				if len(config.Args) != 0 {
					wide.WriteString("\nArgs:")
					for _, e := range config.Args {
						wide.WriteString(fmt.Sprintf(" %q", e))
					}
				}
			}
		}
		printableNode.WideData = wide.String()
	}

	if t.DetailedMode {
		marshaledExecution, err := yaml.Marshal(deployItem.DeployItem)
		if err != nil {
			return nil, fmt.Errorf("failed marshaling DeployItem %s: %w", deployItem.DeployItem.Name, err)
		}
		printableNode.Description = string(marshaledExecution)
	} else if deployItem.DeployItem.Status.LastError != nil {
		printableNode.Description = fmt.Sprintf("Last error: %s", deployItem.DeployItem.Status.LastError.Message)
	}

	return &printableNode, nil
}

func formatStatus(status string) string {
	switch status {
	case string(lsv1alpha1.InstallationPhaseSucceeded):
		return "✅ " + status

	case string(lsv1alpha1.InstallationPhaseInit),
		string(lsv1alpha1.InstallationPhaseObjectsCreated),
		string(lsv1alpha1.InstallationPhaseProgressing),
		string(lsv1alpha1.InstallationPhaseCompleting),
		string(lsv1alpha1.InstallationPhaseInitDelete),
		string(lsv1alpha1.InstallationPhaseTriggerDelete),
		string(lsv1alpha1.InstallationPhaseDeleting):
		return "🏗️ " + status

	case string(lsv1alpha1.InstallationPhaseFailed),
		string(lsv1alpha1.InstallationPhaseDeleteFailed):
		return "❌ " + status

	default:
		return status
	}
}
