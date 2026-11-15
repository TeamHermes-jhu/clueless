// Package cmd contains the cli commands, and flags for running clueless.
package cmd

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"
)

var (
	secretsFile  string
	overrideFile string
)

var rootCmd = &cobra.Command{
	Use:   "server",
	Short: "Backend server for clueless.",
	Long:  `This starts up the clueless backend server.`,
}

func init() {
	rootCmd.PersistentFlags().StringVar(&secretsFile, "secrets-file", "secrets.env",
		"file with the secrets store connection settings")
	rootCmd.PersistentFlags().StringVar(&overrideFile, "override-file", ".env.override",
		"optional file whose values override the stored secrets")
}

// Execute runs the root command and allows for the cli to work.
func Execute() {
	if err := rootCmd.Execute(); err != nil {
		fmt.Fprint(os.Stderr, err)
	}
}
