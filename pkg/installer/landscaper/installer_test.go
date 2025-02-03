package landscaper

import (
	"os"
	"testing"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

func TestConfig(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "Landscaper Installer Test Suite")
}

var _ = Describe("Landscaper Installer", func() {

	It("should deploy landscaper", func() {
		installer := &Installer{
			BaseURL:           "europe-docker.pkg.dev/sap-gcp-cp-k8s-stable-hub/landscaper",
			LandscaperVersion: "v0.126.0",
			Kubeconfig:        os.Getenv("KUBECONFIG"),
			Namespace:         "ls-system",
			Verbosity:         "info",
		}
		err := installer.Deploy()
		Expect(err).NotTo(HaveOccurred())
	})
})
