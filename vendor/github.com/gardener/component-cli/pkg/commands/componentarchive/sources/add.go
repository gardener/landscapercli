// SPDX-FileCopyrightText: 2020 SAP SE or an SAP affiliate company and Gardener contributors.
//
// SPDX-License-Identifier: Apache-2.0

package sources

import (
	"context"
	"fmt"
	"io"
	"os"
	"path/filepath"

	cdv2 "github.com/gardener/component-spec/bindings-go/apis/v2"
	"github.com/gardener/component-spec/bindings-go/apis/v2/cdutils"
	cdvalidation "github.com/gardener/component-spec/bindings-go/apis/v2/validation"
	"github.com/gardener/component-spec/bindings-go/ctf"
	"github.com/ghodss/yaml"
	"github.com/go-logr/logr"
	"github.com/mandelsoft/vfs/pkg/osfs"
	"github.com/mandelsoft/vfs/pkg/vfs"
	"github.com/spf13/cobra"
	"github.com/spf13/pflag"
	"k8s.io/apimachinery/pkg/util/validation/field"
	yamlutil "k8s.io/apimachinery/pkg/util/yaml"

	"github.com/gardener/component-cli/pkg/commands/componentarchive/input"
	"github.com/gardener/component-cli/pkg/componentarchive"
	"github.com/gardener/component-cli/pkg/logger"
)

// Options defines the options that are used to add resources to a component descriptor
type Options struct {
	componentarchive.BuilderOptions

	// either components can be added by a yaml resource template or by input flags

	// SourceObjectPath defines the path to the resources defined as yaml or json
	SourceObjectPath string
}

// SourceOptions contains options that are used to describe a source
type SourceOptions struct {
	cdv2.Source
	Input *input.BlobInput `json:"input,omitempty"`
}

// NewAddCommand creates a command to add additional resources to a component descriptor.
func NewAddCommand(ctx context.Context) *cobra.Command {
	opts := &Options{}
	cmd := &cobra.Command{
		Use:   "add [component descriptor path]",
		Args:  cobra.MaximumNArgs(1),
		Short: "Adds a source to a component descriptor",
		Long: `
add adds sources to the defined component descriptor.
The sources can be defined in a file or given through stdin.

The source definitions are expected to be a multidoc yaml of the following form

<pre>

---
name: 'myrepo'
type: 'git'
access:
  type: "git"
  repository: github.com/gardener/component-cli
...
---
name: 'myconfig'
type: 'json'
input:
  type: "file"
  path: "some/path"
...
---
name: 'myothersrc'
type: 'json'
input:
  type: "dir"
  path: /my/path
  compress: true # defaults to false
  exclude: "*.txt"
...

</pre>
`,
		Run: func(cmd *cobra.Command, args []string) {
			if err := opts.Complete(args); err != nil {
				fmt.Println(err.Error())
				os.Exit(1)
			}

			if err := opts.Run(ctx, logger.Log, osfs.New()); err != nil {
				fmt.Println(err.Error())
				os.Exit(1)
			}
		},
	}

	opts.AddFlags(cmd.Flags())

	return cmd
}

