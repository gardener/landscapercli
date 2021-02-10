package installations

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"path/filepath"

	ociopts "github.com/gardener/component-cli/ociclient/options"
	"github.com/gardener/component-cli/pkg/commands/constants"
	cdv2 "github.com/gardener/component-spec/bindings-go/apis/v2"
	cdoci "github.com/gardener/component-spec/bindings-go/oci"
	lsv1alpha1 "github.com/gardener/landscaper/apis/core/v1alpha1"
	"github.com/gardener/landscaper/pkg/kubernetes"
	"github.com/gardener/landscaper/pkg/utils"
	"github.com/go-logr/logr"
	"github.com/mandelsoft/vfs/pkg/memoryfs"
	"github.com/mandelsoft/vfs/pkg/osfs"
	"github.com/mandelsoft/vfs/pkg/vfs"
	"github.com/spf13/cobra"
	"github.com/spf13/pflag"
	yamlv3 "gopkg.in/yaml.v3"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime/serializer"
	"sigs.k8s.io/yaml"

	"github.com/gardener/landscapercli/pkg/logger"
	"github.com/gardener/landscapercli/pkg/util"
)

type createOpts struct {
	// baseUrl is the oci registry where the component is stored.
	baseUrl string
	// componentName is the unique name of the component in the registry.
	componentName string
	// version is the component version in the oci registry.
	version string
	// OciOptions contains all exposed options to configure the oci client.
	OciOptions ociopts.Options

	name             string
	renderSchemaInfo bool
}

func NewCreateCommand(ctx context.Context) *cobra.Command {
	opts := &createOpts{}
	cmd := &cobra.Command{
		Use:     "create [baseURL] [componentName] [componentVersion]",
		Args:    cobra.ExactArgs(3),
		Aliases: []string{"c"},
		Example: "landscaper-cli installations create my-registry:5000 github.com/my-component v0.1.0",
		Short:   "create an installation template for a component which is stored in an OCI registry",
		Run: func(cmd *cobra.Command, args []string) {
			if err := opts.Complete(args); err != nil {
				cmd.PrintErr(err.Error())
				os.Exit(1)
			}

			if err := opts.run(ctx, cmd, logger.Log, osfs.New()); err != nil {
				cmd.PrintErr(err.Error())
				os.Exit(1)
			}
		},
	}

	opts.AddFlags(cmd.Flags())

	return cmd
}

func (o *createOpts) run(ctx context.Context, cmd *cobra.Command, log logr.Logger, fs vfs.FileSystem) error {
	repoCtx := cdv2.RepositoryContext{
		Type:    cdv2.OCIRegistryType,
		BaseURL: o.baseUrl,
	}
	ociRef, err := cdoci.OCIRef(repoCtx, o.componentName, o.version)
	if err != nil {
		return fmt.Errorf("invalid component reference: %w", err)
	}

	ociClient, _, err := o.OciOptions.Build(log, fs)
	if err != nil {
		return fmt.Errorf("unable to build oci client: %s", err.Error())
	}

	cdresolver := cdoci.NewResolver().WithOCIClient(ociClient).WithRepositoryContext(repoCtx)
	cd, blobResolver, err := cdresolver.Resolve(ctx, o.componentName, o.version)
	if err != nil {
		return fmt.Errorf("unable to to fetch component descriptor %s: %w", ociRef, err)
	}

	var blueprintRes cdv2.Resource
	for _, resource := range cd.ComponentSpec.Resources {
		if resource.IdentityObjectMeta.Type == lsv1alpha1.BlueprintResourceType || resource.IdentityObjectMeta.Type == lsv1alpha1.OldBlueprintType {
			blueprintRes = resource
		}
	}

	var data bytes.Buffer
	if blueprintRes.Access.GetType() == cdv2.OCIRegistryType {
		ref, ok := blueprintRes.Access.Object["imageReference"].(string)
		if !ok {
			return fmt.Errorf("cannot parse imageReference to string")
		}

		manifest, err := ociClient.GetManifest(ctx, ref)
		if err != nil {
			return fmt.Errorf("cannot get manifest: %w", err)
		}

		err = ociClient.Fetch(ctx, ref, manifest.Layers[0], &data)
		if err != nil {
			return fmt.Errorf("cannot get manifest layer: %w", err)
		}
	} else {
		_, err = blobResolver.Resolve(ctx, blueprintRes, &data)
		if err != nil {
			return fmt.Errorf("unable to to resolve blob of blueprint resource: %w", err)
		}
	}

	memFS := memoryfs.New()
	if err := utils.ExtractTarGzip(&data, memFS, "/"); err != nil {
		return fmt.Errorf("cannot extract blueprint blob: %w", err)
	}

	blueprintData, err := vfs.ReadFile(memFS, lsv1alpha1.BlueprintFileName)
	if err != nil {
		return fmt.Errorf("cannot read %s: %w", lsv1alpha1.BlueprintFileName, err)
	}

	blueprint := &lsv1alpha1.Blueprint{}
	if _, _, err := serializer.NewCodecFactory(kubernetes.LandscaperScheme).UniversalDecoder().Decode(blueprintData, nil, blueprint); err != nil {
		return fmt.Errorf("cannot decode blueprint: %w", err)
	}

	installation := buildInstallation(o.name, cd, blueprintRes, blueprint)

	var marshaledYaml []byte
	if o.renderSchemaInfo {
		commentedYaml, err := annotateInstallationWithSchemaComments(installation, blueprint)
		if err != nil {
			return fmt.Errorf("cannot add JSON schema comment: %w", err)
		}

		marshaledYaml, err = util.MarshalYaml(commentedYaml)
		if err != nil {
			return fmt.Errorf("cannot marshal installation yaml: %w", err)
		}
	} else {
		marshaledYaml, err = yaml.Marshal(installation)
		if err != nil {
			return fmt.Errorf("cannot marshal installation yaml: %w", err)
		}
	}

	cmd.Println(string(marshaledYaml))

	return nil
}

