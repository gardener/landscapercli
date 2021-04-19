package installations

import (
	"testing"

	cdv2 "github.com/gardener/component-spec/bindings-go/apis/v2"
	lsv1alpha1 "github.com/gardener/landscaper/apis/core/v1alpha1"
	"github.com/stretchr/testify/assert"
	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func TestBuildInstallation(t *testing.T) {
	tests := []struct {
		name                 string
		cd                   *cdv2.ComponentDescriptor
		expectedInstallation *lsv1alpha1.Installation
	}{
		{
			name: "test with single repo context",
			cd: &cdv2.ComponentDescriptor{
				ComponentSpec: cdv2.ComponentSpec{
					RepositoryContexts: []cdv2.RepositoryContext{
						{
							Type:    "ociRegistry",
							BaseURL: "first-registry.com",
						},
					},
				},
			},
			expectedInstallation: &lsv1alpha1.Installation{
				TypeMeta: v1.TypeMeta{
					Kind:       "Installation",
					APIVersion: lsv1alpha1.SchemeGroupVersion.String(),
				},
				ObjectMeta: v1.ObjectMeta{
					Name: "test-installation",
				},
				Spec: lsv1alpha1.InstallationSpec{
					ComponentDescriptor: &lsv1alpha1.ComponentDescriptorDefinition{
						Reference: &lsv1alpha1.ComponentDescriptorReference{
							RepositoryContext: &cdv2.RepositoryContext{
								Type:    "ociRegistry",
								BaseURL: "first-registry.com",
							},
						},
					},
					Blueprint: lsv1alpha1.BlueprintDefinition{
						Reference: &lsv1alpha1.RemoteBlueprintReference{
							ResourceName: "blueprint-res",
						},
					},
					Imports: lsv1alpha1.InstallationImports{
						Targets: []lsv1alpha1.TargetImportExport{
							{
								Name:   "test-target",
								Target: "",
							},
						},
						Data: []lsv1alpha1.DataImport{
							{
								Name: "test-data-import",
							},
						},
					},
					Exports: lsv1alpha1.InstallationExports{
						Targets: []lsv1alpha1.TargetImportExport{},
						Data:    []lsv1alpha1.DataExport{},
					},
				},
			},
		},
		{
			name: "use latest repo context if multiple repo contexts exist",
			cd: &cdv2.ComponentDescriptor{
				ComponentSpec: cdv2.ComponentSpec{
					RepositoryContexts: []cdv2.RepositoryContext{
						{
							Type:    "ociRegistry",
							BaseURL: "first-registry.com",
						},
						{
							Type:    "ociRegistry",
							BaseURL: "second-registry.com",
						},
					},
				},
			},
			expectedInstallation: &lsv1alpha1.Installation{
				TypeMeta: v1.TypeMeta{
					Kind:       "Installation",
					APIVersion: lsv1alpha1.SchemeGroupVersion.String(),
				},
				ObjectMeta: v1.ObjectMeta{
					Name: "test-installation",
				},
				Spec: lsv1alpha1.InstallationSpec{
					ComponentDescriptor: &lsv1alpha1.ComponentDescriptorDefinition{
						Reference: &lsv1alpha1.ComponentDescriptorReference{
							RepositoryContext: &cdv2.RepositoryContext{
								Type:    "ociRegistry",
								BaseURL: "second-registry.com",
							},
						},
					},
					Blueprint: lsv1alpha1.BlueprintDefinition{
						Reference: &lsv1alpha1.RemoteBlueprintReference{
							ResourceName: "blueprint-res",
						},
					},
					Imports: lsv1alpha1.InstallationImports{
						Targets: []lsv1alpha1.TargetImportExport{
							{
								Name: "test-target",
							},
						},
						Data: []lsv1alpha1.DataImport{
							{
								Name: "test-data-import",
							},
						},
					},
					Exports: lsv1alpha1.InstallationExports{
						Targets: []lsv1alpha1.TargetImportExport{},
						Data:    []lsv1alpha1.DataExport{},
					},
				},
			},
		},
	}

	blueprint := &lsv1alpha1.Blueprint{
		Imports: lsv1alpha1.ImportDefinitionList{
			{
				FieldValueDefinition: lsv1alpha1.FieldValueDefinition{
					Name:       "test-target",
					TargetType: string(lsv1alpha1.KubernetesClusterTargetType),
				},
			},
			{
				FieldValueDefinition: lsv1alpha1.FieldValueDefinition{
					Name: "test-data-import",
				},
			},
		},
	}

	blueprintResource := cdv2.Resource{
		IdentityObjectMeta: cdv2.IdentityObjectMeta{
			Name: "blueprint-res",
		},
	}

	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			installation := buildInstallation("test-installation", tt.cd, blueprintResource, blueprint)
			assert.Equal(t, tt.expectedInstallation, installation)
		})
	}

}
