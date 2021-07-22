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
	"github.com/gardener/landscaper/pkg/landscaper/installations/executions/template/gotemplate"
	"github.com/gardener/landscaper/pkg/landscaper/installations/executions/template/spiff"
	"github.com/gardener/landscaper/pkg/landscaper/jsonschema"
	componentsregistry "github.com/gardener/landscaper/pkg/landscaper/registry/components"
	"github.com/go-logr/logr"
	"github.com/mandelsoft/vfs/pkg/osfs"
	"github.com/mandelsoft/vfs/pkg/projectionfs"
	"github.com/mandelsoft/vfs/pkg/vfs"
	"github.com/spf13/cobra"
	"github.com/spf13/pflag"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime/serializer"
	"k8s.io/apimachinery/pkg/util/sets"
	"k8s.io/apimachinery/pkg/util/validation/field"
	"sigs.k8s.io/yaml"

	"github.com/gardener/landscaper/apis/core"
	lsv1alpha1 "github.com/gardener/landscaper/apis/core/v1alpha1"
	"github.com/gardener/landscaper/apis/core/validation"
	"github.com/gardener/landscaper/pkg/api"
	"github.com/gardener/landscaper/pkg/landscaper/blueprints"
	"github.com/gardener/landscaper/pkg/landscaper/execution"
	"github.com/gardener/landscaper/pkg/landscaper/installations/executions/template"
	"github.com/gardener/landscaper/pkg/landscaper/installations/subinstallations"
	"github.com/gardener/landscaper/pkg/landscaper/registry/components/cdutils"

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
	values                  Values
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

	blueprint := blueprints.New(o.blueprint, o.blueprintFs)

	if err := o.validateImports(blueprint); err != nil {
		return err
	}

	exampleInstallation := &lsv1alpha1.Installation{}
	exampleInstallation.Spec.Blueprint.Reference = &lsv1alpha1.RemoteBlueprintReference{
		ResourceName: "example-blueprint",
	}
	repoCtx, err := cdv2.NewUnstructured(cdv2.NewOCIRegistryRepository("example.com/components", ""))
	if err != nil {
		return fmt.Errorf("unable to construct repository context: %w", err)
	}
	exampleInstallation.Spec.ComponentDescriptor = &lsv1alpha1.ComponentDescriptorDefinition{
		Reference: &lsv1alpha1.ComponentDescriptorReference{
			RepositoryContext: &repoCtx,
			ComponentName:     "my-example-component",
			Version:           "v0.0.0",
		},
	}

	var blobResolver ctf.BlobResolver
	if o.componentDescriptor != nil {
		exampleInstallation.Spec.ComponentDescriptor.Reference.ComponentName = o.componentDescriptor.GetName()
		exampleInstallation.Spec.ComponentDescriptor.Reference.Version = o.componentDescriptor.GetVersion()
		if len(o.componentDescriptor.RepositoryContexts) != 0 {
			repoCtx := o.componentDescriptor.GetEffectiveRepositoryContext()
			exampleInstallation.Spec.ComponentDescriptor.Reference.RepositoryContext = repoCtx
			o.componentDescriptor, blobResolver, err = o.componentResolver.ResolveWithBlobResolver(ctx,
				o.componentDescriptor.GetEffectiveRepositoryContext(),
				o.componentDescriptor.GetName(),
				o.componentDescriptor.GetVersion())
			if err != nil {
				return fmt.Errorf("unable to resolve component descriptor %s:%s: %w ",
					o.componentDescriptor.GetName(), o.componentDescriptor.GetVersion(), err)
			}
		}
	}

	if o.outputResources.Has(OutputResourceDeployItems) {
		templateStateHandler := template.NewMemoryStateHandler()
		deployItemTemplates, err := template.New(
			gotemplate.New(blobResolver, templateStateHandler),
			spiff.New(templateStateHandler),
		).TemplateDeployExecutions(template.DeployExecutionOptions{
			Imports:              o.values.Imports,
			Blueprint:            blueprint,
			ComponentDescriptor:  o.componentDescriptor,
			ComponentDescriptors: &cdv2.ComponentDescriptorList{},
			Installation:         exampleInstallation,
		})
		if err != nil {
			return fmt.Errorf("unable to template deploy executions: %w", err)
		}

		// print out state
		stateOut := map[string]map[string]json.RawMessage{
			"state": {},
		}
		for key, state := range templateStateHandler {
			stateOut["state"][key] = state
		}
		if err := o.out(fs, stateOut, "state"); err != nil {
			return err
		}

		for _, diTmpl := range deployItemTemplates {
			di := &lsv1alpha1.DeployItem{
				TypeMeta: metav1.TypeMeta{
					APIVersion: lsv1alpha1.SchemeGroupVersion.String(),
					Kind:       "DeployItem",
				},
			}
			execution.ApplyDeployItemTemplate(di, diTmpl)
			if err := o.out(fs, di, DeployItemOutputDir, diTmpl.Name); err != nil {
				return err
			}
		}
	}

	if o.outputResources.Has(OutputResourceSubinstallations) {
		dummyInst := &lsv1alpha1.Installation{
			TypeMeta: metav1.TypeMeta{
				APIVersion: lsv1alpha1.SchemeGroupVersion.String(),
				Kind:       "Installation",
			},
		}
		if len(blueprint.Info.Subinstallations) == 0 {
			fmt.Printf("No subinstallations defined\n")
		}
		for _, subInstTmpl := range blueprint.Info.Subinstallations {
			subInst := &lsv1alpha1.Installation{}
			subInst.Spec = lsv1alpha1.InstallationSpec{
				Imports:            subInstTmpl.Imports,
				ImportDataMappings: subInstTmpl.ImportDataMappings,
				Exports:            subInstTmpl.Exports,
				ExportDataMappings: subInstTmpl.ExportDataMappings,
			}
			subBlueprint, _, err := subinstallations.GetBlueprintDefinitionFromInstallationTemplate(
				dummyInst,
				subInstTmpl.InstallationTemplate,
				o.componentDescriptor,
				cdutils.ComponentReferenceResolverFromList(o.componentDescriptorList))
			if err != nil {
				fmt.Printf("unable to get blueprint: %s\n", err.Error())
			} else if subBlueprint != nil {
				subInst.Spec.Blueprint = *subBlueprint
			}
			if err := o.out(fs, subInst, "subinstallations", subInstTmpl.Name); err != nil {
				return err
			}
		}
	}

	return nil
}

