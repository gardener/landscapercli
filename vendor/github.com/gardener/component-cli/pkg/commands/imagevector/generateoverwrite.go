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

	cdv2 "github.com/gardener/component-spec/bindings-go/apis/v2"
	"github.com/gardener/component-spec/bindings-go/codec"
	"github.com/ghodss/yaml"
	"github.com/go-logr/logr"
	"github.com/mandelsoft/vfs/pkg/osfs"
	"github.com/mandelsoft/vfs/pkg/vfs"
	"github.com/spf13/cobra"
	"github.com/spf13/pflag"

	ociopts "github.com/gardener/component-cli/ociclient/options"
	"github.com/gardener/component-cli/pkg/commands/constants"
	"github.com/gardener/component-cli/pkg/components"
	"github.com/gardener/component-cli/pkg/imagevector"
	"github.com/gardener/component-cli/pkg/logger"
	"github.com/gardener/component-cli/pkg/utils"
)

// GenerateOverwriteOptions defines the options that are used to generate a image vector from component descriptors
type GenerateOverwriteOptions struct {
	LocalComponentDescriptorOption
	RemoteComponentDescriptorOption

	// SubComponentName is the name of the reference that should be used as the main component descriptor.
	// +optional
	SubComponentName string
	// ImageVectorPath defines the path to the image vector defined as yaml or json
	ImageVectorPath string

	// OciOptions contains all exposed options to configure the oci client.
	OciOptions ociopts.Options
}

// ComponentDescriptorOption describes a component descriptor
type ComponentDescriptorOption struct {
	Local  LocalComponentDescriptorOption
	Remote RemoteComponentDescriptorOption
}

// RemoteComponentDescriptorOption defines component descriptors that are fetched by their remote repository, name and version.
type RemoteComponentDescriptorOption struct {
	// BaseURL defines the repository base url of the remote repository
	// +optional
	BaseURL string
	// ComponentName is the name of the remote component to fetch
	// +optional
	ComponentName string
	// ComponentVersion is the version of the remote component to fetch
	// +optional
	ComponentVersion string
}

// LocalComponentDescriptorOption defines component descriptors that are locally available
type LocalComponentDescriptorOption struct {
	// ComponentDescriptorPath is the path to the component descriptor
	// Either component descriptor or a remote component descriptor has to be defined.
	// +optional
	ComponentDescriptorPath string
	// ComponentDescriptorsPath is a list of paths to additional component descriptors
	ComponentDescriptorsPath []string
}

