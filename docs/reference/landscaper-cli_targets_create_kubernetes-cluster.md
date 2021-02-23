## landscaper-cli targets create kubernetes-cluster

create a target of type landscaper.gardener.cloud/kubernetes-cluster

```
landscaper-cli targets create kubernetes-cluster --name [name] --namespace [namespace] --target-kubeconfig [path to target kubeconfig] [flags]
```

### Examples

```
landscaper-cli targets create kubernetes-cluster --name my-target --namespace my-namespace --target-kubeconfig  kubeconfig.yaml
```

### Options

```
  -h, --help                       help for kubernetes-cluster
      --name string                name of the target
      --namespace string           namespace of the target
  -o, --output-file string         file path for the resulting target yaml
      --target-kubeconfig string   path to the kubeconfig where the created target object will point to
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

* [landscaper-cli targets create](landscaper-cli_targets_create.md)	 - command for creating different types of targets

