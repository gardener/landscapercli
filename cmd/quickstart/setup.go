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
	"github.com/go-logr/logr"
	"github.com/spf13/cobra"
	"github.com/spf13/pflag"
	corev1 "k8s.io/api/core/v1"
	k8sErrors "k8s.io/apimachinery/pkg/api/errors"
	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/tools/clientcmd"
)

const (
	defaultNamespace = "landscaper"
)

type setupOptions struct {
	kubeconfigPath       string
	namespace            string
	setupOCIRegistry     bool
	landscaperValuesPath string
}

func NewSetupCommand(ctx context.Context) *cobra.Command {
	opts := &setupOptions{}
	cmd := &cobra.Command{
		Use:     "setup",
		Aliases: []string{"s"},
		Short:   "",
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

func (o *setupOptions) run(ctx context.Context, log logr.Logger) error {
	cfg, err := clientcmd.BuildConfigFromFlags("", o.kubeconfigPath)
	if err != nil {
		return err
	}

	k8sClient, err := kubernetes.NewForConfig(cfg)
	if err != nil {
		return err
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
				return err
			}
		} else {
			return err
		}
	}

	if o.setupOCIRegistry {
		fmt.Println("Installing OCI registry...")
		err = setupOCIRegistry(ctx, o.namespace, k8sClient)
		if err != nil {
			return err
		}
		fmt.Println("OCI registry installation succeeded!")
	}

	fmt.Println("Installing Landscaper...")
	err = setupLandscaper(ctx, o.kubeconfigPath, o.namespace, o.landscaperValuesPath)
	if err != nil {
		return err
	}
	fmt.Println("Landscaper installation succeeded!")

	return nil
}

func (o *setupOptions) Complete(args []string) error {
	return nil
}

func (o *setupOptions) AddFlags(fs *pflag.FlagSet) {
	fs.StringVar(&o.kubeconfigPath, "kubeconfig", "", "path to the kubeconfig of the target cluster")
	fs.StringVar(&o.namespace, "namespace", defaultNamespace, "namespace where landscaper and OCI registry are installed (default: "+defaultNamespace+")")
	fs.StringVar(&o.landscaperValuesPath, "landscaper-values", "", "path to values.yaml for the Landscaper Helm installation")
	fs.BoolVar(&o.setupOCIRegistry, "setup-oci-registry", false, "setup an internal OCI registry in the target cluster")
}

func setupLandscaper(ctx context.Context, kubeconfigPath, namespace, landscaperValues string) error {
	landscaperChartURI := "eu.gcr.io/gardener-project/landscaper/charts/landscaper-controller:v0.4.0-dev-203919cd11175450d6032552d116cab8c86023cc"
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

func setupOCIRegistry(ctx context.Context, namespace string, k8sClient kubernetes.Interface) error {
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
	fmt.Printf("Sucessfully executed: %s\n\n", string(command))

	return nil
}
