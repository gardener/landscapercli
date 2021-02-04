package util

import (
	cdv2 "github.com/gardener/component-spec/bindings-go/apis/v2"
)

func CreateComponentDescriptor(name, version, baseURL string) *cdv2.ComponentDescriptor {
	cd := &cdv2.ComponentDescriptor{
		Metadata: cdv2.Metadata{
			Version: cdv2.SchemaVersion,
		},
		ComponentSpec: cdv2.ComponentSpec{
			ObjectMeta: cdv2.ObjectMeta{
				Name:    name,
				Version: version,
			},
			Provider: cdv2.InternalProvider,
			RepositoryContexts: []cdv2.RepositoryContext{
				{
					Type:    cdv2.OCIRegistryType,
					BaseURL: baseURL,
				},
			},
			Resources:           []cdv2.Resource{},
			Sources:             []cdv2.Source{},
			ComponentReferences: []cdv2.ComponentReference{},
		},
	}
	return cd
}
