package tree

import (
	"testing"

	"github.com/gardener/landscaper/test/utils/envtest"
	"github.com/stretchr/testify/assert"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func TestCollector(t *testing.T) {
	fakeClient, state, err := envtest.NewFakeClientFromPath("./testdata")
	assert.NoError(t, err)

	collector := Collector{
		K8sClient: fakeClient,
	}
	expectedInstallationAgg := state.Installations["inttest/my-aggregation"]
	expectedInstallationAgg.TypeMeta = metav1.TypeMeta{}
	expectedInstallationServer := state.Installations["inttest/server-gw64l"]
	expectedInstallationServer.TypeMeta = metav1.TypeMeta{}
	expectedInstallationIngress := state.Installations["inttest/ingress-hhsjf"]
	expectedInstallationIngress.TypeMeta = metav1.TypeMeta{}
	expectedExecutionServer := state.Executions["inttest/server-gw64l"]
	expectedExecutionServer.TypeMeta = metav1.TypeMeta{}
	expectedExecutionIngress := state.Executions["inttest/ingress-hhsjf"]
	expectedExecutionIngress.TypeMeta = metav1.TypeMeta{}
	expectedDeployitemServer := state.DeployItems["inttest/server-gw64l-deploy-7mhc2"]
	expectedDeployitemServer.TypeMeta = metav1.TypeMeta{}
	expectedDeployitemIngress := state.DeployItems["inttest/ingress-hhsjf-deploy-kptjl"]
	expectedDeployitemIngress.TypeMeta = metav1.TypeMeta{}

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

	t.Run("Collection of landscaper CR (installations, executions, deployitems)", func(t *testing.T) {
		assert.Equal(t, expectedStructure, actualStructure)
	})

	actualStructure, err = collector.CollectInstallationsInCluster("", "*")
	assert.NoError(t, err)
	expectedFakeInstallation := state.Installations["default/fakeinst"]
	expectedFakeInstallation.TypeMeta = metav1.TypeMeta{}
	// verify that all of the previous installations as well as the fakeinst from the default namespace are contained in the actual structure
	t.Run("Collection of landscaper CR (installations, executions, deployitems) across all namespaces", func(t *testing.T) {
		assert.Contains(t, actualStructure, &InstallationTree{Installation: expectedFakeInstallation})
		for _, it := range expectedStructure {
			assert.Contains(t, actualStructure, it)
		}
	})
}
