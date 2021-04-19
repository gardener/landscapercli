package util

import (
	"errors"
	"testing"

	cdv2 "github.com/gardener/component-spec/bindings-go/apis/v2"
	lscore "github.com/gardener/landscaper/apis/core"
	"github.com/stretchr/testify/assert"
)

func TestGetBlueprintResource(t *testing.T) {
	tests := []struct {
		name             string
		cd               *cdv2.ComponentDescriptor
		resourceName     string
		expectedResource *cdv2.Resource
		expectedErr      error
	}{
		{
			name: "resource name not specified with only one blueprint in the cd",
			cd: &cdv2.ComponentDescriptor{
				ComponentSpec: cdv2.ComponentSpec{
					Resources: []cdv2.Resource{
						{
							IdentityObjectMeta: cdv2.IdentityObjectMeta{
								Name:    "my-blueprint",
								Version: "v0.1.0",
								Type:    lscore.BlueprintType,
							},
						},
					},
				},
			},
			resourceName: "",
			expectedResource: &cdv2.Resource{
				IdentityObjectMeta: cdv2.IdentityObjectMeta{
					Name:    "my-blueprint",
					Version: "v0.1.0",
					Type:    lscore.BlueprintType,
				},
			},
		},
		{
			name: "old blueprint type",
			cd: &cdv2.ComponentDescriptor{
				ComponentSpec: cdv2.ComponentSpec{
					Resources: []cdv2.Resource{
						{
							IdentityObjectMeta: cdv2.IdentityObjectMeta{
								Name:    "my-blueprint",
								Version: "v0.1.0",
								Type:    lscore.OldBlueprintType,
							},
						},
					},
				},
			},
			resourceName: "",
			expectedResource: &cdv2.Resource{
				IdentityObjectMeta: cdv2.IdentityObjectMeta{
					Name:    "my-blueprint",
					Version: "v0.1.0",
					Type:    lscore.OldBlueprintType,
				},
			},
		},
		{
			name: "resource name explicitly specified",
			cd: &cdv2.ComponentDescriptor{
				ComponentSpec: cdv2.ComponentSpec{
					Resources: []cdv2.Resource{
						{
							IdentityObjectMeta: cdv2.IdentityObjectMeta{
								Name:    "my-blueprint",
								Version: "v0.1.0",
								Type:    lscore.BlueprintType,
							},
						},
						{
							IdentityObjectMeta: cdv2.IdentityObjectMeta{
								Name:    "my-blueprint-2",
								Version: "v0.1.0",
								Type:    lscore.BlueprintType,
							},
						},
					},
				},
			},
			resourceName: "my-blueprint",
			expectedResource: &cdv2.Resource{
				IdentityObjectMeta: cdv2.IdentityObjectMeta{
					Name:    "my-blueprint",
					Version: "v0.1.0",
					Type:    lscore.BlueprintType,
				},
			},
		},
		{
			name: "invalid resource name",
			cd: &cdv2.ComponentDescriptor{
				ComponentSpec: cdv2.ComponentSpec{
					Resources: []cdv2.Resource{
						{
							IdentityObjectMeta: cdv2.IdentityObjectMeta{
								Name:    "my-blueprint",
								Version: "v0.1.0",
								Type:    lscore.BlueprintType,
							},
						},
						{
							IdentityObjectMeta: cdv2.IdentityObjectMeta{
								Name:    "my-blueprint-2",
								Version: "v0.1.0",
								Type:    lscore.BlueprintType,
							},
						},
					},
				},
			},
			resourceName: "my-blueprint-3",
			expectedErr:  errors.New("blueprint my-blueprint-3 is not defined as a resource in the component descriptor"),
		},
		{
			name: "resource name not specified with multiple blueprints in the cd",
			cd: &cdv2.ComponentDescriptor{
				ComponentSpec: cdv2.ComponentSpec{
					Resources: []cdv2.Resource{
						{
							IdentityObjectMeta: cdv2.IdentityObjectMeta{
								Name:    "my-blueprint",
								Version: "v0.1.0",
								Type:    lscore.BlueprintType,
							},
						},
						{
							IdentityObjectMeta: cdv2.IdentityObjectMeta{
								Name:    "my-blueprint-2",
								Version: "v0.1.0",
								Type:    lscore.BlueprintType,
							},
						},
					},
				},
			},
			resourceName: "",
			expectedErr:  errors.New("the blueprint resource name must be defined since multiple blueprint resources exist in the component descriptor"),
		},
		{
			name: "no blueprint resources defined in the cd",
			cd: &cdv2.ComponentDescriptor{
				ComponentSpec: cdv2.ComponentSpec{
					Resources: []cdv2.Resource{},
				},
			},
			resourceName: "",
			expectedErr:  errors.New("no blueprint resources defined in the component descriptor"),
		},
	}

	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			res, err := GetBlueprintResource(tt.cd, tt.resourceName)
			assert.Equal(t, tt.expectedErr, err)
			assert.Equal(t, tt.expectedResource, res)
		})
	}
}
