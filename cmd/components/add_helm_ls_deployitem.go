// SPDX-FileCopyrightText: 2020 SAP SE or an SAP affiliate company and Gardener contributors.
//
// SPDX-License-Identifier: Apache-2.0

package components

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"text/template"

	"github.com/gardener/component-cli/pkg/commands/componentarchive/input"
	cdresources "github.com/gardener/component-cli/pkg/commands/componentarchive/resources"
	cd "github.com/gardener/component-spec/bindings-go/apis/v2"
	"github.com/go-logr/logr"
	"github.com/spf13/cobra"
	"github.com/spf13/pflag"

	"github.com/gardener/landscapercli/pkg/blueprints"
	"github.com/gardener/landscapercli/pkg/components"
	"github.com/gardener/landscapercli/pkg/logger"
	"github.com/gardener/landscapercli/pkg/util"
)

const addHelmLSDeployItemUse = `deployitem \
    [deployitem name] \
   `

const addHelmLSDeployItemExample = `
landscaper-cli component add helm-ls deployitem \
  nginx \
  --component-directory .../myComponent \
  --oci-reference eu.gcr.io/gardener-project/landscaper/tutorials/charts/ingress-nginx:v0.1.0 \
  --resource-version v0.1.0 \
  --cluster-param target-cluster \
  --target-ns-param target-namespace

or 

landscaper-cli component add helm-ls deployitem \
  nginx \
  --component-directory .../myComponent \
  --chart-directory .../charts/echo-server \
  --resource-version v0.1.0 \
  --cluster-param target-cluster \
  --target-ns-param target-namespace
`

const addHelmLSDeployItemShort = `
Command to add a deploy item skeleton to the blueprint of a component`

const helm = "helm"

var identityKeyValidationRegexp = regexp.MustCompile("^[a-z0-9]([-_+a-z0-9]*[a-z0-9])?$")

type addHelmLsDeployItemOptions struct {
	componentPath  string
	deployItemName string

	ociReference       string
	chartDirectoryPath string

	resourceVersion string

	clusterParam  string
	targetNsParam string
}

// NewCreateCommand creates a new blueprint command to create a blueprint
func NewAddHelmLSDeployItemCommand(ctx context.Context) *cobra.Command {
	opts := &addHelmLsDeployItemOptions{}
	cmd := &cobra.Command{
		Use:     addHelmLSDeployItemUse,
		Example: addHelmLSDeployItemExample,
		Short:   addHelmLSDeployItemShort,
		Args:    cobra.ExactArgs(1),

		Run: func(cmd *cobra.Command, args []string) {
			if err := opts.Complete(args); err != nil {
				fmt.Println(err.Error())
				os.Exit(1)
			}

			if err := opts.run(ctx, logger.Log); err != nil {
				fmt.Println(err.Error())
				os.Exit(1)
			}

			fmt.Printf("Deploy item added")
			fmt.Printf("  \n- deploy item definition in blueprint folder in file %s created", util.ExecutionFileName(opts.deployItemName))
			fmt.Printf("  \n- file reference to deploy item definition added to blueprint")
			fmt.Printf("  \n- import definitions added to blueprint")
			fmt.Printf("  \n- helm chart resource added to resources.yaml")
		},
	}

	opts.AddFlags(cmd.Flags())

	return cmd
}

func (o *addHelmLsDeployItemOptions) Complete(args []string) error {
	o.deployItemName = args[0]

	err:= o.validate()
	if err != nil {
		return err
	}

	if o.chartDirectoryPath != "" {
		o.chartDirectoryPath = filepath.Dir(o.chartDirectoryPath)
	}

	return nil
}

