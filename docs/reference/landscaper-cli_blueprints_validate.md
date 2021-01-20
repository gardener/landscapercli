## landscaper-cli blueprints validate

validates a local blueprint filesystem

### Synopsis

The validate command validates a Blueprint in a local directory. The blueprint directory must contain a file with name blueprint.yaml.

```
landscaper-cli blueprints validate [path to Blueprint directory] [flags]
```

### Examples

```
landscaper-cli blueprints validate path/to/blueprint/directory
```

### Options

```
  -h, --help   help for validate
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

