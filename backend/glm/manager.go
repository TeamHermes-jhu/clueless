// Package glm implements the game logic manager for the Clue/Cluedo game.
// It manages the state of the game, including the case file, player hands, player locations, and turn order.
// It provides methods to initialize the game state, handle player actions (move, suggest, accuse),
// and manage the flow of the game.
package glm

import (
	"errors"
	"fmt"
	"math/rand/v2"
	"slices"

	cardsv1 "github.com/TeamHermes-jhu/clueless/backend/internal/gen/cards/v1"
	glmv1 "github.com/TeamHermes-jhu/clueless/backend/internal/gen/glm/v1"
	suggestionv1 "github.com/TeamHermes-jhu/clueless/backend/internal/gen/suggestion/v1"
)

const (
	// text formats for logging.
	suggestFmt = "player %s has suggested: %s with the %s in the %s"
	accuseFmt  = "player %s has made an accusation: %s with the %s in the %s"

	// number of each type of card in the game.
	numSuspects = 6
	numRooms    = 9
	numWeapons  = 6

	// card type numbers.
	cardSuspect = 1
	cardRoom    = 2
	cardWeapon  = 3

	// number of halls in the game.
	numHalls = 12

	// text formats for game initialization.
	caseFileFmt = "Spoiler alert! case file: suspect %s, room %s, weapon %s"
)

var errInvalidTurn = errors.New("not players turn")

// GameLogicManager manages the game logic for a session, including the case file, player hands, player locations, and the list of player IDs.
type GameLogicManager struct {
	// needs to maintain a case file (suspect, room, weapon).
	caseFile *glmv1.CaseFile

	// needs to maintain the turn order and current turn.
	turnOrder []string
	turnIdx   int

	// needs to maintain all the players' hands.
	playerStates map[string]*glmv1.PlayerGameState

	// tracks Players that are eliminated.
	eliminated map[string]bool

	rng Rand
}

// Option is passed to New for GameLogicManager to affect it in someway.
type Option func(*GameLogicManager)

// New creates a new GameLogicManager with the given inputs.
func New(opts ...Option) *GameLogicManager {
	gm := &GameLogicManager{
		rng: globalRand{},
	}

	for _, o := range opts {
		o(gm)
	}

	return gm
}

// WithRand is an optional random number generator to use.
func WithRand(r *rand.Rand) Option {
	return func(g *GameLogicManager) { g.rng = r }
}

// InitializeGameState sets up the initial state of the game, including the turn order, player hands, and player locations.
func (g *GameLogicManager) InitializeGameState(playerIDs []string) {
	g.turnIdx = 0
	g.playerStates = make(map[string]*glmv1.PlayerGameState)

	g.turnOrder = playerIDs

	g.rng.Shuffle(len(g.turnOrder), func(i, j int) {
		g.turnOrder[i], g.turnOrder[j] = g.turnOrder[j], g.turnOrder[i]
	})

	// initialize eliminated tracker
	g.eliminated = make(map[string]bool)

	g.setupGame(playerIDs)
}

// AdvanceTurn updates the current turn index.
func (g *GameLogicManager) AdvanceTurn() {
	if len(g.turnOrder) == 0 || g.PlayersRemaining() == 0 {
		return
	}
	
	for _ := range len(g.turnOrder) {
		g.turnIdx = (g.turnIdx + 1) % len(g.turnOrder)

		if !g.eliminated[g.turnOrder[g.turnIdx]]{
			return
		}
	}
}

// PlayersRemaining checks to see how many players are remaining.
func(g *GameLogicManager) PlayersRemaining() int {
	count := 0

	for _, playerID := range g.turnOrder{
		if !g.eliminated[playerID] {
			count++
		}
	}
	
	return count
}

// Suggest handles a player's suggestion and returns a message indicating the suggestion.
func (g *GameLogicManager) Suggest(suggestion *suggestionv1.SuggestionRequest) string {
	// Do something with GameLogicManager so the linter doesn't complain.
	_ = g
	// TODO: figure out how to handle suggestions.
	return fmt.Sprintf(
		suggestFmt,
		suggestion.GetPlayerId(),
		suggestion.GetCharacter().String(),
		suggestion.GetWeapon().String(),
		suggestion.GetRoom().String(),
	)
}

// Accuse handles a player's accusation and returns whether accusation was correct and whether there were any errors.
func (g *GameLogicManager) Accuse(accusation *suggestionv1.AccusationRequest) (bool, error) {
	// check to see if accusation and suggestion are valid, and return an error if either are missing.
	if accusation == nil || accusation.GetSuggestion() == nil {
		return false, errors.New("invalid accusation")
	} 

	// Get the suggestion containing the player's accusation.
	guess := accusation.GetSuggestion()

	// Get the playerID of player who made accusation.
	playerID := guess.GetPlayerId()

	// Check to see if it is the player's turn, and return an error if it is not their turn.
	if !g.isValidTurn(playerID){
		return false, errInvalidTurn
	}

	// check to see if character, weapon, and room in the accusation matches the case file.
	correct :=
		guess.GetCharacter() == g.caseFile.GetCharacter() &&
		guess.GetWeapon() == g.caseFile.GetWeapon() &&
		guess.GetRoom() == g.caseFile.GetRoom()

	// if accusation is incorrect, mark the player as eliminated.
	if !correct {
		g.eliminated[playerID] = true
	}

	return correct, nil
}