func (o *createOpts) Complete(args []string) error {
	o.baseUrl = args[0]
	o.componentName = args[1]
	o.version = args[2]

	cliHomeDir, err := constants.CliHomeDir()
	if err != nil {
		return err
	}
	o.OciOptions.CacheDir = filepath.Join(cliHomeDir, "components")
	if err := os.MkdirAll(o.OciOptions.CacheDir, os.ModePerm); err != nil {
		return fmt.Errorf("unable to create cache directory %s: %w", o.OciOptions.CacheDir, err)
	}

	if len(o.baseUrl) == 0 {
		return errors.New("the base url must be defined")
	}
	if len(o.componentName) == 0 {
		return errors.New("a component name must be defined")
	}
	if len(o.version) == 0 {
		return errors.New("a component's Version must be defined")
	}
	return nil
}

func annotateInstallationWithSchemaComments(installation *lsv1alpha1.Installation, blueprint *lsv1alpha1.Blueprint) (*yamlv3.Node, error) {
	out, err := yaml.Marshal(installation)
	if err != nil {
		return nil, fmt.Errorf("cannot marshal installation yaml: %w", err)
	}

	commentedInstallationYaml := &yamlv3.Node{}
	err = yamlv3.Unmarshal(out, commentedInstallationYaml)
	if err != nil {
		return nil, fmt.Errorf("cannot unmarshal installation yaml: %w", err)
	}

	err = addImportSchemaComments(commentedInstallationYaml, blueprint)
	if err != nil {
		return nil, fmt.Errorf("cannot add schema comments for imports: %w", err)
	}

	err = addExportSchemaComments(commentedInstallationYaml, blueprint)
	if err != nil {
		return nil, fmt.Errorf("cannot add schema comments for exports: %w", err)
	}

	return commentedInstallationYaml, nil
}

func addExportSchemaComments(commentedInstallationYaml *yamlv3.Node, blueprint *lsv1alpha1.Blueprint) error {
	_, exportsDataValueNode := util.FindNodeByPath(commentedInstallationYaml, "spec.exports.data")
	if exportsDataValueNode != nil {

		for _, dataImportNode := range exportsDataValueNode.Content {
			n1, n2 := util.FindNodeByPath(dataImportNode, "name")
			exportName := n2.Value

			var expdef lsv1alpha1.ExportDefinition
			for _, bpexp := range blueprint.Exports {
				if bpexp.Name == exportName {
					expdef = bpexp
					break
				}
			}
			prettySchema, err := json.MarshalIndent(expdef.Schema, "", "  ")
			if err != nil {
				return fmt.Errorf("cannot marshal JSON schema: %w", err)
			}
			n1.HeadComment = "JSON schema\n" + string(prettySchema)
		}
	}

	_, exportTargetsValueNode := util.FindNodeByPath(commentedInstallationYaml, "spec.exports.targets")
	if exportTargetsValueNode != nil {
		for _, targetExportNode := range exportTargetsValueNode.Content {
			n1, n2 := util.FindNodeByPath(targetExportNode, "name")
			targetName := n2.Value

			var expdef lsv1alpha1.ExportDefinition
			for _, bpexp := range blueprint.Exports {
				if bpexp.Name == targetName {
					expdef = bpexp
					break
				}
			}
			n1.HeadComment = "Target type: " + expdef.TargetType
		}
	}

	return nil
}

