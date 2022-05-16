// SPDX-FileCopyrightText: 2022 SAP SE or an SAP affiliate company and Gardener contributors.
//
// SPDX-License-Identifier: Apache-2.0
package signatures

import (
	"bytes"
	"encoding/hex"
	"encoding/pem"
	"errors"
	"fmt"
	"io/ioutil"
	"net/http"

	cdv2 "github.com/gardener/component-spec/bindings-go/apis/v2"
	cdv2signatures "github.com/gardener/component-spec/bindings-go/apis/v2/signatures"
	"sigs.k8s.io/yaml"
)

const (
	AcceptHeader = "Accept"
)

type SigningServerSigner struct {
	Url      string `json:"url"`
	Username string `json:"username"`
	Password string `json:"password"`
}

func NewSigningServerSignerFromConfigFile(configFilePath string) (*SigningServerSigner, error) {
	configBytes, err := ioutil.ReadFile(configFilePath)
	if err != nil {
		return nil, fmt.Errorf("failed reading config file: %w", err)
	}
	var signer SigningServerSigner
	if err := yaml.Unmarshal(configBytes, &signer); err != nil {
		return nil, fmt.Errorf("failed parsing config yaml: %w", err)
	}
	return &signer, nil
}

func (signer *SigningServerSigner) Sign(componentDescriptor cdv2.ComponentDescriptor, digest cdv2.DigestSpec) (*cdv2.SignatureSpec, error) {
	decodedHash, err := hex.DecodeString(digest.Value)
	if err != nil {
		return nil, fmt.Errorf("failed decoding hash: %w", err)
	}

	req, err := http.NewRequest(http.MethodPost, fmt.Sprintf("%s/sign", signer.Url), bytes.NewBuffer(decodedHash))
	if err != nil {
		return nil, fmt.Errorf("failed building http request: %w", err)
	}
	req.Header.Add(AcceptHeader, cdv2.MediaTypePEM)
	req.SetBasicAuth(signer.Username, signer.Password)

	client := http.Client{}
	res, err := client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed sending request: %w", err)
	}
	defer res.Body.Close()

	responseBodyBytes, err := ioutil.ReadAll(res.Body)
	if err != nil {
		return nil, fmt.Errorf("failed reading response body: %w", err)
	}

	if res.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("request returned with response code %d: %s", res.StatusCode, string(responseBodyBytes))
	}

	signaturePemBlocks, err := cdv2signatures.GetSignaturePEMBlocks(responseBodyBytes)
	if err != nil {
		return nil, fmt.Errorf("failed getting signature pem block from response: %w", err)
	}

	if len(signaturePemBlocks) != 1 {
		return nil, fmt.Errorf("expected 1 signature pem block, found %d", len(signaturePemBlocks))
	}
	signatureBlock := signaturePemBlocks[0]

	signature := signatureBlock.Bytes
	if len(signature) == 0 {
		return nil, errors.New("invalid response: signature block doesn't contain signature")
	}

	algorithm := signatureBlock.Headers[cdv2.SignaturePEMBlockAlgorithmHeader]
	if algorithm == "" {
		return nil, fmt.Errorf("invalid response: %s header is empty", cdv2.SignaturePEMBlockAlgorithmHeader)
	}

	encodedSignature := pem.EncodeToMemory(signatureBlock)

	return &cdv2.SignatureSpec{
		Algorithm: algorithm,
		Value:     string(encodedSignature),
		MediaType: cdv2.MediaTypePEM,
	}, nil
}
