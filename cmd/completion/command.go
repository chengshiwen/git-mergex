package completion

import (
	"os"

	"github.com/spf13/cobra"
)

const longDesc = `To load completions:

Bash:

  $ source <(git-mergex completion bash)

  # To load completions for each session, execute once:
  # Linux:
  $ git-mergex completion bash > /etc/bash_completion.d/git-mergex
  # macOS:
  $ git-mergex completion bash > /usr/local/etc/bash_completion.d/git-mergex

Zsh:

  # If shell completion is not already enabled in your environment,
  # you will need to enable it.  You can execute the following once:

  $ echo "autoload -U compinit; compinit" >> ~/.zshrc

  # To load completions for each session, execute once:
  $ git-mergex completion zsh > "${fpath[1]}/_git-mergex"

  # You will need to start a new shell for this setup to take effect.

fish:

  $ git-mergex completion fish | source

  # To load completions for each session, execute once:
  $ git-mergex completion fish > ~/.config/fish/completions/git-mergex.fish

PowerShell:

  PS> git-mergex completion powershell | Out-String | Invoke-Expression

  # To load completions for every new session, run:
  PS> git-mergex completion powershell > git-mergex.ps1
  # and source this file from your PowerShell profile.
`

func NewCommand() *cobra.Command {
	cobraCmd := &cobra.Command{
		Use:                   "completion [bash|zsh|fish|powershell]",
		Short:                 "Generate completion script",
		Long:                  longDesc,
		DisableFlagsInUseLine: true,
		ValidArgsFunction: func(cmd *cobra.Command, args []string, toComplete string) ([]string, cobra.ShellCompDirective) {
			if len(args) > 0 {
				return nil, cobra.ShellCompDirectiveNoFileComp
			}
			return []string{"bash", "zsh", "fish", "powershell"}, cobra.ShellCompDirectiveNoFileComp
		},
		Args: cobra.ExactValidArgs(1),
		Run: func(c *cobra.Command, args []string) {
			switch args[0] {
			case "bash":
				c.Root().GenBashCompletion(os.Stdout)
			case "zsh":
				c.Root().GenZshCompletion(os.Stdout)
			case "fish":
				c.Root().GenFishCompletion(os.Stdout, true)
			case "powershell":
				c.Root().GenPowerShellCompletionWithDesc(os.Stdout)
			}
		},
	}
	return cobraCmd
}
