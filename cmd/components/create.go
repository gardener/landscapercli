// SPDX-FileCopyrightText: 2020 SAP SE or an SAP affiliate company and Gardener contributors.
//
// SPDX-License-Identifier: Apache-2.0

package components

import (
	"context"
	"fmt"
	"os"

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
		Use:  "create [component directory path] [component name] [component version]",
		Args: cobra.ExactArgs(3),
		Example: "landscaper-cli component create \\\n" +
			"    . \\\n" +
			"    github.com/gardener/landscapercli/nginx \\\n" +
			"    v0.1.0",
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

			fmt.Printf("Successfully created")
		},
	}

	opts.AddFlags(cmd.Flags())

	return cmd
}

func (o *createOptions) Complete(args []string) error {
	o.componentPath = args[0]
	o.componentName = args[1]
	o.componentVersion = args[2]
	return nil
}

func (o *createOptions) AddFlags(fs *pflag.FlagSet) {
}

func (o *createOptions) run(ctx context.Context, log logr.Logger) error {
	err := o.checkPreconditions()
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

func (o *createOptions) checkPreconditions() error {
	// Check that the component directory exists
	fileInfo, err := os.Stat(o.componentPath)
	if err != nil {
		if os.IsNotExist(err) {
			err = os.MkdirAll(o.componentPath, 0755)
			if err != nil {
				return err
			}

			fileInfo, err = os.Stat(o.componentPath)
			if err != nil {
				return err
			}

		} else {
			return err
		}
	}

	// Check that the path points to a directory
	if !fileInfo.IsDir() {
		return fmt.Errorf("Path is not a directory")
	}

	// Check that the component directory is empty
	empty, err := util.IsDirectoryEmpty(o.componentPath)
	if err != nil {
		return err
	}
	if !empty {
		return fmt.Errorf("Component directory is not empty")
	}

	return nil
}

func (o *createOptions) buildInitialComponentDescriptor() *cd.ComponentDescriptor {
	return &cd.ComponentDescriptor{
		Metadata: cd.Metadata{
			Version: cd.SchemaVersion,
		},
		ComponentSpec: cd.ComponentSpec{
			ObjectMeta: cd.ObjectMeta{
				Name:    o.componentName,
				Version: o.componentVersion,
			},
			RepositoryContexts: []cd.RepositoryContext{
				{
					Type:    cd.OCIRegistryType,
					BaseURL: "",
				},
			},
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
				MediaType:        v1alpha1.BlueprintArtifactsMediaType,
			},
		},
	}
}
