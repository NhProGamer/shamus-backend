package entities

import "encoding/json"

type ActionType string

type ActionChannel string

const (
	ActionChannelAction ActionChannel = "action"
	ActionChannelError  ActionChannel = "error"
)

type Action[T any] struct {
	ID      int16
	Channel ActionChannel `json:"channel"`
	Type    ActionType    `json:"type"`
	Data    T             `json:"data"`
}

type RawAction = Action[json.RawMessage]

func (e Action[T]) ToRawAction() RawAction {
	data, _ := json.Marshal(e.Data)
	return RawAction{
		ID:      e.ID,
		Channel: e.Channel,
		Type:    e.Type,
		Data:    data,
	}
}
