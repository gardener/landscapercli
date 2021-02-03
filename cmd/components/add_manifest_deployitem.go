// SPDX-FileCopyrightText: 2020 SAP SE or an SAP affiliate company and Gardener contributors.
//
// SPDX-License-Identifier: Apache-2.0

package components

import (
	"context"
	"encoding/json"
	"fmt"
	"html/template"
	"io/ioutil"
	"os"
	"strings"

	"github.com/gardener/landscaper/apis/core/v1alpha1"
	"github.com/gardener/landscaper/apis/deployer/manifest"
	"github.com/go-logr/logr"
	"github.com/spf13/cobra"
	"github.com/spf13/pflag"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/util/uuid"
	"sigs.k8s.io/yaml"

	"github.com/gardener/landscapercli/pkg/blueprints"
	"github.com/gardener/landscapercli/pkg/logger"
	"github.com/gardener/landscapercli/pkg/util"
)

const addManifestDeployItemUse = `deployitem \
    [deployitem name] \
   `

const addManifestDeployItemExample = `
landscaper-cli component add manifest deployitem \
  nginx \
  --component-path ~/myComponent \
  --file ./deployment.yaml \
  --file ./service.yaml \
  --import-param replicas:integer
  --cluster-param target-cluster
  --target-ns-param target-namespace
`

const addManifestDeployItemShort = `
Command to add a deploy item skeleton to the blueprint of a component`

//var identityKeyValidationRegexp = regexp.MustCompile("^[a-z0-9]([-_+a-z0-9]*[a-z0-9])?$")

type addManifestDeployItemOptions struct {
	componentPath string

	deployItemName string

	// names of manifest files
	files *[]string

	// import parameter definitions in the format "name:type"
	importParams *[]string

	// parsed import parameter definitions
	importDefinitions []v1alpha1.ImportDefinition

	// a map that assigns with each import parameter name a uuid
	replacement map[string]string

	updateStrategy string

	policy string

	clusterParam string

	targetNsParam string
}

// NewCreateCommand creates a new blueprint command to create a blueprint
func NewAddManifestDeployItemCommand(ctx context.Context) *cobra.Command {
	opts := &addManifestDeployItemOptions{}
	cmd := &cobra.Command{
		Use:     addManifestDeployItemUse,
		Example: addManifestDeployItemExample,
		Short:   addManifestDeployItemShort,
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

			fmt.Printf("Successfully added deploy item")
		},
	}

	opts.AddFlags(cmd.Flags())

	return cmd
}

func (o *addManifestDeployItemOptions) Complete(args []string) error {
	o.deployItemName = args[0]

	o.importDefinitions = []v1alpha1.ImportDefinition{}
	o.replacement = map[string]string{}
	if o.importParams != nil {
		importer := blueprints.NewImporter()

		for _, p := range *o.importParams {
			importDefinition, err := importer.ParseImportDefinition(p)
			if err != nil {
				return err
			}

			o.importDefinitions = append(o.importDefinitions, *importDefinition)

			if _, ok := o.replacement[importDefinition.Name]; ok {
				return fmt.Errorf("import parameter %s occurs more than once", importDefinition.Name)
			}

			o.replacement[importDefinition.Name] = string(uuid.NewUUID())
		}
	}

	return o.validate()
}

func (o *addManifestDeployItemOptions) AddFlags(fs *pflag.FlagSet) {
	fs.StringVar(&o.componentPath,
		"component-path",
		".",
		"path to component directory")
	o.files = fs.StringArray(
		"file",
		[]string{},
		"manifest file")
	o.importParams = fs.StringArray(
		"import-param",
		[]string{},
		"import parameter")
	fs.StringVar(&o.updateStrategy,
		"update-strategy",
		"update",
		"update stategy")
	fs.StringVar(&o.policy,
		"policy",
		"manage",
		"policy")
	fs.StringVar(&o.clusterParam,
		"cluster-param",
		"targetCluster",
		"import parameter name for the target resource containing the access data of the target cluster")
	fs.StringVar(&o.targetNsParam,
		"target-ns-param",
		"",
		"target namespace")
}

func (o *addManifestDeployItemOptions) validate() error {
	if !identityKeyValidationRegexp.Match([]byte(o.deployItemName)) {
		return fmt.Errorf("the deploy item name must consist of lower case alphanumeric characters, '-', '_' " +
			"or '+', and must start and end with an alphanumeric character")
	}

	if o.clusterParam == "" {
		return fmt.Errorf("cluster-param is missing")
	}

	if o.targetNsParam == "" {
		return fmt.Errorf("target-ns-param is missing")
	}

	return nil
}

