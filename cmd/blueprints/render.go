// SPDX-FileCopyrightText: 2020 SAP SE or an SAP affiliate company and Gardener contributors.
//
// SPDX-License-Identifier: Apache-2.0

package blueprints

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"path"
	"path/filepath"
	"strings"

	ociclientopts "github.com/gardener/component-cli/ociclient/options"
	"github.com/gardener/component-cli/pkg/commands/componentarchive/resources"
	cdv2 "github.com/gardener/component-spec/bindings-go/apis/v2"
	"github.com/gardener/component-spec/bindings-go/codec"
	"github.com/gardener/landscaper/apis/core"
	lsv1alpha1 "github.com/gardener/landscaper/apis/core/v1alpha1"
	"github.com/gardener/landscaper/apis/core/validation"
	"github.com/gardener/landscaper/pkg/api"
	"github.com/gardener/landscaper/pkg/components/model"
	"github.com/gardener/landscaper/pkg/components/registries"
	"github.com/gardener/landscaper/pkg/landscaper/blueprints"
	lsutils "github.com/gardener/landscaper/pkg/utils/landscaper"
	"github.com/go-logr/logr"
	"github.com/mandelsoft/vfs/pkg/layerfs"
	"github.com/mandelsoft/vfs/pkg/memoryfs"
	"github.com/mandelsoft/vfs/pkg/osfs"
	"github.com/mandelsoft/vfs/pkg/projectionfs"
	"github.com/mandelsoft/vfs/pkg/vfs"
	"github.com/spf13/cobra"
	"github.com/spf13/pflag"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime/serializer"
	"k8s.io/apimachinery/pkg/util/sets"
	"sigs.k8s.io/yaml"

	"github.com/gardener/landscapercli/pkg/components"
	"github.com/gardener/landscapercli/pkg/logger"
	"github.com/gardener/landscapercli/pkg/resolver"
)

const (
	OutputResourceDeployItems      = "deployitems"
	OutputResourceSubinstallations = "subinstallations"
	OutputResourceImports          = "imports"
	OutputResourceExports          = "exports"
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
	// ResourcesPath is the path to the resources yaml file
	ResourcesPath string
	// ExportTemplatesPath is the path to the export templates yaml file
	ExportTemplatesPath string
	// ValueFiles is a list of file paths to value yaml files.
	ValueFiles []string
	// OutputFormat defines the format of the output
	OutputFormat string
	// OutDir is the directory where the rendered should be written to.
	OutDir string

	OCIOptions ociclientopts.Options

	outputResources         sets.Set[string]
	blueprint               *lsv1alpha1.Blueprint
	blueprintFs             vfs.FileSystem
	componentDescriptor     *cdv2.ComponentDescriptor
	componentDescriptorList *cdv2.ComponentDescriptorList
	//componentResolver       ctf.ComponentResolver
	registryAccess  model.RegistryAccess
	resources       []resources.ResourceOptions
	exportTemplates lsutils.ExportTemplates
}

