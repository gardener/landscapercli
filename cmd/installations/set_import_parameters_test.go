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
	repoCtx, _ := cdv2.NewUnstructured(cdv2.NewOCIRegistryRepository("sample.cluster.local", ""))
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
					RepositoryContext: &repoCtx,
					ComponentName:     "component name",
					Version:           "v0.0.1",
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

	importParameters := map[string]string{
		"intValue":             "10",
		"floatValue":           "10.1",
		"stringValue":          "testValue",
		"stringValueWithSpace": "test Value",
	}

	err := replaceImportsWithImportParameters(&installation, &importParametersOptions{importParameters: importParameters})
	assert.Nil(t, err)

	t.Run("String replacements", func(t *testing.T) {
		assert.Equal(t, json.RawMessage(`"testValue"`), installation.Spec.ImportDataMappings["stringValue"].RawMessage)
		assert.Equal(t, json.RawMessage(`"test Value"`), installation.Spec.ImportDataMappings["stringValueWithSpace"].RawMessage)
	})
	t.Run("Number replacements", func(t *testing.T) {
		assert.Equal(t, json.RawMessage(importParameters["intValue"]), installation.Spec.ImportDataMappings["intValue"].RawMessage)
		assert.Equal(t, json.RawMessage(importParameters["floatValue"]), installation.Spec.ImportDataMappings["floatValue"].RawMessage)
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
