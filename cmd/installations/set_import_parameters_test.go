package installations

import (
	"encoding/json"
	"testing"

	cdv2 "github.com/gardener/component-spec/bindings-go/apis/v2"
	lsv1alpha1 "github.com/gardener/landscaper/apis/core/v1alpha1"
	"github.com/stretchr/testify/assert"
	"sigs.k8s.io/yaml"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func TestImportParameterFilling(t *testing.T) {
	installation := lsv1alpha1.Installation{
		TypeMeta: metav1.TypeMeta{
			APIVersion: lsv1alpha1.SchemeGroupVersion.String(),
			Kind:       "Installation",
		},
		ObjectMeta: metav1.ObjectMeta{
			Name: "sampleinstallation",
		},
		Spec: lsv1alpha1.InstallationSpec{
			ComponentDescriptor: &lsv1alpha1.ComponentDescriptorDefinition{
				Reference: &lsv1alpha1.ComponentDescriptorReference{
					RepositoryContext: &cdv2.RepositoryContext{
						Type:    cdv2.OCIRegistryType,
						BaseURL: "sample.cluster.local",
					},
					ComponentName: "component name",
					Version:       "v0.0.1",
				},
			},
			Blueprint: lsv1alpha1.BlueprintDefinition{
				Reference: &lsv1alpha1.RemoteBlueprintReference{
					ResourceName: "blueprintressourcename",
				},
			},
			Imports: lsv1alpha1.InstallationImports{
				Data: []lsv1alpha1.DataImport{
					{
						Name: "intValue",
					},
					{
						Name: "floatValue",
					},
					{
						Name: "stringValue",
					},
					{
						Name: "stringValueWithSpace",
					},
				},
			},
		},
	}

	inputParameters := map[string]string{
		"intValue":             "10",
		"floatValue":           "10.1",
		"stringValue":          "testValue",
		"stringValueWithSpace": "test Value",
	}

	replaceImportsWithInputParameters(&installation, &inputParametersOptions{importParameters: inputParameters})

	t.Run("String replacements", func(t *testing.T) {
		assert.Equal(t, json.RawMessage(`"testValue"`), installation.Spec.ImportDataMappings["stringValue"])
		assert.Equal(t, json.RawMessage(`"test Value"`), installation.Spec.ImportDataMappings["stringValueWithSpace"])
	})
	t.Run("Number replacements", func(t *testing.T) {
		assert.Equal(t, json.RawMessage(inputParameters["intValue"]), installation.Spec.ImportDataMappings["intValue"])
		assert.Equal(t, json.RawMessage(inputParameters["floatValue"]), installation.Spec.ImportDataMappings["floatValue"])
	})
	t.Run("Correct removing of imports.data and adding to importDataMapping list", func(t *testing.T) {
		assert.Equal(t, 4, len(installation.Spec.ImportDataMappings))
		assert.Equal(t, 0, len(installation.Spec.Imports.Data))
	})
	t.Run("Marshaling of yaml", func(t *testing.T) {
		_, err := yaml.Marshal(installation)
		assert.Nil(t, err)
	})

}
