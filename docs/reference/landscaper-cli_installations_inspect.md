## landscaper-cli installations inspect

Displays status information for all installations and depending executions and deployItems in cluster and namespace of the current kubectl cluster context. To display only one installation, specify the installation-name.

```
landscaper-cli installations inspect [installation-name] [--namespace namespace] [--kubeconfig kubeconfig.yaml] [flags]
```

### Examples

```
landscaper-cli installations inspect
```

### Options

```
  -h, --help                help for inspect
      --kubeconfig string   path to the kubeconfig for the cluster. Required if the cluster is not the same as the current-context of kubectl.
  -n, --namespace string    namespace of the installation. Required if --kubeconfig is used.
  -j, --ojson               output in json format.
  -y, --oyaml               output in yaml format.
  -d, --show-details        show detailed information about installations, executions and deployitems. Similar to kubectl describe installation installation-name.
  -e, --show-executions     show the executions in the tree. By default, the executions are not shown.
  -f, --show-failed         show only items that are in phase 'Failed'. It also prints parent elements to the failed items.
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

