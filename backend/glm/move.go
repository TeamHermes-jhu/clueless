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
	// ErrHallOccupied is when a hall is occupied and a player tried to move there.
	ErrHallOccupied = errors.New("hall is occupied by another player")

	// ErrNoLocation is when a Move request is missing a locaiton.
	ErrNoLocation = errors.New("you must select a location")

	// ErrNotAdjacent is when to location isn't adjacent to players current location.
	ErrNotAdjacent = errors.New("the location you select is too far")

	// ErrAlreadyMoved is when a player has already moved this turn.
	ErrAlreadyMoved = errors.New("you have already moved this turn")
)

// Move updates the player's location and returns a message indicating the move.
func (g *GameLogicManager) Move(move *sessionv1.Move) string {
	err := g.validateMove(move)
	if err != nil {
		return err.Error()
	}

	g.hasMoved = true

	state := g.playerStates[move.GetPlayerId()]

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
	state, ok := g.playerStates[move.GetPlayerId()]
	if !ok {
		return ErrUnknownPlayer
	}

	if !g.isValidTurn(move.GetPlayerId()) {
		return ErrInvalidTurn
	}

	if g.hasMoved {
		return ErrAlreadyMoved
	}

	var to node

	switch move.WhichMoveTo() {
	case sessionv1.Move_Hall_case:
		to = hallNode(move.GetHall())

	case sessionv1.Move_Room_case:
		to = roomNode(move.GetRoom())

	default:
		return ErrNoLocation
	}

	if !isAdjacent(nodeFromLocation(state.GetLocation()), to) {
		return ErrNotAdjacent
	}

	if !move.HasHall() {
		return nil
	}

	for playerID, state := range g.playerStates {
		if move.GetPlayerId() == playerID {
			continue
		}

		if state.GetLocation().HasRoom() {
			continue
		}

		if move.GetHall() == state.GetLocation().GetHall() {
			return ErrHallOccupied
		}
	}

	return nil
}
