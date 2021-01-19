package imagevector

import (
	"encoding/json"
	"fmt"
	"io"
	"strings"

	"github.com/docker/distribution/reference"
	cdv2 "github.com/gardener/component-spec/bindings-go/apis/v2"
	"github.com/gardener/component-spec/bindings-go/apis/v2/cdutils"
	"k8s.io/apimachinery/pkg/util/yaml"
)

// ParseImageOptions are options to configure the image vector parsing.
type ParseImageOptions struct {
	// ComponentReferencePrefixes are prefixes that are used to identify images from other components.
	// These images are then not added as direct resources but the source repository is used as the component reference.
	ComponentReferencePrefixes []string

	// GenericDependencies define images that should be untouched and not added as real dependency to the component descriptors.
	// These dependencies are added a specific label to the component descriptor.
	GenericDependencies []string
}

// ParseImageVector parses a image vector and generates the corresponding component descriptor resources.
func ParseImageVector(cd *cdv2.ComponentDescriptor, reader io.Reader, opts *ParseImageOptions) error {
	decoder := yaml.NewYAMLOrJSONDecoder(reader, 1024)

	imageVector := &ImageVector{}
	if err := decoder.Decode(imageVector); err != nil {
		return fmt.Errorf("unable to decode image vector: %w", err)
	}

	genericImageVector := &ImageVector{}
	for _, image := range imageVector.Images {
		if entryMatchesPrefix(opts.GenericDependencies, image.Name) {
			genericImageVector.Images = append(genericImageVector.Images, image)
			continue
		}
		if image.Tag == nil {
			continue
		}

		if entryMatchesPrefix(opts.ComponentReferencePrefixes, image.Repository) {
			// add image as component reference
			ref := cdv2.ComponentReference{
				Name:          image.Name,
				ComponentName: image.SourceRepository,
				Version:       *image.Tag,
				ExtraIdentity: map[string]string{
					TagExtraIdentity: *image.Tag,
				},
				Labels: make([]cdv2.Label, 0),
			}
			// add complete image as label
			var err error
			cd.ComponentReferences, err = addComponentReference(cd.ComponentReferences, ref, image)
			if err != nil {
				return fmt.Errorf("unable to add component reference for %q: %w", image.Name, err)
			}
			continue
		}

		res := cdv2.Resource{
			IdentityObjectMeta: cdv2.IdentityObjectMeta{
				Labels: make([]cdv2.Label, 0),
			},
		}
		res.Name = image.Name
		res.Version = *image.Tag
		res.Type = cdv2.OCIImageType
		res.Relation = cdv2.ExternalRelation

		var err error
		res.Labels, err = cdutils.SetLabel(res.Labels, NameLabel, image.Name)
		if err != nil {
			return fmt.Errorf("unable to add name label to resource for image %q: %w", image.Name, err)
		}

		for _, label := range image.Labels {
			res.Labels = cdutils.SetRawLabel(res.Labels, label.Name, label.Value)
		}

		if len(image.Repository) != 0 {
			res.Labels, err = cdutils.SetLabel(res.Labels, RepositoryLabel, image.Repository)
			if err != nil {
				return fmt.Errorf("unable to add repository label to resource for image %q: %w", image.Name, err)
			}
		}
		if len(image.SourceRepository) != 0 {
			res.Labels, err = cdutils.SetLabel(res.Labels, SourceRepositoryLabel, image.SourceRepository)
			if err != nil {
				return fmt.Errorf("unable to add source repository label to resource for image %q: %w", image.Name, err)
			}
		}
		if image.TargetVersion != nil {
			res.Labels, err = cdutils.SetLabel(res.Labels, TargetVersionLabel, image.TargetVersion)
			if err != nil {
				return fmt.Errorf("unable to add target version label to resource for image %q: %w", image.Name, err)
			}
		}
		if image.RuntimeVersion != nil {
			res.Labels, err = cdutils.SetLabel(res.Labels, RuntimeVersionLabel, image.RuntimeVersion)
			if err != nil {
				return fmt.Errorf("unable to add target version label to resource for image %q: %w", image.Name, err)
			}
		}

		// set the tag as identity
		cdutils.SetExtraIdentityField(&res.IdentityObjectMeta, TagExtraIdentity, *image.Tag)

		// todo: also consider digests
		ociImageAccess := cdv2.NewOCIRegistryAccess(image.Repository + ":" + *image.Tag)
		uObj, err := cdutils.ToUnstructuredTypedObject(cdv2.DefaultJSONTypedObjectCodec, ociImageAccess)
		if err != nil {
			return fmt.Errorf("unable to create oci registry access for %q: %w", image.Name, err)
		}
		res.Access = uObj

		// add resource
		id := cd.GetResourceIndex(res)
		if id != -1 {
			cd.Resources[id] = cdutils.MergeResources(cd.Resources[id], res)
		} else {
			cd.Resources = append(cd.Resources, res)
		}
	}

	// parse label
	if len(genericImageVector.Images) != 0 {
		genericImageVectorBytes, err := json.Marshal(genericImageVector)
		if err != nil {
			return fmt.Errorf("unable to parse generic image vector: %w", err)
		}
		cd.Labels = cdutils.SetRawLabel(cd.Labels,
			ImagesLabel, genericImageVectorBytes)
	}

	return nil
}

