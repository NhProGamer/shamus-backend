package entities

import "encoding/json"

type EventType string

type EventChannel string

const (
	EventChannelGameEvent EventChannel = "game_event"
	EventChannelConnexion EventChannel = "conn_event"
	EventChannelSettings  EventChannel = "settings_event"
	EventChannelTimer     EventChannel = "timer_event"
)

type Event[T any] struct {
	Channel EventChannel `json:"channel"`
	Type    EventType    `json:"type"`
	Data    T            `json:"data"`
}

type RawEvent = Event[json.RawMessage]

func (e Event[T]) ToRawEvent() RawEvent {
	data, _ := json.Marshal(e.Data)
	return RawEvent{
		Channel: e.Channel,
		Type:    e.Type,
		Data:    data,
	}
}
