package tests

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	landscaper "github.com/gardener/landscaper/pkg/apis/core/v1alpha1"
	"github.com/gardener/landscapercli/pkg/util"
	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

func RunQuickstartInstallTest(k8sClient client.Client, targetName string) error {
	test := quickstartInstallTest{
		k8sClient: k8sClient,
		targetName: "#"+targetName,
	}
	return test.run()
}

type quickstartInstallTest struct {
	k8sClient     client.Client
	targetName string
}

func (t *quickstartInstallTest) run() error {
	const (
		instName = "test-1"
		instNs = "landscaper"
		sleepTime = 5*time.Second
		maxRetries = 4
	)

	ctx := context.TODO()
	inst := buildInstallation(instName, instNs, t.targetName, "landscaper-example")

	err := t.k8sClient.Create(ctx, inst, &client.CreateOptions{})
	if err != nil {
		return fmt.Errorf("Cannot create installation: %w", err)
	}

	err = util.WaitUntilLandscaperInstallationSucceeded(t.k8sClient, client.ObjectKey{Name: instName, Namespace: instNs}, sleepTime, maxRetries)
	if err != nil {
		return fmt.Errorf("Error while waiting for installation: %w", err)
	}

	err = t.k8sClient.Delete(ctx, inst, &client.DeleteOptions{})
	if err != nil {
		return fmt.Errorf("Cannot delete installation: %w", err)
	}

	err = util.WaitUntilObjectIsDeleted(t.k8sClient, client.ObjectKey{Name: instName, Namespace: instNs}, inst, sleepTime, maxRetries)
	if err != nil {
		return fmt.Errorf("Error while waiting for installation deletion: %w", err)
	}

	return nil
}

func buildInstallation(name, namespace, targetName, appNamespace string) *landscaper.Installation {
	inlineBlueprint := `
        apiVersion: landscaper.gardener.cloud/v1alpha1
        kind: Blueprint

        imports:
        - name: cluster
          targetType: landscaper.gardener.cloud/kubernetes-cluster

        - name: appname
          schema:
            type: string

        - name: appnamespace
          schema:
            type: string

        deployExecutions:
        - name: default
          type: GoTemplate
          template: |
            deployItems:
            - name: deploy
              type: landscaper.gardener.cloud/helm
              target:
                name: {{ .imports.cluster.metadata.name }}
                namespace: {{ .imports.cluster.metadata.namespace }}
              config:
                apiVersion: helm.deployer.landscaper.gardener.cloud/v1alpha1
                kind: ProviderConfiguration

                chart:
                  ref: "oci-registry.landscaper.svc.cluster.local:5000/echo-server-chart:v1.1.0"
        
                updateStrategy: patch

                name: {{ .imports.appname }}
                namespace: {{ .imports.appnamespace }}
    `

	fs := map[string]interface{}{
		"blueprint.yaml": inlineBlueprint,
	}

	marsh, err := json.Marshal(fs)
	if err != nil {
		panic(err)
	}

	obj := &landscaper.Installation{
		ObjectMeta: v1.ObjectMeta{
			Name:      name,
			Namespace: namespace,
		},
		Spec: landscaper.InstallationSpec{
			Blueprint: landscaper.BlueprintDefinition{
				Inline: &landscaper.InlineBlueprint{
					Filesystem: json.RawMessage(marsh),
				},
			},
			Imports: landscaper.InstallationImports{
				Targets: []landscaper.TargetImportExport{
					{
						Name:   "cluster",
						Target: targetName,
					},
				},
			},
			ImportDataMappings: map[string]json.RawMessage{
				"appname":      json.RawMessage(`"echo-server"`),
				"appnamespace": json.RawMessage(fmt.Sprintf(`"%s"`, appNamespace)),
			},
		},
	}

	return obj
}