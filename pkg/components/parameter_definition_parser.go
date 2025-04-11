package components

import (
	"fmt"
	"strings"

	"github.com/gardener/landscaper/apis/core/v1alpha1"
)

type ParameterDefinitionParser struct{}

func (p *ParameterDefinitionParser) ParseImportDefinitions(importParams *[]string) (importDefinitions map[string]*v1alpha1.ImportDefinition, err error) {
	importDefinitions = map[string]*v1alpha1.ImportDefinition{}

	if importParams == nil {
		return importDefinitions, nil
	}

	for _, importParam := range *importParams {
		importDefinition, err := p.ParseImportDefinition(importParam)
		if err != nil {
			return nil, err
		}

		_, exists := importDefinitions[importDefinition.Name]
		if exists {
			return nil, fmt.Errorf("import parameter %s occurs more than once", importDefinition.Name)
		}

		importDefinitions[importDefinition.Name] = importDefinition
	}

	return importDefinitions, nil
}

// ParseImportDefinition creates a new ImportDefinition from a given parameter definition string.
// The parameter definition string must have the format "name:type", for example "replicas:integer".
// The supported types are: string, boolean, integer
func (p *ParameterDefinitionParser) ParseImportDefinition(paramDef string) (*v1alpha1.ImportDefinition, error) {
	fieldValueDef, err := p.ParseFieldValueDefinition(paramDef)
	if err != nil {
		return nil, err
	}

	required := true
	return &v1alpha1.ImportDefinition{
		FieldValueDefinition: *fieldValueDef,
		Required:             &required,
	}, nil
}

func (p *ParameterDefinitionParser) ParseExportDefinitions(exportParams *[]string) (exportDefinitions map[string]*v1alpha1.ExportDefinition, err error) {
	exportDefinitions = map[string]*v1alpha1.ExportDefinition{}

	if exportParams == nil {
		return exportDefinitions, nil
	}

	for _, exportParam := range *exportParams {
		exportDefinition, err := p.ParseExportDefinition(exportParam)
		if err != nil {
			return nil, err
		}

		_, exists := exportDefinitions[exportDefinition.Name]
		if exists {
			return nil, fmt.Errorf("export parameter %s occurs more than once", exportDefinition.Name)
		}

		exportDefinitions[exportDefinition.Name] = exportDefinition
	}

	return exportDefinitions, nil
}

func (p *ParameterDefinitionParser) ParseExportDefinition(paramDef string) (*v1alpha1.ExportDefinition, error) {
	fieldValueDef, err := p.ParseFieldValueDefinition(paramDef)
	if err != nil {
		return nil, err
	}

	return &v1alpha1.ExportDefinition{
		FieldValueDefinition: *fieldValueDef,
	}, nil
}

func isValidType(typ string) bool {
	validTypes := map[string]bool{
		"string":  true,
		"boolean": true,
		"integer": true,
	}
	return validTypes[typ]
}

func (p *ParameterDefinitionParser) ParseFieldValueDefinition(paramDef string) (*v1alpha1.FieldValueDefinition, error) {
	a := strings.Index(paramDef, ":")

	if a == -1 {
		return nil, fmt.Errorf(
			"parameter definition %s has the wrong format; the expected format is name:type",
			paramDef)
	}

	name := paramDef[:a]
	typ := paramDef[a+1:]

	if !isValidType(typ) {
		return nil, fmt.Errorf(
			"parameter definition %s contains an unsupported type; the supported types are string, boolean, integer",
			paramDef)
	}

	return &v1alpha1.FieldValueDefinition{
		Name: name,
		Schema: &v1alpha1.JSONSchemaDefinition{
			RawMessage: []byte(fmt.Sprintf(`{ "type": "%s" }`, typ)),
		},
	}, nil
}
