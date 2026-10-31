package cmd

import (
	"context"
	"errors"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/spf13/cobra"

	"github.com/TeamHermes-jhu/clueless/backend/internal/secrets"
	"github.com/TeamHermes-jhu/clueless/backend/server"
)

const (
	defaultShutdownTime = 15 * time.Second
	secretsLoadTimeout  = 10 * time.Second
)

var serveCmd = &cobra.Command{
	Use:   "serve",
	Short: "Run the backend server.",
	Long:  `Starts the backend server.`,
	Run:   RunServer,
}

func init() {
	serveCmd.Flags().String("addr", "", "address to run server on")
	serveCmd.Flags().String("port", "8000", "port to run server on")

	rootCmd.AddCommand(serveCmd)
}

// RunServer loads secrets, sets up the server on the port and serves.
func RunServer(cmd *cobra.Command, _ []string) {
	l := slog.New(slog.NewJSONHandler(os.Stdout, nil))

	if err := loadSecrets(cmd.Context()); err != nil {
		l.Error("Error loading secrets", "error", err)

		return
	}

	l.Info("Secrets loaded")

	addr, err := cmd.Flags().GetString("addr")
	if err != nil {
		l.Error("Error getting flag", "flag", "addr")

		return
	}

	port, err := cmd.Flags().GetString("port")
	if err != nil {
		l.Error("Error getting flag", "flag", "port")

		return
	}

	s := server.New(l)

	serverErr := make(chan error, 1)

	go func() {
		if err := s.Serve(addr, port); !errors.Is(err, http.ErrServerClosed) {
			serverErr <- err
		}
	}()

	stop := make(chan os.Signal, 1)
	signal.Notify(stop, os.Interrupt, syscall.SIGTERM)

	select {
	case err := <-serverErr:
		l.Info("Server error", "error", err)
	case sig := <-stop:
		l.Info("Received shutdown signal", "signal", sig)
	}

	l.Info("Server is shutting down")

	ctx, cancel := context.WithTimeout(context.Background(), defaultShutdownTime)
	defer cancel()

	if err := s.Shutdown(ctx); err != nil {
		l.Error("Server shutdwn error", "error", err)

		return
	}

	l.Info("Server shutdown sucessfully")
}

// loadSecrets loads the application secrets into the process environment.
func loadSecrets(ctx context.Context) error {
	ctx, cancel := context.WithTimeout(ctx, secretsLoadTimeout)
	defer cancel()

	return secrets.Bootstrap(ctx, secretsFile, overrideFile)
}
