package components

import (
	"fmt"
	"os"
	"strconv"

	cdresources "github.com/gardener/component-cli/pkg/commands/componentarchive/resources"
	cd "github.com/gardener/component-spec/bindings-go/apis/v2"

	"github.com/gardener/landscapercli/pkg/util"
)

type ResourceWriter struct {
	// path of the component directory, which contains the component-descriptor.yaml file
	componentPath string
}

func NewResourceWriter(componentPath string) *ResourceWriter {
	return &ResourceWriter{
		componentPath: componentPath,
	}
}

func (w *ResourceWriter) Write(resourceOptions []cdresources.ResourceOptions) error {
	f, err := os.Create(util.ResourcesFilePath(w.componentPath))
	if err != nil {
		return err
	}

	defer f.Close()

	infoString := ""

	for i := range resourceOptions {
		infoString += "---\n" +
			"type: " + resourceOptions[i].Type + "\n" +
			"name: " + resourceOptions[i].Name + "\n" +
			"version: " + resourceOptions[i].Version + "\n" +
			"relation: " + string(resourceOptions[i].Relation) + "\n"

		if resourceOptions[i].Input != nil {
			infoString += "input: \n" +
				"  type: " + string(resourceOptions[i].Input.Type) + "\n" +
				"  path: " + resourceOptions[i].Input.Path + "\n"

			if resourceOptions[i].Input.MediaType != "" {
				infoString += "  mediaType: " + resourceOptions[i].Input.MediaType + "\n"
			}

			infoString += "  compress: " + strconv.FormatBool(resourceOptions[i].Input.Compress()) + "\n"

		} else if resourceOptions[i].Access != nil {
			data := resourceOptions[i].Access.Raw

			ociImageAccess := &cd.OCIRegistryAccess{}
			if err := cd.NewDefaultCodec().Decode(data, ociImageAccess); err != nil {
				return fmt.Errorf("unable to decode resource access into oci registry access for %q: %w", resourceOptions[i].GetName(), err)
			}

			infoString += "access: \n" +
				"  type: " + ociImageAccess.Type + "\n" +
				"  imageReference: " + ociImageAccess.ImageReference + "\n"
		}

		infoString += "...\n"
	}

	_, err = f.WriteString(infoString)
	if err != nil {
		return fmt.Errorf("unable to write resources.yaml: %w", err)
	}

	return nil
}
