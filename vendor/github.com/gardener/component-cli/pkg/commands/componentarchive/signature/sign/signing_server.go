// SPDX-FileCopyrightText: 2022 SAP SE or an SAP affiliate company and Gardener contributors.
//
// SPDX-License-Identifier: Apache-2.0
package sign

import (
	"context"
	"errors"
	"fmt"
	"os"

	"github.com/go-logr/logr"
	"github.com/mandelsoft/vfs/pkg/osfs"
	"github.com/mandelsoft/vfs/pkg/vfs"
	"github.com/spf13/cobra"
	"github.com/spf13/pflag"

	"github.com/gardener/component-cli/pkg/logger"
	"github.com/gardener/component-cli/pkg/signatures"
)

type SigningServerSignOptions struct {
	// SigningServerConfigPath path to the config file containing signing server configuration
	SigningServerConfigPath string

	GenericSignOptions
}

func NewSigningServerSignCommand(ctx context.Context) *cobra.Command {
	opts := &SigningServerSignOptions{}
	cmd := &cobra.Command{
		Use:   "signing-server BASE_URL COMPONENT_NAME VERSION",
		Args:  cobra.ExactArgs(3),
		Short: "fetch the component descriptor from an oci registry, sign it with a signature provided from a signing server, and re-upload",
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

func (o *SigningServerSignOptions) Run(ctx context.Context, log logr.Logger, fs vfs.FileSystem) error {
	signer, err := signatures.NewSigningServerSignerFromConfigFile(o.SigningServerConfigPath)
	if err != nil {
		return fmt.Errorf("unable to create signing server signer: %w", err)
	}
	return o.SignAndUploadWithSigner(ctx, log, fs, signer)
}

func (o *SigningServerSignOptions) Complete(args []string) error {
	if err := o.GenericSignOptions.Complete(args); err != nil {
		return err
	}

	if o.SigningServerConfigPath == "" {
		return errors.New("a config file which contains the signing server configuration must be given as flag")
	}

	return nil
}

func (o *SigningServerSignOptions) AddFlags(fs *pflag.FlagSet) {
	o.GenericSignOptions.AddFlags(fs)
	fs.StringVar(&o.SigningServerConfigPath, "config", "", "config file which contains the signing server configuration")
}
