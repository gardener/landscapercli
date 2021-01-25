// SPDX-FileCopyrightText: 2020 SAP SE or an SAP affiliate company and Gardener contributors.
//
// SPDX-License-Identifier: Apache-2.0

package blueprints

import (
	"context"
	"fmt"
	"os"
	"path/filepath"

	"github.com/gardener/landscapercli/pkg/blueprints"
	"github.com/gardener/landscapercli/pkg/logger"

	"github.com/gardener/landscaper/pkg/apis/core/v1alpha1"
	"github.com/go-logr/logr"
	"github.com/spf13/cobra"
	"github.com/spf13/pflag"
)

type createOptions struct {
	// blueprintPath is the path to the directory containing the definition.
	blueprintPath string
}

// NewCreateCommand creates a new blueprint command to create a blueprint
func NewCreateCommand(ctx context.Context) *cobra.Command {
	opts := &createOptions{}
	cmd := &cobra.Command{
		Use:     "create [path to Blueprint directory]",
		Args:    cobra.ExactArgs(1),
		Example: "landscaper-cli blueprints create path/to/blueprint/directory",
		Short:   "command to create a blueprint template in the specified directory",
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
	o.blueprintPath = args[0]
	return nil
}

func (o *createOptions) AddFlags(fs *pflag.FlagSet) {
}

func (o *createOptions) run(ctx context.Context, log logr.Logger) error {
	exists, err := o.existsBlueprint()
	if err != nil {
		return err
	}
	if exists {
		return fmt.Errorf("The blueprint already exists")
	}

	blueprint := &v1alpha1.Blueprint{}

	err = blueprints.NewBlueprintWriter(o.blueprintPath).Write(blueprint)
	if err != nil {
		return err
	}

	return nil
}

func (o *createOptions) existsBlueprint() (bool, error) {
	_, err := os.Stat(o.getBlueprintFilePath())
	if err != nil {
		if os.IsNotExist(err) {
			return false, nil
		}
		return false, err
	}

	return true, nil
}

func (o *createOptions) getBlueprintFilePath() string {
	return filepath.Join(o.blueprintPath, v1alpha1.BlueprintFileName)
}
