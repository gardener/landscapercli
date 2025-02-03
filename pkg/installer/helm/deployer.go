package helm

import (
	"errors"
	"fmt"
	"os"

	"helm.sh/helm/v3/pkg/action"
	"helm.sh/helm/v3/pkg/chart"
	"helm.sh/helm/v3/pkg/cli"
	"helm.sh/helm/v3/pkg/release"
	"helm.sh/helm/v3/pkg/storage/driver"
	"ocm.software/ocm/api/ocm"
)

type Deployer struct {
	ComponentVersion  ocm.ComponentVersionAccess
	ChartResourceName string

	CreateNamespace bool
	Namespace       string
	ReleaseName     string
	Values          map[string]interface{}

	Kubeconfig string

	chart *chart.Chart
}

func (d *Deployer) Deploy() (err error) {
	d.chart, err = downloadChart(d.ComponentVersion, d.ChartResourceName)
	if err != nil {
		return err
	}

	_, err = d.installOrUpgradeChart()
	if err != nil {
		return err
	}

	return nil
}

func (d *Deployer) installOrUpgradeChart() (*release.Release, error) {
	_, err := d.getRelease()
	if errors.Is(err, driver.ErrReleaseNotFound) {
		return d.installChart()
	} else if err != nil {
		return nil, err
	}
	return d.upgradeChart()
}

func (d *Deployer) getActionConfiguration() *action.Configuration {
	settings := cli.New()
	settings.KubeConfig = d.Kubeconfig
	settings.SetNamespace(d.Namespace)
	restClientGetter := settings.RESTClientGetter()
	actionConfig := &action.Configuration{}
	actionConfig.Init(
		restClientGetter,
		d.Namespace,
		os.Getenv("HELM_DRIVER"),
		func(msg string, args ...interface{}) { fmt.Printf(msg, args...) },
	)
	return actionConfig
}

func (d *Deployer) getRelease() (*release.Release, error) {
	actionConfig := d.getActionConfiguration()
	getAction := action.NewGet(actionConfig)
	return getAction.Run(d.ReleaseName)
}

func (d *Deployer) installChart() (*release.Release, error) {
	actionConfig := d.getActionConfiguration()
	install := action.NewInstall(actionConfig)
	install.ReleaseName = d.ReleaseName
	install.Namespace = d.Namespace
	install.CreateNamespace = d.CreateNamespace
	return install.Run(d.chart, d.Values)
}

func (d *Deployer) upgradeChart() (*release.Release, error) {
	actionConfig := d.getActionConfiguration()
	upgrade := action.NewUpgrade(actionConfig)
	upgrade.MaxHistory = 10
	upgrade.Namespace = d.Namespace
	return upgrade.Run(d.ReleaseName, d.chart, d.Values)
}
