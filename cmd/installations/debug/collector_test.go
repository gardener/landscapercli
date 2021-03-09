package tree

import (
	"testing"

	"github.com/gardener/landscaper/test/utils/envtest"
	"github.com/stretchr/testify/assert"
)

func TestCollector(t *testing.T) {
	fakeClient, state, err := envtest.NewFakeClientFromPath("./testdata/")
	assert.NoError(t, err)

	collector := Collector{
		K8sClient: fakeClient,
	}
	expectedInstallation := state.Installations["ls-cli-inttest/echo-server"]
	expectedExecution := state.Executions["ls-cli-inttest/echo-server"]
	expectedDeployitem := state.DeployItems["ls-cli-inttest/echo-server-deploy-k4v5p"]

	actualStructure, err := collector.CollectInstallationsInCluster("", "ls-cli-inttest")
	assert.NoError(t, err)

	expectedStructure := []*InstallationTree{
		{
			Installation: expectedInstallation,
			Execution: &ExecutionTree{
				Execution:   expectedExecution,
				DeployItems: []*DeployItemLeaf{{expectedDeployitem}},
			},
		},
	}

	t.Run("Collection of landscaper CR (installations, executions, deployitems", func(t *testing.T) {
		assert.Equal(t, expectedStructure, actualStructure)
	})

}
