package entities

import "encoding/json"

type EventType string

type EventChannel string

const (
	EventChannelGameEvent  EventChannel = "game_event"
	EventChannelConnection EventChannel = "conn_event"
	EventChannelSettings   EventChannel = "settings_event"
	EventChannelTimer      EventChannel = "timer_event"
	EventChannelAction     EventChannel = "action_event"
)

type Event[T any] struct {
	Channel EventChannel `json:"channel"`
	Type    EventType    `json:"type"`
	Data    T            `json:"data"`
}

type RawEvent = Event[json.RawMessage]

func (e Event[T]) ToRawEvent() (RawEvent, error) {
	data, err := json.Marshal(e.Data)
	if err != nil {
		return RawEvent{}, err
	}
	return RawEvent{
		Channel: e.Channel,
		Type:    e.Type,
		Data:    data,
	}, nil
}
