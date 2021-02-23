// SPDX-FileCopyrightText: 2020 SAP SE or an SAP affiliate company and Gardener contributors.
//
// SPDX-License-Identifier: Apache-2.0

package components

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"os"
	"strings"
	"text/template"

	"github.com/Masterminds/semver/v3"
	cdresources "github.com/gardener/component-cli/pkg/commands/componentarchive/resources"
	cd "github.com/gardener/component-spec/bindings-go/apis/v2"

	"github.com/gardener/landscapercli/pkg/components"

	"github.com/gardener/landscaper/apis/core/v1alpha1"
	"github.com/go-logr/logr"
	"github.com/spf13/cobra"
	"github.com/spf13/pflag"
	"sigs.k8s.io/yaml"

	"github.com/gardener/landscapercli/pkg/blueprints"
	"github.com/gardener/landscapercli/pkg/logger"
	"github.com/gardener/landscapercli/pkg/util"
)

const ociImage = "ociImage"

const addContainerDeployItemUse = `deployitem \
    [deployitem name] \
`

const addContainerDeployItemExample = `
landscaper-cli component add container deployitem \
  myDeployItem \
  --resource-version v0.1.0 \
  --component-directory ~/myComponent \
  --image alpine \
  --command sh,-c \
  --args env,ls \
  --import-param replicas:integer \
  --export-param message:string \
  --cluster-param target-cluster \
`

const addContainerDeployItemShort = `
Command to add a deploy item skeleton to the blueprint of a component`

type addContainerDeployItemOptions struct {
	componentPath string

	deployItemName string

	resourceVersion string

	image string

	command *[]string

	args *[]string

	// import parameter definitions in the format "name:type"
	importParams *[]string

	// parsed import parameter definitions
	importDefinitions map[string]*v1alpha1.ImportDefinition

	// export parameter definitions in the format "name:type"
	exportParams *[]string

	// parsed export parameter definitions
	exportDefinitions map[string]*v1alpha1.ExportDefinition

	clusterParam string
}

func NewAddContainerDeployItemCommand(ctx context.Context) *cobra.Command {
	opts := &addContainerDeployItemOptions{}
	cmd := &cobra.Command{
		Use:     addContainerDeployItemUse,
		Example: addContainerDeployItemExample,
		Short:   addContainerDeployItemShort,
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
			fmt.Printf("  \n- import and export definitions added to blueprint")
			fmt.Printf("  \n- reference to image added to resources.yaml")
		},
	}

	opts.AddFlags(cmd.Flags())

	return cmd
}

func (o *addContainerDeployItemOptions) Complete(args []string) error {
	o.deployItemName = args[0]

	if err := o.parseImportDefinitions(); err != nil {
		return err
	}

	if err := o.parseExportDefinitions(); err != nil {
		return err
	}

	return o.validate()
}

func (o *addContainerDeployItemOptions) AddFlags(fs *pflag.FlagSet) {
	fs.StringVar(&o.resourceVersion,
		"resource-version",
		"",
		"resource version")
	fs.StringVar(&o.componentPath,
		"component-directory",
		".",
		"path to component directory (optional, default is current directory)")

	fs.StringVar(&o.image,
		"image",
		"",
		"image")

	o.command = &[]string{}
	fs.StringSliceVarP(o.command,
		"command",
		"c",
		[]string{},
		"command (optional, multi-value)")

	o.args = &[]string{}
	fs.StringSliceVarP(o.args,
		"args",
		"a",
		[]string{},
		"arguments (optional, multi-value)")

	o.importParams = &[]string{}
	fs.StringSliceVarP(o.importParams,
		"import-param",
		"i",
		[]string{},
		"import parameter as name:integer|string|boolean, e.g. replicas:integer (optional, multi-value)")

	o.exportParams = &[]string{}
	fs.StringSliceVarP(o.exportParams,
		"export-param",
		"e",
		[]string{},
		"export parameter as name:integer|string|boolean, e.g. replicas:integer (optional, multi-value)")

	fs.StringVar(&o.clusterParam,
		"cluster-param",
		"",
		"import parameter name for the target resource containing the access data of the target cluster (optional)")
}

