// SPDX-FileCopyrightText: 2022 SAP SE or an SAP affiliate company and Gardener contributors.
//
// SPDX-License-Identifier: Apache-2.0
package verify

import (
	"context"
	"crypto/rsa"
	"errors"
	"fmt"
	"os"

	"github.com/go-logr/logr"
	"github.com/mandelsoft/vfs/pkg/osfs"
	"github.com/mandelsoft/vfs/pkg/vfs"
	"github.com/spf13/cobra"
	"github.com/spf13/pflag"

	cdv2 "github.com/gardener/component-spec/bindings-go/apis/v2"
	cdv2Sign "github.com/gardener/component-spec/bindings-go/apis/v2/signatures"

	"github.com/gardener/component-cli/pkg/logger"
	"github.com/gardener/component-cli/pkg/signatures"
)

type X509CertificateVerifyOptions struct {
	rootCACertPath           string
	intermediateCAsCertsPath string
	signingCertPath          string

	GenericVerifyOptions
}

func NewX509CertificateVerifyCommand(ctx context.Context) *cobra.Command {
	opts := &X509CertificateVerifyOptions{}
	cmd := &cobra.Command{
		Use:   "x509 BASE_URL COMPONENT_NAME VERSION",
		Args:  cobra.ExactArgs(3),
		Short: fmt.Sprintf("fetch the component descriptor from an oci registry and verify its integrity based on a x509 certificate chain and a %s signature", cdv2.SignatureAlgorithmRSAPKCS1v15),
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

func (o *X509CertificateVerifyOptions) Run(ctx context.Context, log logr.Logger, fs vfs.FileSystem) error {
	cert, err := signatures.CreateAndVerifyX509CertificateFromFiles(o.signingCertPath, o.intermediateCAsCertsPath, o.rootCACertPath)
	if err != nil {
		return fmt.Errorf("unable to create certificate from files: %w", err)
	}

	publicKey, ok := cert.PublicKey.(*rsa.PublicKey)
	if !ok {
		return fmt.Errorf("public key is not of type *rsa.PublicKey: %T", cert.PublicKey)
	}

	verifier, err := cdv2Sign.CreateRsaVerifier(publicKey)
	if err != nil {
		return fmt.Errorf("failed creating rsa verifier: %w", err)
	}

	if err := o.GenericVerifyOptions.VerifyWithVerifier(ctx, log, fs, verifier); err != nil {
		return fmt.Errorf("failed verifying cd: %w", err)
	}
	return nil
}

func (o *X509CertificateVerifyOptions) Complete(args []string) error {
	if err := o.GenericVerifyOptions.Complete(args); err != nil {
		return err
	}

	if o.signingCertPath == "" {
		return errors.New("a path to a signing certificate file must be given as flag")
	}

	return nil
}

func (o *X509CertificateVerifyOptions) AddFlags(fs *pflag.FlagSet) {
	o.GenericVerifyOptions.AddFlags(fs)
	fs.StringVar(&o.signingCertPath, "signing-cert", "", "path to a file containing the signing certificate file in PEM format")
	fs.StringVar(&o.intermediateCAsCertsPath, "intermediate-cas-certs", "", "[OPTIONAL] path to a file containing the intermediate CAs certificates in PEM format")
	fs.StringVar(&o.rootCACertPath, "root-ca-cert", "", "[OPTIONAL] path to a file containing the root CA certificate in PEM format. if empty, the system root CA certificate pool is used.")
}
