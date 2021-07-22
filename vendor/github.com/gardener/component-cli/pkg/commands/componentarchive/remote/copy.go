// SPDX-FileCopyrightText: 2020 SAP SE or an SAP affiliate company and Gardener contributors.
//
// SPDX-License-Identifier: Apache-2.0

package remote

import (
	"context"
	"errors"
	"fmt"
	"io"
	"os"

	cdv2 "github.com/gardener/component-spec/bindings-go/apis/v2"
	"github.com/gardener/component-spec/bindings-go/ctf"
	cdoci "github.com/gardener/component-spec/bindings-go/oci"
	"github.com/go-logr/logr"
	"github.com/mandelsoft/vfs/pkg/osfs"
	"github.com/mandelsoft/vfs/pkg/vfs"
	"github.com/opencontainers/go-digest"
	ocispecv1 "github.com/opencontainers/image-spec/specs-go/v1"
	"github.com/spf13/cobra"
	"github.com/spf13/pflag"

	"github.com/gardener/component-cli/ociclient"
	"github.com/gardener/component-cli/ociclient/cache"

	"github.com/gardener/component-cli/pkg/components"

	ociopts "github.com/gardener/component-cli/ociclient/options"
	"github.com/gardener/component-cli/pkg/logger"
	"github.com/gardener/component-cli/pkg/utils"
)

// CopyOptions contains all options to copy a component descriptor.
type CopyOptions struct {
	ComponentName    string
	ComponentVersion string
	SourceRepository string
	TargetRepository string

	// Recursive specifies if all component references should also be copied.
	Recursive bool
	// Force forces an overwrite in the target registry if the component descriptor is already uploaded.
	Force bool
	// CopyByValue defines if all oci images and artifacts should be copied by value or reference.
	// LocalBlobs are still copied by value.
	CopyByValue bool

	// OciOptions contains all exposed options to configure the oci client.
	OciOptions ociopts.Options
}

// NewCopyCommand creates a new definition command to push definitions
func NewCopyCommand(ctx context.Context) *cobra.Command {
	opts := &CopyOptions{}
	cmd := &cobra.Command{
		Use:   "copy COMPONENT_NAME VERSION --from SOURCE_REPOSITORY --to TARGET_REPOSITORY",
		Args:  cobra.ExactArgs(2),
		Short: "copies a component descriptor from a context repository to another",
		Long: `
copies a component descriptor and its blobs from the source repository to the target repository.

By default the component descriptor and all its component references are recursively copied.
This behavior can be overwritten by specifying "--recursive=false"

`,
		Run: func(cmd *cobra.Command, args []string) {
			if err := opts.Complete(args); err != nil {
				fmt.Println(err.Error())
				os.Exit(1)
			}

			if err := opts.run(ctx, logger.Log, osfs.New()); err != nil {
				logger.Log.Error(err, "")
				os.Exit(1)
			}
		},
	}

	opts.AddFlags(cmd.Flags())

	return cmd
}

func (o *CopyOptions) run(ctx context.Context, log logr.Logger, fs vfs.FileSystem) error {
	ctx = logr.NewContext(ctx, log)
	ociClient, cache, err := o.OciOptions.Build(log, fs)
	if err != nil {
		return fmt.Errorf("unable to build oci client: %s", err.Error())
	}
	defer cache.Close()

	c := Copier{
		SrcRepoCtx:    cdv2.NewOCIRegistryRepository(o.SourceRepository, ""),
		TargetRepoCtx: cdv2.NewOCIRegistryRepository(o.TargetRepository, ""),
		CompResolver:  cdoci.NewResolver(ociClient),
		OciClient:     ociClient,
		Cache:         cache,
		Recursive:     o.Recursive,
		Force:         o.Force,
	}

	if err := c.Copy(ctx, o.ComponentName, o.ComponentVersion); err != nil {
		return err
	}

	fmt.Printf("Successfully copied component descriptor %s:%s from %s to %s\n", o.ComponentName, o.ComponentVersion, o.SourceRepository, o.TargetRepository)
	return nil
}

func (o *CopyOptions) Complete(args []string) error {
	o.ComponentName = args[0]
	o.ComponentVersion = args[1]

	var err error
	o.OciOptions.CacheDir, err = utils.CacheDir()
	if err != nil {
		return fmt.Errorf("unable to get oci cache directory: %w", err)
	}

	if err := o.Validate(); err != nil {
		return err
	}
	return nil
}

// Validate validates push options
func (o *CopyOptions) Validate() error {
	if len(o.SourceRepository) == 0 {
		return errors.New("a source repository has to be specified")
	}
	if len(o.TargetRepository) == 0 {
		return errors.New("a target repository has to be specified")
	}
	if o.CopyByValue {
		return errors.New("copy by value is currently not supported")
	}
	return nil
}

