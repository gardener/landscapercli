## landscaper-cli blueprints get

command to download a blueprint from an oci registry

### Synopsis

The get command downloads a Blueprint from an OCI registry. The reference to the OCI artifact consists of the base URL of the OCI registry, the repository (namespace), and the tag.

```
landscaper-cli blueprints get [ref to OCI artifact] [flags]
```

### Examples

```
landscapercli blueprints get my-registry/my-repository:v1.0.0
```

### Options

```
      --allow-plain-http   allows the fallback to http if the oci registry does not support https
  -h, --help               help for get
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