func addImportSchemaComments(commentedInstallationYaml *yamlv3.Node, blueprint *lsv1alpha1.Blueprint) error {
	_, importDataValueNode := util.FindNodeByPath(commentedInstallationYaml, "spec.imports.data")
	if importDataValueNode != nil {
		for _, dataImportNode := range importDataValueNode.Content {
			n1, n2 := util.FindNodeByPath(dataImportNode, "name")
			importName := n2.Value

			var impdef lsv1alpha1.ImportDefinition
			for _, bpimp := range blueprint.Imports {
				if bpimp.Name == importName {
					impdef = bpimp
					break
				}
			}
			prettySchema, err := json.MarshalIndent(impdef.Schema, "", "  ")
			if err != nil {
				return fmt.Errorf("cannot marshal JSON schema: %w", err)
			}
			n1.HeadComment = "JSON schema\n" + string(prettySchema)
		}
	}

	_, targetsValueNode := util.FindNodeByPath(commentedInstallationYaml, "spec.imports.targets")
	if targetsValueNode != nil {
		for _, targetImportNode := range targetsValueNode.Content {
			n1, n2 := util.FindNodeByPath(targetImportNode, "name")
			targetName := n2.Value

			var impdef lsv1alpha1.ImportDefinition
			for _, bpimp := range blueprint.Imports {
				if bpimp.Name == targetName {
					impdef = bpimp
					break
				}
			}
			n1.HeadComment = "Target type: " + impdef.TargetType
		}
	}

	return nil
}

func (o *createOpts) AddFlags(fs *pflag.FlagSet) {
	fs.StringVar(&o.name, "name", "my-installation", "name of the installation")
	fs.BoolVar(&o.renderSchemaInfo, "render-schema-info", false, "render schema information of the component's imports and exports as comments into the installation")
	o.OciOptions.AddFlags(fs)
}

func buildInstallation(name string, cd *cdv2.ComponentDescriptor, blueprintRes cdv2.Resource, blueprint *lsv1alpha1.Blueprint) *lsv1alpha1.Installation {
	dataImports := []lsv1alpha1.DataImport{}
	targetImports := []lsv1alpha1.TargetImportExport{}
	for _, imp := range blueprint.Imports {
		if imp.TargetType != "" {
			targetImport := lsv1alpha1.TargetImportExport{
				Name: imp.Name,
			}
			targetImports = append(targetImports, targetImport)
		} else {
			dataImport := lsv1alpha1.DataImport{
				Name: imp.Name,
			}
			dataImports = append(dataImports, dataImport)
		}
	}

	dataExports := []lsv1alpha1.DataExport{}
	targetExports := []lsv1alpha1.TargetImportExport{}
	for _, exp := range blueprint.Exports {
		if exp.TargetType != "" {
			targetExport := lsv1alpha1.TargetImportExport{
				Name: exp.Name,
			}
			targetExports = append(targetExports, targetExport)
		} else {
			dataExport := lsv1alpha1.DataExport{
				Name: exp.Name,
			}
			dataExports = append(dataExports, dataExport)
		}
	}

	obj := &lsv1alpha1.Installation{
		TypeMeta: metav1.TypeMeta{
			APIVersion: lsv1alpha1.SchemeGroupVersion.String(),
			Kind:       "Installation",
		},
		ObjectMeta: metav1.ObjectMeta{
			Name: name,
		},
		Spec: lsv1alpha1.InstallationSpec{
			ComponentDescriptor: &lsv1alpha1.ComponentDescriptorDefinition{
				Reference: &lsv1alpha1.ComponentDescriptorReference{
					RepositoryContext: &cd.RepositoryContexts[0],
					ComponentName:     cd.ObjectMeta.Name,
					Version:           cd.ObjectMeta.Version,
				},
			},
			Blueprint: lsv1alpha1.BlueprintDefinition{
				Reference: &lsv1alpha1.RemoteBlueprintReference{
					ResourceName: blueprintRes.Name,
				},
			},
			Imports: lsv1alpha1.InstallationImports{
				Data:    dataImports,
				Targets: targetImports,
			},
			Exports: lsv1alpha1.InstallationExports{
				Data:    dataExports,
				Targets: targetExports,
			},
		},
	}

	return obj
}
