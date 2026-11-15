package glm

import (
	"errors"
	"fmt"
	"slices"

	"google.golang.org/protobuf/proto"

	cardsv1 "github.com/TeamHermes-jhu/clueless/backend/internal/gen/cards/v1"
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
	validCards  []*cardsv1.Card
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

// BeginDisprove ensures disproverID must show one of validCards for suggestion.
// Called by suggest turn when it has found the first player after the suggester who holds a matching card?
func (g *GameLogicManager) BeginDisprove(
	suggestion *suggestionv1.SuggestionRequest,
	disproverID string,
	validCards []*cardsv1.Card,
) {
	g.pendingDisprove = &disproveState{
		suggestion:  suggestion,
		disproverID: disproverID,
		validCards:  validCards,
	}
}

// GetPendingDisprover returns the player who has to show a card, if a suggestion is waiting.
func (g *GameLogicManager) GetPendingDisprover() (string, bool) {
	if g.pendingDisprove == nil {
		return "", false
	}

	return g.pendingDisprove.disproverID, true
}

// GetDisproveOptions returns the cards the asked player may show, if a suggestion is waiting.
func (g *GameLogicManager) GetDisproveOptions() []*cardsv1.Card {
	if g.pendingDisprove == nil {
		return nil
	}

	return g.pendingDisprove.validCards
}

// CancelDisprove drops a pending disprove request, for when the asked player cannot answer
// anymore. It returns the player that made the suggestion and whether there was anything to
// cancel.
func (g *GameLogicManager) CancelDisprove() (string, bool) {
	if g.pendingDisprove == nil {
		return "", false
	}

	suggesterID := g.pendingDisprove.suggestion.GetPlayerId()
	g.pendingDisprove = nil

	return suggesterID, true
}

func (g *GameLogicManager) disproveOptions(
	playerID string,
	suggestion *suggestionv1.SuggestionRequest,
) []*cardsv1.Card {
	var cards []*cardsv1.Card

	hand := g.GetPlayerHand(playerID)

	if character := suggestion.GetCharacter(); slices.Contains(hand.GetCharacters(), character) {
		cards = append(cards, cardsv1.Card_builder{Character: &character}.Build())
	}

	if weapon := suggestion.GetWeapon(); slices.Contains(hand.GetWeapons(), weapon) {
		cards = append(cards, cardsv1.Card_builder{Weapon: &weapon}.Build())
	}

	if room := suggestion.GetRoom(); slices.Contains(hand.GetRooms(), room) {
		cards = append(cards, cardsv1.Card_builder{Room: &room}.Build())
	}

	return cards
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

	if !isValidCard(pending.validCards, resp) {
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

func isValidCard(validCards []*cardsv1.Card, resp *suggestionv1.SuggestionResponse) bool {
	shown := cardFromResponse(resp)
	if shown == nil {
		return false
	}

	return slices.ContainsFunc(validCards, func(card *cardsv1.Card) bool {
		return proto.Equal(card, shown)
	})
}

func cardFromResponse(resp *suggestionv1.SuggestionResponse) *cardsv1.Card {
	switch resp.WhichCard() {
	case suggestionv1.SuggestionResponse_Character_case:
		character := resp.GetCharacter()

		return cardsv1.Card_builder{Character: &character}.Build()

	case suggestionv1.SuggestionResponse_Weapon_case:
		weapon := resp.GetWeapon()

		return cardsv1.Card_builder{Weapon: &weapon}.Build()

	case suggestionv1.SuggestionResponse_Room_case:
		room := resp.GetRoom()

		return cardsv1.Card_builder{Room: &room}.Build()

	default:
		return nil
	}
}
