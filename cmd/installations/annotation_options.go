package installations

import (
	"context"
	"fmt"

	"github.com/gardener/landscaper/apis/core/v1alpha1"
	"github.com/gardener/landscaper/pkg/landscaper/installations"
	"github.com/spf13/pflag"
	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"sigs.k8s.io/controller-runtime/pkg/client"

	"github.com/gardener/landscapercli/pkg/util"
)

type annotationOptions struct {
	annotationKey         string
	annotationValue       string
	rootInstallationsOnly bool

	kubeconfig       string
	installationName string
	namespace        string
}

func (o *annotationOptions) run(ctx context.Context) error {
	kubeClient, namespace, err := util.BuildKubeClientFromConfigOrCurrentClusterContext(o.kubeconfig, scheme)
	if err != nil {
		return fmt.Errorf("cannot build k8s client from config or current cluster context: %w", err)
	}

	if namespace != "" && o.namespace == "" {
		o.namespace = namespace
	}

	if o.namespace == "" {
		return fmt.Errorf("namespace was not defined. Use --namespace to specify a namespace")
	}

	if o.installationName == "" {
		return fmt.Errorf("installationName was not defined.")
	}

	installationKey := client.ObjectKey{Namespace: o.namespace, Name: o.installationName}
	installation := &v1alpha1.Installation{}
	if err := kubeClient.Get(ctx, installationKey, installation); err != nil {
		return fmt.Errorf("failed to read installation: %w", err)
	}

	if o.rootInstallationsOnly && !installations.IsRootInstallation(installation) {
		return fmt.Errorf("the command is only supported for root installations")
	}

	v1.SetMetaDataAnnotation(&installation.ObjectMeta, o.annotationKey, o.annotationValue)
	if err := kubeClient.Update(ctx, installation); err != nil {
		return fmt.Errorf("failed to update installation: %w", err)
	}

	return nil
}

func (o *annotationOptions) validateArgs(args []string) error {
	if len(args) == 1 {
		o.installationName = args[0]
	}
	return nil
}

func (o *annotationOptions) AddFlags(fs *pflag.FlagSet) {
	fs.StringVar(&o.kubeconfig, "kubeconfig", "", "path to the kubeconfig for the cluster. Required if the cluster is not the same as the current-context of kubectl.")
	fs.StringVarP(&o.namespace, "namespace", "n", "", "namespace of the installation. Required if --kubeconfig is used.")
}
