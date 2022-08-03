## landscaper-cli installations reconcile

Starts a new reconciliation of the specified root installation. If the command is invoked while a reconciliation is already running, the new reconciliation is postponed until the current one has finished. The command is only supported for root installations.

```
landscaper-cli installations reconcile [installation-name] [--namespace namespace] [--kubeconfig kubeconfig.yaml] [flags]
```

### Examples

```
landscaper-cli installations reconcile MY_INSTALLATION --namespace MY_NAMESPACE
```

### Options

```
  -h, --help                help for reconcile
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

