package quickstart

import (
	"context"
	"errors"
	"fmt"
	"io/ioutil"
	"os"
	"os/exec"
	"path"
	"strings"

	"github.com/gardener/landscapercli/pkg/logger"
	"github.com/gardener/landscapercli/pkg/util"
	"github.com/go-logr/logr"
	"github.com/spf13/cobra"
	"github.com/spf13/pflag"
	corev1 "k8s.io/api/core/v1"
	k8sErrors "k8s.io/apimachinery/pkg/api/errors"
	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/tools/clientcmd"
	"sigs.k8s.io/yaml"
)

const (
	defaultNamespace = "landscaper"
	defaultLandscaperChartVersion = "v0.4.0"
)

type installOptions struct {
	kubeconfigPath         string
	namespace              string
	installOCIRegistry     bool
	landscaperValuesPath   string
	landscaperChartVersion string
}

func NewInstallCommand(ctx context.Context) *cobra.Command {
	opts := &installOptions{}
	cmd := &cobra.Command{
		Use:     "install",
		Aliases: []string{"i"},
		Short:   "command to install the landscaper (and optionally an OCI registry) in a target cluster",
		Example: "landscaper-cli quickstart install --kubeconfig ./kubconfig.yaml --landscaper-values ./landscaper-values.yaml --namespace landscaper --install-oci-registry",
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

func (o *installOptions) run(ctx context.Context, log logr.Logger) error {
	cfg, err := clientcmd.BuildConfigFromFlags("", o.kubeconfigPath)
	if err != nil {
		return fmt.Errorf("Cannot parse K8s config: %w", err)
	}

	k8sClient, err := kubernetes.NewForConfig(cfg)
	if err != nil {
		return fmt.Errorf("Cannot build K8s clientset: %w", err)
	}

	_, err = k8sClient.CoreV1().Namespaces().Get(ctx, o.namespace, v1.GetOptions{})
	if err != nil {
		if k8sErrors.IsNotFound(err) {
			namespace := &corev1.Namespace{
				ObjectMeta: v1.ObjectMeta{
					Name: o.namespace,
				},
			}
			_, err = k8sClient.CoreV1().Namespaces().Create(ctx, namespace, v1.CreateOptions{})
			if err != nil {
				return fmt.Errorf("Cannot create namespace %s: %w", o.namespace, err)
			}
		} else {
			return fmt.Errorf("Cannot get namespace %s: %w", o.namespace, err)
		}
	}

	if o.installOCIRegistry {
		fmt.Println("Installing OCI registry...")
		err = installOCIRegistry(ctx, o.namespace, k8sClient)
		if err != nil {
			return fmt.Errorf("Cannot install OCI registry: %w", err)
		}
		fmt.Print("OCI registry installation succeeded!\n\n")
	}

	fmt.Println("Installing Landscaper...")
	err = installLandscaper(ctx, o.kubeconfigPath, o.namespace, o.landscaperValuesPath, o.landscaperChartVersion)
	if err != nil {
		return fmt.Errorf("Cannot install landscaper: %w", err)
	}
	fmt.Println("Landscaper installation succeeded!")

	err = runPostInstallChecks(o, log)
	if err != nil {
		msg := fmt.Sprintf("Found problems during post-install checks: %s", err)
		log.V(util.LogLevelWarning).Info(msg)
	}

	return nil
}

func runPostInstallChecks(opts *installOptions, log logr.Logger) error {
	var allowPlainHttpRegistries, ok bool
	if opts.landscaperValuesPath != "" {
		valuesFileContent, err := ioutil.ReadFile(opts.landscaperValuesPath)
		if err != nil {
			return fmt.Errorf("Cannot read file: %w", err)
		}

		values := map[string]interface{}{}
		err = yaml.UnmarshalStrict(valuesFileContent, &values)
		if err != nil {
			return fmt.Errorf("Cannot unmarshall values.yaml: %w", err)
		}

		valuePath := "landscaper.registryConfig.allowPlainHttpRegistries"
		val, err := util.GetValueFromNestedMap(values, valuePath)
		if err != nil {
			return fmt.Errorf("Cannot access Helm values: %w", err)
		}

		allowPlainHttpRegistries, ok = val.(bool)
		if !ok {
			return fmt.Errorf("Value for path %s must be of type boolean", valuePath)
		}
	}

	if opts.installOCIRegistry && !allowPlainHttpRegistries {
		log.V(util.LogLevelWarning).Info("You installed the cluster-internal OCI registry without enabling support for HTTP-only registries on the Landscaper.")
		log.V(util.LogLevelWarning).Info("For using the landscaper with the cluster-internal OCI registry, please set landscaper.registryConfig.allowPlainHttpRegistries in the Landscaper Helm values to true.")
	}

	return nil
}

func (o *installOptions) Complete(args []string) error {
	return nil
}

func (o *installOptions) AddFlags(fs *pflag.FlagSet) {
	fs.StringVar(&o.kubeconfigPath, "kubeconfig", "", "path to the kubeconfig of the target cluster")
	fs.StringVar(&o.namespace, "namespace", defaultNamespace, "namespace where the landscaper and the OCI registry are installed")
	fs.StringVar(&o.landscaperValuesPath, "landscaper-values", "", "path to values.yaml for the landscaper Helm installation")
	fs.BoolVar(&o.installOCIRegistry, "install-oci-registry", false, "install an internal OCI registry in the target cluster")
	fs.StringVar(&o.landscaperChartVersion, "landscaper-chart-version", defaultLandscaperChartVersion, "use a custom landscaper chart version")
}

func installLandscaper(ctx context.Context, kubeconfigPath, namespace, landscaperValues string, landscaperChartVersion string) error {
	landscaperChartURI := fmt.Sprintf("eu.gcr.io/gardener-project/landscaper/charts/landscaper-controller:%s", landscaperChartVersion)
	pullCmd := fmt.Sprintf("helm chart pull %s", landscaperChartURI)
	err := execute(pullCmd)
	if err != nil {
		return err
	}

	tempDir, err := ioutil.TempDir(".", "landscaper-chart-tmp-*")
	defer func() {
		err = os.RemoveAll(tempDir)
		if err != nil {
			fmt.Printf("cannot remove temporary directory %s: %s", tempDir, err.Error())
		}
	}()

	exportCmd := fmt.Sprintf("helm chart export %s -d %s", landscaperChartURI, tempDir)
	err = execute(exportCmd)
	if err != nil {
		return err
	}

	fileInfos, err := ioutil.ReadDir(tempDir)
	if err != nil {
		return err
	}

	if len(fileInfos) != 1 {
		return errors.New("found more than 1 item in temp directory for Helm Chart export")
	}

	chartPath := path.Join(tempDir, fileInfos[0].Name())

	err = execute(fmt.Sprintf("helm upgrade --install --namespace %s landscaper %s -f %s --kubeconfig %s", namespace, chartPath, landscaperValues, kubeconfigPath))
	if err != nil {
		return err
	}

	return nil
}

func installOCIRegistry(ctx context.Context, namespace string, k8sClient kubernetes.Interface) error {
	ociRegistry := NewOCIRegistry(namespace, k8sClient)
	return ociRegistry.install(ctx)
}

func execute(command string) error {
	fmt.Printf("Executing: %s\n", command)

	arr := strings.Split(command, " ")

	cmd := exec.Command(arr[0], arr[1:]...)
	cmd.Env = []string{"HELM_EXPERIMENTAL_OCI=1", "HOME=" + os.Getenv("HOME"), "PATH=" + os.Getenv("PATH")}
	out, err := cmd.CombinedOutput()

	if err != nil {
		fmt.Printf("Failed with error: %s:\n%s\n", err, string(out))
		return err
	}
	fmt.Println("Executed sucessfully!")

	return nil
}
