package installations

import (
	"context"
	"fmt"
	"os"
	"strconv"
	"strings"

	lsv1alpha1 "github.com/gardener/landscaper/apis/core/v1alpha1"
	"github.com/gardener/landscaper/pkg/api"
	"k8s.io/apimachinery/pkg/runtime/serializer"
	"sigs.k8s.io/yaml"

	"github.com/go-logr/logr"
	"github.com/spf13/cobra"
	"github.com/spf13/pflag"

	"github.com/gardener/landscapercli/pkg/logger"
)

type importParametersOptions struct {
	installationPath string

	//input parameters that should be used for the import values
	importParameters map[string]string

	//outputPath is the path to write the installation.yaml to
	outputPath string
}

// NewSetImportParametersCommand sets input parameters from an installation to hardcoded values (as importDataMappings)
func NewSetImportParametersCommand(ctx context.Context) *cobra.Command {
	opts := &importParametersOptions{}
	cmd := &cobra.Command{
		Use:     "set-import-parameters [path to installation.yaml] [key1=value1] [key2=value2]",
		Aliases: []string{"sip"},
		Short:   "Set import parameters for an installation. Quote values containing spaces in double quotation marks.",
		Example: `landscaper-cli installations set-import-parameters <path-to-installation>.yaml importName1="string value with spaces" importName2=42`,
		Args:    cobra.MinimumNArgs(2),
		Run: func(cmd *cobra.Command, args []string) {
			if err := opts.validateArguments(args); err != nil {
				cmd.PrintErr(err.Error())
				os.Exit(1)
			}

			if err := opts.run(ctx, logger.Log, cmd); err != nil {
				cmd.PrintErr(err.Error())
				os.Exit(1)
			}
		},
	}
	cmd.SetOut(os.Stdout)

	opts.AddFlags(cmd.Flags())

	return cmd
}

func (o *importParametersOptions) validateArguments(args []string) error {
	o.installationPath = args[0]

	o.importParameters = make(map[string]string)
	for _, v := range args[1:] {
		keyValue := strings.SplitN(v, "=", 2)
		if len(keyValue) != 2 {
			return fmt.Errorf("cannot split the import parameter %s at the = character.\nDid you enquote the value if it contains spaces?", keyValue)
		}
		o.importParameters[keyValue[0]] = keyValue[1]
	}
	return nil
}

func (o *importParametersOptions) run(ctx context.Context, log logr.Logger, cmd *cobra.Command) error {
	installation := lsv1alpha1.Installation{}

	err := readInstallationFromFile(o, &installation)
	if err != nil {
		return err
	}

	err = replaceImportsWithImportParameters(&installation, o)
	if err != nil {
		return fmt.Errorf("error setting the import parameters: %w", err)
	}

	marshaledYaml, err := yaml.Marshal(installation)
	if err != nil {
		return fmt.Errorf("cannot marshal yaml: %w", err)
	}
	outputPath := o.installationPath
	if o.outputPath != "" {
		outputPath = o.outputPath
	}
	f, err := os.Create(outputPath)
	if err != nil {
		return fmt.Errorf("error creating file %s: %w", outputPath, err)
	}
	_, err = f.Write(marshaledYaml)
	if err != nil {
		return fmt.Errorf("error writing file %s: %w", outputPath, err)
	}

	cmd.Printf("Wrote installation to %s", o.outputPath)
	return nil
}

func (o *importParametersOptions) AddFlags(fs *pflag.FlagSet) {
	fs.StringVarP(&o.outputPath, "output-file", "o", "", "file path for the resulting installation yaml (default: overwrite the given installation file)")
}

func replaceImportsWithImportParameters(installation *lsv1alpha1.Installation, o *importParametersOptions) error {
	validImportDataMappings := make(map[string]lsv1alpha1.AnyJSON)

	//find all imports.data that are specified in importParameters
	for _, importData := range installation.Spec.Imports.Data {
		if importParameter, ok := o.importParameters[importData.Name]; ok {
			validImportDataMappings[importData.Name] = createJSONRawMessageValueWithStringOrNumericType(importParameter)
		}
	}

	//check for any not used importParameters
	for k := range o.importParameters {
		if _, ok := validImportDataMappings[k]; !ok {
			return fmt.Errorf(`import parameter '%s' not found in the installation`, k)
		}
	}

	//modify the installation
	for importName, importDataMappingValue := range validImportDataMappings {
		if installation.Spec.ImportDataMappings == nil {
			installation.Spec.ImportDataMappings = make(map[string]lsv1alpha1.AnyJSON)
		}
		//add to importDataMappings
		installation.Spec.ImportDataMappings[importName] = importDataMappingValue

		//remove from imports.data
		for i, importData := range installation.Spec.Imports.Data {
			if importData.Name == importName {
				installation.Spec.Imports.Data = append(installation.Spec.Imports.Data[:i], installation.Spec.Imports.Data[i+1:]...)
			}
		}
	}

	return nil
}

func createJSONRawMessageValueWithStringOrNumericType(parameter string) lsv1alpha1.AnyJSON {
	if _, err := strconv.ParseFloat(parameter, 64); err == nil {
		return lsv1alpha1.AnyJSON{RawMessage: []byte(parameter)}
	}
	return lsv1alpha1.AnyJSON{RawMessage: []byte(fmt.Sprintf(`"%s"`, parameter))}

}

func readInstallationFromFile(o *importParametersOptions, installation *lsv1alpha1.Installation) error {
	installationFileData, err := os.ReadFile(o.installationPath)
	if err != nil {
		return err
	}
	if _, _, err := serializer.NewCodecFactory(api.LandscaperScheme).UniversalDecoder().Decode(installationFileData, nil, installation); err != nil {
		return err
	}
	return nil
}
