// Package secrets loads secrets from MongoDB so real credentials are never in the repo.
package secrets

import (
	"context"
	"errors"
	"fmt"
	"maps"
	"os"

	"github.com/joho/godotenv"
)

// ErrInvalidSecretsEnv is returned when the stored .env text is not valid.
var ErrInvalidSecretsEnv = errors.New("secrets are not valid .env syntax")

// Provider is the service adapter for a remote secret store.
// MongoProvider is real implementation, tests use a fake.
type Provider interface {
	// Returns raw contents of .env file stored for an environment (e.g. "dev")
	Fetch(ctx context.Context, environment string) (string, error)
	// Releases connection to store
	Close(ctx context.Context) error
}

// Manager turns stored .env text into process environment variables.
type Manager struct {
	provider  Provider
	overrides map[string]string
}

// NewManager creates a Manager that reads from given Provider.
func NewManager(p Provider) *Manager {
	return &Manager{provider: p}
}

// Load fetches and parses .env contents for an environment.
func (m *Manager) Load(ctx context.Context, environment string) (map[string]string, error) {
	raw, err := m.provider.Fetch(ctx, environment)
	if err != nil {
		return nil, err
	}

	values, err := godotenv.Unmarshal(raw)
	if err != nil {
		// Error may contain secret
		return nil, fmt.Errorf("%w for environment %q", ErrInvalidSecretsEnv, environment)
	}

	return values, nil
}

// LoadIntoEnv loads secrets and sets them as environment variables.
// Variables already set are left alone.
func (m *Manager) LoadIntoEnv(ctx context.Context, environment string) error {
	values, err := m.Load(ctx, environment)
	if err != nil {
		return err
	}

	maps.Copy(values, m.overrides) // local overrides replace stored values

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

// WithOverrides sets local values that win over stored secrets.
func (m *Manager) WithOverrides(overrides map[string]string) *Manager {
	m.overrides = overrides
	return m
}
