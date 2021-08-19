// SPDX-FileCopyrightText: 2020 SAP SE or an SAP affiliate company and Gardener contributors.
//
// SPDX-License-Identifier: Apache-2.0

package blueprints

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	ociclientopts "github.com/gardener/component-cli/ociclient/options"
	cdv2 "github.com/gardener/component-spec/bindings-go/apis/v2"
	"github.com/gardener/component-spec/bindings-go/codec"
	"github.com/gardener/component-spec/bindings-go/ctf"
	componentsregistry "github.com/gardener/landscaper/pkg/landscaper/registry/components"
	"github.com/go-logr/logr"
	"github.com/mandelsoft/vfs/pkg/layerfs"
	"github.com/mandelsoft/vfs/pkg/memoryfs"
	"github.com/mandelsoft/vfs/pkg/osfs"
	"github.com/mandelsoft/vfs/pkg/projectionfs"
	"github.com/mandelsoft/vfs/pkg/vfs"
	"github.com/spf13/cobra"
	"github.com/spf13/pflag"
	"k8s.io/apimachinery/pkg/runtime/serializer"
	"k8s.io/apimachinery/pkg/util/sets"
	"sigs.k8s.io/yaml"

	"github.com/gardener/landscaper/apis/core"
	lsv1alpha1 "github.com/gardener/landscaper/apis/core/v1alpha1"
	"github.com/gardener/landscaper/apis/core/validation"
	"github.com/gardener/landscaper/pkg/api"
	lsutils "github.com/gardener/landscaper/pkg/utils/landscaper"

	"github.com/gardener/landscapercli/pkg/logger"
)

const (
	OutputResourceDeployItems      = "deployitems"
	OutputResourceSubinstallations = "subinstallations"
)

var (
	OutputResourceAllTerms              = sets.NewString("all")
	OutputResourceDeployItemsTerms      = sets.NewString(OutputResourceDeployItems, "di")
	OutputResourceSubinstallationsTerms = sets.NewString(OutputResourceSubinstallations, "subinst", "inst")
)

const (
	JSONOut = "json"
	YAMLOut = "yaml"
)

const DeployItemOutputDir = "deployitems"
const SubinstallationOutputDir = "subinstallations"

// RenderOptions describes the options for the render command.
type RenderOptions struct {
	// BlueprintPath is the path to the directory containing the definition.
	BlueprintPath string
	// ComponentDescriptorPath is the path to the component descriptor to be used
	ComponentDescriptorPath string
	// AdditionalComponentDescriptorPath is the path to the component descriptor to be used
	AdditionalComponentDescriptorPath []string
	// ValueFiles is a list of file paths to value yaml files.
	ValueFiles []string
	// OutputFormat defines the format of the output
	OutputFormat string
	// OutDir is the directory where the rendered should be written to.
	OutDir string

	OCIOptions ociclientopts.Options

	outputResources         sets.String
	blueprint               *lsv1alpha1.Blueprint
	blueprintFs             vfs.FileSystem
	componentDescriptor     *cdv2.ComponentDescriptor
	componentDescriptorList *cdv2.ComponentDescriptorList
	componentResolver       ctf.ComponentResolver
}

// NewRenderCommand creates a new local command to render a blueprint instance locally
func NewRenderCommand(ctx context.Context) *cobra.Command {
	opts := &RenderOptions{}
	cmd := &cobra.Command{
		Use:     "render",
		Args:    cobra.RangeArgs(1, 2),
		Example: "landscaper-cli blueprints render BLUEPRINT_DIR [all,deployitems,subinstallations]",
		Short:   "renders the given blueprint",
		Long: fmt.Sprintf(`
Renders the blueprint with the given values files.
All value files are merged whereas the later defined will overwrite the values of the previous ones

By default all rendered resources are printed to stdout.
Specific resources can be printed by adding a second argument.
landscapercli local render [path to Blueprint directory] [resource]
Available resources are
- %s: renders all available resources
- %s: renders deployitems
- %s: renders subinstallations
`,
			strings.Join(OutputResourceAllTerms.List(), "|"),
			strings.Join(OutputResourceDeployItemsTerms.List(), "|"),
			strings.Join(OutputResourceSubinstallationsTerms.List(), "|")),
		Run: func(cmd *cobra.Command, args []string) {
			if err := opts.Complete(logger.Log, args, osfs.New()); err != nil {
				fmt.Println(err.Error())
				os.Exit(1)
			}

			if err := opts.Run(ctx, logger.Log, osfs.New()); err != nil {
				fmt.Println(err.Error())
				os.Exit(1)
			}
		},
	}

	opts.AddFlags(cmd.Flags())
	return cmd
}

