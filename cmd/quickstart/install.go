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

	"sigs.k8s.io/yaml"

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
		Example: "landscaper-cli quickstart install --kubeconfig ./kubconfig.yaml --landscaper-values ./landscaper-values.yaml --namespace landscaper --install-oci-registry --install-registry-ingress --registry-username testuser --registry-password some-pw",
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
	fs.BoolVar(&o.instRegistryIngress, "install-registry-ingress", false, `install an ingress for accessing the OCI registry. 
the credentials must be provided via the flags "--registry-username" and "--registry-password".
the Landscaper instance will then be automatically configured with these credentials.
prerequisites (!):
 - the target cluster must be a Gardener Shoot (TLS is provided via the Gardener cert manager)
 - a nginx ingress controller must be deployed in the target cluster
 - the command "htpasswd" must be installed on your local machine`)
	fs.StringVar(&o.landscaperChartVersion, "landscaper-chart-version", version.LandscaperChartVersion,
		"use a custom Landscaper chart version (corresponds to Landscaper Github release with the same version number)")
	fs.StringVar(&o.registryUsername, "registry-username", "", "username for authenticating at the OCI registry")
	fs.StringVar(&o.registryPassword, "registry-password", "", "password for authenticating at the OCI registry")
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

	if err := o.installLandscaper(ctx); err != nil {
		return fmt.Errorf("cannot install landscaper: %w", err)
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
	defaultDeployers := ""
	if len(o.landscaperValues.Landscaper.Landscaper.Deployers) == 0 {
		defaultDeployers = `
  deployers:
  - container
  - helm
  - manifest`
	}

	landscaperValuesOverride := fmt.Sprintf(`
landscaper:%s
  landscaper:
    deployerManagement:
      namespace: %s
      agent:
        namespace: %s
`, defaultDeployers, o.namespace, o.namespace)

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

func (o *installOptions) installLandscaper(ctx context.Context) error {
	fmt.Println("Installing Landscaper")

	tempDir, err := os.MkdirTemp(".", "landscaper-chart-tmp-*")
	if err != nil {
		return err
	}
	defer func() {
		if err := os.RemoveAll(tempDir); err != nil {
			fmt.Printf("cannot remove temporary directory %s: %s", tempDir, err.Error())
		}
	}()

	landscaperChartURI := fmt.Sprintf("oci://eu.gcr.io/gardener-project/landscaper/charts/landscaper --untar --version %s", o.landscaperChartVersion)
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
