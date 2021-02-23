package tests

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	lsv1alpha1 "github.com/gardener/landscaper/apis/core/v1alpha1"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"sigs.k8s.io/controller-runtime/pkg/client"

	inttestutil "github.com/gardener/landscapercli/integration-test/util"
	"github.com/gardener/landscapercli/pkg/util"
)

func RunQuickstartInstallTest(k8sClient client.Client, target *lsv1alpha1.Target, helmChartRef string, config *inttestutil.Config) error {
	// cleanup before
	err := util.DeleteNamespace(k8sClient, config.TestNamespace, config.SleepTime, config.MaxRetries)
	if err != nil {
		return fmt.Errorf("cannot delete namespace before test: %w", err)
	}

	test := quickstartInstallTest{
		k8sClient:    k8sClient,
		target:       target,
		helmChartRef: helmChartRef,
		namespace:    config.TestNamespace,
		sleepTime:    config.SleepTime,
		maxRetries:   config.MaxRetries,
	}
	err = test.run()
	if err != nil {
		// do not cleanup after erroneous test run to keep failed resources on the cluster
		return fmt.Errorf("test failed: %w", err)
	}

	// cleanup after successful test run
	err = util.DeleteNamespace(k8sClient, config.TestNamespace, config.SleepTime, config.MaxRetries)
	if err != nil {
		return fmt.Errorf("cannot delete namespace after test: %w", err)
	}

	return nil
}

type quickstartInstallTest struct {
	k8sClient    client.Client
	target       *lsv1alpha1.Target
	helmChartRef string
	namespace    string
	sleepTime    time.Duration
	maxRetries   int
}

func (t *quickstartInstallTest) run() error {
	const (
		instName = "echo-server"
	)

	ctx := context.TODO()

	namespace := &corev1.Namespace{
		ObjectMeta: metav1.ObjectMeta{
			Name: t.namespace,
		},
	}
	fmt.Printf("Creating namespace %s\n", t.namespace)
	err := t.k8sClient.Create(ctx, namespace, &client.CreateOptions{})
	if err != nil {
		return fmt.Errorf("cannot create namespace %s: %w", t.namespace, err)
	}

	t.target.Namespace = t.namespace
	err = t.k8sClient.Create(ctx, t.target, &client.CreateOptions{})
	if err != nil {
		return fmt.Errorf("cannot create target: %w", err)
	}

	inst, err := buildHelmInstallation(instName, t.target, t.helmChartRef, t.namespace)
	if err != nil {
		return fmt.Errorf("cannot build helm installation: %w", err)
	}

	fmt.Printf("Creating installation %s in namespace %s\n", inst.Name, inst.Namespace)
	err = t.k8sClient.Create(ctx, inst, &client.CreateOptions{})
	if err != nil {
		return fmt.Errorf("cannot create installation: %w", err)
	}

	timeout, err := util.CheckAndWaitUntilLandscaperInstallationSucceeded(t.k8sClient, client.ObjectKey{Name: inst.Name, Namespace: inst.Namespace}, t.sleepTime, t.maxRetries)
	if err != nil {
		return fmt.Errorf("error while waiting for installation to succeed: %w", err)
	}
	if timeout {
		return fmt.Errorf("timeout at waiting for installation")
	}
	fmt.Println("Installation successful")

	return nil
}

func buildHelmInstallation(name string, target *lsv1alpha1.Target, helmChartRef, appNamespace string) (*lsv1alpha1.Installation, error) {
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

	obj := &lsv1alpha1.Installation{
		ObjectMeta: metav1.ObjectMeta{
			Name:      name,
			Namespace: target.Namespace,
		},
		Spec: lsv1alpha1.InstallationSpec{
			Blueprint: lsv1alpha1.BlueprintDefinition{
				Inline: &lsv1alpha1.InlineBlueprint{
					Filesystem: lsv1alpha1.AnyJSON{marshalledFilesystem},
				},
			},
			Imports: lsv1alpha1.InstallationImports{
				Targets: []lsv1alpha1.TargetImportExport{
					{
						Name:   "cluster",
						Target: "#" + target.Name,
					},
				},
			},
			ImportDataMappings: map[string]lsv1alpha1.AnyJSON{
				"appname":      lsv1alpha1.AnyJSON{[]byte(`"echo-server"`)},
				"appnamespace": lsv1alpha1.AnyJSON{[]byte(fmt.Sprintf(`"%s"`, appNamespace))},
			},
		},
	}

	return obj, nil
}
