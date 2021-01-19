// SPDX-FileCopyrightText: 2020 SAP SE or an SAP affiliate company and Gardener contributors.
//
// SPDX-License-Identifier: Apache-2.0

package componentarchive

import (
	"compress/gzip"
	"context"
	"fmt"
	"os"

	"github.com/gardener/component-spec/bindings-go/ctf"
	"github.com/mandelsoft/vfs/pkg/osfs"
	"github.com/mandelsoft/vfs/pkg/projectionfs"
	"github.com/mandelsoft/vfs/pkg/vfs"
	"github.com/spf13/cobra"
	"github.com/spf13/pflag"

	"github.com/gardener/component-cli/pkg/componentarchive"
	"github.com/gardener/component-cli/pkg/utils"
)

const defaultOutputPath = "./componentarchive"

// ExportOptions defines all options for the export command.
type ExportOptions struct {
	// ComponentArchivePath defines the path to the component archive
	ComponentArchivePath string
	// OutputPath defines the path where the exported component archive should be written to.
	OutputPath string
	// OutputFormat defines the output format of the component archive.
	OutputFormat componentarchive.OutputFormat
}

// NewExportCommand creates a new export command that packages a component archive and
// exports is as tar or compressed tar.
func NewExportCommand(ctx context.Context) *cobra.Command {
	opts := &ExportOptions{}
	cmd := &cobra.Command{
		Use:   "export [component-archive-path] [-o output-dir/file] [-f {fs|tar|tgz}]",
		Args:  cobra.ExactArgs(1),
		Short: "Exports a component archive as defined by CTF",
		Long: `
Export command exports a component archive as defined by CTF (CNUDIE Transport Format).
If the given component-archive path points to a directory, the archive is expected to be a extracted component-archive on the filesystem.
Then it is exported as tar or optionally as compressed tar.

If the given path points to a file, the archive is read as tar or compressed tar (tar.gz) and exported as filesystem to the given location.
`,
		Run: func(cmd *cobra.Command, args []string) {
			if err := opts.Complete(args); err != nil {
				fmt.Println(err.Error())
				os.Exit(1)
			}
			if err := opts.Run(ctx, osfs.New()); err != nil {
				fmt.Println(err.Error())
				os.Exit(1)
			}
			fmt.Printf("Successfully exported component archive to %s\n", opts.OutputPath)
		},
	}
	opts.AddFlags(cmd.Flags())
	return cmd
}

// Run runs the export for a component archive.
func (o *ExportOptions) Run(ctx context.Context, fs vfs.FileSystem) error {
	fileinfo, err := fs.Stat(o.ComponentArchivePath)
	if err != nil {
		return fmt.Errorf("unable to read %q: %s", o.ComponentArchivePath, err.Error())
	}

	if fileinfo.IsDir() {
		ca, err := o.caAsDir(fs)
		if err != nil {
			return err
		}
		return o.export(fs, ca, componentarchive.OutputFormatTar)
	} else {
		ca, err := o.caAsFile(fs)
		if err != nil {
			return nil
		}
		if err := ca.WriteToFilesystem(fs, o.OutputPath); err != nil {
			return fmt.Errorf("unable to write componant archive to %q: %s", o.OutputPath, err.Error())
		}
		return o.export(fs, ca, componentarchive.OutputFormatFilesystem)
	}
}

// caAsDir imports the given component archive as filesystem and outputs it as tar.
func (o *ExportOptions) caAsDir(fs vfs.FileSystem) (*ctf.ComponentArchive, error) {
	archiveFs, err := projectionfs.New(fs, o.ComponentArchivePath)
	if err != nil {
		return nil, fmt.Errorf("unable to create filesystem from %s: %s", o.ComponentArchivePath, err.Error())
	}
	ca, err := ctf.NewComponentArchiveFromFilesystem(archiveFs)
	if err != nil {
		return nil, fmt.Errorf("unable to parse component archive from %s: %s", o.ComponentArchivePath, err.Error())
	}
	return ca, nil
}

// caAsFile imports the given component archive as tar and outputs it as filesystem.
func (o *ExportOptions) caAsFile(fs vfs.FileSystem) (*ctf.ComponentArchive, error) {
	mimetype, err := utils.GetFileType(fs, o.ComponentArchivePath)
	if err != nil {
		return nil, fmt.Errorf("unable to get mimetype of %q: %s", o.ComponentArchivePath, err.Error())
	}
	file, err := fs.Open(o.ComponentArchivePath)
	if err != nil {
		return nil, fmt.Errorf("unable to read component archive rom %q: %s", o.ComponentArchivePath, err.Error())
	}

	switch mimetype {
	case "application/x-gzip":
		zr, err := gzip.NewReader(file)
		if err != nil {
			return nil, fmt.Errorf("unable to open gzip reader: %w", err)
		}
		ca, err := ctf.NewComponentArchiveFromTarReader(zr)
		if err != nil {
			return nil, fmt.Errorf("unable to unzip componentarchive: %s", err.Error())
		}
		if err := zr.Close(); err != nil {
			return nil, fmt.Errorf("unable to close gzip reader: %w", err)
		}
		if err := file.Close(); err != nil {
			return nil, fmt.Errorf("unable to close file reader: %w", err)
		}
		return ca, nil
	case "application/octet-stream": // expect that is has to be a tar
		ca, err := ctf.NewComponentArchiveFromTarReader(file)
		if err != nil {
			return nil, fmt.Errorf("unable to unzip componentarchive: %s", err.Error())
		}
		if err := file.Close(); err != nil {
			return nil, fmt.Errorf("unable to close file reader: %w", err)
		}
		return ca, nil
	default:
		return nil, fmt.Errorf("unsupported file type %q. Expected a tar or a tar.gz", mimetype)
	}
}

func (o *ExportOptions) export(fs vfs.FileSystem, ca *ctf.ComponentArchive, defaultFormat componentarchive.OutputFormat) error {
	if len(o.OutputFormat) == 0 {
		o.OutputFormat = defaultFormat
	}

	return componentarchive.Write(fs, o.OutputPath, ca, o.OutputFormat)
}

// Complete parses the given command arguments and applies default options.
func (o *ExportOptions) Complete(args []string) error {
	if len(args) != 1 {
		return fmt.Errorf("expected exactly one argument that contains the path to the component archive")
	}
	o.ComponentArchivePath = args[0]

	if len(o.OutputPath) == 0 {
		o.OutputPath = defaultOutputPath
	}

	return o.validate()
}

func (o *ExportOptions) validate() error {
	return componentarchive.ValidateOutputFormat(o.OutputFormat, true)
}

func (o *ExportOptions) AddFlags(fs *pflag.FlagSet) {
	fs.StringVarP(&o.OutputPath, "out", "o", "", "writes the resulting archive to the given path")
	componentarchive.OutputFormatVarP(fs, &o.OutputFormat, "format", "f", "", componentarchive.DefaultOutputFormatUsage)
}
