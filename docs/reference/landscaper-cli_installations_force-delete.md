## landscaper-cli installations force-delete

Deletes an installations and the depending executions and deployItems in cluster and namespace of the current kubectl cluster context. Concerning the deployed software no guarantees could be given if it is uninstalled or not.

```
landscaper-cli installations force-delete [installation-name] [--namespace namespace] [--kubeconfig kubeconfig.yaml] [flags]
```

### Examples

```
landscaper-cli installations force-delete
```

### Options

```
  -h, --help                help for force-delete
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

