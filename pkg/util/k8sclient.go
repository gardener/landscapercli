package util

import (
	"fmt"

	"k8s.io/apimachinery/pkg/runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"

	"k8s.io/client-go/tools/clientcmd"

	lsv1alpha1 "github.com/gardener/landscaper/apis/core/v1alpha1"
	clientgoscheme "k8s.io/client-go/kubernetes/scheme"
)

//GetK8sClientFromCurrentConfiguredCluster returns a k8sClient and a namespace from the current context of the kubectl program.
func GetK8sClientFromCurrentConfiguredCluster() (client.Client, string, error) {
	rules := clientcmd.NewDefaultClientConfigLoadingRules()
	rules.DefaultClientConfig = &clientcmd.DefaultClientConfig

	overrides := &clientcmd.ConfigOverrides{ClusterDefaults: clientcmd.ClusterDefaults}
	clientConfig := clientcmd.NewNonInteractiveDeferredLoadingClientConfig(rules, overrides)
	cfg, err := clientConfig.ClientConfig()

	if err != nil {
		return nil, "", fmt.Errorf("cannot build k8s config %w", err)
	}
	namespace, _, err := clientConfig.Namespace()
	if err != nil {
		return nil, "", fmt.Errorf("error extracting namespace from current k8s context. %w", err)
	}

	scheme := runtime.NewScheme()
	_ = clientgoscheme.AddToScheme(scheme)
	_ = lsv1alpha1.AddToScheme(scheme)
	k8sClient, err := client.New(cfg, client.Options{
		Scheme: scheme,
	})
	if err != nil {
		return nil, "", fmt.Errorf("cannot build K8s client: %w", err)
	}

	return k8sClient, namespace, nil
}

func BuildKubeClientFromConfigOrCurrentClusterContext(kubeconfig string, scheme *runtime.Scheme) (client.Client, string, error) {
	var err error
	namespace := ""
	var k8sClient client.Client
	if kubeconfig != "" {
		cfg, err := clientcmd.BuildConfigFromFlags("", kubeconfig)
		if err != nil {
			return nil, namespace, fmt.Errorf("cannot parse K8s config: %w", err)
		}
		k8sClient, err = client.New(cfg, client.Options{
			Scheme: scheme,
		})
		if err != nil {
			return nil, namespace, fmt.Errorf("cannot build K8s client: %w", err)
		}
	} else {
		k8sClient, namespace, err = GetK8sClientFromCurrentConfiguredCluster()
		if err != nil {
			return nil, namespace, fmt.Errorf("cannot build K8s client from current cluster config: %w", err)
		}
	}

	return k8sClient, namespace, nil
}
