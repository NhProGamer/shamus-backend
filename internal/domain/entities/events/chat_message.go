package events

import "shamus-backend/internal/domain/entities"

const EventTypeChatMessage entities.EventType = "chat_message"

type ChatMessageEvent struct {
	PlayerID entities.PlayerID `json:"playerID"`
	Nickname string            `json:"nickname,omitempty"`
	Message  string            `json:"message"`
	Channel  string            `json:"channel"`
}

func NewChatMessageEvent(playerID entities.PlayerID, nickname string, message string, channel string) entities.Event[ChatMessageEvent] {
	return entities.Event[ChatMessageEvent]{
		Channel: entities.EventChannelGameEvent,
		Type:    EventTypeChatMessage,
		Data: ChatMessageEvent{
			PlayerID: playerID,
			Nickname: nickname,
			Message:  message,
			Channel:  channel,
		},
	}
}
