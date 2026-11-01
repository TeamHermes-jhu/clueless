// Package glm implements the game logic manager for the Clue/Cluedo game.
// It manages the state of the game, including the case file, player hands, player locations, and turn order.
// It provides methods to initialize the game state, handle player actions (move, suggest, accuse),
// and manage the flow of the game.
package glm

import (
	"fmt"
	"math/rand/v2"
	"slices"

	cardsv1 "github.com/TeamHermes-jhu/clueless/backend/internal/gen/cards/v1"
	glmv1 "github.com/TeamHermes-jhu/clueless/backend/internal/gen/glm/v1"
	sessionv1 "github.com/TeamHermes-jhu/clueless/backend/internal/gen/session/v1"
	suggestionv1 "github.com/TeamHermes-jhu/clueless/backend/internal/gen/suggestion/v1"
)

const (
	// text formats for logging.
	suggestFmt = "player %s has suggested: %s with the %s in the %s"
	accuseFmt  = "player %s has made an accusation: %s with the %s in the %s"
	moveFmt    = "player %s has moved to %s"

	// number of each type of card in the game.
	numSuspects = 6
	numRooms    = 9
	numWeapons  = 6

	// card type numbers.
	cardSuspect = 1
	cardRoom    = 2
	cardWeapon  = 3

	// number of halls in the game.
	// numHalls = 12 will be used later.

	// text formats for game initialization.
	caseFileFmt       = "Spoiler alert! case file: suspect %s, room %s, weapon %s"
	playerHandFmt     = "Your hand: suspects %v, rooms %v, weapons %v"
	playerLocationFmt = "player %s location: %s"
)

// GameLogicManager manages the game logic for a session, including the case file, player hands, player locations, and the list of player IDs.
type GameLogicManager struct {
	// needs to maintain a case file (suspect, room, weapon)
	caseFile *glmv1.CaseFile

	// needs to maintain the turn order and current turn
	turnOrder []string
	turnIdx   int

	// needs to maintain all the players' hands
	playerStates map[string]*glmv1.PlayerGameState

	// needs to maintain the list of player IDs
	playerIDs []string
}

// InitializeGameState sets up the initial state of the game, including the turn order, player hands, and player locations.
func (g *GameLogicManager) InitializeGameState(playerIDs []string) {
	g.playerIDs = playerIDs
	g.turnIdx = 0
	g.playerStates = make(map[string]*glmv1.PlayerGameState)

	r := rand.Perm(len(g.playerIDs))

	g.turnOrder = make([]string, len(g.playerIDs))
	for i, v := range r {
		g.turnOrder[v] = g.playerIDs[i]
	}

	g.setupGame()
}

// SetTurn updates the current turn index and returns a message indicating whose turn it is.
func (g *GameLogicManager) SetTurn(turnIdx int) string {
	g.turnIdx = (turnIdx % len(g.turnOrder))
	return fmt.Sprintf("It is now %s's turn", g.turnOrder[g.turnIdx])
}

// Move updates the player's location and returns a message indicating the move.
func (g *GameLogicManager) Move(move *sessionv1.Move) string {
	// Do something with GameLogicManager so the linter doesn't complain.
	_ = g
	// TODO: validate its a good move
	to := move.GetRoom().String()

	if move.WhichMoveTo() == sessionv1.Move_Hall_case {
		to = move.GetRoom().String()
	}

	return fmt.Sprintf(moveFmt, move.GetPlayerId(), to)
}

// Suggest handles a player's suggestion and returns a message indicating the suggestion.
func (g *GameLogicManager) Suggest(suggestion *suggestionv1.SuggestionRequest) string {
	// Do something with GameLogicManager so the linter doesn't complain.
	_ = g
	// TODO: figure out how to handle suggestions
	return fmt.Sprintf(
		suggestFmt,
		suggestion.GetPlayerId(),
		suggestion.GetCharacter().String(),
		suggestion.GetWeapon().String(),
		suggestion.GetRoom().String(),
	)
}

// Accuse handles a player's accusation and returns a message indicating the accusation.
func (g *GameLogicManager) Accuse(accusation *suggestionv1.AccusationRequest) string {
	// Do something with GameLogicManager so the linter doesn't complain.
	_ = g
	// TODO: validation + logic
	return fmt.Sprintf(
		accuseFmt,
		accusation.GetSuggestion().GetPlayerId(),
		accusation.GetSuggestion().GetCharacter().String(),
		accusation.GetSuggestion().GetWeapon().String(),
		accusation.GetSuggestion().GetRoom().String(),
	)
}

// GetCurrentPlayer returns the ID of the player whose turn it is.
func (g *GameLogicManager) GetCurrentPlayer() string {
	if g.turnIdx < 0 || g.turnIdx >= len(g.turnOrder) {
		return ""
	}

	return g.turnOrder[g.turnIdx]
}

