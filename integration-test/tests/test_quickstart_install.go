package tests

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	landscaper "github.com/gardener/landscaper/pkg/apis/core/v1alpha1"
	"github.com/gardener/landscapercli/pkg/util"
	corev1 "k8s.io/api/core/v1"
	k8sErrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

func RunQuickstartInstallTest(k8sClient client.Client, targetKey types.NamespacedName, helmChartRef string) error {
	test := quickstartInstallTest{
		k8sClient:    k8sClient,
		targetKey:    targetKey,
		helmChartRef: helmChartRef,
	}
	return test.run()
}

type quickstartInstallTest struct {
	k8sClient    client.Client
	targetKey    types.NamespacedName
	helmChartRef string
}

func (t *quickstartInstallTest) run() error {
	const (
		instName     = "echo-server"
		appNamespace = "inttest-qs-install"
		sleepTime    = 5 * time.Second
		maxRetries   = 4
	)

	ctx := context.TODO()

	namespace := &corev1.Namespace{
		ObjectMeta: metav1.ObjectMeta{
			Name: appNamespace,
		},
	}
	fmt.Printf("Creating namespace %s\n", appNamespace)
	err := t.k8sClient.Create(ctx, namespace, &client.CreateOptions{})
	if err != nil {
		if k8sErrors.IsAlreadyExists(err) {
			fmt.Printf("Namespace %s already exists...Skipping\n", appNamespace)
		} else {
			return fmt.Errorf("cannot create namespace %s: %w", appNamespace, err)
		}
	}

	inst, err := buildHelmInstallation(instName, t.targetKey, t.helmChartRef, appNamespace)
	if err != nil {
		return fmt.Errorf("cannot build echo-server installation")
	}

	fmt.Printf("Creating installation %s in namespace %s\n", inst.Name, inst.Namespace)
	err = t.k8sClient.Create(ctx, inst, &client.CreateOptions{})
	if err != nil {
		return fmt.Errorf("cannot create installation: %w", err)
	}

	err = util.WaitUntilLandscaperInstallationSucceeded(t.k8sClient, client.ObjectKey{Name: inst.Name, Namespace: inst.Namespace}, sleepTime, maxRetries)
	if err != nil {
		return fmt.Errorf("error while waiting for installation: %w", err)
	}
	fmt.Println("Installation successful")

	fmt.Println("Deleting installation")
	err = t.k8sClient.Delete(ctx, inst, &client.DeleteOptions{})
	if err != nil {
		return fmt.Errorf("cannot delete installation: %w", err)
	}

	err = util.WaitUntilObjectIsDeleted(t.k8sClient, client.ObjectKey{Name: inst.Name, Namespace: inst.Namespace}, inst, sleepTime, maxRetries)
	if err != nil {
		return fmt.Errorf("error while waiting for installation deletion: %w", err)
	}
	fmt.Println("Installation successfully deleted")

	return nil
}

func buildHelmInstallation(name string, targetKey types.NamespacedName, helmChartRef, appNamespace string) (*landscaper.Installation, error) {
	inlineBlueprint := fmt.Sprintf(`
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
                  ref: "%s"

                updateStrategy: patch

                name: {{ .imports.appname }}
                namespace: {{ .imports.appnamespace }}
    `, helmChartRef)

	filesystem := map[string]interface{}{
		"blueprint.yaml": inlineBlueprint,
	}

	marshalledFilesystem, err := json.Marshal(filesystem)
	if err != nil {
		return nil, fmt.Errorf("cannot marshall inline blueprint: %w", err)
	}

	obj := &landscaper.Installation{
		ObjectMeta: metav1.ObjectMeta{
			Name:      name,
			Namespace: targetKey.Namespace,
		},
		Spec: landscaper.InstallationSpec{
			Blueprint: landscaper.BlueprintDefinition{
				Inline: &landscaper.InlineBlueprint{
					Filesystem: marshalledFilesystem,
				},
			},
			Imports: landscaper.InstallationImports{
				Targets: []landscaper.TargetImportExport{
					{
						Name:   "cluster",
						Target: "#"+targetKey.Name,
					},
				},
			},
			ImportDataMappings: map[string]json.RawMessage{
				"appname":      json.RawMessage(`"echo-server"`),
				"appnamespace": json.RawMessage(fmt.Sprintf(`"%s"`, appNamespace)),
			},
		},
	}

	return obj, nil
}
