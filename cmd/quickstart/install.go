package quickstart

import (
	"context"
	"encoding/base64"
	"errors"
	"fmt"
	"io/ioutil"
	"os"
	"path"
	"strings"

	"github.com/gardener/landscapercli/pkg/version"

	"github.com/go-logr/logr"
	"github.com/spf13/cobra"
	"github.com/spf13/pflag"
	corev1 "k8s.io/api/core/v1"
	k8sErrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/tools/clientcmd"
	"sigs.k8s.io/controller-runtime/pkg/client"

	"github.com/gardener/landscapercli/pkg/logger"
	"github.com/gardener/landscapercli/pkg/util"
)

const (
	defaultNamespace = "landscaper"
)

type installOptions struct {
	// CLI flags
	kubeconfigPath         string
	namespace              string
	instOCIRegistry        bool
	instRegistryIngress    bool
	landscaperValuesPath   string
	landscaperChartVersion string
	registryUsername       string
	registryPassword       string

	// set during execution
	registryIngressHost string
	landscaperValues    map[string]interface{}
}

func NewInstallCommand(ctx context.Context) *cobra.Command {
	opts := &installOptions{}
	cmd := &cobra.Command{
		Use:     "install --kubeconfig [kubconfig.yaml] [--install-oci-registry]",
		Aliases: []string{"i"},
		Short:   "command to install Landscaper (and optionally an OCI registry) in a target cluster",
		Example: "landscaper-cli quickstart install --kubeconfig ./kubconfig.yaml --install-oci-registry --landscaper-values ./landscaper-values.yaml --namespace landscaper",
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

func (o *installOptions) AddFlags(fs *pflag.FlagSet) {
	fs.StringVar(&o.kubeconfigPath, "kubeconfig", "", "path to the kubeconfig of the target cluster")
	fs.StringVar(&o.namespace, "namespace", defaultNamespace, "namespace where Landscaper and the OCI registry will get installed")
	fs.StringVar(&o.landscaperValuesPath, "landscaper-values", "", "path to values.yaml for the Landscaper Helm installation")
	fs.BoolVar(&o.instOCIRegistry, "install-oci-registry", false, "install an OCI registry in the target cluster")
	fs.BoolVar(&o.instRegistryIngress, "install-registry-ingress", false, `install an ingress for accessing the OCI registry without port-forwarding. 
the credentials must be provided via the flags "--registry-username" and "--registry-password".
the Landscaper instance will then be automatically configured with these credentials.
prerequisites (!):
 - the target cluster must be a Gardener Shoot as TLS is provided via the Gardener cert manager
 - a nginx ingress controller must be deployed in the target cluster
 - "htpasswd" must be installed on your local machine`)
	fs.StringVar(&o.landscaperChartVersion, "landscaper-chart-version", version.LandscaperChartVersion,
		"use a custom Landscaper chart version (corresponds to Landscaper Github release with the same version number)")
	fs.StringVar(&o.registryUsername, "registry-username", "", "username for authenticating at the OCI registry")
	fs.StringVar(&o.registryPassword, "registry-password", "", "password for authenticating at the OCI registry")
}

func (o *installOptions) run(ctx context.Context, log logr.Logger) error {
	cfg, err := clientcmd.BuildConfigFromFlags("", o.kubeconfigPath)
	if err != nil {
		return fmt.Errorf("Cannot parse K8s config: %w", err)
	}

	k8sClient, err := client.New(cfg, client.Options{
		Scheme: scheme,
	})
	if err != nil {
		return fmt.Errorf("cannot build K8s client: %w", err)
	}

	landscaperValues, err := util.ReadYamlFile(o.landscaperValuesPath)
	if err != nil {
		return fmt.Errorf("cannot build K8s client: %w", err)
	}
	o.landscaperValues = landscaperValues

	fmt.Printf("Creating namespace %s", o.namespace)
	ns := &corev1.Namespace{
		ObjectMeta: metav1.ObjectMeta{
			Name: o.namespace,
		},
	}
	err = k8sClient.Create(ctx, ns, &client.CreateOptions{})
	if err != nil {
		if k8sErrors.IsAlreadyExists(err) {
			fmt.Printf("Namespace %s already exists...Skipping\n", o.namespace)
		} else {
			return fmt.Errorf("Cannot create namespace: %w", err)
		}
	}

	if o.instOCIRegistry {
		registryIngressHost := strings.Replace(cfg.Host, "https://api", "o.ingress", 1)
		if len(registryIngressHost) > 64 {
			return fmt.Errorf("No certificate could be created because your domain exceeds 64 characters: " + registryIngressHost)
		}
		o.registryIngressHost = registryIngressHost

		fmt.Println("Installing OCI registry...")
		err = o.installOCIRegistry(ctx, k8sClient)
		if err != nil {
			return fmt.Errorf("Cannot install OCI registry: %w", err)
		}
		fmt.Print("OCI registry installation succeeded!\n\n")
	}

	fmt.Println("Installing Landscaper...")
	err = o.installLandscaper(ctx)
	if err != nil {
		return fmt.Errorf("Cannot install landscaper: %w", err)
	}
	fmt.Printf("Landscaper installation succeeded!\n\n")

	err = o.runPostInstallChecks(log)
	if err != nil {
		msg := fmt.Sprintf("Found problems during post-install checks: %s", err)
		log.V(util.LogLevelWarning).Info(msg)
	}

	fmt.Println("Installation succeeded")

	if o.instOCIRegistry {
		if o.instRegistryIngress {
			fmt.Println("You can access the OCI registry via the URL https://" + o.registryIngressHost)
			fmt.Println("It might take some minutes until the TLS certificate is created")
		} else {
			fmt.Println("You can access the OCI registry via kubectl port-forward <oci-registry-pod> 5000:5000")
		}
	}

	return nil
}

func (o *installOptions) readValues() error {
	return nil
}

func (o *installOptions) runPostInstallChecks(log logr.Logger) error {
	var allowPlainHttpRegistries, ok bool
	if o.landscaperValuesPath != "" {
		valuePath := "landscaper.registryConfig.allowPlainHttpRegistries"
		val, err := util.GetValueFromNestedMap(o.landscaperValues, valuePath)
		if err != nil {
			return fmt.Errorf("Cannot access Helm values: %w", err)
		}

		allowPlainHttpRegistries, ok = val.(bool)
		if !ok {
			return fmt.Errorf("Value for path %s must be of type boolean", valuePath)
		}
	}

	if o.instOCIRegistry && !allowPlainHttpRegistries {
		log.V(util.LogLevelWarning).Info("You installed the cluster-internal OCI registry without enabling support for HTTP-only registries on the Landscaper.")
		log.V(util.LogLevelWarning).Info("For using the landscaper with the cluster-internal OCI registry, please set landscaper.registryConfig.allowPlainHttpRegistries in the Landscaper Helm values to true.")
	}

	return nil
}

func (o *installOptions) Complete(args []string) error {
	if o.instRegistryIngress == true {
		if o.registryUsername == "" {
			return fmt.Errorf("username must be provided if --install-ingress is set to true")
		}
		if o.registryPassword == "" {
			return fmt.Errorf("password must be provided if --install-ingress is set to true")
		}
	}
	return nil
}

func (o *installOptions) installLandscaper(ctx context.Context) error {
	landscaperChartURI := fmt.Sprintf("eu.gcr.io/gardener-project/landscaper/charts/landscaper-controller:%s", o.landscaperChartVersion)
	pullCmd := fmt.Sprintf("helm chart pull %s", landscaperChartURI)
	err := util.ExecCommandBlocking(pullCmd)
	if err != nil {
		return err
	}

	tempDir, err := ioutil.TempDir(".", "landscaper-chart-tmp-*")
	if err != nil {
		return err
	}
	defer func() {
		err := os.RemoveAll(tempDir)
		if err != nil {
			fmt.Printf("cannot remove temporary directory %s: %s", tempDir, err.Error())
		}
	}()

	exportCmd := fmt.Sprintf("helm chart export %s -d %s", landscaperChartURI, tempDir)
	err = util.ExecCommandBlocking(exportCmd)
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

	installCommand := fmt.Sprintf("helm upgrade --install --namespace %s landscaper %s --kubeconfig %s -f %s", o.namespace, chartPath, o.kubeconfigPath, o.landscaperValuesPath)

	if o.instRegistryIngress {
		credentials := fmt.Sprintf("%s:%s", o.registryUsername, o.registryPassword)
		encodedCredentials := base64.StdEncoding.EncodeToString([]byte(credentials))

		valuePath := "landscaper.registryConfig.secrets"
		val, err := util.GetValueFromNestedMap(o.landscaperValues, valuePath)
		fmt.Println(val)

		landscaperValuesOverride := fmt.Sprintf(`
landscaper:
  registryConfig:
    allowPlainHttpRegistries: false
    secrets: 
      default: {
        "auths": {
          "%s":  {
            "auth": "%s"
          }
        }
      }
`, o.registryIngressHost, encodedCredentials)

		tmpFile, err := ioutil.TempFile(".", "landscaper-values-override-")
		if err != nil {
			return fmt.Errorf("cannot create temporary file: %w", err)
		}
		defer func() {
			err := os.Remove(tmpFile.Name())
			if err != nil {
				fmt.Printf("cannot remove temporary file %s: %s", tmpFile.Name(), err.Error())
			}
		}()

		err = ioutil.WriteFile(tmpFile.Name(), []byte(landscaperValuesOverride), os.ModePerm)
		if err != nil {
			return fmt.Errorf("cannot write to file: %w", err)
		}

		installCommand = fmt.Sprintf("%s -f %s", installCommand, tmpFile.Name())
	}

	err = util.ExecCommandBlocking(installCommand)
	if err != nil {
		return err
	}

	return nil
}

func (o *installOptions) installOCIRegistry(ctx context.Context, k8sClient client.Client) error {
	ociRegistryOpts := &ociRegistryOpts{
		namespace:      o.namespace,
		installIngress: o.instRegistryIngress,
		ingressHost:    o.registryIngressHost,
		username:       o.registryUsername,
		password:       o.registryPassword,
	}
	ociRegistry := NewOCIRegistry(ociRegistryOpts, k8sClient)
	return ociRegistry.install(ctx)
}
