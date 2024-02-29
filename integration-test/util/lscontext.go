package util

import (
	"context"
	"encoding/base64"
	"encoding/json"
	"fmt"
	cdv2 "github.com/gardener/component-spec/bindings-go/apis/v2"
	lsv1alpha1 "github.com/gardener/landscaper/apis/core/v1alpha1"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"sigs.k8s.io/controller-runtime/pkg/client"
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

	repoCtx := &cdv2.UnstructuredTypedObject{}
	repoCtx.UnmarshalJSON([]byte(fmt.Sprintf(`{"type": "OCIRegistry", "baseUrl": "%s"}`, externalRegistryBaseURL)))
	lscontext := &lsv1alpha1.Context{
		ObjectMeta: metav1.ObjectMeta{
			Name:      contextName,
			Namespace: contextNamespace,
		},
		ContextConfiguration: lsv1alpha1.ContextConfiguration{
			RegistryPullSecrets: []corev1.LocalObjectReference{corev1.LocalObjectReference{
				Name: contextName,
			}},
			RepositoryContext: repoCtx,
			UseOCM:            true,
		},
	}

	err = k8sClient.Create(ctx, lscontext)
	if err != nil {
		return err
	}

	return nil
}
