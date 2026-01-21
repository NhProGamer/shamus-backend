package events

import "shamus-backend/internal/domain/entities"

const EventTypeChatMessage entities.EventType = "chat_message"

// ChatChannel represents the different chat channels available in the game
type ChatChannel string

const (
	// ChatChannelVillage is the main chat channel for all players during day
	ChatChannelVillage ChatChannel = "village"
	// ChatChannelWerewolf is the private channel for werewolves during night
	ChatChannelWerewolf ChatChannel = "werewolf"
	// ChatChannelLovers is the private channel for lovers during night
	ChatChannelLovers ChatChannel = "lovers"
)

// IsValidChatChannel checks if the given channel is a valid chat channel
func IsValidChatChannel(channel string) bool {
	switch ChatChannel(channel) {
	case ChatChannelVillage, ChatChannelWerewolf, ChatChannelLovers:
		return true
	default:
		return false
	}
}

type ChatMessageEventData struct {
	PlayerID entities.PlayerID `json:"playerID"`
	Nickname string            `json:"nickname,omitempty"`
	Message  string            `json:"message"`
	Channel  ChatChannel       `json:"channel"`
}

func NewChatMessageEvent(playerID entities.PlayerID, nickname string, message string, channel ChatChannel) entities.Event[ChatMessageEventData] {
	return entities.Event[ChatMessageEventData]{
		Channel: entities.EventChannelGameEvent,
		Type:    EventTypeChatMessage,
		Data: ChatMessageEventData{
			PlayerID: playerID,
			Nickname: nickname,
			Message:  message,
			Channel:  channel,
		},
	}
}
