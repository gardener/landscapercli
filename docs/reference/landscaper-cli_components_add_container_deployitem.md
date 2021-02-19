## landscaper-cli components add container deployitem


Command to add a deploy item skeleton to the blueprint of a component

```
landscaper-cli components add container deployitem \
    [deployitem name] \
 [flags]
```

### Examples

```

landscaper-cli component add container deployitem \
  myDeployItem \
  --resource-version v0.1.0 \
  --component-directory ~/myComponent \
  --image alpine \
  --command sh,-c \
  --args env,ls \
  --import-param replicas:integer \
  --export-param message:string \
  --cluster-param target-cluster \

```

### Options

```
  -a, --args strings                 arguments (optional, multi-value)
      --cluster-param string         import parameter name for the target resource containing the access data of the target cluster (optional)
  -c, --command strings              command (multi-value)
      --component-directory string   path to component directory (optional, default is current directory) (default ".")
  -e, --export-param strings         export parameter as name:integer|string|boolean, e.g. replicas:integer
  -h, --help                         help for deployitem
      --image string                 image
  -i, --import-param strings         import parameter as name:integer|string|boolean, e.g. replicas:integer
      --resource-version string      resource version
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

* [landscaper-cli components add container](landscaper-cli_components_add_container.md)	 - command to add parts to a component concerning a container deployment

