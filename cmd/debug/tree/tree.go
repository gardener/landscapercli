package tree

import (
	lsv1alpha1 "github.com/gardener/landscaper/apis/core/v1alpha1"
)

type InstallationTree struct {
	SubInstallations []*InstallationTree      `json:"subInstallations,omitempty"`
	Execution        *ExecutionTree           `json:"execution,omitempty"`
	Installation     *lsv1alpha1.Installation `json:"installation,omitempty"`
}

type ExecutionTree struct {
	DeployItems []*DeployItemLeaf     `json:"deployItems,omitempty"`
	Execution   *lsv1alpha1.Execution `json:"execution,omitempty"`
}

type DeployItemLeaf struct {
	DeployItem *lsv1alpha1.DeployItem `json:"deployItems,omitempty"`
}
