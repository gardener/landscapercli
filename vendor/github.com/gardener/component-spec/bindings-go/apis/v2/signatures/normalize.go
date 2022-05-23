package signatures

import (
	"bytes"
	"context"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"reflect"
	"sort"

	v2 "github.com/gardener/component-spec/bindings-go/apis/v2"
)

// Entry is used for normalisation and has to contain one key
type Entry map[string]interface{}

// AddDigestsToComponentDescriptor adds digest to componentReferences and resources as returned in the resolver functions. If a digest already exists, a mismatch against the resolved digest will return an error.
func AddDigestsToComponentDescriptor(ctx context.Context, cd *v2.ComponentDescriptor,
	compRefResolver func(context.Context, v2.ComponentDescriptor, v2.ComponentReference) (*v2.DigestSpec, error),
	resResolver func(context.Context, v2.ComponentDescriptor, v2.Resource) (*v2.DigestSpec, error)) error {

	for i, reference := range cd.ComponentReferences {
		digest, err := compRefResolver(ctx, *cd, reference)
		if err != nil {
			return fmt.Errorf("failed resolving componentReference for %s:%s: %w", reference.Name, reference.Version, err)
		}
		if reference.Digest != nil && !reflect.DeepEqual(reference.Digest, digest) {
			return fmt.Errorf("calculated cd reference digest mismatches existing digest %s:%s", reference.ComponentName, reference.Version)
		}
		cd.ComponentReferences[i].Digest = digest
	}

	for i, res := range cd.Resources {
		// special digest notation indicates to not digest the content
		if res.Digest != nil && reflect.DeepEqual(res.Digest, v2.NewExcludeFromSignatureDigest()) {
			continue
		}

		digest, err := resResolver(ctx, *cd, res)
		if err != nil {
			return fmt.Errorf("failed resolving resource for %s:%s: %w", res.Name, res.Version, err)
		}
		if res.Digest != nil && !reflect.DeepEqual(res.Digest, digest) {
			return fmt.Errorf("calculated resource digest mismatches existing digest %s:%s", res.Name, res.Version)
		}
		cd.Resources[i].Digest = digest
	}
	return nil
}

// HashForComponentDescriptor return the hash for the component-descriptor, if it is normaliseable
// (= componentReferences and resources contain digest field)
func HashForComponentDescriptor(cd v2.ComponentDescriptor, hash Hasher) (*v2.DigestSpec, error) {
	normalisedComponentDescriptor, err := normalizeComponentDescriptor(cd)
	if err != nil {
		return nil, fmt.Errorf("failed normalising component descriptor %w", err)
	}
	hash.HashFunction.Reset()
	if _, err = hash.HashFunction.Write(normalisedComponentDescriptor); err != nil {
		return nil, fmt.Errorf("failed hashing the normalisedComponentDescriptorJson: %w", err)
	}
	return &v2.DigestSpec{
		HashAlgorithm:          hash.AlgorithmName,
		NormalisationAlgorithm: string(v2.JsonNormalisationV1),
		Value:                  hex.EncodeToString(hash.HashFunction.Sum(nil)),
	}, nil
}

