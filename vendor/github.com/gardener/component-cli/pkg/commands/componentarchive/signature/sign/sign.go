// SPDX-FileCopyrightText: 2022 SAP SE or an SAP affiliate company and Gardener contributors.
//
// SPDX-License-Identifier: Apache-2.0
package sign

import (
	"context"
	"errors"
	"fmt"
	"os"
	"path/filepath"

	cdv2 "github.com/gardener/component-spec/bindings-go/apis/v2"
	cdv2Sign "github.com/gardener/component-spec/bindings-go/apis/v2/signatures"
	"github.com/gardener/component-spec/bindings-go/ctf"
	cdoci "github.com/gardener/component-spec/bindings-go/oci"
	"github.com/go-logr/logr"
	"github.com/mandelsoft/vfs/pkg/vfs"
	"github.com/spf13/cobra"
	"github.com/spf13/pflag"

	ociopts "github.com/gardener/component-cli/ociclient/options"
	"github.com/gardener/component-cli/pkg/commands/constants"
	"github.com/gardener/component-cli/pkg/logger"
	"github.com/gardener/component-cli/pkg/signatures"
)

// NewSignCommand creates a new command to interact with signatures.
func NewSignCommand(ctx context.Context) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "sign",
		Short: "command to sign component descriptors",
	}

	cmd.AddCommand(NewRSASignCommand(ctx))
	cmd.AddCommand(NewSigningServerSignCommand(ctx))
	return cmd
}

type GenericSignOptions struct {
	// BaseUrl is the oci registry where the component is stored.
	BaseUrl string
	// ComponentName is the unique name of the component in the registry.
	ComponentName string
	// Version is the component Version in the oci registry.
	Version string

	// SignatureName defines the name for the generated signature
	SignatureName string

	// UploadBaseUrlForSigned is the base url where the signed component descriptor will be uploaded
	UploadBaseUrlForSigned string

	// Force to overwrite component descriptors on upload
	Force bool

	// RecursiveSigning to enable/disable signing and uploading of all referenced components
	RecursiveSigning bool

	// SkipAccessTypes defines the access types that will be ignored for signing
	SkipAccessTypes []string

	// OciOptions contains all exposed options to configure the oci client.
	OciOptions ociopts.Options
}

//Complete validates the arguments and flags from the command line
func (o *GenericSignOptions) Complete(args []string) error {
	o.BaseUrl = args[0]
	o.ComponentName = args[1]
	o.Version = args[2]

	cliHomeDir, err := constants.CliHomeDir()
	if err != nil {
		return err
	}

	o.OciOptions.CacheDir = filepath.Join(cliHomeDir, "components")
	if err := os.MkdirAll(o.OciOptions.CacheDir, os.ModePerm); err != nil {
		return fmt.Errorf("unable to create cache directory %s: %w", o.OciOptions.CacheDir, err)
	}

	if len(o.BaseUrl) == 0 {
		return errors.New("the base url must be defined")
	}
	if len(o.ComponentName) == 0 {
		return errors.New("a component name must be defined")
	}
	if len(o.Version) == 0 {
		return errors.New("a component's Version must be defined")
	}
	if o.UploadBaseUrlForSigned == "" {
		return errors.New("upload-base-url must be defined")
	}
	if o.SignatureName == "" {
		return errors.New("a signature name must be provided")
	}
	return nil
}

func (o *GenericSignOptions) AddFlags(fs *pflag.FlagSet) {
	fs.StringVar(&o.SignatureName, "signature-name", "", "name of the signature")
	fs.StringVar(&o.UploadBaseUrlForSigned, "upload-base-url", "", "target repository context to upload the signed cd")
	fs.StringSliceVar(&o.SkipAccessTypes, "skip-access-types", []string{}, "[OPTIONAL] comma separated list of access types that will not be digested and signed")
	fs.BoolVar(&o.Force, "force", false, "[OPTIONAL] force overwrite of already existing component descriptors")
	fs.BoolVar(&o.RecursiveSigning, "recursive", false, "[OPTIONAL] recursively sign and upload all referenced component descriptors")
	o.OciOptions.AddFlags(fs)
}

func (o *GenericSignOptions) SignAndUploadWithSigner(ctx context.Context, log logr.Logger, fs vfs.FileSystem, signer cdv2Sign.Signer) error {
	repoCtx := cdv2.NewOCIRegistryRepository(o.BaseUrl, "")

	ociClient, cache, err := o.OciOptions.Build(log, fs)
	if err != nil {
		return fmt.Errorf("unable to build oci client: %s", err.Error())
	}

	cdresolver := cdoci.NewResolver(ociClient)
	cd, blobResolver, err := cdresolver.ResolveWithBlobResolver(ctx, repoCtx, o.ComponentName, o.Version)
	if err != nil {
		return fmt.Errorf("unable to to fetch component descriptor %s:%s: %w", o.ComponentName, o.Version, err)
	}

	blobResolvers := map[string]ctf.BlobResolver{}
	blobResolvers[fmt.Sprintf("%s:%s", cd.Name, cd.Version)] = blobResolver

	skipAccessTypesMap := map[string]bool{}
	for _, v := range o.SkipAccessTypes {
		skipAccessTypesMap[v] = true
	}

	digestedCds, err := signatures.RecursivelyAddDigestsToCd(cd, *repoCtx, ociClient, blobResolvers, context.TODO(), skipAccessTypesMap)
	if err != nil {
		return fmt.Errorf("failed adding digests to cd: %w", err)
	}

	targetRepoCtx := cdv2.NewOCIRegistryRepository(o.UploadBaseUrlForSigned, "")

	if o.RecursiveSigning {
		for _, digestedCd := range digestedCds {
			hasher, err := cdv2Sign.HasherForName(cdv2Sign.SHA256)
			if err != nil {
				return fmt.Errorf("failed creating hasher: %w", err)
			}

			if err := cdv2Sign.SignComponentDescriptor(digestedCd, signer, *hasher, o.SignatureName); err != nil {
				return fmt.Errorf("failed signing component descriptor: %w", err)
			}
			logger.Log.Info(fmt.Sprintf("CD Signed %s %s", digestedCd.Name, digestedCd.Version))

			logger.Log.Info(fmt.Sprintf("Uploading to %s %s %s", o.UploadBaseUrlForSigned, digestedCd.Name, digestedCd.Version))

			if err := signatures.UploadCDPreservingLocalOciBlobs(ctx, *digestedCd, *targetRepoCtx, ociClient, cache, blobResolvers, o.Force, log); err != nil {
				return fmt.Errorf("failed uploading cd: %w", err)
			}
		}
	} else {
		hasher, err := cdv2Sign.HasherForName(cdv2Sign.SHA256)
		if err != nil {
			return fmt.Errorf("failed creating hasher: %w", err)
		}

		if err := cdv2Sign.SignComponentDescriptor(cd, signer, *hasher, o.SignatureName); err != nil {
			return fmt.Errorf("failed signing component descriptor: %w", err)
		}
		logger.Log.Info(fmt.Sprintf("CD Signed %s %s", cd.Name, cd.Version))

		logger.Log.Info(fmt.Sprintf("Uploading to %s %s %s", o.UploadBaseUrlForSigned, cd.Name, cd.Version))

		if err := signatures.UploadCDPreservingLocalOciBlobs(ctx, *cd, *targetRepoCtx, ociClient, cache, blobResolvers, o.Force, log); err != nil {
			return fmt.Errorf("failed uploading cd: %w", err)
		}
	}
	return nil
}
