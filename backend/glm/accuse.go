package glm

import (
	"fmt"

	suggestionv1 "github.com/TeamHermes-jhu/clueless/backend/internal/gen/suggestion/v1"
)

// accuseFmt is the public notice saying what a player accused.
const accuseFmt = "player %s has made an accusation: %s with the %s in the %s"

// AccusationResult is what the session needs to report an accusation back to the players.
type AccusationResult struct {
	// Announcement to broadcast saying what was accused.
	Announcement string
	// IsCorrect reports whether the accusation matched the case file.
	IsCorrect bool
}

// Accuse handles a player's accusation and returns whether accusation was correct and whether there were any errors.
func (g *GameLogicManager) Accuse(accusation *suggestionv1.AccusationRequest) (*AccusationResult, error) {
	if accusation == nil || accusation.GetSuggestion() == nil {
		return nil, errInvalidAccusation
	}

	guess := accusation.GetSuggestion()

	playerID := guess.GetPlayerId()

	if !g.isValidTurn(playerID) {
		return nil, errInvalidTurn
	}

	// check to see if character, weapon, and room in the accusation matches the case file.
	isCorrect := guess.GetCharacter() == g.caseFile.GetCharacter() &&
		guess.GetWeapon() == g.caseFile.GetWeapon() &&
		guess.GetRoom() == g.caseFile.GetRoom()

	// if accusation is incorrect, mark the player as eliminated.
	if !isCorrect {
		g.eliminated[playerID] = struct{}{}
	}

	return &AccusationResult{
		Announcement: fmt.Sprintf(
			accuseFmt,
			guess.GetPlayerId(),
			guess.GetCharacter(),
			guess.GetWeapon(),
			guess.GetRoom(),
		),
		IsCorrect: isCorrect,
	}, nil
}
