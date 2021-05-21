## landscaper-cli components-cli ctf add

Adds component archives to a ctf

```
landscaper-cli components-cli ctf add CTF_PATH [-f component-archive]... [flags]
```

### Options

```
  -f, --component-archive stringArray   path to the component archives to be added. Note that the component archives have to be tar archives.
      --format CAOutputFormat           archive format of the component archive. Can be "tar" or "tgz" (default tar)
  -h, --help                            help for add
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

* [landscaper-cli components-cli ctf](landscaper-cli_components-cli_ctf.md)	 - 

