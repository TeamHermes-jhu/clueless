package cmd

import (
	"context"
	"fmt"
	"os"
	"time"

	"github.com/spf13/cobra"

	"github.com/TeamHermes-jhu/clueless/backend/internal/secrets"
)

var serveCmd = &cobra.Command{
	Use:   "serve",
	Short: "Run the backend server.",
	Long:  `Starts the backend server.`,
	RunE: func(cmd *cobra.Command, _ []string) error {
		ctx, cancel := context.WithTimeout(cmd.Context(), 10*time.Second)
		defer cancel()

		if err := secrets.Bootstrap(ctx, "secrets.env"); err != nil {
			return fmt.Errorf("load secrets: %w", err)
		}

		// temp check to prove secrets loaded without printing the value
		// TODO: delete once server uses real settings
		fmt.Println("EXAMPLE_SECRET loaded:", os.Getenv("EXAMPLE_SECRET") != "")

		fmt.Println("The server has started")
		return nil
	},
}

func init() {
	rootCmd.AddCommand(serveCmd)
}
