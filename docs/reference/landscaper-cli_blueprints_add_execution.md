## landscaper-cli blueprints add execution

command to add a deploy execution skeleton to the blueprint in the specified directory

```
landscaper-cli blueprints add execution [path to Blueprint directory] [name] [flags]
```

### Examples

```
landscaper-cli blueprints add execution path/to/blueprint/directory default
```

### Options

```
  -h, --help   help for execution
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

* [landscaper-cli blueprints add](landscaper-cli_blueprints_add.md)	 - command to add parts to a blueprint

