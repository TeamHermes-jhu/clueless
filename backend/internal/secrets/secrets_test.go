package secrets_test

import (
	"context"
	"errors"
	"os"
	"testing"

	"github.com/TeamHermes-jhu/clueless/backend/internal/secrets"
)

// ErrStoreUnreachable is a sentinel error used by fakeProvider to simulate a store connection failure.
var errStoreUnreachable = errors.New("store unreachable")

type fakeProvider struct {
	dotenv string
	err    error
}

func (f fakeProvider) Fetch(context.Context, string) (string, error) {
	return f.dotenv, f.err
}

func (fakeProvider) Close(context.Context) error {
	return nil
}

func TestLoadIntoEnvSetsNewVariables(t *testing.T) {
	// t.Setenv registers the cleanup that restores original state after test.
	t.Setenv("SECRETS_TEST_NEW", "")

	_ = os.Unsetenv("SECRETS_TEST_NEW")

	m := secrets.NewManager(fakeProvider{dotenv: "SECRETS_TEST_NEW=value\n"})
	if err := m.LoadIntoEnv(context.Background(), "dev"); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if got := os.Getenv("SECRETS_TEST_NEW"); got != "value" {
		t.Errorf("SECRETS_TEST_NEW = %q, want %q", got, "value")
	}
}

func TestLoadIntoEnvDoesNotOverrideExisting(t *testing.T) {
	t.Setenv("SECRETS_TEST_EXISTING", "keep")

	m := secrets.NewManager(fakeProvider{dotenv: "SECRETS_TEST_EXISTING=overwritten\n"})
	if err := m.LoadIntoEnv(context.Background(), "dev"); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if got := os.Getenv("SECRETS_TEST_EXISTING"); got != "keep" {
		t.Errorf("SECRETS_TEST_EXISTING = %q, want %q", got, "keep")
	}
}

func TestLoadReturnsProviderError(t *testing.T) {
	wantErr := errStoreUnreachable
	m := secrets.NewManager(fakeProvider{err: wantErr})

	if _, err := m.Load(context.Background(), "dev"); !errors.Is(err, wantErr) {
		t.Errorf("error = %v, want %v", err, wantErr)
	}
}

func TestLoadRejectsInvalidDotenv(t *testing.T) {
	m := secrets.NewManager(fakeProvider{dotenv: "THIS LINE IS NOT VALID\n=\n"})

	if _, err := m.Load(context.Background(), "dev"); err == nil {
		t.Error("expected an error for invalid .env text, got nil")
	}
}

func TestOverridesWinOverStoredValues(t *testing.T) {
	t.Setenv("SECRETS_TEST_OVERRIDDEN", "")

	_ = os.Unsetenv("SECRETS_TEST_OVERRIDDEN")

	m := secrets.NewManager(fakeProvider{dotenv: "SECRETS_TEST_OVERRIDDEN=from-store\n"}).
		WithOverrides(map[string]string{"SECRETS_TEST_OVERRIDDEN": "from-override"})
	if err := m.LoadIntoEnv(context.Background(), "dev"); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if got := os.Getenv("SECRETS_TEST_OVERRIDDEN"); got != "from-override" {
		t.Errorf("got %q, want %q", got, "from-override")
	}
}

func TestRealEnvironmentBeatsOverrides(t *testing.T) {
	t.Setenv("SECRETS_TEST_PRIORITY", "from-shell")

	m := secrets.NewManager(fakeProvider{dotenv: "SECRETS_TEST_PRIORITY=from-store\n"}).
		WithOverrides(map[string]string{"SECRETS_TEST_PRIORITY": "from-override"})
	if err := m.LoadIntoEnv(context.Background(), "dev"); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if got := os.Getenv("SECRETS_TEST_PRIORITY"); got != "from-shell" {
		t.Errorf("got %q, want %q", got, "from-shell")
	}
}