// GetCurrentPlayer returns the ID of the player whose turn it is.
func (g *GameLogicManager) GetCurrentPlayer() string {
	if g.turnIdx < 0 || g.turnIdx >= len(g.turnOrder) {
		return ""
	}

	return g.turnOrder[g.turnIdx]
}

// GetTurnMessage returns a message indicating whose turn it is.
func (g *GameLogicManager) GetTurnMessage() string {
	return fmt.Sprintf("It is now %s's turn", g.turnOrder[g.turnIdx])
}

// GetCaseFile returns a string representation of the current case file.
func (g *GameLogicManager) GetCaseFile() string {
	return fmt.Sprintf(
		caseFileFmt,
		g.caseFile.GetCharacter().String(),
		g.caseFile.GetWeapon().String(),
		g.caseFile.GetRoom().String(),
	)
}

// GetPlayerHand returns a player's hand as a protobuf message.
func (g *GameLogicManager) GetPlayerHand(playerID string) *glmv1.PlayerHand {
	state, ok := g.playerStates[playerID]
	if !ok {
		return nil
	}

	return state.GetHand()
}

// GetPlayerLocation returns a player's location as a protobuf message.
func (g *GameLogicManager) GetPlayerLocation(playerID string) *glmv1.PlayerLocation {
	state, ok := g.playerStates[playerID]
	if !ok {
		return nil
	}

	return state.GetLocation()
}

// setupGame initializes the game by setting up the case file, dealing cards, and placing players in their starting rooms.
func (g *GameLogicManager) setupGame(playerIDs []string) {
	g.setupCaseFile()
	g.initializePlayerStates(playerIDs)
}

// setupCaseFile creates the case file with a random suspect, room, and weapon.
func (g *GameLogicManager) setupCaseFile() {
	suspect := cardsv1.Characters(g.rng.IntN(numSuspects) + 1)
	room := cardsv1.Rooms(g.rng.IntN(numRooms) + 1)
	weapon := cardsv1.Weapons(g.rng.IntN(numWeapons) + 1)

	g.caseFile = glmv1.CaseFile_builder{
		Character: &suspect,
		Room:      &room,
		Weapon:    &weapon,
	}.Build()
}

func (g *GameLogicManager) initializePlayerStates(playerIDs []string) {
	g.playerStates = make(map[string]*glmv1.PlayerGameState)
	for _, playerID := range playerIDs {
		g.playerStates[playerID] = glmv1.PlayerGameState_builder{
			Hand: glmv1.PlayerHand_builder{}.Build(),
		}.Build()
	}

	g.dealCards()
	g.setupPlayerLocations()
}

// dealCards distributes the cards to the players in a round-robin fashion.
func (g *GameLogicManager) dealCards() {
	// we have 18 cards left, 5 suspects/weapons and 8 rooms
	allCards := make([]int32, numSuspects+numRooms+numWeapons)
	for i := range allCards {
		allCards[i] = int32(i + 1)
	}

	allCards = removeCardFromDeck(allCards, int32(g.caseFile.GetCharacter()), cardSuspect)
	allCards = removeCardFromDeck(allCards, int32(g.caseFile.GetRoom()), cardRoom)
	allCards = removeCardFromDeck(allCards, int32(g.caseFile.GetWeapon()), cardWeapon)

	g.rng.Shuffle(len(allCards), func(i, j int) {
		allCards[i], allCards[j] = allCards[j], allCards[i]
	})

	i := 0

	for i < len(allCards) {
		for _, state := range g.playerStates {
			if i >= len(allCards) {
				break
			}

			card := allCards[i]

			switch {
			case card <= numSuspects:
				// assign suspect card to current player
				currentCharacters := state.GetHand().GetCharacters()
				state.GetHand().SetCharacters(
					append(currentCharacters, cardsv1.Characters(card)),
				)

			case card <= numSuspects+numRooms:
				// assign room card to current player
				card -= numSuspects
				currentRooms := state.GetHand().GetRooms()
				state.GetHand().SetRooms(
					append(currentRooms, cardsv1.Rooms(card)),
				)

			default:
				// assign weapon card to current player
				card -= numSuspects + numRooms
				currentWeapons := state.GetHand().GetWeapons()
				state.GetHand().SetWeapons(
					append(currentWeapons, cardsv1.Weapons(card)),
				)
			}

			i++
		}
	}
}

// removeCardFromDeck removes a specific card from the deck based on its type and number.
func removeCardFromDeck(deck []int32, cardNum int32, cardType int) []int32 {
	// cardType == 1 -> suspect
	// cardType == 2 -> room
	// cardType == 3 -> weapon
	switch cardType {
	case cardSuspect:
		// suspect cards are at the beginning of the deck
	case cardRoom:
		// room cards are after the suspect cards in the deck
		cardNum += numSuspects
	case cardWeapon:
		// weapon cards are after the room cards in the deck
		cardNum += numSuspects + numRooms
	}

	// remove the card from the deck
	deck = slices.DeleteFunc(deck, func(c int32) bool {
		return c == cardNum
	})

	return deck
}

// setupPlayerLocations places each player in a random starting room.
func (g *GameLogicManager) setupPlayerLocations() {
	for _, state := range g.playerStates {
		state.SetLocation(glmv1.PlayerLocation_builder{
			Hall: new(cardsv1.Halls(g.rng.IntN(numHalls) + 1)),
		}.Build())
	}
}

func (g *GameLogicManager) isValidTurn(id string) bool {
	return g.turnOrder[g.turnIdx] == id
}
