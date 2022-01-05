package resolver

import (
	"context"

	"github.com/mandelsoft/vfs/pkg/vfs"

	cdresources "github.com/gardener/component-cli/pkg/commands/componentarchive/resources"
	cdv2 "github.com/gardener/component-spec/bindings-go/apis/v2"
	v2 "github.com/gardener/component-spec/bindings-go/apis/v2"
	"github.com/gardener/component-spec/bindings-go/ctf"
)

func NewRenderComponentResolver(
	innerResolver ctf.ComponentResolver,
	cd *cdv2.ComponentDescriptor,
	resources []cdresources.ResourceOptions,
	resourcesPath string,
	fs vfs.FileSystem,
) *renderComponentResolver {
	return &renderComponentResolver{
		innerResolver: innerResolver,
		cd:            cd,
		resources:     resources,
		resourcesPath: resourcesPath,
		fs:            fs,
	}
}

type renderComponentResolver struct {
	innerResolver ctf.ComponentResolver
	cd            *cdv2.ComponentDescriptor
	resources     []cdresources.ResourceOptions
	resourcesPath string
	fs            vfs.FileSystem
}

func (r *renderComponentResolver) Resolve(ctx context.Context, repoCtx v2.Repository, name, version string) (*v2.ComponentDescriptor, error) {
	if !r.isOwnComponent(name, version) {
		return r.innerResolver.Resolve(ctx, repoCtx, name, version)
	}

	return r.cd, nil
}

func (r *renderComponentResolver) ResolveWithBlobResolver(ctx context.Context, repoCtx v2.Repository, name, version string) (
	*cdv2.ComponentDescriptor, ctf.BlobResolver, error,
) {
	if !r.isOwnComponent(name, version) {
		return r.innerResolver.ResolveWithBlobResolver(ctx, repoCtx, name, version)
	}

	// Does this fail if the component descriptor exists only locally?
	_, innerBlobResolver, err := r.innerResolver.ResolveWithBlobResolver(ctx, repoCtx, name, version)
	if err != nil {
		innerBlobResolver = nil
	}

	blobResolver := NewRenderBlobResolver(innerBlobResolver, r.resourcesPath, r.fs)

	return r.cd, blobResolver, nil
}

func (r *renderComponentResolver) isOwnComponent(name, version string) bool {
	return name == r.cd.ComponentSpec.Name && version == r.cd.ComponentSpec.Version
}