func (o *addHelmLsDeployItemOptions) AddFlags(fs *pflag.FlagSet) {
	fs.StringVar(&o.componentPath,
		"component-directory",
		".",
		"path to component directory")
	fs.StringVar(&o.ociReference,
		"oci-reference",
		"",
		"reference to oci artifact containing the helm chart")
	fs.StringVar(&o.chartDirectoryPath,
		"chart-directory",
		"",
		"path to chart directory")
	fs.StringVar(&o.resourceVersion,
		"resource-version",
		"",
		"helm chart version")
	fs.StringVar(&o.clusterParam,
		"cluster-param",
		"targetCluster",
		"import parameter name for the target resource containing the access data of the target cluster")
	fs.StringVar(&o.targetNsParam,
		"target-ns-param",
		"",
		"target namespace")
}

func (o *addHelmLsDeployItemOptions) validate() error {
	if !identityKeyValidationRegexp.Match([]byte(o.deployItemName)) {
		return fmt.Errorf("the deploy item name must consist of lower case alphanumeric characters, '-', '_' " +
			"or '+', and must start and end with an alphanumeric character")
	}

	if o.ociReference == "" && o.chartDirectoryPath == "" {
		return fmt.Errorf("oci-reference and chart-directory not set, exactly one needs to be specified")
	}

	if o.ociReference != "" && o.chartDirectoryPath != "" {
		return fmt.Errorf("both oci-reference and chart-directory are set, exactly one needs to be specified")
	}

	if o.resourceVersion == "" {
		return fmt.Errorf("resource-version is missing")
	}

	if o.clusterParam == "" {
		return fmt.Errorf("cluster-param is missing")
	}

	if o.targetNsParam == "" {
		return fmt.Errorf("target-ns-param is missing")
	}

	return nil
}

func (o *addHelmLsDeployItemOptions) run(ctx context.Context, log logr.Logger) error {
	blueprintPath := util.BlueprintDirectoryPath(o.componentPath)
	blueprint, err := blueprints.NewBlueprintReader(blueprintPath).Read()
	if err != nil {
		return err
	}

	err = o.addResource()
	if err != nil {
		return err
	}

	blueprintBuilder := blueprints.NewBlueprintBuilder(blueprint)

	if blueprintBuilder.ExistsDeployExecution(o.deployItemName) {
		return fmt.Errorf("The blueprint already contains a deploy item %s\n", o.deployItemName)
	}

	exists, err := o.existsExecutionFile()
	if err != nil {
		return err
	}
	if exists {
		return fmt.Errorf("Deploy execution file %s already exists\n", util.ExecutionFilePath(o.componentPath, o.deployItemName))
	}

	err = o.createExecutionFile()
	if err != nil {
		return err
	}

	blueprintBuilder.AddDeployExecution(o.deployItemName)
	blueprintBuilder.AddImportForTarget(o.clusterParam)
	blueprintBuilder.AddImportForElementaryType(o.targetNsParam, "string")

	return blueprints.NewBlueprintWriter(blueprintPath).Write(blueprint)
}

func (o *addHelmLsDeployItemOptions) addResource() error {
	resourceReader := components.NewResourceReader(o.componentPath)

	resources, err := resourceReader.Read()
	if err != nil {
		return err
	}

	resource, err := o.createResources()
	if err != nil {
		return err
	}

	resources = append(resources, *resource)

	resourceWriter := components.NewResourceWriter(o.componentPath)
	err = resourceWriter.Write(resources)

	return err
}

func (o *addHelmLsDeployItemOptions) existsExecutionFile() (bool, error) {
	fileInfo, err := os.Stat(util.ExecutionFilePath(o.componentPath, o.deployItemName))
	if err != nil {
		if os.IsNotExist(err) {
			return false, nil
		}
		return false, err
	}

	if fileInfo.IsDir() {
		return false, fmt.Errorf("There already exists a directory %s\n", util.ExecutionFileName(o.deployItemName))
	}

	return true, nil
}

func (o *addHelmLsDeployItemOptions) createExecutionFile() error {
	f, err := os.Create(util.ExecutionFilePath(o.componentPath, o.deployItemName))
	if err != nil {
		return err
	}

	defer f.Close()

	err = o.writeExecution(f)

	return err
}

