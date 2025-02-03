package landscaper

import (
	"fmt"

	"ocm.software/ocm/api/ocm"

	"github.com/gardener/landscapercli/pkg/installer/components"
	"github.com/gardener/landscapercli/pkg/installer/helm"
)

type Installer struct {
	// BaseURL is the base URL of the repository where the components are stored.
	BaseURL string

	// LandscaperVersion is the version of the ocm component "github.com/gardener/landscaper".
	LandscaperVersion string

	// Kubeconfig is the path to the kubeconfig file of the cluster where the landscaper should be installed.
	Kubeconfig string

	// Namespace is the namespace where the landscaper should be installed.
	Namespace string

	// Verbosity is the verbosity level of the installation.
	Verbosity string
}

// Deploy installs the landscaper and its deployers.
// TODO: image pull secrets are not created yet. (But the secret names are added to the helm values.)
func (i *Installer) Deploy() error {

	const landscaperComponentName = "github.com/gardener/landscaper"

	repo, err := components.RetrieveRepository(i.BaseURL)
	if err != nil {
		return err
	}
	defer repo.Close()

	cvLandscaper, err := components.RetrieveComponentVersion(repo, landscaperComponentName, i.LandscaperVersion)
	if err != nil {
		return err
	}
	defer cvLandscaper.Close()

	err = i.deployLandscaper(repo, cvLandscaper)
	if err != nil {
		return err
	}

	err = i.deployHelmDeployer(repo, cvLandscaper)
	if err != nil {
		return err
	}

	err = i.deployManifestDeployer(repo, cvLandscaper)
	if err != nil {
		return err
	}

	err = i.deployContainerDeployer(repo, cvLandscaper)
	if err != nil {
		return err
	}

	return nil
}

//// Retrieving the landscaper component version from the landscaper-service component version.
//
//const componentName = "github.com/gardener/landscaper-service"
//cvLandscaperService, err := components.RetrieveComponentVersion(repo, componentName, i.LandscaperServiceVersion)
//if err != nil {
//	return err
//}
//defer cvLandscaperService.Close()
//
//cvLandscaperInstance, err := components.RetrieveReferencedComponentVersion(repo, cvLandscaperService, "landscaper-instance")
//if err != nil {
//	return err
//}
//defer cvLandscaperInstance.Close()
//
//cvLandscaper, err := components.RetrieveReferencedComponentVersion(repo, cvLandscaperInstance, "landscaper")
//if err != nil {
//	return err
//}
//defer cvLandscaper.Close()

func (i *Installer) deployLandscaper(repo ocm.Repository, cvLandscaper ocm.ComponentVersionAccess) error {
	const (
		releaseName                     = "landscaper"
		chartResourceName               = "landscaper-chart"
		controllerImageResourceName     = "landscaper-controller"
		webhooksServerImageResourceName = "landscaper-webhooks-server"
	)

	values, err := helm.NewValuesBuilderWithValues(map[string]any{
		"landscaper": map[string]any{
			"landscaper": map[string]any{
				"verbosity": i.Verbosity,
			},
		},
	}).
		WithImagePullSecrets([]string{fmt.Sprintf("%s-img-credentials", releaseName)}, "landscaper", "imagePullSecrets").
		WithImage(repo, cvLandscaper, controllerImageResourceName, "landscaper", "controller", "image").
		WithImage(repo, cvLandscaper, webhooksServerImageResourceName, "landscaper", "webhooksServer", "image").
		Values()
	if err != nil {
		return err
	}

	d := helm.Deployer{
		ComponentVersion:  cvLandscaper,
		ChartResourceName: chartResourceName,
		CreateNamespace:   true,
		Kubeconfig:        i.Kubeconfig,
		Namespace:         i.Namespace,
		ReleaseName:       releaseName,
		Values:            values,
	}
	if err := d.Deploy(); err != nil {
		return err
	}

	return nil
}

func (i *Installer) deployHelmDeployer(repo ocm.Repository, cvLandscaper ocm.ComponentVersionAccess) error {
	const (
		componentRef      = "helm-deployer"
		releaseName       = "helm-deployer"
		chartResourceName = "helm-deployer-chart"
		imageResourceName = "helm-deployer-image"
	)

	comp, err := components.RetrieveReferencedComponentVersion(repo, cvLandscaper, componentRef)
	if err != nil {
		return err
	}
	defer comp.Close()

	values, err := helm.NewValuesBuilder().
		WithImagePullSecrets([]string{fmt.Sprintf("%s-img-credentials", releaseName)}, "imagePullSecrets").
		WithImage(repo, comp, imageResourceName, "image").
		Values()
	if err != nil {
		return err
	}

	d := helm.Deployer{
		ComponentVersion:  comp,
		ChartResourceName: chartResourceName,
		CreateNamespace:   true,
		Kubeconfig:        i.Kubeconfig,
		Namespace:         i.Namespace,
		ReleaseName:       releaseName,
		Values:            values,
	}
	if err := d.Deploy(); err != nil {
		return err
	}

	return nil
}

func (i *Installer) deployManifestDeployer(repo ocm.Repository, cvLandscaper ocm.ComponentVersionAccess) error {
	const (
		componentRef      = "manifest-deployer"
		releaseName       = "manifest-deployer"
		chartResourceName = "manifest-deployer-chart"
		imageResourceName = "manifest-deployer-image"
	)

	comp, err := components.RetrieveReferencedComponentVersion(repo, cvLandscaper, componentRef)
	if err != nil {
		return err
	}
	defer comp.Close()

	values, err := helm.NewValuesBuilder().
		WithImagePullSecrets([]string{fmt.Sprintf("%s-img-credentials", releaseName)}, "imagePullSecrets").
		WithImage(repo, comp, imageResourceName, "image").
		Values()
	if err != nil {
		return err
	}

	d := helm.Deployer{
		ComponentVersion:  comp,
		ChartResourceName: chartResourceName,
		CreateNamespace:   true,
		Kubeconfig:        i.Kubeconfig,
		Namespace:         i.Namespace,
		ReleaseName:       releaseName,
		Values:            values,
	}
	if err := d.Deploy(); err != nil {
		return err
	}

	return nil
}

func (i *Installer) deployContainerDeployer(repo ocm.Repository, cvLandscaper ocm.ComponentVersionAccess) error {
	const (
		componentRef          = "container-deployer"
		releaseName           = "container-deployer"
		chartResourceName     = "container-deployer-chart"
		imageResourceName     = "container-deployer-image"
		initImageResourceName = "container-init-image"
		waitImageResourceName = "container-wait-image"
	)

	comp, err := components.RetrieveReferencedComponentVersion(repo, cvLandscaper, componentRef)
	if err != nil {
		return err
	}
	defer comp.Close()

	values, err := helm.NewValuesBuilder().
		WithImagePullSecrets([]string{fmt.Sprintf("%s-img-credentials", releaseName)}, "imagePullSecrets").
		WithImage(repo, comp, imageResourceName, "image").
		WithImage(repo, comp, initImageResourceName, "deployer", "initContainer", "image").
		WithImage(repo, comp, waitImageResourceName, "deployer", "waitContainer", "image").
		Values()
	if err != nil {
		return err
	}

	d := helm.Deployer{
		ComponentVersion:  comp,
		ChartResourceName: chartResourceName,
		CreateNamespace:   true,
		Kubeconfig:        i.Kubeconfig,
		Namespace:         i.Namespace,
		ReleaseName:       releaseName,
		Values:            values,
	}
	if err := d.Deploy(); err != nil {
		return err
	}

	return nil
}