func (o *RenderOptions) AddFlags(fs *pflag.FlagSet) {
	fs.StringVarP(&o.ComponentDescriptorPath, "component-descriptor", "c", "", "Path to the local component descriptor")
	fs.StringArrayVarP(&o.AdditionalComponentDescriptorPath, "additional-component-descriptor", "a", []string{}, "Path to additional local component descriptors")
	fs.StringArrayVarP(&o.ValueFiles, "file", "f", []string{}, "List of filepaths to value yaml files that define the imports")
	fs.StringVarP(&o.OutputFormat, "output", "o", YAMLOut, "The format of the output. Can be json or yaml.")
	fs.StringVarP(&o.OutDir, "write", "w", "", "The output directory where the rendered files should be written to")
	o.OCIOptions.AddFlags(fs)
}

func (o *RenderOptions) Run(ctx context.Context, log logr.Logger, fs vfs.FileSystem) error {
	log.V(3).Info(fmt.Sprintf("rendering %s", strings.Join(o.outputResources.List(), ", ")))

	overlayFs := layerfs.New(memoryfs.New(), fs)
	if err := overlayFs.MkdirAll("/apptmp", os.ModePerm); err != nil {
		return err
	}

	renderArgs := lsutils.BlueprintRenderArgs{
		Fs:                          overlayFs,
		BlueprintPath:               o.BlueprintPath,
		ComponentDescriptorFilepath: o.ComponentDescriptorPath,
		ComponentResolver:           o.componentResolver,
		ComponentDescriptorList:     o.componentDescriptorList,
	}

	if err := o.setupImports(overlayFs, &renderArgs); err != nil {
		return err
	}
	out, err := lsutils.RenderBlueprint(renderArgs)
	if err != nil {
		return err
	}

	if o.outputResources.Has(OutputResourceDeployItems) {
		if len(out.DeployItems) == 0 {
			fmt.Println("No deploy items defined")
		}
		// print out state
		stateOut := map[string]map[string]json.RawMessage{
			"state": {},
		}
		for key, state := range out.DeployItemTemplateState {
			stateOut["state"][key] = state
		}
		if err := o.out(fs, stateOut, DeployItemOutputDir, "state"); err != nil {
			return err
		}

		for _, diTmpl := range out.DeployItems {
			if err := o.out(fs, diTmpl, DeployItemOutputDir, diTmpl.Name); err != nil {
				return err
			}
		}
	}

	if o.outputResources.Has(OutputResourceSubinstallations) {
		if len(out.Installations) == 0 {
			fmt.Println("No subinstallations defined")
		}

		// print out state
		stateOut := map[string]map[string]json.RawMessage{
			"state": {},
		}
		for key, state := range out.DeployItemTemplateState {
			stateOut["state"][key] = state
		}
		if err := o.out(fs, stateOut, SubinstallationOutputDir, "state"); err != nil {
			return err
		}
		for _, inst := range out.Installations {
			if err := o.out(fs, inst, SubinstallationOutputDir, inst.Name); err != nil {
				return err
			}
		}
	}

	return nil
}

func (o *RenderOptions) setupImports(fs vfs.FileSystem, args *lsutils.BlueprintRenderArgs) error {
	if len(o.ValueFiles) == 0 {
		return nil
	}
	values := lsutils.Imports{}
	for _, filePath := range o.ValueFiles {
		data, err := vfs.ReadFile(fs, filePath)
		if err != nil {
			return fmt.Errorf("unable to read values file '%s': %w", filePath, err)
		}
		tmpValues := &lsutils.Imports{}
		if err := yaml.Unmarshal(data, tmpValues); err != nil {
			return fmt.Errorf("unable to parse values file '%s': %w", filePath, err)
		}

		lsutils.MergeImports(&values, tmpValues)
	}
	args.Imports = &values
	return nil
}

