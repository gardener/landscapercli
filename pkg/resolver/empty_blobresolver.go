package resolver

import (
	"context"
	"fmt"
	"io"

	v2 "github.com/gardener/component-spec/bindings-go/apis/v2"
	"github.com/gardener/component-spec/bindings-go/ctf"
)

type emptyBlobResolver struct {
}

func (r emptyBlobResolver) Info(_ context.Context, res v2.Resource) (*ctf.BlobInfo, error) {
	return nil, fmt.Errorf("trying to get info on non-local resource %s with local blob resolver", res.Name)
}

func (r emptyBlobResolver) Resolve(_ context.Context, res v2.Resource, _ io.Writer) (*ctf.BlobInfo, error) {
	return nil, fmt.Errorf("trying to resolve non-local resource %s with local blob resolver", res.Name)
}
