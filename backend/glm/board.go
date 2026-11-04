package glm

import (
	cardsv1 "github.com/TeamHermes-jhu/clueless/backend/internal/gen/cards/v1"
	glmv1 "github.com/TeamHermes-jhu/clueless/backend/internal/gen/glm/v1"
)

type node struct {
	room cardsv1.Rooms
	hall cardsv1.Halls
}

func roomNode(r cardsv1.Rooms) node { return node{room: r} }
func hallNode(h cardsv1.Halls) node { return node{hall: h} }

var adjacency = buildAdjacency()

func buildAdjacency() map[node]map[node]bool {
	edges := [][2]node{
		// horizontal halls
		{hallNode(cardsv1.Halls_ONE), roomNode(cardsv1.Rooms_STUDY)},
		{hallNode(cardsv1.Halls_ONE), roomNode(cardsv1.Rooms_HALL)},
		{hallNode(cardsv1.Halls_TWO), roomNode(cardsv1.Rooms_HALL)},
		{hallNode(cardsv1.Halls_TWO), roomNode(cardsv1.Rooms_LOUNGE)},
		{hallNode(cardsv1.Halls_THREE), roomNode(cardsv1.Rooms_LIBRARY)},
		{hallNode(cardsv1.Halls_THREE), roomNode(cardsv1.Rooms_BILLIARD_ROOM)},
		{hallNode(cardsv1.Halls_FOUR), roomNode(cardsv1.Rooms_BILLIARD_ROOM)},
		{hallNode(cardsv1.Halls_FOUR), roomNode(cardsv1.Rooms_DINING_ROOM)},
		{hallNode(cardsv1.Halls_FIVE), roomNode(cardsv1.Rooms_CONSERVATORY)},
		{hallNode(cardsv1.Halls_FIVE), roomNode(cardsv1.Rooms_BALLROOM)},
		{hallNode(cardsv1.Halls_SIX), roomNode(cardsv1.Rooms_BALLROOM)},
		{hallNode(cardsv1.Halls_SIX), roomNode(cardsv1.Rooms_KITCHEN)},

		// vertical halls
		{hallNode(cardsv1.Halls_SEVEN), roomNode(cardsv1.Rooms_STUDY)},
		{hallNode(cardsv1.Halls_SEVEN), roomNode(cardsv1.Rooms_LIBRARY)},
		{hallNode(cardsv1.Halls_EIGHT), roomNode(cardsv1.Rooms_LIBRARY)},
		{hallNode(cardsv1.Halls_EIGHT), roomNode(cardsv1.Rooms_CONSERVATORY)},
		{hallNode(cardsv1.Halls_NINE), roomNode(cardsv1.Rooms_HALL)},
		{hallNode(cardsv1.Halls_NINE), roomNode(cardsv1.Rooms_BILLIARD_ROOM)},
		{hallNode(cardsv1.Halls_TEN), roomNode(cardsv1.Rooms_BILLIARD_ROOM)},
		{hallNode(cardsv1.Halls_TEN), roomNode(cardsv1.Rooms_BALLROOM)},
		{hallNode(cardsv1.Halls_ELEVEN), roomNode(cardsv1.Rooms_LOUNGE)},
		{hallNode(cardsv1.Halls_ELEVEN), roomNode(cardsv1.Rooms_DINING_ROOM)},
		{hallNode(cardsv1.Halls_TWELVE), roomNode(cardsv1.Rooms_DINING_ROOM)},
		{hallNode(cardsv1.Halls_TWELVE), roomNode(cardsv1.Rooms_KITCHEN)},

		// secret passages: direct room-to-room links between opposite corners
		{roomNode(cardsv1.Rooms_STUDY), roomNode(cardsv1.Rooms_KITCHEN)},
		{roomNode(cardsv1.Rooms_LOUNGE), roomNode(cardsv1.Rooms_CONSERVATORY)},
	}

	adj := make(map[node]map[node]bool, len(edges)*2)
	for _, e := range edges {
		a, b := e[0], e[1]

		if adj[a] == nil {
			adj[a] = make(map[node]bool)
		}

		if adj[b] == nil {
			adj[b] = make(map[node]bool)
		}

		adj[a][b] = true
		adj[b][a] = true
	}

	return adj
}

// isAdjacent reports whether to can be reached from from in a single move.
func isAdjacent(from, to node) bool {
	return adjacency[from][to]
}

// nodeFromLocation converts a players current location into a node.
func nodeFromLocation(loc *glmv1.PlayerLocation) node {
	if loc.HasHall() {
		return hallNode(loc.GetHall())
	}

	return roomNode(loc.GetRoom())
}