func (o *CopyOptions) AddFlags(fs *pflag.FlagSet) {
	fs.StringVar(&o.SourceRepository, "from", "", "source repository base url.")
	fs.StringVar(&o.TargetRepository, "to", "", "target repository where the components are copied to.")
	fs.BoolVar(&o.Recursive, "recursive", true, "Recursively copy the component descriptor and its references.")
	fs.BoolVar(&o.Force, "force", false, "Forces the tool to overwrite already existing component descriptors.")
	fs.BoolVar(&o.CopyByValue, "copy-by-value", false, "[EXPERIMENTAL] copies all references oci images and artifacts by value and not by reference.")
	o.OciOptions.AddFlags(fs)
}

// Copier copies a component descriptor from a target repo to another.
type Copier struct {
	SrcRepoCtx, TargetRepoCtx cdv2.Repository
	Cache                     cache.Cache
	OciClient                 ociclient.Client
	CompResolver              ctf.ComponentResolver

	// Recursive specifies if all component references should also be copied.
	Recursive bool
	// Force forces an overwrite in the target registry if the component descriptor is already uploaded.
	Force bool
}

func (c *Copier) Copy(ctx context.Context, name, version string) error {
	log := logr.FromContextOrDiscard(ctx).WithValues("component", name, "version", version)
	log.V(3).Info("start copy")
	cd, blobs, err := c.CompResolver.ResolveWithBlobResolver(ctx, c.SrcRepoCtx, name, version)
	if err != nil {
		return err
	}

	if c.Recursive {
		log.V(3).Info("copy referenced components")
		for _, ref := range cd.ComponentReferences {
			if err := c.Copy(ctx, ref.ComponentName, ref.Version); err != nil {
				return err
			}
		}
	}

	// check if the component descriptor already exists
	if !c.Force {
		if _, err := c.CompResolver.Resolve(ctx, c.TargetRepoCtx, name, version); err == nil {
			log.V(3).Info("Component already exists. Nothing to copy.")
			return nil
		}
	}

	if err := cdv2.InjectRepositoryContext(cd, c.TargetRepoCtx); err != nil {
		return fmt.Errorf("unble to inject target repository: %w", err)
	}

	manifest, err := cdoci.NewManifestBuilder(c.Cache, ctf.NewComponentArchive(cd, nil)).Build(ctx)
	if err != nil {
		return fmt.Errorf("unable to build oci artifact for component acrchive: %w", err)
	}

	blobToResource := map[string]*cdv2.Resource{}
	for _, res := range cd.Resources {
		if res.Access.GetType() != cdv2.LocalOCIBlobType {
			// skip
			continue
		}

		localBlob := &cdv2.LocalOCIBlobAccess{}
		if err := res.Access.DecodeInto(localBlob); err != nil {
			return fmt.Errorf("unable to decode resource %s: %w", res.Name, err)
		}
		blobInfo, err := blobs.Info(ctx, res)
		if err != nil {
			return fmt.Errorf("unable to get blob info for resource %s: %w", res.Name, err)
		}
		d, err := digest.Parse(blobInfo.Digest)
		if err != nil {
			return fmt.Errorf("unable to parse digest for resource %s: %w", res.Name, err)
		}
		manifest.Layers = append(manifest.Layers, ocispecv1.Descriptor{
			MediaType: blobInfo.MediaType,
			Digest:    d,
			Size:      blobInfo.Size,
			Annotations: map[string]string{
				"resource": res.Name,
			},
		})
		blobToResource[blobInfo.Digest] = res.DeepCopy()
	}

	ref, err := components.OCIRef(c.TargetRepoCtx, name, version)
	if err != nil {
		return fmt.Errorf("invalid component reference: %w", err)
	}

	store := ociclient.GenericStore(func(ctx context.Context, desc ocispecv1.Descriptor, writer io.Writer) error {
		log := log.WithValues("digest", desc.Digest.String(), "mediaType", desc.MediaType)
		res, ok := blobToResource[desc.Digest.String()]
		if !ok {
			// default to cache
			log.V(4).Info("copying resource from cache")
			rc, err := c.Cache.Get(desc)
			if err != nil {
				return err
			}
			defer func() {
				if err := rc.Close(); err != nil {
					log.Error(err, "unable to close blob reader")
				}
			}()
			if _, err := io.Copy(writer, rc); err != nil {
				return err
			}
			return nil
		}

		log.V(4).Info("copying resource", "resource", res.Name)
		_, err := blobs.Resolve(ctx, *res, writer)
		return err
	})

	log.V(3).Info("Upload component.", "ref", ref)
	if err := c.OciClient.PushManifest(ctx, ref, manifest, ociclient.WithStore(store)); err != nil {
		return err
	}

	return nil
}
