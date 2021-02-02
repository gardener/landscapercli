// SPDX-FileCopyrightText: 2020 SAP SE or an SAP affiliate company and Gardener contributors.
//
// SPDX-License-Identifier: Apache-2.0

package components

import (
	"bufio"
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"html/template"
	"io/ioutil"
	"os"
	"strings"

	"k8s.io/apimachinery/pkg/util/uuid"
	"github.com/gardener/landscaper/apis/core/v1alpha1"
	"github.com/gardener/landscaper/apis/deployer/manifest"
	"github.com/go-logr/logr"
	"github.com/spf13/cobra"
	"github.com/spf13/pflag"
	"k8s.io/apimachinery/pkg/runtime"
	"sigs.k8s.io/yaml"

	"github.com/gardener/landscapercli/pkg/blueprints"
	"github.com/gardener/landscapercli/pkg/logger"
	"github.com/gardener/landscapercli/pkg/util"
)

const addManifestDeployItemUse = `deployitem \
    [component directory path] \
    [deployitem name] \
   `

const addManifestDeployItemExample = `
landscaper-cli component add manifest deployitem \
  . \
  nginx \
  --file ./deployment.yaml \
  --file ./service.yaml \
  --import-param replicas:integer
`

const addManifestDeployItemShort = `
Command to add a deploy item skeleton to the blueprint of a component`

//var identityKeyValidationRegexp = regexp.MustCompile("^[a-z0-9]([-_+a-z0-9]*[a-z0-9])?$")

type addManifestDeployItemOptions struct {
	componentPath  string
	deployItemName string

	files        *[]string
	importParams *[]string
	importDefinitions []v1alpha1.ImportDefinition

	updateStrategy string
	policy         string

	clusterParam  string
	targetNsParam string
}

