package tree

import (
	lsv1alpha1 "github.com/gardener/landscaper/apis/core/v1alpha1"
)

//InstallationTree contains the Installation and the references to subInstallations and the Execution.
type InstallationTree struct {
	SubInstallations []*InstallationTree      `json:"subInstallations,omitempty"`
	Execution        *ExecutionTree           `json:"execution,omitempty"`
	Installation     *lsv1alpha1.Installation `json:"installation,omitempty"`
}

//ExecutionTree contains the Execution and the references to all deployItems.
type ExecutionTree struct {
	DeployItems []*DeployItemLeaf     `json:"deployItems,omitempty"`
	Execution   *lsv1alpha1.Execution `json:"execution,omitempty"`
}

//DeployItemLeaf contains a DeployItem.
type DeployItemLeaf struct {
	DeployItem *lsv1alpha1.DeployItem `json:"deployItem,omitempty"`
}
