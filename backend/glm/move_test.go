package glm_test

import (
	"math/rand/v2"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/TeamHermes-jhu/clueless/backend/glm"
	cardsv1 "github.com/TeamHermes-jhu/clueless/backend/internal/gen/cards/v1"
	sessionv1 "github.com/TeamHermes-jhu/clueless/backend/internal/gen/session/v1"
)

func TestMove(t *testing.T) {
	// using a deterministic rand so player 2 is always going first
	r := rand.New(rand.NewPCG(1, 2))
	gm := glm.New(glm.WithRand(r))

	gm.InitializeGameState([]string{"1", "2", "3"})

	// player 1 never moves so using their location as the occupied room.
	occupiedLoc, err := gm.GetPlayerLocation("1")
	require.NoError(t, err)

	occupiedRoom := occupiedLoc.GetRoom()

	player3Loc, err := gm.GetPlayerLocation("3")
	require.NoError(t, err)

	freeRoom := cardsv1.Rooms_UNSPECIFIED

	for v := 1; v <= 9; v++ {
		candidate := cardsv1.Rooms(int32(v))

		if candidate != occupiedRoom && candidate != player3Loc.GetRoom() {
			freeRoom = candidate

			break
		}
	}

	require.NotEqual(t, cardsv1.Rooms_UNSPECIFIED, freeRoom)

	tt := []struct {
		name  string
		move  *sessionv1.Move
		want  string
		moved bool
	}{
		{
			name: "not players turn",
			move: sessionv1.Move_builder{
				PlayerId: new("1"),
				Hall:     cardsv1.Halls_ONE.Enum(),
			}.Build(),
			want: "not players turn",
		},
		{
			name: "invalid player id",
			move: sessionv1.Move_builder{
				PlayerId: new("0"),
				Hall:     cardsv1.Halls_ONE.Enum(),
			}.Build(),
			want: "invalid player id",
		},
		{
			name: "no location selected",
			move: sessionv1.Move_builder{
				PlayerId: new("2"),
			}.Build(),
			want: "no location selected",
		},
		{
			name: "room occupied by another player",
			move: sessionv1.Move_builder{
				PlayerId: new("2"),
				Room:     occupiedRoom.Enum(),
			}.Build(),
			want: "room is occupied by player",
		},
		{
			name: "valid hall move",
			move: sessionv1.Move_builder{
				PlayerId: new("2"),
				Hall:     cardsv1.Halls_ONE.Enum(),
			}.Build(),
			want:  "player 2 has moved to HALLS_ONE",
			moved: true,
		},
		{
			name: "valid room move",
			move: sessionv1.Move_builder{
				PlayerId: new("2"),
				Room:     freeRoom.Enum(),
			}.Build(),
			want:  "player 2 has moved to " + freeRoom.String(),
			moved: true,
		},
	}

	for _, test := range tt {
		t.Run(test.name, func(t *testing.T) {
			before, beforeErr := gm.GetPlayerLocation(test.move.GetPlayerId())
			if beforeErr == nil {
				if test.move.HasHall() {
					assert.NotEqual(t, test.move.GetHall(), before.GetHall())
				} else if test.move.HasRoom() {
					assert.NotEqual(t, test.move.GetRoom(), before.GetRoom())
				}
			}

			got := gm.Move(test.move)

			assert.Equal(t, test.want, got)

			if !test.moved {
				return
			}

			after, err := gm.GetPlayerLocation(test.move.GetPlayerId())
			require.NoError(t, err)

			if test.move.HasHall() {
				assert.Equal(t, test.move.GetHall(), after.GetHall())
			} else if test.move.HasRoom() {
				assert.Equal(t, test.move.GetRoom(), after.GetRoom())
			}
		})
	}
}
