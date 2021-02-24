## landscaper-cli quickstart install

command to install the landscaper (and optionally an OCI registry) in a target cluster

```
landscaper-cli quickstart install --kubeconfig [kubconfig.yaml] [--install-oci-registry] [flags]
```

### Examples

```
landscaper-cli quickstart install --kubeconfig ./kubconfig.yaml --install-oci-registry --landscaper-values ./landscaper-values.yaml --namespace landscaper
```

### Options

```
  -h, --help                              help for install
      --install-oci-registry              install an internal OCI registry in the target cluster
      --kubeconfig string                 path to the kubeconfig of the target cluster
      --landscaper-chart-version string   use a custom landscaper chart version (corresponds to landscaper github release with the same version number) (default "v0.7.0-dev-febf39db3ec93a69a1bbb1e100422ee2434f412f")
      --landscaper-values string          path to values.yaml for the landscaper Helm installation
      --namespace string                  namespace where the landscaper and the OCI registry are installed (default "landscaper")
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

* [landscaper-cli quickstart](landscaper-cli_quickstart.md)	 - useful commands for getting quickly up and running with the landscaper

