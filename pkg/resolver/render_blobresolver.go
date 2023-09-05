package resolver

import (
	"context"
	"fmt"
	"io"

	v2 "github.com/gardener/component-spec/bindings-go/apis/v2"
	"github.com/gardener/component-spec/bindings-go/ctf"
	"github.com/mandelsoft/vfs/pkg/vfs"
)

func NewRenderBlobResolver(
	innerBlobResolver ctf.BlobResolver,
	resourcesPath string,
	fs vfs.FileSystem,
) *renderBlobResolver {
	return &renderBlobResolver{
		innerBlobResolver: innerBlobResolver,
		resourcesPath:     resourcesPath,
		fs:                fs,
	}
}

type renderBlobResolver struct {
	innerBlobResolver ctf.BlobResolver
	resourcesPath     string
	fs                vfs.FileSystem
}

func (r *renderBlobResolver) Info(ctx context.Context, res v2.Resource) (*ctf.BlobInfo, error) {
	if res.Access.Type != LocalFilesystemResourceType {
		return r.innerBlobResolver.Info(ctx, res)
	}

	blobInput, err := convertAccessToInput(res.Access)
	if err != nil {
		return nil, err
	}
	if blobInput == nil {
		return nil, fmt.Errorf("resource %s has no input definition", res.Name)
	}

	blob, err := blobInput.Read(context.Background(), r.fs, r.resourcesPath)
	if err != nil {
		return nil, err
	}

	return &ctf.BlobInfo{
		MediaType: blobInput.MediaType,
		Digest:    blob.Digest,
		Size:      blob.Size,
	}, nil
}

func (r *renderBlobResolver) Resolve(ctx context.Context, res v2.Resource, writer io.Writer) (*ctf.BlobInfo, error) {
	if res.Access.Type != LocalFilesystemResourceType {
		return r.innerBlobResolver.Resolve(ctx, res, writer)
	}

	blobInput, err := convertAccessToInput(res.Access)
	if err != nil {
		return nil, err
	}

	blob, err := blobInput.Read(context.Background(), r.fs, r.resourcesPath)
	if err != nil {
		return nil, err
	}
	if blobInput == nil {
		return nil, fmt.Errorf("resource %s has no input definition", res.Name)
	}

	_, err = io.Copy(writer, blob.Reader)
	if err != nil {
		return nil, err
	}

	err = blob.Reader.Close()
	if err != nil {
		return nil, err
	}

	return &ctf.BlobInfo{
		MediaType: blobInput.MediaType,
		Digest:    blob.Digest,
		Size:      blob.Size,
	}, nil
}
