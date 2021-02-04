package util

import (
	"context"
	"fmt"

	"github.com/gardener/component-cli/ociclient"
	"github.com/gardener/component-cli/ociclient/cache"
	"github.com/gardener/component-cli/pkg/utils"
	cdv2 "github.com/gardener/component-spec/bindings-go/apis/v2"
	"github.com/gardener/component-spec/bindings-go/ctf"
	cdoci "github.com/gardener/component-spec/bindings-go/oci"
	"github.com/gardener/landscapercli/pkg/logger"
	"github.com/go-logr/logr"
)

func UploadComponentArchive(archiveDir, uploadRef string) error {
	// TODO: parse url and get registry base url
	const registryBaseURL = "localhost:5000"

	ctx := context.TODO()
	ociClient, cache, err := buildOCIClient(logger.Log)
	if err != nil {
		return fmt.Errorf("unable to build oci client: %w", err)
	}

	archive, err := ctf.ComponentArchiveFromPath(archiveDir)
	if err != nil {
		return fmt.Errorf("unable to build component archive: %w", err)
	}
	// update repository context
	archive.ComponentDescriptor.RepositoryContexts = utils.AddRepositoryContext(archive.ComponentDescriptor.RepositoryContexts, cdv2.OCIRegistryType, registryBaseURL)

	// manifest, err := cdoci.NewManifestBuilder(cache, archive).Build(ctx)
	manifest, err := cdoci.NewManifestBuilder(cache, archive).Build(ctx)
	if err != nil {
		return fmt.Errorf("unable to build oci artifact for component acrchive: %w", err)
	}

	return ociClient.PushManifest(ctx, uploadRef, manifest)
}

func buildOCIClient(log logr.Logger) (ociclient.Client, cache.Cache, error) {
	cache, err := cache.NewCache(logger.Log)
	if err != nil {
		return nil, nil, err
	}

	ociOpts := []ociclient.Option{
		ociclient.WithCache{Cache: cache},
		ociclient.WithKnownMediaType(cdoci.ComponentDescriptorConfigMimeType),
		ociclient.WithKnownMediaType(cdoci.ComponentDescriptorTarMimeType),
		ociclient.WithKnownMediaType(cdoci.ComponentDescriptorJSONMimeType),
		ociclient.AllowPlainHttp(true),
	}

	ociClient, err := ociclient.NewClient(log, ociOpts...)
	if err != nil {
		return nil, nil, fmt.Errorf("unable to build oci client: %w", err)
	}
	return ociClient, cache, nil
}
