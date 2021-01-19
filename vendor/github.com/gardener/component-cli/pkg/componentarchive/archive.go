// SPDX-FileCopyrightText: 2020 SAP SE or an SAP affiliate company and Gardener contributors.
//
// SPDX-License-Identifier: Apache-2.0

package componentarchive

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"

	cdv2 "github.com/gardener/component-spec/bindings-go/apis/v2"
	"github.com/gardener/component-spec/bindings-go/ctf"
	"github.com/mandelsoft/vfs/pkg/projectionfs"
	"github.com/mandelsoft/vfs/pkg/vfs"
	"github.com/spf13/pflag"

	"github.com/gardener/component-cli/pkg/commands/constants"
)

type BuilderOptions struct {
	ComponentArchivePath string

	Name    string
	Version string
	BaseUrl string
}

func (o *BuilderOptions) AddFlags(fs *pflag.FlagSet) {
	fs.StringVarP(&o.ComponentArchivePath, "archive", "a", "", "path to the component archive directory")
	fs.StringVar(&o.Name, "component-name", "", "name of the component")
	fs.StringVar(&o.Version, "component-version", "", "version of the component")
	fs.StringVar(&o.BaseUrl, "repo-ctx", "", "repository context url for component to upload. The repository url will be automatically added to the repository contexts.")
}

// Default applies defaults to the builder options
func (o *BuilderOptions) Default() {
	// default component path to env var
	if len(o.ComponentArchivePath) == 0 {
		o.ComponentArchivePath = filepath.Dir(os.Getenv(constants.ComponentArchivePathEnvName))
	}
}

// Validate validates the component archive builder options.
func (o *BuilderOptions) Validate() error {
	if len(o.ComponentArchivePath) == 0 {
		return errors.New("a component archive path must be defined")
	}

	if len(o.Name) != 0 {
		if len(o.Version) == 0 {
			return errors.New("a version has to be provided for a minimal component descriptor")
		}
		if len(o.BaseUrl) == 0 {
			return errors.New("a repository context base url has to be provided for a minimal component descriptor")
		}
	}
	return nil
}

// Build creates a component archives with the given configuration
func (o *BuilderOptions) Build(fs vfs.FileSystem) (*ctf.ComponentArchive, error) {
	o.Default()
	if err := o.Validate(); err != nil {
		return nil, err
	}

	compDescFilePath := filepath.Join(o.ComponentArchivePath, ctf.ComponentDescriptorFileName)
	_, err := fs.Stat(compDescFilePath)
	if err != nil && !os.IsNotExist(err) {
		return nil, err
	}

	if err == nil {
		// add the input to the ctf format
		archiveFs, err := projectionfs.New(fs, o.ComponentArchivePath)
		if err != nil {
			return nil, fmt.Errorf("unable to create projectionfilesystem: %w", err)
		}
		archive, err := ctf.NewComponentArchiveFromFilesystem(archiveFs)
		if err != nil {
			return nil, fmt.Errorf("unable to parse component archive from %s: %w", o.ComponentArchivePath, err)
		}
		return archive, nil
	}

	// build minimal archive

	if err := fs.MkdirAll(o.ComponentArchivePath, os.ModePerm); err != nil {
		return nil, fmt.Errorf("unable to create component-archive path %q: %w", o.ComponentArchivePath, err)
	}
	archiveFs, err := projectionfs.New(fs, o.ComponentArchivePath)
	if err != nil {
		return nil, fmt.Errorf("unable to create projectionfilesystem: %w", err)
	}

	cd := &cdv2.ComponentDescriptor{}
	cd.Metadata.Version = cdv2.SchemaVersion
	cd.ComponentSpec.Name = o.Name
	cd.ComponentSpec.Version = o.Version
	cd.Provider = cdv2.InternalProvider
	cd.RepositoryContexts = []cdv2.RepositoryContext{
		{
			Type:    cdv2.OCIRegistryType,
			BaseURL: o.BaseUrl,
		},
	}
	if err := cdv2.DefaultComponent(cd); err != nil {
		return nil, fmt.Errorf("unable to default component descriptor: %w", err)
	}

	return ctf.NewComponentArchive(cd, archiveFs), nil
}