// addComponentReference adds the given component to the list of component references.
// if the component is already declared, the given image entry is appended to the images label
func addComponentReference(refs []cdv2.ComponentReference, new cdv2.ComponentReference, entry ImageEntry) ([]cdv2.ComponentReference, error) {
	for i, ref := range refs {
		if ref.Name == new.Name && ref.Version == new.Version {
			// parse current images and add the image
			imageVector := &ImageVector{
				Images: []ImageEntry{entry},
			}
			data, ok := ref.GetLabels().Get(ImagesLabel)
			if ok {
				if err := json.Unmarshal(data, imageVector); err != nil {
					return nil, err
				}
				imageVector.Images = append(imageVector.Images, entry)
			}
			var err error
			ref.Labels, err = cdutils.SetLabel(ref.Labels, ImagesLabel, imageVector)
			if err != nil {
				return nil, err
			}
			refs[i] = ref
			return refs, nil
		}
	}

	imageVector := ImageVector{
		Images: []ImageEntry{entry},
	}
	var err error
	new.Labels, err = cdutils.SetLabel(new.Labels, ImagesLabel, imageVector)
	if err != nil {
		return nil, err
	}
	return append(refs, new), nil
}

// parseResourceAccess parses a resource's access and sets the repository and tag on the given image entry.
// Currently only access of type 'ociRegistry' is supported.
func parseResourceAccess(imageEntry *ImageEntry, res cdv2.Resource) error {
	access := &cdv2.OCIRegistryAccess{}
	if err := cdv2.NewCodec(nil, nil, nil).Decode(res.Access.Raw, access); err != nil {
		return fmt.Errorf("unable to decode ociRegistry access: %w", err)
	}

	ref, err := reference.Parse(access.ImageReference)
	if err != nil {
		return fmt.Errorf("unable to parse image reference %q: %w", access.ImageReference, err)
	}

	named, ok := ref.(reference.Named)
	if !ok {
		return fmt.Errorf("unable to get repository for %q", ref.String())
	}
	imageEntry.Repository = named.Name()

	switch r := ref.(type) {
	case reference.Tagged:
		tag := r.Tag()
		imageEntry.Tag = &tag
	case reference.Digested:
		tag := r.Digest().String()
		imageEntry.Tag = &tag
	}
	return nil
}

func getLabel(labels cdv2.Labels, name string, into interface{}) (bool, error) {
	val, ok := labels.Get(name)
	if !ok {
		return false, nil
	}

	if err := json.Unmarshal(val, into); err != nil {
		return false, err
	}
	return true, nil
}

func entryMatchesPrefix(prefixes []string, val string) bool {
	for _, pref := range prefixes {
		if strings.HasPrefix(val, pref) {
			return true
		}
	}
	return false
}
