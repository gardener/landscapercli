## landscaper-cli installations interrupt

Interrupts the processing of an installations and its subobjects. All of these objects with an unfinished phase (i.e. a phase which is neither 'Succeeded' nor 'Failed' nor 'DeleteFailed') are changed to phase 'Failed'. Note that the command affects only the status of Landscaper objects, but does not interrupt a running installation process, for example a helm deployment.

```
landscaper-cli installations interrupt [installation-name] [--namespace namespace] [--kubeconfig kubeconfig.yaml] [flags]
```

### Examples

```
landscaper-cli installations interrupt MY_INSTALLATION --namespace MY_NAMESPACE
```

### Options

```
  -h, --help                help for interrupt
      --kubeconfig string   path to the kubeconfig for the cluster. Required if the cluster is not the same as the current-context of kubectl.
  -n, --namespace string    namespace of the installation. Required if --kubeconfig is used.
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

