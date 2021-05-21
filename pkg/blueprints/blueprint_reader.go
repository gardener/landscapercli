package blueprints

import (
	"io/ioutil"
	"path/filepath"

	"github.com/gardener/landscaper/apis/core/v1alpha1"
	"github.com/gardener/landscaper/pkg/api"
	"k8s.io/apimachinery/pkg/runtime/serializer"
)

type BlueprintReader struct {
	// path of the blueprint directory, which contains the blueprint.yaml file
	blueprintPath string
}

func NewBlueprintReader(blueprintPath string) *BlueprintReader {
	return &BlueprintReader{
		blueprintPath: blueprintPath,
	}
}

func (r *BlueprintReader) Read() (*v1alpha1.Blueprint, error) {
	blueprintFilePath := filepath.Join(r.blueprintPath, v1alpha1.BlueprintFileName)
	data, err := ioutil.ReadFile(blueprintFilePath)
	if err != nil {
		return nil, err
	}

	blueprint := &v1alpha1.Blueprint{}
	if _, _, err := serializer.NewCodecFactory(api.LandscaperScheme).UniversalDecoder().Decode(data, nil, blueprint); err != nil {
		return nil, err
	}

	return blueprint, nil
}
