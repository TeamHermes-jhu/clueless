package glm

import (
	"errors"
	"fmt"
	"slices"

	suggestionv1 "github.com/TeamHermes-jhu/clueless/backend/internal/gen/suggestion/v1"
)

// disproveFmt is the public notice saying who disproved the suggestion.
const disproveFmt = "player %s disproved player %s's suggestion"

var (
	// ErrNoPendingDisprove is returned when nobody has been asked to show a card.
	ErrNoPendingDisprove = errors.New("there is nothing to disprove right now")
	// ErrNotDisprover is returned when someone other than the asked player answers.
	ErrNotDisprover = errors.New("it is not your turn to disprove")
	// ErrWrongSuggestion is returned when the answer is for a different suggestion.
	ErrWrongSuggestion = errors.New("that is not the suggestion you were asked about")
	// ErrMustShowCard is returned when the asked player does not show a card.
	ErrMustShowCard = errors.New("you must show a card")
	// ErrCardNotAllowed is returned when the card is not in the player's hand or is not
	// part of the suggestion.
	ErrCardNotAllowed = errors.New("that card does not disprove the suggestion")
)

// disproveState is a suggestion that is waiting for one player to show a card.
type disproveState struct {
	suggestion  *suggestionv1.SuggestionRequest
	disproverID string
}

// DisproveResult is what the session needs after a successful disprove.
type DisproveResult struct {
	// SuggesterID is the only player who may see the card.
	SuggesterID string
	// Message to broadcast saying who disproved the suggestion.
	Message string
	// Reveal has the shown card and who showed it, which is only sent to the suggester.
	Reveal *suggestionv1.SuggestionResponse
}

// BeginDisprove ensures disproverID must show a card for suggestion.
// Called by suggest turn when it has found the first player after the suggester who holds a matching card?
func (g *GameLogicManager) BeginDisprove(suggestion *suggestionv1.SuggestionRequest, disproverID string) {
	g.pendingDisprove = &disproveState{
		suggestion:  suggestion,
		disproverID: disproverID,
	}
}

// GetPendingDisprover returns the player who has to show a card, if a suggestion is waiting.
func (g *GameLogicManager) GetPendingDisprover() (string, bool) {
	if g.pendingDisprove == nil {
		return "", false
	}

	return g.pendingDisprove.disproverID, true
}

// Disprove handles the answer to a disprove request.
// The asked player shows a card.
func (g *GameLogicManager) Disprove(playerID string, resp *suggestionv1.SuggestionResponse) (*DisproveResult, error) {
	pending := g.pendingDisprove
	if pending == nil {
		return nil, ErrNoPendingDisprove
	}

	if playerID != pending.disproverID {
		return nil, ErrNotDisprover
	}

	if resp.GetSuggestionId() != pending.suggestion.GetSuggestionId() {
		return nil, ErrWrongSuggestion
	}

	if !resp.GetHasCard() {
		return nil, ErrMustShowCard
	}

	if !g.holdsMatchingCard(playerID, pending.suggestion, resp) {
		return nil, ErrCardNotAllowed
	}

	g.pendingDisprove = nil

	// The server decides who showed the card, whatever the client put in the message.
	resp.SetResponderId(playerID)

	return &DisproveResult{
		SuggesterID: pending.suggestion.GetPlayerId(),
		Message:     fmt.Sprintf(disproveFmt, playerID, pending.suggestion.GetPlayerId()),
		Reveal:      resp,
	}, nil
}

// holdsMatchingCard says whether the shown card is part of the suggestion and is in the player's hand.
func (g *GameLogicManager) holdsMatchingCard(
	playerID string,
	suggestion *suggestionv1.SuggestionRequest,
	resp *suggestionv1.SuggestionResponse,
) bool {
	hand := g.GetPlayerHand(playerID)

	switch resp.WhichCard() {
	case suggestionv1.SuggestionResponse_Character_case:
		card := resp.GetCharacter()

		return card == suggestion.GetCharacter() && slices.Contains(hand.GetCharacters(), card)

	case suggestionv1.SuggestionResponse_Weapon_case:
		card := resp.GetWeapon()

		return card == suggestion.GetWeapon() && slices.Contains(hand.GetWeapons(), card)

	case suggestionv1.SuggestionResponse_Room_case:
		card := resp.GetRoom()

		return card == suggestion.GetRoom() && slices.Contains(hand.GetRooms(), card)

	default:
		return false
	}
}
