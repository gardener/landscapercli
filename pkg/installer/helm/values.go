package helm

import (
	"fmt"
	"strings"

	v1 "k8s.io/api/core/v1"
	"ocm.software/ocm/api/ocm"
	metav1 "ocm.software/ocm/api/ocm/compdesc/meta/v1"
	"ocm.software/ocm/api/ocm/extensions/artifacttypes"
	"ocm.software/ocm/api/ocm/ocmutils"
)

type ValuesBuilder struct {
	values map[string]any
	error  error
}

func NewValuesBuilder() *ValuesBuilder {
	return &ValuesBuilder{
		values: make(map[string]any),
	}
}

func NewValuesBuilderWithValues(values map[string]any) *ValuesBuilder {
	return &ValuesBuilder{
		values: values,
	}
}

func (b *ValuesBuilder) Values() (map[string]any, error) {
	return b.values, b.error
}

func (b *ValuesBuilder) WithImagePullSecrets(secretNames []string, path ...string) *ValuesBuilder {
	if b.error != nil {
		return b
	}

	injectImagePullSecrets(secretNames, b.values, path)
	return b
}

func (b *ValuesBuilder) WithImage(repo ocm.Repository, comp ocm.ComponentVersionAccess, imageResourceName string, path ...string) *ValuesBuilder {
	if b.error != nil {
		return b
	}

	b.values, b.error = injectImage(repo, comp, imageResourceName, b.values, path)
	if b.error != nil {
		return b
	}
	return b
}

func injectImagePullSecrets(secretNames []string, values map[string]any, path []string) {
	secrets := make([]v1.LocalObjectReference, len(secretNames))
	for i, secretName := range secretNames {
		secrets[i] = v1.LocalObjectReference{Name: secretName}
	}
	injectValue(values, path, secrets)
}

func injectImage(repo ocm.Repository, comp ocm.ComponentVersionAccess, imageResourceName string, values map[string]any, path []string) (map[string]any, error) {
	res, err := comp.GetResource(metav1.NewIdentity(imageResourceName))
	if err != nil {
		return nil, err
	}
	ref, err := ocmutils.GetOCIArtifactRef(repo.GetContext(), res)
	if err != nil {
		return nil, err
	}
	acc, err := res.Access()
	if err != nil {
		return nil, err
	}
	imageRepo := ""
	imageVers := ""
	if acc.GetType() == artifacttypes.OCI_ARTIFACT {
		imageRepo, imageVers, err = splitOCIArtifactRef(ref)
		if err != nil {
			return nil, err
		}
	}
	image := map[string]any{
		"image":      imageRepo,
		"tag":        imageVers,
		"pullPolicy": "IfNotPresent",
	}

	return injectValue(values, path, image)
}

func injectValue(values map[string]any, path []string, value any) (map[string]any, error) {
	if len(path) == 0 {
		return nil, fmt.Errorf("path must not be empty")
	}

	if len(path) == 1 {
		values[path[0]] = value
		return values, nil
	}

	var subValuesMap map[string]any
	if v, found := values[path[0]]; !found || v == nil {
		subValuesMap = make(map[string]any)
	} else if m, ok := values[path[0]].(map[string]any); ok {
		subValuesMap = m
	} else {
		subValuesMap = make(map[string]any)
	}

	var err error
	values[path[0]], err = injectValue(subValuesMap, path[1:], value)
	if err != nil {
		return nil, err
	}
	return values, nil
}

func splitOCIArtifactRef(ref string) (repository string, version string, err error) {
	s := strings.SplitN(ref, ":", 2)
	if len(s) != 2 {
		return "", "", fmt.Errorf("invalid OCI artifact reference: %s", ref)
	}
	return s[0], s[1], nil
}
