package resolver

import (
	"encoding/json"

	"github.com/gardener/component-cli/pkg/commands/componentarchive/input"
	cdresources "github.com/gardener/component-cli/pkg/commands/componentarchive/resources"
	cdv2 "github.com/gardener/component-spec/bindings-go/apis/v2"
	v2 "github.com/gardener/component-spec/bindings-go/apis/v2"
)

// AddLocalResourcesForRender adds the resources from the resources yaml file to the component descriptor.
func AddLocalResourcesForRender(cd *cdv2.ComponentDescriptor, resources []cdresources.ResourceOptions) (*cdv2.ComponentDescriptor, error) {
	newCD := cd.DeepCopy()

	for _, resource := range resources {
		if resource.Input != nil {
			access, err := convertInputToAccess(resource.Input)
			if err != nil {
				return nil, err
			}

			newRes := resource.Resource.DeepCopy()
			newRes.Access = access
			setAppendResource(newCD, *newRes)
		} else {
			setAppendResource(newCD, resource.Resource)
		}
	}

	return newCD, nil
}

func setAppendResource(cd *cdv2.ComponentDescriptor, resource cdv2.Resource) {
	for i := range cd.Resources {
		if cd.Resources[i].Name == resource.Name {
			cd.Resources[i] = resource
			return
		}
	}

	cd.Resources = append(cd.Resources, resource)
}

const LocalFilesystemResourceType = "localFilesystemResource"

// localFilesystemResourceAccess describes the access for a resource on the local file system
// during the execution of command "landscaper-cli blueprints render ...".
type localFilesystemResourceAccess struct {
	cdv2.ObjectType `json:",inline"`
	Input           *input.BlobInput `json:"input"`
}

func (_ *localFilesystemResourceAccess) GetType() string {
	return LocalFilesystemResourceType
}

func convertInputToAccess(in *input.BlobInput) (*v2.UnstructuredTypedObject, error) {
	acc := localFilesystemResourceAccess{
		ObjectType: cdv2.ObjectType{
			Type: LocalFilesystemResourceType,
		},
		Input: in,
	}

	raw, err := json.Marshal(acc)
	if err != nil {
		return nil, err
	}

	obj := map[string]interface{}{}
	err = json.Unmarshal(raw, &obj)
	if err != nil {
		return nil, err
	}

	return &v2.UnstructuredTypedObject{
		ObjectType: v2.ObjectType{Type: LocalFilesystemResourceType},
		Raw:        raw,
		Object:     obj,
	}, nil

}

func convertAccessToInput(access *v2.UnstructuredTypedObject) (*input.BlobInput, error) {
	var raw []byte

	if len(access.Raw) > 0 {
		raw = access.Raw
	} else {
		var err error
		raw, err = json.Marshal(access.Object)
		if err != nil {
			return nil, err
		}
	}

	obj := localFilesystemResourceAccess{}
	err := json.Unmarshal(raw, &obj)
	if err != nil {
		return nil, err
	}

	return obj.Input, nil
}