func (o *RenderOptions) Complete(log logr.Logger, args []string, fs vfs.FileSystem) error {
	if len(o.ValueFiles) > 0 {
		for i := range o.ValueFiles {
			absPath, err := filepath.Abs(o.ValueFiles[i])
			if err != nil {
				return fmt.Errorf("unable get absolute values file path for %s: %w", o.ValueFiles[i], err)
			}
			o.ValueFiles[i] = absPath
		}
	}

	absBlueprintPath, err := filepath.Abs(args[0])
	if err != nil {
		return fmt.Errorf("unable get absolute blueprint path for %s: %w", args[0], err)
	}

	o.BlueprintPath = absBlueprintPath

	data, err := vfs.ReadFile(fs, filepath.Join(o.BlueprintPath, lsv1alpha1.BlueprintFileName))
	if err != nil {
		return fmt.Errorf("unable to read blueprint from %s: %w", filepath.Join(o.BlueprintPath, lsv1alpha1.BlueprintFileName), err)
	}
	o.blueprint = &lsv1alpha1.Blueprint{}
	if _, _, err := serializer.NewCodecFactory(api.LandscaperScheme).UniversalDecoder().Decode(data, nil, o.blueprint); err != nil {
		return err
	}
	o.blueprintFs, err = projectionfs.New(fs, o.BlueprintPath)
	if err != nil {
		return fmt.Errorf("unable to construct blueprint filesystem: %w", err)
	}

	if len(o.ComponentDescriptorPath) != 0 {
		absComponentDescriptorPath, err := filepath.Abs(o.ComponentDescriptorPath)
		if err != nil {
			return fmt.Errorf("unable get absolute component descriptor path for %s: %w", o.ComponentDescriptorPath, err)
		}

		o.ComponentDescriptorPath = absComponentDescriptorPath

		data, err := vfs.ReadFile(fs, o.ComponentDescriptorPath)
		if err != nil {
			return fmt.Errorf("unable to read component descriptor from %s: %w", o.ComponentDescriptorPath, err)
		}
		cd := &cdv2.ComponentDescriptor{}
		if err := codec.Decode(data, cd); err != nil {
			return fmt.Errorf("unable to decode component descriptor: %w", err)
		}
		o.componentDescriptor = cd
	}

	o.componentDescriptorList = &cdv2.ComponentDescriptorList{}
	for _, cdPath := range o.AdditionalComponentDescriptorPath {
		data, err := vfs.ReadFile(fs, cdPath)
		if err != nil {
			return fmt.Errorf("unable to read component descriptor from %s: %w", o.ComponentDescriptorPath, err)
		}
		cd := cdv2.ComponentDescriptor{}
		if err := codec.Decode(data, &cd); err != nil {
			return fmt.Errorf("unable to decode component descriptor: %w", err)
		}
		o.componentDescriptorList.Components = append(o.componentDescriptorList.Components, cd)
	}

	if err := o.parseOutputResources(args); err != nil {
		return err
	}

	// build component resolver with oci client
	ociClient, _, err := o.OCIOptions.Build(log, fs)
	if err != nil {
		return err
	}

	o.componentResolver, err = componentsregistry.NewOCIRegistryWithOCIClient(log, ociClient)
	if err != nil {
		return err
	}

	return o.Validate()
}

// Validate validates push options
func (o *RenderOptions) Validate() error {
	blueprint := &core.Blueprint{}
	if err := lsv1alpha1.Convert_v1alpha1_Blueprint_To_core_Blueprint(o.blueprint, blueprint, nil); err != nil {
		return err
	}
	if errList := validation.ValidateBlueprint(blueprint); len(errList) != 0 {
		return errList.ToAggregate()
	}

	if o.OutputFormat != YAMLOut && o.OutputFormat != JSONOut {
		return fmt.Errorf("output format is expected to be json or yaml but got '%s'", o.OutputFormat)
	}

	return nil
}

func (o *RenderOptions) parseOutputResources(args []string) error {
	allResources := sets.NewString(OutputResourceDeployItems, OutputResourceSubinstallations)
	if len(args) == 1 {
		o.outputResources = allResources
		return nil
	}
	if len(args) > 1 {
		resources := strings.Split(args[1], ",")
		o.outputResources = sets.NewString()
		for _, res := range resources {
			if OutputResourceAllTerms.Has(res) {
				o.outputResources = allResources
				return nil
			} else if OutputResourceDeployItemsTerms.Has(res) {
				o.outputResources.Insert(OutputResourceDeployItems)
			} else if OutputResourceSubinstallationsTerms.Has(res) {
				o.outputResources.Insert(OutputResourceSubinstallations)
			} else {
				return fmt.Errorf("unknown resource '%s'", res)
			}
		}
	}
	return nil
}

func (o *RenderOptions) out(fs vfs.FileSystem, obj interface{}, names ...string) error {

	var data []byte
	switch o.OutputFormat {
	case YAMLOut:
		var err error
		data, err = yaml.Marshal(obj)
		if err != nil {
			return err
		}
	case JSONOut:
		var err error
		data, err = json.MarshalIndent(obj, "", "  ")
		if err != nil {
			return err
		}
	default:
		return fmt.Errorf("unknown output format '%s'", o.OutputFormat)
	}

	// print to stdout if no directory is given
	if len(o.OutDir) == 0 {

		if len(names) != 0 {
			fmt.Println("--------------------------------------")
			fmt.Printf("-- %s\n", strings.Join(names, " "))
			fmt.Println("--------------------------------------")
		}
		fmt.Printf("%s\n", data)
		return nil
	}

	objFilePath := filepath.Join(append([]string{o.OutDir}, names...)...)
	if err := fs.MkdirAll(filepath.Dir(objFilePath), os.ModePerm); err != nil {
		return fmt.Errorf("unable to create path %s", o.OutDir)
	}
	return vfs.WriteFile(fs, objFilePath, data, os.ModePerm)
}