// GetCurrentTurnIndex returns the index of the current turn in the turn order.
func (g *GameLogicManager) GetCurrentTurnIndex() int {
	return g.turnIdx
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

func (g *GameLogicManager) getPlayerState(playerID string) *glmv1.PlayerGameState {
	// Return nil if the player does not exist.
	state, ok := g.playerStates[playerID]
	if !ok {
		return nil
	}

	return state
}

// GetPlayerHand returns a string representation of a player's hand.
func (g *GameLogicManager) GetPlayerHand(playerID string) string {
	state := g.getPlayerState(playerID)
	if state == nil {
		return ""
	}

	hand := state.GetHand()

	return fmt.Sprintf(
		playerHandFmt,
		hand.GetCharacters(),
		hand.GetRooms(),
		hand.GetWeapons(),
	)
}

// GetPlayerLocation returns a string representation of a player's location.
func (g *GameLogicManager) GetPlayerLocation(playerID string) string {
	state := g.getPlayerState(playerID)
	if state == nil {
		return ""
	}

	location := state.GetLocation()
	isHall := location.HasHall()

	var locationName string
	if isHall {
		locationName = location.GetHall().String()
	} else {
		locationName = location.GetRoom().String()
	}

	return fmt.Sprintf(
		playerLocationFmt,
		playerID,
		locationName,
	)
}

// setupGame initializes the game by setting up the case file, dealing cards, and placing players in their starting rooms.
func (g *GameLogicManager) setupGame() {
	g.setupCaseFile()
	g.initializePlayerStates()
	g.dealCards()
	g.setupPlayerLocations()
}

// setupCaseFile creates the case file with a random suspect, room, and weapon.
func (g *GameLogicManager) setupCaseFile() {
	//nolint:gosec // rand.IntN is used for selecting random cards, not for security-sensitive purposes.
	suspect := cardsv1.Characters(rand.IntN(numSuspects) + 1)
	//nolint:gosec // rand.IntN is used for selecting random cards, not for security-sensitive purposes.
	room := cardsv1.Rooms(rand.IntN(numRooms) + 1)
	//nolint:gosec // rand.IntN is used for selecting random cards, not for security-sensitive purposes.
	weapon := cardsv1.Weapons(rand.IntN(numWeapons) + 1)

	g.caseFile = glmv1.CaseFile_builder{
		Character: &suspect,
		Room:      &room,
		Weapon:    &weapon,
	}.Build()
}

func (g *GameLogicManager) initializePlayerStates() {
	g.playerStates = make(map[string]*glmv1.PlayerGameState)
	for _, playerID := range g.playerIDs {
		g.playerStates[playerID] = glmv1.PlayerGameState_builder{}.Build()
	}

	g.initializeHands()
}

func (g *GameLogicManager) initializeHands() {
	for _, playerID := range g.playerIDs {
		g.playerStates[playerID].SetHand(glmv1.PlayerHand_builder{
			Characters: []cardsv1.Characters{},
			Rooms:      []cardsv1.Rooms{},
			Weapons:    []cardsv1.Weapons{},
		}.Build())
	}
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
	// we want to deal them sequentially to players, uniformly randomly (all shuffled together)

	rand.Shuffle(len(allCards), func(i, j int) {
		allCards[i], allCards[j] = allCards[j], allCards[i]
	})

	currentplayerIDxToDeal := 1

	for len(allCards) > 0 {
		cardToDeal := allCards[0]
		allCards = allCards[1:]
		// assign the card to the current player
		switch {
		case cardToDeal <= numSuspects:
			// assign suspect card to current player
			currentCharacters := g.playerStates[g.playerIDs[currentplayerIDxToDeal-1]].GetHand().GetCharacters()
			g.playerStates[g.playerIDs[currentplayerIDxToDeal-1]].GetHand().SetCharacters(
				append(currentCharacters, cardsv1.Characters(cardToDeal)))
		case cardToDeal <= numSuspects+numRooms:
			// assign room card to current player
			cardToDeal -= numSuspects
			currentRooms := g.playerStates[g.playerIDs[currentplayerIDxToDeal-1]].GetHand().GetRooms()
			g.playerStates[g.playerIDs[currentplayerIDxToDeal-1]].GetHand().SetRooms(
				append(currentRooms, cardsv1.Rooms(cardToDeal)))
		default:
			// assign weapon card to current player
			cardToDeal -= numSuspects + numRooms
			currentWeapons := g.playerStates[g.playerIDs[currentplayerIDxToDeal-1]].GetHand().GetWeapons()
			g.playerStates[g.playerIDs[currentplayerIDxToDeal-1]].GetHand().SetWeapons(
				append(currentWeapons, cardsv1.Weapons(cardToDeal)))
		}

		currentplayerIDxToDeal = (currentplayerIDxToDeal % len(g.playerIDs)) + 1
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
	for _, playerID := range g.playerIDs {
		//nolint:gosec // rand.IntN is used for selecting random rooms, not for security-sensitive purposes.
		room := cardsv1.Rooms(rand.IntN(numRooms) + 1)
		g.playerStates[playerID].SetLocation(glmv1.PlayerLocation_builder{
			Room: &room,
		}.Build())
	}
}
