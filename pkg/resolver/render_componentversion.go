package resolver

import (
	"context"
	"fmt"

	cdv2 "github.com/gardener/component-spec/bindings-go/apis/v2"
	"github.com/gardener/component-spec/bindings-go/ctf"
	lsv1alpha1 "github.com/gardener/landscaper/apis/core/v1alpha1"
	"github.com/gardener/landscaper/pkg/components/cnudie"
	"github.com/gardener/landscaper/pkg/components/model"
	"github.com/gardener/landscaper/pkg/components/model/componentoverwrites"
	"github.com/gardener/landscaper/pkg/components/model/types"
	"github.com/mandelsoft/vfs/pkg/vfs"
)

func NewRenderComponentVersion(
	innerComponentVersion model.ComponentVersion,
	registryAccess model.RegistryAccess,
	componentDescriptor *cdv2.ComponentDescriptor,
	resourcesPath string,
	fs vfs.FileSystem,
) *renderComponentVersion {
	return &renderComponentVersion{
		innerComponentVersion: innerComponentVersion,
		registryAccess:        registryAccess,
		componentDescriptor:   componentDescriptor,
		resourcesPath:         resourcesPath,
		fs:                    fs,
	}
}

type renderComponentVersion struct {
	innerComponentVersion model.ComponentVersion
	registryAccess        model.RegistryAccess
	componentDescriptor   *cdv2.ComponentDescriptor
	resourcesPath         string
	fs                    vfs.FileSystem
}

var _ model.ComponentVersion = &renderComponentVersion{}

func (c *renderComponentVersion) GetName() string {
	return c.componentDescriptor.GetName()
}

func (c *renderComponentVersion) GetVersion() string {
	return c.componentDescriptor.GetVersion()
}

func (c *renderComponentVersion) GetComponentDescriptor() (*types.ComponentDescriptor, error) {
	var componentDescriptor types.ComponentDescriptor = *c.componentDescriptor
	return &componentDescriptor, nil
}

func (c *renderComponentVersion) GetRepositoryContext() (*types.UnstructuredTypedObject, error) {
	return c.componentDescriptor.GetEffectiveRepositoryContext(), nil
}

func (c *renderComponentVersion) GetComponentReferences() ([]types.ComponentReference, error) {
	return c.componentDescriptor.ComponentReferences, nil
}

func (c *renderComponentVersion) GetComponentReference(name string) (*types.ComponentReference, error) {
	refs, err := c.GetComponentReferences()
	if err != nil {
		return nil, err
	}

	for i := range refs {
		ref := &refs[i]
		if ref.GetName() == name {
			return ref, nil
		}
	}

	return nil, nil
}

func (c *renderComponentVersion) GetReferencedComponentVersion(ctx context.Context, componentRef *types.ComponentReference, repositoryContext *types.UnstructuredTypedObject, overwriter componentoverwrites.Overwriter) (model.ComponentVersion, error) {
	cdRef := &lsv1alpha1.ComponentDescriptorReference{
		RepositoryContext: repositoryContext,
		ComponentName:     componentRef.ComponentName,
		Version:           componentRef.Version,
	}

	return model.GetComponentVersionWithOverwriter(ctx, c.registryAccess, cdRef, overwriter)
}

func (c *renderComponentVersion) GetResource(name string, selectors map[string]string) (model.Resource, error) {
	resources, err := c.componentDescriptor.GetResourcesByName(name, cdv2.Identity(selectors))
	if err != nil {
		return nil, err
	}
	if len(resources) < 1 {
		return nil, fmt.Errorf("no resource with name %s and extra identities %v found", name, selectors)
	}
	if len(resources) > 1 {
		return nil, fmt.Errorf("there is more than one resource with name %s and extra identities %v", name, selectors)
	}

	blobResolver, err := c.GetBlobResolver()
	if err != nil {
		return nil, err
	}

	return cnudie.NewResource(&resources[0], blobResolver), nil
}

func (c *renderComponentVersion) GetBlobResolver() (model.BlobResolver, error) {
	var (
		innerBlobResolver ctf.BlobResolver
		err               error
	)

	if c.innerComponentVersion != nil {
		innerBlobResolver, err = c.innerComponentVersion.GetBlobResolver()
		if err != nil {
			return nil, err
		}
	} else {
		innerBlobResolver = emptyBlobResolver{}
	}

	return NewRenderBlobResolver(innerBlobResolver, c.resourcesPath, c.fs), nil
}
