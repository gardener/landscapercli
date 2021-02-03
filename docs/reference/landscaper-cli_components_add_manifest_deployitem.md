## landscaper-cli components add manifest deployitem


Command to add a deploy item skeleton to the blueprint of a component

```
landscaper-cli components add manifest deployitem \
    [component directory path] \
    [deployitem name] \
    [flags]
```

### Examples

```

landscaper-cli component add manifest deployitem \
  . \
  nginx \
  --file ./deployment.yaml \
  --file ./service.yaml \
  --import-param replicas:integer

```

### Options

```
      --cluster-param string       import parameter name for the target resource containing the access data of the target cluster (default "targetCluster")
      --file stringArray           manifest file
  -h, --help                       help for deployitem
      --import-param stringArray   import parameter
      --policy string              policy (default "manage")
      --target-ns-param string     target namespace
      --update-strategy string     update stategy (default "update")
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

* [landscaper-cli components add manifest](landscaper-cli_components_add_manifest.md)	 - command to add parts to a component concerning a manifest deployment

