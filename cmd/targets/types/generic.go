// SPDX-FileCopyrightText: 2022 SAP SE or an SAP affiliate company and Gardener contributors.
//
// SPDX-License-Identifier: Apache-2.0

package types

import (
	"context"
	"fmt"
	"os"
	"strings"

	lsv1alpha1 "github.com/gardener/landscaper/apis/core/v1alpha1"
	"github.com/spf13/cobra"
	"github.com/spf13/pflag"
	"sigs.k8s.io/yaml"

	"github.com/gardener/landscapercli/pkg/util"
)

// TypedTargetCreator is a function which takes a context and the common target creation options
// and returns the content which should be put into the target, as well as the target's type.
type TypedTargetCreator func(context.Context, *TargetCreateOpts) ([]byte, lsv1alpha1.TargetType, error)

// TargetCreateOpts contains the options which all target creation subcommands have in common.
type TargetCreateOpts struct {
	// Name is the name of the target. It is required.
	Name string
	// Namespace is the namespace of the target.
	Namespace string
	// OutputPath defines where to write the generated target to.
	// If it is empty, the generated target will be printed to stdout,
	// otherwise it will be put into a file at the specified path.
	OutputPath string
	// SecretName defines the name of the secret which should be used to store the target's content.
	// If empty, the content will be put into the target's spec directly.
	SecretName string
}

func (o *TargetCreateOpts) AddFlags(fs *pflag.FlagSet) {
	fs.StringVar(&o.Name, "name", "", "name of the target (required)")
	fs.StringVarP(&o.Namespace, "namespace", "n", "", "namespace of the target")
	fs.StringVarP(&o.OutputPath, "output-file", "o", "", "file path for the resulting target yaml, leave empty for stdout")
	fs.StringVarP(&o.SecretName, "secret", "s", "", "name of the secret to store the target's content in (content will be stored in target spec directly, if empty)")
}

// NewTargetCreateSubcommand wraps target-type-specific code with generic functionality for target creation.
// It overwrites the given command's Run field with a function that
//   - uses the given TypedTargetCreator function to generate the target's content
//   - generates the Target yaml and potentially secret yaml, depending on the options
//   - either prints the yaml file(s) to stdout or the given path, depending on the options
func NewTargetCreateSubcommand(ctx context.Context, cmd *cobra.Command, run TypedTargetCreator, opts *TargetCreateOpts) *cobra.Command {
	handleError := func(err error) {
		cmd.PrintErr(err)
		os.Exit(1)
	}

	cmd.Run = func(cmd *cobra.Command, args []string) {
		content, targetType, err := run(ctx, opts)
		if err != nil {
			cmd.PrintErr(err.Error())
			os.Exit(1)
		}

		target, secret := util.BuildTargetWithContent(opts.Name, opts.Namespace, targetType, content, opts.SecretName)

		res := strings.Builder{}

		if secret != nil {
			marshalledSecret, err := yaml.Marshal(secret)
			if err != nil {
				handleError(fmt.Errorf("cannot marshal secret yaml: %w", err))
			}
			res.WriteString(string(marshalledSecret))
			if !strings.HasSuffix(res.String(), "\n") {
				res.WriteString("\n")
			}
			res.WriteString("---\n")
		}

		marshalledTarget, err := yaml.Marshal(target)
		if err != nil {
			handleError(fmt.Errorf("cannot marshal target yaml: %w", err))
		}
		res.WriteString(string(marshalledTarget))

		if opts.OutputPath == "" {
			cmd.Println(res.String())
		} else {
			f, err := os.Create(opts.OutputPath)
			if err != nil {
				handleError(fmt.Errorf("error creating file %s: %w", opts.OutputPath, err))
			}
			_, err = f.WriteString(res.String())
			if err != nil {
				handleError(fmt.Errorf("error writing file %s: %w", opts.OutputPath, err))
			}
			cmd.Printf("Wrote target to %s", opts.OutputPath)
		}
	}

	return cmd
}
