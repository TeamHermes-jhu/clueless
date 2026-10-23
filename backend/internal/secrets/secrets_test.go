package secrets

import (
	"context"
	"errors"
	"os"
	"testing"
)

type fakeProvider struct {
	dotenv string
	err    error
}

func (f fakeProvider) Fetch(context.Context, string) (string, error) {
	return f.dotenv, f.err
}

func (f fakeProvider) Close(context.Context) error {
	return nil
}

func TestLoadIntoEnvSetsNewVariables(t *testing.T) {
	// t.Setenv registers the cleanup that restores original state after test
	t.Setenv("SECRETS_TEST_NEW", "")
	_ = os.Unsetenv("SECRETS_TEST_NEW")

	m := NewManager(fakeProvider{dotenv: "SECRETS_TEST_NEW=value\n"})
	if err := m.LoadIntoEnv(context.Background(), "dev"); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if got := os.Getenv("SECRETS_TEST_NEW"); got != "value" {
		t.Errorf("SECRETS_TEST_NEW = %q, want %q", got, "value")
	}
}

func TestLoadIntoEnvDoesNotOverrideExisting(t *testing.T) {
	t.Setenv("SECRETS_TEST_EXISTING", "keep")

	m := NewManager(fakeProvider{dotenv: "SECRETS_TEST_EXISTING=overwritten\n"})
	if err := m.LoadIntoEnv(context.Background(), "dev"); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if got := os.Getenv("SECRETS_TEST_EXISTING"); got != "keep" {
		t.Errorf("SECRETS_TEST_EXISTING = %q, want %q", got, "keep")
	}
}

func TestLoadReturnsProviderError(t *testing.T) {
	wantErr := errors.New("store unreachable")
	m := NewManager(fakeProvider{err: wantErr})

	if _, err := m.Load(context.Background(), "dev"); !errors.Is(err, wantErr) {
		t.Errorf("error = %v, want %v", err, wantErr)
	}
}

func TestLoadRejectsInvalidDotenv(t *testing.T) {
	m := NewManager(fakeProvider{dotenv: "THIS LINE IS NOT VALID\n=\n"})

	if _, err := m.Load(context.Background(), "dev"); err == nil {
		t.Error("expected an error for invalid .env text, got nil")
	}
}
