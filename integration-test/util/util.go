package util

import (
	"fmt"

	"k8s.io/apimachinery/pkg/util/uuid"

	cdv2 "github.com/gardener/component-spec/bindings-go/apis/v2"
	"github.com/stretchr/testify/assert"
)

const (
	Testuser = "testuser"
)

var Testpw string

func init() {
	Testpw = string(uuid.NewUUID())
}

// DummyTestingT is a utility struct for using assertions from "github.com/stretchr/testify/assert"
// in the integration tests (outside of unit tests)
type DummyTestingT struct{}

var _ assert.TestingT = DummyTestingT{}

func (t DummyTestingT) Errorf(format string, args ...interface{}) {
	fmt.Printf(format, args...)
}

func CreateComponentDescriptor(name, version, baseURL string) *cdv2.ComponentDescriptor {
	repoCtx, _ := cdv2.NewUnstructured(cdv2.NewOCIRegistryRepository(baseURL, ""))
	cd := &cdv2.ComponentDescriptor{
		Metadata: cdv2.Metadata{
			Version: cdv2.SchemaVersion,
		},
		ComponentSpec: cdv2.ComponentSpec{
			ObjectMeta: cdv2.ObjectMeta{
				Name:    name,
				Version: version,
			},
			Provider:            "internal",
			RepositoryContexts:  []*cdv2.UnstructuredTypedObject{&repoCtx},
			Resources:           []cdv2.Resource{},
			Sources:             []cdv2.Source{},
			ComponentReferences: []cdv2.ComponentReference{},
		},
	}
	return cd
}