func normalizeComponentDescriptor(cd v2.ComponentDescriptor) ([]byte, error) {
	if err := isNormaliseable(cd); err != nil {
		return nil, fmt.Errorf("can not normalise component-descriptor %s:%s: %w", cd.Name, cd.Version, err)
	}

	meta := []Entry{
		{"schemaVersion": cd.Metadata.Version},
	}

	componentReferences := []interface{}{}
	for _, ref := range cd.ComponentSpec.ComponentReferences {
		extraIdentity := buildExtraIdentity(ref.ExtraIdentity)

		digest := []Entry{
			{"hashAlgorithm": ref.Digest.HashAlgorithm},
			{"normalisationAlgorithm": ref.Digest.NormalisationAlgorithm},
			{"value": ref.Digest.Value},
		}

		componentReference := []Entry{
			{"componentName": ref.ComponentName},
			{"name": ref.Name},
			{"version": ref.Version},
			{"extraIdentity": extraIdentity},
			{"digest": digest},
		}
		componentReferences = append(componentReferences, componentReference)
	}

	resources := []interface{}{}
	for _, res := range cd.ComponentSpec.Resources {
		extraIdentity := buildExtraIdentity(res.ExtraIdentity)

		//ignore access.type=None for normalisation and hash calculation
		if res.Access == nil || res.Access.Type == "None" {
			resource := []Entry{
				{"name": res.Name},
				{"version": res.Version},
				{"type": res.Type},
				{"relation": res.Relation},
				{"extraIdentity": extraIdentity},
			}
			resources = append(resources, resource)
			continue
		}

		digest := []Entry{
			{"hashAlgorithm": res.Digest.HashAlgorithm},
			{"normalisationAlgorithm": res.Digest.NormalisationAlgorithm},
			{"value": res.Digest.Value},
		}

		resource := []Entry{
			{"name": res.Name},
			{"version": res.Version},
			{"type": res.Type},
			{"relation": res.Relation},
			{"extraIdentity": extraIdentity},
			{"digest": digest},
		}
		resources = append(resources, resource)
	}

	componentSpec := []Entry{
		{"name": cd.ComponentSpec.Name},
		{"version": cd.ComponentSpec.Version},
		{"provider": cd.ComponentSpec.Provider},
		{"componentReferences": componentReferences},
		{"resources": resources},
	}

	normalizedComponentDescriptor := []Entry{
		{"meta": meta},
		{"component": componentSpec},
	}

	if err := deepSort(normalizedComponentDescriptor); err != nil {
		return nil, fmt.Errorf("failed sorting during normalisation: %w", err)
	}

	byteBuffer := bytes.NewBuffer([]byte{})
	encoder := json.NewEncoder(byteBuffer)
	encoder.SetEscapeHTML(false)

	if err := encoder.Encode(normalizedComponentDescriptor); err != nil {
		return nil, err
	}

	normalizedJson := byteBuffer.Bytes()

	// encoder.Encode appends a newline that we do not want
	if normalizedJson[len(normalizedJson)-1] == 10 {
		normalizedJson = normalizedJson[:len(normalizedJson)-1]
	}

	return normalizedJson, nil
}

func buildExtraIdentity(identity v2.Identity) []Entry {
	var extraIdentities []Entry
	for k, v := range identity {
		extraIdentities = append(extraIdentities, Entry{k: v})
	}
	return extraIdentities
}

// deepSort sorts Entry, []Enry and [][]Entry interfaces recursively, lexicographicly by key(Entry).
func deepSort(in interface{}) error {
	switch castIn := in.(type) {
	case []Entry:
		// sort the values recursively for every entry
		for _, entry := range castIn {
			val := getOnlyValueInEntry(entry)
			if err := deepSort(val); err != nil {
				return err
			}
		}
		// sort the entries based on the key
		sort.SliceStable(castIn, func(i, j int) bool {
			return getOnlyKeyInEntry(castIn[i]) < getOnlyKeyInEntry(castIn[j])
		})
	case Entry:
		val := getOnlyValueInEntry(castIn)
		if err := deepSort(val); err != nil {
			return err
		}
	case []interface{}:
		for _, v := range castIn {
			if err := deepSort(v); err != nil {
				return err
			}
		}
	case string:
		break
	case v2.ProviderType:
		break
	case v2.ResourceRelation:
		break
	default:
		return fmt.Errorf("unknown type in sorting. This should not happen")
	}
	return nil
}

func getOnlyKeyInEntry(entry Entry) string {
	var key string
	for k := range entry {
		key = k
	}
	return key
}

func getOnlyValueInEntry(entry Entry) interface{} {
	var value interface{}
	for _, v := range entry {
		value = v
	}
	return value
}

// isNormaliseable checks if componentReferences and resources contain digest.
// Resources are allowed to omit the digest, if res.access.type == None or res.access == nil.
// Does NOT verify if the digests are correct
func isNormaliseable(cd v2.ComponentDescriptor) error {
	// check for digests on component references
	for _, reference := range cd.ComponentReferences {
		if reference.Digest == nil || reference.Digest.HashAlgorithm == "" || reference.Digest.NormalisationAlgorithm == "" || reference.Digest.Value == "" {
			return fmt.Errorf("missing digest in componentReference for %s:%s", reference.Name, reference.Version)
		}
	}
	for _, res := range cd.Resources {
		if (res.Access != nil && res.Access.Type != "None") && res.Digest == nil {
			return fmt.Errorf("missing digest in resource for %s:%s", res.Name, res.Version)
		}
		if (res.Access == nil || res.Access.Type == "None") && res.Digest != nil {
			return fmt.Errorf("digest for resource with emtpy (None) access not allowed %s:%s", res.Name, res.Version)
		}
	}
	return nil
}
