package blueprints

import (
	"os"
	"path/filepath"

	"github.com/gardener/landscaper/pkg/apis/core/v1alpha1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"sigs.k8s.io/yaml"
)

type BlueprintWriter struct {
	// path of the blueprint directory, which contains the blueprint.yaml file
	blueprintPath string
}

func NewBlueprintWriter(blueprintPath string) *BlueprintWriter {
	return &BlueprintWriter{
		blueprintPath: blueprintPath,
	}
}

func (w *BlueprintWriter) Write(blueprint *v1alpha1.Blueprint) error {
	blueprintFilePath := filepath.Join(w.blueprintPath, v1alpha1.BlueprintFileName)
	f, err := os.Create(blueprintFilePath)
	if err != nil {
		return err
	}

	defer f.Close()

	if err := w.writeTypeMeta(f); err != nil {
		return err
	}

	if err := w.writeList("imports", blueprint.Imports, len(blueprint.Imports), f); err != nil {
		return err
	}

	if err := w.writeList("exports", blueprint.Exports, len(blueprint.Exports), f); err != nil {
		return err
	}

	if err := w.writeList("exportExecutions", blueprint.ExportExecutions, len(blueprint.ExportExecutions), f); err != nil {
		return err
	}

	if err := w.writeList("subinstallations", blueprint.Subinstallations, len(blueprint.Subinstallations), f); err != nil {
		return err
	}

	if err := w.writeList("deployExecutions", blueprint.DeployExecutions, len(blueprint.DeployExecutions), f); err != nil {
		return err
	}

	return nil
}

func (w *BlueprintWriter) writeTypeMeta(f *os.File) error {
	typeMeta := metav1.TypeMeta{
		APIVersion: "landscaper.gardener.cloud/v1alpha1",
		Kind:       "Blueprint",
	}

	data, err := yaml.Marshal(&typeMeta)
	if err != nil {
		return err
	}

	_, err = f.Write(data)
	return err
}

func (w *BlueprintWriter) writeList(name string, list interface{}, length int, f *os.File) error {
	if _, err := f.WriteString("\n"); err != nil {
		return err
	}

	if length == 0 {
		list = []interface{}{}
	}

	m := map[string]interface{}{
		name: list,
	}

	data, err := yaml.Marshal(m)
	if err != nil {
		return err
	}

	_, err = f.Write(data)
	return err
}
