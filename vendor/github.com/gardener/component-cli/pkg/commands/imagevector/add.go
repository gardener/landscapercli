// SPDX-FileCopyrightText: 2020 SAP SE or an SAP affiliate company and Gardener contributors.
//
// SPDX-License-Identifier: Apache-2.0

package imagevector

import (
	"context"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	cdv2 "github.com/gardener/component-spec/bindings-go/apis/v2"
	cdvalidation "github.com/gardener/component-spec/bindings-go/apis/v2/validation"
	"github.com/gardener/component-spec/bindings-go/codec"
	"github.com/ghodss/yaml"
	"github.com/go-logr/logr"
	"github.com/mandelsoft/vfs/pkg/osfs"
	"github.com/mandelsoft/vfs/pkg/vfs"
	"github.com/spf13/cobra"
	"github.com/spf13/pflag"

	"github.com/gardener/component-cli/pkg/commands/constants"
	"github.com/gardener/component-cli/pkg/imagevector"
	"github.com/gardener/component-cli/pkg/logger"
)

// AddOptions defines the options that are used to add resources defined by a image vector to a component descriptor
type AddOptions struct {
	// ComponentDescriptorPath is the path to the component descriptor
	ComponentDescriptorPath string
	// ImageVectorPath defines the path to the image vector defined as yaml or json
	ImageVectorPath string

	imagevector.ParseImageOptions
	// GenericDependencies is a comma separated list of generic dependency names.
	// The list will be merged with the parse image options names.
	GenericDependencies string
}

// NewAddCommand creates a command to add additional resources to a component descriptor.
func NewAddCommand(ctx context.Context) *cobra.Command {
	opts := &AddOptions{}
	cmd := &cobra.Command{
		Use:   "add --comp-desc component-descriptor-file --image-vector images.yaml [--component-prefixes \"github.com/gardener/myproj\"]... [--generic-dependency image-source-name]... [--generic-dependencies \"image-name1,image-name2\"]",
		Short: "Adds all resources of a image vector to the component descriptor",
		Long: `
add parses a image vector and generates the corresponding component descriptor resources.

<pre>

images:
- name: pause-container
  sourceRepository: github.com/kubernetes/kubernetes/blob/master/build/pause/Dockerfile
  repository: gcr.io/google_containers/pause-amd64
  tag: "3.1"

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

func (o *AddOptions) Run(ctx context.Context, log logr.Logger, fs vfs.FileSystem) error {
	data, err := vfs.ReadFile(fs, o.ComponentDescriptorPath)
	if err != nil {
		return fmt.Errorf("unable to read component descriptor from %q: %s", o.ComponentDescriptorPath, err.Error())
	}

	// add the input to the ctf format
	cd := &cdv2.ComponentDescriptor{}
	if err := codec.Decode(data, cd); err != nil {
		return fmt.Errorf("unable to decode component descriptor from %q: %s", o.ComponentDescriptorPath, err.Error())
	}

	if err := o.parseImageVector(cd, fs); err != nil {
		return err
	}

	if err := cdvalidation.Validate(cd); err != nil {
		return fmt.Errorf("invalid component descriptor: %w", err)
	}

	data, err = yaml.Marshal(cd)
	if err != nil {
		return fmt.Errorf("unable to encode component descriptor: %w", err)
	}
	if err := vfs.WriteFile(fs, o.ComponentDescriptorPath, data, 0664); err != nil {
		return fmt.Errorf("unable to write modified comonent descriptor: %w", err)
	}
	log.V(2).Info("Successfully added all resources from the image vector to component descriptor")
	return nil
}

func (o *AddOptions) Complete(args []string) error {

	// default component path to env var
	if len(o.ComponentDescriptorPath) == 0 {
		o.ComponentDescriptorPath = filepath.Dir(os.Getenv(constants.ComponentDescriptorPathEnvName))
	}

	// parse generic dependencies
	if len(o.GenericDependencies) != 0 {
		for _, genericDepName := range strings.Split(o.GenericDependencies, ",") {
			o.ParseImageOptions.GenericDependencies = append(o.ParseImageOptions.GenericDependencies, strings.TrimSpace(genericDepName))
		}
	}

	return o.validate()
}

func (o *AddOptions) validate() error {
	if len(o.ComponentDescriptorPath) == 0 {
		return errors.New("component descriptor path must be provided")
	}
	if len(o.ImageVectorPath) == 0 {
		return errors.New("images path must be provided")
	}
	return nil
}

func (o *AddOptions) AddFlags(set *pflag.FlagSet) {
	set.StringVar(&o.ComponentDescriptorPath, "comp-desc", "", "path to the component descriptor directory")
	set.StringVar(&o.ImageVectorPath, "image-vector", "", "The path to the resources defined as yaml or json")
	set.StringArrayVar(&o.ParseImageOptions.ComponentReferencePrefixes, "component-prefixes", []string{}, "Specify all prefixes that define a image  from another component")
	set.StringArrayVar(&o.ParseImageOptions.GenericDependencies, "generic-dependency", []string{}, "Specify all image source names that are a generic dependency.")
	set.StringVar(&o.GenericDependencies, "generic-dependencies", "", "Specify all prefixes that define a image  from another component")
}

// parseImageVector parses the given image vector and returns a list of all resources.
func (o *AddOptions) parseImageVector(cd *cdv2.ComponentDescriptor, fs vfs.FileSystem) error {
	file, err := fs.Open(o.ImageVectorPath)
	if err != nil {
		return fmt.Errorf("unable to open image vector file: %q: %w", o.ImageVectorPath, err)
	}
	defer file.Close()
	return imagevector.ParseImageVector(cd, file, &o.ParseImageOptions)
}
