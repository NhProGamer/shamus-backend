package app

import (
	"shamus-backend/internal/domain/entities"
	"shamus-backend/internal/domain/entities/events"
	"shamus-backend/internal/domain/ports"
)

// ChatServiceImpl implements the ChatService interface
type ChatServiceImpl struct{}

// NewChatService creates a new ChatService
func NewChatService() ports.ChatService {
	return &ChatServiceImpl{}
}

// CanSendToChannel checks if a player can send messages to a specific channel
// Rules:
// - Village channel: only during day phases (Day/Vote), player must be alive
// - Werewolf channel: only during night, player must be alive and a werewolf
// - Lovers channel: only during night, player must be alive and a lover
func (s *ChatServiceImpl) CanSendToChannel(
	sender *entities.Player,
	channel events.ChatChannel,
	gamePhase entities.GamePhase,
) bool {
	// Dead players cannot send messages
	if !sender.IsAlive {
		return false
	}

	switch channel {
	case events.ChatChannelVillage:
		return s.isDayPhase(gamePhase)

	case events.ChatChannelWerewolf:
		return gamePhase == entities.PhaseNight && s.isWerewolf(sender)

	case events.ChatChannelLovers:
		return gamePhase == entities.PhaseNight && s.isLover(sender)

	default:
		return false
	}
}

// GetChannelRecipients returns the list of players who can receive messages
// on a specific channel
// Rules:
// - Village channel: all players receive (during day phases)
// - Werewolf channel: only werewolves receive (during night)
// - Lovers channel: only lovers receive (during night)
func (s *ChatServiceImpl) GetChannelRecipients(
	channel events.ChatChannel,
	allPlayers []*entities.Player,
	gamePhase entities.GamePhase,
) []*entities.Player {
	recipients := make([]*entities.Player, 0)

	switch channel {
	case events.ChatChannelVillage:
		// During day, everyone can receive village messages
		if s.isDayPhase(gamePhase) {
			return allPlayers
		}
		// During night, no one receives village messages
		return recipients

	case events.ChatChannelWerewolf:
		// Only werewolves receive werewolf channel messages during night
		if gamePhase != entities.PhaseNight {
			return recipients
		}
		for _, p := range allPlayers {
			if s.isWerewolf(p) {
				recipients = append(recipients, p)
			}
		}
		return recipients

	case events.ChatChannelLovers:
		// Only lovers receive lovers channel messages during night
		if gamePhase != entities.PhaseNight {
			return recipients
		}
		for _, p := range allPlayers {
			if s.isLover(p) {
				recipients = append(recipients, p)
			}
		}
		return recipients

	default:
		return recipients
	}
}

// isDayPhase returns true if the phase is during daytime (Day or Vote)
func (s *ChatServiceImpl) isDayPhase(phase entities.GamePhase) bool {
	return phase == entities.PhaseDay || phase == entities.PhaseVote
}

// isWerewolf checks if a player belongs to the werewolf clan
func (s *ChatServiceImpl) isWerewolf(p *entities.Player) bool {
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

// isLover checks if a player has the lover status
// TODO: Implement lover detection when Cupid role is added
func (s *ChatServiceImpl) isLover(p *entities.Player) bool {
	if p == nil || p.Role == nil {
		return false
	}
	// For now, check if player has ClanLovers
	// This will be set by Cupid's ability when implemented
	for _, clan := range p.Role.GetClans() {
		if clan == entities.ClanLovers {
			return true
		}
	}
	return false
}
