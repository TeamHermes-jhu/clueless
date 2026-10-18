// Package cmd contains the cli commands, and flags for running clueless.
package cmd

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"
)

var rootCmd = &cobra.Command{
	Use:   "server",
	Short: "Backend server for clueless.",
	Long:  `This starts up the clueless backend server.`,
}

// Execute runs the root command and allows for the cli to work.
func Execute() {
	if err := rootCmd.Execute(); err != nil {
		fmt.Fprint(os.Stderr, err)
	}
}
