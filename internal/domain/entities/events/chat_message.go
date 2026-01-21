package events

import "shamus-backend/internal/domain/entities"

const EventTypeChatMessage entities.EventType = "chat_message"

type ChatMessageEventData struct {
	PlayerID entities.PlayerID `json:"playerID"`
	Nickname string            `json:"nickname,omitempty"`
	Message  string            `json:"message"`
	Channel  string            `json:"channel"`
}

func NewChatMessageEvent(playerID entities.PlayerID, nickname string, message string, channel string) entities.Event[ChatMessageEventData] {
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
