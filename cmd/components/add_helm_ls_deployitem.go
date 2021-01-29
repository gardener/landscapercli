// SPDX-FileCopyrightText: 2020 SAP SE or an SAP affiliate company and Gardener contributors.
//
// SPDX-License-Identifier: Apache-2.0

package components

import (
	"context"
	"encoding/json"
	"fmt"
	"html/template"
	"os"
	"regexp"

	"github.com/gardener/component-cli/pkg/commands/componentarchive/input"
	cdresources "github.com/gardener/component-cli/pkg/commands/componentarchive/resources"
	cd "github.com/gardener/component-spec/bindings-go/apis/v2"
	"github.com/gardener/landscaper/pkg/apis/core/v1alpha1"

	"github.com/go-logr/logr"
	"github.com/spf13/cobra"
	"github.com/spf13/pflag"

	"github.com/gardener/landscapercli/pkg/blueprints"
	"github.com/gardener/landscapercli/pkg/components"
	"github.com/gardener/landscapercli/pkg/logger"
	"github.com/gardener/landscapercli/pkg/util"
)

const addHelmLSDeployItemUse = `deployitem \
    [component directory path] \
    [deployitem name] \
   `

const addHelmLSDeployItemExample = `
landscaper-cli component add helm-ls deployitem \
  . \
  nginx \
  --oci-reference eu.gcr.io/gardener-project/landscaper/tutorials/charts/ingress-nginx:v0.1.0 \
  --chart-version v0.1.0`

const addHelmLSDeployItemShort = `
Command to add a deploy item skeleton to the blueprint of a component`

const helm = "helm"

var identityKeyValidationRegexp = regexp.MustCompile("^[a-z0-9]([-_+a-z0-9]*[a-z0-9])?$")

type addHelmLsDeployItemOptions struct {
	componentPath  string
	deployItemName string

	ociReference       string
	chartDirectoryPath string

	chartVersion string

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
		Args:    cobra.ExactArgs(2),

		Run: func(cmd *cobra.Command, args []string) {
			if err := opts.Complete(args); err != nil {
				fmt.Println(err.Error())
				os.Exit(1)
			}

			if err := opts.run(ctx, logger.Log); err != nil {
				fmt.Println(err.Error())
				os.Exit(1)
			}

			fmt.Printf("Successfully added deploy item")
		},
	}

	opts.AddFlags(cmd.Flags())

	return cmd
}

func (o *addHelmLsDeployItemOptions) Complete(args []string) error {
	o.componentPath = args[0]
	o.deployItemName = args[1]

	return o.validate()
}

func (o *addHelmLsDeployItemOptions) AddFlags(fs *pflag.FlagSet) {
	fs.StringVar(&o.ociReference,
		"oci-reference",
		"",
		"reference to oci artifact containing the helm chart")
	fs.StringVar(&o.chartDirectoryPath,
		"chart-directory",
		"",
		"path to chart directory")
	fs.StringVar(&o.chartVersion,
		"chart-version",
		"",
		"helm chart version")
	fs.StringVar(&o.clusterParam,
		"cluster-param",
		"targetCluster",
		"target cluster")
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

	if o.chartVersion == "" {
		return fmt.Errorf("chart-version is missing")
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

	if o.existsExecution(blueprint) {
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

	o.addExecution(blueprint)
	o.addImports(blueprint)

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

func (o *addHelmLsDeployItemOptions) existsExecution(blueprint *v1alpha1.Blueprint) bool {
	for i := range blueprint.DeployExecutions {
		execution := &blueprint.DeployExecutions[i]
		if execution.Name == o.deployItemName {
			return true
		}
	}

	return false
}

func (o *addHelmLsDeployItemOptions) addExecution(blueprint *v1alpha1.Blueprint) {
	blueprint.DeployExecutions = append(blueprint.DeployExecutions, v1alpha1.TemplateExecutor{
		Name: o.deployItemName,
		Type: v1alpha1.GOTemplateType,
		File: "/" + util.ExecutionFileName(o.deployItemName),
	})
}

func (o *addHelmLsDeployItemOptions) addImports(blueprint *v1alpha1.Blueprint) {
	o.addTargetImport(blueprint, o.clusterParam)
	o.addStringImport(blueprint, o.targetNsParam)
}

func (o *addHelmLsDeployItemOptions) addTargetImport(blueprint *v1alpha1.Blueprint, name string) {
	for i := range blueprint.Imports {
		if blueprint.Imports[i].Name == name {
			return
		}
	}

	required := true

	blueprint.Imports = append(blueprint.Imports, v1alpha1.ImportDefinition{
		FieldValueDefinition: v1alpha1.FieldValueDefinition{
			Name:       name,
			TargetType: string(v1alpha1.KubernetesClusterTargetType),
		},
		Required: &required,
	})
}

func (o *addHelmLsDeployItemOptions) addStringImport(blueprint *v1alpha1.Blueprint, name string) {
	for i := range blueprint.Imports {
		if blueprint.Imports[i].Name == name {
			return
		}
	}

	required := true

	blueprint.Imports = append(blueprint.Imports, v1alpha1.ImportDefinition{
		FieldValueDefinition: v1alpha1.FieldValueDefinition{
			Name:   name,
			Schema: v1alpha1.JSONSchemaDefinition("{ \"type\": \"string\" }"),
		},
		Required: &required,
	})
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

const executionTemplate = `deployItems:
- name: {{.DeployItemName}}
  type: landscaper.gardener.cloud/helm
  target:
    name: {{"{{"}} .imports.{{.ClusterParam}}.metadata.name {{"}}"}}
    namespace: {{"{{"}} .imports.{{.ClusterParam}}.metadata.namespace {{"}}"}}
  config:
    apiVersion: helm.deployer.landscaper.gardener.cloud/v1alpha1
    kind: ProviderConfiguration

    chart:
      ref: {{"{{"}} with (getResource .cd "name" "{{.DeployItemName}}-chart") {{"}}"}} {{"{{"}} .access.imageReference {{"}}"}} {{"{{"}} end {{"}}"}}

    updateStrategy: patch

    name: {{.DeployItemName}}
    namespace: {{"{{"}} .imports.{{.TargetNsParam}} {{"}}"}}
`

func (o *addHelmLsDeployItemOptions) writeExecution(f *os.File) error {
	t, err := template.New("").Parse(executionTemplate)
	if err != nil {
		return err
	}

	data := struct {
		ClusterParam   string
		TargetNsParam  string
		DeployItemName string
	}{
		ClusterParam:   o.clusterParam,
		TargetNsParam:  o.targetNsParam,
		DeployItemName: o.deployItemName,
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
				Version: o.chartVersion,
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
				Version: o.chartVersion,
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
