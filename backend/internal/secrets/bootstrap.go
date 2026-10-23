package secrets

import (
	"context"
	"errors"
	"fmt"
	"io/fs"
	"os"
	"time"

	"github.com/joho/godotenv"
)

var (
	// ErrInvalidEnvSyntax is returned when a .env file is not valid.
	ErrInvalidEnvSyntax = errors.New("the file is not valid .env syntax")
	// ErrMongoURIUnset is returned when the MongoDB URI is not set.
	ErrMongoURIUnset = errors.New("MongoDB URI is not set")
)

// Bootstrap settings
// These values are handed out by secrets.env.

// SECRETS_MONGO_URI is the connection string (don't log).
// SECRETS_DB and SECRETS_COLLECTION say where the secrets live.
// SECRETS_ENV picks the document.
const (
	envMongoURI    = "SECRETS_MONGO_URI"
	envDatabase    = "SECRETS_DB"
	envCollection  = "SECRETS_COLLECTION"
	envEnvironment = "SECRETS_ENV"
)

const bootstrapTimeout = 5 * time.Second

// Bootstrap reads bootstrap file (if it exists), connects to secret store, and loads secrets into process environment.
func Bootstrap(ctx context.Context, bootstrapFile string) error {
	provider, environment, err := openProvider(bootstrapFile)
	if err != nil {
		return err
	}
	defer closeQuietly(provider)

	return NewManager(provider).LoadIntoEnv(ctx, environment)
}

// Push checks that dotenv is valid and uploads it to secret store under given environment name, replacing whatever was stored there.
// If environment is empty, SECRETS_ENV (or "dev") is used.
func Push(ctx context.Context, bootstrapFile, environment, dotenv string) error {
	if _, err := godotenv.Unmarshal(dotenv); err != nil {
		// Parser error can contain part of secret
		return ErrInvalidEnvSyntax
	}

	provider, defaultEnvironment, err := openProvider(bootstrapFile)
	if err != nil {
		return err
	}
	defer closeQuietly(provider)

	if environment == "" {
		environment = defaultEnvironment
	}

	return provider.Push(ctx, environment, dotenv)
}

// Reads bootstrap file (if present) and connects to secret store.
// Also returns environment name configured in SECRETS_ENV.
func openProvider(bootstrapFile string) (*MongoProvider, string, error) {
	// godotenv.Load never overwrites variables that are already set
	if err := godotenv.Load(bootstrapFile); err != nil && !errors.Is(err, fs.ErrNotExist) {
		return nil, "", fmt.Errorf("read %s: %w", bootstrapFile, err)
	}

	uri := os.Getenv(envMongoURI)
	if uri == "" {
		return nil, "", fmt.Errorf("%w: %s; copy secrets.env.example to %s and fill it in",
			ErrMongoURIUnset, envMongoURI, bootstrapFile)
	}

	provider, err := NewMongoProvider(
		uri,
		valueOrDefault(envDatabase, "secrets"),
		valueOrDefault(envCollection, "env_files"),
	)
	if err != nil {
		return nil, "", err
	}

	return provider, valueOrDefault(envEnvironment, "dev"), nil
}

func closeQuietly(p *MongoProvider) {
	ctx, cancel := context.WithTimeout(context.Background(), bootstrapTimeout)
	defer cancel()

	_ = p.Close(ctx)
}

func valueOrDefault(key, fallback string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}

	return fallback
}
