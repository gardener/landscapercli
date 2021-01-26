package blueprints

import (
	"io/ioutil"

	"github.com/gardener/landscapercli/pkg/util"

	cd "github.com/gardener/component-spec/bindings-go/apis/v2"
	"sigs.k8s.io/yaml"
)

type ComponentDescriptorReader struct {
	// path of the component directory, which contains the component-descriptor.yaml file
	componentPath string
}

func NewComponentDescriptorReader(componentPath string) *ComponentDescriptorReader {
	return &ComponentDescriptorReader{
		componentPath: componentPath,
	}
}

func (r *ComponentDescriptorReader) Read() (*cd.ComponentDescriptor, error) {
	componentDescriptorFilePath := util.ComponentDescriptorFilePath(r.componentPath)
	data, err := ioutil.ReadFile(componentDescriptorFilePath)
	if err != nil {
		return nil, err
	}

	componentDescriptor := &cd.ComponentDescriptor{}
	err = yaml.Unmarshal(data, componentDescriptor)
	if err != nil {
		return nil, err
	}

	return componentDescriptor, nil
}