// NewCreateCommand creates a new blueprint command to create a blueprint
func NewAddManifestDeployItemCommand(ctx context.Context) *cobra.Command {
	opts := &addManifestDeployItemOptions{}
	cmd := &cobra.Command{
		Use:     addManifestDeployItemUse,
		Example: addManifestDeployItemExample,
		Short:   addManifestDeployItemShort,
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

func (o *addManifestDeployItemOptions) Complete(args []string) error {
	o.componentPath = args[0]
	o.deployItemName = args[1]

	o.importDefinitions = []v1alpha1.ImportDefinition{}
	if o.importParams != nil {
		for _, p := range *o.importParams {
			importDefinition, err := o.parseImportDefinition(p)
			if err != nil {
				return err
			}

			o.importDefinitions = append(o.importDefinitions, *importDefinition)
		}
	}

	return o.validate()
}

func (o *addManifestDeployItemOptions) AddFlags(fs *pflag.FlagSet) {
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

	err = o.addImports(blueprint)
	if err != nil {
		return err
	}

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

func (o *addManifestDeployItemOptions) addImports(blueprint *v1alpha1.Blueprint) error {
	o.addTargetImport(blueprint, o.clusterParam)
	o.addStringImport(blueprint, o.targetNsParam)

	for _, importDefinition := range o.importDefinitions {
		err := o.addImport(blueprint, &importDefinition)
		if err != nil {
			return err
		}
	}

	return nil
}

func (o *addManifestDeployItemOptions) addTargetImport(blueprint *v1alpha1.Blueprint, name string) {
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

func (o *addManifestDeployItemOptions) addStringImport(blueprint *v1alpha1.Blueprint, name string) {
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

func (o *addManifestDeployItemOptions) addImport(blueprint *v1alpha1.Blueprint, importDefinition *v1alpha1.ImportDefinition) error {
	for i := range blueprint.Imports {
		if blueprint.Imports[i].Name == importDefinition.Name {
			// todo: check that the type has not changed
			return nil
		}
	}

	blueprint.Imports = append(blueprint.Imports, *importDefinition)
	return nil
}

// parseImportDefinition creates a new ImportDefinition from a given parameter definition string.
// The parameter definition string must have the format "name:type", for example "replicas:integer".
// The supported types are: string, boolean, integer
func (o *addManifestDeployItemOptions) parseImportDefinition(paramDef string) (*v1alpha1.ImportDefinition, error) {
	a := strings.Index(paramDef, ":")

	if a == -1 {
		return nil, fmt.Errorf(
			"import parameter definition %s has the wrong format; the expected format is name:type",
			paramDef)
	}

	name := paramDef[:a]
	typ := paramDef[a+1:]

	if !(typ == "string" || typ == "boolean" || typ == "integer") {
		return nil, fmt.Errorf(
			"import parameter definition %s contains an unsupported type; the supported types are string, boolean, integer",
			paramDef)
	}

	schema := fmt.Sprintf("{ \"type\": \"%s\" }", typ)
	required := true

	return &v1alpha1.ImportDefinition{
		FieldValueDefinition: v1alpha1.FieldValueDefinition{
			Name:   name,
			Schema: v1alpha1.JSONSchemaDefinition(schema),
		},
		Required: &required,
	}, nil
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
	f, err := os.Create(util.ExecutionFilePath(o.componentPath, o.deployItemName))
	if err != nil {
		return err
	}

	defer f.Close()

	err = o.writeExecution(f)

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
    {{- getManifests }}
`

func (o *addManifestDeployItemOptions) writeExecution(f *os.File) error {
	funcs := map[string]interface{}{
		"getManifests": func() (string, error) {
			data, err := o.getManifestsYaml()
			if err != nil {
				return "", err
			}

			data, err = indentLines(data, 4)
			if err != nil {
				return "", err
			}

			return string(data), nil
		},
	}

	t, err := template.New("").Funcs(funcs).Parse(manifestExecutionTemplate)
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

func indentLines(data []byte, n int) ([]byte, error) {
	prefix := strings.Repeat(" ", n)

	writer := bytes.Buffer{}

	reader := bytes.NewReader(data)
	scanner := bufio.NewScanner(reader)
	scanner.Split(bufio.ScanLines)

	for scanner.Scan() {
		line := scanner.Text()
		writer.WriteString("\n" + prefix + line)
	}

	if err := scanner.Err(); err != nil {
		return nil, err
	}

	return writer.Bytes(), nil
}

func (o *addManifestDeployItemOptions) getManifestsYaml() ([]byte, error) {
	replacement := o.createReplacement()

	manifests, err := o.readManifests(replacement)
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

	data = o.replaceUUIDsByImportTemplates(data, replacement)

	return data, nil
}

func (o *addManifestDeployItemOptions) readManifests(replacement map[string]string) ([]manifest.Manifest, error) {
	manifests := []manifest.Manifest{}

	if o.files == nil {
		return manifests, nil
	}

	for _, filename := range *o.files {
		m, err := o.readManifest(filename, replacement)
		if err != nil {
			return manifests, err
		}

		manifests = append(manifests, *m)
	}

	return manifests, nil
}

func (o *addManifestDeployItemOptions) readManifest(filename string, replacement map[string]string) (*manifest.Manifest, error) {
	yamlData, err := ioutil.ReadFile(filename)
	if err != nil {
		return nil, err
	}

	var m interface{}
	err = yaml.Unmarshal(yamlData, &m)
	if err != nil {
		return nil, err
	}

	m = o.replaceParamsByUUIDs(m, replacement)

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

func (o *addManifestDeployItemOptions) createReplacement() map[string]string {
	replacement := map[string]string{}

	for _, importDef := range o.importDefinitions {
		replacement[importDef.Name] = string(uuid.NewUUID())
	}

	return replacement
}

func (o *addManifestDeployItemOptions) replaceParamsByUUIDs(in interface{}, replacement map[string]string) interface{} {
	switch m := in.(type) {
	case map[string]interface{}:
		for k, _ := range m {
			m[k] = o.replaceParamsByUUIDs(m[k], replacement)
		}
		return m

	case []interface{}:
		for k, _ := range m {
			m[k] = o.replaceParamsByUUIDs(m[k], replacement)
		}
		return m

	case string:
		newValue, ok := replacement[m]
		if ok {
			return newValue
		}
		return m

	default:
		return m
	}
}

func (o *addManifestDeployItemOptions) replaceUUIDsByImportTemplates(data []byte, replacement map[string]string) []byte {
	s := string(data)

	for paramName, uuid := range replacement {
		newValue := fmt.Sprintf("{{ .imports.%s }}", paramName)
		s = strings.ReplaceAll(s, uuid, newValue)
	}

	return []byte(s)
}