const executionTemplateExternalRef = `deployItems:
- name: {{.DeployItemName}}
  type: landscaper.gardener.cloud/helm
  target:
    name: {{.TargetNameExpression}}
    namespace: {{.TargetNamespaceExpression}}
  config:
    apiVersion: helm.deployer.landscaper.gardener.cloud/v1alpha1
    kind: ProviderConfiguration

    chart:
      ref: {{"{{"}} with (getResource .cd "name" "{{.DeployItemName}}-chart") {{"}}"}} {{"{{"}} .access.imageReference {{"}}"}} {{"{{"}} end {{"}}"}}

    updateStrategy: patch

    name: {{.DeployItemName}}
    namespace: {{.ApplicationNamespaceExpression}}
`

const executionTemplateLocally = `deployItems:
- name: {{.DeployItemName}}
  type: landscaper.gardener.cloud/helm
  target:
    name: {{.TargetNameExpression}}
    namespace: {{.TargetNamespaceExpression}}
  config:
    apiVersion: helm.deployer.landscaper.gardener.cloud/v1alpha1
    kind: ProviderConfiguration

    chart:
      fromResource: 
{{"{{"}} toYaml .componentDescriptorDef | indent 8 {{"}}"}}
        resourceName: {{.DeployItemName}}-chart

    updateStrategy: patch

    name: {{.DeployItemName}}
    namespace: {{.ApplicationNamespaceExpression}}
`

func (o *addHelmLsDeployItemOptions) writeExecution(f *os.File) error {
	templateString := executionTemplateExternalRef

	if o.chartDirectoryPath != "" {
		templateString = executionTemplateLocally
	}

	t, err := template.New("").Parse(templateString)
	if err != nil {
		return err
	}

	data := struct {
		ApplicationNamespaceExpression string
		TargetNameExpression           string
		TargetNamespaceExpression      string
		DeployItemName                 string
	}{
		ApplicationNamespaceExpression: blueprints.GetImportExpression(o.targetNsParam),
		TargetNameExpression:           blueprints.GetTargetNameExpression(o.clusterParam),
		TargetNamespaceExpression:      blueprints.GetTargetNamespaceExpression(o.clusterParam),
		DeployItemName:                 o.deployItemName,
	}

	err = t.Execute(f, data)
	if err != nil {
		return err
	}

	return nil
}

func (o *addHelmLsDeployItemOptions) createResources() (*cdresources.ResourceOptions, error) {
	if o.ociReference != "" {
		return o.createOciResource()
	}

	resource, err := o.createDirectoryResource()

	return resource, err
}

func (o *addHelmLsDeployItemOptions) createOciResource() (*cdresources.ResourceOptions, error) {
	ociRegistryRef := cd.OCIRegistryAccess{
		ObjectType:     cd.ObjectType{Type: cd.OCIRegistryType},
		ImageReference: o.ociReference,
	}

	data, err := json.Marshal(&ociRegistryRef)
	if err != nil {
		return nil, err
	}

	resource := &cdresources.ResourceOptions{

		Resource: cd.Resource{
			IdentityObjectMeta: cd.IdentityObjectMeta{
				Name:    o.deployItemName + "-" + "chart",
				Version: o.resourceVersion,
				Type:    helm,
			},
			Relation: cd.ExternalRelation,
			Access: &cd.UnstructuredAccessType{
				ObjectType: cd.ObjectType{Type: cd.OCIRegistryType},
				Raw:        data,
			},
		},
	}

	return resource, nil
}

func (o *addHelmLsDeployItemOptions) createDirectoryResource() (*cdresources.ResourceOptions, error) {
	compress := true
	resource := &cdresources.ResourceOptions{
		Resource: cd.Resource{
			IdentityObjectMeta: cd.IdentityObjectMeta{
				Name:    o.deployItemName + "-" + "chart",
				Version: o.resourceVersion,
				Type:    helm,
			},
			Relation: cd.ExternalRelation,
		},

		Input: &input.BlobInput{
			Type:             input.DirInputType,
			Path:             o.chartDirectoryPath,
			CompressWithGzip: &compress,
		},
	}

	return resource, nil
}
