## landscaper-cli blueprints create

command to create a component template in the specified directory

```
landscaper-cli blueprints create [component directory path] [component name] [component version] [flags]
```

### Examples

```
landscaper-cli blueprints create \
    . \
    github.com/gardener/landscapercli/nginx \
    v0.1.0
```

### Options

```
  -h, --help   help for create
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

