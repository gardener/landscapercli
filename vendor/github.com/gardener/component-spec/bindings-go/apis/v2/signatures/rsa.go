package signatures

import (
	"crypto/rand"
	"crypto/rsa"
	"crypto/x509"
	"encoding/hex"
	"encoding/pem"
	"errors"
	"fmt"
	"io/ioutil"

	v2 "github.com/gardener/component-spec/bindings-go/apis/v2"
)

// RsaSigner is a signatures.Signer compatible struct to sign with RSASSA-PKCS1-V1_5.
type RsaSigner struct {
	privateKey rsa.PrivateKey
}

// CreateRsaSignerFromKeyFile creates an Instance of RsaSigner with the given private key.
// The private key has to be in the PKCS #1, ASN.1 DER form, see x509.ParsePKCS1PrivateKey.
func CreateRsaSignerFromKeyFile(pathToPrivateKey string) (*RsaSigner, error) {
	privKeyFile, err := ioutil.ReadFile(pathToPrivateKey)
	if err != nil {
		return nil, fmt.Errorf("failed opening private key file %w", err)
	}

	block, _ := pem.Decode([]byte(privKeyFile))
	if block == nil {
		return nil, fmt.Errorf("failed decoding PEM formatted block in key %w", err)
	}
	untypedPrivateKey, err := x509.ParsePKCS8PrivateKey(block.Bytes)
	if err != nil {
		return nil, fmt.Errorf("failed parsing key %w", err)
	}

	key, ok := untypedPrivateKey.(*rsa.PrivateKey)
	if !ok {
		return nil, fmt.Errorf("parsed key is not of type *rsa.PrivateKey: %T", untypedPrivateKey)
	}

	return &RsaSigner{
		privateKey: *key,
	}, nil
}

// Sign returns the signature for the data for the component-descriptor.
func (s RsaSigner) Sign(componentDescriptor v2.ComponentDescriptor, digest v2.DigestSpec) (*v2.SignatureSpec, error) {
	decodedHash, err := hex.DecodeString(digest.Value)
	if err != nil {
		return nil, fmt.Errorf("failed decoding hash to bytes")
	}
	// ensure length of hash is correct
	if len(decodedHash) != 32 {
		return nil, fmt.Errorf("hash to sign has invalid length")
	}
	signature, err := rsa.SignPKCS1v15(rand.Reader, &s.privateKey, 0, decodedHash)
	if err != nil {
		return nil, fmt.Errorf("failed signing hash, %w", err)
	}
	return &v2.SignatureSpec{
		Algorithm: v2.SignatureAlgorithmRSAPKCS1v15,
		Value:     hex.EncodeToString(signature),
		MediaType: v2.MediaTypeRSASignature,
	}, nil
}

// RsaVerifier is a signatures.Verifier compatible struct to verify RSASSA-PKCS1-V1_5 signatures.
type RsaVerifier struct {
	publicKey rsa.PublicKey
}

// CreateRsaVerifier creates an instance of RsaVerifier from a given rsa public key.
func CreateRsaVerifier(publicKey *rsa.PublicKey) (*RsaVerifier, error) {
	if publicKey == nil {
		return nil, errors.New("public key must not be nil")
	}

	verifier := RsaVerifier{
		publicKey: *publicKey,
	}

	return &verifier, nil
}

// CreateRsaVerifierFromKeyFile creates an instance of RsaVerifier from a rsa public key file.
// The private key has to be in the PKIX, ASN.1 DER form, see x509.ParsePKIXPublicKey.
func CreateRsaVerifierFromKeyFile(pathToPublicKey string) (*RsaVerifier, error) {
	publicKey, err := ioutil.ReadFile(pathToPublicKey)
	if err != nil {
		return nil, fmt.Errorf("failed opening public key file %w", err)
	}
	block, _ := pem.Decode([]byte(publicKey))
	if block == nil {
		return nil, fmt.Errorf("failed decoding PEM formatted block in key %w", err)
	}
	untypedKey, err := x509.ParsePKIXPublicKey(block.Bytes)
	if err != nil {
		return nil, fmt.Errorf("failed parsing key %w", err)
	}
	switch key := untypedKey.(type) {
	case *rsa.PublicKey:
		return CreateRsaVerifier(key)
	default:
		return nil, fmt.Errorf("public key format is not supported. Only rsa.PublicKey is supported")
	}
}

// Verify checks the signature, returns an error on verification failure
func (v RsaVerifier) Verify(componentDescriptor v2.ComponentDescriptor, signature v2.Signature) error {
	var signatureBytes []byte
	var err error
	switch signature.Signature.MediaType {
	case v2.MediaTypeRSASignature:
		signatureBytes, err = hex.DecodeString(signature.Signature.Value)
		if err != nil {
			return fmt.Errorf("unable to get signature value: failed decoding hash %s: %w", signature.Digest.Value, err)
		}
	case v2.MediaTypePEM:
		signaturePemBlocks, err := GetSignaturePEMBlocks([]byte(signature.Signature.Value))
		if err != nil {
			return fmt.Errorf("unable to get signature pem blocks: %w", err)
		}
		if len(signaturePemBlocks) != 1 {
			return fmt.Errorf("expected 1 signature pem block, found %d", len(signaturePemBlocks))
		}
		signatureBytes = signaturePemBlocks[0].Bytes
	default:
		return fmt.Errorf("invalid signature mediaType %s", signature.Signature.MediaType)
	}

	decodedHash, err := hex.DecodeString(signature.Digest.Value)
	if err != nil {
		return fmt.Errorf("failed decoding hash %s: %w", signature.Digest.Value, err)
	}
	// ensure length of hash is correct
	if len(decodedHash) != 32 {
		return fmt.Errorf("hash to verify has invalid length")
	}
	if err := rsa.VerifyPKCS1v15(&v.publicKey, 0, decodedHash, signatureBytes); err != nil {
		return fmt.Errorf("signature verification failed, %w", err)
	}
	return nil
}

// GetSignaturePEMBlocks returns all signature pem blocks from a list of pem blocks
func GetSignaturePEMBlocks(pemData []byte) ([]*pem.Block, error) {
	if len(pemData) == 0 {
		return []*pem.Block{}, nil
	}

	signatureBlocks := []*pem.Block{}
	for {
		var currentBlock *pem.Block
		currentBlock, pemData = pem.Decode(pemData)
		if currentBlock == nil && len(pemData) > 0 {
			return nil, fmt.Errorf("unable to decode pem block %s", string(pemData))
		}

		if currentBlock.Type == v2.SignaturePEMBlockType {
			signatureBlocks = append(signatureBlocks, currentBlock)
		}

		if len(pemData) == 0 {
			break
		}
	}

	return signatureBlocks, nil
}
