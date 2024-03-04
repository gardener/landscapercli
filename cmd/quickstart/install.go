package quickstart

import (
	"context"
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"path"
	"strings"
	"time"

	version2 "github.com/gardener/landscapercli/pkg/version"

	"k8s.io/apimachinery/pkg/util/wait"

	"sigs.k8s.io/yaml"

	"github.com/go-logr/logr"
	"github.com/spf13/cobra"
	"github.com/spf13/pflag"
	corev1 "k8s.io/api/core/v1"
	extv1 "k8s.io/apiextensions-apiserver/pkg/apis/apiextensions/v1"
	k8sErrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/tools/clientcmd"
	"sigs.k8s.io/controller-runtime/pkg/client"

	"github.com/gardener/landscapercli/pkg/logger"
	"github.com/gardener/landscapercli/pkg/util"
)

const (
	defaultNamespace = "landscaper"

	latestRelease = "latest release"

	installExample = `
landscaper-cli quickstart install --kubeconfig ./kubconfig.yaml

landscaper-cli quickstart install --kubeconfig ./kubconfig.yaml --landscaper-values ./landscaper-values.yaml --namespace landscaper --install-oci-registry --install-registry-ingress --registry-username testuser --registry-password some-pw
`
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
	landscaperValues    landscaperValues
}

type landscaperValues struct {
	Landscaper landscaperconfig `json:"landscaper"`
}

type landscaperconfig struct {
	Landscaper landscaper `json:"landscaper"`
}

type landscaper struct {
	RegistryConfig registryConfig `json:"registryConfig"`
	Deployers      []string       `json:"deployers,omitempty"`
}

type registryConfig struct {
	AllowPlainHttpRegistries bool    `json:"allowPlainHttpRegistries"`
	Secrets                  secrets `json:"secrets"`
}

type secrets struct {
	Defaults defaults `json:"default"`
}

type defaults struct {
	Auths map[string]interface{} `json:"auths"`
}

