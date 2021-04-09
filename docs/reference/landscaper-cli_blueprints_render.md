## landscaper-cli blueprints render

renders the given blueprint

### Synopsis


Renders the blueprint with the given values files.
All value files are merged whereas the later defined will overwrite the values of the previous ones

By default all rendered resources are printed to stdout.
Specific resources can be printed by adding a second argument.
landscapercli local render [path to Blueprint directory] [resource]
Available resources are
- all: renders all available resources
- deployitems|di: renders deployitems
- inst|subinst|subinstallations: renders subinstallations


```
landscaper-cli blueprints render [flags]
```

### Examples

```
landscaper-cli blueprints render BLUEPRINT_DIR [all,deployitems,subinstallations]
```

### Options

```
  -a, --additional-component-descriptor stringArray   Path to additional local component descriptors
      --allow-plain-http                              allows the fallback to http if the oci registry does not support https
      --cc-config string                              path to the local concourse config file
  -c, --component-descriptor string                   Path to the local component descriptor
  -f, --file stringArray                              List of filepaths to value yaml files that define the imports
  -h, --help                                          help for render
  -o, --output string                                 The format of the output. Can be json or yaml. (default "yaml")
      --registry-config string                        path to the dockerconfig.json with the oci registry authentication information
  -w, --write string                                  The output directory where the rendered files should be written to
```

### Options inherited from parent commands

```
      --cli                  logger runs as cli logger. enables cli logging
      --dev                  enable development logging which result in console encoding, enabled stacktrace and enabled caller
      --disable-caller       disable the caller of logs (default true)
      --disable-stacktrace   disable the stacktrace of error logs (default true)
      --disable-timestamp    disable timestamp output (default true)
  -v, --verbosity int        number for the log level verbosity (default 1)
```

### SEE ALSO

* [landscaper-cli blueprints](landscaper-cli_blueprints.md)	 - command to interact with blueprints stored in an oci registry