// NewGenerateOverwriteCommand creates a command to add additional resources to a component descriptor.
func NewGenerateOverwriteCommand(ctx context.Context) *cobra.Command {
	opts := &GenerateOverwriteOptions{}
	cmd := &cobra.Command{
		Use:     "generate-overwrite",
		Aliases: []string{"go"},
		Short:   "Get parses a component descriptor and returns the defined image vector",
		Long: `
generate-overwrite parses images defined in a component descriptor and returns them as image vector.

Images can be defined in a component descriptor in 3 different ways:
1. as 'ociImage' resource: The image is defined a default resource of type 'ociImage' with a access of type 'ociRegistry'.
   It is expected that the resource contains the following labels to be identified as image vector image.
   The resulting image overwrite will contain the repository and the tag/digest from the access method.
<pre>
resources:
- name: pause-container
  version: "3.1"
  type: ociImage
  relation: external
  extraIdentity:
    "imagevector-gardener-cloud+tag": "3.1"
  labels:
  - name: imagevector.gardener.cloud/name
    value: pause-container
  - name: imagevector.gardener.cloud/repository
    value: gcr.io/google_containers/pause-amd64
  - name: imagevector.gardener.cloud/source-repository
    value: github.com/kubernetes/kubernetes/blob/master/build/pause/Dockerfile
  - name: imagevector.gardener.cloud/target-version
    value: "< 1.16"
  access:
    type: ociRegistry
    imageReference: gcr.io/google_containers/pause-amd64:3.1
</pre>

2. as component reference: The images are defined in a label "imagevector.gardener.cloud/images".
   The resulting image overwrite will contain all images defined in the images label.
   Their repository and tag/digest will be matched from the resources defined in the actual component's resources.

   Note: The images from the label are matched to the resources using their name and version. The original image reference do not exit anymore.

<pre>
componentReferences:
- name: cluster-autoscaler-abc
  componentName: github.com/gardener/autoscaler
  version: v0.10.1
  labels:
  - name: imagevector.gardener.cloud/images
    value:
      images:
      - name: cluster-autoscaler
        repository: eu.gcr.io/gardener-project/gardener/autoscaler/cluster-autoscaler
        tag: "v0.10.1"
</pre>

3. as generic images from the component descriptor labels.
   Generic images are images that do not directly result in a resource.
   They will be matched with another component descriptor that actually defines the images.
   The other component descriptor MUST have the "imagevector.gardener.cloud/name" label in order to be matched.

<pre>
meta:
  schemaVersion: 'v2'
component:
  labels:
  - name: imagevector.gardener.cloud/images
    value:
      images:
      - name: hyperkube
        repository: k8s.gcr.io/hyperkube
        targetVersion: "< 1.19"
</pre>

<pre>
meta:
  schemaVersion: 'v2'
component:
  resources:
  - name: hyperkube
    version: "v1.19.4"
    type: ociImage
    extraIdentity:
      "imagevector-gardener-cloud+tag": "v1.19.4"
    labels:
    - name: imagevector.gardener.cloud/name
      value: hyperkube
    - name: imagevector.gardener.cloud/repository
      value: k8s.gcr.io/hyperkube
    access:
	  type: ociRegistry
	  imageReference: my-registry/hyperkube:v1.19.4
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
	utils.CleanMarkdownUsageFunc(cmd)
	return cmd
}

func (o *GenerateOverwriteOptions) Run(ctx context.Context, log logr.Logger, fs vfs.FileSystem) error {
	ociClient, _, err := o.OciOptions.Build(log, fs)
	if err != nil {
		return err
	}
	compResolver := components.New(log, fs, ociClient)

	root, main, err := o.getComponentDescriptor(ctx, fs, compResolver)
	if err != nil {
		return fmt.Errorf("unable to get component descriptor: %s", err.Error())
	}

	// parse all given additional component descriptors
	cdList, err := o.getComponentDescriptors(ctx, fs, compResolver, root)
	if err != nil {
		return fmt.Errorf("unable to get component descriptors: %s", err.Error())
	}

	// if the root component descriptor is not the main component descriptor add it to the list of component descriptos
	if root != main {
		cdList.Components = append(cdList.Components, *root)
	}

	imageVector, err := imagevector.GenerateImageOverwrite(main, cdList)
	if err != nil {
		return fmt.Errorf("unable to parse image vector: %s", err.Error())
	}

	data, err := yaml.Marshal(imageVector)
	if err != nil {
		return fmt.Errorf("unable to encode image vector: %w", err)
	}
	if len(o.ImageVectorPath) != 0 {
		if err := fs.MkdirAll(filepath.Dir(o.ImageVectorPath), os.ModePerm); err != nil {
			return fmt.Errorf("unable to create directories for %q: %s", o.ImageVectorPath, err.Error())
		}
		if err := vfs.WriteFile(fs, o.ImageVectorPath, data, 06444); err != nil {
			return fmt.Errorf("unable to write image vector: %w", err)
		}
		fmt.Printf("Successfully generated image vector from component descriptor")
	} else {
		fmt.Println(string(data))
	}
	return nil
}

func (o *GenerateOverwriteOptions) Complete(args []string) error {

	// default component path to env var
	if len(o.ComponentDescriptorPath) == 0 {
		o.ComponentDescriptorPath = filepath.Dir(os.Getenv(constants.ComponentDescriptorPathEnvName))
	}
	if len(o.BaseURL) == 0 {
		o.BaseURL = os.Getenv(constants.ComponentRepositoryRepositoryBaseUrlEnvName)
	}

	return o.validate()
}

func (o *GenerateOverwriteOptions) validate() error {
	if len(o.ComponentDescriptorPath) == 0 && len(o.ComponentName) == 0 {
		return errors.New("component descriptor path or a remote component descriptor must be provided")
	}

	if len(o.ComponentName) == 0 {
		if len(o.ComponentVersion) != 0 {
			return errors.New("a component version has to be defined for a upstream component")
		}
		if len(o.BaseURL) != 0 {
			return errors.New("a base url has to be defined for a upstream component")
		}
	}

	return nil
}

func (o *GenerateOverwriteOptions) AddFlags(fs *pflag.FlagSet) {
	fs.StringVar(&o.ComponentDescriptorPath, "comp", "", "path to the component descriptor directory")
	fs.StringArrayVar(&o.ComponentDescriptorsPath, "add-comp", []string{}, "path to the component descriptor directory")

	fs.StringVar(&o.BaseURL, "repo-ctx", "", "base url of the component repository")
	fs.StringVar(&o.ComponentName, "component-name", "", "name of the remote component")
	fs.StringVar(&o.ComponentVersion, "component-version", "", "version of the remote component")

	fs.StringVarP(&o.ImageVectorPath, "Output", "O", "", "The path to the image vector that will be written.")
	fs.StringVar(&o.SubComponentName, "sub-component", "", "name of the sub component that should be used as the main component descriptor")
	o.OciOptions.AddFlags(fs)
}

func (o *GenerateOverwriteOptions) getComponentDescriptor(ctx context.Context, fs vfs.FileSystem, compResolver components.ComponentResolver) (root, main *cdv2.ComponentDescriptor, err error) {
	var cd *cdv2.ComponentDescriptor
	if len(o.ComponentDescriptorPath) != 0 {
		data, err := vfs.ReadFile(fs, o.ComponentDescriptorPath)
		if err != nil {
			return nil, nil, fmt.Errorf("unable to read component descriptor from %q: %s", o.ComponentDescriptorPath, err.Error())
		}

		// add the input to the ctf format
		cd = &cdv2.ComponentDescriptor{}
		if err := codec.Decode(data, cd); err != nil {
			return nil, nil, fmt.Errorf("unable to decode component descriptor from %q: %s", o.ComponentDescriptorPath, err.Error())
		}
	} else {
		var err error
		cd, err = compResolver.Resolve(ctx, cdv2.RepositoryContext{
			Type:    cdv2.OCIRegistryType,
			BaseURL: o.BaseURL,
		}, o.ComponentName, o.ComponentVersion)
		if err != nil {
			return nil, nil, fmt.Errorf("unable to resolve upstream component descriptor %q:%q: %s", o.ComponentName, o.ComponentVersion, err.Error())
		}
	}

	// use the defined subcomponents as main component
	if len(o.SubComponentName) != 0 {

		ref, err := cd.GetComponentReferencesByName(o.SubComponentName)
		if err != nil {
			return nil, nil, fmt.Errorf("unable to resolve subcomponent from root component descriptor: %w", err)
		}
		if len(ref) != 1 {
			return nil, nil, fmt.Errorf("the sub component %q is not unique in the component descriptor", o.SubComponentName)
		}

		main, err := compResolver.Resolve(ctx, cdv2.RepositoryContext{
			Type:    cdv2.OCIRegistryType,
			BaseURL: o.BaseURL,
		}, ref[0].ComponentName, ref[0].Version)
		if err != nil {
			return nil, nil, fmt.Errorf("unable to resolve upstream subcomponent descriptor %q:%q: %s", ref[0].ComponentName, ref[0].Version, err.Error())
		}
		return cd, main, nil
	}

	return cd, cd, nil
}

func (o *GenerateOverwriteOptions) getComponentDescriptors(ctx context.Context, fs vfs.FileSystem, compResolver components.ComponentResolver, cd *cdv2.ComponentDescriptor) (*cdv2.ComponentDescriptorList, error) {
	if len(o.ComponentDescriptorsPath) != 0 {
		// parse all given additional component descriptors
		cdList := &cdv2.ComponentDescriptorList{}
		for _, cdPath := range o.ComponentDescriptorsPath {
			data, err := vfs.ReadFile(fs, cdPath)
			if err != nil {
				return nil, fmt.Errorf("unable to read component descriptor from %q: %s", cdPath, err.Error())
			}

			// add the input to the ctf format
			cd := cdv2.ComponentDescriptor{}
			if err := codec.Decode(data, &cd); err != nil {
				return nil, fmt.Errorf("unable to decode component descriptor from %q: %s", cdPath, err.Error())
			}
			cdList.Components = append(cdList.Components, cd)
		}
		return cdList, nil
	}

	return components.ResolveTransitiveComponentDescriptors(ctx, compResolver, cd)
}
