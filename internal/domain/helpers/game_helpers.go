package helpers

import "shamus-backend/internal/domain/entities"

// IsDayPhase returns true if the phase is during daytime (Day or Vote)
func IsDayPhase(phase entities.GamePhase) bool {
	return phase == entities.PhaseDay || phase == entities.PhaseVote || phase == entities.PhaseStart
}

// IsPlayerInClan checks if a player belongs to a specific clan
func IsPlayerInClan(p *entities.Player, clan entities.Clan) bool {
	if p == nil || p.Role == nil {
		return false
	}
	for _, c := range p.Role.GetClans() {
		if c == clan {
			return true
		}
	}
	return false
}

// IsWerewolf checks if a player belongs to the werewolf clan
func IsWerewolf(p *entities.Player) bool {
	return IsPlayerInClan(p, entities.ClanWerewolf)
}

// IsLover checks if a player has the lovers clan
func IsLover(p *entities.Player) bool {
	return IsPlayerInClan(p, entities.ClanLovers)
}
