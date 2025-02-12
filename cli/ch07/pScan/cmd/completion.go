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
	"strings"

	"github.com/spf13/cobra"
)

// completionCmd represents the completion command
var completionCmd = &cobra.Command{
	Use:   "completion",
	Short: "Generate bash completion for your command",
	Long: `To load your completions run:
source <(pScan completion)

To load completions automatically on login, add this line to your shell config:
- Bash:   echo 'source <(pScan completion --bash)' >> ~/.bashrc
- Zsh:    echo 'source <(pScan completion --zsh)' >> ~/.zshrc
- Fish:   pScan completion --fish | source ~/.config/fish/config.fis
- PowerShell:  pScan completion --powershell | Out-File -Append $PROFILE
`,
	RunE: func(cmd *cobra.Command, args []string) error {
		shells := map[string]bool{
			"bash":       false,
			"zsh":        false,
			"fish":       false,
			"powershell": false,
		}

		for shell := range shells {
			value, err := cmd.Flags().GetBool(shell)
			if err != nil {
				return err
			}
			shells[shell] = value
		}

		var shell string
		for s, enabled := range shells {
			if enabled {
				shell = s
				break
			}
		}

		if shell == "" {
			var err error
			shell, err = getShell()
			if err != nil {
				return err
			}
		}

		return completionAction(os.Stdout, shell)
	},
}

func init() {
	rootCmd.AddCommand(completionCmd)

	rootCmd.Flags().BoolP("bash", "b", false, "bash shell")
	rootCmd.Flags().BoolP("zsh", "z", false, "zsh shell")
	rootCmd.Flags().BoolP("fish", "f", false, "fish shell")
	rootCmd.Flags().BoolP("powershell", "ps", false, "powershell shell")
}

func getShell() (string, error) {
	var shell string

	if runtime.GOOS == "windows" {
		if strings.Contains(strings.ToLower(os.Getenv("TERM_PROGRAM")), "powershell") {
			return "powershell", nil
		}
		if strings.Contains(strings.ToLower(os.Getenv("ComSpec")), "cmd.exe") {
			return "cmd", nil
		}
		return "", fmt.Errorf("unsupported shell on Windows")
	}

	shell = os.Getenv("SHELL")

	if shell == "" {
		return "", fmt.Errorf("unknown shell")
	}

	switch shell {
	case "/bin/bash":
		return "bash", nil
	case "/bin/zsh":
		return "zsh", nil
	case "/usr/bin/fish":
		return "fish", nil
	default:
		return "", fmt.Errorf("unsupported shell")
	}
}

func completionAction(out io.Writer, shell string) error {
	switch shell {
	case "bash":
		return rootCmd.GenBashCompletion(out)
	case "fish":
		return rootCmd.GenZshCompletion(out)
	case "zsh":
		return rootCmd.GenFishCompletion(out, false)
	case "powershell":
		return rootCmd.GenPowerShellCompletion(out)
	case "cmd":
		return fmt.Errorf("command-line completion is not supported in cmd.exe")
	default:
		return fmt.Errorf("unknown shell: %s", shell)
	}
}
