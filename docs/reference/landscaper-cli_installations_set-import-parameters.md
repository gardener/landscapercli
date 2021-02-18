## landscaper-cli installations set-import-parameters

Set import parameters for an installation. Quote values containing spaces in double quotation marks.

```
landscaper-cli installations set-import-parameters [flags]
```

### Examples

```
landscaper-cli installations set-import-parameters <path-to-installation>.yaml importName1="string value with spaces" importName2=42
```

### Options

```
  -h, --help                 help for set-import-parameters
  -o, --output-file string   file path for the resulting installation yaml (default: overwrite the given installation file)
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

* [landscaper-cli installations](landscaper-cli_installations.md)	 - commands to interact with installations

