## landscaper-cli completion

Generate completion script

### Synopsis

To load completions:

Bash:

  $ source <(landscaper-cli completion bash)

  # To load completions for each session, execute once:
  # Linux:
  $ landscaper-cli completion bash > /etc/bash_completion.d/landscaper-cli
  # macOS:
  $ landscaper-cli completion bash > /usr/local/etc/bash_completion.d/landscaper-cli

Zsh:

  # If shell completion is not already enabled in your environment,
  # you will need to enable it.  You can execute the following once:

  $ echo "autoload -U compinit; compinit" >> ~/.zshrc

  # To load completions for each session, execute once:
  $ landscaper-cli completion zsh > "${fpath[1]}/_landscaper-cli"

  # You will need to start a new shell for this setup to take effect.

fish:

  $ landscaper-cli completion fish | source

  # To load completions for each session, execute once:
  $ landscaper-cli completion fish > ~/.config/fish/completions/landscaper-cli.fish

PowerShell:

  PS> landscaper-cli completion powershell | Out-String | Invoke-Expression

  # To load completions for every new session, run:
  PS> landscaper-cli completion powershell > landscaper-cli.ps1
  # and source this file from your PowerShell profile.


```
landscaper-cli completion [bash|zsh|fish|powershell]
```

### Options

```
  -h, --help   help for completion
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

* [landscaper-cli](landscaper-cli.md)	 - landscaper cli

