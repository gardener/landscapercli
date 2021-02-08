// SPDX-FileCopyrightText: 2020 SAP SE or an SAP affiliate company and Gardener contributors.
//
// SPDX-License-Identifier: Apache-2.0

package blueprints

import (
	"context"
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"

	"github.com/mandelsoft/vfs/pkg/osfs"
	"github.com/mandelsoft/vfs/pkg/projectionfs"
	"github.com/spf13/cobra"
	"k8s.io/apimachinery/pkg/runtime/serializer"

	"github.com/gardener/landscaper/apis/core"
	lsv1alpha1 "github.com/gardener/landscaper/apis/core/v1alpha1"
	"github.com/gardener/landscaper/apis/core/validation"
	"github.com/gardener/landscaper/pkg/kubernetes"
)

type validationOptions struct {
	// blueprintPath is the path to the directory containing the definition.
	blueprintPath string
}

// NewValidationCommand creates a new blueprint command to validate blueprints.
func NewValidationCommand(_ context.Context) *cobra.Command {
	opts := &validationOptions{}
	cmd := &cobra.Command{
		Use:     "validate [path to Blueprint directory]",
		Args:    cobra.ExactArgs(1),
		Example: "landscaper-cli blueprints validate path/to/blueprint/directory",
		Short:   "validates a local blueprint filesystem",
		Long: "The validate command validates a Blueprint in a local directory. " +
			"The blueprint directory must contain a file with name blueprint.yaml.",
		Run: func(cmd *cobra.Command, args []string) {
			if err := opts.Complete(args); err != nil {
				fmt.Println(err.Error())
				os.Exit(1)
			}

			if err := opts.run(); err != nil {
				fmt.Println(err.Error())
				os.Exit(1)
			}

			fmt.Printf("Blueprint validated without errors\n")
		},
	}

	return cmd
}

func (o *validationOptions) run() error {
	data, err := ioutil.ReadFile(filepath.Join(o.blueprintPath, lsv1alpha1.BlueprintFileName))
	if err != nil {
		return err
	}
	blueprint := &core.Blueprint{}
	if _, _, err := serializer.NewCodecFactory(kubernetes.LandscaperScheme).UniversalDecoder().Decode(data, nil, blueprint); err != nil {
		return err
	}

	blueprintFs, err := projectionfs.New(osfs.New(), o.blueprintPath)
	if err != nil {
		return fmt.Errorf("unable to construct blueprint filesystem: %w", err)
	}
	if errList := validation.ValidateBlueprint(blueprintFs, blueprint); len(errList) != 0 {
		return errList.ToAggregate()
	}

	return nil
}

func (o *validationOptions) Complete(args []string) error {
	o.blueprintPath = args[0]
	return nil
}
