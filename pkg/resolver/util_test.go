package resolver

import (
	"github.com/gardener/component-cli/pkg/commands/componentarchive/input"
	"github.com/gardener/landscaper/apis/mediatype"
	"github.com/stretchr/testify/assert"
	"k8s.io/utils/pointer"
	"testing"
)

func TestConversion(t *testing.T) {
	blobInput := input.BlobInput{
		Type:             input.FileInputType,
		MediaType:        mediatype.JSONSchemaArtifactsMediaTypeV1,
		Path:             "./resources/schemas/vpa.json",
		CompressWithGzip: pointer.Bool(false),
	}

	access, err := convertInputToAccess(&blobInput)
	assert.NoError(t, err)

	blobInputResult, err := convertAccessToInput(access)
	assert.NoError(t, err)
	assert.Equal(t, blobInput.Type, blobInputResult.Type)
	assert.Equal(t, blobInput.MediaType, blobInputResult.MediaType)
	assert.Equal(t, blobInput.Path, blobInputResult.Path)
	assert.Equal(t, blobInput.CompressWithGzip, blobInputResult.CompressWithGzip)
}