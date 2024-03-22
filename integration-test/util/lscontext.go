package util

import (
	"context"
	"encoding/base64"
	"encoding/json"
	"fmt"

	lsv1alpha1 "github.com/gardener/landscaper/apis/core/v1alpha1"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/yaml"
)

type DockerConfigJson struct {
	Auths map[string]DockerConfigEntry `json:"auths"`
}

type DockerConfigEntry struct {
	Username string `json:"username"`
	Password string `json:"password"`
	Email    string `json:"email,omitempty"`
	Auth     string `json:"auth"`
}

func CreateContext(ctx context.Context, k8sClient client.Client, externalRegistryBaseURL, contextNamespace, contextName string) error {
	auth := base64.StdEncoding.EncodeToString([]byte(Testuser + ":" + Testpw))
	dockerConfigJson := DockerConfigJson{
		Auths: map[string]DockerConfigEntry{
			externalRegistryBaseURL: {
				Username: Testuser,
				Password: Testpw,
				Email:    "test@test.com",
				Auth:     auth,
			},
		},
	}

	dockerConfigJsonBytes, _ := json.Marshal(dockerConfigJson)
	dockerConfigJsonString := string(dockerConfigJsonBytes)

	secret := &corev1.Secret{
		ObjectMeta: metav1.ObjectMeta{
			Name:      contextName,
			Namespace: contextNamespace,
		},
		Data: map[string][]byte{
			".dockerconfigjson": []byte(dockerConfigJsonString),
		},
		Type: corev1.SecretTypeDockerConfigJson,
	}

	err := k8sClient.Create(ctx, secret)
	if err != nil {
		return err
	}

	// The type of field repositoryContext (UnstructuredTypedObject) belongs to the component-spec. Avoid to import it.
	contextConfigRaw := fmt.Sprintf(`
registryPullSecrets:
  - name: "%s"
repositoryContext:
  type: OCIRegistry
  baseUrl: "%s"
useOCM: true
`, contextName, externalRegistryBaseURL)

	contextConfig := lsv1alpha1.ContextConfiguration{}
	if err := yaml.Unmarshal([]byte(contextConfigRaw), &contextConfig); err != nil {
		return err
	}

	lsContext := &lsv1alpha1.Context{
		ObjectMeta: metav1.ObjectMeta{
			Name:      contextName,
			Namespace: contextNamespace,
		},
		ContextConfiguration: contextConfig,
	}

	err = k8sClient.Create(ctx, lsContext)
	if err != nil {
		return err
	}

	return nil
}