func (o *Options) Run(ctx context.Context, log logr.Logger, fs vfs.FileSystem) error {
	compDescFilePath := filepath.Join(o.ComponentArchivePath, ctf.ComponentDescriptorFileName)

	archive, err := o.BuilderOptions.Build(fs)
	if err != nil {
		return err
	}

	sources, err := o.generateSources(fs)
	if err != nil {
		return err
	}

	for _, src := range sources {
		if src.Input != nil {
			log.Info(fmt.Sprintf("add input blob from %q", src.Input.Path))
			if err := o.addInputBlob(fs, archive, src); err != nil {
				return err
			}
		} else {
			id := archive.ComponentDescriptor.GetSourceIndex(src.Source)
			if id != -1 {
				mergedSrc := cdutils.MergeSources(archive.ComponentDescriptor.Sources[id], src.Source)
				if errList := cdvalidation.ValidateSource(field.NewPath(""), mergedSrc); len(errList) != 0 {
					return fmt.Errorf("invalid component reference: %w", errList.ToAggregate())
				}
				archive.ComponentDescriptor.Sources[id] = mergedSrc
			} else {
				if errList := cdvalidation.ValidateSource(field.NewPath(""), src.Source); len(errList) != 0 {
					return fmt.Errorf("invalid component reference: %w", errList.ToAggregate())
				}
				archive.ComponentDescriptor.Sources = append(archive.ComponentDescriptor.Sources, src.Source)
			}
		}
		log.V(3).Info(fmt.Sprintf("Successfully added source %q to component descriptor", src.Name))
	}

	if err := cdvalidation.Validate(archive.ComponentDescriptor); err != nil {
		return fmt.Errorf("invalid component descriptor: %w", err)
	}

	data, err := yaml.Marshal(archive.ComponentDescriptor)
	if err != nil {
		return fmt.Errorf("unable to encode component descriptor: %w", err)
	}
	if err := vfs.WriteFile(fs, compDescFilePath, data, 0664); err != nil {
		return fmt.Errorf("unable to write modified comonent descriptor: %w", err)
	}
	log.V(1).Info("Successfully added all sources to component descriptor")
	return nil
}

func (o *Options) Complete(args []string) error {
	if len(args) != 0 {
		o.BuilderOptions.ComponentArchivePath = args[0]
	}
	o.BuilderOptions.Default()
	return o.validate()
}

func (o *Options) validate() error {
	return o.BuilderOptions.Validate()
}

func (o *Options) AddFlags(fs *pflag.FlagSet) {
	o.BuilderOptions.AddFlags(fs)
	// specify the resource
	fs.StringVarP(&o.SourceObjectPath, "resource", "r", "", "The path to the resources defined as yaml or json")
}

// generateSources parses component references from the given path and stdin.
func (o *Options) generateSources(fs vfs.FileSystem) ([]SourceOptions, error) {
	sources := make([]SourceOptions, 0)
	if len(o.SourceObjectPath) != 0 {
		resourceObjectReader, err := fs.Open(o.SourceObjectPath)
		if err != nil {
			return nil, fmt.Errorf("unable to read resource object from %s: %w", o.SourceObjectPath, err)
		}
		defer resourceObjectReader.Close()
		sources, err = generateSourcesFromReader(resourceObjectReader)
		if err != nil {
			return nil, fmt.Errorf("unable to read sources from %s: %w", o.SourceObjectPath, err)
		}
	}

	stdinInfo, err := os.Stdin.Stat()
	if err != nil {
		return nil, fmt.Errorf("unable to read from stdin: %w", err)
	}
	if (stdinInfo.Mode()&os.ModeNamedPipe != 0) || stdinInfo.Size() != 0 {
		stdinSources, err := generateSourcesFromReader(os.Stdin)
		if err != nil {
			return nil, fmt.Errorf("unable to read from stdin: %w", err)
		}
		sources = append(sources, stdinSources...)
	}
	return sources, nil
}

// generateSourcesFromReader generates a resource given resource options and a resource template file.
func generateSourcesFromReader(reader io.Reader) ([]SourceOptions, error) {
	sources := make([]SourceOptions, 0)
	yamldecoder := yamlutil.NewYAMLOrJSONDecoder(reader, 1024)
	for {
		src := SourceOptions{}
		if err := yamldecoder.Decode(&src); err != nil {
			if err == io.EOF {
				break
			}
			return nil, fmt.Errorf("unable to decode src: %w", err)
		}
		sources = append(sources, src)
	}

	return sources, nil
}

func (o *Options) addInputBlob(fs vfs.FileSystem, archive *ctf.ComponentArchive, src SourceOptions) error {
	blob, err := src.Input.Read(fs, o.SourceObjectPath)
	if err != nil {
		return err
	}

	err = archive.AddSource(&src.Source, ctf.BlobInfo{
		MediaType: src.Type,
		Digest:    blob.Digest,
		Size:      blob.Size,
	}, blob.Reader)
	if err != nil {
		blob.Reader.Close()
		return fmt.Errorf("unable to add input blob to archive: %w", err)
	}
	if err := blob.Reader.Close(); err != nil {
		return fmt.Errorf("unable to close input file: %w", err)
	}
	return nil
}
