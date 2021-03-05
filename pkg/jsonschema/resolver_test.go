package jsonschema

import (
	"context"
	"path"
	"testing"

	v2 "github.com/gardener/component-spec/bindings-go/apis/v2"
	"github.com/gardener/landscaper/pkg/landscaper/jsonschema"
	componentsregistry "github.com/gardener/landscaper/pkg/landscaper/registry/components"
	"github.com/gardener/landscaper/pkg/landscaper/registry/components/cdutils"
	"github.com/gardener/landscapercli/pkg/blueprints"
	logrtesting "github.com/go-logr/logr/testing"
	"github.com/mandelsoft/vfs/pkg/osfs"
	"github.com/mandelsoft/vfs/pkg/projectionfs"
	"github.com/stretchr/testify/assert"
)

func TestResolve(t *testing.T) {
	tests := []struct {
		name            string
		componentName   string
		assetsPath      string
		expectedSchemas JSONSchemaList
	}{
		{
			name:       "inline",
			componentName: "example.com/inline",
			assetsPath: "./testdata/registry/inline",
			expectedSchemas: JSONSchemaList{
				{
					Ref: "root",
					Schema: map[string]interface{}{
						"$schema": "https://json-schema.org/draft/2019-09/schema",
						"type":    "string",
					},
				},
			},
		},
		{
			name:       "local-types",
			componentName: "example.com/local-types",
			assetsPath: "./testdata/registry/local-types",
			expectedSchemas: JSONSchemaList{
				{
					Ref: "root",
					Schema: map[string]interface{}{
						"$ref": "local://my-type",
					},
				},
				{
					Ref: "local://my-type",
					Schema: map[string]interface{}{
						"type": "string",
					},
				},
			},
		},
		{
			name:       "blueprint-fs",
			componentName: "example.com/blueprint-fs",
			assetsPath: "./testdata/registry/blueprint-fs",
			expectedSchemas: JSONSchemaList{
				{
					Ref: "root",
					Schema: map[string]interface{}{
						"$ref": "blueprint://my-type.json",
					},
				},
				{
					Ref: "blueprint://my-type.json",
					Schema: map[string]interface{}{
						"type": "string",
					},
				},
			},
		},
		// {
		// 	name:       "distinct-component",
		// 	componentName: "example.com/distinct-component/bp-component",
		// 	assetsPath: "./testdata/registry/distinct-component/bp-component",
		// 	expectedSchemas: JSONSchemaList{
		// 		{
		// 			Ref: "root",
		// 			Schema: map[string]interface{}{
		// 				"$ref": "blueprint://my-type.json",
		// 			},
		// 		},
		// 		{
		// 			Ref: "blueprint://my-type.json",
		// 			Schema: map[string]interface{}{
		// 				"type": "string",
		// 			},
		// 		},
		// 	},
		// },
	}

	fakeCompRepo, err := componentsregistry.NewLocalClient(logrtesting.NullLogger{}, "./testdata/registry")
	assert.NoError(t, err)

	repoCtx := v2.RepositoryContext{
		Type:    "local",
		BaseURL: "./testdata",
	}
	refResolver := cdutils.ComponentReferenceResolverFromResolver(fakeCompRepo, repoCtx)

	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			cd, _, err := fakeCompRepo.Resolve(context.TODO(), repoCtx, tt.componentName, "v0.1.0")
			assert.NoError(t, err)

			bpDir := path.Join(tt.assetsPath, "blueprint")

			blueprintFs, err := projectionfs.New(osfs.New(), bpDir)
			assert.NoError(t, err)

			blueprintReader := blueprints.NewBlueprintReader(bpDir)
			bp, err := blueprintReader.Read()
			assert.NoError(t, err)

			loaderConfig:= jsonschema.LoaderConfig{
				LocalTypes:                 bp.LocalTypes,
				BlueprintFs:                blueprintFs,
				ComponentDescriptor:        cd,
				ComponentResolver:          fakeCompRepo,
				ComponentReferenceResolver: refResolver,
			}

			resolver := NewJSONSchemaResolver(&loaderConfig, 2)

			schemas, err := resolver.Resolve(bp.Imports[0].Schema)
			assert.NoError(t, err)

			assert.Equal(t, tt.expectedSchemas, schemas)
		})
	}

}
