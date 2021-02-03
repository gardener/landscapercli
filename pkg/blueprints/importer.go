package blueprints

import (
	"fmt"
	"strings"

	"github.com/gardener/landscaper/apis/core/v1alpha1"
)

type Importer struct{}

func NewImporter() *Importer {
	return &Importer{}
}

func (imp *Importer) AddImports(blueprint *v1alpha1.Blueprint, importDefinitions []v1alpha1.ImportDefinition) {
	for _, importDefinition := range importDefinitions {
		imp.AddImport(blueprint, &importDefinition)
	}
}

func (imp *Importer) AddImport(blueprint *v1alpha1.Blueprint, importDefinition *v1alpha1.ImportDefinition) {
	if imp.existsImport(blueprint, importDefinition.Name) {
		return
	}

	blueprint.Imports = append(blueprint.Imports, *importDefinition)
}

func (imp *Importer) existsImport(blueprint *v1alpha1.Blueprint, name string) bool {
	for i := range blueprint.Imports {
		if blueprint.Imports[i].Name == name {
			return true
		}
	}

	return false
}

func (imp *Importer) AddImportForTarget(blueprint *v1alpha1.Blueprint, paramName string) {
	required := true
	importDefinition := &v1alpha1.ImportDefinition{
		FieldValueDefinition: v1alpha1.FieldValueDefinition{
			Name:       paramName,
			TargetType: string(v1alpha1.KubernetesClusterTargetType),
		},
		Required: &required,
	}
	imp.AddImport(blueprint, importDefinition)
}

func (imp *Importer) AddImportForElementaryType(blueprint *v1alpha1.Blueprint, paramName, paramType string) {
	required := true
	importDefinition := &v1alpha1.ImportDefinition{
		FieldValueDefinition: v1alpha1.FieldValueDefinition{
			Name:   paramName,
			Schema: imp.buildElementarySchema(paramType),
		},
		Required: &required,
	}
	imp.AddImport(blueprint, importDefinition)
}

// parseImportDefinition creates a new ImportDefinition from a given parameter definition string.
// The parameter definition string must have the format "name:type", for example "replicas:integer".
// The supported types are: string, boolean, integer
func (imp *Importer) ParseImportDefinition(paramDef string) (*v1alpha1.ImportDefinition, error) {
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

	required := true

	return &v1alpha1.ImportDefinition{
		FieldValueDefinition: v1alpha1.FieldValueDefinition{
			Name:   name,
			Schema: imp.buildElementarySchema(typ),
		},
		Required: &required,
	}, nil
}

func (imp *Importer) buildElementarySchema(elementaryType string) v1alpha1.JSONSchemaDefinition {
	schema := fmt.Sprintf("{ \"type\": \"%s\" }", elementaryType)
	return v1alpha1.JSONSchemaDefinition(schema)
}
