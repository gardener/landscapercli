// SPDX-FileCopyrightText: 2022 SAP SE or an SAP affiliate company and Gardener contributors.
//
// SPDX-License-Identifier: Apache-2.0
package signatures

import (
	"context"
	"encoding/hex"
	"fmt"
	"io"
	"io/ioutil"
	"net/http"
	"reflect"

	"github.com/gardener/component-cli/ociclient"
	"github.com/gardener/component-cli/pkg/logger"

	cdv2 "github.com/gardener/component-spec/bindings-go/apis/v2"
	"github.com/gardener/component-spec/bindings-go/apis/v2/signatures"
	cdoci "github.com/gardener/component-spec/bindings-go/oci"
)

type Digester struct {
	ociClient ociclient.Client
	hasher    signatures.Hasher
}

func NewDigester(ociClient ociclient.Client, hasher signatures.Hasher) *Digester {
	return &Digester{
		ociClient: ociClient,
		hasher:    hasher,
	}

}

func (d *Digester) DigestForResource(ctx context.Context, cd cdv2.ComponentDescriptor, res cdv2.Resource) (*cdv2.DigestSpec, error) {
	// return the digest for a resource that is defined to be ignored for signing
	if res.Digest != nil && reflect.DeepEqual(res.Digest, cdv2.NewExcludeFromSignatureDigest()) {
		return res.Digest, nil
	}

	switch res.Access.Type {
	case cdv2.OCIRegistryType:
		return d.digestForOciArtifact(ctx, cd, res)
	case cdv2.LocalOCIBlobType:
		return d.digestForLocalOciBlob(ctx, cd, res)
	case cdv2.S3AccessType:
		return d.digestForS3Access(ctx, cd, res)
	case "None":
		logger.Log.V(5).Info(fmt.Sprintf("access type %s found in %s %s", res.Access.Type, cd.Name, cd.Version))
		return nil, nil
	default:
		return nil, fmt.Errorf("access type %s not supported", res.Access.Type)
	}
}

func (d *Digester) digestForLocalOciBlob(ctx context.Context, componentDescriptor cdv2.ComponentDescriptor, res cdv2.Resource) (*cdv2.DigestSpec, error) {
	if res.Access.GetType() != cdv2.LocalOCIBlobType {
		return nil, fmt.Errorf("unsupported access type: %s", res.Access.Type)
	}

	repoctx := cdv2.OCIRegistryRepository{}
	if err := componentDescriptor.GetEffectiveRepositoryContext().DecodeInto(&repoctx); err != nil {
		return nil, fmt.Errorf("unable to decode repository context: %w", err)
	}

	tmpfile, err := ioutil.TempFile("", "")
	if err != nil {
		return nil, fmt.Errorf("unable to create tempfile: %w", err)
	}
	defer tmpfile.Close()

	resolver := cdoci.NewResolver(d.ociClient)
	_, blobResolver, err := resolver.ResolveWithBlobResolver(ctx, &repoctx, componentDescriptor.Name, componentDescriptor.Version)
	if err != nil {
		return nil, fmt.Errorf("unable to resolve component descriptor: %w", err)
	}
	if _, err := blobResolver.Resolve(ctx, res, tmpfile); err != nil {
		return nil, fmt.Errorf("unable to resolve blob: %w", err)
	}

	if _, err := tmpfile.Seek(0, io.SeekStart); err != nil {
		return nil, fmt.Errorf("unable to seek to beginning of tempfile: %w", err)
	}
	d.hasher.HashFunction.Reset()

	if _, err := io.Copy(d.hasher.HashFunction, tmpfile); err != nil {
		return nil, fmt.Errorf("unable to hash blob: %w", err)
	}
	return &cdv2.DigestSpec{
		HashAlgorithm:          d.hasher.AlgorithmName,
		NormalisationAlgorithm: string(cdv2.GenericBlobDigestV1),
		Value:                  hex.EncodeToString((d.hasher.HashFunction.Sum(nil))),
	}, nil
}

func (d *Digester) digestForOciArtifact(ctx context.Context, componentDescriptor cdv2.ComponentDescriptor, res cdv2.Resource) (*cdv2.DigestSpec, error) {
	if res.Access.GetType() != cdv2.OCIRegistryType {
		return nil, fmt.Errorf("unsupported access type: %s", res.Access.Type)
	}

	ociAccess := &cdv2.OCIRegistryAccess{}
	if err := res.Access.DecodeInto(ociAccess); err != nil {
		return nil, fmt.Errorf("unable to decode resource access: %w", err)
	}

	_, bytes, err := d.ociClient.GetRawManifest(ctx, ociAccess.ImageReference)
	if err != nil {
		return nil, fmt.Errorf("failed getting oci raw manifest: %w", err)
	}

	d.hasher.HashFunction.Reset()
	if _, err = d.hasher.HashFunction.Write(bytes); err != nil {
		return nil, fmt.Errorf("failed hashing oci raw manifest, %w", err)
	}

	return &cdv2.DigestSpec{
		HashAlgorithm:          d.hasher.AlgorithmName,
		NormalisationAlgorithm: string(cdv2.OciArtifactDigestV1),
		Value:                  hex.EncodeToString((d.hasher.HashFunction.Sum(nil))),
	}, nil
}

func (d *Digester) digestForS3Access(ctx context.Context, componentDescriptor cdv2.ComponentDescriptor, res cdv2.Resource) (*cdv2.DigestSpec, error) {
	log := logger.Log.WithValues("componentDescriptor", componentDescriptor.ComponentSpec.ObjectMeta, "resource.name", res.Name, "resource.version", res.Version, "resource.extraIdentity", res.ExtraIdentity)

	if res.Access.GetType() != cdv2.S3AccessType {
		return nil, fmt.Errorf("unsupported access type for s3 Access Digester: %s", res.Access.Type)
	}
	s3Access := &cdv2.S3Access{}
	if err := res.Access.DecodeInto(s3Access); err != nil {
		return nil, fmt.Errorf("unable to decode resource access: %w", err)
	}

	url := fmt.Sprintf("https://%s.s3.amazonaws.com/%s", s3Access.BucketName, s3Access.ObjectKey)
	resp, err := http.Get(url)
	if err != nil {
		return nil, fmt.Errorf("unable to access s3 access with url %s: %w", url, err)
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("unable to access s3 access with url %s, response code %d", url, resp.StatusCode)
	}
	log.V(5).Info(fmt.Sprintf("downloading and hashing %s bytes from s3 access", resp.Header.Get("Content-Length")))
	d.hasher.HashFunction.Reset()
	if _, err := io.Copy(d.hasher.HashFunction, resp.Body); err != nil {
		return nil, fmt.Errorf("unable to hash s3 access with url %s and hash function %s: %w", url, d.hasher.AlgorithmName, err)
	}
	return &cdv2.DigestSpec{
		HashAlgorithm:          d.hasher.AlgorithmName,
		NormalisationAlgorithm: string(cdv2.GenericBlobDigestV1),
		Value:                  hex.EncodeToString((d.hasher.HashFunction.Sum(nil))),
	}, nil

}
