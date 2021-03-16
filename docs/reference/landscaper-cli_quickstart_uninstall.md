## landscaper-cli quickstart uninstall

command to uninstall Landscaper and OCI registry (from the install command) in a target cluster

```
landscaper-cli quickstart uninstall --kubeconfig [kubconfig.yaml] [flags]
```

### Examples

```
landscaper-cli quickstart uninstall --kubeconfig ./kubconfig.yaml --namespace landscaper
```

### Options

```
  -h, --help                help for uninstall
      --kubeconfig string   path to the kubeconfig of the target cluster
      --namespace string    namespace where the landscaper and the OCI registry are installed (default "landscaper")
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

* [landscaper-cli quickstart](landscaper-cli_quickstart.md)	 - useful commands for getting quickly up and running with Landscaper

