// Copyright (c) 2026 Julien Briault
//
// This program is free software: you can redistribute it and/or modify
// it under the terms of the GNU General Public License as published by
// the Free Software Foundation, either version 3 of the License, or
// (at your option) any later version.
//
// This program is distributed in the hope that it will be useful,
// but WITHOUT ANY WARRANTY; without even the implied warranty of
// MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the
// GNU General Public License for more details.
//
// You should have received a copy of the GNU General Public License
// along with this program.  If not, see <https://www.gnu.org/licenses/>.

package cli

import (
	"os"

	"github.com/spf13/cobra"
)

var completionCmd = &cobra.Command{
	Use:   "completion [bash|zsh|fish|powershell]",
	Short: "Generate shell completion scripts",
	Long: `Generate shell completion scripts for resumectl.

Usage examples:
  resumectl completion bash        # Output bash completion script
  resumectl completion zsh         # Output zsh completion script
  resumectl completion fish        # Output fish completion script
  resumectl completion powershell  # Output PowerShell completion script

To load completions:

Bash:
  $ source <(resumectl completion bash)

  # To load completions for each session, execute once:
  # Linux:
  $ resumectl completion bash > /etc/bash_completion.d/resumectl
  # macOS:
  $ resumectl completion bash > $(brew --prefix)/etc/bash_completion.d/resumectl

Zsh:
  # If shell completion is not already enabled in your environment,
  # you will need to enable it. You can execute the following once:
  $ echo "autoload -U compinit; compinit" >> ~/.zshrc

  # To load completions for each session, execute once:
  $ resumectl completion zsh > "${fpath[1]}/_resumectl"

  # You will need to start a new shell for this setup to take effect.

Fish:
  $ resumectl completion fish | source

  # To load completions for each session, execute once:
  $ resumectl completion fish > ~/.config/fish/completions/resumectl.fish

PowerShell:
  PS> resumectl completion powershell | Out-String | Invoke-Expression

  # To load completions for every new session, run:
  PS> resumectl completion powershell > resumectl.ps1
  # and source this file from your PowerShell profile.
`,
	DisableFlagsInUseLine: true,
	ValidArgs:             []string{"bash", "zsh", "fish", "powershell"},
	Args:                  cobra.MatchAll(cobra.ExactArgs(1), cobra.OnlyValidArgs),
	Run: func(cmd *cobra.Command, args []string) {
		switch args[0] {
		case "bash":
			cmd.Root().GenBashCompletion(os.Stdout)
		case "zsh":
			cmd.Root().GenZshCompletion(os.Stdout)
		case "fish":
			cmd.Root().GenFishCompletion(os.Stdout, true)
		case "powershell":
			cmd.Root().GenPowerShellCompletionWithDesc(os.Stdout)
		}
	},
}

func init() {
	rootCmd.AddCommand(completionCmd)
}
