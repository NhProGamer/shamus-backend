package app

import (
	"shamus-backend/internal/domain/entities"
	"shamus-backend/internal/domain/entities/events"
	"shamus-backend/internal/domain/ports"
)

// VisibilityServiceImpl implements the VisibilityService interface
type VisibilityServiceImpl struct{}

// NewVisibilityService creates a new VisibilityService
func NewVisibilityService() ports.VisibilityService {
	return &VisibilityServiceImpl{}
}

// BuildPlayersDetailsForPlayer returns the player list with appropriate role visibility
// based on the viewer's role and the current game phase.
func (s *VisibilityServiceImpl) BuildPlayersDetailsForPlayer(
	viewer *entities.Player,
	allPlayers []*entities.Player,
	gamePhase entities.GamePhase,
) []events.PlayersDetailsData {
	result := make([]events.PlayersDetailsData, 0, len(allPlayers))

	viewerIsWerewolf := s.isWerewolf(viewer)

	for _, p := range allPlayers {
		detail := events.PlayersDetailsData{
			ID:             p.ID,
			Username:       p.Username,
			Alive:          p.IsAlive,
			Target:         p.VotedFor,
			ConnexionState: p.ConnectionState,
			Role:           nil, // Role is hidden by default
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
func (s *VisibilityServiceImpl) shouldRevealRole(
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
	if !target.IsAlive && s.isDayPhase(phase) {
		return true
	}

	// 3. Werewolves see other werewolves' roles
	if viewerIsWerewolf && s.isWerewolf(target) {
		return true
	}

	return false
}

// isDayPhase returns true if the phase is during daytime (Day or Vote)
func (s *VisibilityServiceImpl) isDayPhase(phase entities.GamePhase) bool {
	return phase == entities.PhaseDay || phase == entities.PhaseVote
}

// isWerewolf checks if a player belongs to the werewolf clan
func (s *VisibilityServiceImpl) isWerewolf(p *entities.Player) bool {
	if p == nil || p.Role == nil {
		return false
	}
	for _, clan := range p.Role.GetClans() {
		if clan == entities.ClanWerewolf {
			return true
		}
	}
	return false
}
