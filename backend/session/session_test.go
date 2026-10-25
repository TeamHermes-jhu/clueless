// Package session_test provides tests for the session package.
package session_test

import (
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/TeamHermes-jhu/clueless/backend/internal/gen/session/v1"
	"github.com/TeamHermes-jhu/clueless/backend/session"
)

func TestMakeSession(t *testing.T) {
	sl := new(SessionList)
	sl, s, err := makeSession(sl, 1)
	require.NoError(t, err)
	assert.Len(t, sl.GetSessions(), 1)
	assert.Equal(t, int32(1), s.GetSessionId())
	assert.Equal(t, SessionStatus_WAITING, s.GetStatus())

	assert.Equal(t, 3, int(s.GetMinPlayers()))
	assert.Equal(t, 6, int(s.GetMaxPlayers()))
	assert.Empty(t, s.GetPlayers())
}

func TestMakeSessionDuplicate(t *testing.T) {
	sl := new(SessionList)
	sl, s1, err1 := makeSession(sl, 1)
	sl, s2, err2 := makeSession(sl, 1)

	require.NoError(t, err1)
	require.Error(t, err2)
	assert.Len(t, sl.GetSessions(), 1)
	assert.Equal(t, int32(1), s1.GetSessionId())
	assert.Equal(t, SessionStatus_WAITING, s1.GetStatus())
	assert.Nil(t, s2)
}

func TestMakeMultipleSessions(t *testing.T) {
	sl := new(SessionList)
	sl, s1, err1 := makeSession(sl, 1)
	sl, s2, err2 := makeSession(sl, 2)

	require.NoError(t, err1)
	require.NoError(t, err2)
	assert.Len(t, sl.GetSessions(), 2)
	assert.Equal(t, int32(1), s1.GetSessionId())
	assert.Equal(t, int32(2), s2.GetSessionId())
}

func TestAddPlayer(t *testing.T) {
	sl := new(SessionList)
	_, s, err := makeSession(sl, 1)
	require.NoError(t, err)

	p1 := new(Player)
	p1.SetPlayerId(1)
	p1.SetPlayerName("Alice")

	err = addPlayer(s, p1)
	require.NoError(t, err)
	assert.Len(t, s.GetPlayers(), 1)
	assert.Equal(t, int32(1), s.GetPlayers()[0].GetPlayerId())
	assert.Equal(t, "Alice", s.GetPlayers()[0].GetPlayerName())
}

func TestAddDuplicatePlayer(t *testing.T) {
	sl := new(SessionList)
	_, s, err := makeSession(sl, 1)
	require.NoError(t, err)

	p1 := new(Player)
	p1.SetPlayerId(1)
	p1.SetPlayerName("Alice")

	err = addPlayer(s, p1)
	require.NoError(t, err)

	err = addPlayer(s, p1)
	require.Error(t, err)
	assert.Len(t, s.GetPlayers(), 1)
}

func TestAddPlayerToFullSession(t *testing.T) {
	sl := new(SessionList)
	_, s, err := makeSession(sl, 1)
	require.NoError(t, err)

	// add 6 players to the session
	for i := 1; i <= 6; i++ {
		p := new(Player)
		p.SetPlayerId(int32(i))
		p.SetPlayerName(fmt.Sprintf("Player%d", i))
		err = addPlayer(s, p)
		require.NoError(t, err)
	}

	// try to add a 7th player
	p7 := new(Player)
	p7.SetPlayerId(7)
	p7.SetPlayerName("Player7")
	err = addPlayer(s, p7)
	require.Error(t, err)
	assert.Len(t, s.GetPlayers(), 6)
}

func TestRemovePlayer(t *testing.T) {
	sl := new(SessionList)
	_, s, err := makeSession(sl, 1)
	require.NoError(t, err)

	p1 := new(Player)
	p1.SetPlayerId(1)
	p1.SetPlayerName("Alice")

	err = addPlayer(s, p1)
	require.NoError(t, err)
	assert.Len(t, s.GetPlayers(), 1)

	err = removePlayer(s, p1)
	require.NoError(t, err)
	assert.Empty(t, s.GetPlayers())
}

func TestRemoveNonexistentPlayer(t *testing.T) {
	sl := new(SessionList)
	_, s, err := makeSession(sl, 1)
	require.NoError(t, err)

	p1 := new(Player)
	p1.SetPlayerId(1)
	p1.SetPlayerName("Alice")

	err = removePlayer(s, p1)
	require.Error(t, err)
	assert.Empty(t, s.GetPlayers())
}

func TestBeginGame(t *testing.T) {
	sl := new(SessionList)
	_, s, err := makeSession(sl, 1)
	require.NoError(t, err)

	// add 3 players to the session
	for i := 1; i <= 3; i++ {
		p := new(Player)
		p.SetPlayerId(int32(i))
		p.SetPlayerName(fmt.Sprintf("Player%d", i))
		err = addPlayer(s, p)
		require.NoError(t, err)
	}

	beginGame(s)
	assert.Equal(t, SessionStatus_ACTIVE, s.GetStatus())
}

func TestBeginGameNotEnoughPlayers(t *testing.T) {
	sl := new(SessionList)
	_, s, err := makeSession(sl, 1)
	require.NoError(t, err)

	// add only 2 players to the session
	for i := 1; i <= 2; i++ {
		p := new(Player)
		p.SetPlayerId(int32(i))
		p.SetPlayerName(fmt.Sprintf("Player%d", i))
		err = addPlayer(s, p)
		require.NoError(t, err)
	}

	err = beginGame(s)
	require.Error(t, err)
	assert.Equal(t, SessionStatus_WAITING, s.GetStatus())
}

func TestEndSession(t *testing.T) {
	sl := new(SessionList)
	sl, s1, err1 := makeSession(sl, 1)
	sl, s2, err2 := makeSession(sl, 2)
	sl, s3, err3 := makeSession(sl, 3)

	require.NoError(t, err1)
	require.NoError(t, err2)
	require.NoError(t, err3)
	assert.Len(t, sl.GetSessions(), 3)

	err := endSession(sl, s2.GetSessionId())
	require.NoError(t, err)
	assert.Len(t, sl.GetSessions(), 2)
	assert.Equal(t, s1.GetSessionId(), sl.GetSessions()[0].GetSessionId())
	assert.Equal(t, s3.GetSessionId(), sl.GetSessions()[1].GetSessionId())
}

func TestEndNonexistentSession(t *testing.T) {
	sl := new(SessionList)
	sl, _, err1 := makeSession(sl, 1)
	sl, _, err2 := makeSession(sl, 2)

	require.NoError(t, err1)
	require.NoError(t, err2)
	assert.Len(t, sl.GetSessions(), 2)

	err := endSession(sl, 3)
	require.Error(t, err)
	assert.Len(t, sl.GetSessions(), 2)
}
