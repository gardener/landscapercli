package resolver

import (
	"context"

	cdv2 "github.com/gardener/component-spec/bindings-go/apis/v2"
	lsv1alpha1 "github.com/gardener/landscaper/apis/core/v1alpha1"
	"github.com/gardener/landscaper/pkg/components/model"
	"github.com/mandelsoft/vfs/pkg/vfs"
)

func NewRenderRegistryAccess(
	innerRegistryAccess model.RegistryAccess,
	cd *cdv2.ComponentDescriptor,
	cdList *cdv2.ComponentDescriptorList,
	resourcesPath string,
	fs vfs.FileSystem,
) *renderRegistryAccess {
	return &renderRegistryAccess{
		innerRegistryAccess: innerRegistryAccess,
		cd:                  cd,
		cdList:              cdList,
		resourcesPath:       resourcesPath,
		fs:                  fs,
	}
}

type renderRegistryAccess struct {
	innerRegistryAccess model.RegistryAccess
	cd                  *cdv2.ComponentDescriptor
	cdList              *cdv2.ComponentDescriptorList
	resourcesPath       string
	fs                  vfs.FileSystem
}

var _ model.RegistryAccess = &renderRegistryAccess{}

func (r *renderRegistryAccess) GetComponentVersion(ctx context.Context, cdRef *lsv1alpha1.ComponentDescriptorReference) (model.ComponentVersion, error) {
	cd := r.getOwnComponentDescriptor(cdRef.ComponentName, cdRef.Version)
	if cd == nil {
		return r.innerRegistryAccess.GetComponentVersion(ctx, cdRef)
	}

	var (
		innerComponentVersion model.ComponentVersion
		err                   error
	)

	if cdRef.RepositoryContext != nil {
		innerComponentVersion, err = r.innerRegistryAccess.GetComponentVersion(ctx, cdRef)
		if err != nil {
			innerComponentVersion = nil
		}
	}

	return NewRenderComponentVersion(innerComponentVersion, r, cd, r.resourcesPath, r.fs), nil
}

func (r *renderRegistryAccess) getOwnComponentDescriptor(name, version string) *cdv2.ComponentDescriptor {
	if r.cd != nil && name == r.cd.ComponentSpec.Name && version == r.cd.ComponentSpec.Version {
		return r.cd
	}

	if r.cdList != nil {
		for _, cd := range r.cdList.Components {
			if name == cd.ComponentSpec.Name && version == cd.ComponentSpec.Version {
				return &cd
			}
		}
	}

	return nil
}
