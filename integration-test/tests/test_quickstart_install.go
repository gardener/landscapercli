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
	"sigs.k8s.io/controller-runtime/pkg/client"
)

func RunQuickstartInstallTest(k8sClient client.Client, target *landscaper.Target, helmChartRef string) error {
	test := quickstartInstallTest{
		k8sClient:    k8sClient,
		target:       target,
		helmChartRef: helmChartRef,
	}
	return test.run()
}

type quickstartInstallTest struct {
	k8sClient    client.Client
	target       *landscaper.Target
	helmChartRef string
}

func (t *quickstartInstallTest) run() error {
	ctx := context.TODO()

	const (
		instName      = "echo-server"
		testNamespace = "qs-install-test"
		sleepTime     = 5 * time.Second
		maxRetries    = 6
	)

	defer util.DeleteNamespace(t.k8sClient, testNamespace, sleepTime, maxRetries)

	util.DeleteNamespace(t.k8sClient, testNamespace, sleepTime, maxRetries)

	namespace := &corev1.Namespace{
		ObjectMeta: metav1.ObjectMeta{
			Name: testNamespace,
		},
	}
	fmt.Printf("Creating namespace %s\n", testNamespace)
	err := t.k8sClient.Create(ctx, namespace, &client.CreateOptions{})
	if err != nil {
		if k8sErrors.IsAlreadyExists(err) {
			fmt.Printf("Namespace %s already exists...Skipping\n", testNamespace)
		} else {
			return fmt.Errorf("cannot create namespace %s: %w", testNamespace, err)
		}
	}

	t.target.Namespace = testNamespace
	err = t.k8sClient.Create(ctx, t.target, &client.CreateOptions{})
	if err != nil {
		return fmt.Errorf("cannot create target: %w", err)
	}

	inst, err := buildHelmInstallation(instName, t.target, t.helmChartRef, testNamespace)
	if err != nil {
		return fmt.Errorf("cannot build echo-server installation")
	}

	fmt.Printf("Creating installation %s in namespace %s\n", inst.Name, inst.Namespace)
	err = t.k8sClient.Create(ctx, inst, &client.CreateOptions{})
	if err != nil {
		return fmt.Errorf("cannot create installation: %w", err)
	}

	timeout, err := util.CheckAndWaitUntilLandscaperInstallationSucceeded(t.k8sClient, client.ObjectKey{Name: inst.Name, Namespace: inst.Namespace}, sleepTime, maxRetries)
	if err != nil {
		return fmt.Errorf("error while waiting for installation: %w", err)
	}
	if timeout {
		return fmt.Errorf("timeout at waiting for installation")
	}
	fmt.Println("Installation successful")

	return nil
}

func buildHelmInstallation(name string, target *landscaper.Target, helmChartRef, appNamespace string) (*landscaper.Installation, error) {
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
			Namespace: target.Namespace,
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
						Target: "#" + target.Name,
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
