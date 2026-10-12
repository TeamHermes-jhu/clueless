package cmd

import (
	"fmt"

	"github.com/spf13/cobra"
)

var serveCmd = &cobra.Command{
	Use:   "serve",
	Short: "Run the backend server.",
	Long:  `Starts the backend server.`,
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Println("The server has started")
	},
}

func init() {
	rootCmd.AddCommand(serveCmd)
}
