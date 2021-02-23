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

func (b *BlueprintBuilder) AddImports(importDefinitions []v1alpha1.ImportDefinition) {
	for _, importDefinition := range importDefinitions {
		b.AddImport(&importDefinition)
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
			Schema: v1alpha1.JSONSchemaDefinition{schema},
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
	})
}
