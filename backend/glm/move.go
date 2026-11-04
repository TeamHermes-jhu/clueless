package glm

import (
	"errors"
	"fmt"

	sessionv1 "github.com/TeamHermes-jhu/clueless/backend/internal/gen/session/v1"
)

const (
	moveFmt = "player %s has moved to %s"
)

var (
	errRoomOccupied = errors.New("room is occupied by player")
	errNoLocation   = errors.New("no location selected")
)

// Move updates the player's location and returns a message indicating the move.
func (g *GameLogicManager) Move(move *sessionv1.Move) string {
	state, ok := g.playerStates[move.GetPlayerId()]
	if !ok {
		return "invalid player id"
	}

	err := g.validateMove(move)
	if err != nil {
		return err.Error()
	}

	var to string
	if move.HasRoom() {
		to = move.GetRoom().String()

		state.GetLocation().SetRoom(move.GetRoom())
	} else {
		to = move.GetHall().String()

		state.GetLocation().SetHall(move.GetHall())
	}

	return fmt.Sprintf(moveFmt, move.GetPlayerId(), to)
}

func (g *GameLogicManager) validateMove(move *sessionv1.Move) error {
	if !g.isValidTurn(move.GetPlayerId()) {
		return errInvalidTurn
	}

	switch move.WhichMoveTo() {
	case sessionv1.Move_Hall_case:
		return nil

	case sessionv1.Move_Room_case:
		for playerID, state := range g.playerStates {
			if move.GetPlayerId() == playerID {
				continue
			}

			if state.GetLocation().HasHall() {
				continue
			}

			if move.GetRoom() == state.GetLocation().GetRoom() {
				return errRoomOccupied
			}
		}

		return nil
	}

	return errNoLocation
}
