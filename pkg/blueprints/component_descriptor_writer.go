package blueprints

import (
	"os"

	cd "github.com/gardener/component-spec/bindings-go/apis/v2"
	"sigs.k8s.io/yaml"

	"github.com/gardener/landscapercli/pkg/util"
)

type ComponentDescriptorWriter struct {
	// path of the component directory, which contains the component-descriptor.yaml file
	componentPath string
}

func NewComponentDescriptorWriter(componentPath string) *ComponentDescriptorWriter {
	return &ComponentDescriptorWriter{
		componentPath: componentPath,
	}
}

func (w *ComponentDescriptorWriter) Write(componentDescriptor *cd.ComponentDescriptor) error {
	f, err := os.Create(util.ComponentDescriptorFilePath(w.componentPath))
	if err != nil {
		return err
	}

	defer f.Close()

	data, err := yaml.Marshal(&componentDescriptor)
	if err != nil {
		return err
	}

	_, err = f.Write(data)
	return err
}
