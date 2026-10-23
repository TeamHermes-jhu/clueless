// Loads secrets from MongoDB so real credentials are never in the repo
package secrets

import (
	"context"
	"fmt"
	"os"

	"github.com/joho/godotenv"
)

// Service adapter for a remote secret store
// MongoProvider is real implementation, tests use a fake
type Provider interface {
	// Returns raw contents of .env file stored for an environment (e.g. "dev")
	Fetch(ctx context.Context, environment string) (string, error)
	// Releases connection to store
	Close(ctx context.Context) error
}

// Turns stored .env text into process environment variables
type Manager struct {
	provider Provider
}

// Creates a Manager that reads from given Provider
func NewManager(p Provider) *Manager {
	return &Manager{provider: p}
}

// Fetches and parses .env contents for an environment
func (m *Manager) Load(ctx context.Context, environment string) (map[string]string, error) {
	raw, err := m.provider.Fetch(ctx, environment)
	if err != nil {
		return nil, err
	}

	values, err := godotenv.Unmarshal(raw)
	if err != nil {
		// Error may contain secret
		return nil, fmt.Errorf("secrets for %q are not valid .env syntax", environment)
	}
	return values, nil
}

// Loads secrets and sets them as environment variables
// Variables already set are left alone
func (m *Manager) LoadIntoEnv(ctx context.Context, environment string) error {
	values, err := m.Load(ctx, environment)
	if err != nil {
		return err
	}

	for key, value := range values {
		if _, exists := os.LookupEnv(key); exists {
			continue
		}
		if err := os.Setenv(key, value); err != nil {
			return fmt.Errorf("set %s: %w", key, err)
		}
	}
	return nil
}
