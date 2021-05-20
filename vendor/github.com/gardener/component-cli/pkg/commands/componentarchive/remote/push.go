// SPDX-FileCopyrightText: 2020 SAP SE or an SAP affiliate company and Gardener contributors.
//
// SPDX-License-Identifier: Apache-2.0

package remote

import (
	"context"
	"errors"
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"

	cdv2 "github.com/gardener/component-spec/bindings-go/apis/v2"
	"github.com/gardener/component-spec/bindings-go/codec"
	"github.com/gardener/component-spec/bindings-go/ctf"
	cdoci "github.com/gardener/component-spec/bindings-go/oci"
	"github.com/go-logr/logr"
	"github.com/mandelsoft/vfs/pkg/osfs"
	"github.com/mandelsoft/vfs/pkg/vfs"
	"github.com/spf13/cobra"
	"github.com/spf13/pflag"

	ociopts "github.com/gardener/component-cli/ociclient/options"
	"github.com/gardener/component-cli/pkg/logger"
	"github.com/gardener/component-cli/pkg/utils"
)

// PushOptions contains all options to upload a component archive.
type PushOptions struct {
	// BaseUrl is the oci registry where the component is stored.
	BaseUrl string
	// ComponentName is the unique name of the component in the registry.
	ComponentName string
	// Version is the component Version in the oci registry.
	Version string
	// ComponentPath is the path to the directory containing the definition.
	ComponentPath string
	// AdditionalTags defines additional tags that the oci artifact should be tagged with.
	AdditionalTags []string

	// cd is the effective component descriptor
	cd *cdv2.ComponentDescriptor

	// OciOptions contains all exposed options to configure the oci client.
	OciOptions ociopts.Options
}

// NewPushCommand creates a new definition command to push definitions
func NewPushCommand(ctx context.Context) *cobra.Command {
	opts := &PushOptions{}
	cmd := &cobra.Command{
		Use:   "push COMPONENT_DESCRIPTOR_PATH",
		Args:  cobra.RangeArgs(1, 4),
		Short: "pushes a component archive to an oci repository",
		Long: `
pushes a component archive with the component descriptor and its local blobs to an oci repository.

The command can be called in 2 different ways:

push [path to component descriptor]
- The cli will read all necessary parameters from the component descriptor.

push [baseurl] [componentname] [Version] [path to component descriptor]
- The cli will add the baseurl as repository context and validate the name and Version.
`,
		Run: func(cmd *cobra.Command, args []string) {
			if err := opts.Complete(args); err != nil {
				fmt.Println(err.Error())
				os.Exit(1)
			}

			if err := opts.run(ctx, logger.Log, osfs.New()); err != nil {
				fmt.Println(err.Error())
				os.Exit(1)
			}
		},
	}

	opts.AddFlags(cmd.Flags())

	return cmd
}

func (o *PushOptions) run(ctx context.Context, log logr.Logger, fs vfs.FileSystem) error {
	ociClient, cache, err := o.OciOptions.Build(log, fs)
	if err != nil {
		return fmt.Errorf("unable to build oci client: %s", err.Error())
	}

	archive, err := ctf.ComponentArchiveFromPath(o.ComponentPath)
	if err != nil {
		return fmt.Errorf("unable to build component archive: %w", err)
	}
	// update repository context
	archive.ComponentDescriptor.RepositoryContexts = utils.AddRepositoryContext(archive.ComponentDescriptor.RepositoryContexts, cdv2.OCIRegistryType, o.BaseUrl)

	manifest, err := cdoci.NewManifestBuilder(cache, archive).Build(ctx)
	if err != nil {
		return fmt.Errorf("unable to build oci artifact for component acrchive: %w", err)
	}

	repoCtx := o.cd.GetEffectiveRepositoryContext()
	ref, err := cdoci.OCIRef(repoCtx, o.cd.Name, o.cd.Version)
	if err != nil {
		return fmt.Errorf("invalid component reference: %w", err)
	}
	if err := ociClient.PushManifest(ctx, ref, manifest); err != nil {
		return err
	}
	log.Info(fmt.Sprintf("Successfully uploaded component descriptor at %q", ref))

	for _, tag := range o.AdditionalTags {
		ref, err := cdoci.OCIRef(repoCtx, o.cd.Name, tag)
		if err != nil {
			return fmt.Errorf("invalid component reference: %w", err)
		}
		if err := ociClient.PushManifest(ctx, ref, manifest); err != nil {
			return err
		}
		log.Info(fmt.Sprintf("Successfully tagged component descriptor %q", ref))
	}
	return nil
}

func (o *PushOptions) Complete(args []string) error {
	switch len(args) {
	case 1:
		o.ComponentPath = args[0]
	case 4:
		o.BaseUrl = args[0]
		o.ComponentName = args[1]
		o.Version = args[2]
		o.ComponentPath = args[3]
	}

	var err error
	o.OciOptions.CacheDir, err = utils.CacheDir()
	if err != nil {
		return fmt.Errorf("unable to get oci cache directory: %w", err)
	}

	if err := o.Validate(); err != nil {
		return err
	}

	info, err := os.Stat(o.ComponentPath)
	if err != nil {
		return fmt.Errorf("unable to get info for %s: %w", o.ComponentPath, err)
	}
	if !info.IsDir() {
		return fmt.Errorf(`%s is not a directory. 
It is expected that the given path points to a diectory that contains the component descriptor as file in "%s" 
`, o.ComponentPath, ctf.ComponentDescriptorFileName)
	}

	data, err := ioutil.ReadFile(filepath.Join(o.ComponentPath, ctf.ComponentDescriptorFileName))
	if err != nil {
		return err
	}
	o.cd = &cdv2.ComponentDescriptor{}
	if err := codec.Decode(data, o.cd); err != nil {
		return err
	}

	if len(o.ComponentName) != 0 {
		if o.cd.Name != o.ComponentName {
			return fmt.Errorf("name in component descriptor '%s' does not match the given name '%s'", o.cd.Name, o.ComponentName)
		}
		if o.cd.Version != o.Version {
			return fmt.Errorf("Version in component descriptor '%s' does not match the given Version '%s'", o.cd.Version, o.Version)
		}
	}

	if len(o.BaseUrl) != 0 {
		o.cd.RepositoryContexts = append(o.cd.RepositoryContexts, cdv2.RepositoryContext{
			Type:    cdv2.OCIRegistryType,
			BaseURL: o.BaseUrl,
		})
	}

	return nil
}

// Validate validates push options
func (o *PushOptions) Validate() error {
	if len(o.ComponentPath) == 0 {
		return errors.New("a path to the component descriptor must be defined")
	}

	// todo: validate references exist
	return nil
}

func (o *PushOptions) AddFlags(fs *pflag.FlagSet) {
	fs.StringVar(&o.BaseUrl, "repo-ctx", "", "repository context url for component to upload. The repository url will be automatically added to the repository contexts.")
	fs.StringArrayVarP(&o.AdditionalTags, "tag", "t", []string{}, "set additional tags on the oci artifact")
	o.OciOptions.AddFlags(fs)
}
