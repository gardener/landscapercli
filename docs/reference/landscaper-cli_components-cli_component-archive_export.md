## landscaper-cli components-cli component-archive export

Exports a component archive as defined by CTF

### Synopsis


Export command exports a component archive as defined by CTF (CNUDIE Transport Format).
If the given component-archive path points to a directory, the archive is expected to be a extracted component-archive on the filesystem.
Then it is exported as tar or optionally as compressed tar.

If the given path points to a file, the archive is read as tar or compressed tar (tar.gz) and exported as filesystem to the given location.


```
landscaper-cli components-cli component-archive export [component-archive-path] [-o output-dir/file] [-f {fs|tar|tgz}] [flags]
```

### Options

```
  -f, --format CAOutputFormat   output format of the component archive. Can be "fs"', "tar"' or "tgz"'
  -h, --help                    help for export
  -o, --out string              writes the resulting archive to the given path
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

* [landscaper-cli components-cli component-archive](landscaper-cli_components-cli_component-archive.md)	 - 

