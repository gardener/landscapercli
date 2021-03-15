## landscaper-cli components-cli component-archive sources add

Adds a source to a component descriptor

### Synopsis


add adds sources to the defined component descriptor.
The sources can be defined in a file or given through stdin.

The source definitions are expected to be a multidoc yaml of the following form

<pre>

---
name: 'myrepo'
type: 'git'
access:
  type: "git"
  repository: github.com/gardener/component-cli
...
---
name: 'myconfig'
type: 'json'
input:
  type: "file"
  path: "some/path"
...
---
name: 'myothersrc'
type: 'git'
input:
  type: "dir"
  path: /my/path
  compress: true # defaults to false
  exclude: "*.txt"
...

</pre>


Templating:
All yaml/json defined resources can be templated using simple envsubst syntax.
Variables are specified after a "--" and follow the syntax "<name>=<value>".

Note: Variable names are case-sensitive.

Example:
<pre>
<command> [args] [--flags] -- MY_VAL=test
</pre>

<pre>

key:
  subkey: "abc ${MY_VAL}"

</pre>




```
landscaper-cli components-cli component-archive sources add [component descriptor path] [source file]... [flags]
```

### Options

```
  -a, --archive string             path to the component archive directory
      --component-name string      name of the component
      --component-version string   version of the component
  -h, --help                       help for add
      --repo-ctx string            [OPTIONAL] repository context url for component to upload. The repository url will be automatically added to the repository contexts.
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

* [landscaper-cli components-cli component-archive sources](landscaper-cli_components-cli_component-archive_sources.md)	 - command to modify sources of a component descriptor

