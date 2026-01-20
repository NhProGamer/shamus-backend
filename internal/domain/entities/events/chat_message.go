package events

import "shamus-backend/internal/domain/entities"

const EventTypeChatMessage entities.EventType = "chat_message"

type ChatMessageEvent struct {
	PlayerID entities.PlayerID `json:"playerID"`
	Message  string            `json:"message"`
	Channel  string            `json:"channel"`
}

func NewChatMessageEvent(playerID entities.PlayerID) entities.Event[ChatMessageEvent] {
	return entities.Event[ChatMessageEvent]{
		Channel: entities.EventChannelGameEvent,
		Type:    EventTypeConnexion,
		Data: ChatMessageEvent{
			PlayerID: playerID,
		},
	}
}
