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
	"strings"

	"github.com/ZeroBl21/cli/ch07/pScan/scan"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

// scanCmd represents the scan command
var scanCmd = &cobra.Command{
	Use:   "scan",
	Short: "Run a port scan on the host",
	RunE: func(cmd *cobra.Command, args []string) error {
		hostsFile := viper.GetString("hosts-file")

		ports, err := cmd.Flags().GetIntSlice("ports")
		if err != nil {
			return err
		}

		return scanAction(os.Stdout, hostsFile, ports)
	},
}

func init() {
	rootCmd.AddCommand(scanCmd)

	scanCmd.Flags().IntSliceP("ports", "p", []int{22, 80, 443}, "ports to scan")
}

func scanAction(out io.Writer, hostsFile string, ports []int) error {
	hl := &scan.HostsList{}

	if err := hl.Load(hostsFile); err != nil {
		return err
	}

	results := scan.Run(hl, ports)

	return printResults(out, results)
}

func printResults(out io.Writer, results []scan.Results) error {
	var sb strings.Builder

	for _, r := range results {
		fmt.Fprintf(&sb, "%s:", r.Host)

		if r.NotFound {
			fmt.Fprintf(&sb, " Host not found\n\n")
			continue
		}

		fmt.Fprintln(&sb)

		for _, p := range r.PortStates {
			fmt.Fprintf(&sb, "\t%d: %s\n", p.Port, p.Open)
		}

		fmt.Fprintln(&sb)
	}

	_, err := out.Write([]byte(sb.String()))
	return err
}
