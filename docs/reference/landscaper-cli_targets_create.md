## landscaper-cli targets create

command for creating different types of targets

### Options

```
  -h, --help                 help for create
      --name string          name of the target (required)
  -n, --namespace string     namespace of the target
  -o, --output-file string   file path for the resulting target yaml, leave empty for stdout
  -s, --secret string        name of the secret to store the target's content in (content will be stored in target spec directly, if empty)
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

* [landscaper-cli targets](landscaper-cli_targets.md)	 - commands for interacting with targets
* [landscaper-cli targets create kubernetes-cluster](landscaper-cli_targets_create_kubernetes-cluster.md)	 - create a target of type landscaper.gardener.cloud/kubernetes-cluster

