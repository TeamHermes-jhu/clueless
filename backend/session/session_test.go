package session

import (
	"testing"
	"github.com/stretchr/testify/assert"
	. "github.com/TeamHermes-jhu/clueless/backend/internal/gen/session/v1"
)

func TestMakeSession(t *testing.T) {

	sl := new(SessionList)
	sl, s, err := makeSession(sl, 1)
	assert.Nil(t, err)
	assert.Equal(t, 1, len(sl.GetSessions()))
	assert.Equal(t, int32(1), s.GetSessionId())
	assert.Equal(t, SessionStatus_WAITING, s.GetStatus())

	assert.Equal(t, 3, int(s.GetMinPlayers()))
	assert.Equal(t, 6, int(s.GetMaxPlayers()))
	assert.Equal(t, 0, len(s.GetPlayers()))
}

func TestMakeSessionDuplicate(t *testing.T) {
	
	sl := new(SessionList)
	sl, s1, err := makeSession(sl, 1)
	sl, s2, err := makeSession(sl, 1)

	assert.NotNil(t, err)
	assert.Equal(t, 1, len(sl.GetSessions()))
	assert.Equal(t, int32(1), s1.GetSessionId())
	assert.Equal(t, SessionStatus_WAITING, s1.GetStatus())
	assert.Nil(t, s2)
}

func TestMakeMultipleSessions(t *testing.T) {
	sl := new(SessionList)
	sl, s1, err1 := makeSession(sl, 1)
	sl, s2, err2 := makeSession(sl, 2)
	
	assert.Nil(t, err1)
	assert.Nil(t, err2)
	assert.Equal(t, 2, len(sl.GetSessions()))
	assert.Equal(t, int32(1), s1.GetSessionId())
	assert.Equal(t, int32(2), s2.GetSessionId())
}
