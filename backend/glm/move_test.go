package glm_test

import (
	"math/rand/v2"
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/TeamHermes-jhu/clueless/backend/glm"
	cardsv1 "github.com/TeamHermes-jhu/clueless/backend/internal/gen/cards/v1"
	sessionv1 "github.com/TeamHermes-jhu/clueless/backend/internal/gen/session/v1"
)

func TestMove(t *testing.T) {
	// using a deterministic rand so player 2 is always going first
	r := rand.New(rand.NewPCG(1, 2))
	gm := glm.New(glm.WithRand(r))

	gm.InitializeGameState([]string{"1", "2", "3"})

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
				Room:     cardsv1.Rooms_BALLROOM.Enum(),
			}.Build(),
			want: "not players turn",
		},
		{
			name: "invalid player id",
			move: sessionv1.Move_builder{
				PlayerId: new("0"),
				Room:     cardsv1.Rooms_BALLROOM.Enum(),
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
			name: "not adjacent move",
			move: sessionv1.Move_builder{
				PlayerId: new("2"),
				Room:     cardsv1.Rooms_LOUNGE.Enum(),
			}.Build(),
			want: "location is not next to the players current location",
		},
		{
			name: "valid room move",
			move: sessionv1.Move_builder{
				PlayerId: new("2"),
				Room:     cardsv1.Rooms_LIBRARY.Enum(),
			}.Build(),
			want:  "player 2 has moved to " + cardsv1.Rooms_LIBRARY.String(),
			moved: true,
		},
	}

	for _, test := range tt {
		t.Run(test.name, func(t *testing.T) {
			before := gm.GetPlayerLocation(test.move.GetPlayerId())
			if before != nil {
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

			after := gm.GetPlayerLocation(test.move.GetPlayerId())

			if test.move.HasHall() {
				assert.Equal(t, test.move.GetHall(), after.GetHall())
			} else if test.move.HasRoom() {
				assert.Equal(t, test.move.GetRoom(), after.GetRoom())
			}
		})
	}
}

func TestOccupiedHall(t *testing.T) {
	// using a deterministic rand so player 2 is always going first
	r := rand.New(rand.NewPCG(1, 2))
	gm := glm.New(glm.WithRand(r))

	gm.InitializeGameState([]string{"1", "2", "3"})

	occupant := gm.GetCurrentPlayer()

	occupantLoc := gm.GetPlayerLocation(occupant)
	occupantLoc.SetHall(cardsv1.Halls_ONE)

	gm.AdvanceTurn()

	mover := gm.GetCurrentPlayer()

	moverLoc := gm.GetPlayerLocation(mover)
	moverLoc.SetRoom(cardsv1.Rooms_STUDY)

	got := gm.Move(
		sessionv1.Move_builder{
			PlayerId: new(mover),
			Hall:     cardsv1.Halls_ONE.Enum(),
		}.Build(),
	)

	assert.Equal(t, "hall is occupied by player", got)
}
