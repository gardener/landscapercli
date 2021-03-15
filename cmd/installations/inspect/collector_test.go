package tree

import (
	"testing"

	"github.com/gardener/landscaper/test/utils/envtest"
	"github.com/stretchr/testify/assert"
)

func TestCollector(t *testing.T) {
	fakeClient, state, err := envtest.NewFakeClientFromPath("./testdata")
	assert.NoError(t, err)

	collector := Collector{
		K8sClient: fakeClient,
	}
	expectedInstallationAgg := state.Installations["inttest/my-aggregation"]
	expectedInstallationServer := state.Installations["inttest/server-gw64l"]
	expectedInstallationIngress := state.Installations["inttest/ingress-hhsjf"]

	expectedExecutionServer := state.Executions["inttest/server-gw64l"]
	expectedExecutionIngress := state.Executions["inttest/ingress-hhsjf"]

	expectedDeployitemServer := state.DeployItems["inttest/server-gw64l-deploy-7mhc2"]
	expectedDeployitemIngress := state.DeployItems["inttest/ingress-hhsjf-deploy-kptjl"]

	actualStructure, err := collector.CollectInstallationsInCluster("", "inttest")
	assert.NoError(t, err)

	expectedStructure := []*InstallationTree{
		{
			Installation: expectedInstallationAgg,
			SubInstallations: []*InstallationTree{
				{
					Installation: expectedInstallationIngress,
					Execution: &ExecutionTree{
						Execution: expectedExecutionIngress,
						DeployItems: []*DeployItemLeaf{
							{
								DeployItem: expectedDeployitemIngress,
							},
						},
					},
				},
				{
					Installation: expectedInstallationServer,
					Execution: &ExecutionTree{
						Execution: expectedExecutionServer,
						DeployItems: []*DeployItemLeaf{
							{
								DeployItem: expectedDeployitemServer,
							},
						},
					},
				},
			},
		},
	}

	t.Run("Collection of landscaper CR (installations, executions, deployitems", func(t *testing.T) {
		assert.Equal(t, expectedStructure, actualStructure)
	})

}
