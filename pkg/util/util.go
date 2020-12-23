package util

import (
	"io"
	"strings"

	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/util/yaml"
)

func ParseUnstructuredK8sManifest(k8sManifest string) ([]*unstructured.Unstructured, error) {
	k8sObjects := []*unstructured.Unstructured{}
	decoder := yaml.NewYAMLOrJSONDecoder(strings.NewReader(k8sManifest), 1024)

	for {
		k8sObj := &unstructured.Unstructured{
			Object: map[string]interface{}{},
		}
		err := decoder.Decode(&k8sObj.Object)
		if err != nil {
			if err == io.EOF {
				break
			} else {
				return nil, err
			}
		}
		k8sObjects = append(k8sObjects, k8sObj)
	}

	return k8sObjects, nil
}