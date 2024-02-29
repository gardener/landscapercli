package tests

import (
	"context"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"time"

	cdv2 "github.com/gardener/component-spec/bindings-go/apis/v2"

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
		k8sClient:               k8sClient,
		target:                  target,
		helmChartRef:            helmChartRef,
		namespace:               config.TestNamespace,
		sleepTime:               config.SleepTime,
		maxRetries:              config.MaxRetries,
		externalRegistryBaseURL: config.ExternalRegistryBaseURL,
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
	k8sClient               client.Client
	target                  *lsv1alpha1.Target
	helmChartRef            string
	namespace               string
	sleepTime               time.Duration
	maxRetries              int
	externalRegistryBaseURL string
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

	contextName, err := t.createContext(ctx)
	if err != nil {
		return fmt.Errorf("cannot create context: %w", err)
	}

	inst, err := buildHelmInstallation(instName, t.target, t.helmChartRef, t.namespace, contextName)
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

type DockerConfigJson struct {
	Auths map[string]DockerConfigEntry `json:"auths"`
}

type DockerConfigEntry struct {
	Username string `json:"username"`
	Password string `json:"password"`
	Email    string `json:"email,omitempty"`
	Auth     string `json:"auth"`
}

func (t *quickstartInstallTest) createContext(ctx context.Context) (string, error) {
	auth := base64.StdEncoding.EncodeToString([]byte(inttestutil.Testuser + ":" + inttestutil.Testpw))
	dockerConfigJson := DockerConfigJson{
		Auths: map[string]DockerConfigEntry{
			t.externalRegistryBaseURL: {
				Username: inttestutil.Testuser,
				Password: inttestutil.Testpw,
				Email:    "test@test.com",
				Auth:     auth,
			},
		},
	}

	dockerConfigJsonBytes, _ := json.Marshal(dockerConfigJson)
	dockerConfigJsonString := string(dockerConfigJsonBytes)

	secretAndContextName := "installcontext"
	secret := &corev1.Secret{
		ObjectMeta: metav1.ObjectMeta{
			Name:      secretAndContextName,
			Namespace: t.namespace,
		},
		Data: map[string][]byte{
			".dockerconfigjson": []byte(dockerConfigJsonString),
		},
		Type: corev1.SecretTypeDockerConfigJson,
	}

	err := t.k8sClient.Create(ctx, secret)
	if err != nil {
		return "", err
	}

	repoCtx := &cdv2.UnstructuredTypedObject{}
	repoCtx.UnmarshalJSON([]byte(fmt.Sprintf(`{"type": "OCIRegistry", "baseUrl": "%s"}`, t.externalRegistryBaseURL)))
	lscontext := &lsv1alpha1.Context{
		ObjectMeta: metav1.ObjectMeta{
			Name:      secretAndContextName,
			Namespace: t.namespace,
		},
		ContextConfiguration: lsv1alpha1.ContextConfiguration{
			RegistryPullSecrets: []corev1.LocalObjectReference{corev1.LocalObjectReference{
				Name: secretAndContextName,
			}},
			RepositoryContext: repoCtx,
			UseOCM:            true,
		},
	}

	err = t.k8sClient.Create(ctx, lscontext)
	if err != nil {
		return "", err
	}

	return secretAndContextName, nil
}

func buildHelmInstallation(name string, target *lsv1alpha1.Target, helmChartRef, appNamespace, contextName string) (*lsv1alpha1.Installation, error) {
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
			Annotations: map[string]string{
				lsv1alpha1.OperationAnnotation: string(lsv1alpha1.ReconcileOperation),
			},
		},
		Spec: lsv1alpha1.InstallationSpec{
			Context: contextName,
			Blueprint: lsv1alpha1.BlueprintDefinition{
				Inline: &lsv1alpha1.InlineBlueprint{
					Filesystem: lsv1alpha1.AnyJSON{RawMessage: marshalledFilesystem},
				},
			},
			Imports: lsv1alpha1.InstallationImports{
				Targets: []lsv1alpha1.TargetImport{
					{
						Name:   "cluster",
						Target: "#" + target.Name,
					},
				},
			},
			ImportDataMappings: map[string]lsv1alpha1.AnyJSON{
				"appname":      lsv1alpha1.AnyJSON{RawMessage: []byte(`"echo-server"`)},
				"appnamespace": lsv1alpha1.AnyJSON{RawMessage: []byte(fmt.Sprintf(`"%s"`, appNamespace))},
			},
		},
	}

	return obj, nil
}
