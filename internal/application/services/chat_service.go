package services

import (
	"shamus-backend/internal/domain/entities"
	"shamus-backend/internal/domain/entities/events"
	"shamus-backend/internal/domain/helpers"
	"shamus-backend/internal/domain/ports"
)

// ChatService implements the ports.ChatService interface
type ChatService struct{}

// NewChatService creates a new ChatService
func NewChatService() ports.ChatService {
	return &ChatService{}
}

// CanSendToChannel checks if a player can send messages to a specific channel
// Rules:
// - Village channel: only during day phases (Day/Vote), player must be alive
// - Werewolf channel: only during night, player must be alive and a werewolf
// - Lovers channel: only during night, player must be alive and a lover
func (s *ChatService) CanSendToChannel(
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
		return helpers.IsDayPhase(gamePhase)

	case events.ChatChannelWerewolf:
		return gamePhase == entities.PhaseNight && helpers.IsWerewolf(sender)

	case events.ChatChannelLovers:
		return gamePhase == entities.PhaseNight && helpers.IsLover(sender)

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
func (s *ChatService) GetChannelRecipients(
	channel events.ChatChannel,
	allPlayers []*entities.Player,
	gamePhase entities.GamePhase,
) []*entities.Player {
	recipients := make([]*entities.Player, 0)

	switch channel {
	case events.ChatChannelVillage:
		// During day, everyone can receive village messages
		if helpers.IsDayPhase(gamePhase) {
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
			if helpers.IsWerewolf(p) {
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
			if helpers.IsLover(p) {
				recipients = append(recipients, p)
			}
		}
		return recipients

	default:
		return recipients
	}
}
