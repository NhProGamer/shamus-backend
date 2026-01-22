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
// TODO: Implement Lovers system - this helper is ready but needs:
// - Cupid role that assigns ClanLovers to two players on first night
// - Death cascade: when one lover dies, the other dies too
// - Win condition: if only lovers survive, they win regardless of their original clans
func IsLover(p *entities.Player) bool {
	return IsPlayerInClan(p, entities.ClanLovers)
}
