package quickstart

import (
	"context"
	"fmt"
	v1 "k8s.io/api/admissionregistration/v1"
	extv1 "k8s.io/apiextensions-apiserver/pkg/apis/apiextensions/v1"
	"os"
	"strings"
	"time"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/util/wait"

	"github.com/go-logr/logr"
	"github.com/spf13/cobra"
	"github.com/spf13/pflag"
	corev1 "k8s.io/api/core/v1"
	k8sErrors "k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/client-go/tools/clientcmd"
	"sigs.k8s.io/controller-runtime/pkg/client"

	"github.com/gardener/landscaper/apis/core/v1alpha1"

	"github.com/gardener/landscapercli/pkg/logger"
	"github.com/gardener/landscapercli/pkg/util"
)

type uninstallOptions struct {
	kubeconfigPath  string
	namespace       string
	deleteNamespace bool
	deleteCrd       bool
}

func NewUninstallCommand(ctx context.Context) *cobra.Command {
	opts := &uninstallOptions{}
	cmd := &cobra.Command{
		Use:     "uninstall --kubeconfig [kubconfig.yaml] --delete-namespace --delete-crd",
		Aliases: []string{"u"},
		Short:   "command to uninstall Landscaper and OCI registry (from the install command) in a target cluster",
		Example: "landscaper-cli quickstart uninstall --kubeconfig ./kubconfig.yaml --namespace landscaper --delete-namespace --delete-crd",
		Run: func(cmd *cobra.Command, args []string) {
			if err := opts.Complete(args); err != nil {
				fmt.Println(err.Error())
				os.Exit(1)
			}

			if err := opts.run(ctx, logger.Log); err != nil {
				fmt.Println(err.Error())
				os.Exit(1)
			}
		},
	}

	opts.AddFlags(cmd.Flags())

	return cmd
}

func (o *uninstallOptions) run(ctx context.Context, log logr.Logger) error {
	cfg, err := clientcmd.BuildConfigFromFlags("", o.kubeconfigPath)
	if err != nil {
		return fmt.Errorf("cannot parse K8s config: %w", err)
	}

	k8sClient, err := client.New(cfg, client.Options{
		Scheme: scheme,
	})
	if err != nil {
		return fmt.Errorf("cannot build K8s client: %w", err)
	}

	key := client.ObjectKey{
		Name: o.namespace,
	}
	ns := corev1.Namespace{}
	if err := k8sClient.Get(ctx, key, &ns); err != nil {
		if k8sErrors.IsNotFound(err) {
			fmt.Printf("Cannot find namespace %s\n", o.namespace)
			return nil
		}
		return fmt.Errorf("cannot get namespace: %w", err)
	}

	fmt.Println("Uninstall OCI Registry")
	if err := o.uninstallOCIRegistry(ctx, k8sClient); err != nil {
		return fmt.Errorf("cannot uninstall OCI registry: %w", err)
	}
	fmt.Print("OCI registry uninstall succeeded!\n\n")

	fmt.Println("Uninstall Landscaper")
	if err := o.uninstallLandscaper(ctx, k8sClient); err != nil {
		return fmt.Errorf("cannot uninstall landscaper: %w", err)
	}
	fmt.Println("Landscaper uninstall succeeded!")

	return nil
}

func (o *uninstallOptions) Complete(args []string) error {
	return nil
}

func (o *uninstallOptions) AddFlags(fs *pflag.FlagSet) {
	fs.StringVar(&o.kubeconfigPath, "kubeconfig", "", "path to the kubeconfig of the target cluster")
	fs.StringVar(&o.namespace, "namespace", defaultNamespace, "namespace where Landscaper and the OCI registry are installed")
	fs.BoolVar(&o.deleteNamespace, "delete-namespace", false, "deletes the namespace (otherwise secrets, service accounts etc. of the landscaper installation in the namespace are not removed)")
	fs.BoolVar(&o.deleteCrd, "delete-crd", false, "deletes the Landscaper CRDs and all CRs of theses types without uninstalling the data deployed by them")

}

