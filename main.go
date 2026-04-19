package main

import (
	"fmt"
	"os"

	"github.com/AmadlaOrg/conduct/cmd"
	"github.com/spf13/cobra"
)

const (
	appName = "conduct"
	version = "1.0.0"
)

var rootCmd = &cobra.Command{
	Use:     appName,
	Short:   "Multi-server orchestrator for the Amadla tool pipeline",
	Version: version,
}

func init() {
	rootCmd.AddCommand(cmd.DeployCmd)
	rootCmd.AddCommand(cmd.StatusCmd)
	rootCmd.AddCommand(cmd.DestroyCmd)
	rootCmd.AddCommand(cmd.ExecCmd)
}

func main() {
	if err := rootCmd.Execute(); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}
