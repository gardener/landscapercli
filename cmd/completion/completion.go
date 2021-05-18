// SPDX-FileCopyrightText: 2020 SAP SE or an SAP affiliate company and Gardener contributors.
//
// SPDX-License-Identifier: Apache-2.0

//go:generate go run -mod=vendor ../hack/generate-docs ../docs/reference

package completion

import (
	"os"

	"github.com/spf13/cobra"
)

var CompletionCmd = &cobra.Command{
	Use:   "completion [bash|zsh|fish|powershell]",
	Short: "Generate completion script",
	Long: `To load completions:

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
`,
	DisableFlagsInUseLine: true,
	ValidArgs:             []string{"bash", "zsh", "fish", "powershell"},
	Args:                  cobra.ExactValidArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		switch args[0] {
		case "bash":
			cmd.Root().GenBashCompletion(os.Stdout)
		case "zsh":
			cmd.Root().GenZshCompletion(os.Stdout)
		case "fish":
			cmd.Root().GenFishCompletion(os.Stdout, true)
		case "powershell":
			cmd.Root().GenPowerShellCompletion(os.Stdout)
		}
	},
}
