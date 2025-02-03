package helm

import (
	"github.com/mandelsoft/vfs/pkg/memoryfs"
	"helm.sh/helm/v3/pkg/chart"
	"ocm.software/ocm/api/ocm"
	metav1 "ocm.software/ocm/api/ocm/compdesc/meta/v1"
	"ocm.software/ocm/api/ocm/extensions/download"
	"ocm.software/ocm/api/tech/helm/loader"
)

func downloadChart(comp ocm.ComponentVersionAccess, chartResourceName string) (*chart.Chart, error) {
	chartResource, err := comp.GetResource(metav1.NewIdentity(chartResourceName))
	if err != nil {
		return nil, err
	}
	fs := memoryfs.New()
	path, err := download.DownloadResource(comp.GetContext(), chartResource, "chart", download.WithFileSystem(fs))
	if err != nil {
		return nil, err
	}
	return loader.Load(path, fs)
}