func (o *addManifestDeployItemOptions) run(ctx context.Context, log logr.Logger) error {
	blueprintPath := util.BlueprintDirectoryPath(o.componentPath)
	blueprint, err := blueprints.NewBlueprintReader(blueprintPath).Read()
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

func (o *addManifestDeployItemOptions) existsExecution(blueprint *v1alpha1.Blueprint) bool {
	for i := range blueprint.DeployExecutions {
		execution := &blueprint.DeployExecutions[i]
		if execution.Name == o.deployItemName {
			return true
		}
	}

	return false
}

func (o *addManifestDeployItemOptions) addExecution(blueprint *v1alpha1.Blueprint) {
	blueprint.DeployExecutions = append(blueprint.DeployExecutions, v1alpha1.TemplateExecutor{
		Name: o.deployItemName,
		Type: v1alpha1.GOTemplateType,
		File: "/" + util.ExecutionFileName(o.deployItemName),
	})
}

// addImports adds the following imports to the blueprint:
// - one import parameter specified by the flag "--cluster-param [parameter name]"
// - one import parameter specified by the flag "--target-ns-param [parameter name]"
// - all import parameters specified  by (multiple) flags "--import-param [parameter name]:[parameter type]"
func (o *addManifestDeployItemOptions) addImports(blueprint *v1alpha1.Blueprint) {
	imp := blueprints.NewImporter()
	imp.AddImportForTarget(blueprint, o.clusterParam)
	imp.AddImportForElementaryType(blueprint, o.targetNsParam, "string")
	imp.AddImports(blueprint, o.importDefinitions)
}

func (o *addManifestDeployItemOptions) existsExecutionFile() (bool, error) {
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

func (o *addManifestDeployItemOptions) createExecutionFile() error {
	err := o.writeExecution()
	if err != nil {
		return err
	}

	f, err := os.OpenFile(util.ExecutionFilePath(o.componentPath, o.deployItemName), os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		return err
	}

	defer f.Close()

	manifests, err := o.getManifests()
	if err != nil {
		return err
	}

	_, err = f.WriteString(manifests)

	return err
}

const manifestExecutionTemplate = `deployItems:
- name: {{.DeployItemName}}
  type: landscaper.gardener.cloud/kubernetes-manifest
  target:
    name: {{"{{"}} .imports.{{.ClusterParam}}.metadata.name {{"}}"}}
    namespace: {{"{{"}} .imports.{{.ClusterParam}}.metadata.namespace {{"}}"}}
  config:
    apiVersion: manifest.deployer.landscaper.gardener.cloud/v1alpha2
    kind: ProviderConfiguration
    updateStrategy: {{.UpdateStrategy}}
`

func (o *addManifestDeployItemOptions) writeExecution() error {
	f, err := os.Create(util.ExecutionFilePath(o.componentPath, o.deployItemName))
	if err != nil {
		return err
	}

	defer f.Close()

	t, err := template.New("").Parse(manifestExecutionTemplate)
	if err != nil {
		return err
	}

	data := struct {
		ClusterParam   string
		TargetNsParam  string
		DeployItemName string
		UpdateStrategy string
	}{
		ClusterParam:   o.clusterParam,
		TargetNsParam:  o.targetNsParam,
		DeployItemName: o.deployItemName,
		UpdateStrategy: o.updateStrategy,
	}

	err = t.Execute(f, data)
	if err != nil {
		return err
	}

	return nil
}

func (o *addManifestDeployItemOptions) getManifests() (string, error) {
	data, err := o.getManifestsYaml()
	if err != nil {
		return "", err
	}

	stringData := string(data)
	stringData = indentLines(stringData, 4)
	return stringData, nil
}

func indentLines(data string, n int) string {
	indent := strings.Repeat(" ", n)
	return indent + strings.ReplaceAll(data, "\n", "\n"+indent)
}

func (o *addManifestDeployItemOptions) getManifestsYaml() ([]byte, error) {
	manifests, err := o.readManifests()
	if err != nil {
		return nil, err
	}

	m := map[string][]manifest.Manifest{
		"manifests": manifests,
	}

	data, err := yaml.Marshal(m)
	if err != nil {
		return nil, err
	}

	data = o.replaceUUIDsByImportTemplates(data)

	return data, nil
}

func (o *addManifestDeployItemOptions) readManifests() ([]manifest.Manifest, error) {
	manifests := []manifest.Manifest{}

	if o.files == nil {
		return manifests, nil
	}

	for _, filename := range *o.files {
		m, err := o.readManifest(filename)
		if err != nil {
			return manifests, err
		}

		manifests = append(manifests, *m)
	}

	return manifests, nil
}

func (o *addManifestDeployItemOptions) readManifest(filename string) (*manifest.Manifest, error) {
	yamlData, err := ioutil.ReadFile(filename)
	if err != nil {
		return nil, err
	}

	var m interface{}
	err = yaml.Unmarshal(yamlData, &m)
	if err != nil {
		return nil, err
	}

	m = o.replaceParamsByUUIDs(m)

	// render to string
	uuidData, err := json.Marshal(m)
	if err != nil {
		return nil, err
	}

	m2 := &manifest.Manifest{
		Policy: manifest.ManifestPolicy(o.policy),
		Manifest: &runtime.RawExtension{
			Raw: uuidData,
		},
	}

	return m2, nil
}

func (o *addManifestDeployItemOptions) replaceParamsByUUIDs(in interface{}) interface{} {
	switch m := in.(type) {
	case map[string]interface{}:
		for k := range m {
			m[k] = o.replaceParamsByUUIDs(m[k])
		}
		return m

	case []interface{}:
		for k := range m {
			m[k] = o.replaceParamsByUUIDs(m[k])
		}
		return m

	case string:
		newValue, ok := o.replacement[m]
		if ok {
			return newValue
		}
		return m

	default:
		return m
	}
}

func (o *addManifestDeployItemOptions) replaceUUIDsByImportTemplates(data []byte) []byte {
	s := string(data)

	for paramName, uuid := range o.replacement {
		newValue := fmt.Sprintf("{{ .imports.%s }}", paramName)
		s = strings.ReplaceAll(s, uuid, newValue)
	}

	return []byte(s)
}
