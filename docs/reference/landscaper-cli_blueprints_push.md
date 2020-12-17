## landscaper-cli blueprints push

command to interact with definitions of an oci registry

```
landscaper-cli blueprints push [flags]
```

### Examples

```
landscapercli blueprints push [ref] [path to Blueprint directory]
```

### Options

```
      --allow-plain-http   allows the fallback to http if the oci registry does not support https
  -h, --help               help for push
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

