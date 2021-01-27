// SPDX-FileCopyrightText: 2020 SAP SE or an SAP affiliate company and Gardener contributors.
//
// SPDX-License-Identifier: Apache-2.0

package components

import (
	"context"
	"fmt"
	"os"

	"github.com/gardener/landscapercli/pkg/blueprints"
	"github.com/gardener/landscapercli/pkg/logger"
	"github.com/gardener/landscapercli/pkg/util"

	"github.com/gardener/landscaper/pkg/apis/core/v1alpha1"
	"github.com/go-logr/logr"
	"github.com/spf13/cobra"
	"github.com/spf13/pflag"
)

const addHelmLSDeployItemUse = `deployitem \
    [component directory path] \
    [execution name] \
    [deployitem name] \
   `

const addHelmLSDeployItemExample = `
landscaper-cli component add helm-ls deployitem \
  . \
  nginx \
  nginx \
  --oci-reference eu.gcr.io/gardener-project/landscaper/tutorials/charts/ingress-nginx:v0.1.0 \
  --chart-version v0.1.0`

const addHelmLSDeployItemShort = `
Command to add a deploy item skeleton to the blueprint of a component`

type addHelmLsDeployItemOptions struct {
	componentPath  string
	executionName  string
	deployItemName string

	ociReference string
	chartVersion string

	clusterParam  string
	targetNsParam string
}

// NewCreateCommand creates a new blueprint command to create a blueprint
func NewAddHelmLSDeployItemCommand(ctx context.Context) *cobra.Command {
	opts := &addHelmLsDeployItemOptions{}
	cmd := &cobra.Command{
		Use:     addHelmLSDeployItemUse,
		Example: addHelmLSDeployItemExample,
		Short:   addHelmLSDeployItemShort,
		Args:    cobra.ExactArgs(3),

		Run: func(cmd *cobra.Command, args []string) {
			if err := opts.Complete(args); err != nil {
				fmt.Println(err.Error())
				os.Exit(1)
			}

			if err := opts.run(ctx, logger.Log); err != nil {
				fmt.Println(err.Error())
				os.Exit(1)
			}

			fmt.Printf("Successfully added deploy item")
		},
	}

	opts.AddFlags(cmd.Flags())

	return cmd
}

func (o *addHelmLsDeployItemOptions) Complete(args []string) error {
	o.componentPath = args[0]
	o.executionName = args[1]
	o.deployItemName = args[2]

	return o.validate()
}

func (o *addHelmLsDeployItemOptions) AddFlags(fs *pflag.FlagSet) {
	fs.StringVar(&o.ociReference,
		"oci-reference",
		"",
		"reference to oci artifact containing the helm chart")
	fs.StringVar(&o.chartVersion,
		"chart-version",
		"",
		"helm chart version")
	fs.StringVar(&o.clusterParam,
		"cluster-param",
		"targetCluster",
		"target cluster")
	fs.StringVar(&o.targetNsParam,
		"target-ns-param",
		"",
		"target namespace")
}

func (o *addHelmLsDeployItemOptions) validate() error {
	if o.ociReference == "" {
		return fmt.Errorf("oci-reference is missing")
	}

	if o.chartVersion == "" {
		return fmt.Errorf("chart-version is missing")
	}

	if o.targetNsParam == "" {
		return fmt.Errorf("target-ns-param is missing")
	}

	return nil
}

func (o *addHelmLsDeployItemOptions) run(ctx context.Context, log logr.Logger) error {
	blueprintPath := util.BlueprintDirectoryPath(o.componentPath)
	blueprint, err := blueprints.NewBlueprintReader(blueprintPath).Read()
	if err != nil {
		return err
	}

	if o.existsExecution(blueprint) {
		return fmt.Errorf("The blueprint already contains a deploy execution %s\n", o.executionName)
	}

	exists, err := o.existsExecutionFile()
	if err != nil {
		return err
	}

	if exists {
		return fmt.Errorf("Deploy execution file %s already exists\n", util.ExecutionFilePath(o.componentPath, o.executionName))
	}

	err = o.createExecutionFile()
	if err != nil {
		return err
	}

	o.addExecution(blueprint)

	return blueprints.NewBlueprintWriter(blueprintPath).Write(blueprint)
}

func (o *addHelmLsDeployItemOptions) existsExecution(blueprint *v1alpha1.Blueprint) bool {
	for i := range blueprint.DeployExecutions {
		execution := &blueprint.DeployExecutions[i]
		if execution.Name == o.executionName {
			return true
		}
	}

	return false
}

func (o *addHelmLsDeployItemOptions) addExecution(blueprint *v1alpha1.Blueprint) {
	blueprint.DeployExecutions = append(blueprint.DeployExecutions, v1alpha1.TemplateExecutor{
		Name: o.executionName,
		Type: "GoTemplate",
		File: "/" + util.ExecutionFileName(o.executionName),
	})
}

func (o *addHelmLsDeployItemOptions) existsExecutionFile() (bool, error) {
	fileInfo, err := os.Stat(util.ExecutionFilePath(o.componentPath, o.executionName))
	if err != nil {
		if os.IsNotExist(err) {
			return false, nil
		}
		return false, err
	}

	if fileInfo.IsDir() {
		return false, fmt.Errorf("There already exists a directory %s\n", util.ExecutionFileName(o.executionName))
	}

	return true, nil
}

func (o *addHelmLsDeployItemOptions) createExecutionFile() error {
	f, err := os.Create(util.ExecutionFilePath(o.componentPath, o.executionName))
	if err != nil {
		return err
	}

	defer f.Close()

	_, err = f.WriteString("deployItems: []\n")

	return err
}