func (o *uninstallOptions) uninstallOCIRegistry(ctx context.Context, k8sClient client.Client) error {
	ociRegistryOpts := &ociRegistryOpts{
		namespace: o.namespace,
	}
	ociRegistry := NewOCIRegistry(ociRegistryOpts, k8sClient)
	return ociRegistry.uninstall(ctx)
}

func (o *uninstallOptions) uninstallLandscaper(ctx context.Context, k8sClient client.Client) error {
	fmt.Println("Removing deployer registrations")

	crdList := &extv1.CustomResourceDefinitionList{}
	if err := k8sClient.List(ctx, crdList); err != nil {
		return err
	}

	if o.containsDeployerRegistration(crdList) {
		deployerRegistrations := v1alpha1.DeployerRegistrationList{}
		if err := k8sClient.List(ctx, &deployerRegistrations); err != nil {
			return err
		}

		for _, registration := range deployerRegistrations.Items {
			if err := k8sClient.Delete(ctx, &registration); err != nil {
				return err
			}
		}

		if err := waitForRegistrationsRemoved(ctx, k8sClient); err != nil {
			return err
		}
	}

	err := util.ExecCommandBlocking(fmt.Sprintf("helm delete --namespace %s helm-deployer --kubeconfig %s", o.namespace, o.kubeconfigPath))
	if err != nil {
		if strings.Contains(err.Error(), "not found") {
			// Ignore error if the release that should be deleted was not found ;)
			fmt.Println("Helm deployer release not found...Skipping")
		} else {
			return err
		}
	}

	err = util.ExecCommandBlocking(fmt.Sprintf("helm delete --namespace %s manifest-deployer --kubeconfig %s", o.namespace, o.kubeconfigPath))
	if err != nil {
		if strings.Contains(err.Error(), "not found") {
			// Ignore error if the release that should be deleted was not found ;)
			fmt.Println("Manifest deployer release not found...Skipping")
		} else {
			return err
		}
	}

	err = util.ExecCommandBlocking(fmt.Sprintf("helm delete --namespace %s container-deployer --kubeconfig %s", o.namespace, o.kubeconfigPath))
	if err != nil {
		if strings.Contains(err.Error(), "not found") {
			// Ignore error if the release that should be deleted was not found ;)
			fmt.Println("Container deployer release not found...Skipping")
		} else {
			return err
		}
	}

	err = util.ExecCommandBlocking(fmt.Sprintf("helm delete --namespace %s landscaper --kubeconfig %s", o.namespace, o.kubeconfigPath))
	if err != nil {
		if strings.Contains(err.Error(), "not found") {
			// Ignore error if the release that should be deleted was not found ;)
			fmt.Println("Landscaper release not found...Skipping")
		} else {
			return err
		}
	}

	fmt.Println("Removing Validating Webhook")
	if err = o.deleteWebhook(ctx, k8sClient); err != nil {
		return err
	}

	if o.deleteCrd {
		fmt.Println("Removing CRDs")
		if err = o.deleteCrds(ctx, k8sClient, crdList); err != nil {
			return err
		}
	}

	if o.deleteNamespace {
		fmt.Println("Removing namespace")
		namespace := corev1.Namespace{
			ObjectMeta: metav1.ObjectMeta{Name: o.namespace},
		}

		if err := k8sClient.Delete(ctx, &namespace); err != nil {
			return err
		}
	}

	return nil
}

func (o *uninstallOptions) containsDeployerRegistration(crds *extv1.CustomResourceDefinitionList) bool {
	for i := range crds.Items {
		if crds.Items[i].Name == "deployerregistrations.landscaper.gardener.cloud" {
			return true
		}
	}

	return false
}

