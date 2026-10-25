package server //nolint:testpackage // were touching an internal field so its fine

import (
	"bytes"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"google.golang.org/protobuf/encoding/protojson"

	sessionv1 "github.com/TeamHermes-jhu/clueless/backend/internal/gen/session/v1"
)

func doJoin(t *testing.T, s *Server, body []byte) *httptest.ResponseRecorder {
	t.Helper()

	req := httptest.NewRequest(http.MethodPost, "/join_session", bytes.NewReader(body)) //nolint:noctx // its a test package
	rec := httptest.NewRecorder()

	s.mux.ServeHTTP(rec, req)

	return rec
}

func joinBody(t *testing.T, username string) []byte {
	t.Helper()

	req := sessionv1.JoinRequest_builder{
		Username: new(username),
	}.Build()

	body, err := protojson.Marshal(req)
	require.NoError(t, err)

	return body
}

func TestJoinSession_MalformedBody(t *testing.T) {
	s := newTestServer()

	rec := doJoin(t, s, []byte("not json"))

	assert.Equal(t, http.StatusBadRequest, rec.Code)
}

func TestJoinSession_MissingUsername(t *testing.T) {
	s := newTestServer()

	rec := doJoin(t, s, []byte("{}"))

	assert.Equal(t, http.StatusBadRequest, rec.Code)
}

func TestJoinSession_CreatesNewSession(t *testing.T) {
	s := newTestServer()

	rec := doJoin(t, s, joinBody(t, "alice"))

	require.Equal(t, http.StatusOK, rec.Code)

	var resp sessionv1.JoinResponse
	require.NoError(t, protojson.Unmarshal(rec.Body.Bytes(), &resp))

	assert.Equal(t, "alice", resp.GetUsername())
	assert.NotEmpty(t, resp.GetSessionId())
	assert.Len(t, s.sessions, 1)
}

func TestJoinSession_SecondPlayerJoinsExistingSession(t *testing.T) {
	s := newTestServer()

	first := doJoin(t, s, joinBody(t, "alice"))
	require.Equal(t, http.StatusOK, first.Code)

	var firstResp sessionv1.JoinResponse
	require.NoError(t, protojson.Unmarshal(first.Body.Bytes(), &firstResp))

	second := doJoin(t, s, joinBody(t, "bob"))
	require.Equal(t, http.StatusOK, second.Code)

	var secondResp sessionv1.JoinResponse
	require.NoError(t, protojson.Unmarshal(second.Body.Bytes(), &secondResp))

	assert.Equal(t, firstResp.GetSessionId(), secondResp.GetSessionId())
	assert.Len(t, s.sessions, 1)
}

func TestJoinSession_AllSessionsFull(t *testing.T) {
	s := newTestServer()

	for i := range maxSessionsCount {
		for j := range 6 {
			username := fmt.Sprintf("session%d_player%d", i, j)

			rec := doJoin(t, s, joinBody(t, username))
			require.Equal(t, http.StatusOK, rec.Code)
		}
	}

	require.Len(t, s.sessions, maxSessionsCount)

	rec := doJoin(t, s, joinBody(t, "latecomer"))

	assert.Equal(t, http.StatusOK, rec.Code)
	assert.Contains(t, rec.Body.String(), "Sorry, all sessions full")
}
