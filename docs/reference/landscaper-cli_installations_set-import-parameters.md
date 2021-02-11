## landscaper-cli installations set-import-parameters

set import parameters for an installation. Enquote string values in double quotation marks.

```
landscaper-cli installations set-import-parameters [flags]
```

### Examples

```
landscapercli installation set-input-parameters <path-to-installation>.yaml importName1="string-value" importName2=42
```

### Options

```
  -h, --help   help for set-import-parameters
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