func (o *addContainerDeployItemOptions) validate() error {
	if !identityKeyValidationRegexp.Match([]byte(o.deployItemName)) {
		return fmt.Errorf("the deploy item name must consist of lower case alphanumeric characters, '-', '_' " +
			"or '+', and must start and end with an alphanumeric character")
	}

	if o.resourceVersion == "" {
		return fmt.Errorf("resource-version is missing")
	}

	_, err := semver.NewVersion(o.resourceVersion)
	if err != nil {
		return fmt.Errorf("resource-version %s is not semver compatible", o.resourceVersion)
	}

	err = o.checkIfDeployItemNotAlreadyAdded()
	if err != nil {
		return err
	}

	if o.image == "" {
		return fmt.Errorf("image is missing")
	}

	return nil
}

func (o *addContainerDeployItemOptions) run(ctx context.Context, log logr.Logger) error {
	err := o.createExecutionFile()
	if err != nil {
		return err
	}

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

	blueprintBuilder.AddDeployExecution(o.deployItemName)
	if o.clusterParam != "" {
		blueprintBuilder.AddImportForTarget(o.clusterParam)
	}
	blueprintBuilder.AddImportsFromMap(o.importDefinitions)
	blueprintBuilder.AddExportsFromMap(o.exportDefinitions)

	return blueprints.NewBlueprintWriter(blueprintPath).Write(blueprint)
}