// NewRenderCommand creates a new local command to render a blueprint instance locally
func NewRenderCommand(ctx context.Context) *cobra.Command {
	opts := &RenderOptions{}
	cmd := &cobra.Command{
		Use:     "render",
		Args:    cobra.RangeArgs(1, 2),
		Example: "landscaper-cli blueprints render BLUEPRINT_DIR [all,deployitems,subinstallations,imports,exports]",
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
	fs.StringVarP(&o.ResourcesPath, "resources", "r", "", "Path to the resources yaml file")
	fs.StringVarP(&o.ExportTemplatesPath, "export-templates", "e", "", "Path to the yaml file, defining the export templates")
	fs.StringArrayVarP(&o.ValueFiles, "file", "f", []string{}, "List of filepaths to value yaml files that define the imports")
	fs.StringVarP(&o.OutputFormat, "output", "o", YAMLOut, "The format of the output. Can be json or yaml.")
	fs.StringVarP(&o.OutDir, "write", "w", "", "The output directory where the rendered files should be written to")
	o.OCIOptions.AddFlags(fs)
}

func formatState(state map[string][]byte) map[string]map[string]json.RawMessage {
	stateOut := map[string]map[string]json.RawMessage{
		"state": {},
	}
	for key, value := range state {
		stateOut["state"][key] = value
	}
	return stateOut
}

type SimulatorCallbacks struct {
	options *RenderOptions
	fs      vfs.FileSystem
}

func (c SimulatorCallbacks) OnInstallation(installationPath string, installation *lsv1alpha1.Installation) {
	fmt.Printf("executing installation %s\n", installationPath)

	if c.options.outputResources.Has(OutputResourceSubinstallations) {
		_ = c.options.out(c.fs, installation, installationPath, "installation")
	}
}

func (c SimulatorCallbacks) OnInstallationTemplateState(installationPath string, state map[string][]byte) {
	if c.options.outputResources.Has(OutputResourceDeployItems) {
		_ = c.options.out(c.fs, formatState(state), installationPath, "installation", "state")
	}
}

func (c SimulatorCallbacks) OnImports(installationPath string, imports map[string]interface{}) {
	if c.options.outputResources.Has(OutputResourceImports) {
		_ = c.options.out(c.fs, imports, installationPath, "imports")
	}
}

func (c SimulatorCallbacks) OnDeployItem(installationPath string, deployItem *lsv1alpha1.DeployItem) {
	fmt.Printf("executing deploy item %s\n", path.Join(installationPath, deployItem.Name))

	if c.options.outputResources.Has(OutputResourceDeployItems) {
		_ = c.options.out(c.fs, deployItem, installationPath, "deployitems", deployItem.Name)
	}
}

func (c SimulatorCallbacks) OnDeployItemTemplateState(installationPath string, state map[string][]byte) {
	if c.options.outputResources.Has(OutputResourceDeployItems) {
		_ = c.options.out(c.fs, formatState(state), installationPath, "deployitems", "state")
	}
}

func (c SimulatorCallbacks) OnExports(installationPath string, exports map[string]interface{}) {
	if c.options.outputResources.Has(OutputResourceExports) {
		_ = c.options.out(c.fs, exports, installationPath, "exports")
	}
}

func (o *RenderOptions) Run(ctx context.Context, log logr.Logger, fs vfs.FileSystem) error {
	log.V(3).Info(fmt.Sprintf("rendering %s", strings.Join(sets.List(o.outputResources), ", ")))

	overlayFs := layerfs.New(memoryfs.New(), fs)
	if err := overlayFs.MkdirAll("/apptmp", os.ModePerm); err != nil {
		return err
	}

	imports, err := o.setupImports(overlayFs)
	if err != nil {
		return err
	}

	blueprint, err := blueprints.NewFromFs(o.blueprintFs)
	if err != nil {
		return err
	}

	var componentVersion model.ComponentVersion
	if o.componentDescriptor != nil {
		componentVersion, err = o.registryAccess.GetComponentVersion(ctx, &lsv1alpha1.ComponentDescriptorReference{
			RepositoryContext: o.componentDescriptor.GetEffectiveRepositoryContext(),
			ComponentName:     o.componentDescriptor.GetName(),
			Version:           o.componentDescriptor.GetVersion(),
		})
		if err != nil {
			return err
		}
	}

	componentVersionList := model.ComponentVersionList{
		Components: make([]model.ComponentVersion, 0, len(o.componentDescriptorList.Components)),
	}

	for _, cd := range o.componentDescriptorList.Components {
		fmt.Printf("resolving component descriptor %s:%s\n", cd.GetName(), cd.GetVersion())
		cv, err := o.registryAccess.GetComponentVersion(ctx, &lsv1alpha1.ComponentDescriptorReference{
			RepositoryContext: cd.GetEffectiveRepositoryContext(),
			ComponentName:     cd.GetName(),
			Version:           cd.GetVersion(),
		})
		if err != nil {
			return err
		}
		componentVersionList.Components = append(componentVersionList.Components, cv)
	}

	if componentVersion != nil {
		refs, err := componentVersion.GetComponentReferences()
		if err != nil {
			return err
		}

		for _, ref := range refs {
			if _, err := componentVersionList.GetComponentVersion(ref.ComponentName, ref.Version); err != nil {
				// not found
				fmt.Printf("resolving component descriptor reference %q (%s:%s)\n", ref.Name, ref.ComponentName, ref.Version)
				cv, err := o.registryAccess.GetComponentVersion(ctx, &lsv1alpha1.ComponentDescriptorReference{
					RepositoryContext: o.componentDescriptor.GetEffectiveRepositoryContext(),
					ComponentName:     ref.ComponentName,
					Version:           ref.Version,
				})
				if err != nil {
					return err
				}
				componentVersionList.Components = append(componentVersionList.Components, cv)
			}
		}
	}

	if len(o.ExportTemplatesPath) != 0 {
		simulator, err := lsutils.NewInstallationSimulator(&componentVersionList, o.registryAccess, nil, o.exportTemplates)
		if err != nil {
			return err
		}
		simulator.SetCallbacks(SimulatorCallbacks{options: o, fs: fs})
		_, err = simulator.Run(componentVersion, blueprint, imports.Imports)
		if err != nil {
			return err
		}
	} else {
		blueprintRenderer := lsutils.NewBlueprintRenderer(&componentVersionList, o.registryAccess, nil)

		installation := &lsv1alpha1.Installation{
			ObjectMeta: metav1.ObjectMeta{
				Name:      "root",
				Namespace: "default",
			},
			Spec: lsv1alpha1.InstallationSpec{
				Blueprint: lsv1alpha1.BlueprintDefinition{
					Reference: &lsv1alpha1.RemoteBlueprintReference{
						ResourceName: "example-blueprint",
					},
				},
			},
		}

		if o.componentDescriptor != nil {
			installation.Spec.ComponentDescriptor = &lsv1alpha1.ComponentDescriptorDefinition{
				Reference: &lsv1alpha1.ComponentDescriptorReference{
					ComponentName: o.componentDescriptor.ComponentSpec.Name,
					Version:       o.componentDescriptor.ComponentSpec.Version,
				},
			}
		} else {
			installation.Spec.ComponentDescriptor = &lsv1alpha1.ComponentDescriptorDefinition{
				Reference: &lsv1alpha1.ComponentDescriptorReference{
					ComponentName: "my-example-component",
					Version:       "v0.0.0",
				},
			}
		}

		out, err := blueprintRenderer.RenderDeployItemsAndSubInstallations(&lsutils.ResolvedInstallation{
			ComponentVersion: componentVersion,
			Installation:     installation,
			Blueprint:        blueprint,
		}, imports.Imports)

		if err != nil {
			return err
		}

		if o.outputResources.Has(OutputResourceDeployItems) {
			if len(out.DeployItems) == 0 {
				fmt.Println("No deploy items defined")
			}

			if err := o.out(fs, formatState(out.DeployItemTemplateState), DeployItemOutputDir, "state"); err != nil {
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

			if err := o.out(fs, formatState(out.InstallationTemplateState), SubinstallationOutputDir, "state"); err != nil {
				return err
			}

			for _, inst := range out.Installations {
				if err := o.out(fs, inst.Installation, SubinstallationOutputDir, inst.Installation.Name); err != nil {
					return err
				}
			}
		}
	}

	return nil
}

type Imports struct {
	Imports map[string]interface{} `json:"imports"`
}

func (o *RenderOptions) setupImports(fs vfs.FileSystem) (*Imports, error) {
	values := &Imports{
		Imports: make(map[string]interface{}),
	}

	if len(o.ValueFiles) == 0 {
		return values, nil
	}
	for _, filePath := range o.ValueFiles {
		data, err := vfs.ReadFile(fs, filePath)
		if err != nil {
			return nil, fmt.Errorf("unable to read values file '%s': %w", filePath, err)
		}
		tmpValues := &Imports{}
		if err := yaml.Unmarshal(data, tmpValues); err != nil {
			return nil, fmt.Errorf("unable to parse values file '%s': %w", filePath, err)
		}

		for key, val := range tmpValues.Imports {
			values.Imports[key] = val
		}
	}
	return values, nil
}

func (o *RenderOptions) Complete(log logr.Logger, args []string, fs vfs.FileSystem) error {
	ctx := context.Background()

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
		absCdPath, err := filepath.Abs(cdPath)
		if err != nil {
			return fmt.Errorf("unable get absolute component descriptor path for %s: %w", cdPath, err)
		}

		data, err := vfs.ReadFile(fs, absCdPath)
		if err != nil {
			return fmt.Errorf("unable to read component descriptor from %s: %w", o.ComponentDescriptorPath, err)
		}
		cd := cdv2.ComponentDescriptor{}
		if err := codec.Decode(data, &cd); err != nil {
			return fmt.Errorf("unable to decode component descriptor: %w", err)
		}
		o.componentDescriptorList.Components = append(o.componentDescriptorList.Components, cd)
	}

	if len(o.ResourcesPath) != 0 {
		absResourcesPath, err := filepath.Abs(o.ResourcesPath)
		if err != nil {
			return fmt.Errorf("unable get absolute resources path for %s: %w", o.ResourcesPath, err)
		}

		o.ResourcesPath = absResourcesPath

		resourceReader := components.NewResourceReader(o.ResourcesPath)
		o.resources, err = resourceReader.Read()
		if err != nil {
			return fmt.Errorf("unable to read resources from file %s: %w", o.ResourcesPath, err)
		}
	}

	if err := o.parseOutputResources(args); err != nil {
		return err
	}

	if len(o.resources) > 0 {
		if o.componentDescriptor == nil {
			return fmt.Errorf("if you specify a resources yaml file (option -r) you must also specify a component descriptor (option -c)")
		}

		o.componentDescriptor, err = resolver.AddLocalResourcesForRender(o.componentDescriptor, o.resources)

		if err != nil {
			return err
		}
	}

	if len(o.ExportTemplatesPath) != 0 {
		absExportTemplatesPath, err := filepath.Abs(o.ExportTemplatesPath)
		if err != nil {
			return fmt.Errorf("unable get absolute exports template path for %s: %w", o.ExportTemplatesPath, err)
		}

		o.ExportTemplatesPath = absExportTemplatesPath

		data, err := vfs.ReadFile(fs, o.ExportTemplatesPath)
		if err != nil {
			return fmt.Errorf("failed to read exports template path %s: %w", o.ExportTemplatesPath, err)
		}

		err = yaml.Unmarshal(data, &o.exportTemplates)
		if err != nil {
			return fmt.Errorf("failed to parse export templates: %w", err)
		}
	}

	// build component resolver with oci client
	//if o.componentDescriptor == nil {

	o.registryAccess, err = registries.NewFactory().NewRegistryAccessFromOciOptions(ctx, log, fs,
		o.OCIOptions.AllowPlainHttp, o.OCIOptions.SkipTLSVerify, o.OCIOptions.RegistryConfigPath, o.OCIOptions.ConcourseConfigPath)
	if err != nil {
		return err
	}

	//} else {
	//	o.registryAccess, err = registries.NewFactory().NewRegistryAccessFromOciOptions(ctx, log, fs,
	//		o.OCIOptions.AllowPlainHttp, o.OCIOptions.SkipTLSVerify, o.OCIOptions.RegistryConfigPath, o.OCIOptions.ConcourseConfigPath, o.componentDescriptor)
	//	if err != nil {
	//		return err
	//	}
	//}

	o.registryAccess = resolver.NewRenderRegistryAccess(o.registryAccess, o.componentDescriptor, o.componentDescriptorList, o.ResourcesPath, fs)

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
	allResources := sets.New(OutputResourceDeployItems, OutputResourceSubinstallations, OutputResourceImports, OutputResourceExports)
	if len(args) == 1 {
		o.outputResources = allResources
		return nil
	}
	if len(args) > 1 {
		resources := strings.Split(args[1], ",")
		o.outputResources = sets.New[string]()
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
