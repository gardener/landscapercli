package tree

import (
	"github.com/gardener/landscaper/apis/core/v1alpha1"
	lsv1alpha1 "github.com/gardener/landscaper/apis/core/v1alpha1"
)

// InstallationTree contains the Installation and the references to Sub-Installations and the Execution.
type InstallationTree struct {
	SubInstallations []*InstallationTree      `json:"subInstallations,omitempty"`
	Execution        *ExecutionTree           `json:"execution,omitempty"`
	Installation     *lsv1alpha1.Installation `json:"installation,omitempty"`
}

func (i *InstallationTree) filterForFailedInstallation() *InstallationTree {
	filteredSubInstallations := []*InstallationTree{}
	for _, subInstallation := range i.SubInstallations {
		if filteredSubInstallation := subInstallation.filterForFailedInstallation(); filteredSubInstallation != nil {
			filteredSubInstallations = append(filteredSubInstallations, filteredSubInstallation)
		}
	}
	i.SubInstallations = filteredSubInstallations

	if i.Execution != nil {
		i.Execution = i.Execution.filterForFailedExecution()
	}

	if len(filteredSubInstallations) > 0 || i.Execution != nil || i.Installation.Status.InstallationPhase == v1alpha1.InstallationPhases.Failed {
		return i
	}
	return nil

}

// ExecutionTree contains the Execution and the references to all DeployItems.
type ExecutionTree struct {
	DeployItems []*DeployItemLeaf     `json:"deployItems,omitempty"`
	Execution   *lsv1alpha1.Execution `json:"execution,omitempty"`
}

func (e *ExecutionTree) filterForFailedExecution() *ExecutionTree {
	filteredDeployItems := []*DeployItemLeaf{}
	for _, depItem := range e.DeployItems {
		if filteredDeployItem := depItem.filterForFailedDeployItem(); filteredDeployItem != nil {
			filteredDeployItems = append(filteredDeployItems, filteredDeployItem)
		}
	}

	e.DeployItems = filteredDeployItems

	if len(filteredDeployItems) > 0 || e.Execution.Status.ExecutionPhase == v1alpha1.ExecutionPhases.Failed {
		return e
	}
	return nil
}

// DeployItemLeaf contains a DeployItem.
type DeployItemLeaf struct {
	DeployItem *lsv1alpha1.DeployItem `json:"deployItem,omitempty"`
}

func (d *DeployItemLeaf) filterForFailedDeployItem() *DeployItemLeaf {
	if d.DeployItem.Status.Phase.IsFailed() {
		return d
	}
	return nil
}
