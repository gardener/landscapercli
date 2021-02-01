// SPDX-FileCopyrightText: 2020 SAP SE or an SAP affiliate company and Gardener contributors.
//
// SPDX-License-Identifier: Apache-2.0

package remote

import (
	"context"
	"errors"
	"fmt"
	"os"
	"path/filepath"

	cdv2 "github.com/gardener/component-spec/bindings-go/apis/v2"
	cdoci "github.com/gardener/component-spec/bindings-go/oci"
	"github.com/go-logr/logr"
	"github.com/mandelsoft/vfs/pkg/osfs"
	"github.com/mandelsoft/vfs/pkg/vfs"
	"github.com/spf13/cobra"
	"github.com/spf13/pflag"
	"sigs.k8s.io/yaml"

	ociopts "github.com/gardener/component-cli/ociclient/options"
	"github.com/gardener/component-cli/pkg/commands/constants"
	"github.com/gardener/component-cli/pkg/logger"
)

type showOptions struct {
	// baseUrl is the oci registry where the component is stored.
	baseUrl string
	// componentName is the unique name of the component in the registry.
	componentName string
	// version is the component version in the oci registry.
	version string

	// OciOptions contains all exposed options to configure the oci client.
	OciOptions ociopts.Options
}

// NewGetCommand shows definitions and their configuration.
func NewGetCommand(ctx context.Context) *cobra.Command {
	opts := &showOptions{}
	cmd := &cobra.Command{
		Use:   "get [baseurl] [componentname] [Version]",
		Args:  cobra.ExactArgs(3),
		Short: "fetch the component descriptor from a oci registry",
		Long: `
get fetches the component descriptor from a baseurl with the given name and Version.
`,
		Run: func(cmd *cobra.Command, args []string) {
			if err := opts.Complete(args); err != nil {
				fmt.Println(err.Error())
				os.Exit(1)
			}

			if err := opts.run(ctx, logger.Log, osfs.New()); err != nil {
				fmt.Println(err.Error())
				os.Exit(1)
			}
		},
	}

	opts.AddFlags(cmd.Flags())

	return cmd
}

func (o *showOptions) run(ctx context.Context, log logr.Logger, fs vfs.FileSystem) error {
	repoCtx := cdv2.RepositoryContext{
		Type:    cdv2.OCIRegistryType,
		BaseURL: o.baseUrl,
	}
	ociRef, err := cdoci.OCIRef(repoCtx, o.componentName, o.version)
	if err != nil {
		return fmt.Errorf("invalid component reference: %w", err)
	}

	ociClient, _, err := o.OciOptions.Build(log, fs)
	if err != nil {
		return fmt.Errorf("unable to build oci client: %s", err.Error())
	}

	cdresolver := cdoci.NewResolver().WithOCIClient(ociClient).WithRepositoryContext(repoCtx)
	cd, _, err := cdresolver.Resolve(ctx, o.componentName, o.version)
	if err != nil {
		return fmt.Errorf("unable to to fetch component descriptor %s: %w", ociRef, err)
	}

	out, err := yaml.Marshal(cd)
	if err != nil {
		return err
	}

	fmt.Println(string(out))
	return nil
}

func (o *showOptions) Complete(args []string) error {
	// todo: validate args
	o.baseUrl = args[0]
	o.componentName = args[1]
	o.version = args[2]

	cliHomeDir, err := constants.CliHomeDir()
	if err != nil {
		return err
	}
	o.OciOptions.CacheDir = filepath.Join(cliHomeDir, "components")
	if err := os.MkdirAll(o.OciOptions.CacheDir, os.ModePerm); err != nil {
		return fmt.Errorf("unable to create cache directory %s: %w", o.OciOptions.CacheDir, err)
	}

	if len(o.baseUrl) == 0 {
		return errors.New("the base url must be defined")
	}
	if len(o.componentName) == 0 {
		return errors.New("a component name must be defined")
	}
	if len(o.version) == 0 {
		return errors.New("a component's Version must be defined")
	}
	return nil
}

func (o *showOptions) AddFlags(fs *pflag.FlagSet) {
	o.OciOptions.AddFlags(fs)
}
