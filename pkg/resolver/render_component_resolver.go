package resolver

import (
	"context"
	"fmt"
	"io"

	"github.com/mandelsoft/vfs/pkg/vfs"

	cdv2 "github.com/gardener/component-spec/bindings-go/apis/v2"
	v2 "github.com/gardener/component-spec/bindings-go/apis/v2"
	"github.com/gardener/component-spec/bindings-go/ctf"
)

func NewRenderComponentResolver(
	innerResolver ctf.ComponentResolver,
	cd *cdv2.ComponentDescriptor,
	cdList *cdv2.ComponentDescriptorList,
	resourcesPath string,
	fs vfs.FileSystem,
) *renderComponentResolver {
	return &renderComponentResolver{
		innerResolver: innerResolver,
		cd:            cd,
		cdList:        cdList,
		resourcesPath: resourcesPath,
		fs:            fs,
	}
}

type emptyBlobResolver struct {
}

func (r emptyBlobResolver) Info(_ context.Context, res v2.Resource) (*ctf.BlobInfo, error) {
	return nil, fmt.Errorf("trying to get info on non-local resource %s with local blob resolver", res.Name)
}

func (r emptyBlobResolver) Resolve(_ context.Context, res v2.Resource, _ io.Writer) (*ctf.BlobInfo, error) {
	return nil, fmt.Errorf("trying to resolve non-local resource %s with local blob resolver", res.Name)
}

type renderComponentResolver struct {
	innerResolver ctf.ComponentResolver
	cd            *cdv2.ComponentDescriptor
	cdList        *cdv2.ComponentDescriptorList
	resourcesPath string
	fs            vfs.FileSystem
}

func (r *renderComponentResolver) Resolve(ctx context.Context, repoCtx v2.Repository, name, version string) (*v2.ComponentDescriptor, error) {
	cd := r.getOwnComponent(name, version)
	if cd == nil {
		return r.innerResolver.Resolve(ctx, repoCtx, name, version)
	}

	return cd, nil
}

func (r *renderComponentResolver) ResolveWithBlobResolver(ctx context.Context, repoCtx v2.Repository, name, version string) (
	*cdv2.ComponentDescriptor, ctf.BlobResolver, error,
) {
	cd := r.getOwnComponent(name, version)
	if cd == nil {
		return r.innerResolver.ResolveWithBlobResolver(ctx, repoCtx, name, version)
	}

	// Does this fail if the component descriptor exists only locally?
	_, innerBlobResolver, err := r.innerResolver.ResolveWithBlobResolver(ctx, repoCtx, name, version)
	if err != nil {
		innerBlobResolver = emptyBlobResolver{}
	}

	blobResolver := NewRenderBlobResolver(innerBlobResolver, r.resourcesPath, r.fs)

	return cd, blobResolver, nil
}

func (r *renderComponentResolver) getOwnComponent(name, version string) *cdv2.ComponentDescriptor {
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
