package glm

import (
	"errors"

	suggestionv1 "github.com/TeamHermes-jhu/clueless/backend/internal/gen/suggestion/v1"
)

// ErrDisproveNotImplemented is returned by the disprove stubs until the exchange is built.
var ErrDisproveNotImplemented = errors.New("disproving a suggestion is not implemented yet")

// TODO: send the disprove request, with the matching cards, to the disproving player.
func (g *GameLogicManager) disproveRequest(_ string, _ *suggestionv1.SuggestionRequest) {
	_ = g
}

// HandleDisproveResponse is the stub for the card a disproving player chose to show.
// TODO: check the chosen card against the suggestion and report the result to the suggesting player.
func (g *GameLogicManager) HandleDisproveResponse(_ *suggestionv1.SuggestionResponse) (*SuggestionResult, error) {
	_ = g

	return nil, ErrDisproveNotImplemented
}
