## landscaper-cli installations inspect

displays status information for Installations and depending Executions and DeployItems

```
landscaper-cli installations inspect [installationName] [--namespace namespace] [--kubeconfig kubeconfig.yaml] [flags]
```

### Examples

```
landscaper-cli installations inspect
```

### Options

```
  -h, --help                help for inspect
      --kubeconfig string   path to the kubeconfig of the cluster
  -n, --namespace string    namespace of the installation. Required if --kubeconfig is used.
  -j, --ojson               output in json format
  -y, --oyaml               output in yaml format
  -d, --show-details        show detailed information about installations, executions and deployitems
  -e, --show-executions     show the executions in the tree
  -f, --show-failed         show only failed items
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

