package services

import (
	"shamus-backend/internal/domain/entities"
	"shamus-backend/internal/domain/entities/events"
	"shamus-backend/internal/domain/helpers"
	"shamus-backend/internal/domain/ports"
)

// VisibilityService implements the ports.VisibilityService interface
type VisibilityService struct{}

// NewVisibilityService creates a new VisibilityService
func NewVisibilityService() ports.VisibilityService {
	return &VisibilityService{}
}

// BuildPlayersDetailsForPlayer returns the player list with appropriate role visibility
// based on the viewer's role and the current game phase.
func (s *VisibilityService) BuildPlayersDetailsForPlayer(
	viewer *entities.Player,
	allPlayers []*entities.Player,
	gamePhase entities.GamePhase,
) []events.PlayersDetailsData {
	result := make([]events.PlayersDetailsData, 0, len(allPlayers))

	viewerIsWerewolf := helpers.IsWerewolf(viewer)

	for _, p := range allPlayers {
		detail := events.PlayersDetailsData{
			ID:              p.ID,
			Username:        p.Username,
			Alive:           p.IsAlive,
			Target:          p.VotedFor,
			ConnectionState: p.ConnectionState,
			Role:            nil, // Role is hidden by default
		}

		// Determine if role should be revealed
		if s.shouldRevealRole(viewer, p, viewerIsWerewolf, gamePhase) {
			roleType := p.Role.GetType()
			detail.Role = &roleType
		}

		result = append(result, detail)
	}
	return result
}

// shouldRevealRole determines if target's role should be visible to viewer
func (s *VisibilityService) shouldRevealRole(
	viewer, target *entities.Player,
	viewerIsWerewolf bool,
	phase entities.GamePhase,
) bool {
	// Target has no role assigned yet (game not started)
	if target.Role == nil {
		return false
	}

	// 1. A player always sees their own role
	if viewer.ID == target.ID {
		return true
	}

	// 2. Dead players' roles are revealed during day phases (Day or Vote)
	if !target.IsAlive && helpers.IsDayPhase(phase) {
		return true
	}

	// 3. Werewolves see other werewolves' roles
	if viewerIsWerewolf && helpers.IsWerewolf(target) {
		return true
	}

	return false
}
