package cmd

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"time"

	"github.com/spf13/cobra"

	"github.com/TeamHermes-jhu/clueless/backend/internal/secrets"
)

var (
	secretsPushFile string
	secretsPushEnv  string
)

const commandTimeout = 10 * time.Second

var secretsPushCmd = &cobra.Command{
	Use:   "secrets-push",
	Short: "Upload a local .env file to the secrets store.",
	Long: `Reads a local .env file and stores its contents in the secrets store (MongoDB)
under the given environment name, replacing whatever was stored there.`,
	RunE: func(cmd *cobra.Command, _ []string) error {
		raw, err := os.ReadFile(filepath.Clean(secretsPushFile))
		if err != nil {
			return fmt.Errorf("read %s: %w", secretsPushFile, err)
		}

		ctx, cancel := context.WithTimeout(cmd.Context(), commandTimeout)
		defer cancel()

		if err := secrets.Push(ctx, "secrets.env", secretsPushEnv, string(raw)); err != nil {
			return fmt.Errorf("upload secrets: %w", err)
		}

		fmt.Printf("Uploaded %s to the secrets store.\n", secretsPushFile)

		return nil
	},
}

func init() {
	secretsPushCmd.Flags().StringVar(&secretsPushFile, "file", ".env",
		"path to the .env file to upload")
	secretsPushCmd.Flags().StringVar(&secretsPushEnv, "env", "",
		`environment name to store it under (default: SECRETS_ENV, or "dev")`)
	rootCmd.AddCommand(secretsPushCmd)
}
