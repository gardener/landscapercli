## landscaper-cli components add helm-ls deployitem


Command to add a deploy item skeleton to the blueprint of a component

```
landscaper-cli components add helm-ls deployitem \
    [component directory path] \
    [execution name] \
    [deployitem name] \
    [flags]
```

### Examples

```

landscaper-cli component add helm-ls deployitem \
  . \
  nginx \
  nginx \
  --oci-reference eu.gcr.io/gardener-project/landscaper/tutorials/charts/ingress-nginx:v0.1.0 \
  --chart-version v0.1.0
```

### Options

```
      --chart-directory string   path to chart directory
      --chart-version string     helm chart version
      --cluster-param string     target cluster (default "targetCluster")
  -h, --help                     help for deployitem
      --oci-reference string     reference to oci artifact containing the helm chart
      --target-ns-param string   target namespace
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

* [landscaper-cli components add helm-ls](landscaper-cli_components_add_helm-ls.md)	 - command to add parts to a component concerning a helm landscaper deployment

