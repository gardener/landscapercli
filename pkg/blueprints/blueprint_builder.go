package blueprints

import (
	"fmt"

	"github.com/gardener/landscaper/apis/core/v1alpha1"

	"github.com/gardener/landscapercli/pkg/util"
)

type BlueprintBuilder struct {
	blueprint *v1alpha1.Blueprint
}

func NewBlueprintBuilder(blueprint *v1alpha1.Blueprint) *BlueprintBuilder {
	return &BlueprintBuilder{
		blueprint: blueprint,
	}
}

func (b *BlueprintBuilder) AddImportsFromMap(importDefinitions map[string]*v1alpha1.ImportDefinition) {
	for _, importDefinition := range importDefinitions {
		b.AddImport(importDefinition)
	}
}

func (b *BlueprintBuilder) AddExportsFromMap(exportDefinitions map[string]*v1alpha1.ExportDefinition) {
	for _, exportDefinition := range exportDefinitions {
		b.AddExport(exportDefinition)
	}
}

func (b *BlueprintBuilder) AddImport(importDefinition *v1alpha1.ImportDefinition) {
	if b.existsImport(importDefinition.Name) {
		return
	}

	b.blueprint.Imports = append(b.blueprint.Imports, *importDefinition)
}

func (b *BlueprintBuilder) AddExport(exportDefinition *v1alpha1.ExportDefinition) {
	if b.existsExport(exportDefinition.Name) {
		return
	}

	b.blueprint.Exports = append(b.blueprint.Exports, *exportDefinition)
}

func (b *BlueprintBuilder) existsImport(name string) bool {
	for i := range b.blueprint.Imports {
		if b.blueprint.Imports[i].Name == name {
			return true
		}
	}

	return false
}

func (b *BlueprintBuilder) existsExport(name string) bool {
	for i := range b.blueprint.Exports {
		if b.blueprint.Exports[i].Name == name {
			return true
		}
	}

	return false
}

func (b *BlueprintBuilder) AddImportForTarget(paramName string) {
	required := true
	importDefinition := &v1alpha1.ImportDefinition{
		FieldValueDefinition: v1alpha1.FieldValueDefinition{
			Name:       paramName,
			TargetType: string(v1alpha1.KubernetesClusterTargetType),
		},
		Required: &required,
	}
	b.AddImport(importDefinition)
}

func (b *BlueprintBuilder) AddImportForElementaryType(paramName, paramType string) {
	required := true

	schema := []byte(fmt.Sprintf("{ \"type\": \"%s\" }", paramType))

	importDefinition := &v1alpha1.ImportDefinition{
		FieldValueDefinition: v1alpha1.FieldValueDefinition{
			Name:   paramName,
			Schema: &v1alpha1.JSONSchemaDefinition{RawMessage: schema},
		},
		Required: &required,
	}
	b.AddImport(importDefinition)
}

func (b *BlueprintBuilder) ExistsDeployExecution(executionName string) bool {
	for i := range b.blueprint.DeployExecutions {
		execution := &b.blueprint.DeployExecutions[i]
		if execution.Name == executionName {
			return true
		}
	}

	return false
}

func (b *BlueprintBuilder) AddDeployExecution(deployItemName string) {
	b.blueprint.DeployExecutions = append(b.blueprint.DeployExecutions, v1alpha1.TemplateExecutor{
		Name: deployItemName,
		Type: v1alpha1.GOTemplateType,
		File: "/" + util.ExecutionFileName(deployItemName),
		Template: v1alpha1.NewAnyJSON([]byte{}),
	})
}

// AddExportExecution adds one export executions for all export parameters of one deployitem:
// exportExecutions:
// - name: [name of the export execution, here equal to the deployitem name]
//   type: GoTemplate
//   template: |
//     exports:
//       [parameter name]: {{ index .values "deployitems" "[deployitem name]" "[internal parameter name]" }}
func (b *BlueprintBuilder) AddExportExecution(deployItemName string, exportDefinitions map[string]*v1alpha1.ExportDefinition) {
	if len(exportDefinitions) == 0 {
		return
	}

	s := `"exports:\n`
	for exportParameterName := range exportDefinitions {
		s += fmt.Sprintf(`  %s: {{ index .values \"deployitems\" \"%s\" \"%s\" }}\n`,
			exportParameterName, // name of the export parameter as defined in the export section of the blueprint
			deployItemName,      // deployitem that computes the value
			exportParameterName) // (internal) parameter name of the deploy item
		// - for the helm deployer: a key in the exportsFromManifest section
		// - for the container deployer: a key written by the program in the container to $EXPORTS_PATH
	}
	s += `"`

	b.blueprint.ExportExecutions = append(b.blueprint.ExportExecutions, v1alpha1.TemplateExecutor{
		Name: deployItemName, // we give the export execution the same name as the deployitem
		Type: v1alpha1.GOTemplateType,
		Template: v1alpha1.AnyJSON{
			RawMessage: []byte(s),
		},
	})
}
