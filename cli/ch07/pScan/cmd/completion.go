/*
Copyright © 2025 NAME HERE <EMAIL ADDRESS>

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

	http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/
package cmd

import (
	"fmt"
	"io"
	"os"
	"runtime"

	"github.com/spf13/cobra"
)

// completionCmd represents the completion command
var completionCmd = &cobra.Command{
	Use:   "completion",
	Short: "Generate bash completion for your command",
	Long: `To load your completions run
source <(pScan completion)

To load completions automatically on login, add this line to your
.bashrc file:
$ ~/.bashrc source <(pScan completion)
`,
	RunE: func(cmd *cobra.Command, args []string) error {
		return completionAction(os.Stdout)
	},
}

func init() {
	rootCmd.AddCommand(completionCmd)
}

func completionAction(out io.Writer) error {
	var shell string

	if runtime.GOOS == "windows" {
		shell = os.Getenv("ComSpec") // Normalmente "C:\Windows\System32\cmd.exe" o la ruta de PowerShell
	} else {
		shell = os.Getenv("SHELL") // Ejemplo: "/bin/bash", "/usr/bin/fish", "/bin/zsh"
	}

	if shell == "" {
		return fmt.Errorf("unknown terminal")
	}

	switch shell {
	case "/bin/bash":
		return rootCmd.GenBashCompletion(out)
	case "/usr/bin/fish":
		return rootCmd.GenFishCompletion(out, false)
	case "/bin/zsh":
		return rootCmd.GenZshCompletion(out)

	default:
		return fmt.Errorf("unsupported terminal")
	}
}