func NewInstallCommand(ctx context.Context) *cobra.Command {
	opts := &installOptions{}
	cmd := &cobra.Command{
		Use:     "install --kubeconfig [kubconfig.yaml] --landscaper-values [landscaper-values.yaml] --namespace landscaper --install-oci-registry --install-registry-ingress --registry-username testuser --registry-password some-pw",
		Aliases: []string{"i"},
		Short:   "command to install Landscaper (including Container, Helm, and Manifest deployers) in a target cluster. An OCI registry for testing can be optionally installed",
		Example: installExample,
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
	fs.StringVar(&o.namespace, "namespace", defaultNamespace, "namespace where Landscaper and the OCI registry will get installed (optional)")
	fs.StringVar(&o.landscaperValuesPath, "landscaper-values", "", "path to values.yaml for the Landscaper Helm installation (optional)")
	fs.BoolVar(&o.instOCIRegistry, "install-oci-registry", false, "install an OCI registry in the target cluster (optional)")
	fs.BoolVar(&o.instRegistryIngress, "install-registry-ingress", false, `install an ingress for accessing the OCI registry (optional). 
the credentials must be provided via the flags "--registry-username" and "--registry-password".
the Landscaper instance will then be automatically configured with these credentials.
prerequisites (!):
 - the target cluster must be a Gardener Shoot (TLS is provided via the Gardener cert manager)
 - a nginx ingress controller must be deployed in the target cluster
 - the command "htpasswd" must be installed on your local machine`)
	fs.StringVar(&o.landscaperChartVersion, "landscaper-chart-version", latestRelease,
		"use a custom Landscaper chart version (optional)")
	fs.StringVar(&o.registryUsername, "registry-username", "", "username for authenticating at the OCI registry (optional)")
	fs.StringVar(&o.registryPassword, "registry-password", "", "password for authenticating at the OCI registry (optional)")
}

func (o *installOptions) ReadLandscaperValues() error {
	content, err := os.ReadFile(o.landscaperValuesPath)
	if err != nil {
		return fmt.Errorf("cannot read file: %w", err)
	}

	unmarshaledContent := landscaperValues{}
	if err := yaml.Unmarshal(content, &unmarshaledContent); err != nil {
		return fmt.Errorf("cannot unmarshall file content: %w", err)
	}

	if unmarshaledContent.Landscaper.Landscaper.RegistryConfig.Secrets.Defaults.Auths == nil {
		unmarshaledContent.Landscaper.Landscaper.RegistryConfig.Secrets.Defaults.Auths = map[string]interface{}{}
	}

	o.landscaperValues = unmarshaledContent

	return nil
}

func (o *installOptions) run(ctx context.Context, log logr.Logger) error {
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

	if o.instOCIRegistry && o.instRegistryIngress {
		registryIngressHost := strings.Replace(cfg.Host, "https://api", "o.ingress", 1)
		if len(registryIngressHost) > 64 {
			return fmt.Errorf("no TLS certificate could be created because domain exceeds 64 characters: " + registryIngressHost)
		}
		o.registryIngressHost = registryIngressHost
	}

	if o.landscaperValuesPath != "" {
		if err := o.ReadLandscaperValues(); err != nil {
			return fmt.Errorf("cannot read landscaper values: %w", err)
		}
	}

	if err := o.checkConfiguration(); err != nil {
		return fmt.Errorf("invalid configuration: %w", err)
	}

	if err := o.createNamespace(ctx, k8sClient); err != nil {
		return err
	}

	if o.instOCIRegistry {
		if err := o.installOCIRegistry(ctx, k8sClient); err != nil {
			return fmt.Errorf("cannot install OCI registry: %w", err)
		}
	}

	version := o.landscaperChartVersion
	if version == latestRelease {
		version, err = version2.GetRelease()
		if err != nil {
			return fmt.Errorf("cannot get latest Landscaper release: %w", err)
		}
	}

	if err := o.installLandscaper(ctx, version); err != nil {
		return fmt.Errorf("cannot install landscaper: %w", err)
	}

	if err := o.waitForCrds(ctx); err != nil {
		return fmt.Errorf("waiting for crds failed: %w", err)
	}

	if err := o.installDeployer(ctx, "helm", version); err != nil {
		return fmt.Errorf("cannot install helm deployer: %w", err)
	}

	if err := o.installDeployer(ctx, "manifest", version); err != nil {
		return fmt.Errorf("cannot install manifest deployer: %w", err)
	}

	if err := o.installDeployer(ctx, "container", version); err != nil {
		return fmt.Errorf("cannot install container deployer: %w", err)
	}

	if o.instOCIRegistry {
		if o.instRegistryIngress {
			fmt.Println("The OCI registry can be accessed via the URL https://" + o.registryIngressHost)
			fmt.Println("It might take some minutes until the TLS certificate is created")
		} else {
			fmt.Println("The OCI registry can be accessed via kubectl port-forward <oci-registry-pod> 5000:5000")
		}
	}

	return nil
}

func (o *installOptions) createNamespace(ctx context.Context, k8sClient client.Client) error {
	fmt.Printf("Creating namespace %s\n", o.namespace)

	ns := &corev1.Namespace{
		ObjectMeta: metav1.ObjectMeta{
			Name: o.namespace,
		},
	}
	if err := k8sClient.Create(ctx, ns, &client.CreateOptions{}); err != nil {
		if k8sErrors.IsAlreadyExists(err) {
			fmt.Printf("Namespace %s already exists...Skipping\n\n", o.namespace)
		} else {
			return fmt.Errorf("cannot create namespace: %w", err)
		}
	} else {
		fmt.Printf("Namespace creation succeeded!\n\n")
	}

	return nil
}

func (o *installOptions) checkConfiguration() error {
	// check that the credentials for the OCI registry we will create are not set in the values from the user
	if o.instOCIRegistry && o.instRegistryIngress {
		registryAuths := o.landscaperValues.Landscaper.Landscaper.RegistryConfig.Secrets.Defaults.Auths
		_, ok := registryAuths[o.registryIngressHost]
		if ok {
			return fmt.Errorf("registry credentials for %s are already set in the provided Landscaper values. please remove them from your configuration as they will get configured automatically", o.registryIngressHost)
		}
	}

	allowHttpRegistries := o.landscaperValues.Landscaper.Landscaper.RegistryConfig.AllowPlainHttpRegistries
	allowHttpRegistriesPath := "landscaper.landscaper.registryConfig.allowPlainHttpRegistries"

	// if OCI registry is exposed via ingress, allowHttpRegistries must be false
	if o.instOCIRegistry && o.instRegistryIngress && allowHttpRegistries {
		return fmt.Errorf("%s must be set to false when installing Landscaper together with the OCI registry with ingress access", allowHttpRegistriesPath)
	}

	// if OCI registry is cluster-internal, allowHttpRegistries must be true
	if o.instOCIRegistry && !o.instRegistryIngress && !allowHttpRegistries {
		return fmt.Errorf("%s must be set to true when installing Landscaper together with the OCI registry without ingress access", allowHttpRegistriesPath)
	}

	return nil
}

func (o *installOptions) Complete(args []string) error {
	if !o.instOCIRegistry && o.instRegistryIngress {
		return fmt.Errorf("you can only set --install-registry-ingress to true together with --install-oci-registry")
	}

	if o.instOCIRegistry {
		if o.instRegistryIngress {
			if o.registryUsername == "" {
				return fmt.Errorf("username must be provided if --install-registry-ingress is set to true")
			}
			if o.registryPassword == "" {
				return fmt.Errorf("password must be provided if --install-registry-ingress is set to true")
			}
		}
	}
	return nil
}

func (o *installOptions) generateLandscaperValuesOverride() ([]byte, error) {

	landscaperValuesOverride := fmt.Sprintf(`
landscaper:
  landscaper:
    deployers: []
    deployerManagement:
      disable: true
      namespace: %s
      agent:
        disable: true
`, o.namespace)

	if o.instRegistryIngress {
		// when installing the ingress, we must add the registry credentials to the Landscaper values file
		// also, we must keep any other set of credentials that might have been configured for another registry

		credentials := fmt.Sprintf("%s:%s", o.registryUsername, o.registryPassword)
		encodedCredentials := base64.StdEncoding.EncodeToString([]byte(credentials))

		if o.landscaperValues.Landscaper.Landscaper.RegistryConfig.Secrets.Defaults.Auths == nil {
			o.landscaperValues.Landscaper.Landscaper.RegistryConfig.Secrets.Defaults.Auths = map[string]interface{}{}
		}
		registryAuths := o.landscaperValues.Landscaper.Landscaper.RegistryConfig.Secrets.Defaults.Auths
		registryAuths[o.registryIngressHost] = map[string]interface{}{
			"auth": encodedCredentials,
		}

		marshaledRegistryAuths, err := json.Marshal(registryAuths)
		if err != nil {
			return nil, fmt.Errorf("cannot marshal registry auths: %w", err)
		}

		landscaperValuesOverride = fmt.Sprintf(`%s
    registryConfig:
      secrets:
        default: {
          "auths": %s
        }
`, landscaperValuesOverride, string(marshaledRegistryAuths))
	}

	return []byte(landscaperValuesOverride), nil
}

func (o *installOptions) installLandscaper(ctx context.Context, version string) error {
	fmt.Println("Installing Landscaper")

	tempDir, err := os.MkdirTemp(".", "landscaper-chart-tmp-*")
	if err != nil {
		return err
	}
	defer func() {
		if err := os.RemoveAll(tempDir); err != nil {
			fmt.Printf("cannot remove temporary directory %s: %s\n", tempDir, err.Error())
		}
	}()

	landscaperChartURI := fmt.Sprintf("oci://europe-docker.pkg.dev/sap-gcp-cp-k8s-stable-hub/landscaper/github.com/gardener/landscaper/charts/landscaper --untar --version %s", version)
	pullCmd := fmt.Sprintf("helm pull %s -d %s", landscaperChartURI, tempDir)
	if err := util.ExecCommandBlocking(pullCmd); err != nil {
		return err
	}

	fileInfos, err := os.ReadDir(tempDir)
	if err != nil {
		return err
	}

	if len(fileInfos) != 1 {
		return errors.New("found more than 1 item in temp directory for Helm Chart export")
	}
	chartPath := path.Join(tempDir, fileInfos[0].Name())

	landscaperValuesOverride, err := o.generateLandscaperValuesOverride()
	if err != nil {
		return fmt.Errorf("error generating landscaper values override: %w", err)
	}

	tmpFile, err := os.CreateTemp(".", "landscaper-values-override-")
	if err != nil {
		return fmt.Errorf("cannot create temporary file: %w", err)
	}
	defer func() {
		if err := os.Remove(tmpFile.Name()); err != nil {
			fmt.Printf("cannot remove temporary file %s: %s\n", tmpFile.Name(), err.Error())
		}
	}()

	if err := os.WriteFile(tmpFile.Name(), landscaperValuesOverride, os.ModePerm); err != nil {
		return fmt.Errorf("cannot write to file: %w", err)
	}

	installCommand := fmt.Sprintf(
		"helm upgrade --install --namespace %s landscaper %s --kubeconfig %s -f %s -f %s",
		o.namespace,
		chartPath,
		o.kubeconfigPath,
		o.landscaperValuesPath,
		tmpFile.Name(),
	)

	if err := util.ExecCommandBlocking(installCommand); err != nil {
		return err
	}

	fmt.Printf("Landscaper installation succeeded!\n\n")

	return nil
}

func (o *installOptions) installDeployer(ctx context.Context, deployer, version string) error {
	fmt.Printf("Installing %s deployer\n", deployer)

	tempDir, err := os.MkdirTemp(".", fmt.Sprintf("%s-deployer-chart-tmp-*", deployer))
	if err != nil {
		return err
	}
	defer func() {
		if err := os.RemoveAll(tempDir); err != nil {
			fmt.Printf("cannot remove temporary directory %s: %s\n", tempDir, err.Error())
		}
	}()

	landscaperChartURI := fmt.Sprintf("oci://europe-docker.pkg.dev/sap-gcp-cp-k8s-stable-hub/landscaper/github.com/gardener/landscaper/%s-deployer/charts/%s-deployer --untar --version %s",
		deployer, deployer, version)
	pullCmd := fmt.Sprintf("helm pull %s -d %s", landscaperChartURI, tempDir)
	if err := util.ExecCommandBlocking(pullCmd); err != nil {
		return err
	}

	fileInfos, err := os.ReadDir(tempDir)
	if err != nil {
		return err
	}

	if len(fileInfos) != 1 {
		return fmt.Errorf("found more than 1 item in temp directory for %s Chart export", deployer)
	}
	chartPath := path.Join(tempDir, fileInfos[0].Name())

	installCommand := fmt.Sprintf(
		"helm upgrade --install --namespace %s %s-deployer %s --kubeconfig %s",
		o.namespace,
		deployer,
		chartPath,
		o.kubeconfigPath,
	)

	if err := util.ExecCommandBlocking(installCommand); err != nil {
		return err
	}

	fmt.Printf("%s installation succeeded!\n\n", deployer)

	return nil
}

func (o *installOptions) installOCIRegistry(ctx context.Context, k8sClient client.Client) error {
	fmt.Println("Installing OCI registry")

	ociRegistryOpts := &ociRegistryOpts{
		namespace:      o.namespace,
		installIngress: o.instRegistryIngress,
		ingressHost:    o.registryIngressHost,
		username:       o.registryUsername,
		password:       o.registryPassword,
	}
	ociRegistry := NewOCIRegistry(ociRegistryOpts, k8sClient)

	if err := ociRegistry.install(ctx); err != nil {
		return err
	}

	fmt.Print("OCI registry installation succeeded!\n\n")
	return nil
}

func (o *installOptions) waitForCrds(ctx context.Context) error {
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

	err = wait.PollUntilContextTimeout(ctx, 1*time.Second, 1*time.Minute, true, func(ctx context.Context) (done bool, err error) {
		fmt.Println("waiting for CRDs")

		crdName := "contexts.landscaper.gardener.cloud"
		if crdExists := o.crdExists(ctx, k8sClient, crdName); !crdExists {
			return false, nil
		}

		crdName = "dataobjects.landscaper.gardener.cloud"
		if crdExists := o.crdExists(ctx, k8sClient, crdName); !crdExists {
			return false, nil
		}

		crdName = "deployitems.landscaper.gardener.cloud"
		if crdExists := o.crdExists(ctx, k8sClient, crdName); !crdExists {
			return false, nil
		}

		crdName = "targets.landscaper.gardener.cloud"
		if crdExists := o.crdExists(ctx, k8sClient, crdName); !crdExists {
			return false, nil
		}

		return true, nil
	})

	return err
}

func (o *installOptions) crdExists(ctx context.Context, k8sClient client.Client, crdName string) bool {
	crd := &extv1.CustomResourceDefinition{}
	crd.SetName(crdName)

	if err := k8sClient.Get(ctx, client.ObjectKeyFromObject(crd), crd); err != nil {
		fmt.Printf("Crd not found: %s\n\n", crdName)
		return false
	}

	return true
}
