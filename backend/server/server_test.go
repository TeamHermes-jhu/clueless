package server

import (
	"context"
	"io"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/TeamHermes-jhu/clueless/backend/session"
)

func newTestServer() *Server {
	l := slog.New(slog.NewTextHandler(io.Discard, nil))

	return New(l)
}

func TestNew_RegistersRoutes(t *testing.T) {
	s := newTestServer()

	rec := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodPost, "/join_session", nil)
	s.mux.ServeHTTP(rec, req)

	assert.NotEqual(t, http.StatusNotFound, rec.Code)

	rec = httptest.NewRecorder()
	req = httptest.NewRequest(http.MethodGet, "/not-a-real-route", nil)
	s.mux.ServeHTTP(rec, req)

	assert.Equal(t, http.StatusNotFound, rec.Code)
}

func TestShutdown_NoRegisteredClients(t *testing.T) {
	s := newTestServer()

	l := slog.New(slog.NewTextHandler(io.Discard, nil))
	sess := session.New(l)
	s.sessions[sess.GetID()] = sess

	require.True(t, sess.Add("alice"))

	err := s.Shutdown(context.Background())
	require.NoError(t, err)
}