// validateImports validates the type of all imports.
// todo(schrodit): use central landscaper validation function as soon as the new landscaper version is released
func (o *RenderOptions) validateImports(bp *blueprints.Blueprint) error {
	jsonschemaValidator := &jsonschema.Validator{
		Config: &jsonschema.LoaderConfig{
			LocalTypes:                 bp.Info.LocalTypes,
			BlueprintFs:                bp.Fs,
			ComponentDescriptor:        o.componentDescriptor,
			ComponentResolver:          o.componentResolver,
			ComponentReferenceResolver: cdutils.ComponentReferenceResolverFromList(o.componentDescriptorList),
		},
	}

	var allErr field.ErrorList
	for _, importDef := range bp.Info.Imports {
		fldPath := field.NewPath(importDef.Name)
		value, ok := o.values.Imports[importDef.Name]
		if !ok {
			if *importDef.Required {
				allErr = append(allErr, field.Required(fldPath, "Import is required"))
			}
			continue
		}
		if importDef.Schema != nil {
			if err := jsonschemaValidator.ValidateGoStruct(importDef.Schema.RawMessage, value); err != nil {
				allErr = append(allErr, field.Invalid(
					fldPath,
					value,
					fmt.Sprintf("invalid imported value: %s", err.Error())))
			}
		} else {
			// import is a target import
			targetObj, ok := value.(map[string]interface{})
			if !ok {
				allErr = append(allErr, field.Invalid(fldPath, value, "a target is expected to be an object"))
				continue
			}
			targetType, _, err := unstructured.NestedString(targetObj, "spec", "type")
			if err != nil {
				allErr = append(allErr, field.Invalid(
					fldPath,
					value,
					fmt.Sprintf("unable to get type of target: %s", err.Error())))
				continue
			}
			if targetType != importDef.TargetType {
				allErr = append(allErr, field.Invalid(
					fldPath,
					targetType,
					fmt.Sprintf("expected target type to be %q but got %q", importDef.TargetType, targetType)))
				continue
			}
		}
	}

	return allErr.ToAggregate()
}

func (o *RenderOptions) Complete(log logr.Logger, args []string, fs vfs.FileSystem) error {
	o.BlueprintPath = args[0]
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

	o.values = Values{}
	for _, filePath := range o.ValueFiles {
		data, err := vfs.ReadFile(fs, filePath)
		if err != nil {
			return fmt.Errorf("unable to read values file '%s': %w", filePath, err)
		}
		values := &Values{}
		if err := yaml.Unmarshal(data, values); err != nil {
			return fmt.Errorf("unable to parse values file '%s': %w", filePath, err)
		}

		MergeValues(&o.values, values)
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

	if err := ValidateValues(o.values); err != nil {
		return err
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
