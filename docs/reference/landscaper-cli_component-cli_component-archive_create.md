## landscaper-cli component-cli component-archive create

Creates a component archive with a component descriptor

### Synopsis


Create command creates a new component archive directory with a "component-descriptor.yaml" file.


```
landscaper-cli component-cli component-archive create COMPONENT_ARCHIVE_PATH [flags]
```

### Options

```
  -a, --archive string                  path to the component archive directory
      --component-name string           name of the component
      --component-name-mapping string   [OPTIONAL] repository context name mapping (default "urlPath")
      --component-version string        version of the component
  -h, --help                            help for create
  -w, --overwrite                       overwrites the existing component
      --repo-ctx string                 [OPTIONAL] repository context url for component to upload. The repository url will be automatically added to the repository contexts.
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

* [landscaper-cli component-cli component-archive](landscaper-cli_component-cli_component-archive.md)	 - 

