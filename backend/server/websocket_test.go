package server //nolint:testpackage // were touching an internal field so its fine

import (
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/gorilla/websocket"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"google.golang.org/protobuf/encoding/protojson"

	envelopev1 "github.com/TeamHermes-jhu/clueless/backend/internal/gen/envelope/v1"
	sessionv1 "github.com/TeamHermes-jhu/clueless/backend/internal/gen/session/v1"
)

// wsURL turns an httptest server URL into a ws:// URL pointing at /ws
// with the given query string.
func wsURL(t *testing.T, ts *httptest.Server, query string) string {
	t.Helper()

	return "ws" + strings.TrimPrefix(ts.URL, "http") + "/ws?" + query
}

func dial(t *testing.T, url string) (*websocket.Conn, *http.Response, error) {
	t.Helper()

	header := http.Header{}
	header.Set("Origin", "http://localhost:8000")

	return websocket.DefaultDialer.Dial(url, header)
}

func TestHandleWS_MissingQueryParams(t *testing.T) {
	s := newTestServer()

	ts := httptest.NewServer(s.mux)
	defer ts.Close()

	conn, resp, err := dial(t, wsURL(t, ts, ""))
	require.Error(t, err)
	require.Nil(t, conn)
	require.NotNil(t, resp)

	assert.Equal(t, http.StatusBadRequest, resp.StatusCode)
}

func TestHandleWS_InvalidSessionID(t *testing.T) {
	s := newTestServer()

	ts := httptest.NewServer(s.mux)
	defer ts.Close()

	url := wsURL(t, ts, "session=not-a-uuid&player=alice")

	conn, resp, err := dial(t, url)
	require.Error(t, err)
	require.Nil(t, conn)
	require.NotNil(t, resp)

	assert.Equal(t, http.StatusBadRequest, resp.StatusCode)
}

func TestHandleWS_UnknownSession(t *testing.T) {
	s := newTestServer()

	ts := httptest.NewServer(s.mux)
	defer ts.Close()

	url := wsURL(t, ts, "session=00000000-0000-0000-0000-000000000000&player=alice")

	conn, resp, err := dial(t, url)
	require.Error(t, err)
	require.Nil(t, conn)
	require.NotNil(t, resp)

	assert.Equal(t, http.StatusNotFound, resp.StatusCode)
}

func TestHandleWS_SuccessfulHandshakeAndRoundTrip(t *testing.T) {
	s := newTestServer()

	ts := httptest.NewServer(s.mux)
	defer ts.Close()

	sessionID := joinAndGetSessionID(t, ts, "alice")

	url := wsURL(t, ts, fmt.Sprintf("session=%s&player=alice", sessionID))

	conn, resp, err := dial(t, url)
	require.NoError(t, err)

	defer conn.Close()

	require.Equal(t, http.StatusSwitchingProtocols, resp.StatusCode)

	// A message sent before the game has started should come back as an
	// error envelope, proving the read/write pumps are wired up.
	env := envelopev1.Envelope_builder{
		Msg: sessionv1.Message_builder{Msg: new("hello")}.Build(),
	}.Build()

	data, err := protojson.Marshal(env)
	require.NoError(t, err)
	require.NoError(t, conn.WriteMessage(websocket.TextMessage, data))

	require.NoError(t, conn.SetReadDeadline(time.Now().Add(2*time.Second)))

	_, reply, err := conn.ReadMessage()
	require.NoError(t, err)

	var replyEnv envelopev1.Envelope
	require.NoError(t, protojson.Unmarshal(reply, &replyEnv))
	assert.Equal(t, "error game not started", replyEnv.GetError().GetError())
}

// joinAndGetSessionID calls /join_session over HTTP and returns the
// session id assigned to the given username.
func joinAndGetSessionID(t *testing.T, ts *httptest.Server, username string) string {
	t.Helper()

	body := joinBody(t, username)

	resp, err := http.Post(ts.URL+"/join_session", "application/json", strings.NewReader(string(body))) //nolint:noctx // test package
	require.NoError(t, err)

	defer resp.Body.Close()

	require.Equal(t, http.StatusOK, resp.StatusCode)

	data, err := io.ReadAll(resp.Body)
	require.NoError(t, err)

	var joinResp sessionv1.JoinResponse
	require.NoError(t, protojson.Unmarshal(data, &joinResp))

	return joinResp.GetSessionId()
}
