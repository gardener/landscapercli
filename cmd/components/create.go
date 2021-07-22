// SPDX-FileCopyrightText: 2020 SAP SE or an SAP affiliate company and Gardener contributors.
//
// SPDX-License-Identifier: Apache-2.0

package components

import (
	"context"
	"fmt"
	"os"
	"regexp"

	"github.com/Masterminds/semver/v3"
	"github.com/gardener/landscaper/apis/mediatype"

	"github.com/gardener/component-cli/pkg/commands/componentarchive/input"
	"github.com/gardener/landscaper/apis/core/v1alpha1"

	cdresources "github.com/gardener/component-cli/pkg/commands/componentarchive/resources"
	cd "github.com/gardener/component-spec/bindings-go/apis/v2"

	"github.com/gardener/landscapercli/pkg/blueprints"
	"github.com/gardener/landscapercli/pkg/components"
	"github.com/gardener/landscapercli/pkg/logger"
	"github.com/gardener/landscapercli/pkg/util"

	"github.com/go-logr/logr"
	"github.com/spf13/cobra"
	"github.com/spf13/pflag"
)

var componentNameValidationRegexp = regexp.MustCompile("^[a-z0-9.\\-]+[.][a-z]{2,4}/[-a-z0-9/_.]*$") // nolint

type createOptions struct {
	// componentPath is the path to the directory containing the componentDescriptor.yaml
	// and blueprint directory.
	componentPath    string
	componentName    string
	componentVersion string
}

// NewCreateCommand creates a new blueprint command to create a blueprint
func NewCreateCommand(ctx context.Context) *cobra.Command {
	opts := &createOptions{}
	cmd := &cobra.Command{
		Use:  "create [component name] [component semver version]",
		Args: cobra.ExactArgs(2),
		Example: "landscaper-cli component create \\\n" +
			"    github.com/gardener/landscapercli/nginx \\\n" +
			"    v0.1.0 \\\n" +
			"    --component-directory ~/myComponent",
		Short: "command to create a component template in the specified directory",
		Run: func(cmd *cobra.Command, args []string) {
			if err := opts.Complete(args); err != nil {
				fmt.Println(err.Error())
				os.Exit(1)
			}

			if err := opts.run(ctx, logger.Log); err != nil {
				fmt.Println(err.Error())
				os.Exit(1)
			}

			fmt.Printf("Component with blueprint created")
			fmt.Printf("  \n- blueprint folder with blueprint yaml created")
			fmt.Printf("  \n- component descriptor yaml created")
			fmt.Printf("  \n- resources yaml created")
		},
	}

	opts.AddFlags(cmd.Flags())

	return cmd
}

func (o *createOptions) Complete(args []string) error {
	o.componentName = args[0]
	o.componentVersion = args[1]

	return o.validate()
}

func (o *createOptions) validate() error {
	if !componentNameValidationRegexp.Match([]byte(o.componentName)) {
		return fmt.Errorf("the component name does not match pattern '^[a-z0-9.\\-]+[.][a-z]{2,4}/[-a-z0-9/_.]*$'")
	}

	_, err := semver.NewVersion(o.componentVersion)
	if err != nil {
		return fmt.Errorf("component version %s is not semver compatible", o.componentVersion)
	}

	fileInfo, err := os.Stat(o.componentPath)
	if err != nil {
		if os.IsNotExist(err) {
			return nil
		}
		return err
	}

	// Check that the path points to a directory
	if !fileInfo.IsDir() {
		return fmt.Errorf("component-directory is not a directory")
	}

	_, err = os.Stat(util.BlueprintDirectoryPath(o.componentPath))
	if err != nil {
		if !os.IsNotExist(err) {
			return err
		}
	} else {
		return fmt.Errorf("blueprint file or folder already exists in %s ", util.BlueprintDirectoryPath(o.componentPath))
	}

	_, err = os.Stat(util.ResourcesFilePath(o.componentPath))
	if err != nil {
		if !os.IsNotExist(err) {
			return err
		}
	} else {
		return fmt.Errorf("resources.yaml file already exists in %s ", util.ResourcesFilePath(o.componentPath))
	}

	_, err = os.Stat(util.ComponentDescriptorFilePath(o.componentPath))
	if err != nil {
		if !os.IsNotExist(err) {
			return err
		}
	} else {
		return fmt.Errorf("resources.yaml file already exists in %s ", util.ResourcesFilePath(o.componentPath))
	}

	return nil
}

func (o *createOptions) AddFlags(fs *pflag.FlagSet) {
	fs.StringVar(&o.componentPath,
		"component-directory",
		".",
		"path to component directory")
}

func (o *createOptions) run(ctx context.Context, log logr.Logger) error {
	err := o.createComponentDir()
	if err != nil {
		return err
	}

	// Create blueprint directory
	blueprintDirectoryPath := util.BlueprintDirectoryPath(o.componentPath)
	err = os.Mkdir(blueprintDirectoryPath, 0755)
	if err != nil {
		return err
	}

	// Create blueprint file
	blueprint := &v1alpha1.Blueprint{}
	err = blueprints.NewBlueprintWriter(blueprintDirectoryPath).Write(blueprint)
	if err != nil {
		return err
	}

	// Create component-descriptor file
	componentDescriptor := o.buildInitialComponentDescriptor()
	err = components.NewComponentDescriptorWriter(o.componentPath).Write(componentDescriptor)
	if err != nil {
		return err
	}

	// Create resources.yaml
	resourceOptions := o.buildInitialResources()
	err = components.NewResourceWriter(o.componentPath).Write(resourceOptions)
	if err != nil {
		return err
	}

	return nil
}

func (o *createOptions) createComponentDir() error {
	_, err := os.Stat(o.componentPath)
	if err != nil {
		if os.IsNotExist(err) {
			err = os.MkdirAll(o.componentPath, 0755)
			if err != nil {
				return err
			}

			_, err = os.Stat(o.componentPath)
			if err != nil {
				return err
			}

		} else {
			return err
		}
	}

	return nil
}

func (o *createOptions) buildInitialComponentDescriptor() *cd.ComponentDescriptor {
	repoCtx, _ := cd.NewUnstructured(cd.NewOCIRegistryRepository("", ""))
	return &cd.ComponentDescriptor{
		Metadata: cd.Metadata{
			Version: cd.SchemaVersion,
		},
		ComponentSpec: cd.ComponentSpec{
			ObjectMeta: cd.ObjectMeta{
				Name:    o.componentName,
				Version: o.componentVersion,
			},
			RepositoryContexts:  []*cd.UnstructuredTypedObject{&repoCtx},
			Provider:            cd.InternalProvider,
			Sources:             []cd.Source{},
			ComponentReferences: []cd.ComponentReference{},
			Resources:           []cd.Resource{},
		},
	}
}

func (o *createOptions) buildInitialResources() []cdresources.ResourceOptions {
	compress := true

	return []cdresources.ResourceOptions{
		{
			Resource: cd.Resource{
				IdentityObjectMeta: cd.IdentityObjectMeta{
					Name:    "blueprint",
					Version: o.componentVersion,
					Type:    v1alpha1.BlueprintResourceType,
				},
				Relation: "local",
			},
			Input: &input.BlobInput{
				Type:             input.DirInputType,
				Path:             "./blueprint",
				CompressWithGzip: &compress,
				MediaType: mediatype.NewBuilder(mediatype.BlueprintArtifactsLayerMediaTypeV1).
					Compression(mediatype.GZipCompression).
					String(),
			},
		},
	}
}
