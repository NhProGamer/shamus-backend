package helpers

import (
	"fmt"
	"math/rand"
	"shamus-backend/internal/domain/entities"
	"shamus-backend/internal/domain/entities/factories"
)

// AssignRoles distributes roles to players based on game settings.
// It shuffles the roles randomly and assigns one to each player.
// Returns an error if the number of roles doesn't match the number of players.
func AssignRoles(players []*entities.Player, settings entities.GameSettings) error {
	// Build the list of roles from settings
	roles := make([]entities.RoleType, 0)
	for roleType, count := range settings.Roles {
		for i := 0; i < count; i++ {
			roles = append(roles, roleType)
		}
	}

	// Defensive validation: ensure role count matches player count
	if len(roles) != len(players) {
		return fmt.Errorf("role count mismatch: %d roles for %d players", len(roles), len(players))
	}

	// Shuffle roles using Fisher-Yates algorithm
	rand.Shuffle(len(roles), func(i, j int) {
		roles[i], roles[j] = roles[j], roles[i]
	})

	// Assign roles to players
	for i, player := range players {
		roleType := roles[i]
		role := factories.GetNewRole(roleType)
		player.AssignRole(role)
	}

	return nil
}