func (o *addContainerDeployItemOptions) addResource() error {
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

func (o *addContainerDeployItemOptions) createResources() (*cdresources.ResourceOptions, error) {
	return o.createOciResource()
}

func (o *addContainerDeployItemOptions) createOciResource() (*cdresources.ResourceOptions, error) {
	ociRegistryRef := cd.OCIRegistryAccess{
		ObjectType:     cd.ObjectType{Type: cd.OCIRegistryType},
		ImageReference: o.image,
	}

	data, err := json.Marshal(&ociRegistryRef)
	if err != nil {
		return nil, err
	}

	resource := &cdresources.ResourceOptions{

		Resource: cd.Resource{
			IdentityObjectMeta: cd.IdentityObjectMeta{
				Name:    o.deployItemName,
				Version: o.resourceVersion,
				Type:    ociImage,
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

func (o *addContainerDeployItemOptions) checkIfDeployItemNotAlreadyAdded() error {
	_, err := os.Stat(util.ExecutionFilePath(o.componentPath, o.deployItemName))
	if err != nil {
		if os.IsNotExist(err) {
			return nil
		}
		return err
	}

	return fmt.Errorf("Deploy item was already added. The corresponding deploy execution file %s already exists\n",
		util.ExecutionFilePath(o.componentPath, o.deployItemName))
}

func (o *addContainerDeployItemOptions) parseImportDefinitions() error {
	o.importDefinitions = map[string]*v1alpha1.ImportDefinition{}
	if o.importParams != nil {
		for _, p := range *o.importParams {
			importDefinition, err := o.parseImportDefinition(p)
			if err != nil {
				return err
			}

			_, exists := o.importDefinitions[importDefinition.Name]
			if exists {
				return fmt.Errorf("import parameter %s occurs more than once", importDefinition.Name)
			}

			o.importDefinitions[importDefinition.Name] = importDefinition
		}
	}

	return nil
}

// parseImportDefinition creates a new ImportDefinition from a given parameter definition string.
// The parameter definition string must have the format "name:type", for example "replicas:integer".
// The supported types are: string, boolean, integer
func (o *addContainerDeployItemOptions) parseImportDefinition(paramDef string) (*v1alpha1.ImportDefinition, error) {
	fieldValueDef, err := o.parseFieldValueDefinition(paramDef)
	if err != nil {
		return nil, err
	}

	required := true

	return &v1alpha1.ImportDefinition{
		FieldValueDefinition: *fieldValueDef,
		Required:             &required,
	}, nil
}

func (o *addContainerDeployItemOptions) parseExportDefinitions() error {
	o.exportDefinitions = map[string]*v1alpha1.ExportDefinition{}
	if o.exportParams != nil {
		for _, p := range *o.exportParams {
			exportDefinition, err := o.parseExportDefinition(p)
			if err != nil {
				return err
			}

			_, exists := o.exportDefinitions[exportDefinition.Name]
			if exists {
				return fmt.Errorf("export parameter %s occurs more than once", exportDefinition.Name)
			}

			o.exportDefinitions[exportDefinition.Name] = exportDefinition
		}
	}

	return nil
}

func (o *addContainerDeployItemOptions) parseExportDefinition(paramDef string) (*v1alpha1.ExportDefinition, error) {
	fieldValueDef, err := o.parseFieldValueDefinition(paramDef)
	if err != nil {
		return nil, err
	}

	return &v1alpha1.ExportDefinition{
		FieldValueDefinition: *fieldValueDef,
	}, nil
}

func (o *addContainerDeployItemOptions) parseFieldValueDefinition(paramDef string) (*v1alpha1.FieldValueDefinition, error) {
	a := strings.Index(paramDef, ":")

	if a == -1 {
		return nil, fmt.Errorf(
			"parameter definition %s has the wrong format; the expected format is name:type",
			paramDef)
	}

	name := paramDef[:a]
	typ := paramDef[a+1:]

	if !(typ == "string" || typ == "boolean" || typ == "integer") {
		return nil, fmt.Errorf(
			"parameter definition %s contains an unsupported type; the supported types are string, boolean, integer",
			paramDef)
	}

	return &v1alpha1.FieldValueDefinition{
		Name:   name,
		Schema: v1alpha1.JSONSchemaDefinition{RawMessage: []byte(fmt.Sprintf("{ \"type\": \"%s\" }", typ))},
	}, nil
}

func (o *addContainerDeployItemOptions) createExecutionFile() error {
	f, err := os.Create(util.ExecutionFilePath(o.componentPath, o.deployItemName))
	if err != nil {
		return err
	}

	defer f.Close()

	err = o.writeExecution(f)
	if err != nil {
		return err
	}

	return nil
}

const containerExecutionTemplateWithTarget = `deployItems:
- name: {{.DeployItemName}}
  type: landscaper.gardener.cloud/container
  target:
    name: {{.TargetNameExpression}}
    namespace: {{.TargetNamespaceExpression}}
  config:
    apiVersion: container.deployer.landscaper.gardener.cloud/v1alpha1
    kind: ProviderConfiguration
    image: {{.Image}}
`

const containerExecutionTemplateWithoutTarget = `deployItems:
- name: {{.DeployItemName}}
  type: landscaper.gardener.cloud/container
  config:
    apiVersion: container.deployer.landscaper.gardener.cloud/v1alpha1
    kind: ProviderConfiguration
    image: {{.Image}}
`

func (o *addContainerDeployItemOptions) writeExecution(f io.Writer) error {
	containerExecutionTemplate := containerExecutionTemplateWithTarget
	if o.clusterParam == "" {
		containerExecutionTemplate = containerExecutionTemplateWithoutTarget
	}

	t, err := template.New("").Parse(containerExecutionTemplate)
	if err != nil {
		return err
	}

	commandSection, err := o.getCommandSection()
	if err != nil {
		return err
	}

	importValuesSection, err := o.getImportValuesSection()
	if err != nil {
		return err
	}

	sections := indentLines(string(commandSection)+string(importValuesSection), 4)

	data := struct {
		DeployItemName            string
		TargetNameExpression      string
		TargetNamespaceExpression string
		ImportValuesExpression    string
		Image                     string
	}{
		DeployItemName:            o.deployItemName,
		TargetNameExpression:      blueprints.GetTargetNameExpression(o.clusterParam),
		TargetNamespaceExpression: blueprints.GetTargetNamespaceExpression(o.clusterParam),
		Image:                     o.image,
	}

	err = t.Execute(f, data)
	if err != nil {
		return fmt.Errorf("could not template deploy execution file: %w", err)
	}

	_, err = f.Write([]byte(sections))
	if err != nil {
		return fmt.Errorf("could not write command section of deploy execution file: %w", err)
	}

	return nil
}

func (o *addContainerDeployItemOptions) getCommandSection() ([]byte, error) {
	commandSection := map[string][]string{
		"command": *o.command,
		"args":    *o.args,
	}

	data, err := yaml.Marshal(commandSection)
	if err != nil {
		return nil, fmt.Errorf("could not write command and arguments: %w", err)
	}

	return []byte(string(data)), nil
}

func (o *addContainerDeployItemOptions) getImportValuesSection() ([]byte, error) {
	b := strings.Builder{}

	if len(o.importDefinitions) == 0 {
		if _, err := b.WriteString("importValues: []\n"); err != nil {
			return nil, fmt.Errorf("could not write import values: %w", err)
		}
	} else {
		b.WriteString("importValues:\n")
		for parameterName := range o.importDefinitions {
			b.WriteString("  " + parameterName + ": " + blueprints.GetImportExpression(parameterName) + "\n")
		}
	}

	return []byte(b.String()), nil
}
