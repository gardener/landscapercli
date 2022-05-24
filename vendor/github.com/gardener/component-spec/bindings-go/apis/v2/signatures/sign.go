package signatures

import (
	"fmt"
	"reflect"

	v2 "github.com/gardener/component-spec/bindings-go/apis/v2"
)

// SignComponentDescriptor signs the given component-descriptor with the signer.
// The component-descriptor has to contain digests for componentReferences and resources.
func SignComponentDescriptor(cd *v2.ComponentDescriptor, signer Signer, hasher Hasher, signatureName string) error {
	hashedDigest, err := HashForComponentDescriptor(*cd, hasher)
	if err != nil {
		return fmt.Errorf("failed getting hash for cd: %w", err)
	}

	signature, err := signer.Sign(*cd, *hashedDigest)
	if err != nil {
		return fmt.Errorf("failed signing hash of normalised component descriptor, %w", err)
	}
	cd.Signatures = append(cd.Signatures, v2.Signature{
		Name:      signatureName,
		Digest:    *hashedDigest,
		Signature: *signature,
	})
	return nil
}

// VerifySignedComponentDescriptor verifies the signature (selected by signatureName) and hash of the component-descriptor (as specified in the signature).
// Does NOT resolve resources or referenced component-descriptors.
// Returns error if verification fails.
func VerifySignedComponentDescriptor(cd *v2.ComponentDescriptor, verifier Verifier, signatureName string) error {
	//find matching signature
	matchingSignature, err := SelectSignatureByName(cd, signatureName)
	if err != nil {
		return fmt.Errorf("failed checking signature: %w", err)
	}

	//Verify author of signature
	err = verifier.Verify(*cd, *matchingSignature)
	if err != nil {
		return fmt.Errorf("failed verifying: %w", err)
	}

	//get hasher by algorithm name
	hasher, err := HasherForName(matchingSignature.Digest.HashAlgorithm)
	if err != nil {
		return fmt.Errorf("failed creating hasher for %s: %w", matchingSignature.Digest.HashAlgorithm, err)
	}

	//Verify normalised cd to given (and verified) hash
	calculatedDigest, err := HashForComponentDescriptor(*cd, *hasher)
	if err != nil {
		return fmt.Errorf("failed hashing cd %s:%s: %w", cd.Name, cd.Version, err)
	}

	if !reflect.DeepEqual(*calculatedDigest, matchingSignature.Digest) {
		return fmt.Errorf("normalised component-descriptor does not match hash from signature")
	}

	return nil
}

// SelectSignatureByName returns the Signature (Digest and SigantureSpec) matching the given name
func SelectSignatureByName(cd *v2.ComponentDescriptor, signatureName string) (*v2.Signature, error) {
	for _, signature := range cd.Signatures {
		if signature.Name == signatureName {
			return &signature, nil
		}
	}
	return nil, fmt.Errorf("signature with name %s not found in component-descriptor", signatureName)

}
