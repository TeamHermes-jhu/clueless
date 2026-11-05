package server //nolint:testpackage // uses the unexported test helpers in this package

import (
	"fmt"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/gorilla/websocket"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"google.golang.org/protobuf/encoding/protojson"

	cardsv1 "github.com/TeamHermes-jhu/clueless/backend/internal/gen/cards/v1"
	envelopev1 "github.com/TeamHermes-jhu/clueless/backend/internal/gen/envelope/v1"
	glmv1 "github.com/TeamHermes-jhu/clueless/backend/internal/gen/glm/v1"
)

// connectPlayer joins the session as username and opens a websocket for them.
func connectPlayer(t *testing.T, ts *httptest.Server, username string) *websocket.Conn {
	t.Helper()

	sessionID := joinAndGetSessionID(t, ts, username)

	conn, _, err := dial(t, wsURL(t, ts, fmt.Sprintf("session=%s&player=%s", sessionID, username)))
	require.NoError(t, err)

	t.Cleanup(func() { _ = conn.Close() })

	return conn
}

// sendSelect sends a character selection request.
func sendSelect(t *testing.T, conn *websocket.Conn, character cardsv1.Characters) {
	t.Helper()

	env := envelopev1.Envelope_builder{
		SelectCharacter: glmv1.CharacterSelectRequest_builder{
			Character: &character,
		}.Build(),
	}.Build()

	data, err := protojson.Marshal(env)
	require.NoError(t, err)
	require.NoError(t, conn.WriteMessage(websocket.TextMessage, data))
}

// readUntil reads envelopes until one satisfies match. Other messages, such as the game
// start countdown, are skipped.
func readUntil(t *testing.T, conn *websocket.Conn, match func(*envelopev1.Envelope) bool) *envelopev1.Envelope {
	t.Helper()

	require.NoError(t, conn.SetReadDeadline(time.Now().Add(5*time.Second)))

	for {
		_, data, err := conn.ReadMessage()
		require.NoError(t, err)

		var env envelopev1.Envelope
		require.NoError(t, protojson.Unmarshal(data, &env))

		if match(&env) {
			return &env
		}
	}
}

func TestCharacterSelection_BroadcastsToAllPlayers(t *testing.T) {
	s := newTestServer()

	ts := httptest.NewServer(s.mux)
	defer ts.Close()

	alice := connectPlayer(t, ts, "alice")
	bob := connectPlayer(t, ts, "bob")

	sendSelect(t, alice, cardsv1.Characters_PLUM)

	for _, conn := range []*websocket.Conn{alice, bob} {
		env := readUntil(t, conn, func(e *envelopev1.Envelope) bool {
			return e.HasCharacterSelection()
		})

		assignments := env.GetCharacterSelection().GetAssignments()
		require.Len(t, assignments, 1)
		assert.Equal(t, "alice", assignments[0].GetPlayerId())
		assert.Equal(t, cardsv1.Characters_PLUM, assignments[0].GetCharacter())
	}
}

func TestCharacterSelection_RejectsTakenCharacter(t *testing.T) {
	s := newTestServer()

	ts := httptest.NewServer(s.mux)
	defer ts.Close()

	alice := connectPlayer(t, ts, "alice")
	bob := connectPlayer(t, ts, "bob")

	sendSelect(t, alice, cardsv1.Characters_PLUM)

	// Wait until bob has seen alice's pick, so the next request is definitely a conflict.
	readUntil(t, bob, func(e *envelopev1.Envelope) bool {
		return e.HasCharacterSelection()
	})

	sendSelect(t, bob, cardsv1.Characters_PLUM)

	env := readUntil(t, bob, func(e *envelopev1.Envelope) bool {
		return e.HasError()
	})

	assert.Equal(t, "character already taken", env.GetError().GetError())
}
