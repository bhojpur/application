package cmd

// Copyright (c) 2018 Bhojpur Consulting Private Limited, India. All rights reserved.

// Permission is hereby granted, free of charge, to any person obtaining a copy
// of this software and associated documentation files (the "Software"), to deal
// in the Software without restriction, including without limitation the rights
// to use, copy, modify, merge, publish, distribute, sublicense, and/or sell
// copies of the Software, and to permit persons to whom the Software is
// furnished to do so, subject to the following conditions:

// The above copyright notice and this permission notice shall be included in
// all copies or substantial portions of the Software.

// THE SOFTWARE IS PROVIDED "AS IS", WITHOUT WARRANTY OF ANY KIND, EXPRESS OR
// IMPLIED, INCLUDING BUT NOT LIMITED TO THE WARRANTIES OF MERCHANTABILITY,
// FITNESS FOR A PARTICULAR PURPOSE AND NONINFRINGEMENT. IN NO EVENT SHALL THE
// AUTHORS OR COPYRIGHT HOLDERS BE LIABLE FOR ANY CLAIM, DAMAGES OR OTHER
// LIABILITY, WHETHER IN AN ACTION OF CONTRACT, TORT OR OTHERWISE, ARISING FROM,
// OUT OF OR IN CONNECTION WITH THE SOFTWARE OR THE USE OR OTHER DEALINGS IN
// THE SOFTWARE.

import (
	"os"

	"github.com/spf13/cobra"
)

var completionExample = `
	# Installing bash completion on macOS using homebrew
	## If running Bash 3.2 included with macOS
	brew install bash-completion
	## or, if running Bash 4.1+
	brew install bash-completion@2
	## Add the completion to your completion directory
	appctl completion bash > $(brew --prefix)/etc/bash_completion.d/appctl
	source ~/.bash_profile

	# Installing bash completion on Linux
	## If bash-completion is not installed on Linux, please install the 'bash-completion' package
	## via your distribution's package manager.
	## Load the Bhojpur Application completion code for bash into the current shell
	source <(appctl completion bash)
	## Write bash completion code to a file and source if from .bash_profile
	appctl completion bash > ~/.bhojpur/completion.bash.inc
	printf "
	## appctl shell completion
	source '$HOME/.bhojpur/completion.bash.inc'
	" >> $HOME/.bash_profile
	source $HOME/.bash_profile

	# Installing zsh completion on macOS using homebrew
	## If zsh-completion is not installed on macOS, please install the 'zsh-completion' package
	brew install zsh-completions
	## Set the Bhojpur Application completion code for zsh[1] to autoload on startup
	appctl completion zsh > "${fpath[1]}/_appctl"
	source ~/.zshrc

	# Installing zsh completion on Linux
	## If zsh-completion is not installed on Linux, please install the 'zsh-completion' package
	## via your distribution's package manager.
	## Load the Bhojpur Application completion code for zsh into the current shell
	source <(appctl completion zsh)
	# Set the appctl completion code for zsh[1] to autoload on startup
  	appctl completion zsh > "${fpath[1]}/_appctl"

	# Installing powershell completion on Windows
	## Create $PROFILE if it not exists
	if (!(Test-Path -Path $PROFILE )){ New-Item -Type File -Path $PROFILE -Force }
	## Add the completion to your profile
	appctl completion powershell >> $PROFILE
`

func newCompletionCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:     "completion",
		Short:   "Generates shell completion scripts",
		Example: completionExample,
		Run: func(cmd *cobra.Command, args []string) {
			cmd.Help()
		},
	}

	cmd.AddCommand(
		newCompletionBashCmd(),
		newCompletionZshCmd(),
		newCompletionPowerShellCmd(),
	)

	cmd.Flags().BoolP("help", "h", false, "Print this help message")

	return cmd
}

func newCompletionBashCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "bash",
		Short: "Generates bash completion scripts",
		Run: func(cmd *cobra.Command, args []string) {
			rootCmd.GenBashCompletion(os.Stdout)
		},
	}

	cmd.Flags().BoolP("help", "h", false, "Print this help message")

	return cmd
}

func newCompletionZshCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "zsh",
		Short: "Generates zsh completion scripts",
		Run: func(cmd *cobra.Command, args []string) {
			rootCmd.GenZshCompletion(os.Stdout)
		},
	}
	cmd.Flags().BoolP("help", "h", false, "Print this help message")

	return cmd
}

func newCompletionPowerShellCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "powershell",
		Short: "Generates powershell completion scripts",
		Run: func(cmd *cobra.Command, args []string) {
			rootCmd.GenPowerShellCompletion(os.Stdout)
		},
	}
	cmd.Flags().BoolP("help", "h", false, "Print this help message")

	return cmd
}

func init() {
	rootCmd.AddCommand(newCompletionCmd())
}
