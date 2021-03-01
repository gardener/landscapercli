package tree

import (
	lsv1alpha1 "github.com/gardener/landscaper/apis/core/v1alpha1"
)

type InstallationTree struct {
	SubInstallations []InstallationTree `json:"subInstallations,omitempty"`
	Execution        *ExecutionTree     `json:"executions,omitempty"`
	Installation     *lsv1alpha1.Installation
}

type ExecutionTree struct {
	DeployItems []DeployItemTree `json:"deployItems,omitempty"`
	Execution   *lsv1alpha1.Execution
}

type DeployItemTree struct {
	DeployItem *lsv1alpha1.DeployItem
}