func (o *uninstallOptions) deleteCrds(ctx context.Context, k8sClient client.Client, crds *extv1.CustomResourceDefinitionList) error {
	for i := range crds.Items {
		nextCrd := &crds.Items[i]

		found := true

		nextName := nextCrd.Name
		switch nextName {
		case "componentversionoverwrites.landscaper.gardener.cloud":
			objectList := &v1alpha1.ComponentVersionOverwritesList{}
			if err := removeCrdComponentversionoverwrites(ctx, k8sClient, nextName, objectList, nextCrd); err != nil {
				return err
			}
		case "contexts.landscaper.gardener.cloud":
			objectList := &v1alpha1.ContextList{}
			if err := removeCrdContext(ctx, k8sClient, nextName, objectList, nextCrd); err != nil {
				return err
			}
		case "dataobjects.landscaper.gardener.cloud":
			objectList := &v1alpha1.DataObjectList{}
			if err := removeCrdDataObject(ctx, k8sClient, nextName, objectList, nextCrd); err != nil {
				return err
			}
		case "deployerregistrations.landscaper.gardener.cloud":
			objectList := &v1alpha1.DeployerRegistrationList{}
			if err := removeCrdDeployerRegistration(ctx, k8sClient, nextName, objectList, nextCrd); err != nil {
				return err
			}
		case "deployitems.landscaper.gardener.cloud":
			objectList := &v1alpha1.DeployItemList{}
			if err := removeCrdDeployItem(ctx, k8sClient, nextName, objectList, nextCrd); err != nil {
				return err
			}
		case "environments.landscaper.gardener.cloud":
			objectList := &v1alpha1.EnvironmentList{}
			if err := removeCrdEnvironment(ctx, k8sClient, nextName, objectList, nextCrd); err != nil {
				return err
			}
		case "executions.landscaper.gardener.cloud":
			objectList := &v1alpha1.ExecutionList{}
			if err := removeCrdExecution(ctx, k8sClient, nextName, objectList, nextCrd); err != nil {
				return err
			}
		case "installations.landscaper.gardener.cloud":
			objectList := &v1alpha1.InstallationList{}
			if err := removeCrdInstallation(ctx, k8sClient, nextName, objectList, nextCrd); err != nil {
				return err
			}
		case "lshealthchecks.landscaper.gardener.cloud":
			objectList := &v1alpha1.LsHealthCheckList{}
			if err := removeCrdLsHealthCheck(ctx, k8sClient, nextName, objectList, nextCrd); err != nil {
				return err
			}
		case "syncobjects.landscaper.gardener.cloud":
			objectList := &v1alpha1.SyncObjectList{}
			if err := removeCrdSyncObject(ctx, k8sClient, nextName, objectList, nextCrd); err != nil {
				return err
			}
		case "targets.landscaper.gardener.cloud":
			objectList := &v1alpha1.TargetList{}
			if err := removeCrdTarget(ctx, k8sClient, nextName, objectList, nextCrd); err != nil {
				return err
			}
		case "targetsyncs.landscaper.gardener.cloud":
			objectList := &v1alpha1.TargetSyncList{}
			if err := removeCrdTargetSync(ctx, k8sClient, nextName, objectList, nextCrd); err != nil {
				return err
			}
		default:
			found = false
		}

		if found {
			fmt.Println("Removing CRD: " + nextName)
			if err := k8sClient.Delete(ctx, nextCrd); err != nil {
				if !k8sErrors.IsNotFound(err) {
					return err
				}
			}
		}
	}

	return nil
}

func (o *uninstallOptions) deleteWebhook(ctx context.Context, k8sClient client.Client) error {
	webhookname := "landscaper-validation-webhook"

	for i := 0; i < 10; i++ {
		webhook := &v1.ValidatingWebhookConfiguration{}
		webhook.SetName(webhookname)
		if err := k8sClient.Delete(ctx, webhook); err != nil {
			if k8sErrors.IsNotFound(err) {
				return nil
			}
			return err
		}

		time.Sleep(time.Second * 2)
	}

	return fmt.Errorf("webhook could not be removed %s", webhookname)
}

func waitForRegistrationsRemoved(ctx context.Context, k8sClient client.Client) error {
	err := wait.PollUntilContextTimeout(ctx, 10*time.Second, 4*time.Minute, true, func(ctx context.Context) (done bool, err error) {
		fmt.Println("waiting for deployer registrations removed")

		deployerRegistrations := v1alpha1.DeployerRegistrationList{}
		if err := k8sClient.List(ctx, &deployerRegistrations); err != nil {
			return false, err
		}

		if len(deployerRegistrations.Items) == 0 {
			return true, nil
		}

		return false, nil
	})

	if err != nil {
		return fmt.Errorf("error while waiting for deployer registrations being removed: %w", err)
	}

	return nil
}
