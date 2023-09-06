package util

import (
	"errors"
	"testing"

	cdv2 "github.com/gardener/component-spec/bindings-go/apis/v2"
	"github.com/gardener/landscaper/apis/mediatype"
	modeltypes "github.com/gardener/landscaper/pkg/components/model/types"
	"github.com/stretchr/testify/assert"
)

func TestGetBlueprintResource(t *testing.T) {
	tests := []struct {
		name                 string
		cd                   *modeltypes.ComponentDescriptor
		resourceName         string
		expectedResourceName string
		expectedErr          error
	}{
		{
			name: "only one blueprint in the cd",
			cd: &cdv2.ComponentDescriptor{
				ComponentSpec: cdv2.ComponentSpec{
					Resources: []cdv2.Resource{
						{
							IdentityObjectMeta: cdv2.IdentityObjectMeta{
								Name:    "my-blueprint",
								Version: "v0.1.0",
								Type:    mediatype.BlueprintType,
							},
						},
					},
				},
			},
			expectedResourceName: "my-blueprint",
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
								Type:    mediatype.OldBlueprintType,
							},
						},
					},
				},
			},
			expectedResourceName: "my-blueprint",
		},
		{
			name: "multiple blueprints in the component descriptor",
			cd: &cdv2.ComponentDescriptor{
				ComponentSpec: cdv2.ComponentSpec{
					Resources: []cdv2.Resource{
						{
							IdentityObjectMeta: cdv2.IdentityObjectMeta{
								Name:    "my-blueprint",
								Version: "v0.1.0",
								Type:    mediatype.BlueprintType,
							},
						},
						{
							IdentityObjectMeta: cdv2.IdentityObjectMeta{
								Name:    "my-blueprint-2",
								Version: "v0.1.0",
								Type:    mediatype.BlueprintType,
							},
						},
					},
				},
			},
			expectedResourceName: "",
			expectedErr:          errors.New("the blueprint resource name must be defined since multiple blueprint resources exist in the component descriptor"),
		},
		{
			name: "no blueprint in the component descriptor",
			cd: &cdv2.ComponentDescriptor{
				ComponentSpec: cdv2.ComponentSpec{
					Resources: []cdv2.Resource{},
				},
			},
			expectedResourceName: "",
			expectedErr:          errors.New("no blueprint resources defined in the component descriptor"),
		},
	}

	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			blueprintResourceName, err := GetBlueprintResourceName(tt.cd)
			assert.Equal(t, tt.expectedErr, err)
			assert.Equal(t, tt.expectedResourceName, blueprintResourceName)
		})
	}
}
