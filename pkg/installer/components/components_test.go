package components

import (
	"fmt"
	"testing"

	"github.com/mandelsoft/vfs/pkg/memoryfs"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	metav1 "ocm.software/ocm/api/ocm/compdesc/meta/v1"
	"ocm.software/ocm/api/ocm/extensions/download"
	"ocm.software/ocm/api/tech/helm/loader"
	"ocm.software/ocm/api/utils/tarutils"
)

func TestConfig(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "Components Test Suite")
}

var _ = Describe("Components", func() {

	It("should retrieve a component version", func() {
		baseURL := "europe-docker.pkg.dev/sap-gcp-cp-k8s-stable-hub/landscaper"
		componentName := "github.com/gardener/landscaper-service"
		componentVersion := "v0.121.0"

		repo, err := RetrieveRepository(baseURL)
		Expect(err).NotTo(HaveOccurred())
		defer repo.Close()

		cvLandscaperService, err := RetrieveComponentVersion(repo, componentName, componentVersion)
		Expect(err).NotTo(HaveOccurred())
		defer cvLandscaperService.Close()

		cvLandscaperInstance, err := RetrieveReferencedComponentVersion(repo, cvLandscaperService, "landscaper-instance")
		Expect(err).NotTo(HaveOccurred())
		defer cvLandscaperInstance.Close()

		cvLandscaper, err := RetrieveReferencedComponentVersion(repo, cvLandscaperInstance, "landscaper")
		Expect(err).NotTo(HaveOccurred())
		defer cvLandscaper.Close()

		cvManifestDeployer, err := RetrieveReferencedComponentVersion(repo, cvLandscaper, "manifest-deployer")
		Expect(err).NotTo(HaveOccurred())
		defer cvManifestDeployer.Close()

		resChart, err := cvManifestDeployer.GetResource(metav1.NewIdentity("manifest-deployer-chart"))
		Expect(err).NotTo(HaveOccurred())

		fs := memoryfs.New()

		path, err := download.DownloadResource(cvManifestDeployer.GetContext(), resChart, "chart", download.WithFileSystem(fs))
		Expect(err).NotTo(HaveOccurred())
		Expect(path).NotTo(BeNil())

		// report found files
		files, err := tarutils.ListArchiveContent(path, fs)
		Expect(err).NotTo(HaveOccurred())

		fmt.Printf("files for helm chart:\n")
		for _, f := range files {
			fmt.Printf("- %s\n", f)
		}

		chart, err := loader.Load(path, fs)
		Expect(err).NotTo(HaveOccurred())
		Expect(chart).NotTo(BeNil())

		//release, err := helm.InstallChart(chart, helm.ManifestDeployerOptions())
		//Expect(err).NotTo(HaveOccurred())
		//Expect(release).NotTo(BeNil())
	})
})
