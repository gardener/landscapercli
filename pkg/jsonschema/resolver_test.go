package jsonschema

import (
	"context"
	"path"
	"testing"

	cdv2 "github.com/gardener/component-spec/bindings-go/apis/v2"
	"github.com/gardener/landscaper/pkg/landscaper/jsonschema"
	componentsregistry "github.com/gardener/landscaper/pkg/landscaper/registry/components"
	"github.com/go-logr/logr"
	"github.com/mandelsoft/vfs/pkg/osfs"
	"github.com/mandelsoft/vfs/pkg/projectionfs"
	"github.com/stretchr/testify/assert"

	"github.com/gardener/landscapercli/pkg/blueprints"
)

func TestResolve(t *testing.T) {
	tests := []struct {
		name            string
		componentName   string
		assetsPath      string
		expectedSchemas JSONSchemaList
	}{
		{
			name:          "inline",
			componentName: "example.com/inline",
			assetsPath:    "./testdata/registry/inline",
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
			name:          "local-types",
			componentName: "example.com/local-types",
			assetsPath:    "./testdata/registry/local-types",
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
			name:          "blueprint-fs",
			componentName: "example.com/blueprint-fs",
			assetsPath:    "./testdata/registry/blueprint-fs",
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
		{
			name:          "cyclic reference",
			componentName: "example.com/inline",
			assetsPath:    "./testdata/registry/cyclic-ref",
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
						"$schema": "http://json-schema.org/draft-04/schema#",
						"definitions": map[string]interface{}{
							"author": map[string]interface{}{
								"properties": map[string]interface{}{
									"name": map[string]interface{}{
										"type": "string",
									},
									"co-author": map[string]interface{}{
										"$ref": "#/definitions/author",
									},
								},
							},
						},
						"type": "object",
						"properties": map[string]interface{}{
							"author": map[string]interface{}{
								"$ref": "#/definitions/author",
							},
						},
					},
				},
			},
		},
	}

	fakeCompRepo, err := componentsregistry.NewLocalClient(logr.Discard(), "./testdata/registry")
	assert.NoError(t, err)

	repoCtx, _ := cdv2.NewUnstructured(componentsregistry.NewLocalRepository("./testdata/registry"))

	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			cd, err := fakeCompRepo.Resolve(context.TODO(), &repoCtx, tt.componentName, "v0.1.0")
			assert.NoError(t, err)

			bpDir := path.Join(tt.assetsPath, "blueprint")

			blueprintFs, err := projectionfs.New(osfs.New(), bpDir)
			assert.NoError(t, err)

			blueprintReader := blueprints.NewBlueprintReader(bpDir)
			bp, err := blueprintReader.Read()
			assert.NoError(t, err)

			loaderConfig := jsonschema.LoaderConfig{
				LocalTypes:          bp.LocalTypes,
				BlueprintFs:         blueprintFs,
				ComponentDescriptor: cd,
				ComponentResolver:   fakeCompRepo,
			}

			resolver := NewJSONSchemaResolver(&loaderConfig, 2)

			schemas, err := resolver.Resolve(bp.Imports[0].Schema)
			assert.NoError(t, err)

			assert.Equal(t, tt.expectedSchemas, schemas)
		})
	}
}
