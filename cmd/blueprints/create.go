// SPDX-FileCopyrightText: 2020 SAP SE or an SAP affiliate company and Gardener contributors.
//
// SPDX-License-Identifier: Apache-2.0

package blueprints

import (
	"context"
	"fmt"
	"os"
	"path/filepath"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	"github.com/go-logr/logr"
	"github.com/spf13/cobra"
	"github.com/spf13/pflag"
	"sigs.k8s.io/yaml"

	"github.com/gardener/landscapercli/pkg/logger"
)

const blueprintFilename = "blueprint.yaml"

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

func (o *createOptions) run(ctx context.Context, log logr.Logger) error {
	typeMeta := metav1.TypeMeta{
		APIVersion: "landscaper.gardener.cloud/v1alpha1",
		Kind:       "Blueprint",
	}

	blueprintFilePath := filepath.Join(o.blueprintPath, blueprintFilename)
	f, err := os.Create(blueprintFilePath)
	if err != nil {
		return err
	}

	defer f.Close()

	typeMetaBytes, err := yaml.Marshal(typeMeta)
	if err != nil {
		return err
	}

	_, err = f.Write(typeMetaBytes)
	if err != nil {
		return err
	}

	_, err = f.WriteString("\nimports: []")
	if err != nil {
		return err
	}

	_, err = f.WriteString("\n\nexports: []")
	if err != nil {
		return err
	}

	_, err = f.WriteString("\n\nexportExecutions: []")
	if err != nil {
		return err
	}

	_, err = f.WriteString("\n\nsubinstallations: []")
	if err != nil {
		return err
	}

	_, err = f.WriteString("\n\ndeployExecutions: []")
	if err != nil {
		return err
	}

	return nil
}

func (o *createOptions) Complete(args []string) error {
	o.blueprintPath = args[0]
	return nil
}

func (o *createOptions) AddFlags(fs *pflag.FlagSet) {
}
